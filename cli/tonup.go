package cli

import (
	"time"

	"github.com/cmingxu/dedust/bot"
	"github.com/cmingxu/dedust/utils"
	cli2 "github.com/urfave/cli/v2"
	"github.com/xssnick/tonutils-go/tlb"
	"github.com/xssnick/tonutils-go/ton/wallet"
)

var tonupCmd = &cli2.Command{
	Name:        "tonup",
	Description: "transfer amount of ton to wallet",
	Flags: []cli2.Flag{
		&tonConfig,
		&mainWalletSeed,
		&walletSeed,
		&amount,
		&botType,
	},
	Action: func(c *cli2.Context) error {
		if err := utils.SetupLogger(c.String("loglevel")); err != nil {
			return err
		}
		return tonup(c)
	},
}

func tonup(c *cli2.Context) error {
	var (
		err error
	)
	mainWalletSeeds := MustLoadSeeds(c.String("main-wallet-seed"))
	walletSeeds := MustLoadSeeds(c.String("wallet-seed"))
	botType := mustLoadBotType(c.String("bot-type"))

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

	amount, err := tlb.FromTON(c.String("amount"))
	if err != nil {
		return err
	}

	return bot.Tonup(
		ctx,
		client,
		mainWallet,
		pkFromSeed(walletSeeds),
		botType,
		amount,
	)
}
