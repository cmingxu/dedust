package detector

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/cmingxu/dedust/model"
	"github.com/patrickmn/go-cache"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/samber/lo"
	"github.com/xssnick/tonutils-go/tlb"
)

var (
	MinmumTradeAmount = big.NewInt(1 * 1e9)
	MinGasCost        = tlb.MustFromTON("0.1")
)

var (
	ErrRiskPool               = errors.New("risk pool")
	ErrNotBuyTrade            = errors.New("not buy trade")
	ErrTradeAmountTooSmall    = errors.New("trade amount too small")
	ErrPairCanMakeTradePass   = errors.New("pair can make trade pass")
	ErrNoPairProfitable       = errors.New("no pair profitable")
	ErrLimitTooHigh           = errors.New("limit too high")
	ErrLikelyABot             = errors.New("likely a bot")
	ErrDuplicated             = errors.New("duplicated")
	ErrNoPairAfterRangeFilter = errors.New("no pair after range filter")

	ErrRecentSellDetect = errors.New("recent sell detect")

	ErrRiskAsTradeIsBundle = errors.New("risk as trade is bundle")
)

func (d *Detector) p(format string, args ...interface{}) {
	fmt.Fprintf(d.out, format, args...)
}

// to check any trade bundleable or not
func (d *Detector) BuildBundleChance(pool *model.Pool, trade *model.Trade) (*model.BundleChance, error) {
	chance := &model.BundleChance{
		VictimTx:        trade.Hash,
		VictimAccountId: trade.Address,
		PoolAddress:     pool.Address,
		VictimAmount:    trade.Amount,
		VictimLimit:     trade.Limit,

		LatestReserve0: pool.Asset0Reserve,
		LatestReserve1: pool.Asset1Reserve,
		LatestLt:       pool.Lt,
		CreatedAt:      time.Now(),
	}

	d.p("=== %s ==== %s ====== \n", pool.Address, trade.Address)
	d.p("- now is: %s \n", time.Now().Format(time.RFC3339Nano))
	d.p("- VictimTx %s\n", trade.Hash)
	d.p("- VictimAccountId %s\n", trade.Address)
	d.p("- VictimAmount %s\n", trade.Amount)
	d.p("- VictimLimit %s\n", trade.Limit)
	d.p("- LatestReserve0 %s\n", pool.Asset0Reserve)
	d.p("- LatestReserve1 %s\n", pool.Asset1Reserve)

	if trade.HasNextStep {
		d.p("$ %s is bundle, skip\n", trade.Address)
		return nil, ErrRiskAsTradeIsBundle
	}

	h := hashOfChance(pool.Address, trade.Amount, trade.Limit)
	d.p("- Hash %s\n", h)
	if _, found := d.chanceCache.Get(h); found {
		d.p("$ %s duplicated, skip\n", trade.Address)
		return nil, ErrDuplicated
	}
	d.chanceCache.Set(h, struct{}{}, cache.DefaultExpiration)

	// if pool is valid, should be white list
	if !poolInWhiteList(pool) {
		d.p("$ %s risk pool, skip \n", pool.Address)
		return nil, ErrRiskPool
	}

	if trade.TradeType != model.TradeTypeBuy {
		d.p("$ %s not BUY trade, skip \n", trade.Address)
		return nil, ErrNotBuyTrade
	}

	if _, ok := d.sellingCache.Get(pool.Address); ok {
		d.p("$ %s selling signal in 50s, skip \n", trade.Address)
		return nil, ErrRecentSellDetect
	}

	tonAmount := stringToBigInt(trade.Amount)

	x := stringToBigInt(trade.LatestReserve0)
	y := stringToBigInt(trade.LatestReserve1)
	limit := stringToBigInt(trade.Limit)

	d.p("- X %s\n", x.String())
	d.p("- Y %s\n", y.String())
	d.p("- Limit %s\n", limit.String())

	pairs := make([]InOut, 0)
	initial := tlb.MustFromTON("2").Nano()
	terminator := tlb.MustFromTON("200").Nano()
	step := new(big.Int).Div(new(big.Int).Sub(terminator, initial), big.NewInt(200))

	model := NewModel(x, y, limit, tonAmount)
	log.Debug().Msgf("trade in: %s", model.TradeIn.String())
	log.Debug().Msgf("trade in without fee: %s", model.TradeInWithoutFee.String())
	log.Debug().Msgf("trade limit: %s", limit.String())
	log.Debug().Msgf("trade actual out should: %s", model.TradeActualOut().String())
	log.Debug().Msgf("limit actual out ratio should: %s", model.LimitActualOutRatio().String())

	d.p("- TradeIn %s\n", model.TradeIn.String())
	d.p("- TradeInWithoutFee %s\n", model.TradeInWithoutFee.String())
	d.p("- TradeActualOut %s\n", model.TradeActualOut().String())
	d.p("- LimitActualOutRatio %s\n", model.LimitActualOutRatio().String())

	chance.VictimJettonOut = model.TradeActualOut().String()
	chance.VictimLimitRatio = model.LimitActualOutRatio().Uint64()

	// if actual out is less than limit, trade not going to success ignore
	if model.TradeActualOut().Cmp(limit) < 0 {
		d.p("$ %s limit too high, skip \n", trade.Address)
		return nil, ErrLimitTooHigh
	}

	// if actual out is too close to actual out, might be an bot
	if model.LimitActualOutRatio().Cmp(big.NewInt(9990)) > 0 {
		return nil, ErrLikelyABot
	}

	for i := 0; i < 200; i++ {
		mI := new(big.Int).Add(initial, new(big.Int).Mul(step, big.NewInt(int64(i))))
		botJettonOut, tradeJettonOut, _, botTonOut, tradeSuccess := model.IfBotBuyAmount(mI)

		// log.Debug().Msgf("===================================")
		// log.Debug().Msgf("botAmountIn: %s ", mI.String())
		// log.Debug().Msgf("botJettonOut: %s", botJettonOut.String())
		// log.Debug().Msgf("tradeJettonOut: %s", tradeJettonOut.String())
		// log.Debug().Msgf("trade limit: %s", limit.String())
		// log.Debug().Msgf("tradeSuccess: %t", tradeSuccess)
		// log.Debug().Msgf("limitVsTradeJettonOut: %s", limitVsTradeJettonOut.String())
		// log.Debug().Msgf("botTonOut: %s", botTonOut.String())
		// log.Debug().Msgf("profit: %s", new(big.Int).Sub(botTonOut, mI).String())

		// d.p("*******************\n")
		// d.p("botAmount: %s\n", mI.String())
		// d.p("botJettonOut: %s\n", botJettonOut.String())
		// d.p("tradeJettonOut: %s\n", tradeJettonOut.String())
		// d.p("trade limit: %s\n", limit.String())
		// d.p("tradeSuccess: %t\n", tradeSuccess)
		// d.p("botTonOut: %s\n", botTonOut.String())
		// d.p("profit: %s\n", new(big.Int).Sub(botTonOut, mI).String())

		pairs = append(pairs, InOut{
			TradeSuccess:      tradeSuccess,
			TradeJettonAmount: tradeJettonOut,
			BotIn:             mI,
			BotTonOut:         botTonOut,
			BotJettonOut:      botJettonOut,
			Profit:            new(big.Int).Sub(botTonOut, mI),
			Roi:               new(big.Int).Div(new(big.Int).Mul(new(big.Int).Sub(botTonOut, mI), big.NewInt(10000)), mI),
		})
	}

	// for _, pair := range pairs {
	// 	log.Debug().Msgf(pair.Dump())
	// }

	// 不能影响真正的买家购买成功
	pairCanMakeTradePass := lo.Filter(pairs, func(pair InOut, index int) bool {
		return pair.TradeSuccess
	})

	d.p("$ pairCanMakeTradePass: %d\n", len(pairCanMakeTradePass))

	if len(pairCanMakeTradePass) == 0 {
		return nil, ErrPairCanMakeTradePass
	}

	// 得赚钱， 不赚钱的不要
	pairsAreProfitable := lo.Filter(pairCanMakeTradePass, func(pair InOut, index int) bool {
		return pair.Profit.Cmp(big.NewInt(0)) > 0
	})

	d.p("$ pairsAreProfitable: %d\n", len(pairsAreProfitable))

	if len(pairsAreProfitable) == 0 {
		return nil, ErrNoPairProfitable
	}

	// 赚钱要够多，要大于整个 Gas 手续费
	pairsAreProfitableEnough := lo.Filter(pairsAreProfitable, func(pair InOut, index int) bool {
		return pair.Profit.Cmp(MinGasCost.Nano()) > 0
	})

	d.p("$ pairsAreProfitableEnough(more then gas): %d\n", len(pairsAreProfitableEnough))

	if len(pairsAreProfitableEnough) == 0 {
		return nil, ErrNoPairProfitable
	}

	// 要和 reserve0（TON） 的数量比较，成比例
	pairsAreProfitableEnough = lo.Filter(pairsAreProfitableEnough, func(pair InOut, index int) bool {
		botInVsReserve0Ratio := new(big.Int).Div(new(big.Int).Mul(pair.BotIn, big.NewInt(10000)), x)
		// 如果 botIn 小于 2 TON，可以试试， 跟池子 TON 大小无关
		if pair.BotIn.Cmp(tlb.MustFromTON("3").Nano()) < 0 {
			return true
		}

		if pair.BotIn.Cmp(tlb.MustFromTON("5").Nano()) < 0 && botInVsReserve0Ratio.Cmp(big.NewInt(5000)) < 0 {
			return true
		}

		if pair.BotIn.Cmp(tlb.MustFromTON("10").Nano()) < 0 && botInVsReserve0Ratio.Cmp(big.NewInt(2000)) < 0 {
			return true
		}

		// 20  30%
		if pair.BotIn.Cmp(tlb.MustFromTON("20").Nano()) < 0 && botInVsReserve0Ratio.Cmp(big.NewInt(1500)) < 0 {
			return true
		}

		// 如果 botIn 大于 20 TON， botInVsReserve0Ratio 小于 20%可以试试
		if pair.BotIn.Cmp(tlb.MustFromTON("50").Nano()) < 0 && botInVsReserve0Ratio.Cmp(big.NewInt(1000)) < 0 {
			return true
		}

		if pair.BotIn.Cmp(tlb.MustFromTON("100").Nano()) < 0 && botInVsReserve0Ratio.Cmp(big.NewInt(500)) < 0 {
			return true
		}

		// 如果 botIn 大于 100 TON， botInVsReserve0Ratio 小于 5%可以试试
		if botInVsReserve0Ratio.Cmp(big.NewInt(500)) < 0 {
			return true
		}

		return false
	})

	d.p("$ pairsAreProfitableEnough(after range filter): %d\n", len(pairsAreProfitableEnough))

	for _, pair := range pairsAreProfitableEnough {
		log.Debug().Msgf(pair.Dump())
		d.p("***  %s\n", pair.Dump())
	}

	if len(pairsAreProfitableEnough) == 0 {
		return nil, ErrNoPairAfterRangeFilter
	}

	// 选出最大的那个
	maxProfitPair := MaxProfit(pairsAreProfitableEnough)
	log.Debug().Msgf("max profit pair: %s", maxProfitPair.Dump())

	d.p(" ***************** ")
	d.p("max profit pair: %s\n", maxProfitPair.Dump())
	d.p(" ***************** ")

	chance.BotIn = maxProfitPair.BotIn.String()
	chance.BotTonOut = maxProfitPair.BotTonOut.String()
	chance.BotJettonOut = maxProfitPair.BotJettonOut.String()
	chance.Profit = maxProfitPair.Profit.String()
	chance.Roi = maxProfitPair.Roi.String()

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
	BotIn             *big.Int
	BotTonOut         *big.Int
	BotJettonOut      *big.Int
	TradeJettonAmount *big.Int
	TradeSuccess      bool
	Profit            *big.Int
	PossibleLoss      *big.Int
	Roi               *big.Int
}

func (io *InOut) Dump() string {
	var sb strings.Builder
	sb.WriteString("In: ")
	sb.WriteString(io.BotIn.String())
	sb.WriteString(", Ton Out: ")
	sb.WriteString(io.BotTonOut.String())
	sb.WriteString(", Jetton Out: ")
	sb.WriteString(io.BotJettonOut.String())
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
