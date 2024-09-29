package bot

import (
	"context"
	"crypto/ed25519"
	"encoding/base64"

	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/tlb"
	"github.com/xssnick/tonutils-go/ton/wallet"
	"github.com/xssnick/tonutils-go/tvm/cell"
)

// my lucky number
const SubwalletID = 39

// Bot contract is w4r2 based contracts with my features
func DeployBot(ctx context.Context,
	mainWallet *wallet.Wallet,
	botPrivateKey ed25519.PrivateKey) (*address.Address, error) {
	msgBody := cell.BeginCell().EndCell()

	addr, _, _, err := mainWallet.DeployContractWaitTransaction(context.Background(),
		tlb.MustFromTON("0.02"),
		msgBody,
		getCode(),
		getData(botPrivateKey.Public().(ed25519.PublicKey)))
	if err != nil {
		return nil, err
	}

	return addr, nil
}

func getCode() *cell.Cell {
	boc := `
te6ccgECFAEAAs4AART/APSkE/S88sgLAQIBIAIDAgFIBAUE+PKDCNcYINMf0x/THwL4I7vyZO1E0NMf0x/T//QE0VFDuvKhUVG68qIF+QFUEGT5EPKj+AAkpMjLH1JAyx9SMMv/UhD0AMntVPgPAdMHIcAAn2xRkyDXSpbTB9QC+wDoMOAhwAHjACHAAuMAAcADkTDjDQOkyMsfEssfy/8QERITAubQAdDTAyFxsJJfBOAi10nBIJJfBOAC0x8hghBwbHVnvSKCEGRzdHK9sJJfBeAD+kAwIPpEAcjKB8v/ydDtRNCBAUDXIfQEMFyBAQj0Cm+hMbOSXwfgBdM/yCWCEHBsdWe6kjgw4w0DghBkc3RyupJfBuMNBgcCASAICQByAfoA9AQw+CdvIjBQCqEhvvLgUIIQ8Gx1Z3CAGFAEywUmzxZY+gIZ9ADLaRfLH1Jgyz8gyYBA+wAGAIRQBIEBCPRZMO1E0IEBQNcgyAHPFvQAye1UAXKwjiCCEORzdHJwgBhQBcsFUAPPFiP6AhPLassfyz/JgED7AJJfA+ICASAKCwBZvSQrb2omhAgKBrkPoCGEcNQICEekk30pkQzmkD6f+YN4EoAbeBAUiYcVnzGEAgFYDA0AEbjJftRNDXCx+AA9sp37UTQgQFA1yH0BDACyMoHy//J0AGBAQj0Cm+hMYAIBIA4PABmtznaiaEAga5Drhf/AABmvHfaiaEAQa5DrhY/AAG7SB/oA1NQi+QAFyMoHFcv/ydB3dIAYyMsFywIizxZQBfoCFMtrEszMyXP7AMhAFIEBCPRR8qcCAHCBAQjXGPoA0z/IVCBHgQEI9FHyp4IQbm90ZXB0gBjIywXLAlAGzxZQBPoCFMtqEssfyz/Jc/sAAgBsgQEI1xj6ANM/MFIkgQEI9Fnyp4IQZHN0cnB0gBjIywXLAlAFzxZQA/oCE8tqyx8Syz/Jc/sAAAr0AMntVA==
`

	bytes, err := base64.StdEncoding.DecodeString(boc)
	if err != nil {
		panic(err)
	}

	code, err := cell.FromBOC(bytes)
	if err != nil {
		panic(err)
	}

	return code
}

func getData(publicKey ed25519.PublicKey) *cell.Cell {
	dataCell := cell.BeginCell().
		MustStoreUInt(0, 32).           // Seqno
		MustStoreUInt(SubwalletID, 32). // Subwallet ID
		MustStoreSlice(publicKey, 256). // Public Key
		MustStoreDict(nil).
		EndCell()

	return dataCell
}
