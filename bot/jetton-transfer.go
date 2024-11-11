package bot

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/binary"
	"fmt"

	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/tlb"
	"github.com/xssnick/tonutils-go/ton"
	"github.com/xssnick/tonutils-go/ton/wallet"
	"github.com/xssnick/tonutils-go/tvm/cell"
	"golang.org/x/net/context"
)

func JettonTransferFromG(ctx context.Context,
	client ton.APIClientWrapped,
	botprivateKey ed25519.PrivateKey,
	pkOfG ed25519.PrivateKey,
	jettonWalletOfGAddr *address.Address,
	destAddr *address.Address,
	amount tlb.Coins,
) error {
	botAddr := botAddress(botprivateKey.Public().(ed25519.PublicKey))
	gAddr := gAddress(pkOfG.Public().(ed25519.PublicKey), botAddr)

	masterBlock, err := client.GetMasterchainInfo(ctx)
	if err != nil {
		return err
	}

	account, err := client.WaitForBlock(masterBlock.SeqNo).GetAccount(ctx, masterBlock, gAddr)
	if err != nil {
		return err
	}

	fmt.Println("Dest address:", destAddr.String())
	fmt.Println("jetton wallet of G", jettonWalletOfGAddr.String())
	fmt.Println("G address:", gAddr.String())
	fmt.Println("G balance:", account.State.Balance)
	seqno, err := getSeqno(ctx, client, masterBlock, gAddr)
	if err != nil {
		return err
	}
	fmt.Println("Bot seqno:", seqno)

	gWallet := NewWallet(ctx, client, G, pkOfG, botAddr, seqno)

	msg, err := gWallet.BuildJettonTransfer(jettonWalletOfGAddr, destAddr, amount, tlb.MustFromTON("0.1"))
	if err != nil {
		return err
	}

	fmt.Println("Sending message", msg)
	return gWallet.Send(ctx, 0, msg)
}

type TransferPayload struct {
	_                   tlb.Magic        `tlb:"#0f8a7ea5"`
	QueryID             uint64           `tlb:"## 64"`
	Amount              tlb.Coins        `tlb:"."`
	Destination         *address.Address `tlb:"addr"`
	ResponseDestination *address.Address `tlb:"addr"`
	CustomPayload       *cell.Cell       `tlb:"maybe ^"`
	ForwardTONAmount    tlb.Coins        `tlb:"."`
	ForwardPayload      *cell.Cell       `tlb:"either . ^"`
}

func (w *Wallet) BuildJettonTransfer(jettonWalletAddr *address.Address,
	destAddr *address.Address,
	jettonAmount tlb.Coins, amountForwardTON tlb.Coins) (_ *wallet.Message, err error) {
	body, err := w.buildTransferPayloadV2(destAddr, nil, jettonAmount, amountForwardTON, nil, nil)
	if err != nil {
		return nil, err
	}

	return &wallet.Message{
		Mode: wallet.PayGasSeparately + wallet.IgnoreErrors,
		InternalMessage: &tlb.InternalMessage{
			IHRDisabled: true,
			Bounce:      false,
			DstAddr:     jettonWalletAddr,
			Amount:      tlb.MustFromTON("0.3"),
			Body:        body,
		},
	}, nil
}

func (w *Wallet) buildTransferPayloadV2(to, responseTo *address.Address, amountCoins, amountForwardTON tlb.Coins, payloadForward, customPayload *cell.Cell) (*cell.Cell, error) {
	if payloadForward == nil {
		payloadForward = cell.BeginCell().EndCell()
	}

	buf := make([]byte, 8)
	if _, err := rand.Read(buf); err != nil {
		return nil, err
	}
	rnd := binary.LittleEndian.Uint64(buf)

	body, err := tlb.ToCell(TransferPayload{
		QueryID:             rnd,
		Amount:              amountCoins,
		Destination:         to,
		ResponseDestination: responseTo,
		CustomPayload:       customPayload,
		ForwardTONAmount:    amountForwardTON,
		ForwardPayload:      payloadForward,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to convert TransferPayload to cell: %w", err)
	}

	return body, nil
}
