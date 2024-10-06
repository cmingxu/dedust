package detector

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"math/big"
	"strings"

	"github.com/cmingxu/dedust/model"
	"github.com/patrickmn/go-cache"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/samber/lo"
	"github.com/xssnick/tonutils-go/tlb"
)

var (
	MinmumTradeAmount = big.NewInt(1 * 1e9)
	MinGasCost        = tlb.MustFromTON("0.12")
)

var (
	ErrRiskPool             = errors.New("risk pool")
	ErrNotBuyTrade          = errors.New("not buy trade")
	ErrTradeAmountTooSmall  = errors.New("trade amount too small")
	ErrPairCanMakeTradePass = errors.New("pair can make trade pass")
	ErrNoPairProfitable     = errors.New("no pair profitable")
	ErrLimitTooHigh         = errors.New("limit too high")
	ErrLikelyABot           = errors.New("likely a bot")
	ErrDuplicated           = errors.New("duplicated")
)

type BundleChance struct {
	VictimTx         string `json:"victim_tx"`
	VictimAccountId  string `json:"victim_account_id"`
	PoolAddress      string `json:"pool_address"`
	VictimAmount     string `json:"victim_amount"`
	VictimLimit      string `json:"victim_limit"`
	VictimLimitRatio uint64 `json:"victim_limit_ratio"`

	LatestReserve0 string `json:"latest_reserve0"`
	LatestReserve1 string `json:"latest_reserve1"`
	LatestLt       uint64 `json:"latest_lt"`

	VictimJettonOut string `json:"victim_jetton_out"`
	BotIn           string `json:"bot_in"`
	BotOut          string `json:"bot_out"`
	Profit          string `json:"profit"`
}

func (b *BundleChance) Dump() string {
	buf := bytes.NewBuffer(nil)
	err := json.NewEncoder(buf).Encode(b)
	if err != nil {
		return ""
	}

	return buf.String()
}

