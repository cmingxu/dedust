package cli

import (
	"crypto/ed25519"
	"crypto/hmac"
	"crypto/sha512"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/cmingxu/dedust/bot"
	"github.com/cmingxu/dedust/utils"
	_ "github.com/go-sql-driver/mysql"
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
	var (
		err             error
		mainWalletSeeds []string
		botWalletSeeds  []string
	)
	if mainWalletSeeds, err = loadSeeds(c.String("main-wallet-seed")); err != nil {
		return err
	}

	if len(mainWalletSeeds) != 24 {
		return errors.New("main wallet seeds are required")
	}

	if botWalletSeeds, err = loadSeeds(c.String("bot-wallet-seed")); err != nil {
		return err
	}

	if len(botWalletSeeds) != 24 {
		return errors.New("new wallet seeds are required")
	}

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

	// calculate new PK for new wallet
	mac := hmac.New(sha512.New, []byte(strings.Join(botWalletSeeds, " ")))
	mac.Write([]byte(""))
	hash := mac.Sum(nil)

	p := pbkdf2.Key(hash, []byte(_BasicSalt), _Iterations/256, 1, sha512.New)
	if p[0] != 0 {
		return errors.New("invalid new wallet seed")
	}
	pk := ed25519.NewKeyFromSeed(pbkdf2.Key(hash, []byte(_Salt), _Iterations, 32, sha512.New))

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
		err            error
		botWalletSeeds []string
	)
	if botWalletSeeds, err = loadSeeds(c.String("bot-wallet-seed")); err != nil {
		return err
	}

	if len(botWalletSeeds) != 24 {
		return errors.New("new wallet seeds are required")
	}

	// establish connection to the server
	pool, ctx, err := utils.GetConnectionPool(c.String("ton-config"))
	if err != nil {
		return err
	}
	client := utils.GetAPIClientWithTimeout(pool, time.Second*10)

	// calculate new PK for new wallet
	mac := hmac.New(sha512.New, []byte(strings.Join(botWalletSeeds, " ")))
	mac.Write([]byte(""))
	hash := mac.Sum(nil)

	p := pbkdf2.Key(hash, []byte(_BasicSalt), _Iterations/256, 1, sha512.New)
	if p[0] != 0 {
		return errors.New("invalid new wallet seed")
	}
	pk := ed25519.NewKeyFromSeed(pbkdf2.Key(hash, []byte(_Salt), _Iterations, 32, sha512.New))
	return bot.InfoBot(ctx, client, pk)
}

// tonupBot ton up bot
func tonupBot(c *cli2.Context) error {
	var (
		err             error
		mainWalletSeeds []string
		botWalletSeeds  []string
	)
	if mainWalletSeeds, err = loadSeeds(c.String("main-wallet-seed")); err != nil {
		return err
	}

	if len(mainWalletSeeds) != 24 {
		return errors.New("main wallet seeds are required")
	}

	if botWalletSeeds, err = loadSeeds(c.String("bot-wallet-seed")); err != nil {
		return err
	}

	if len(botWalletSeeds) != 24 {
		return errors.New("new wallet seeds are required")
	}

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

	// calculate new PK for new wallet
	mac := hmac.New(sha512.New, []byte(strings.Join(botWalletSeeds, " ")))
	mac.Write([]byte(""))
	hash := mac.Sum(nil)

	p := pbkdf2.Key(hash, []byte(_BasicSalt), _Iterations/256, 1, sha512.New)
	if p[0] != 0 {
		return errors.New("invalid new wallet seed")
	}
	pk := ed25519.NewKeyFromSeed(pbkdf2.Key(hash, []byte(_Salt), _Iterations, 32, sha512.New))

	amount, err := tlb.FromTON(c.String("amount"))
	if err != nil {
		return err
	}

	return bot.Tonup(
		ctx,
		client,
		mainWallet,
		pk,
		amount,
	)
}

func botTransfer(c *cli2.Context) error {
	var (
		err            error
		botWalletSeeds []string
	)
	if botWalletSeeds, err = loadSeeds(c.String("bot-wallet-seed")); err != nil {
		return err
	}

	if len(botWalletSeeds) != 24 {
		return errors.New("new wallet seeds are required")
	}

	// establish connection to the server
	pool, ctx, err := utils.GetConnectionPool(c.String("ton-config"))
	if err != nil {
		return err
	}
	client := utils.GetAPIClientWithTimeout(pool, time.Second*10)

	// calculate new PK for new wallet
	mac := hmac.New(sha512.New, []byte(strings.Join(botWalletSeeds, " ")))
	mac.Write([]byte(""))
	hash := mac.Sum(nil)

	p := pbkdf2.Key(hash, []byte(_BasicSalt), _Iterations/256, 1, sha512.New)
	if p[0] != 0 {
		return errors.New("invalid new wallet seed")
	}
	pk := ed25519.NewKeyFromSeed(pbkdf2.Key(hash, []byte(_Salt), _Iterations, 32, sha512.New))

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
		pk,
		destAddr,
		amount,
	)
}

func loadSeeds(seedsOrSeedsFile string) ([]string, error) {
	seeds := []string{}

	if len(seedsOrSeedsFile) == 0 {
		return seeds, errors.New("seeds or seeds file is required")
	}

	seeds = strings.SplitN(seedsOrSeedsFile, " ", -1)
	if len(seeds) == 24 {
		return seeds, nil
	}

	if _, err := os.Stat(seedsOrSeedsFile); err == nil {
		seedsFromFile, err := os.ReadFile(seedsOrSeedsFile)
		if err != nil {
			return []string{}, err
		}

		return strings.Split(string(strings.Trim(string(seedsFromFile), "\n")), " "), nil
	}

	return []string{}, nil
}
