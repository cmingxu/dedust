package cli

import (
	"crypto/ed25519"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/cmingxu/dedust/bot"
	"github.com/cmingxu/dedust/utils"
	cli2 "github.com/urfave/cli/v2"
)

var collectGCmd = &cli2.Command{
	Name: "collect-g",
	Flags: []cli2.Flag{
		&host,
		&port,
		&user,
		&password,
		&database,
		&walletSeed,
		&privateKeyOfG,
		&tonConfig,
	},
	Description: "collect G",
	Action: func(c *cli2.Context) error {
		return botCollectG(c)
	},
}

func botCollectG(c *cli2.Context) error {
	var (
		err error
	)
	botWalletSeeds := MustLoadSeeds(c.String("wallet-seed"))

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
