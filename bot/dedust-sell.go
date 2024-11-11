package bot

import (
	"crypto/ed25519"
	"fmt"

	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/tlb"
	"github.com/xssnick/tonutils-go/ton"
	"github.com/xssnick/tonutils-go/ton/jetton"
	"golang.org/x/net/context"
)

func DedustSell(
	ctx context.Context,
	client ton.APIClientWrapped,
	privateKey ed25519.PrivateKey,
	botType BotType,
	jettonMasterAddr *address.Address,
	dedustVaultAddr *address.Address,
	poolAddr *address.Address,
) error {
	addr := WalletAddress(privateKey.Public().(ed25519.PublicKey), nil, botType)
	fmt.Println("Bot address:", addr.String())
	fmt.Println("Jetton master address:", jettonMasterAddr.String())
	fmt.Println("Dedust vault address:", dedustVaultAddr.String())
	fmt.Println("Pool address:", poolAddr.String())

	jettonWalletAddr, jettonAmount, err := jettonWalletAddrAndAmount(ctx,
		client,
		jettonMasterAddr,
		addr)
	if err != nil {
		return err
	}

	fmt.Println("Jetton wallet address:", jettonWalletAddr.String())
	fmt.Println("Jetton amount:", jettonAmount.String())

	masterBlock, err := client.GetMasterchainInfo(ctx)
	if err != nil {
		return err
	}

	account, err := client.WaitForBlock(masterBlock.SeqNo).GetAccount(ctx, masterBlock, addr)
	if err != nil {
		return err
	}

	fmt.Println("Address:", addr.String())
	fmt.Println("Balance:", account.State.Balance)
	seqno, err := getSeqno(ctx, client, masterBlock, addr)
	if err != nil {
		return err
	}
	fmt.Println("Bot seqno:", seqno)

	wallet := NewWallet(ctx, client, botType, privateKey, nil, seqno)
	msg := wallet.BuildDedustSell(jettonWalletAddr,
		dedustVaultAddr,
		poolAddr,
		jettonAmount,
		tlb.MustFromTON("0.00001"), // ton limit expected
	)

	fmt.Println("Dedust sell message:", msg)

	return wallet.Send(ctx, 0, msg, true)
}

func jettonWalletAddrAndAmount(ctx context.Context,
	client ton.APIClientWrapped,
	jettonMasterAddr *address.Address,
	botAddr *address.Address) (*address.Address, tlb.Coins, error) {
	master := jetton.NewJettonMasterClient(client, jettonMasterAddr)

	tokenWallet, err := master.GetJettonWallet(ctx, botAddr)
	if err != nil {
		return nil, tlb.Coins{}, err
	}

	tokenBalance, err := tokenWallet.GetBalance(ctx)
	if err != nil {
		return nil, tlb.Coins{}, err
	}

	return tokenWallet.Address(), tlb.MustFromNano(tokenBalance, 9), nil
}
