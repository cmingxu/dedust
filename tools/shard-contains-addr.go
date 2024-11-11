package main

import (
	"encoding/binary"

	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/tlb"
)

func main() {
	shard := tlb.ShardID(0xe000000000000000)
	addr := address.MustParseAddr("EQDYq9DNZfs3ovNMWnaGO0Y-IUq-6aRsNGX_uYckE200mKh9")

	if shard.ContainsAddress(addr) {
		println("Address is in shard")
	} else {
		println("Address is not in shard")
	}

	println(binary.BigEndian.Uint64(addr.Data()))

	print("origin: ")
	printBinary(addr.Data())

	print("negtive: ")
	i := binary.BigEndian.Uint64(addr.Data())
	bytes := make([]byte, 8)
	binary.BigEndian.PutUint64(bytes, bitNegtive64(i))
	printBinary(bytes)

	print("lower: ")
	binary.BigEndian.PutUint64(bytes, lowerBit64(i))
	printBinary(bytes)

	print("shard: ")
	binary.BigEndian.PutUint64(bytes, uint64(shard))
	printBinary(bytes)

	print("parent: ")
	binary.BigEndian.PutUint64(bytes, uint64(shard.GetParent()))
	printBinary(bytes)

	print("child left:")
	binary.BigEndian.PutUint64(bytes, uint64(shard.GetChild(true)))
	printBinary(bytes)

	print("child right:")
	binary.BigEndian.PutUint64(bytes, uint64(shard.GetChild(false)))
	printBinary(bytes)
}

func bitNegtive64(i uint64) uint64 {
	return ^i + 1
}

func lowerBit64(i uint64) uint64 {
	return i & bitNegtive64(i)
}

func printBinary(b []byte) {
	for i := 0; i < len(b); i++ {
		for j := 0; j < 8; j++ {
			if b[i]&(1<<uint(j)) != 0 {
				print("1")
			} else {
				print("0")
			}
		}
		print(" ")
	}
	println()
}
