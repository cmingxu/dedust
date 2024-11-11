package bot

import (
	"crypto/ed25519"
	"encoding/hex"
	"fmt"
	"os"

	"github.com/cmingxu/dedust/model"
	"github.com/jmoiron/sqlx"
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/liteclient"
	"github.com/xssnick/tonutils-go/tlb"
	"github.com/xssnick/tonutils-go/ton"
	"github.com/xssnick/tonutils-go/ton/wallet"
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
	db *sqlx.DB,
) error {
	botAddr := botAddress(botprivateKey.Public().(ed25519.PublicKey))
	fmt.Println("Bot address:", botAddr.String())

	masterBlock, err := client.GetMasterchainInfo(ctx)
	if err != nil {
		return err
	}

	seqno, err := GetSeqno(ctx, client, masterBlock, botAddr)
	if err != nil {
		return err
	}

	botWallet := NewWallet(ctx, client, Bot, botprivateKey, nil, seqno)
	var pk ed25519.PrivateKey

	if os.Getenv("PKOFG") != "" {
		pkRaw := os.Getenv("PKOFG")
		pk, _ = hex.DecodeString(pkRaw)
	} else {
		var p model.Pool
		err := db.Get(&p, "SELECT * FROM pools WHERE address = ?", poolAddr.String())
		if err != nil {
			return err
		}

		pk, err = hex.DecodeString(p.PrivateKeyOfG.String)
		if err != nil {
			return err
		}
	}

	fmt.Println("PK: ", hex.EncodeToString(pk))
	deployGMsg, gAddr, err := BuildG(botAddr, pk, tlb.MustFromTON("0.3"))
	if err != nil {
		return err
	}

	nextLimit := tonIn
	msg := botWallet.BuildBundle(poolAddr, tonIn.Nano(), limit.Nano(),
		nextLimit.Nano(), 0, gAddr)

	fmt.Println("G address:", gAddr.String())
	fmt.Println("G deploy message:", deployGMsg)

	fmt.Println("Bundle message:", msg)

	return botWallet.SendMany(ctx, 5, []*wallet.Message{deployGMsg, msg}, false)
}
