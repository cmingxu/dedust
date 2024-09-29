package detector

import (
	"testing"
	"encoding/hex"

	"github.com/pkg/errors"
	"github.com/xssnick/tonutils-go/tvm/cell"
)

func TestDecodeDedustNativeSwap(t *testing.T) {
	NativeSwapBOC := "b5ee9c7241010201004000016bea06185df4ce7117000000004453fdf518007cbff951bbf9e6d86d93fe8de62ac5556a37322908b5ad84d97bcc93ab14ab103627409401000966f4cd3802faedb3f6"


	boc, err := hex.DecodeString(NativeSwapBOC)
	if err != nil {
		t.Fatal(errors.Wrap(err, "decode"))
	}


	cell, err := cell.FromBOC(boc)
	if err != nil {
		t.Fatal(errors.Wrap(err, "from boc"))
	}

	println(cell.Dump())

	nativeSwap, err := decodeDedustNativeSwap(cell)
	if err != nil {
		t.Fatal(err)
	}

	println(nativeSwap.QueryId)

}

func TestDecodeJettonTransfer(t *testing.T) {
	SellSwapBOC := "b5ee9c724101040100980001b00f8a7ea5540cf95ed20fe01845a2687e48003fc026511547bfa29b001d06c42356a6f4d161ba8bc96f23fba34852f24505b50007b0d978d5908c660a5e87c4197117c80e164e22d39d997f162a95afd0e0625e4811e1a301010155e3a0d482801b9369d232d6f5c65d5cdf65149eacd97d48475af60f7d9d5a488bfb31d01302c4385ec18f4002020966f4dd1d0e030300082e1cfa82455b33cf"
	boc, err := hex.DecodeString(SellSwapBOC)
	if err != nil {
		t.Fatal(errors.Wrap(err, "decode"))
	}

	cell, err := cell.FromBOC(boc)
	if err != nil {
		t.Fatal(errors.Wrap(err, "from boc"))
	}

	println(cell.Dump())

	jettonTransfer, err := decodeJettonTransfer(cell)
	if err != nil {
		t.Fatal(err)
	}

	println(jettonTransfer.QueryId)
}
