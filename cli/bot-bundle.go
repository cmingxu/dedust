package cli

import (
	"time"

	"github.com/cmingxu/dedust/bot"
	"github.com/cmingxu/dedust/utils"
	"github.com/jmoiron/sqlx"
	cli2 "github.com/urfave/cli/v2"
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/tlb"
)

var botBundleCmd = &cli2.Command{
	Name:        "bot-bundle",
	Description: "to bundle some ton from bot",
	Flags: []cli2.Flag{
		&tonConfig,
		&walletSeed,
		&poolAddr,
		&amount,
		&host,
		&port,
		&user,
		&password,
		&database,
	},
	Action: func(c *cli2.Context) error {
		if err := utils.SetupLogger(c.String("loglevel")); err != nil {
			return err
		}
		return botBundle(c)
	},
}

func botBundle(c *cli2.Context) error {
	var (
		err error
	)
	botWalletSeeds := MustLoadSeeds(c.String("wallet-seed"))

	// establish connection to the server
	connPool, ctx, err := utils.GetConnectionPool(c.String("ton-config"))
	if err != nil {
		return err
	}
	client := utils.GetAPIClientWithTimeout(connPool, time.Second*10)

	amount, err := tlb.FromTON(c.String("amount"))
	if err != nil {
		return err
	}

	poolAddr, err := address.ParseAddr(c.String("pool-addr"))
	if err != nil {
		return err
	}

	db, err := sqlx.Connect("mysql", utils.ConstructDSN(c))
	if err != nil {
		return err
	}
	defer db.Close()

	return bot.Bundle(
		ctx,
		connPool,
		client,
		pkFromSeed(botWalletSeeds),
		poolAddr,
		amount,
		tlb.MustFromTON("0.00000001"),
		db,
	)
}
