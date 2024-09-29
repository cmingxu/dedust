package bot

import (
	"bytes"
	"errors"
	"fmt"
	"time"

	"github.com/xssnick/tonutils-go/tlb"
	"github.com/xssnick/tonutils-go/ton"
	"github.com/xssnick/tonutils-go/ton/wallet"
	"github.com/xssnick/tonutils-go/tvm/cell"
	"golang.org/x/net/context"
)

func (w *BotWallet) Send(ctx context.Context, message *wallet.Message, waitConfirmation ...bool) error {
	return w.SendMany(ctx, []*wallet.Message{message}, waitConfirmation...)
}

func (w *BotWallet) SendMany(ctx context.Context, messages []*wallet.Message, waitConfirmation ...bool) error {
	_, _, _, err := w.sendMany(ctx, messages, waitConfirmation...)
	return err
}

// SendManyGetInMsgHash returns hash of external incoming message payload.
func (w *BotWallet) SendManyGetInMsgHash(ctx context.Context, messages []*wallet.Message, waitConfirmation ...bool) ([]byte, error) {
	_, _, inMsgHash, err := w.sendMany(ctx, messages, waitConfirmation...)
	return inMsgHash, err
}

// SendManyWaitTxHash always waits for tx block confirmation and returns found tx hash in block.
func (w *BotWallet) SendManyWaitTxHash(ctx context.Context, messages []*wallet.Message) ([]byte, error) {
	tx, _, _, err := w.sendMany(ctx, messages, true)
	if err != nil {
		return nil, err
	}
	return tx.Hash, err
}

// SendManyWaitTransaction always waits for tx block confirmation and returns found tx.
func (w *BotWallet) SendManyWaitTransaction(ctx context.Context, messages []*wallet.Message) (*tlb.Transaction, *ton.BlockIDExt, error) {
	tx, block, _, err := w.sendMany(ctx, messages, true)
	return tx, block, err
}

// SendWaitTransaction always waits for tx block confirmation and returns found tx.
func (w *BotWallet) SendWaitTransaction(ctx context.Context, message *wallet.Message) (*tlb.Transaction, *ton.BlockIDExt, error) {
	return w.SendManyWaitTransaction(ctx, []*wallet.Message{message})
}

func (w *BotWallet) sendMany(ctx context.Context, messages []*wallet.Message, waitConfirmation ...bool) (tx *tlb.Transaction, block *ton.BlockIDExt, inMsgHash []byte, err error) {
	block, err = w.api.CurrentMasterchainInfo(ctx)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to get block: %w", err)
	}

	ext, err := w.BuildExternalMessageForMany(ctx, messages)
	if err != nil {
		return nil, nil, nil, err
	}
	inMsgHash = ext.Body.Hash()

	if err = w.api.SendExternalMessage(ctx, ext); err != nil {
		return nil, nil, nil, fmt.Errorf("failed to send message: %w", err)
	}

	fmt.Println("Sent external message", inMsgHash)

	if len(waitConfirmation) > 0 && waitConfirmation[0] {
		fmt.Println("Waiting for confirmation...")
		acc, err := w.api.WaitForBlock(block.SeqNo).GetAccount(ctx, block, w.addr)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("failed to get account state: %w", err)
		}

		tx, block, err = w.waitConfirmation(ctx, block, acc, ext)
		if err != nil {
			return nil, nil, nil, err
		}
	}

	return tx, block, inMsgHash, nil
}

func (w *BotWallet) BuildExternalMessageForMany(ctx context.Context, messages []*wallet.Message) (*tlb.ExternalMessage, error) {
	return w.PrepareExternalMessageForMany(ctx, messages)
}

// PrepareExternalMessageForMany - Prepares external message for wallet
// can be used directly for offline signing but custom fetchers should be defined in this case
func (w *BotWallet) PrepareExternalMessageForMany(ctx context.Context, messages []*wallet.Message) (_ *tlb.ExternalMessage, err error) {

	msg, err := w.BuildMessage(ctx, messages)
	if err != nil {
		return nil, err
	}

	// fuck *tlb.StateInit is not same as &tlb.StateInit{}
	var stateInit *tlb.StateInit

	return &tlb.ExternalMessage{
		DstAddr:   w.addr,
		StateInit: stateInit,
		Body:      msg,
	}, nil
}

