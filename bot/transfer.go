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
	privateKey ed25519.PrivateKey,
	botType BotType,
	destAddr *address.Address,
	amount tlb.Coins,
) error {
	addr := WalletAddress(privateKey.Public().(ed25519.PublicKey), nil, botType)

	masterBlock, err := client.GetMasterchainInfo(ctx)
	if err != nil {
		return err
	}

	account, err := client.WaitForBlock(masterBlock.SeqNo).GetAccount(ctx, masterBlock, addr)
	if err != nil {
		return err
	}

	fmt.Println("Dest address:", destAddr.String())
	fmt.Println("Address:", addr.String())
	fmt.Println("Type", botType)
	fmt.Println("Balance:", account.State.Balance)
	seqno, err := getSeqno(ctx, client, masterBlock, addr)
	if err != nil {
		return err
	}
	fmt.Println("Bot seqno:", seqno)

	if account.State.Balance.Nano().Cmp(amount.Nano()) < 0 {
		return fmt.Errorf("not enough balance")
	}

	wallet := NewWallet(ctx, client, botType, privateKey, nil, seqno)
	return wallet.TransferNoBounce(ctx, destAddr, amount, "(^)", true)
}
