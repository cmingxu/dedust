package main

import (
	"encoding/hex"
	"fmt"
	"math/big"
)

func main4() {
	h1 := bigIntFromString("c553a7ee17190d9205ddaf6faa9cafd5cc0ec91faa9b6d1ce05b094664cb9291")
	h2 := bigIntFromString("f960d72253a2f1352c002ee9d3d3ce81618ac060156fa7eae1b05ba6c5b14f12")
	h3 := bigIntFromString("46335ea090937c0ff34822fd2708934c34448a05b818e7d67b7fb4808813a2d9")
	h4 := bigIntFromString("39544a2a895112109f4a56c799423ddeccab32d4b598f79894f1a514a6755b75")
	h5 := bigIntFromString("edd31d0dbd4cf9ec66cfcffefa99e4de6e54edf37dd133bd4498555d6f13421c")

	fmt.Println("h1", h1.String())

	fmt.Println(h1.Cmp(h2))
	fmt.Println(h1.Cmp(h3))
	fmt.Println(h1.Cmp(h4))
	fmt.Println(h1.Cmp(h5))
}

func bigIntFromString(s string) *big.Int {
	b, err := hex.DecodeString(s)
	if err != nil {
		panic(err)
	}

	return new(big.Int).SetBytes(b)
}
