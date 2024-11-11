package bot

import (
	"crypto/ed25519"
	"encoding/base64"
	"encoding/hex"
	"fmt"

	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/tlb"
	"github.com/xssnick/tonutils-go/ton"
	"github.com/xssnick/tonutils-go/ton/wallet"
	"golang.org/x/net/context"
)

func DummyTransfer(ctx context.Context,
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

	botWallet := NewWallet(ctx, client, Bot, botprivateKey, nil, 1143)
	// return botWallet.TransferNoBounce(ctx, destAddr, amount, "you deserved it", true)

	transferInternalMsg, err := botWallet.BuildTransfer(destAddr, amount, false, "(_))")
	if err != nil {
		return err
	}

	c := context.Background()
	externalMsg, err := botWallet.BuildExternalMessageForMany(c, 0, []*wallet.Message{transferInternalMsg})
	if err != nil {
		return err
	}

	cell, err := tlb.ToCell(externalMsg)
	if err != nil {
		return err
	}

	boc := cell.ToBOC()

	fmt.Println("boc")
	fmt.Println(hex.EncodeToString(boc))

	fmt.Println("")
	fmt.Println("")

	fmt.Println("base64", base64.RawStdEncoding.EncodeToString(boc))

	fmt.Println("base64")
	fmt.Println("base64", base64.StdEncoding.EncodeToString(boc))

	return nil
}
