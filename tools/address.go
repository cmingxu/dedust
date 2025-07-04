package main

import (
	"fmt"
	"math/big"

	"github.com/xssnick/tonutils-go/address"
)

func main3() {
	addr := address.MustParseAddr("EQDQ1I0oC04dvmbewqUkt1n7zzxijS2hXmIv0FMv_JJZjQNI")

	fmt.Println("Address:", addr.String())
	fmt.Println("Workchain:", addr.Workchain())
	fmt.Println("Bounceable:", addr.IsBounceable())
	fmt.Println("IsTestonly", addr.IsTestnetOnly())

	fmt.Println(len(addr.Data()))

	bi := new(big.Int).SetBytes(addr.Data())
	fmt.Println("Big int:", bi)

	m := address.MustParseAddr("EQDa8hg1R1vD8aN_u1-AuvLz5T2SNntC2dhOK8ne0g3_g3t-")
	bim := new(big.Int).SetBytes(m.Data())

	d4i := address.MustParseAddr("EQDapaPu3mdjy0pKA7LY931i76lexBFAcctYY-Pez5i9kd4i")
	d4im := new(big.Int).SetBytes(d4i.Data())

	fmt.Println(bi.Cmp(bim))
	fmt.Println(bi.Cmp(d4im))

	xwA := address.MustParseAddr("EQDYINWQNq8dSrqkAIgjc73rkGlDr9o-qJb6mu0xYXPk4xwA")
	xwAm := new(big.Int).SetBytes(xwA.Data())

	fmt.Println(xwAm.Cmp(bim))
	fmt.Println(xwAm.Cmp(d4im))

	abb := address.MustParseAddr("EQDvDz84DtQ7fwi13axZ3FHpCeHyGCFuo_bt3nrvGwcOAABT")
	abbi := new(big.Int).SetBytes(abb.Data())

	// fmt.Println(abbi.Cmp(bi))
	// fmt.Println(abbi.Cmp(bim))
	// fmt.Println(abbi.Cmp(xwAm))

	n := address.MustParseAddr("EQDQsm_hYZKJ0p1C0d3PMZtzIWfVveqXpr2PcinpBe5PXIlt")
	nm := new(big.Int).SetBytes(n.Data())
	fmt.Println("n:", n.String())
	fmt.Println(nm.Cmp(bi))
	fmt.Println(nm.Cmp(bim))
	fmt.Println(nm.Cmp(d4im))
	fmt.Println(nm.Cmp(xwAm))
	fmt.Println(nm.Cmp(abbi))

	// d5l := address.MustParseAddr("EQDwMhZgWS6IceQ_upIWtMMUHZ3-4qlu9dNvcgpeRW9Ihd5l")
	// d5li := new(big.Int).SetBytes(d5l.Data())

	tgn := address.MustParseAddr("EQDh9HoI_XrAYeuvUZLrlH2BGfYKvF4Jz_z67zkzG7_7uTgN")
	tgni := new(big.Int).SetBytes(tgn.Data())
	fmt.Println(nm.Cmp(tgni))

	zoz := address.MustParseAddr("UQDQ7jqqGUsLNDYwTTHo-E14ehHBPv1oVIw3Jam7_7SZBZoX")
	zozi := new(big.Int).SetBytes(zoz.Data())

	fmt.Println(nm.Cmp(zozi))

	oaju := address.MustParseAddr("EQDU42903hp_1pVhXxnjIjnbm80jToPqkxaMC_sAzph-oajU")
	oajui := new(big.Int).SetBytes(oaju.Data())

	fmt.Println(nm.Cmp(oajui))

	fmt.Println(oajui.Cmp(bi))

}
