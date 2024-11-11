package cli

import (
	"fmt"

	"github.com/cmingxu/dedust/utils"
	cli2 "github.com/urfave/cli/v2"
	"github.com/xssnick/tonutils-go/ton/wallet"
)

var (
	newSeedCmd = &cli2.Command{
		Name:        "new-seed",
		Description: "to generate a new seed for a wallet",
		Flags:       []cli2.Flag{},
		Action: func(c *cli2.Context) error {
			if err := utils.SetupLogger(c.String("loglevel")); err != nil {
				return err
			}

			fmt.Println(wallet.NewSeed())
			return nil
		},
	}
)
