package cli

import (
	"crypto/ed25519"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/cmingxu/dedust/bot"
	"github.com/cmingxu/dedust/printer"
	"github.com/cmingxu/dedust/utils"
	"github.com/jmoiron/sqlx"
	cli2 "github.com/urfave/cli/v2"
)

func botCollectG(c *cli2.Context) error {
	var (
		err error
	)
	botWalletSeeds := MustLoadSeeds(c.String("bot-wallet-seed"))

	gPKStr := c.String("private-key-of-g")
	if len(gPKStr) == 0 {
		return fmt.Errorf("private-key-of-g is required")
	}

	gpkRaw, err := hex.DecodeString(gPKStr)
	if err != nil {
		return err
	}

	gpk := ed25519.PrivateKey(gpkRaw)

	// establish connection to the server
	connPool, ctx, err := utils.GetConnectionPool(c.String("ton-config"))
	if err != nil {
		return err
	}
	client := utils.GetAPIClientWithTimeout(connPool, time.Second*30)

	return bot.CollectG(
		ctx,
		client,
		pkFromSeed(botWalletSeeds),
		gpk,
	)
}

func botCollectGAuto(c *cli2.Context) error {
	var (
		err error
	)
	botWalletSeeds := MustLoadSeeds(c.String("bot-wallet-seed"))
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
	botAddr := bot.BotAddress(botPk.Public().(ed25519.PublicKey))

	collector := printer.NewGCollector(ctx, client, db, botPk, botAddr)
	return collector.Run()
}
