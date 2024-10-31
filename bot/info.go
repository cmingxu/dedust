package bot

import (
	"context"
	"crypto/ed25519"
	"fmt"

	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/tlb"
	"github.com/xssnick/tonutils-go/ton"
)

func InfoBot(ctx context.Context,
	client ton.APIClientWrapped,
	botprivateKey ed25519.PrivateKey) error {

	botAddr := botAddress(botprivateKey.Public().(ed25519.PublicKey))
	fmt.Println("Bot address:", botAddr.String())

	masterBlock, err := client.GetMasterchainInfo(ctx)
	if err != nil {
		return err
	}

	fmt.Println("Masterchain last block:", masterBlock.SeqNo)
	account, err := client.WaitForBlock(masterBlock.SeqNo).GetAccount(ctx, masterBlock, botAddr)
	if err != nil {
		return err
	}

	fmt.Println("Bot state:", account.IsActive)
	fmt.Println("Bot balance:", account.State.Balance)

	seqno, err := getSeqno(ctx, client, masterBlock, botAddr)
	if err != nil {
		return err
	}
	fmt.Println("Bot seqno:", seqno)

	return nil
}

func BotAddress(publicKey ed25519.PublicKey) *address.Address {
	return botAddress(publicKey)
}

func GAddress(publicKey ed25519.PublicKey, botAddr *address.Address) *address.Address {
	return gAddress(publicKey, botAddr)
}

func gAddress(publicKey ed25519.PublicKey, botAddr *address.Address) *address.Address {
	stateInit := &tlb.StateInit{
		Code: getGCode(),
		Data: getGData(publicKey, botAddr),
	}

	stateCell, err := tlb.ToCell(stateInit)
	if err != nil {
		panic(err)
	}

	return address.NewAddress(0, 0, stateCell.Hash())
}

func botAddress(publicKey ed25519.PublicKey) *address.Address {
	stateInit := &tlb.StateInit{
		Code: getCode(),
		Data: getData(publicKey),
	}

	stateCell, err := tlb.ToCell(stateInit)
	if err != nil {
		panic(err)
	}

	return address.NewAddress(0, 0, stateCell.Hash())
}

func GetSeqno(ctx context.Context,
	client ton.APIClientWrapped,
	masterBlock *ton.BlockIDExt,
	addr *address.Address) (uint64, error) {
	return getSeqno(ctx, client, masterBlock, addr)
}

func getSeqno(ctx context.Context,
	client ton.APIClientWrapped,
	masterBlock *ton.BlockIDExt,
	addr *address.Address) (uint64, error) {

	stack, err := client.RunGetMethod(ctx, masterBlock, addr, "seqno")
	if err != nil {
		return 0, err
	}

	seqBI, err := stack.Int(0)
	if err != nil {
		return 0, err
	}

	return seqBI.Uint64(), nil
}
