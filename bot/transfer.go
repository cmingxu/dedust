package bot

import (
	"context"
	"crypto/ed25519"
	"fmt"

	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/tlb"
	"github.com/xssnick/tonutils-go/ton"
)

func Transfer(ctx context.Context,
	client ton.APIClientWrapped,
	botprivateKey ed25519.PrivateKey,
	destAddr *address.Address,
	amount tlb.Coins,
) error {
	botAddr := botAddress(botprivateKey.Public().(ed25519.PublicKey))

	masterBlock, err := client.GetMasterchainInfo(ctx)
	if err != nil {
		return err
	}

	account, err := client.WaitForBlock(masterBlock.SeqNo).GetAccount(ctx, masterBlock, botAddr)
	if err != nil {
		return err
	}

	fmt.Println("Dest address:", destAddr.String())
	fmt.Println("Bot address:", botAddr.String())
	fmt.Println("Bot balance:", account.State.Balance)

	if account.State.Balance.Nano().Cmp(amount.Nano()) < 0 {
		return fmt.Errorf("not enough balance")
	}

	botWallet := NewBotWallet(ctx, client, botAddr, botprivateKey, 4)
	return botWallet.TransferNoBounce(ctx, destAddr, amount, "you deserved it", true)
}