// to check any trade bundleable or not
func (d *Detector) BuildBundleChance(pool *model.Pool, trade *model.Trade) (*BundleChance, error) {
	chance := &BundleChance{
		VictimTx:        trade.Hash,
		VictimAccountId: trade.Address,
		PoolAddress:     pool.Address,
		VictimAmount:    trade.Amount,
		VictimLimit:     trade.Limit,

		LatestReserve0: pool.Asset0Reserve,
		LatestReserve1: pool.Asset1Reserve,
		LatestLt:       pool.Lt,
	}

	h := hashOfChance(pool.Address, trade.Amount, trade.Limit)
	if _, found := d.chanceCache.Get(h); found {
		return nil, ErrDuplicated
	}
	d.chanceCache.Set(h, struct{}{}, cache.DefaultExpiration)

	// if pool is valid, should be white list
	if !poolInWhiteList(pool) {
		return nil, ErrRiskPool
	}

	if trade.TradeType != model.TradeTypeBuy {
		return nil, ErrNotBuyTrade
	}

	tonAmount := stringToBigInt(trade.Amount)
	// if amount is too small, ignore
	if tonAmount.Cmp(MinmumTradeAmount) < 0 {
		return nil, ErrTradeAmountTooSmall
	}

	x := stringToBigInt(trade.LatestReserve0)
	y := stringToBigInt(trade.LatestReserve1)
	limit := stringToBigInt(trade.Limit)

	pairs := make([]InOut, 0)
	initial := tlb.MustFromTON("2").Nano()
	terminator := tlb.MustFromTON("250").Nano()
	step := new(big.Int).Div(new(big.Int).Sub(terminator, initial), big.NewInt(200))

	model := NewModel(x, y, limit, tonAmount)
	log.Debug().Msgf("trade in: %s", model.TradeIn.String())
	log.Debug().Msgf("trade in without fee: %s", model.TradeInWithoutFee.String())
	log.Debug().Msgf("trade limit: %s", limit.String())
	log.Debug().Msgf("trade actual out should: %s", model.TradeActualOut().String())
	log.Debug().Msgf("limit actual out ratio should: %s", model.LimitActualOutRatio().String())

	chance.VictimJettonOut = model.TradeActualOut().String()
	chance.VictimLimitRatio = model.LimitActualOutRatio().Uint64()

	// if actual out is less than limit, trade not going to success ignore
	if model.TradeActualOut().Cmp(limit) < 0 {
		return nil, ErrLimitTooHigh
	}

	// if actual out is too close to actual out, might be an bot
	if model.LimitActualOutRatio().Cmp(big.NewInt(9900)) > 0 {
		return nil, ErrLikelyABot
	}

	for i := 0; i < 250; i++ {
		mI := new(big.Int).Add(initial, new(big.Int).Mul(step, big.NewInt(int64(i))))
		botJettonOut, tradeJettonOut, limitVsTradeJettonOut, botTonOut, tradeSuccess := model.IfBotBuyAmount(mI)

		log.Debug().Msgf("===================================")
		log.Debug().Msgf("botAmountIn: %s ", mI.String())
		log.Debug().Msgf("botJettonOut: %s", botJettonOut.String())
		log.Debug().Msgf("tradeJettonOut: %s", tradeJettonOut.String())
		log.Debug().Msgf("trade limit: %s", limit.String())
		log.Debug().Msgf("tradeSuccess: %t", tradeSuccess)
		log.Debug().Msgf("limitVsTradeJettonOut: %s", limitVsTradeJettonOut.String())
		log.Debug().Msgf("botTonOut: %s", botTonOut.String())
		log.Debug().Msgf("profit: %s", new(big.Int).Sub(botTonOut, mI).String())

		pairs = append(pairs, InOut{
			TradeSuccess:      tradeSuccess,
			TradeJettonAmount: tradeJettonOut,
			In:                mI,
			Out:               botTonOut,
			Profit:            new(big.Int).Sub(botTonOut, mI),
			Roi:               new(big.Int).Div(new(big.Int).Mul(new(big.Int).Sub(botTonOut, mI), big.NewInt(10000)), mI),
		})
	}

	// for _, pair := range pairs {
	// 	log.Debug().Msgf(pair.Dump())
	// }

	// do not affect actual trade
	pairCanMakeTradePass := lo.Filter(pairs, func(pair InOut, index int) bool {
		return pair.TradeSuccess
	})

	if len(pairCanMakeTradePass) == 0 {
		return nil, ErrPairCanMakeTradePass
	}

	// profitable
	pairsAreProfitable := lo.Filter(pairCanMakeTradePass, func(pair InOut, index int) bool {
		return pair.Profit.Cmp(big.NewInt(0)) > 0
	})

	if len(pairsAreProfitable) == 0 {
		return nil, ErrNoPairProfitable
	}

	pairsAreProfitableEnough := lo.Filter(pairsAreProfitable, func(pair InOut, index int) bool {
		return pair.Profit.Cmp(MinGasCost.Nano()) > 0
	})

	if len(pairsAreProfitableEnough) == 0 {
		return nil, ErrNoPairProfitable
	}

	maxProfitPair := MaxProfit(pairsAreProfitableEnough)

	chance.BotIn = maxProfitPair.In.String()
	chance.BotOut = maxProfitPair.Out.String()
	chance.Profit = maxProfitPair.Profit.String()

	return chance, nil
}

func poolInWhiteList(pool *model.Pool) bool {
	return true
}

func withFeeDeducted(amount *big.Int) *big.Int {
	// 0.25% fee
	fee := new(big.Int).Mul(amount, big.NewInt(25))
	fee = new(big.Int).Div(fee, big.NewInt(10000))

	return new(big.Int).Sub(amount, fee)
}

func stringToBigInt(s string) *big.Int {
	bi := new(big.Int)
	bi.SetString(s, 10)

	return bi
}

type BundleModel struct {
	X                 *big.Int
	Y                 *big.Int
	K                 *big.Int
	Limit             *big.Int
	TradeIn           *big.Int
	TradeInWithoutFee *big.Int
}

