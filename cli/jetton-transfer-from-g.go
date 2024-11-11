package cli

import (
	"crypto/ed25519"
	"encoding/hex"
	"time"

	"github.com/cmingxu/dedust/bot"
	"github.com/cmingxu/dedust/utils"
	cli2 "github.com/urfave/cli/v2"
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/tlb"
)

var jettonTransferFromGCmd = &cli2.Command{
	Name: "jetton-transfer-from-g",
	Flags: []cli2.Flag{
		&amount,
		&destAddr,
		&walletSeed,
		&tonConfig,
		&privateKeyOfG,
		&jettonWalletAddr,
	},
	Description: "collect G auto",
	Action: func(c *cli2.Context) error {
		return jettonTransferFromG(c)
	},
}

func jettonTransferFromG(c *cli2.Context) error {
	var (
		err error
	)
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

	pkOfGStr := c.String("private-key-of-g")
	pkOfGRaw, err := hex.DecodeString(pkOfGStr)
	if err != nil {
		return err
	}

	pkOfG := ed25519.PrivateKey(pkOfGRaw)
	jettonWalletAddr, err := address.ParseAddr(c.String("jetton-wallet-addr"))
	if err != nil {
		return err
	}

	return bot.JettonTransferFromG(
		ctx,
		client,
		pkFromSeed(botWalletSeeds),
		pkOfG,
		jettonWalletAddr,
		destAddr,
		amount,
	)
}
