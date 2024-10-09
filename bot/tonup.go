package bot

import (
	"context"
	"crypto/ed25519"
	"fmt"

	"github.com/xssnick/tonutils-go/tlb"
	"github.com/xssnick/tonutils-go/ton"
	"github.com/xssnick/tonutils-go/ton/wallet"
)

var (
	// guard from sending too much money
	MaxTonupValue = tlb.MustFromTON("200")
)

func Tonup(ctx context.Context,
	client ton.APIClientWrapped,
	mainWallet *wallet.Wallet,
	botprivateKey ed25519.PrivateKey,
	amount tlb.Coins) error {

	if amount.Nano().Cmp(MaxTonupValue.Nano()) > 0 {
		return fmt.Errorf("too much value to send")
	}

	botAddr := botAddress(botprivateKey.Public().(ed25519.PublicKey))
	fmt.Println("Bot address:", botAddr.String())

	masterBlock, err := client.GetMasterchainInfo(ctx)
	if err != nil {
		return err
	}

	balance, err := mainWallet.GetBalance(ctx, masterBlock)
	if err != nil {
		return err
	}

	if balance.Nano().Cmp(amount.Nano()) < 0 {
		return fmt.Errorf("not enough balance")
	}

	fmt.Printf("transfer %s of TON to %s\n", amount.String(), botAddr.String())

	return mainWallet.TransferNoBounce(ctx, botAddr, amount, "6-_-9", true)
}
