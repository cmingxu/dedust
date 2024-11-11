package cli

import (
	"crypto/ed25519"
	"fmt"
	"time"

	"github.com/cmingxu/dedust/bot"
	"github.com/cmingxu/dedust/utils"
	"github.com/urfave/cli/v2"
	cli2 "github.com/urfave/cli/v2"
	"github.com/xssnick/tonutils-go/ton/wallet"
)

var deployCmd = &cli.Command{
	Name:        "deploy",
	Description: "to deploy a bot wallet",
	Flags: []cli2.Flag{
		&tonConfig,
		&mainWalletSeed,
		&walletSeed,
		&botType,
	},
	Action: func(c *cli2.Context) error {
		if err := utils.SetupLogger(c.String("loglevel")); err != nil {
			return err
		}
		return deploy(c)
	},
}

func deploy(c *cli2.Context) error {
	var err error

	mainWalletSeeds := MustLoadSeeds(c.String("main-wallet-seed"))
	botWalletSeeds := MustLoadSeeds(c.String("wallet-seed"))

	// establish connection to the server
	pool, ctx, err := utils.GetConnectionPool(c.String("ton-config"))
	if err != nil {
		return err
	}
	client := utils.GetAPIClientWithTimeout(pool, time.Second*10)

	// initialize main wallet
	mainWallet, err := wallet.FromSeed(client, mainWalletSeeds, wallet.V4R2)
	if err != nil {
		return err
	}

	pk := pkFromSeed(botWalletSeeds)
	fmt.Println("Main wallet address:", mainWallet.Address().String())
	fmt.Println("Deloy wallet public key:", pk.Public().(ed25519.PublicKey))


	botType :=  mustLoadBotType(c.String("bot-type"))
	addr, err := bot.DeployBot(ctx, mainWallet, pk, botType)
	if err != nil {
		return err
	}

	fmt.Println("Bot wallet address:", addr.String())

	return nil
}
