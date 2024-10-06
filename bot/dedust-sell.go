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
	botprivateKey ed25519.PrivateKey,
	jettonMasterAddr *address.Address,
	dedustVaultAddr *address.Address,
	poolAddr *address.Address,
) error {
	botAddr := botAddress(botprivateKey.Public().(ed25519.PublicKey))
	fmt.Println("Bot address:", botAddr.String())
	fmt.Println("Jetton master address:", jettonMasterAddr.String())
	fmt.Println("Dedust vault address:", dedustVaultAddr.String())
	fmt.Println("Pool address:", poolAddr.String())

	botJettonWalletAddr, jettonAmountOfBot, err := botJettonWalletAddrAndAmount(ctx,
		client,
		jettonMasterAddr,
		botAddr)
	if err != nil {
		return err
	}

	fmt.Println("Bot jetton wallet address:", botJettonWalletAddr.String())
	fmt.Println("Jetton amount of bot:", jettonAmountOfBot.String())

	botWallet := NewBotWallet(ctx, client, botAddr, botprivateKey, 2)
	msg := botWallet.BuildDedustSell(botJettonWalletAddr,
		dedustVaultAddr,
		poolAddr,
		jettonAmountOfBot,
		tlb.MustFromTON("0.00001"), // ton limit expected
	)

	fmt.Println("Dedust sell message:", msg)

	return botWallet.Send(ctx, msg, true)
}

func botJettonWalletAddrAndAmount(ctx context.Context,
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
