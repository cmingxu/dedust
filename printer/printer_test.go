package printer

import (
	"context"
	"crypto/ed25519"
	"encoding/hex"
	"fmt"
	"testing"
	"time"

	"github.com/cmingxu/dedust/bot"
	"github.com/cmingxu/dedust/model"
	"github.com/cmingxu/dedust/utils"
	"github.com/cmingxu/dedust/wallet"
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/tlb"
	"github.com/xssnick/tonutils-go/tvm/cell"
)

func TestGGeneration(t *testing.T) {
	poolAddr := address.MustParseAddr("EQD4KR7TNQeNGfDZw3WF0WndkGNsNVaus8tk4jC6WQep2jpz")
	asset1JettonMasterAddr := address.MustParseAddr("EQCUTVCHyfnc46GF3EhItU-wjH7APK95hGjxvvGNfBC2Rf2_")
	asset1Vault := address.MustParseAddr("EQDZ_trLm3dpmwZp8rgy9riCT0cnzJZpxcMuD9upS76bAG8d")
	asset1JettonWalletCodeHash := "p2DWKdU0PnbQRQF9ncIW/IoweoN3gV/rKwpcSQ5zNIY="

	code, _ := model.WalletCodeBOCs[asset1JettonWalletCodeHash]
	content, _ := hex.DecodeString(code)
	codeCell, _ := cell.FromBOC(content)
	ownerAddr := asset1Vault

	jettonWalletCell := wallet.CalculateUserJettonWalletAddress(
		ownerAddr,
		asset1JettonMasterAddr,
		codeCell,
	)

	println("asset1 wallet addr", wallet.CellToAddress(jettonWalletCell).String())
	ctx := context.Background()

	connPool, ctx, _ := utils.GetConnectionPool("/tmp/global-config.json")
	client := utils.GetAPIClientWithTimeout(connPool, time.Second*10)

	botprivateKey := ed25519.PrivateKey{0x1, 0x2, 0x3, 0x4}
	botWallet := bot.NewBotWallet(ctx, client, poolAddr, botprivateKey, 1)

	fmt.Println("begin: ", time.Now())
	for {
		_, pk, _ := ed25519.GenerateKey(nil)
		_, gAddr, _ := botWallet.BuildG(poolAddr, pk, tlb.MustFromTON("0.3"))

		// println("gAddr", gAddr.String())
		gJettonWalletCell := wallet.CalculateUserJettonWalletAddress(
			gAddr,
			asset1JettonMasterAddr,
			codeCell,
		)
		// println("gJettonWalletCell", wallet.CellToAddress(gJettonWalletCell).String())

		if match(
			wallet.CellToAddress(jettonWalletCell).Data(),
			wallet.CellToAddress(gJettonWalletCell).Data(),
			gAddr.Data()) {
			break
		}
	}
	fmt.Println("end: ", time.Now())

	t.Error("jettonWalletCell", jettonWalletCell.Dump())
}

func match(a, b, c []byte) bool {
	fa := a[0]
	fb := b[0]
	fc := c[0]

	if fa == fb && fb == fc {
		return true
	}
	return false
}
