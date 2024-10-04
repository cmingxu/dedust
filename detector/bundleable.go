package detector

import (
	"bytes"
	"encoding/json"
	"math/big"

	"github.com/cmingxu/dedust/model"
	"github.com/pkg/errors"
)

var (
	ErrRiskPool = errors.New("risk pool")
)

type BundleChance struct {
	VictimTx        string  `json:"victim_tx"`
	VictimAccountId string  `json:"victim_account_id"`
	PoolAddress     string  `json:"pool_address"`
	VictimAmount    float32 `json:"victim_amount"`
	VictimLimit     string  `json:"victim_limit"`
	LatestReserve0  string  `json:"latest_reserve0"`
	LatestReserve1  string  `json:"latest_reserve1"`
	LatestLt        uint64  `json:"latest_lt"`
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
func BuildBundleChance(pool *model.Pool, trade *model.Trade) (*BundleChance, error) {
	chance := &BundleChance{
		VictimTx:        trade.Hash,
		VictimAccountId: trade.Address,
		PoolAddress:     pool.Address,
		VictimAmount:    trade.Amount,
		VictimLimit:     trade.Limit,
	}

	// if pool is valid, should be white list
	if !poolInWhiteList(pool) {
		return nil, nil
	}

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
