package utils

import (
	"github.com/xssnick/tonutils-go/ton"
	"github.com/xssnick/tonutils-go/ton/wallet"
)

func MainWallet(api *ton.APIClient, seeds []string) (w *wallet.Wallet, err error) {
	return wallet.FromSeed(api, seeds, wallet.V4R2)
}
