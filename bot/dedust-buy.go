package bot

import (
	"crypto/ed25519"
	"fmt"
	"time"

	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/tlb"
	"github.com/xssnick/tonutils-go/ton"
	"github.com/xssnick/tonutils-go/ton/wallet"
	"golang.org/x/net/context"
)

func DedustBuy(
	ctx context.Context,
	client ton.APIClientWrapped,
	botprivateKey ed25519.PrivateKey,
	botType BotType,
	poolAddr *address.Address,
	amount tlb.Coins,
	limit tlb.Coins,
) error {
	addr := WalletAddress(botprivateKey.Public().(ed25519.PublicKey), nil, botType)

	fmt.Println("Bot address:", addr.String())
	fmt.Println("Bot type:", botType)
	fmt.Println("Pool address:", poolAddr.String())
	fmt.Println("Amount Ton: ", amount.Nano())
	fmt.Println("Limit Ton: ", limit.Nano())

	masterBlock, err := client.GetMasterchainInfo(ctx)
	if err != nil {
		return err
	}
	account, err := client.WaitForBlock(masterBlock.SeqNo).GetAccount(ctx, masterBlock, addr)
	if err != nil {
		return err
	}
	fmt.Println("Acc balance:", account.State.Balance)

	seqno, err := getSeqno(ctx, client, masterBlock, addr)
	if err != nil {
		return err
	}
	fmt.Println("Acc seqno:", seqno)
	v4Wallet := NewWallet(ctx, client, V4R2, botprivateKey, nil, seqno)
	fmt.Println("V4 wallet:", v4Wallet)

	botAddr := botAddress(botprivateKey.Public().(ed25519.PublicKey))
	fmt.Println("Bot address:", botAddr.String())

	deadline := time.Now().Unix() - 10

	msg := v4Wallet.BuildDedustBuy(poolAddr, amount.Nano(), limit.Nano(), uint64(deadline))

	fmt.Println("Bundle message:", msg)

	return v4Wallet.SendMany(ctx, 0, []*wallet.Message{msg}, false)
}
