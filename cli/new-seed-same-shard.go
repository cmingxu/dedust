package cli

import (
	"crypto/ed25519"
	"encoding/hex"
	"fmt"

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
	}
)

func NewSeedSameShardWith(addr *address.Address) error {
	for {
		seeds := wallet.NewSeed()
		pk := pkFromSeed(seeds)
		newBotAddr := bot.WalletAddress(pk.Public().(ed25519.PublicKey), nil, bot.Bot)
		if shardIdFromAddr(newBotAddr) == shardIdFromAddr(addr) {
			fmt.Println("Dest Hash: ", hex.EncodeToString(addr.Data()))
			fmt.Println("New Hash:", hex.EncodeToString(newBotAddr.Data()))
			fmt.Println("Address:", newBotAddr.String())
			fmt.Println("Seed:", seeds)
			return nil
		}
	}
}

func shardIdFromAddr(addr *address.Address) string {
	return hex.EncodeToString(addr.Data())[:2]
}
