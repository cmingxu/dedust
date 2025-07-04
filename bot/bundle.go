package bot

import (
	"bytes"
	"crypto/ed25519"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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

	op := int64(0)
	nextLimit := tonIn
	msg := botWallet.BuildBundle(poolAddr, tonIn.Nano(), limit.Nano(),
		nextLimit.Nano(), 0, gAddr, op)

	fmt.Println("G address:", gAddr.String())
	fmt.Println("G deploy message:", deployGMsg)
	fmt.Println("Bundle message:", msg)

	addr := address.MustParseAddr("UQCwSxqefElovEPlpZ8bIEL_KXqWuqoOhwb65uYjos9bCDcM")
	amount := tlb.MustFromTON("0.00000001")
	comment, _ := botWallet.BuildTransfer(addr, amount, true, "c")

	fmt.Println("Comment message:", comment)
	//return botWallet.SendMany(ctx, op, []*wallet.Message{deployGMsg, msg, comment}, false)

	externalMsg, err := botWallet.BuildExternalMessageForMany(context.Background(),
		0,
		[]*wallet.Message{deployGMsg, msg, comment})
	if err != nil {
		return err
	}
	cell, err := tlb.ToCell(externalMsg)
	if err != nil {
		return err
	}
	var body struct {
		Boc string `json:"boc"`
	}
	body.Boc = base64.StdEncoding.EncodeToString(cell.ToBOC())

	buf := bytes.NewBuffer(nil)
	if err := json.NewEncoder(buf).Encode(body); err != nil {
		return err
	}

	req, err := http.NewRequest("POST", "https://tonapi.io/v2/blockchain/message", buf)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	const TOKEN = "AEETAB4AU6BMELIAAAADMMZHBQOIVYFMRL7QZ77HCXATNHS5PF6CIJQJNAQRLC4OG73V2VQ"
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", TOKEN))
	resp, err := http.DefaultClient.Do(req)
	io.Copy(os.Stdout, resp.Body)
	return err
}
