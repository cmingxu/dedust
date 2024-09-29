package wallet

import "github.com/xssnick/tonutils-go/tvm/cell"

type V3Header struct {
	Sig       []byte     `tlb:"bits 512"`
	Subwallet uint64     `tlb:"## 32"`
	Expire    uint64     `tlb:"## 32"`
	Seqno     uint64     `tlb:"## 32"`
	Mode      uint32     `tlb:"## 8"`
	Body      *cell.Cell `tlb:"^"`
}
