package detector

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/cmingxu/dedust/model"
	"github.com/cmingxu/dedust/utils"
	mywallet "github.com/cmingxu/dedust/wallet"
	"github.com/patrickmn/go-cache"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/xssnick/tonutils-go/address"
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

func (d *Detector) parseTrade(msg *tlb.ExternalMessage) (*model.Pool, *model.Trade, error) {
	dest := msg.DestAddr()
	dest.SetBounce(true)
	// log.Debug().Msgf("ExternalIn Dst: %s", msg.DestAddr().String())

	slice := msg.Body.BeginParse()
	if slice.BitsLeft() < 32 {
		return nil, nil, errors.New("magic wallet code not enough bits")
	}

	magic := slice.MustPreloadUInt(32)
	trade := model.Trade{
		Hash:      hex.EncodeToString(msg.Body.Hash()),
		Address:   msg.DestAddr().String(),
		FirstSeen: time.Now(),
	}

	var pool *model.Pool
	var err error

	if msg.Body.RefsNum() != 1 {
		trade.HasMultipleActions = true
	}

	internalMsg := tlb.InternalMessage{}
	if magic == V5R1Magic {
		trade.WalletType = model.WalletTypeV5R1
		msg := mywallet.V5R1Header{}
		if err := tlb.LoadFromCell(&msg, slice); err != nil {
			return nil, &trade, errors.Wrap(err, "failed to load V5R1Header")
		}
		if err := tlb.LoadFromCell(&internalMsg,
			msg.Action.OutMsg.BeginParse()); err != nil {
			return nil, &trade, errors.Wrap(err, "failed to load InternalMessage")
		}
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
		goto FINISH

	CORRECT:
		if err := tlb.LoadFromCell(&internalMsg, internalCell.BeginParse()); err != nil {
			log.Debug().Err(err).Msg("failed to load InternalMessage")
			trade.WalletType = model.WalletTypeBot
			goto FINISH
		}
	}

	pool, err = d.parseInternalMessage(&internalMsg, &trade)
	if err != nil {
		// log.Debug().Err(err).Msg("failed to parse internal message")
		return nil, &trade, err
	}

	if pool != nil {
		trade.PoolAddr = pool.Address
		trade.LatestReserve0 = pool.Asset0Reserve
		trade.LatestReserve1 = pool.Asset1Reserve
		trade.LatestPoolLt = pool.Lt
		trade.PoolUpdateAt = pool.UpdatedAt
	}

FINISH:
	d.p("Finish Internal Message: %+v\n", internalMsg)
	trade.AmountIn = internalMsg.Amount.Nano().String()
	return pool, &trade, d.saveTrade(&trade)
}

func (d *Detector) parseInternalMessage(msg *tlb.InternalMessage, trade *model.Trade) (*model.Pool, error) {
	var (
		poolAddr *address.Address
		pool     *model.Pool = nil
	)

	bodySlice := msg.Body.BeginParse()
	if bodySlice.BitsLeft() < 32 {
		return pool, errors.New("opcode not enough bits")
	}

	opcode := bodySlice.MustPreloadUInt(32)
	switch opcode {
	case 0:
		// log.Debug().Msgf("InternalMessage Opcode 0 transfer from %s to %s amount %s",
		//	trade.Address, msg.DstAddr.String(), msg.Amount.Nano().String())
		d.tonTransferCache.Set(trade.Address, struct{}{}, cache.DefaultExpiration)
	case DedustNativeSwap:
		nativeSwap, err := decodeDedustBuy(msg.Body)
		if err != nil {
			return pool, errors.Wrap(err, "failed to decode DedustNativeSwap")
		}
		log.Debug().Msgf("(Dedust BUY) %s NativeSwap: %+v", trade.Address, nativeSwap)
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

		if nativeSwap.SwapStep.PoolAddr != nil {
			poolAddr = nativeSwap.SwapStep.PoolAddr
		}

	case JettonTransfer:
		d.sellingCache.Set(trade.PoolAddr, struct{}{}, cache.DefaultExpiration)

		transfer, err := decodeDedustSell(msg.Body)
		if err != nil {
			return pool, errors.Wrap(err, "failed to decode JettonTransfer")
		}

		log.Debug().Msgf("(Dedust SELL) JettonTransfer: %s %+v", trade.Address, transfer)
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

		if swapStep.PoolAddr != nil {
			poolAddr = swapStep.PoolAddr
		}

	case JettonBurn:
		log.Debug().Msgf("(Dedust BURN) LP JettonBurn: %s", trade.Address)
		poolAddr, err := decodeDedustWithdrawLP(msg.Body)
		if err != nil {
			d.sellingCache.Set(poolAddr, struct{}{}, cache.DefaultExpiration)
			// https://tonviewer.com/transaction/d5cb3a50271a86222fbbd269c64a65e97d25fea16aaa43ce9ea37a08e7e2d7b0
		}

	default:
		return pool, errors.New("internal unknown opcode")
	}

	if poolAddr != nil {
		var ok bool
		if pool, ok = d.poolMap[utils.RawAddr(poolAddr)]; !ok {
			return pool, fmt.Errorf("pool not found %s", poolAddr.String())
		}
	}

	return pool, nil
}

func (d *Detector) saveTrade(trade *model.Trade) error {
	return trade.SaveToDB(d.db)
}

func decodeDedustBuy(cell *cell.Cell) (*NativeSwap, error) {
	var nativeSwap NativeSwap
	if err := tlb.LoadFromCell(&nativeSwap, cell.BeginParse()); err != nil {
		return nil, errors.Wrap(err, "failed to load SwapRequest")
	}
	return &nativeSwap, nil
}

// aka. this is the sell transaction
func decodeDedustSell(cell *cell.Cell) (*JettonTransferParams, error) {
	var transfer JettonTransferParams
	if err := tlb.LoadFromCell(&transfer, cell.BeginParse()); err != nil {
		return nil, errors.Wrap(err, "failed to load JettonTransfer")
	}
	return &transfer, nil
}

func decodeDedustWithdrawLP(cell *cell.Cell) (poolAddr string, err error) {
	return "", nil
}
