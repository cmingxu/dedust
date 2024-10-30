package cli

import (
	"crypto/ed25519"
	"encoding/hex"
	"fmt"

	"github.com/cmingxu/dedust/bot"
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/ton/wallet"
)

func NewSeedSameShardWith(addr *address.Address) error {
	for {
		seeds := wallet.NewSeed()
		pk := pkFromSeed(seeds)
		newBotAddr := bot.BotAddress(pk.Public().(ed25519.PublicKey))
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
