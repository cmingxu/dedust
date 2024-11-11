package wallet

import (
	"github.com/xssnick/tonutils-go/tvm/cell"
)

type V4R2Header struct {
	Sig       []byte     `tlb:"bits 512"`
	Subwallet uint64     `tlb:"## 32"`
	Expire    uint64     `tlb:"## 32"`
	Seqno     uint64     `tlb:"## 32"`
	Op        uint32     `tlb:"## 8"`
	Mode      uint32     `tlb:"## 8"`
	Body      *cell.Cell `tlb:"^"`
}
