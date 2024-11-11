package cli

import (
	"time"

	"github.com/cmingxu/dedust/bot"
	"github.com/cmingxu/dedust/utils"
	cli2 "github.com/urfave/cli/v2"
	"github.com/xssnick/tonutils-go/address"
)

var dedustSellCmd = &cli2.Command{
	Name:        "dedust-sell",
	Description: "to sell some ton from wallet",
	Flags: []cli2.Flag{
		&tonConfig,
		&walletSeed,
		&jettonMasterAddr,
		&vaultAddr,
		&poolAddr,
		&botType,
	},
	Action: func(c *cli2.Context) error {
		if err := utils.SetupLogger(c.String("loglevel")); err != nil {
			return err
		}
		return dedustSell(c)
	},
}

func dedustSell(c *cli2.Context) error {
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

	jettonMaterAddr, err := address.ParseAddr(c.String("jetton-master-addr"))
	if err != nil {
		return err
	}

	vaultAddr, err := address.ParseAddr(c.String("vault-addr"))
	if err != nil {
		return err
	}

	poolAddr, err := address.ParseAddr(c.String("pool-addr"))
	if err != nil {
		return err
	}

	botType := mustLoadBotType(c.String("bot-type"))

	return bot.DedustSell(
		ctx,
		client,
		pkFromSeed(walletSeeds),
		botType,
		jettonMaterAddr,
		vaultAddr,
		poolAddr,
	)
}
