package main

import (
	"fmt"

	"github.com/xssnick/tonutils-go/address"
)

func main3() {
	addr := address.MustParseAddr("EQDQ1I0oC04dvmbewqUkt1n7zzxijS2hXmIv0FMv_JJZjQNI")

	fmt.Println("Address:", addr.String())
	fmt.Println("Workchain:", addr.Workchain())
	fmt.Println("Bounceable:", addr.IsBounceable())
	fmt.Println("IsTestonly", addr.IsTestnetOnly())
}
