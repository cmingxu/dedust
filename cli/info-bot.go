package cli

import (
	"time"

	"github.com/cmingxu/dedust/bot"
	"github.com/cmingxu/dedust/utils"
	cli2 "github.com/urfave/cli/v2"
)

// infoBot prints bot info
func infoBot(c *cli2.Context) error {
	var (
		err error
	)
	botWalletSeeds := MustLoadSeeds(c.String("wallet-seed"))

	// establish connection to the server
	pool, ctx, err := utils.GetConnectionPool(c.String("ton-config"))
	if err != nil {
		return err
	}
	client := utils.GetAPIClientWithTimeout(pool, time.Second*30)

	botType := mustLoadBotType(c.String("bot-type"))
	return bot.WalletInfo(ctx, client, pkFromSeed(botWalletSeeds), botType)
}
