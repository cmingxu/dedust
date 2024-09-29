package cli

import (
	"fmt"
	"os"

	"github.com/cmingxu/dedust/utils"
	cli2 "github.com/urfave/cli/v2"
	"github.com/xssnick/tonutils-go/ton/wallet"
)

var (
	flagLogLevel = cli2.StringFlag{
		Name:  "loglevel",
		Value: "info",
		Usage: "Set the log level (debug, info, warn, error, fatal, panic)",
	}

	host = cli2.StringFlag{
		Name:  "host",
		Value: "localhost",
	}

	port = cli2.IntFlag{
		Name:  "port",
		Value: 3306,
	}

	user = cli2.StringFlag{
		Name:  "user",
		Value: "root",
	}

	password = cli2.StringFlag{
		Name:  "password",
		Value: "password",
	}

	database = cli2.StringFlag{
		Name:  "database",
		Value: "mydb",
	}

	tonConfig = cli2.StringFlag{
		Name:  "ton-config",
		Value: "https://ton.org/global-config.json",
		Usage: "Set the TON config url path or local file path",
	}

	mainWalletSeed = cli2.StringFlag{
		Name:  "main-wallet-seed",
		Value: "./main-wallet-seed.txt",
		Usage: "Set the main wallet seed file path or seed itself",
	}

	botWalletSeed = cli2.StringFlag{
		Name:  "bot-wallet-seed",
		Value: "./bot-wallet-seed.txt",
		Usage: "Set the new wallet seed file path or seed itself",
	}

	amount = cli2.StringFlag{
		Name:  "amount",
		Value: "0.1",
		Usage: "Set the amount to send",
	}

	destAddr = cli2.StringFlag{
		Name:  "dest-addr",
		Value: "UQCwSxqefElovEPlpZ8bIEL_KXqWuqoOhwb65uYjos9bCDcM",
		Usage: "main wallet address to send",
	}

	poolAddr = cli2.StringFlag{
		Name:  "pool-addr",
		Usage: "pool address to trade",
	}
)

func Run(args []string) int {
	app := &cli2.App{
		Name:  "dedust",
		Usage: "A CLI tool to dedust your code",
		Flags: []cli2.Flag{
			&flagLogLevel,
		},
		Commands: []*cli2.Command{
			{
				Name:  "info",
				Flags: []cli2.Flag{&host, &port, &user, &password, &database},
				Action: func(c *cli2.Context) error {
					if err := utils.SetupLogger(c.String("loglevel")); err != nil {
						return err
					}

					return info(c)
				},
			},

			{
				Name: "bootstrap",
				Flags: []cli2.Flag{
					&host,
					&port,
					&user,
					&password,
					&database,
				},
				Action: func(c *cli2.Context) error {
					return bootstrap(c)
				},
			},

			{
				Name: "sync-pool",
				Flags: []cli2.Flag{
					&host,
					&port,
					&user,
					&password,
					&database,
				},
				Action: func(c *cli2.Context) error {
					if err := utils.SetupLogger(c.String("loglevel")); err != nil {
						return err
					}
					return syncPool(c)
				},
			},

			{
				Name:        "detect",
				Description: "to detect and save dedust trading infomation",
				Flags: []cli2.Flag{
					&host,
					&port,
					&user,
					&password,
					&database,
				},
				Action: func(c *cli2.Context) error {
					if err := utils.SetupLogger(c.String("loglevel")); err != nil {
						return err
					}
					return detect(c)
				},
			},

			{
				Name:        "bot-deploy",
				Description: "to deploy a bot wallet",
				Flags: []cli2.Flag{
					&tonConfig,
					&mainWalletSeed,
					&botWalletSeed,
				},
				Action: func(c *cli2.Context) error {
					if err := utils.SetupLogger(c.String("loglevel")); err != nil {
						return err
					}
					return deployBot(c)
				},
			},

			{
				Name:        "bot-info",
				Description: "to get bot info",
				Flags: []cli2.Flag{
					&tonConfig,
					&botWalletSeed,
				},
				Action: func(c *cli2.Context) error {
					if err := utils.SetupLogger(c.String("loglevel")); err != nil {
						return err
					}
					return infoBot(c)
				},
			},

			{
				Name:        "bot-tonup",
				Description: "to get bot info",
				Flags: []cli2.Flag{
					&tonConfig,
					&mainWalletSeed,
					&botWalletSeed,
					&amount,
				},
				Action: func(c *cli2.Context) error {
					if err := utils.SetupLogger(c.String("loglevel")); err != nil {
						return err
					}
					return tonupBot(c)
				},
			},

			{
				Name:        "bot-transfer",
				Description: "to transfer some ton from bot",
				Flags: []cli2.Flag{
					&tonConfig,
					&botWalletSeed,
					&destAddr,
					&amount,
				},
				Action: func(c *cli2.Context) error {
					if err := utils.SetupLogger(c.String("loglevel")); err != nil {
						return err
					}
					return botTransfer(c)
				},
			},
			{
				Name:        "bot-bundle",
				Description: "to bundle some ton from bot",
				Flags: []cli2.Flag{
					&tonConfig,
					&botWalletSeed,
					&poolAddr,
					&amount,
				},
				Action: func(c *cli2.Context) error {
					if err := utils.SetupLogger(c.String("loglevel")); err != nil {
						return err
					}
					return botBundle(c)
				},
			},

			{
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
			},
		},
	}

	if err := app.Run(args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	return 0
}
