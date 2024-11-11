package cli

import (
	"crypto/ed25519"

	"github.com/cmingxu/dedust/bot"
	"github.com/cmingxu/dedust/printer"
	"github.com/cmingxu/dedust/utils"
	cli2 "github.com/urfave/cli/v2"
)

var printerCmd = &cli2.Command{
	Name: "printer",
	Flags: []cli2.Flag{
		&tonConfig,
		&walletSeed,
		&wsEndpoint,
		&printerOutPath,
		&sendCnt,
		&useTonAPI,
		&useTonCenter,
		&useANDL,
		&limit,
		&host,
		&port,
		&user,
		&password,
		&database,
	},
	Description: "to print money",
	Action: func(c *cli2.Context) error {
		return startPrinter(c)
	},
}

func startPrinter(c *cli2.Context) error {
	if err := utils.SetupLogger(c.String("loglevel")); err != nil {
		return err
	}

	botWalletSeeds := MustLoadSeeds(c.String("wallet-seed"))

	botprivateKey := pkFromSeed(botWalletSeeds)
	botAddr := bot.WalletAddress(botprivateKey.Public().(ed25519.PublicKey), nil, bot.Bot)

	p, err := printer.NewPrinter(
		c.String("ton-config"),
		botAddr,
		botprivateKey,
		c.String("ws-endpoint"),
		c.String("out-path"),
		uint32(c.Uint("send-cnt")),
		c.Bool("use-tonapi"),
		c.Bool("use-toncenter"),
		c.Bool("use-andl"),
		c.String("limit"),
		utils.ConstructDSN(c),
	)
	if err != nil {
		return err
	}

	return p.Run()
}
