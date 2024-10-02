package detector

import (
	"encoding/base64"
	"encoding/hex"

	"github.com/cmingxu/dedust/model"
	"github.com/cmingxu/dedust/utils"
	mywallet "github.com/cmingxu/dedust/wallet"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/xssnick/tonutils-go/tlb"
	"github.com/xssnick/tonutils-go/tvm/cell"
)

const V5R1Magic = 0x7369676e

// turn a BOC string into a tlb.Message
func (d *Detector) outerMessageFromBOC(boc string) (*tlb.Message, error) {
	var msg tlb.Message
	rawBoc, err := base64.StdEncoding.DecodeString(boc)
	if err != nil {
		return nil, err
	}

	c, err := cell.FromBOC(rawBoc)
	if err != nil {
		return nil, err
	}

	if err := tlb.LoadFromCell(&msg, c.BeginParse()); err != nil {
		return nil, err
	}

	return &msg, nil
}

// handleExternalMsg handles external messages
func (d *Detector) handleExternalMsg(pool *model.Pool, msg *tlb.ExternalMessage) error {
	log.Debug().Msgf("=========================================")
	log.Debug().Msgf("=======         BEGIN        ============")
	log.Debug().Msgf("=========================================")

	log.Debug().Msgf("ExternalIn Dst: %s", msg.DestAddr().String())
	// log.Debug().Msgf("%s", msg.Body.Dump())

	slice := msg.Body.BeginParse()
	magic := slice.MustPreloadUInt(32)

	trade := model.Trade{
		Hash:     hex.EncodeToString(msg.Body.Hash()),
		PoolAddr: pool.Address,
		Address:  msg.DestAddr().String(),
	}

	internalMsg := tlb.InternalMessage{}
	if magic == V5R1Magic {
		trade.WalletType = model.WalletTypeV5R1
		msg := mywallet.V5R1Header{}
		if err := tlb.LoadFromCell(&msg, slice); err != nil {
			return errors.Wrap(err, "failed to load V5R1Header")
		}
		if err := tlb.LoadFromCell(&internalMsg,
			msg.Action.OutMsg.BeginParse()); err != nil {
			return errors.Wrap(err, "failed to load InternalMessage")
		}
	} else {
		var internalCell *cell.Cell
		msgv3 := mywallet.V3Header{}
		msgv4 := mywallet.V4R2Header{}
		if err := tlb.LoadFromCell(&msgv4, slice); err == nil {
			trade.WalletType = model.WalletTypeV4R2
			internalCell = msgv4.Body
			goto CORRECT
		}

		if err := tlb.LoadFromCell(&msgv3, slice); err == nil {
			trade.WalletType = model.WalletTypeV3
			internalCell = msgv3.Body
			goto CORRECT
		}

		trade.WalletType = model.WalletTypeBot
		goto FINISH

	CORRECT:
		if err := tlb.LoadFromCell(&internalMsg, internalCell.BeginParse()); err != nil {
			log.Debug().Err(err).Msg("failed to load InternalMessage")
			trade.WalletType = model.WalletTypeBot
			goto FINISH
		}
	}

	if err := d.parseInternalMessage(&internalMsg, &trade); err != nil {
		log.Debug().Err(err).Msg("failed to parse internal message")
	}

FINISH:
	trade.AmountIn = utils.CoinsToFloatTON(internalMsg.Amount)
	return d.saveTrade(&trade)
}

func (d *Detector) parseInternalMessage(msg *tlb.InternalMessage, trade *model.Trade) error {
	opcode := msg.Body.BeginParse().MustPreloadUInt(32)

	switch opcode {
	case DedustNativeSwap:
		nativeSwap, err := decodeDedustNativeSwap(msg.Body)
		if err != nil {
			return errors.Wrap(err, "failed to decode DedustNativeSwap")
		}
		log.Debug().Msgf("(BUY) NativeSwap: %+v", nativeSwap)

		trade.TradeType = model.TradeTypeBuy
		trade.SwapType = model.SwapTypeNative
		trade.Amount = utils.CoinsToFloatTON(nativeSwap.Amount)
		trade.Limit = nativeSwap.SwapStep.SwapStepParams.Limit.String()
		trade.Recipient = nativeSwap.SwapParams.Recipient.String()
		trade.Referrer = nativeSwap.SwapParams.Referrer.String()
		if nativeSwap.SwapParams.FullfillPayload != nil {
			trade.FullfillBOC = hex.EncodeToString(nativeSwap.SwapParams.FullfillPayload.ToBOC())
		}
		if nativeSwap.SwapParams.RejectPayload != nil {
			trade.RejectBOC = hex.EncodeToString(nativeSwap.SwapParams.RejectPayload.ToBOC())
		}

	case JettonTransfer:
		transfer, err := decodeJettonTransfer(msg.Body)
		if err != nil {
			return errors.Wrap(err, "failed to decode JettonTransfer")
		}
		log.Debug().Msgf("(SELL) JettonTransfer: %+v", transfer)
		trade.TradeType = model.TradeTypeSell
		trade.SwapType = model.SwapTypeJetton
		trade.Amount = utils.CoinsToFloatTON(transfer.Amount)
		trade.TokenAmount = transfer.Amount.String()
		swapStep := transfer.ForwardPayload.SwapStep
		swapParams := transfer.ForwardPayload.SwapParams
		trade.Limit = swapStep.SwapStepParams.Limit.String()
		trade.Recipient = swapParams.Recipient.String()
		trade.Referrer = swapParams.Referrer.String()
		if swapParams.FullfillPayload != nil {
			trade.FullfillBOC = hex.EncodeToString(swapParams.FullfillPayload.ToBOC())
		}

		if swapParams.RejectPayload != nil {
			trade.RejectBOC = hex.EncodeToString(swapParams.RejectPayload.ToBOC())
		}
	default:
		return errors.New("unknown opcode")
	}

	return nil
}

func (d *Detector) saveTrade(trade *model.Trade) error {
	return trade.SaveToDB(d.db)
}

func decodeDedustNativeSwap(cell *cell.Cell) (*NativeSwap, error) {
	var nativeSwap NativeSwap
	if err := tlb.LoadFromCell(&nativeSwap, cell.BeginParse()); err != nil {
		return nil, errors.Wrap(err, "failed to load SwapRequest")
	}
	return &nativeSwap, nil
}

// aka. this is the sell transaction
func decodeJettonTransfer(cell *cell.Cell) (*JettonTransferParams, error) {
	var transfer JettonTransferParams
	if err := tlb.LoadFromCell(&transfer, cell.BeginParse()); err != nil {
		return nil, errors.Wrap(err, "failed to load JettonTransfer")
	}
	return &transfer, nil
}
