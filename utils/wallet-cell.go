package utils

import (
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/tvm/cell"
)

func CellToAddress(c *cell.Cell) *address.Address {
	builder := c.BeginParse()
	flags := builder.MustLoadUInt(3)
	workchain := builder.MustLoadUInt(8)
	addrSlice := builder.MustLoadSlice(256)

	return address.NewAddress(
		byte(flags),
		byte(workchain),
		addrSlice)
}

func AddressToCell(addr *address.Address) *cell.Cell {
	return cell.BeginCell().
		MustStoreUInt(4, 3).
		MustStoreUInt(0, 8). // workchain
		MustStoreSlice(addr.Data(), 256).
		EndCell()
}
