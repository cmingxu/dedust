package cli

import (
	"time"

	"github.com/cmingxu/dedust/bot"
	"github.com/cmingxu/dedust/utils"
	cli2 "github.com/urfave/cli/v2"
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/tlb"
)

var tonTransferCmd = &cli2.Command{
	Name:        "transfer",
	Description: "to transfer ton from wallet to dest-addr",
	Flags: []cli2.Flag{
		&tonConfig,
		&walletSeed,
		&destAddr,
		&amount,
		&botType,
	},
	Action: func(c *cli2.Context) error {
		if err := utils.SetupLogger(c.String("loglevel")); err != nil {
			return err
		}
		return tonTransfer(c)
	},
}

func tonTransfer(c *cli2.Context) error {
	var (
		err error
	)
	walletSeeds := MustLoadSeeds(c.String("wallet-seed"))
	botType := mustLoadBotType(c.String("bot-type"))

	// establish connection to the server
	pool, ctx, err := utils.GetConnectionPool(c.String("ton-config"))
	if err != nil {
		return err
	}
	client := utils.GetAPIClientWithTimeout(pool, time.Second*10)

	amount, err := tlb.FromTON(c.String("amount"))
	if err != nil {
		return err
	}

	destAddr, err := address.ParseAddr(c.String("dest-addr"))
	if err != nil {
		return err
	}

	return bot.Transfer(
		ctx,
		client,
		pkFromSeed(walletSeeds),
		botType,
		destAddr,
		amount,
	)
}
