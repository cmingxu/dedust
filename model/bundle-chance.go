package model

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"
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
	IOR0            string `json:"ior0"` // BotInOverReserve0

	CreatedAt time.Time `json:"created_at"`
	FirstSeen time.Time `json:"first_seen"`
}

func (b *BundleChance) Dump() string {
	buf := bytes.NewBuffer(nil)
	err := json.NewEncoder(buf).Encode(b)
	if err != nil {
		return ""
	}

	return buf.String()
}

func (b *BundleChance) ShortDump() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("| VictimAccountId: %s", b.VictimAccountId))
	sb.WriteString(fmt.Sprintf("| PoolAddress: %s", b.PoolAddress))
	sb.WriteString(fmt.Sprintf("| VictimAmount: %s", b.VictimAmount))
	sb.WriteString(fmt.Sprintf("| VictimLimit: %s", b.VictimLimit))
	sb.WriteString(fmt.Sprintf("| VictimLimitRatio: %d", b.VictimLimitRatio))
	sb.WriteString(fmt.Sprintf("| LatestReserve0: %s", b.LatestReserve0))
	sb.WriteString(fmt.Sprintf("| LatestReserve1: %s", b.LatestReserve1))
	sb.WriteString(fmt.Sprintf("| LatestLt: %d", b.LatestLt))
	sb.WriteString(fmt.Sprintf("| VictimJettonOut: %s", b.VictimJettonOut))
	sb.WriteString(fmt.Sprintf("| BotIn: %s", b.BotIn))
	sb.WriteString(fmt.Sprintf("| BotJettonOut: %s", b.BotJettonOut))
	sb.WriteString(fmt.Sprintf("| BotTonOut: %s", b.BotTonOut))
	sb.WriteString(fmt.Sprintf("| Profit: %s", b.Profit))
	sb.WriteString(fmt.Sprintf("| Roi: %s", b.Roi))
	sb.WriteString(fmt.Sprintf("| IOR0: %s", b.IOR0))
	sb.WriteString(fmt.Sprintf("| CreatedAt: %s", b.CreatedAt.Format(time.RFC3339Nano)))
	sb.WriteString(fmt.Sprintf("| FirstSeen: %s", b.FirstSeen.Format(time.RFC3339Nano)))
	sb.WriteString(fmt.Sprintf("| Now: %s", time.Now().Format(time.RFC3339Nano)))

	return sb.String()
}

func (b *BundleChance) CSV(w io.Writer) error {
	_, err := fmt.Fprintf(w, "%s,%s,%s,%s,%s,%d,%s,%s,%s,%d,%s,%s,%s,%s,%s,%s, %s, %s, %s\n",
		b.VictimTx, b.VictimAccountId, b.PoolAddress, b.VictimAmount, b.VictimLimit, b.VictimLimitRatio,
		b.LatestReserve0, b.LatestReserve1, b.LatestLt,
		b.VictimJettonOut, b.BotIn, b.BotJettonOut, b.BotTonOut, b.Profit, b.Roi, b.IOR0, b.CreatedAt.Format(time.RFC3339Nano), b.FirstSeen.Format(time.RFC3339Nano), time.Now().Format(time.RFC3339Nano),
	)
	return err
}
