package cli

import (
	"crypto/ed25519"
	"encoding/hex"
	"fmt"
	"math/big"
	"os"

	"github.com/cmingxu/dedust/bot"
	"github.com/cmingxu/dedust/utils"
	cli2 "github.com/urfave/cli/v2"
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/ton/wallet"
)

var (
	newSeedSameShardCmd = &cli2.Command{
		Name:        "new-seed-same-shard",
		Description: "to generate a new seed for a wallet",
		Flags: []cli2.Flag{
			&destAddr,
			&knownSmallestAddr,
		},
		Action: func(c *cli2.Context) error {
			if err := utils.SetupLogger(c.String("loglevel")); err != nil {
				return err
			}

			dest, err := address.ParseAddr(c.String("dest-addr"))
			if err != nil {
				return err
			}

			smallest, err := address.ParseAddr(c.String("known-smallest-addr"))
			if err != nil {
				return err
			}

			return NewSeedSameShardWith(dest, smallest)
		},
	}
)

func NewSeedSameShardWith(dest, smallest *address.Address) error {
	fmt.Println("smallest", smallest.String())
	fmt.Println("dest", dest.String())

	smallestValue := new(big.Int).SetBytes(smallest.Data())

	for {
		seeds := wallet.NewSeed()
		pk := pkFromSeed(seeds)
		newBotAddr := bot.WalletAddress(pk.Public().(ed25519.PublicKey), nil, bot.Bot)

		fmt.Print(".")
		os.Stdout.Sync()
		newValue := new(big.Int).SetBytes(newBotAddr.Data())
		// fmt.Println("New Value:", newValue.String())
		// fmt.Println("Smallest Value:", smallestValue.String())
		// fmt.Println(newValue.Cmp(smallestValue))

		if newValue.Cmp(smallestValue) < 0 {
			fmt.Println("3")
			fmt.Println(shardIdFromAddr(newBotAddr), shardIdFromAddr(dest))
			if shardIdFromAddr(newBotAddr) == shardIdFromAddr(dest) {
				fmt.Println("Dest Hash: ", hex.EncodeToString(dest.Data()))
				fmt.Println("New Hash:", hex.EncodeToString(newBotAddr.Data()))
				fmt.Println("Address:", newBotAddr.String())
				fmt.Println("Seed:", seeds)
				return nil
			}
		}
	}
}

func shardIdFromAddr(addr *address.Address) string {
	return hex.EncodeToString(addr.Data())[:1]
}