func NewModel(x, y, limit, tradeIn *big.Int) *BundleModel {
	model := &BundleModel{
		X:       x,
		Y:       y,
		Limit:   limit,
		TradeIn: tradeIn,
	}

	model.K = new(big.Int).Mul(x, y)
	model.TradeInWithoutFee = withFeeDeducted(tradeIn)
	return model
}

func (m *BundleModel) TradeActualOut() *big.Int {
	xHat := new(big.Int).Add(m.TradeInWithoutFee, m.X)
	yHat := new(big.Int).Div(m.K, xHat)
	return new(big.Int).Sub(m.Y, yHat)
}

func (m *BundleModel) LimitActualOutRatio() *big.Int {
	actualYOut := m.TradeActualOut()
	return new(big.Int).Div(new(big.Int).Mul(m.Limit, big.NewInt(10000)), actualYOut)
}

func (m *BundleModel) IfBotBuyAmount(amount *big.Int) (*big.Int, *big.Int, *big.Int, *big.Int, bool) {
	// bot buy
	botTonIn := withFeeDeducted(amount)
	xHat := new(big.Int).Add(botTonIn, m.X)
	yHat := new(big.Int).Div(m.K, xHat)
	botJettonOut := new(big.Int).Sub(m.Y, yHat)
	yAfterBotBuy := new(big.Int).Sub(m.Y, botJettonOut)
	xAfterBotBuy := new(big.Int).Add(m.X, botTonIn)

	// trade buy
	xHat = new(big.Int).Add(m.TradeInWithoutFee, xAfterBotBuy)
	yHat = new(big.Int).Div(m.K, xHat)
	tradeJettonOut := new(big.Int).Sub(yAfterBotBuy, yHat)
	limitVsTradeJettonOut := new(big.Int).Div(new(big.Int).Mul(m.Limit, big.NewInt(10000)), tradeJettonOut)
	xAfterTradeBuy := new(big.Int).Add(xAfterBotBuy, m.TradeInWithoutFee)
	yAfterTradeBuy := new(big.Int).Sub(yAfterBotBuy, tradeJettonOut)

	botSellJettonAfterFee := withFeeDeducted(botJettonOut)
	yHat = new(big.Int).Add(yAfterTradeBuy, botSellJettonAfterFee)
	xHat = new(big.Int).Div(m.K, yHat)
	botTonOut := new(big.Int).Sub(xAfterTradeBuy, xHat)

	tradeSuccess := tradeJettonOut.Cmp(m.Limit) >= 0

	return botJettonOut, tradeJettonOut, limitVsTradeJettonOut, botTonOut, tradeSuccess
}

type InOut struct {
	In                *big.Int
	Out               *big.Int
	TradeJettonAmount *big.Int
	TradeSuccess      bool
	Profit            *big.Int
	PossibleLoss      *big.Int
	Roi               *big.Int
}

func (io *InOut) Dump() string {
	var sb strings.Builder
	sb.WriteString("In: ")
	sb.WriteString(io.In.String())
	sb.WriteString(", Out: ")
	sb.WriteString(io.Out.String())
	sb.WriteString(", Profit: ")
	sb.WriteString(io.Profit.String())
	sb.WriteString(", PossibleLoss: ")
	sb.WriteString(io.PossibleLoss.String())
	sb.WriteString(", Roi: ")
	sb.WriteString(io.Roi.String())

	return sb.String()
}

func MaxProfit(pairs []InOut) *InOut {
	m := pairs[0]
	for _, pair := range pairs {
		if pair.Profit.Cmp(m.Profit) > 0 {
			m = pair
		}
	}

	return &m
}

func hashOfChance(args ...string) string {
	buf := bytes.NewBuffer(nil)
	for _, arg := range args {
		buf.WriteString(arg)
	}

	hash := md5.New()
	hash.Write(buf.Bytes())
	return hex.EncodeToString(hash.Sum(nil))
}
