package cli

import (
	"fmt"
	"os"

	"github.com/cmingxu/dedust/utils"
	cli2 "github.com/urfave/cli/v2"
	"github.com/xssnick/tonutils-go/address"
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
		Value: "192.168.8.200",
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
		Value: "dedust",
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

	jettonMasterAddr = cli2.StringFlag{
		Name:  "jetton-master-addr",
		Usage: "jetton master address",
	}

	jettonWalletAddr = cli2.StringFlag{
		Name: "jetton-wallet-addr",
	}

	vaultAddr = cli2.StringFlag{
		Name:  "vault-addr",
		Usage: "dedust vault address",
	}

	preUpdateReserve = cli2.BoolFlag{
		Name:  "pre-update-reserve",
		Value: false,
		Usage: "whether update reserve before detect(with runGetMethod(get_reserves)",
	}

	wsEndpoint = cli2.StringFlag{
		Name:  "ws-endpoint",
		Value: "ws://170.187.148.100:8080/ws",
		Usage: "Set the websocket endpoint",
	}

	printerOutPath = cli2.StringFlag{
		Name:  "out-path",
		Value: "out.csv",
		Usage: "Set the output file path",
	}

	output = cli2.StringFlag{
		Name:  "detect-output",
		Value: "detect-out.txt",
		Usage: "Set the output file path",
	}

	sendCnt = cli2.IntFlag{
		Name:  "send-cnt",
		Value: 2,
		Usage: "Set the send count in printer",
	}

	useTonAPI = cli2.BoolFlag{
		Name:  "use-tonapi",
		Value: false,
		Usage: "Set whether use ton api",
	}

	useTonCenter = cli2.BoolFlag{
		Name:  "use-toncenter",
		Value: false,
		Usage: "Set whether use ton center",
	}

	useANDL = cli2.BoolFlag{
		Name:  "use-andl",
		Value: true,
		Usage: "Set whether use andl",
	}

	limit = cli2.StringFlag{
		Name:  "limit",
		Value: "50",
		Usage: "Set the limit amount of TON to bundle",
	}

	privateKeyOfG = cli2.StringFlag{
		Name:  "private-key-of-g",
		Value: "",
		Usage: "Set the private key",
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
				Name: "info",
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
					&tonConfig,
					&botWalletSeed,
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
					&tonConfig,
					&preUpdateReserve,
					&output,
				},
				Action: func(c *cli2.Context) error {
					if err := utils.SetupLogger(c.String("loglevel")); err != nil {
						return err
					}
					return detect(c)
				},
			},

			{
				Name:        "mem-pool",
				Description: "to check tonapi mempool",
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
					return memPoolCheck(c)
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
					return botBundle(c)
				},
			},

			{
				Name:        "bot-dedust-sell",
				Description: "to sell some ton from bot",
				Flags: []cli2.Flag{
					&tonConfig,
					&botWalletSeed,
					&jettonMasterAddr,
					&vaultAddr,
					&poolAddr,
				},
				Action: func(c *cli2.Context) error {
					if err := utils.SetupLogger(c.String("loglevel")); err != nil {
						return err
					}
					return botDedustSell(c)
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
			{
				Name:        "new-seed-same-shard",
				Description: "to generate a new seed for a wallet",
				Flags: []cli2.Flag{
					&destAddr,
				},
				Action: func(c *cli2.Context) error {
					if err := utils.SetupLogger(c.String("loglevel")); err != nil {
						return err
					}

					a, err := address.ParseAddr(c.String("dest-addr"))
					if err != nil {
						return err
					}

					return NewSeedSameShardWith(a)
				},
			},
			{
				Name: "printer",
				Flags: []cli2.Flag{
					&tonConfig,
					&botWalletSeed,
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
			},
			{
				Name: "liteserver-ips",
				Flags: []cli2.Flag{
					&tonConfig,
				},
				Description: "to get list all ip of ton-config",
				Action: func(c *cli2.Context) error {
					return LiteserverIps(c)
				},
			},
			{
				Name: "dummy-transfer",
				Flags: []cli2.Flag{
					&tonConfig,
					&botWalletSeed,
					&destAddr,
					&amount,
				},
				Description: "generate base64 send message(external)",
				Action: func(c *cli2.Context) error {
					return buildTransferMessage(c)
				},
			},
			{
				Name: "generate-g",
				Flags: []cli2.Flag{
					&host,
					&port,
					&user,
					&password,
					&database,
					&botWalletSeed,
				},
				Description: "generate G",
				Action: func(c *cli2.Context) error {
					return GenGForPools(c)
				},
			},
			{
				Name: "collect-g",
				Flags: []cli2.Flag{
					&host,
					&port,
					&user,
					&password,
					&database,
					&botWalletSeed,
					&privateKeyOfG,
					&tonConfig,
				},
				Description: "collect G",
				Action: func(c *cli2.Context) error {
					return botCollectG(c)
				},
			},
			{
				Name: "collect-g-auto",
				Flags: []cli2.Flag{
					&host,
					&port,
					&user,
					&password,
					&database,
					&botWalletSeed,
					&tonConfig,
				},
				Description: "collect G auto",
				Action: func(c *cli2.Context) error {
					return botCollectGAuto(c)
				},
			},
			{
				Name: "jetton-transfer-from-g",
				Flags: []cli2.Flag{
					&amount,
					&destAddr,
					&botWalletSeed,
					&tonConfig,
					&privateKeyOfG,
					&jettonWalletAddr,
				},
				Description: "collect G auto",
				Action: func(c *cli2.Context) error {
					return jettonTransferFromG(c)
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
