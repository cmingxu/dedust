package cli

import (
	"fmt"
	"os"

	cli2 "github.com/urfave/cli/v2"
)

func Run(args []string) int {
	app := &cli2.App{
		Name:  "dedust",
		Usage: "A CLI tool to dedust your code",
		Flags: []cli2.Flag{
			&flagLogLevel,
		},
		Commands: []*cli2.Command{
			newSeedCmd,
			newSeedSameShardCmd,
			infoCmd,
			bootstrapCmd,
			syncPoolCmd,
			detectCmd,
			memPoolCmd,
			deployCmd,
			walletInfoCmd,
			tonupCmd,
			tonTransferCmd,
			printerCmd,
			botBundleCmd,
			dedustSellCmd,
			dedustBuyCmd,
			collectGCmd,
			collectGAutoCmd,
			LiteserverIpsCmd,
			dummyTransferCmd,
			generateGCmd,
			jettonTransferFromGCmd,
			checkPoolReserveCmd,
		},
	}

	if err := app.Run(args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	return 0
}
