package wallet

import (
	"github.com/xssnick/tonutils-go/tlb"
	"github.com/xssnick/tonutils-go/tvm/cell"
)

// https://github.com/xssnick/tonutils-go/blob/master/ton/wallet/v5r1.go
type Action struct {
	_      tlb.Magic  `tlb:"#0ec3c86d"`
	Mode   uint64     `tlb:"## 8"`
	Empty  *cell.Cell `tlb:"^"`
	OutMsg *cell.Cell `tlb:"^"`
}

type V5R1Header struct {
	_          tlb.Magic `tlb:"#7369676e"`
	WalletId   uint64    `tlb:"## 32"`
	ExpireAt   uint64    `tlb:"## 32"`
	Seqno      uint64    `tlb:"## 32"`
	ActionPref uint8     `tlb:"## 1"`
	Action     *Action   `tlb:"^"`
}
