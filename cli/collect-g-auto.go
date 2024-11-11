package cli

import (
	"crypto/ed25519"
	"time"

	"github.com/cmingxu/dedust/bot"
	"github.com/cmingxu/dedust/printer"
	"github.com/cmingxu/dedust/utils"
	"github.com/jmoiron/sqlx"
	cli2 "github.com/urfave/cli/v2"
)

var collectGAutoCmd = &cli2.Command{
	Name: "bot-collect-g-auto",
	Flags: []cli2.Flag{
		&host,
		&port,
		&user,
		&password,
		&database,
		&walletSeed,
		&tonConfig,
	},
	Description: "bot collect G auto",
	Action: func(c *cli2.Context) error {
		return botCollectGAuto(c)
	},
}

func botCollectGAuto(c *cli2.Context) error {
	var (
		err error
	)
	botWalletSeeds := MustLoadSeeds(c.String("wallet-seed"))
	// establish connection to the server
	connPool, ctx, err := utils.GetConnectionPool(c.String("ton-config"))
	if err != nil {
		return err
	}
	client := utils.GetAPIClientWithTimeout(connPool, time.Second*30)

	db, err := sqlx.Connect("mysql", utils.ConstructDSN(c))
	if err != nil {
		return err
	}

	botPk := pkFromSeed(botWalletSeeds)
	botAddr := bot.WalletAddress(botPk.Public().(ed25519.PublicKey), nil, bot.Bot)

	collector := printer.NewGCollector(ctx, client, db, botPk, botAddr)
	return collector.Run()
}
