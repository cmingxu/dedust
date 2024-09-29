package detector

import (
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/tlb"
	"github.com/xssnick/tonutils-go/tvm/cell"
)

// https://docs.dedust.io/reference/tlb-schemes
const DedustNativeSwap = 0xea06185d
const DedustJettonSwap = 0xe3a0d482
const JettonTransfer = 0xf8a7ea5
const JettonBurn = 0x595f07bc
const DedustSwapExternal = 0x61ee542d

type SwapParams struct {
	Deadline        uint64           `tlb:"## 32"`
	Recipient       *address.Address `tlb:"addr"`
	Referrer        *address.Address `tlb:"addr"`
	FullfillPayload *cell.Cell       `tlb:"maybe ^"`
	RejectPayload   *cell.Cell       `tlb:"maybe ^"`
}

type SwapStepParams struct {
	SwapKind uint8     `tlb:"## 1"`
	Limit    tlb.Coins `tlb:"."`
	Next     *SwapStep `tlb:"maybe ^"`
}

type SwapStep struct {
	PoolAddr       *address.Address `tlb:"addr"`
	SwapStepParams *SwapStepParams  `tlb:"."`
}

type NativeSwap struct {
	_          tlb.Magic   `tlb:"#ea06185d"`
	QueryId    uint64      `tlb:"## 64"`
	Amount     tlb.Coins   `tlb:"."`
	SwapStep   *SwapStep   `tlb:"."`
	SwapParams *SwapParams `tlb:"^"`
}

type JettonSwap struct {
	_          tlb.Magic   `tlb:"#e3a0d482"`
	SwapStep   *SwapStep   `tlb:"."`
	SwapParams *SwapParams `tlb:"^"`
}

// https://github.com/ton-blockchain/token-contract/blob/2c13d3ef61ca4288293ad65bf0cfeaed83879b93/ft/jetton-wallet.fc#L55'
type JettonTransferParams struct {
	IsRight            uint8            `tlb:"## 4"`
	_                  tlb.Magic        `tlb:"#f8a7ea5"`
	QueryId            uint64           `tlb:"## 64"`
	Amount             tlb.Coins        `tlb:"."`
	Destination        *address.Address `tlb:"addr"`
	ResponseDestinaton *address.Address `tlb:"addr"`
	CustomPayload      *cell.Cell       `tlb:"maybe ^"`
	ForwardTonAmount   tlb.Coins        `tlb:"."`
	ForwardPayload     *JettonSwap      `tlb:"either . ^"`
}
