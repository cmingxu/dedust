package cli

import (
	"crypto/ed25519"
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/cmingxu/dedust/bot"
	"github.com/cmingxu/dedust/utils"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	cli2 "github.com/urfave/cli/v2"
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/tlb"
	"github.com/xssnick/tonutils-go/ton/wallet"
	"golang.org/x/crypto/pbkdf2"
)

const (
	_Iterations   = 100000
	_Salt         = "TON default seed"
	_BasicSalt    = "TON seed version"
	_PasswordSalt = "TON fast seed version"
)

func deployBot(c *cli2.Context) error {
	var err error

	mainWalletSeeds := MustLoadSeeds(c.String("main-wallet-seed"))
	botWalletSeeds := MustLoadSeeds(c.String("bot-wallet-seed"))

	// establish connection to the server
	pool, ctx, err := utils.GetConnectionPool(c.String("ton-config"))
	if err != nil {
		return err
	}
	client := utils.GetAPIClientWithTimeout(pool, time.Second*10)

	// initialize main wallet
	mainWallet, err := wallet.FromSeed(client, mainWalletSeeds, wallet.V4R2)
	if err != nil {
		return err
	}

	pk := pkFromSeed(botWalletSeeds)
	fmt.Println("Main wallet address:", mainWallet.Address().String())
	fmt.Println("Bot wallet public key:", pk.Public().(ed25519.PublicKey))

	addr, err := bot.DeployBot(ctx, mainWallet, pk)
	if err != nil {
		return err
	}

	fmt.Println("Bot wallet address:", addr.String())

	return nil
}

// infoBot prints bot info
func infoBot(c *cli2.Context) error {
	var (
		err error
	)
	botWalletSeeds := MustLoadSeeds(c.String("bot-wallet-seed"))

	// establish connection to the server
	pool, ctx, err := utils.GetConnectionPool(c.String("ton-config"))
	if err != nil {
		return err
	}
	client := utils.GetAPIClientWithTimeout(pool, time.Second*30)

	return bot.InfoBot(ctx, client, pkFromSeed(botWalletSeeds))
}

// tonupBot ton up bot
func tonupBot(c *cli2.Context) error {
	var (
		err error
	)
	mainWalletSeeds := MustLoadSeeds(c.String("main-wallet-seed"))
	botWalletSeeds := MustLoadSeeds(c.String("bot-wallet-seed"))

	// establish connection to the server
	pool, ctx, err := utils.GetConnectionPool(c.String("ton-config"))
	if err != nil {
		return err
	}
	client := utils.GetAPIClientWithTimeout(pool, time.Second*10)

	// initialize main wallet
	mainWallet, err := wallet.FromSeed(client, mainWalletSeeds, wallet.V4R2)
	if err != nil {
		return err
	}

	amount, err := tlb.FromTON(c.String("amount"))
	if err != nil {
		return err
	}

	return bot.Tonup(
		ctx,
		client,
		mainWallet,
		pkFromSeed(botWalletSeeds),
		amount,
	)
}

func botTransfer(c *cli2.Context) error {
	var (
		err error
	)
	botWalletSeeds := MustLoadSeeds(c.String("bot-wallet-seed"))

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

	return bot.Transfer(
		ctx,
		client,
		pkFromSeed(botWalletSeeds),
		destAddr,
		amount,
	)
}

func botBundle(c *cli2.Context) error {
	var (
		err error
	)
	botWalletSeeds := MustLoadSeeds(c.String("bot-wallet-seed"))

	// establish connection to the server
	connPool, ctx, err := utils.GetConnectionPool(c.String("ton-config"))
	if err != nil {
		return err
	}
	client := utils.GetAPIClientWithTimeout(connPool, time.Second*10)

	amount, err := tlb.FromTON(c.String("amount"))
	if err != nil {
		return err
	}

	poolAddr, err := address.ParseAddr(c.String("pool-addr"))
	if err != nil {
		return err
	}

	db, err := sqlx.Connect("mysql", utils.ConstructDSN(c))
	if err != nil {
		return err
	}
	defer db.Close()

	return bot.Bundle(
		ctx,
		connPool,
		client,
		pkFromSeed(botWalletSeeds),
		poolAddr,
		amount,
		tlb.MustFromTON("0.00000001"),
		db,
	)
}

func botDedustSell(c *cli2.Context) error {
	var (
		err error
	)
	botWalletSeeds := MustLoadSeeds(c.String("bot-wallet-seed"))

	// establish connection to the server
	connPool, ctx, err := utils.GetConnectionPool(c.String("ton-config"))
	if err != nil {
		return err
	}
	client := utils.GetAPIClientWithTimeout(connPool, time.Second*30)

	jettonMaterAddr, err := address.ParseAddr(c.String("jetton-master-addr"))
	if err != nil {
		return err
	}

	vaultAddr, err := address.ParseAddr(c.String("vault-addr"))
	if err != nil {
		return err
	}

	poolAddr, err := address.ParseAddr(c.String("pool-addr"))
	if err != nil {
		return err
	}

	return bot.DedustSell(
		ctx,
		client,
		pkFromSeed(botWalletSeeds),
		jettonMaterAddr,
		vaultAddr,
		poolAddr,
	)
}

func botCollectG(c *cli2.Context) error {
	var (
		err error
	)
	botWalletSeeds := MustLoadSeeds(c.String("bot-wallet-seed"))

	gPKStr := c.String("private-key-of-g")
	if len(gPKStr) == 0 {
		return fmt.Errorf("private-key-of-g is required")
	}

	gpkRaw, err := hex.DecodeString(gPKStr)
	if err != nil {
		return err
	}

	gpk := ed25519.PrivateKey(gpkRaw)

	// establish connection to the server
	connPool, ctx, err := utils.GetConnectionPool(c.String("ton-config"))
	if err != nil {
		return err
	}
	client := utils.GetAPIClientWithTimeout(connPool, time.Second*30)

	return bot.CollectG(
		ctx,
		client,
		pkFromSeed(botWalletSeeds),
		gpk,
	)
}

func MustLoadSeeds(seedsOrSeedsFile string) []string {
	seeds := []string{}

	if len(seedsOrSeedsFile) == 0 {
		panic("seeds are required")
	}

	if _, err := os.Stat(seedsOrSeedsFile); err == nil {
		seedsFromFile, err := os.ReadFile(seedsOrSeedsFile)
		if err != nil {
			panic("failed to read seeds file")
		}

		return strings.Split(string(strings.Trim(string(seedsFromFile), "\n")), " ")
	} else {
		seeds = strings.SplitN(seedsOrSeedsFile, " ", -1)
	}

	if len(seeds) != 24 {
		panic("invalid seeds")
	}

	return seeds
}

func pkFromSeed(seeds []string) ed25519.PrivateKey {
	// calculate new PK for new wallet
	mac := hmac.New(sha512.New, []byte(strings.Join(seeds, " ")))
	mac.Write([]byte(""))
	hash := mac.Sum(nil)

	p := pbkdf2.Key(hash, []byte(_BasicSalt), _Iterations/256, 1, sha512.New)
	if p[0] != 0 {
		panic("invalid new wallet seed")
	}
	return ed25519.NewKeyFromSeed(pbkdf2.Key(hash, []byte(_Salt), _Iterations, 32, sha512.New))

}
