package model

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"time"
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
	BotJettonOut    string `json:"bot_jetton_out"`
	BotTonOut       string `json:"bot_ton_out"`
	Profit          string `json:"profit"`
	Roi             string `json:"roi"`

	CreatedAt time.Time `json:"created_at"`
}

func (b *BundleChance) Dump() string {
	buf := bytes.NewBuffer(nil)
	err := json.NewEncoder(buf).Encode(b)
	if err != nil {
		return ""
	}

	return buf.String()
}

func (b *BundleChance) CSV(w io.Writer) error {
	_, err := fmt.Fprintf(w, "%s,%s,%s,%s,%s,%d,%s,%s,%s,%d,%s,%s,%s,%s,%s,%s, %s\n",
		b.VictimTx, b.VictimAccountId, b.PoolAddress, b.VictimAmount, b.VictimLimit, b.VictimLimitRatio,
		b.LatestReserve0, b.LatestReserve1, b.LatestLt,
		b.VictimJettonOut, b.BotIn, b.BotJettonOut, b.BotTonOut, b.Profit, b.Roi, b.CreatedAt.Format(time.RFC3339),
	)
	return err
}
