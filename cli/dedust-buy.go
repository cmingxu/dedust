package cli

import (
	"time"

	"github.com/cmingxu/dedust/bot"
	"github.com/cmingxu/dedust/utils"
	cli2 "github.com/urfave/cli/v2"
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/tlb"
)

var dedustBuyCmd = &cli2.Command{
	Name:        "dedust-buy",
	Description: "buy dedust from pool",
	Flags: []cli2.Flag{
		&tonConfig,
		&walletSeed,
		&poolAddr,
		&amount,
		&limit,
		&botType,
	},
	Action: func(c *cli2.Context) error {
		if err := utils.SetupLogger(c.String("loglevel")); err != nil {
			return err
		}
		return dedustBuy(c)
	},
}

func dedustBuy(c *cli2.Context) error {
	var (
		err error
	)
	walletSeeds := MustLoadSeeds(c.String("wallet-seed"))

	// establish connection to the server
	connPool, ctx, err := utils.GetConnectionPool(c.String("ton-config"))
	if err != nil {
		return err
	}
	client := utils.GetAPIClientWithTimeout(connPool, time.Second*30)

	poolAddr, err := address.ParseAddr(c.String("pool-addr"))
	if err != nil {
		return err
	}

	amount := tlb.MustFromTON(c.String("amount"))
	limit := tlb.MustFromTON(c.String("limit"))

	botType := mustLoadBotType(c.String("bot-type"))

	return bot.DedustBuy(
		ctx,
		client,
		pkFromSeed(walletSeeds),
		botType,
		poolAddr,
		amount,
		limit,
	)
}
