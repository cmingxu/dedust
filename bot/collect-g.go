package bot

import (
	"crypto/ed25519"
	"fmt"

	"github.com/xssnick/tonutils-go/tlb"
	"github.com/xssnick/tonutils-go/ton"
	"github.com/xssnick/tonutils-go/ton/wallet"
	"github.com/xssnick/tonutils-go/tvm/cell"
	"golang.org/x/net/context"
)

func CollectG(
	ctx context.Context,
	client ton.APIClientWrapped,
	botPk ed25519.PrivateKey,
	gPk ed25519.PrivateKey,
) error {
	botAddr := botAddress(botPk.Public().(ed25519.PublicKey))
	gAddr := gAddress(gPk.Public().(ed25519.PublicKey), botAddr)

	fmt.Println("Bot address:", botAddr.String())
	fmt.Println("G address:", gAddr.String())

	masterBlock, err := client.GetMasterchainInfo(ctx)
	if err != nil {
		return err
	}

	seqno, err := GetSeqno(ctx, client, masterBlock, botAddr)
	if err != nil {
		return err
	}

	fmt.Println("G seqno:", seqno)

	botWallet := NewWallet(ctx, client, Bot, botPk, nil, seqno)
	msgBody := cell.BeginCell().
		MustStoreUInt(0x474f86cd, 32).
		EndCell()

	msg := &wallet.Message{
		Mode: wallet.PayGasSeparately + wallet.IgnoreErrors,
		InternalMessage: &tlb.InternalMessage{
			IHRDisabled: true,
			Bounce:      false,
			DstAddr:     gAddr,
			Amount:      tlb.MustFromTON("0.1"),
			Body:        msgBody,
		},
	}

	return botWallet.Send(ctx, 0, msg, true)
}
