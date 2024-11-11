package cli

import (
	"time"

	"github.com/cmingxu/dedust/bot"
	"github.com/cmingxu/dedust/utils"
	cli2 "github.com/urfave/cli/v2"
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/tlb"
)

var dummyTransferCmd = &cli2.Command{
	Name: "dummy-transfer",
	Flags: []cli2.Flag{
		&tonConfig,
		&walletSeed,
		&destAddr,
		&amount,
	},
	Description: "generate base64 send message(external)",
	Action: func(c *cli2.Context) error {
		return dummyTransfer(c)
	},
}

func dummyTransfer(c *cli2.Context) error {
	botWalletSeeds := MustLoadSeeds(c.String("wallet-seed"))

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

	return bot.DummyTransfer(
		ctx,
		client,
		pkFromSeed(botWalletSeeds),
		destAddr,
		amount,
	)
}
