package bot

import (
	"context"
	"crypto/ed25519"
	"fmt"

	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/tlb"
	"github.com/xssnick/tonutils-go/ton"
)

func WalletInfo(ctx context.Context,
	client ton.APIClientWrapped,
	privateKey ed25519.PrivateKey,
	botType BotType) error {
	addr := WalletAddress(privateKey.Public().(ed25519.PublicKey), nil, botType)
	fmt.Println("Wallet address:", addr.String())
	fmt.Println("Wallet Type: ", botType)

	masterBlock, err := client.GetMasterchainInfo(ctx)
	if err != nil {
		return err
	}

	fmt.Println("Masterchain last block:", masterBlock.SeqNo)
	account, err := client.WaitForBlock(masterBlock.SeqNo).GetAccount(ctx, masterBlock, addr)
	if err != nil {
		return err
	}

	fmt.Println("State:", account.IsActive)
	fmt.Println("Balance:", account.State.Balance)

	seqno, err := getSeqno(ctx, client, masterBlock, addr)
	if err != nil {
		return err
	}
	fmt.Println("Seqno:", seqno)

	return nil
}

func WalletAddress(publicKey ed25519.PublicKey,
	botAddr *address.Address,
	botType BotType) *address.Address {
	if botType == V4R2 {
		return v4Address(publicKey)
	} else if botType == Bot {
		return botAddress(publicKey)
	} else if botType == G {
		return gAddress(publicKey, botAddr)
	} else {
		panic("unknown bot type")
	}
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

func v4Address(publicKey ed25519.PublicKey) *address.Address {
	stateInit := &tlb.StateInit{
		Code: getV4Code(),
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
