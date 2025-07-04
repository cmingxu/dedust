package cli

import (
	cli2 "github.com/urfave/cli/v2"
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
		Value: "./config/global-config.json",
		Usage: "Set the TON config url path or local file path",
	}

	mainWalletSeed = cli2.StringFlag{
		Name:  "main-wallet-seed",
		Value: "./main-wallet-seed.txt",
		Usage: "Set the main wallet seed file path or seed itself",
	}

	walletSeed = cli2.StringFlag{
		Name:  "wallet-seed",
		Value: "./wallet-seed.txt",
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

	knownSmallestAddr = cli2.StringFlag{
		Name:  "known-smallest-addr",
		Value: "EQDQ1I0oC04dvmbewqUkt1n7zzxijS2hXmIv0FMv_JJZjQNI",
		Usage: "known smallest address",
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
		Value: true,
		Usage: "Set whether use ton api",
	}

	useTonAPIBlockchain = cli2.BoolFlag{
		Name:  "use-tonapi-blockchain",
		Value: true,
		Usage: "Set whether use ton api blockchain",
	}

	useTonCenter = cli2.BoolFlag{
		Name:  "use-toncenter",
		Value: true,
		Usage: "Set whether use ton center",
	}

	useTonCenterV3 = cli2.BoolFlag{
		Name:  "use-toncenter-v3",
		Value: true,
		Usage: "Set whether use ton center v3",
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

	floor = cli2.StringFlag{
		Name:  "floor",
		Value: "1",
		Usage: "Set the floor amount of TON to bundle",
	}

	privateKeyOfG = cli2.StringFlag{
		Name:  "private-key-of-g",
		Value: "",
		Usage: "Set the private key",
	}

	botType = cli2.StringFlag{
		Name:  "bot-type",
		Value: "v4",
		Usage: "Set the bot type, valid values are bot, v4, g",
	}

	terminator = cli2.StringFlag{
		Name:  "terminator",
		Value: "200",
		Usage: "Set the terminator",
	}

	enableTracing = cli2.BoolFlag{
		Name:  "enable-tracing",
		Value: false,
		Usage: "Set whether enable tracing",
	}

	walletMode = cli2.IntFlag{
		Name:  "wallet-mode",
		Value: 5,
		Usage: "Set the wallet mode",
	}

	tonapiIp = cli2.StringFlag{
		Name:  "tonapi-ip",
		Value: "",
		Usage: "Set the tonapi ip, instead of using domain",
	}

	skippedPool = cli2.StringFlag{
		Name:  "skipped-pool",
		Value: "EQAjNbON8Bw_1H56jc8YMtM-jd1VS94nFczAWE6U85Hj6hDk,EQDvBFeKDjBqz3DAc5wbHsMs1IFfQVKYYGH8Du2MbAkqXtx7",
		Usage: "Set the skipped pool",
	}

	skippedVictims = cli2.StringFlag{
		Name:  "skipped-victims",
		Value: "EQDa6ts3gUF9CquZdVGLGkdOY4mluPmP2k_T2ROOCUZiiNNA,EQARVRr--lCMqduE458pl_U93AzyRcxusxMS9KYBqT-bCU41",
		Usage: "Set the skipped victims",
	}

	generateG = cli2.BoolFlag{
		Name:  "generate-g",
		Value: false,
		Usage: "Set whether generate G",
	}
)
