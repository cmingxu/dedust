package cli

import (
	"crypto/ed25519"
	"time"

	"github.com/cmingxu/dedust/bot"
	printerPkg "github.com/cmingxu/dedust/printer"
	"github.com/cmingxu/dedust/utils"
	cli2 "github.com/urfave/cli/v2"
)

func printer(c *cli2.Context) error {
	botWalletSeeds := MustLoadSeeds(c.String("bot-wallet-seed"))

	// establish connection to the server
	pool, ctx, err := utils.GetConnectionPool(c.String("ton-config"))
	if err != nil {
		return err
	}
	client := utils.GetAPIClientWithTimeout(pool, time.Second*10)

	botprivateKey := pkFromSeed(botWalletSeeds)
	botAddr := bot.BotAddress(botprivateKey.Public().(ed25519.PublicKey))

	p, err := printerPkg.NewPrinter(ctx,
		pool,
		client,
		botAddr,
		botprivateKey,
		c.String("ws-endpoint"),
		c.String("out-path"),
		uint32(c.Uint("send-cnt")),
	)
	if err != nil {
		return err
	}

	return p.Run()
}
