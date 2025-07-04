package wallet

import (
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/tlb"
	"github.com/xssnick/tonutils-go/tvm/cell"
)

var (
	ZeroCoins = tlb.MustFromTON("0")
)

var (
	WalletCodes = map[string]string{
		"": "",
	}
)

// https://github.com/ton-blockchain/token-contract/blob/2c13d3ef61ca4288293ad65bf0cfeaed83879b93/ft/jetton-utils.fc

func PackJettonWalletData(
	balance tlb.Coins,
	ownerAddress *address.Address,
	jettonMasterAddress *address.Address,
	jettonWalletCode *cell.Cell) *cell.Cell {
	return cell.BeginCell().MustStoreCoins(balance.Nano().Uint64()).MustStoreAddr(ownerAddress).MustStoreAddr(jettonMasterAddress).MustStoreRef(jettonWalletCode).EndCell()
}

func CalculateJettonWalletStateInit(
	ownerAddress *address.Address,
	jettonMasterAddress *address.Address,
	jettonWalletCode *cell.Cell) *cell.Cell {
	return cell.BeginCell().
		MustStoreUInt(0, 2).
		MustStoreMaybeRef(jettonWalletCode).
		MustStoreMaybeRef(PackJettonWalletData(ZeroCoins, ownerAddress, jettonMasterAddress, jettonWalletCode)).
		MustStoreUInt(0, 1).
		EndCell()
}

func CalculateJettonWalletAddress(stateInit *cell.Cell) *cell.Cell {
	stateInitHash := stateInit.Hash()

	return cell.BeginCell().
		MustStoreUInt(4, 3).
		MustStoreInt(0, 8). // workchain
		MustStoreSlice(stateInitHash, 256).
		EndCell()
}

func CalculateUserJettonWalletAddress(ownerAddress *address.Address,
	jettonMasterAddress *address.Address,
	jettonWalletCode *cell.Cell) *cell.Cell {
	return CalculateJettonWalletAddress(CalculateJettonWalletStateInit(ownerAddress, jettonMasterAddress, jettonWalletCode))
}
