package bot

import (
	"crypto/ed25519"
	"fmt"

	"github.com/cmingxu/dedust/utils"
	mywallet "github.com/cmingxu/dedust/wallet"
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/tlb"
	"github.com/xssnick/tonutils-go/ton/wallet"
	"github.com/xssnick/tonutils-go/tvm/cell"
)

// https://github.com/dedust-io/sdk/blob/main/src/contracts/dex/vault/VaultNative.ts
func BuildG(botAddr *address.Address,
	gPrivateKey ed25519.PrivateKey,
	amount tlb.Coins,
) (*wallet.Message, *address.Address, error) {

	body, _ := createCommentCell("G")

	contractCode := getGCode()
	contractData := getGData(gPrivateKey.Public().(ed25519.PublicKey), botAddr)
	state := &tlb.StateInit{
		Data: contractData,
		Code: contractCode,
	}

	stateCell, err := tlb.ToCell(state)
	if err != nil {
		return nil, nil, err
	}

	addr := address.NewAddress(0, 0, stateCell.Hash())

	message := wallet.Message{
		Mode: wallet.PayGasSeparately + wallet.IgnoreErrors,
		InternalMessage: &tlb.InternalMessage{
			IHRDisabled: true,
			Bounce:      false,
			DstAddr:     addr,
			Amount:      amount,
			Body:        body,
			StateInit:   state,
		},
	}

	return &message, addr, nil
}

func BuildGBestFitInShard(
	botAddr *address.Address,
	ownerAddr *address.Address,
	jettonMasterAddr *address.Address,
	vaultJettonWalletAddr *address.Address,
	codeCell *cell.Cell,
) (ed25519.PrivateKey, *address.Address, error) {

	var pk ed25519.PrivateKey
	var gAddr *address.Address

	for {
		_, pk, _ = ed25519.GenerateKey(nil)
		_, gAddr, _ = BuildG(botAddr, pk, tlb.MustFromTON("0.1"))
		gJettonWalletCell := mywallet.CalculateUserJettonWalletAddress(
			gAddr,
			jettonMasterAddr,
			codeCell,
		)

		if match(
			vaultJettonWalletAddr.Data(),
			utils.CellToAddress(gJettonWalletCell).Data(),
			gAddr.Data()) {
			break
		}
	}

	return pk, gAddr, nil
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

func createCommentCell(text string) (*cell.Cell, error) {
	// comment ident
	root := cell.BeginCell().MustStoreUInt(0, 32)

	if err := root.StoreStringSnake(text); err != nil {
		return nil, fmt.Errorf("failed to build comment: %w", err)
	}

	return root.EndCell(), nil
}
