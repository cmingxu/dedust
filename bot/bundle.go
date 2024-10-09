package bot

import (
	"crypto/ed25519"
	"fmt"

	"github.com/rs/zerolog/log"
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/liteclient"
	"github.com/xssnick/tonutils-go/tlb"
	"github.com/xssnick/tonutils-go/ton"
	"golang.org/x/net/context"
)

func Bundle(
	ctx context.Context,
	pool *liteclient.ConnectionPool,
	client ton.APIClientWrapped,
	botprivateKey ed25519.PrivateKey,
	poolAddr *address.Address,
	tonIn tlb.Coins,
	limit tlb.Coins,
) error {
	botAddr := botAddress(botprivateKey.Public().(ed25519.PublicKey))

	fmt.Println("Bot address:", botAddr.String())

	botWallet := NewBotWallet(ctx, client, botAddr, botprivateKey, 277)

	nextLimit := tonIn
	msg := botWallet.BuildBundle(poolAddr, tonIn.Nano(), limit.Nano(), nextLimit.Nano())

	// botWallet.Send(ctx, msg, false)

	i := 0
	for i <= 5 {
		ctx := context.WithValue(context.Background(), "foo", struct{}{})
		nodeCtx, err := pool.StickyContextNextNodeBalanced(ctx)
		if err != nil {
			log.Error().Err(err).Msg("failed to get next node")
			break
		}

		err = botWallet.Send(nodeCtx, msg, false)
		if err != nil {
			log.Error().Err(err).Msg("failed to send bundle")
		}

		nodeId, _ := nodeCtx.Value("_ton_node_sticky").(uint32)
		log.Debug().Msgf("sent bundle to node %d[%d]", nodeId, i)
		i++
	}

	return nil
}
