package wallet

import (
	"bytes"
	"encoding/hex"
	"testing"

	"github.com/cmingxu/dedust/utils"
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/tvm/cell"
)

func TestCalculateUserJettonWalletAddress(t *testing.T) {
	// vault
	ownerAddress := address.MustParseAddr("EQADhEildWcgRr6VdFpSi8Izn0BYmlIB5zYOr5y7768EdWke")
	// jetton master
	jettonMasterAddress := address.MustParseAddr("EQCIiZ6b5fYSHLfCCAqi4KsIeqQAD3H_HSP-V5r4OUgTEuAs")

	// jetton wallet code
	jettonWalletCodeBoc := "b5ee9c7201021201000334000114ff00f4a413f4bcf2c80b0102016202030202cc0405001ba0f605da89a1f401f481f481a8610201d40607020148080900c30831c02497c138007434c0c05c6c2544d7c0fc02f83e903e900c7e800c5c75c87e800c7e800c1cea6d0000b4c7e08403e29fa954882ea54c4d167c0238208405e3514654882ea58c511100fc02780d60841657c1ef2ea4d67c02b817c12103fcbc2000113e910c1c2ebcb853600201200a0b020120101101f100f4cffe803e90087c007b51343e803e903e90350c144da8548ab1c17cb8b04a30bffcb8b0950d109c150804d50500f214013e809633c58073c5b33248b232c044bd003d0032c032483e401c1d3232c0b281f2fff274013e903d010c7e800835d270803cb8b11de0063232c1540233c59c3e8085f2dac4f3200c03f73b51343e803e903e90350c0234cffe80145468017e903e9014d6f1c1551cdb5c150804d50500f214013e809633c58073c5b33248b232c044bd003d0032c0327e401c1d3232c0b281f2fff274140371c1472c7cb8b0c2be80146a2860822625a020822625a004ad8228608239387028062849f8c3c975c2c070c008e00d0e0f00ae8210178d4519c8cb1f19cb3f5007fa0222cf165006cf1625fa025003cf16c95005cc2391729171e25008a813a08208e4e1c0aa008208989680a0a014bcf2e2c504c98040fb001023c85004fa0258cf1601cf16ccc9ed5400705279a018a182107362d09cc8cb1f5230cb3f58fa025007cf165007cf16c9718010c8cb0524cf165006fa0215cb6a14ccc971fb0010241023000e10491038375f040076c200b08e218210d53276db708010c8cb055008cf165004fa0216cb6a12cb1f12cb3fc972fb0093356c21e203c85004fa0258cf1601cf16ccc9ed5400db3b51343e803e903e90350c01f4cffe803e900c145468549271c17cb8b049f0bffcb8b0a0823938702a8005a805af3cb8b0e0841ef765f7b232c7c572cfd400fe8088b3c58073c5b25c60063232c14933c59c3e80b2dab33260103ec01004f214013e809633c58073c5b3327b55200083200835c87b51343e803e903e90350c0134c7e08405e3514654882ea0841ef765f784ee84ac7cb8b174cfcc7e800c04e81408f214013e809633c58073c5b3327b5520"

	jettonWalletAddress := address.MustParseAddr("EQBnZ_RxxkX-1wHTD5wQnFPY6JCX9za0lBxwevWIaZfC8tkM")

	decoded, err := hex.DecodeString(jettonWalletCodeBoc)
	if err != nil {
		t.Error(err)
	}

	jettonWalletCode, err := cell.FromBOC(decoded)
	if err != nil {
		t.Error(err)
	}

	stateInit := CalculateJettonWalletStateInit(ownerAddress, jettonMasterAddress, jettonWalletCode)
	jettonWalletAddrCell := CalculateUserJettonWalletAddress(ownerAddress, jettonMasterAddress, jettonWalletCode)

	println("===========")
	println("vault : ", ownerAddress.String())
	println("jetton master: ", jettonMasterAddress.String())
	println("jetton wallet: ", jettonWalletAddress.String())
	println("generated wallet addr:", utils.CellToAddress(jettonWalletAddrCell).String())

	//println("jetton wallet cell: ", jettonWalletAddrCell.Dump())
	//println("state init cell: ", stateInit.Dump())
	t.Log(hex.EncodeToString(stateInit.Hash()))
	t.Log(len(stateInit.Hash()))

	if bytes.Equal(utils.CellToAddress(jettonWalletAddrCell).Data(),
		jettonWalletAddress.Data()) {
		t.Error("no equal")
	}
}
