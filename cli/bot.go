package cli

import (
	"crypto/ed25519"
	"crypto/hmac"
	"crypto/sha512"
	"os"
	"strings"

	"github.com/cmingxu/dedust/bot"
	_ "github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/pbkdf2"
)

const (
	_Iterations   = 100000
	_Salt         = "TON default seed"
	_BasicSalt    = "TON seed version"
	_PasswordSalt = "TON fast seed version"
)

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

func mustLoadBotType(botType string) bot.BotType {
	switch botType {
	case "g":
		return bot.G
	case "bot":
		return bot.Bot
	case "v4":
		return bot.V4R2
	default:
		panic("invalid bot type")
	}
}
