package main

import "math/big"

func main1() {
	botIn := big.NewInt(200)
	x := big.NewInt(1001)

	println(new(big.Int).Div(x, botIn).Cmp(big.NewInt(10)) >= 0)
}
