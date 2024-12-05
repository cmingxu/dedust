package detector

import (
	"encoding/base64"
	"encoding/hex"
	"time"

	"github.com/cmingxu/dedust/model"
	mywallet "github.com/cmingxu/dedust/wallet"
	"github.com/patrickmn/go-cache"
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

func (d *Detector) parseTrade(pool *model.Pool, msg *tlb.ExternalMessage) (*model.Trade, error) {
	log.Debug().Msgf("=========================================")
	log.Debug().Msgf("=======         BEGIN        ============")
	log.Debug().Msgf("=========================================")

	dest := msg.DestAddr()
	dest.SetBounce(true)
	log.Debug().Msgf("ExternalIn Dst: %s", msg.DestAddr().String())
	log.Debug().Msgf("ExternalIn Dst(Bounceable): %s", dest.String())
	log.Debug().Msgf("ExternalIn Dst: %s", msg.DestAddr().String())
	// log.Debug().Msgf("%s", msg.Body.Dump())
	log.Debug().Msgf("External Body RefNum: %d", msg.Body.RefsNum())

	d.p("=========================================\n")
	d.p("=======         BEGIN        ============\n")
	d.p("=========================================\n")
	dstAddr := msg.DestAddr()
	dstAddr.SetBounce(true)
	d.p("ExternalIn Dst: %s %s\n", dstAddr.String(), time.Now())
	d.p("ExternalIn Dst(Bounceable): %s", dest.String())
	d.p("External Body RefNum: %d\n", msg.Body.RefsNum())

	slice := msg.Body.BeginParse()
	magic := slice.MustPreloadUInt(32)

	trade := model.Trade{
		Hash:           hex.EncodeToString(msg.Body.Hash()),
		PoolAddr:       pool.Address,
		Address:        msg.DestAddr().String(),
		LatestReserve0: pool.Asset0Reserve,
		LatestReserve1: pool.Asset1Reserve,
		LatestPoolLt:   pool.Lt,
		PoolUpdateAt:   pool.UpdatedAt,
		FirstSeen:      time.Now(),
	}

	if msg.Body.RefsNum() != 1 {
		trade.HasMultipleActions = true
	}

	internalMsg := tlb.InternalMessage{}
	if magic == V5R1Magic {
		trade.WalletType = model.WalletTypeV5R1
		msg := mywallet.V5R1Header{}
		if err := tlb.LoadFromCell(&msg, slice); err != nil {
			return &trade, errors.Wrap(err, "failed to load V5R1Header")
		}
		if err := tlb.LoadFromCell(&internalMsg,
			msg.Action.OutMsg.BeginParse()); err != nil {
			return &trade, errors.Wrap(err, "failed to load InternalMessage")
		}
		d.p("V5 Internal Cell: %s\n", msg.Action.OutMsg.Dump())
		d.p("V5 Slice BitsLeft: %d RefNum %d \n", slice.BitsLeft(), slice.RefsNum())
		d.p("V5 msg empty bitsleft: %d, refnums:  %d\n", msg.Action.Empty.BitsSize(), msg.Action.Empty.RefsNum())
		if msg.Action.Empty.RefsNum() > 0 {
			trade.HasMultipleActions = true
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
		d.p("Unknown WalletType: (%s) %s\n", trade.WalletType, msg.DstAddr.String())
		goto FINISH

	CORRECT:
		d.p("Trade WalletType %s\n", trade.WalletType)
		d.p("V3/V4 branch Internal Cell: %s\n", internalCell.Dump())
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
	d.p("Finish Internal Message: %+v\n", internalMsg)
	trade.AmountIn = internalMsg.Amount.Nano().String()
	return &trade, d.saveTrade(&trade)
}

func (d *Detector) parseInternalMessage(msg *tlb.InternalMessage, trade *model.Trade) error {
	bodySlice := msg.Body.BeginParse()
	if bodySlice.BitsLeft() < 32 {
		return errors.New("not enough bits")
	}

	opcode := bodySlice.MustPreloadUInt(32)

	switch opcode {
	case DedustNativeSwap:
		nativeSwap, err := decodeDedustNativeSwap(msg.Body)
		if err != nil {
			return errors.Wrap(err, "failed to decode DedustNativeSwap")
		}
		log.Debug().Msgf("(BUY) NativeSwap: %+v", nativeSwap)
		d.p("NativeSwap: %+v\n", nativeSwap)
		d.p("NativeSwap(SwapParams): %+v\n", nativeSwap.SwapParams)
		d.p("NativeSwap(SwapStep): %+v\n", nativeSwap.SwapStep)

		trade.TradeType = model.TradeTypeBuy
		trade.SwapType = model.SwapTypeNative
		trade.Amount = nativeSwap.Amount.Nano().String()
		trade.Limit = nativeSwap.SwapStep.SwapStepParams.Limit.Nano().String()
		trade.Deadline = nativeSwap.SwapParams.Deadline
		trade.Recipient = nativeSwap.SwapParams.Recipient.String()
		trade.Referrer = nativeSwap.SwapParams.Referrer.String()
		if nativeSwap.SwapParams.FullfillPayload != nil {
			trade.FullfillBOC = hex.EncodeToString(nativeSwap.SwapParams.FullfillPayload.ToBOC())
		}
		if nativeSwap.SwapParams.RejectPayload != nil {
			trade.RejectBOC = hex.EncodeToString(nativeSwap.SwapParams.RejectPayload.ToBOC())
		}

		// 这说明这个 trade 是可以多步交易的，也是一个夹子， 这样的交易是有风险的
		if nativeSwap.SwapStep.SwapStepParams.Next != nil &&
			nativeSwap.SwapStep.SwapStepParams.Next.SwapStepParams != nil {
			trade.HasNextStep = true
		}

	case JettonTransfer:
		d.sellingCache.Set(trade.PoolAddr, struct{}{}, cache.DefaultExpiration)

		transfer, err := decodeJettonTransfer(msg.Body)
		if err != nil {
			return errors.Wrap(err, "failed to decode JettonTransfer")
		}
		log.Debug().Msgf("(SELL) JettonTransfer: %+v", transfer)
		trade.TradeType = model.TradeTypeSell
		trade.SwapType = model.SwapTypeJetton
		trade.Amount = transfer.Amount.Nano().String()
		trade.TokenAmount = transfer.Amount.String()
		swapStep := transfer.ForwardPayload.SwapStep
		swapParams := transfer.ForwardPayload.SwapParams
		trade.Limit = swapStep.SwapStepParams.Limit.Nano().String()
		trade.Deadline = swapParams.Deadline
		trade.Recipient = swapParams.Recipient.String()
		trade.Referrer = swapParams.Referrer.String()
		if swapParams.FullfillPayload != nil {
			trade.FullfillBOC = hex.EncodeToString(swapParams.FullfillPayload.ToBOC())
		}

		if swapParams.RejectPayload != nil {
			trade.RejectBOC = hex.EncodeToString(swapParams.RejectPayload.ToBOC())
		}
	case JettonBurn:
		d.sellingCache.Set(trade.PoolAddr, struct{}{}, cache.DefaultExpiration)
		// https://tonviewer.com/transaction/d5cb3a50271a86222fbbd269c64a65e97d25fea16aaa43ce9ea37a08e7e2d7b0
		log.Debug().Msg("(LP Burn)")
		d.p("LP Burn: %s\n", msg.Body.Dump())

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
