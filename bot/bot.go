package bot

import (
	"crypto/ed25519"

	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/tlb"
	"github.com/xssnick/tonutils-go/ton"
	"github.com/xssnick/tonutils-go/ton/wallet"
	"github.com/xssnick/tonutils-go/tvm/cell"
	"golang.org/x/net/context"
)

var (
	MessageTTL = 60 * 3
)

type BotWallet struct {
	addr *address.Address
	api  ton.APIClientWrapped
	pk   ed25519.PrivateKey

	seq uint64
}

func NewBotWallet(
	ctx context.Context,
	client ton.APIClientWrapped,
	addr *address.Address,
	botprivateKey ed25519.PrivateKey,
	seq uint64,
) *BotWallet {
	return &BotWallet{
		addr: addr,
		api:  client,
		pk:   botprivateKey,
		seq:  seq,
	}
}

func (b *BotWallet) Info(ctx context.Context) error {
	return nil
}

func (b *BotWallet) TransferNoBounce(ctx context.Context, to *address.Address,
	amount tlb.Coins, comment string, wait bool) error {

	return b.transfer(ctx, to, amount, comment, false, wait)
}

func (b *BotWallet) Transfer(ctx context.Context, to *address.Address,
	amount tlb.Coins, comment string, wait bool) error {

	return b.transfer(ctx, to, amount, comment, true, wait)
}

func (w *BotWallet) transfer(ctx context.Context, to *address.Address, amount tlb.Coins, comment string, bounce bool, waitConfirmation ...bool) (err error) {
	transfer, err := w.BuildTransfer(to, amount, bounce, comment)
	if err != nil {
		return err
	}
	return w.Send(ctx, transfer, waitConfirmation...)
}

func (w *BotWallet) BuildTransfer(to *address.Address, amount tlb.Coins, bounce bool, comment string) (_ *wallet.Message, err error) {
	var body *cell.Cell
	if comment != "" {
		body, err = wallet.CreateCommentCell(comment)
		if err != nil {
			return nil, err
		}
	}

	return &wallet.Message{
		Mode: wallet.PayGasSeparately + wallet.IgnoreErrors,
		InternalMessage: &tlb.InternalMessage{
			IHRDisabled: true,
			Bounce:      bounce,
			DstAddr:     to,
			Amount:      amount,
			Body:        body,
		},
	}, nil
}