func (w *BotWallet) BuildMessage(ctx context.Context, messages []*wallet.Message) (*cell.Cell, error) {
	if len(messages) > 4 {
		return nil, errors.New("for this type of wallet max 4 messages can be sent in the same time")
	}

	payload := cell.BeginCell().MustStoreUInt(uint64(SubwalletID), 32).
		MustStoreUInt(uint64(time.Now().Add(time.Duration(MessageTTL)*time.Second).UTC().Unix()), 32).
		MustStoreUInt(uint64(w.seq), 32).
		MustStoreInt(0, 8) // op

	for i, message := range messages {
		intMsg, err := tlb.ToCell(message.InternalMessage)
		if err != nil {
			return nil, fmt.Errorf("failed to convert internal message %d to cell: %w", i, err)
		}

		payload.MustStoreUInt(uint64(message.Mode), 8).MustStoreRef(intMsg)
	}

	sign := payload.EndCell().Sign(w.pk)
	msg := cell.BeginCell().MustStoreSlice(sign, 512).MustStoreBuilder(payload).EndCell()

	return msg, nil
}

func (w *BotWallet) waitConfirmation(ctx context.Context, block *ton.BlockIDExt, acc *tlb.Account, ext *tlb.ExternalMessage) (*tlb.Transaction, *ton.BlockIDExt, error) {
	if _, hasDeadline := ctx.Deadline(); !hasDeadline {
		// fallback timeout to not stuck forever with background context
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), 180*time.Second)
		defer cancel()
	}
	till, _ := ctx.Deadline()

	ctx = w.api.Client().StickyContext(ctx)

	for time.Now().Before(till) {
		blockNew, err := w.api.WaitForBlock(block.SeqNo + 1).GetMasterchainInfo(ctx)
		if err != nil {
			continue
		}

		accNew, err := w.api.WaitForBlock(blockNew.SeqNo).GetAccount(ctx, blockNew, w.addr)
		if err != nil {
			continue
		}
		block = blockNew

		if accNew.LastTxLT == acc.LastTxLT {
			// if not in block, maybe LS lost our message, send it again
			if err = w.api.SendExternalMessage(ctx, ext); err != nil {
				continue
			}

			continue
		}

		lastLt, lastHash := accNew.LastTxLT, accNew.LastTxHash

		// it is possible that > 5 new not related transactions will happen, and we should not lose our scan offset,
		// to prevent this we will scan till we reach last seen offset.
		for time.Now().Before(till) {
			// we try to get last 5 transactions, and check if we have our new there.
			txList, err := w.api.WaitForBlock(block.SeqNo).ListTransactions(ctx, w.addr, 5, lastLt, lastHash)
			if err != nil {
				continue
			}

			sawLastTx := false
			for i, transaction := range txList {
				if i == 0 {
					// get previous of the oldest tx, in case if we need to scan deeper
					lastLt, lastHash = txList[0].PrevTxLT, txList[0].PrevTxHash
				}

				if !sawLastTx && transaction.PrevTxLT == acc.LastTxLT &&
					bytes.Equal(transaction.PrevTxHash, acc.LastTxHash) {
					sawLastTx = true
				}

				if transaction.IO.In != nil && transaction.IO.In.MsgType == tlb.MsgTypeExternalIn {
					extIn := transaction.IO.In.AsExternalIn()
					if ext.StateInit != nil {
						if extIn.StateInit == nil {
							continue
						}

						if !bytes.Equal(ext.StateInit.Data.Hash(), extIn.StateInit.Data.Hash()) {
							continue
						}

						if !bytes.Equal(ext.StateInit.Code.Hash(), extIn.StateInit.Code.Hash()) {
							continue
						}
					}

					if !bytes.Equal(extIn.Body.Hash(), ext.Body.Hash()) {
						continue
					}

					return transaction, block, nil
				}
			}

			if sawLastTx {
				break
			}
		}
		acc = accNew
	}

	return nil, nil, errors.New("tx not confirmed")
}
