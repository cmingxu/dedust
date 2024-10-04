package bot

import (
	"crypto/ed25519"
	"math/big"

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

var (
	// dedust native swap gas
	GasAmount = tlb.MustFromTON("0.1").Nano()

	// dedust native vault address
	DedustNativeVault = address.MustParseAddr("EQDa4VOnTYlLvDJ0gZjNYm5PXfSmmtL6Vs6A_CZEtXCNICq_")

	DedustNativeSwapMaigc = uint64(0xea06185d)
	DedustJettonSwapMagic = uint64(0xe3a0d482)
)

var (
	JettonTransferMagic = uint64(0xf8a7ea5)
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

// Transfer
// https://github.com/dedust-io/sdk/blob/main/src/contracts/jettons/JettonWallet.ts
func (w *BotWallet) BuildDedustSell(
	to *address.Address, // bot jetton wallet
	otherMainWallet *address.Address,
	poolAddr *address.Address,
	amount tlb.Coins,
	limitOfTon tlb.Coins,
) (_ *wallet.Message) {

	swapParams := cell.BeginCell().
		MustStoreUInt(0, 32).   // deadline
		MustStoreAddr(w.addr).  // receipent address
		MustStoreAddr(nil).     // referer address
		MustStoreMaybeRef(nil). // fulfillPayload
		MustStoreMaybeRef(nil). // rejectPayload
		EndCell()

	//https://github.com/dedust-io/sdk/blob/ecd9ee9a8ddbee74c7432f8755e4b3192932d5d5/src/contracts/dex/vault/VaultJetton.ts#L63
	swapRef := cell.BeginCell().
		MustStoreUInt(DedustJettonSwapMagic, 32).   // magic
		MustStoreAddr(poolAddr).                    // pool address
		MustStoreUInt(0, 1).                        // kind
		MustStoreCoins(limitOfTon.Nano().Uint64()). // ton out
		MustStoreMaybeRef(nil).
		MustStoreRef(swapParams).
		EndCell()

	body := cell.BeginCell().
		MustStoreUInt(JettonTransferMagic, 32).                 // transfer magic
		MustStoreUInt(0, 64).                                   // queryId
		MustStoreCoins(amount.Nano().Uint64()).                 // amount of jetton to transfer
		MustStoreAddr(otherMainWallet).                         // other main wallet address - not jetton wallet
		MustStoreAddr(w.addr).                                  // responseAddress
		MustStoreMaybeRef(nil).                                 // custom payload
		MustStoreCoins(tlb.MustFromTON("0.2").Nano().Uint64()). // foward amount - here is the swap fee 0.2
		MustStoreMaybeRef(swapRef).
		EndCell()

	return &wallet.Message{
		Mode: wallet.PayGasSeparately + wallet.IgnoreErrors,
		InternalMessage: &tlb.InternalMessage{
			IHRDisabled: true,
			Bounce:      true,
			DstAddr:     to,
			Amount:      tlb.MustFromTON("0.3"), //0.2 forward for swap, jetton transfer let's say 0.1
			Body:        body,
		},
	}
}

// https://github.com/dedust-io/sdk/blob/main/src/contracts/dex/vault/VaultNative.ts
func (w *BotWallet) BuildBundle(poolAddr *address.Address,
	amount *big.Int, limit *big.Int, nextLimit *big.Int) (_ *wallet.Message) {

	passingPoolAddr := cell.BeginCell().
		MustStoreAddr(poolAddr).
		EndCell()

	swapParamsRef := cell.BeginCell().
		MustStoreUInt(0, 32).               // deadline
		MustStoreAddr(w.addr).              // receipent address
		MustStoreAddr(nil).                 // referer address
		MustStoreMaybeRef(passingPoolAddr). // fulfillPayload
		MustStoreMaybeRef(passingPoolAddr). // rejectPayload
		EndCell()

		// sell imeidiately
	next := cell.BeginCell().
		MustStoreAddr(poolAddr). // next pool addr
		MustStoreUInt(0, 1).
		MustStoreCoins(nextLimit.Uint64()). // next limit
		MustStoreMaybeRef(nil).
		EndCell()

	body := cell.BeginCell().
		MustStoreUInt(DedustNativeSwapMaigc, 32). // magic
		MustStoreUInt(0, 64).                     // queryId
		MustStoreCoins(amount.Uint64()).          // amount
		MustStoreAddr(poolAddr).                  // poolAddr
		MustStoreUInt(0, 1).                      // Kind
		MustStoreCoins(limit.Uint64()).           // Fee
		MustStoreMaybeRef(next).
		MustStoreRef(swapParamsRef).
		EndCell()

	return &wallet.Message{
		Mode: wallet.PayGasSeparately + wallet.IgnoreErrors,
		InternalMessage: &tlb.InternalMessage{
			IHRDisabled: true,
			Bounce:      true,
			DstAddr:     DedustNativeVault,
			Amount:      tlb.FromNanoTON(new(big.Int).Add(amount, GasAmount)),
			Body:        body,
		},
	}
}
