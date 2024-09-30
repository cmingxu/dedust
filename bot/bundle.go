package bot

import (
	"crypto/ed25519"
	"fmt"

	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/tlb"
	"github.com/xssnick/tonutils-go/ton"
	"golang.org/x/net/context"
)

func Bundle(
	ctx context.Context,
	client ton.APIClientWrapped,
	botprivateKey ed25519.PrivateKey,
	poolAddr *address.Address,
	tonIn tlb.Coins,
	limit tlb.Coins,
) error {
	botAddr := botAddress(botprivateKey.Public().(ed25519.PublicKey))

	fmt.Println("Bot address:", botAddr.String())

	botWallet := NewBotWallet(ctx, client, botAddr, botprivateKey, 6)
	nextLimit := tlb.MustFromTON("0.00001").Nano()
	msg := botWallet.BuildBundle(poolAddr, tonIn.Nano(), limit.Nano(), nextLimit)

	return botWallet.Send(ctx, msg, true)
}
