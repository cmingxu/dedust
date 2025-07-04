package bot

import (
	"context"
	"crypto/ed25519"
	"encoding/base64"
	"encoding/hex"

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
	botPrivateKey ed25519.PrivateKey,
	botType BotType,
) (*address.Address, error) {
	msgBody := cell.BeginCell().EndCell()

	code := getCode()
	if botType == G {
		code = getGCode()
	}

	if botType == V4R2 {
		code = getV4Code()
	}

	addr, _, _, err := mainWallet.DeployContractWaitTransaction(context.Background(),
		tlb.MustFromTON("0.02"),
		msgBody,
		code,
		getData(botPrivateKey.Public().(ed25519.PublicKey)))
	if err != nil {
		return nil, err
	}

	return addr, nil
}

func getCode() *cell.Cell {
	// v1
	_ = `
te6ccgECEwEAAoQAART/APSkE/S88sgLAQIBIAIDAgFIBAUE+PKDCNcYINMf0x/THwL4I7vyZO1E0NMf0x/T//QE0VFDuvKhUVG68qIF+QFUEGT5EPKj+AAkpMjLH1JAyx9SMMv/UhD0AMntVPgPAdMHIcAAn2xRkyDXSpbTB9QC+wDoMOAhwAHjACHAAuMAAcADkTDjDQOkyMsfEssfy/8PEBESAWLQMtDTAwFxsJFb4CHXScEgkVvgAdMfMCCCEHNi0Jy9kVvgAfpAAoIQc2LQnLqRW+MNBgIBIAcIAOwB0x/6APpA1DDQ+kAwcFRwAMjLH/gozxbLAcsAywDJcCCCEOOg1ILIyx9QBM8WE8sAIvoCEssAzMlwghAPin6lyMsfFcs/UAP6AgHPFvgozxYSywCCCTEtAPoCAdDPFslxgBjIywVQA88WcPoCEstqzMmAQPsAAgEgCQoAWb0kK29qJoQICga5D6AhhHDUCAhHpJN9KZEM5pA+n/mDeBKAG3gQFImHFZ8xhAIBWAsMABG4yX7UTQ1wsfgAPbKd+1E0IEBQNch9AQwAsjKB8v/ydABgQEI9ApvoTGACASANDgAZrc52omhAIGuQ64X/wAAZrx32omhAEGuQ64WPwABu0gf6ANTUIvkABcjKBxXL/8nQd3SAGMjLBcsCIs8WUAX6AhTLaxLMzMlz+wDIQBSBAQj0UfKnAgBwgQEI1xj6ANM/yFQgR4EBCPRR8qeCEG5vdGVwdIAYyMsFywJQBs8WUAT6AhTLahLLH8s/yXP7AAIAbIEBCNcY+gDTPzBSJIEBCPRZ8qeCEGRzdHJwdIAYyMsFywJQBc8WUAP6AhPLassfEss/yXP7AAAK9ADJ7VQ=
`

	_ = `
te6ccgECEwEAAoUAART/APSkE/S88sgLAQIBIAIDAgFIBAUE+PKDCNcYINMf0x/THwL4I7vyZO1E0NMf0x/T//QE0VFDuvKhUVG68qIF+QFUEGT5EPKj+AAkpMjLH1JAyx9SMMv/UhD0AMntVPgPAdMHIcAAn2xRkyDXSpbTB9QC+wDoMOAhwAHjACHAAuMAAcADkTDjDQOkyMsfEssfy/8PEBESAWTQMtDTAwFxsJFb4CHXScEgkVvgAdMfIYIQc2LQnL2SXwPgAvpAMAGCEHNi0Jy6kVvjDQYCASAHCADsAdMf+gD6QNQw0PpAMHBUcADIyx/4KM8WywHLAMsAyXAgghDjoNSCyMsfUATPFhPLACL6AhLLAMzJcIIQD4p+pcjLHxXLP1AD+gIBzxb4KM8WEssAggkxLQD6AgHQzxbJcYAYyMsFUAPPFnD6AhLLaszJgED7AAIBIAkKAFm9JCtvaiaECAoGuQ+gIYRw1AgIR6STfSmRDOaQPp/5g3gSgBt4EBSJhxWfMYQCAVgLDAARuMl+1E0NcLH4AD2ynftRNCBAUDXIfQEMALIygfL/8nQAYEBCPQKb6ExgAgEgDQ4AGa3OdqJoQCBrkOuF/8AAGa8d9qJoQBBrkOuFj8AAbtIH+gDU1CL5AAXIygcVy//J0Hd0gBjIywXLAiLPFlAF+gIUy2sSzMzJc/sAyEAUgQEI9FHypwIAcIEBCNcY+gDTP8hUIEeBAQj0UfKnghBub3RlcHSAGMjLBcsCUAbPFlAE+gIUy2oSyx/LP8lz+wACAGyBAQjXGPoA0z8wUiSBAQj0WfKnghBkc3RycHSAGMjLBcsCUAXPFlAD+gITy2rLHxLLP8lz+wAACvQAye1U
`

	_ = `
te6ccgECEwEAAn0AART/APSkE/S88sgLAQIBIAIDAgFIBAUE+PKDCNcYINMf0x/THwL4I7vyZO1E0NMf0x/T//QE0VFDuvKhUVG68qIF+QFUEGT5EPKj+AAkpMjLH1JAyx9SMMv/UhD0AMntVPgPAdMHIcAAn2xRkyDXSpbTB9QC+wDoMOAhwAHjACHAAuMAAcADkTDjDQOkyMsfEssfy/8PEBESAWTQMtDTAwFxsJFb4CHXScEgkVvgAdMfIYIQc2LQnL2SXwPgAvpAMAGCEHNi0Jy6kVvjDQYCASAHCADcAdMf+gD6QDBwVHAAyMsf+CjPFssBywDLAMlwUwCCEOOg1ILIyx/LAcsAIfoCywDMyXCCEA+KfqXIyx8Vyz9QA/oCAc8W+CjPFhLLAIIJMS0A+gIB0M8WyXGAGMjLBVADzxZw+gISy2rMyYBA+wACASAJCgBZvSQrb2omhAgKBrkPoCGEcNQICEekk30pkQzmkD6f+YN4EoAbeBAUiYcVnzGEAgFYCwwAEbjJftRNDXCx+AA9sp37UTQgQFA1yH0BDACyMoHy//J0AGBAQj0Cm+hMYAIBIA0OABmtznaiaEAga5Drhf/AABmvHfaiaEAQa5DrhY/AAG7SB/oA1NQi+QAFyMoHFcv/ydB3dIAYyMsFywIizxZQBfoCFMtrEszMyXP7AMhAFIEBCPRR8qcCAHCBAQjXGPoA0z/IVCBHgQEI9FHyp4IQbm90ZXB0gBjIywXLAlAGzxZQBPoCFMtqEssfyz/Jc/sAAgBsgQEI1xj6ANM/MFIkgQEI9Fnyp4IQZHN0cnB0gBjIywXLAlAFzxZQA/oCE8tqyx8Syz/Jc/sAAAr0AMntVA==
`

	_ = `
te6ccgECEwEAAn0AART/APSkE/S88sgLAQIBIAIDAgFIBAUE+PKDCNcYINMf0x/THwL4I7vyZO1E0NMf0x/T//QE0VFDuvKhUVG68qIF+QFUEGT5EPKj+AAkpMjLH1JAyx9SMMv/UhD0AMntVPgPAdMHIcAAn2xRkyDXSpbTB9QC+wDoMOAhwAHjACHAAuMAAcADkTDjDQOkyMsfEssfy/8PEBESAWTQMtDTAwFxsJFb4CHXScEgkVvgAdMfIYIQc2LQnL2SXwPgAvpAMAGCEHNi0Jy6kVvjDQYCASAHCADcAdMf+gD6QDBwVHAAyMsf+CjPFssBywDLAMlwUwCCEOOg1ILIyx/LAcsAIfoCywDMyXCCEA+KfqXIyx8Vyz9QA/oCAc8W+CjPFhLLAIIJMS0A+gIB0M8WyXGAGMjLBVADzxZw+gISy2rMyYBA+wACASAJCgBZvSQrb2omhAgKBrkPoCGEcNQICEekk30pkQzmkD6f+YN4EoAbeBAUiYcVnzGEAgFYCwwAEbjJftRNDXCx+AA9sp37UTQgQFA1yH0BDACyMoHy//J0AGBAQj0Cm+hMYAIBIA0OABmtznaiaEAga5Drhf/AABmvHfaiaEAQa5DrhY/AAG7SB/oA1NQi+QAFyMoHFcv/ydB3dIAYyMsFywIizxZQBfoCFMtrEszMyXP7AMhAFIEBCPRR8qcCAHCBAQjXGPoA0z/IVCBHgQEI9FHyp4IQbm90ZXB0gBjIywXLAlAGzxZQBPoCFMtqEssfyz/Jc/sAAgBsgQEI1xj6ANM/MFIkgQEI9Fnyp4IQZHN0cnB0gBjIywXLAlAFzxZQA/oCE8tqyx8Syz/Jc/sAAAr0AMntVA==
`

	_ = `
te6ccgECEwEAAoUAART/APSkE/S88sgLAQIBIAIDAgFIBAUE+PKDCNcYINMf0x/THwL4I7vyZO1E0NMf0x/T//QE0VFDuvKhUVG68qIF+QFUEGT5EPKj+AAkpMjLH1JAyx9SMMv/UhD0AMntVPgPAdMHIcAAn2xRkyDXSpbTB9QC+wDoMOAhwAHjACHAAuMAAcADkTDjDQOkyMsfEssfy/8PEBESAWTQMtDTAwFxsJFb4CHXScEgkVvgAdMfIYIQc2LQnL2SXwPgAvpAMAGCEHNi0Jy6kVvjDQYCASAHCADsAdM/+gD6QNQw0PpAMHBUcADIyx/4KM8WywHLAMsAyXAgghDjoNSCyMsfUATPFhPLACL6AhLLAMzJcIIQD4p+pcjLHxXLP1AD+gIBzxb4KM8WggkxLQD6AhLLAAHQzxbJcYAYyMsFUAPPFnD6AhLLaszJgED7AAIBIAkKAFm9JCtvaiaECAoGuQ+gIYRw1AgIR6STfSmRDOaQPp/5g3gSgBt4EBSJhxWfMYQCAVgLDAARuMl+1E0NcLH4AD2ynftRNCBAUDXIfQEMALIygfL/8nQAYEBCPQKb6ExgAgEgDQ4AGa3OdqJoQCBrkOuF/8AAGa8d9qJoQBBrkOuFj8AAbtIH+gDU1CL5AAXIygcVy//J0Hd0gBjIywXLAiLPFlAF+gIUy2sSzMzJc/sAyEAUgQEI9FHypwIAcIEBCNcY+gDTP8hUIEeBAQj0UfKnghBub3RlcHSAGMjLBcsCUAbPFlAE+gIUy2oSyx/LP8lz+wACAGyBAQjXGPoA0z8wUiSBAQj0WfKnghBkc3RycHSAGMjLBcsCUAXPFlAD+gITy2rLHxLLP8lz+wAACvQAye1U
`

	_ = `
te6ccgECEwEAAoUAART/APSkE/S88sgLAQIBIAIDAgFIBAUE+PKDCNcYINMf0x/THwL4I7vyZO1E0NMf0x/T//QE0VFDuvKhUVG68qIF+QFUEGT5EPKj+AAkpMjLH1JAyx9SMMv/UhD0AMntVPgPAdMHIcAAn2xRkyDXSpbTB9QC+wDoMOAhwAHjACHAAuMAAcADkTDjDQOkyMsfEssfy/8PEBESAWTQMtDTAwFxsJFb4CHXScEgkVvgAdMfIYIQc2LQnL2SXwPgAvpAMAGCEHNi0Jy6kVvjDQYCASAHCADsAdM/+gD6QNQw0PpAMHBUcADIyx/4KM8WywHLAMsAyXAgghDjoNSCyMsfUATPFhPLACL6AhLLAMzJcIIQD4p+pcjLHxXLP1AD+gIBzxb4KM8WggkxLQD6AhLLAAHQzxbJcYAYyMsFUAPPFnD6AhLLaszJgED7AAIBIAkKAFm9JCtvaiaECAoGuQ+gIYRw1AgIR6STfSmRDOaQPp/5g3gSgBt4EBSJhxWfMYQCAVgLDAARuMl+1E0NcLH4AD2ynftRNCBAUDXIfQEMALIygfL/8nQAYEBCPQKb6ExgAgEgDQ4AGa3OdqJoQCBrkOuF/8AAGa8d9qJoQBBrkOuFj8AAbtIH+gDU1CL5AAXIygcVy//J0Hd0gBjIywXLAiLPFlAF+gIUy2sSzMzJc/sAyEAUgQEI9FHypwIAcIEBCNcY+gDTP8hUIEeBAQj0UfKnghBub3RlcHSAGMjLBcsCUAbPFlAE+gIUy2oSyx/LP8lz+wACAGyBAQjXGPoA0z8wUiSBAQj0WfKnghBkc3RycHSAGMjLBcsCUAXPFlAD+gITy2rLHxLLP8lz+wAACvQAye1U
`
	_ = `
te6ccgECEwEAAogAART/APSkE/S88sgLAQIBIAIDAgFIBAUE+PKDCNcYINMf0x/THwL4I7vyZO1E0NMf0x/T//QE0VFDuvKhUVG68qIF+QFUEGT5EPKj+AAkpMjLH1JAyx9SMMv/UhD0AMntVPgPAdMHIcAAn2xRkyDXSpbTB9QC+wDoMOAhwAHjACHAAuMAAcADkTDjDQOkyMsfEssfy/8PEBESAWTQMtDTAwFxsJFb4CHXScEgkVvgAdMfIYIQc2LQnL2SXwPgAvpAMAGCEHNi0Jy6kVvjDQYCASAHCADyAdM/+gD6QNQw0PpAMHBUcADIyx/4KM8WywHLAMsAyXAgghDjoNSCyMsfUATPFhPLACL6AhLLAMzJcXCCEA+KfqXIyx8Wyz9QBPoCWM8W+CjPFhPLAIIJMS0A+gLLAAHQzxbJcYAYyMsFUAPPFnD6AhLLaszJgED7AAIBIAkKAFm9JCtvaiaECAoGuQ+gIYRw1AgIR6STfSmRDOaQPp/5g3gSgBt4EBSJhxWfMYQCAVgLDAARuMl+1E0NcLH4AD2ynftRNCBAUDXIfQEMALIygfL/8nQAYEBCPQKb6ExgAgEgDQ4AGa3OdqJoQCBrkOuF/8AAGa8d9qJoQBBrkOuFj8AAbtIH+gDU1CL5AAXIygcVy//J0Hd0gBjIywXLAiLPFlAF+gIUy2sSzMzJc/sAyEAUgQEI9FHypwIAcIEBCNcY+gDTP8hUIEeBAQj0UfKnghBub3RlcHSAGMjLBcsCUAbPFlAE+gIUy2oSyx/LP8lz+wACAGyBAQjXGPoA0z8wUiSBAQj0WfKnghBkc3RycHSAGMjLBcsCUAXPFlAD+gITy2rLHxLLP8lz+wAACvQAye1U
`

	_ = `
te6ccgECEwEAAokAART/APSkE/S88sgLAQIBIAIDAgFIBAUE+PKDCNcYINMf0x/THwL4I7vyZO1E0NMf0x/T//QE0VFDuvKhUVG68qIF+QFUEGT5EPKj+AAkpMjLH1JAyx9SMMv/UhD0AMntVPgPAdMHIcAAn2xRkyDXSpbTB9QC+wDoMOAhwAHjACHAAuMAAcADkTDjDQOkyMsfEssfy/8PEBESAWTQMtDTAwFxsJFb4CHXScEgkVvgAdMfIYIQc2LQnL2SXwPgAvpAMAGCEHNi0Jy6kVvjDQYCASAHCAD0AdM/+gD6QNQw0PpAMHBUcADIyx/4KM8WywHLAMsAyXFwIIIQ46DUgsjLH1AFzxYUywAj+gITywASywDMyXFwghAPin6lyMsfFss/UAT6AljPFvgozxYTywCCCTEtAPoCywDMyXGAGMjLBVADzxZw+gISy2rMyYBA+wACASAJCgBZvSQrb2omhAgKBrkPoCGEcNQICEekk30pkQzmkD6f+YN4EoAbeBAUiYcVnzGEAgFYCwwAEbjJftRNDXCx+AA9sp37UTQgQFA1yH0BDACyMoHy//J0AGBAQj0Cm+hMYAIBIA0OABmtznaiaEAga5Drhf/AABmvHfaiaEAQa5DrhY/AAG7SB/oA1NQi+QAFyMoHFcv/ydB3dIAYyMsFywIizxZQBfoCFMtrEszMyXP7AMhAFIEBCPRR8qcCAHCBAQjXGPoA0z/IVCBHgQEI9FHyp4IQbm90ZXB0gBjIywXLAlAGzxZQBPoCFMtqEssfyz/Jc/sAAgBsgQEI1xj6ANM/MFIkgQEI9Fnyp4IQZHN0cnB0gBjIywXLAlAFzxZQA/oCE8tqyx8Syz/Jc/sAAAr0AMntVA==
`
	_ = `
te6ccgECEwEAAogAART/APSkE/S88sgLAQIBIAIDAgFIBAUE+PKDCNcYINMf0x/THwL4I7vyZO1E0NMf0x/T//QE0VFDuvKhUVG68qIF+QFUEGT5EPKj+AAkpMjLH1JAyx9SMMv/UhD0AMntVPgPAdMHIcAAn2xRkyDXSpbTB9QC+wDoMOAhwAHjACHAAuMAAcADkTDjDQOkyMsfEssfy/8PEBESAWTQMtDTAwFxsJFb4CHXScEgkVvgAdMfIYIQc2LQnL2SXwPgAvpAMAGCEHNi0Jy6kVvjDQYCASAHCADyAdM/+gD6QNQw0PpAMHBUcAAgyMsfywHLAcsAywDJcXAgghDjoNSCyMsfUAXPFhTLACP6AhPLABLLAMzJcXCCEA+KfqXIyx8Wyz9QBPoCWM8W+CjPFhPLAIIJMS0A+gLLAMzJcYAYyMsFUAPPFnD6AhLLaszJgED7AAIBIAkKAFm9JCtvaiaECAoGuQ+gIYRw1AgIR6STfSmRDOaQPp/5g3gSgBt4EBSJhxWfMYQCAVgLDAARuMl+1E0NcLH4AD2ynftRNCBAUDXIfQEMALIygfL/8nQAYEBCPQKb6ExgAgEgDQ4AGa3OdqJoQCBrkOuF/8AAGa8d9qJoQBBrkOuFj8AAbtIH+gDU1CL5AAXIygcVy//J0Hd0gBjIywXLAiLPFlAF+gIUy2sSzMzJc/sAyEAUgQEI9FHypwIAcIEBCNcY+gDTP8hUIEeBAQj0UfKnghBub3RlcHSAGMjLBcsCUAbPFlAE+gIUy2oSyx/LP8lz+wACAGyBAQjXGPoA0z8wUiSBAQj0WfKnghBkc3RycHSAGMjLBcsCUAXPFlAD+gITy2rLHxLLP8lz+wAACvQAye1U
`

	_ = `
te6ccgECEwEAAogAART/APSkE/S88sgLAQIBIAIDAgFIBAUE+PKDCNcYINMf0x/THwL4I7vyZO1E0NMf0x/T//QE0VFDuvKhUVG68qIF+QFUEGT5EPKj+AAkpMjLH1JAyx9SMMv/UhD0AMntVPgPAdMHIcAAn2xRkyDXSpbTB9QC+wDoMOAhwAHjACHAAuMAAcADkTDjDQOkyMsfEssfy/8PEBESAWTQMtDTAwFxsJFb4CHXScEgkVvgAdMfIYIQc2LQnL2SXwPgAvpAMAGCEHNi0Jy6kVvjDQYCASAHCADyAdM/+gD6QNQw0PpAMHBUcAAgyMsfywHLAcsAywDJcXAgghDjoNSCyMsfUAXPFhTLACP6AhPLABLLAMzJcXCCEA+KfqXIyx8Wyz9QBPoCWM8W+CjPFhPLAIIJMS0A+gLLAMzJcYAYyMsFUAPPFnD6AhLLaszJgEP7AAIBIAkKAFm9JCtvaiaECAoGuQ+gIYRw1AgIR6STfSmRDOaQPp/5g3gSgBt4EBSJhxWfMYQCAVgLDAARuMl+1E0NcLH4AD2ynftRNCBAUDXIfQEMALIygfL/8nQAYEBCPQKb6ExgAgEgDQ4AGa3OdqJoQCBrkOuF/8AAGa8d9qJoQBBrkOuFj8AAbtIH+gDU1CL5AAXIygcVy//J0Hd0gBjIywXLAiLPFlAF+gIUy2sSzMzJc/sAyEAUgQEI9FHypwIAcIEBCNcY+gDTP8hUIEeBAQj0UfKnghBub3RlcHSAGMjLBcsCUAbPFlAE+gIUy2oSyx/LP8lz+wACAGyBAQjXGPoA0z8wUiSBAQj0WfKnghBkc3RycHSAGMjLBcsCUAXPFlAD+gITy2rLHxLLP8lz+wAACvQAye1U
`

	_ = `
te6ccgECEwEAAoQAART/APSkE/S88sgLAQIBIAIDAgFIBAUE+PKDCNcYINMf0x/THwL4I7vyZO1E0NMf0x/T//QE0VFDuvKhUVG68qIF+QFUEGT5EPKj+AAkpMjLH1JAyx9SMMv/UhD0AMntVPgPAdMHIcAAn2xRkyDXSpbTB9QC+wDoMOAhwAHjACHAAuMAAcADkTDjDQOkyMsfEssfy/8PEBESAWTQMtDTAwFxsJFb4CHXScEgkVvgAdMfIYIQc2LQnL2SXwPgAvpAMAGCEHNi0Jy6kVvjDQYCASAHCADqAdM/+gD6QNQw0PpAMHBUcAAgyMsfywHLAcsAywDJcCCCEOOg1ILIyx9QBM8WE8sAIvoCEssAzMlxcIIQD4p+pcjLHxbLP1AE+gJYzxb4KM8WE8sAggkxLQD6AssAzMlxgBjIywVQA88WcPoCEstqzMmAQ/sAAgEgCQoAWb0kK29qJoQICga5D6AhhHDUCAhHpJN9KZEM5pA+n/mDeBKAG3gQFImHFZ8xhAIBWAsMABG4yX7UTQ1wsfgAPbKd+1E0IEBQNch9AQwAsjKB8v/ydABgQEI9ApvoTGACASANDgAZrc52omhAIGuQ64X/wAAZrx32omhAEGuQ64WPwABu0gf6ANTUIvkABcjKBxXL/8nQd3SAGMjLBcsCIs8WUAX6AhTLaxLMzMlz+wDIQBSBAQj0UfKnAgBwgQEI1xj6ANM/yFQgR4EBCPRR8qeCEG5vdGVwdIAYyMsFywJQBs8WUAT6AhTLahLLH8s/yXP7AAIAbIEBCNcY+gDTPzBSJIEBCPRZ8qeCEGRzdHJwdIAYyMsFywJQBc8WUAP6AhPLassfEss/yXP7AAAK9ADJ7VQ=
`

	_ = `
te6ccgECEwEAAoQAART/APSkE/S88sgLAQIBIAIDAgFIBAUE+PKDCNcYINMf0x/THwL4I7vyZO1E0NMf0x/T//QE0VFDuvKhUVG68qIF+QFUEGT5EPKj+AAkpMjLH1JAyx9SMMv/UhD0AMntVPgPAdMHIcAAn2xRkyDXSpbTB9QC+wDoMOAhwAHjACHAAuMAAcADkTDjDQOkyMsfEssfy/8PEBESAWTQMtDTAwFxsJFb4CHXScEgkVvgAdMfIYIQc2LQnL2SXwPgAvpAMAGCEHNi0Jy6kVvjDQYCASAHCADqAdM/+gD6QNQw0PpAMHBUcAAgyMsfywHLAcsAywDJcCCCEOOg1ILIyx9QBM8WE8sAIvoCEssAzMlxcIIQD4p+pcjLHxbLP1AE+gJYzxb4KM8WE8sAggr68ID6AssAzMlxgBjIywVQA88WcPoCEstqzMmAQ/sAAgEgCQoAWb0kK29qJoQICga5D6AhhHDUCAhHpJN9KZEM5pA+n/mDeBKAG3gQFImHFZ8xhAIBWAsMABG4yX7UTQ1wsfgAPbKd+1E0IEBQNch9AQwAsjKB8v/ydABgQEI9ApvoTGACASANDgAZrc52omhAIGuQ64X/wAAZrx32omhAEGuQ64WPwABu0gf6ANTUIvkABcjKBxXL/8nQd3SAGMjLBcsCIs8WUAX6AhTLaxLMzMlz+wDIQBSBAQj0UfKnAgBwgQEI1xj6ANM/yFQgR4EBCPRR8qeCEG5vdGVwdIAYyMsFywJQBs8WUAT6AhTLahLLH8s/yXP7AAIAbIEBCNcY+gDTPzBSJIEBCPRZ8qeCEGRzdHJwdIAYyMsFywJQBc8WUAP6AhPLassfEss/yXP7AAAK9ADJ7VQ=
`

	_ = `
te6ccgECEwEAAoQAART/APSkE/S88sgLAQIBIAIDAgFIBAUE+PKDCNcYINMf0x/THwL4I7vyZO1E0NMf0x/T//QE0VFDuvKhUVG68qIF+QFUEGT5EPKj+AAkpMjLH1JAyx9SMMv/UhD0AMntVPgPAdMHIcAAn2xRkyDXSpbTB9QC+wDoMOAhwAHjACHAAuMAAcADkTDjDQOkyMsfEssfy/8PEBESAWTQMtDTAwFxsJFb4CHXScEgkVvgAdMfIYIQc2LQnL2SXwPgAvpAMAGCEHNi0Jy6kVvjDQYCASAHCADqAdM/+gD6QNQw0PpAMHBUcAAgyMsfywHLAcsAywDJcCCCEOOg1ILIyx9QBM8WE8sAIvoCEssAzMlxcIIQD4p+pcjLHxbLP1AE+gJYzxb4KM8WE8sAghAL68IA+gLLAMzJcYAYyMsFUAPPFnD6AhLLaszJc/sAAgEgCQoAWb0kK29qJoQICga5D6AhhHDUCAhHpJN9KZEM5pA+n/mDeBKAG3gQFImHFZ8xhAIBWAsMABG4yX7UTQ1wsfgAPbKd+1E0IEBQNch9AQwAsjKB8v/ydABgQEI9ApvoTGACASANDgAZrc52omhAIGuQ64X/wAAZrx32omhAEGuQ64WPwABu0gf6ANTUIvkABcjKBxXL/8nQd3SAGMjLBcsCIs8WUAX6AhTLaxLMzMlz+wDIQBSBAQj0UfKnAgBwgQEI1xj6ANM/yFQgR4EBCPRR8qeCEG5vdGVwdIAYyMsFywJQBs8WUAT6AhTLahLLH8s/yXP7AAIAbIEBCNcY+gDTPzBSJIEBCPRZ8qeCEGRzdHJwdIAYyMsFywJQBc8WUAP6AhPLassfEss/yXP7AAAK9ADJ7VQ=
`

	_ = `
te6ccgECEwEAAoUAART/APSkE/S88sgLAQIBIAIDAgFIBAUE+PKDCNcYINMf0x/THwL4I7vyZO1E0NMf0x/T//QE0VFDuvKhUVG68qIF+QFUEGT5EPKj+AAkpMjLH1JAyx9SMMv/UhD0AMntVPgPAdMHIcAAn2xRkyDXSpbTB9QC+wDoMOAhwAHjACHAAuMAAcADkTDjDQOkyMsfEssfy/8PEBESAWTQMtDTAwFxsJFb4CHXScEgkVvgAdMfIYIQc2LQnL2SXwPgAvpAMAGCEHNi0Jy6kVvjDQYCASAHCADsAdM/+gD6QNQw0PpAMHBUcAAgyMsfywHLAcsAywDJcCCCEOOg1ILIyx9QBM8WE8sAIvoCEssAzMlxcIIQD4p+pcjLHxbLP1AE+gJYzxb4KM8WE8sAghAL68IA+gLLAMzJcYAYyMsFUAPPFnD6AhLLaszJgEP7AAIBIAkKAFm9JCtvaiaECAoGuQ+gIYRw1AgIR6STfSmRDOaQPp/5g3gSgBt4EBSJhxWfMYQCAVgLDAARuMl+1E0NcLH4AD2ynftRNCBAUDXIfQEMALIygfL/8nQAYEBCPQKb6ExgAgEgDQ4AGa3OdqJoQCBrkOuF/8AAGa8d9qJoQBBrkOuFj8AAbtIH+gDU1CL5AAXIygcVy//J0Hd0gBjIywXLAiLPFlAF+gIUy2sSzMzJc/sAyEAUgQEI9FHypwIAcIEBCNcY+gDTP8hUIEeBAQj0UfKnghBub3RlcHSAGMjLBcsCUAbPFlAE+gIUy2oSyx/LP8lz+wACAGyBAQjXGPoA0z8wUiSBAQj0WfKnghBkc3RycHSAGMjLBcsCUAXPFlAD+gITy2rLHxLLP8lz+wAACvQAye1U
`

	_ = `
te6ccgECEwEAAokAART/APSkE/S88sgLAQIBIAIDAgFIBAUE+PKDCNcYINMf0x/THwL4I7vyZO1E0NMf0x/T//QE0VFDuvKhUVG68qIF+QFUEGT5EPKj+AAkpMjLH1JAyx9SMMv/UhD0AMntVPgPAdMHIcAAn2xRkyDXSpbTB9QC+wDoMOAhwAHjACHAAuMAAcADkTDjDQOkyMsfEssfy/8PEBESAWTQMtDTAwFxsJFb4CHXScEgkVvgAdMfIYIQc2LQnL2SXwPgAvpAMAGCEHNi0Jy6kVvjDQYCASAHCAD0AdM/+gD6QNQw0PpAMHBUcAAgyMsfywHLAcsAywDJcCCCEOOg1ILIyx9QBM8WE8sAIvoCEssAzMlxcIIQD4p+pcjLHxbLP1AE+gJYzxb4KM8WE8sAghAL68IA+gLLAMzJcYAYyMsFUAPPFoIQDuaygPoCEstqzMlw+wACASAJCgBZvSQrb2omhAgKBrkPoCGEcNQICEekk30pkQzmkD6f+YN4EoAbeBAUiYcVnzGEAgFYCwwAEbjJftRNDXCx+AA9sp37UTQgQFA1yH0BDACyMoHy//J0AGBAQj0Cm+hMYAIBIA0OABmtznaiaEAga5Drhf/AABmvHfaiaEAQa5DrhY/AAG7SB/oA1NQi+QAFyMoHFcv/ydB3dIAYyMsFywIizxZQBfoCFMtrEszMyXP7AMhAFIEBCPRR8qcCAHCBAQjXGPoA0z/IVCBHgQEI9FHyp4IQbm90ZXB0gBjIywXLAlAGzxZQBPoCFMtqEssfyz/Jc/sAAgBsgQEI1xj6ANM/MFIkgQEI9Fnyp4IQZHN0cnB0gBjIywXLAlAFzxZQA/oCE8tqyx8Syz/Jc/sAAAr0AMntVA==
`

	_ = `
te6ccgECEgEAAoEAART/APSkE/S88sgLAQIBIAIDAgFIBAUD+PKDCNcYINMf0x/THzEB+CO78mTtRNDTH9Mf0//0BNFRUrryogX5AVQQZfkQ8qP4ACCkyMsfUlDLH1JAy/9SMPQAye1U+A8B0wchwACfbFGTINdKltMH1AL7AOgw4CHAAeMAIcAC4wABwAORMOMNpMjLHxPLH8v/9ADJ7VQPEBEBZNAy0NMDAXGwkVvgIddJwSCRW+AB0x8hghBzYtCcvZJfA+AC+kAwAYIQc2LQnLqRW+MNBgIBIAcIAPQB0z/6APpA1DDQ+kAwcFRwACDIyx/LAcsBywDLAMlwIIIQ46DUgsjLH1AEzxYTywAi+gISywDMyXFwghAPin6lyMsfFss/UAT6AljPFvgozxYTywCCEAvrwgD6AssAzMlxgBjIywVQA88WghAO5rKA+gISy2rMyXD7AAIBIAkKAFm9JCtvaiaECAoGuQ+gIYRw1AgIR6STfSmRDOaQPp/5g3gSgBt4EBSJhxWfMYQCAVgLDAARuMl+1E0NcLH4AD2ynftRNCBAUDXIfQEMALIygfL/8nQAYEBCPQKb6ExgAgEgDQ4AGa3OdqJoQCBrkOuF/8AAGa8d9qJoQBBrkOuFj8AAbtIH+gDU1CL5AAXIygcVy//J0Hd0gBjIywXLAiLPFlAF+gIUy2sSzMzJc/sAyEAVgQEI9FHypwMAcIEBCNcY+gDTP8hUIEiBAQj0UfKnghBub3RlcHSAGMjLBcsCUAbPFlAE+gIUy2oSyx/LP8lz+wADAGyBAQjXGPoA0z8wUiWBAQj0WfKnghBkc3RycHSAGMjLBcsCUAXPFlAD+gITy2rLHxPLP8lz+wA=

`

	_ = `
te6ccgECEwEAAokAART/APSkE/S88sgLAQIBIAIDAgFIBAUE+PKDCNcYINMf0x/THwL4I7vyZO1E0NMf0x/T//QE0VFDuvKhUVG68qIF+QFUEGT5EPKj+AAkpMjLH1JAyx9SMMv/UhD0AMntVPgPAdMHIcAAn2xRkyDXSpbTB9QC+wDoMOAhwAHjACHAAuMAAcADkTDjDQOkyMsfEssfy/8PEBESAWTQMtDTAwFxsJFb4CHXScEgkVvgAdMfIYIQc2LQnL2SXwPgAvpAMAGCEHNi0Jy6kVvjDQYCASAHCAD0AdM/+gD6QNQw0PpAMHBUcAAgyMsfywHLAcsAywDJcCCCEOOg1ILIyx9QBM8WE8sAIvoCEssAzMlxcIIQD4p+pcjLHxbLP1AE+gJYzxb4KM8WE8sAghAL68IA+gLLAMzJcYAYyMsFUAPPFoIQDuaygPoCEstqzMlw+wACASAJCgBZvSQrb2omhAgKBrkPoCGEcNQICEekk30pkQzmkD6f+YN4EoAbeBAUiYcVnzGEAgFYCwwAEbjJftRNDXCx+AA9sp37UTQgQFA1yH0BDACyMoHy//J0AGBAQj0Cm+hMYAIBIA0OABmtznaiaEAga5Drhf/AABmvHfaiaEAQa5DrhY/AAG7SB/oA1NQi+QAFyMoHFcv/ydB3dIAYyMsFywIizxZQBfoCFMtrEszMyXP7AMhAFIEBCPRR8qcCAHCBAQjXGPoA0z/IVCBHgQEI9FHyp4IQbm90ZXB0gBjIywXLAlAGzxZQBPoCFMtqEssfyz/Jc/sAAgBsgQEI1xj6ANM/MFIkgQEI9Fnyp4IQZHN0cnB0gBjIywXLAlAFzxZQA/oCE8tqyx8Syz/Jc/sAAAr0AMntVA==
`

	_ = `
te6ccgECEwEAArgAART/APSkE/S88sgLAQIBIAIDAgFIBAUEvvKDCNcYINMf0x/THwL4I7vyZO1E0NMf0x/T//QE0VFDuvKhUVG68qIF+QFUEGT5EPKj+AAkpMjLH1JAyx9SMMv/UhD0AMntVPgPAdMHIcAA4wIhwAHjACHAAuMAAcADDxAREgFk0DLQ0wMBcbCRW+Ah10nBIJFb4AHTHyGCEHNi0Jy9kl8D4AL6QDABghBzYtCcupFb4w0GAgEgBwgA9AHTP/oA+kDUMND6QDBwVHAAIMjLH8sBywHLAMsAyXAgghDjoNSCyMsfUATPFhPLACL6AhLLAMzJcXCCEA+KfqXIyx8Wyz9QBPoCWM8W+CjPFhPLAIIQC+vCAPoCywDMyXGAGMjLBVADzxaCEA7msoD6AhLLaszJcPsAAgEgCQoAWb0kK29qJoQICga5D6AhhHDUCAhHpJN9KZEM5pA+n/mDeBKAG3gQFImHFZ8xhAIBWAsMABG4yX7UTQ1wsfgAPbKd+1E0IEBQNch9AQwAsjKB8v/ydABgQEI9ApvoTGACASANDgAZrc52omhAIGuQ64X/wAAZrx32omhAEGuQ64WPwAB6bFGTINdKjjPTB9Qh0CDXSQHTHwGCEOoGGF+6jhgzghDqBhhdyMsfAabgE9cYMBLPFslY+wCUWwL7AOLoMABu0gf6ANTUIvkABcjKBxXL/8nQd3SAGMjLBcsCIs8WUAX6AhTLaxLMzMlz+wDIQBSBAQj0UfKnAgBwgQEI1xj6ANM/yFQgR4EBCPRR8qeCEG5vdGVwdIAYyMsFywJQBs8WUAT6AhTLahLLH8s/yXP7AAIAlI42gQEI1xj6ANM/MFIkgQEI9Fnyp4IQZHN0cnB0gBjIywXLAlAFzxZQA/oCE8tqyx8Syz/Jc/sAkTDiA6TIyx8Syx/L//QAye1U
`
	_ = `
te6ccgECEwEAArUAART/APSkE/S88sgLAQIBIAIDAgFIBAUEvvKDCNcYINMf0x/THwL4I7vyZO1E0NMf0x/T//QE0VFDuvKhUVG68qIF+QFUEGT5EPKj+AAkpMjLH1JAyx9SMMv/UhD0AMntVPgPAdMHIcAA4wIhwAHjACHAAuMAAcADDxAREgFk0DLQ0wMBcbCRW+Ah10nBIJFb4AHTHyGCEHNi0Jy9kl8D4AL6QDABghBzYtCcupFb4w0GAgEgBwgA9AHTP/oA+kDUMND6QDBwVHAAIMjLH8sBywHLAMsAyXAgghDjoNSCyMsfUATPFhPLACL6AhLLAMzJcXCCEA+KfqXIyx8Wyz9QBPoCWM8W+CjPFhPLAIIQC+vCAPoCywDMyXGAGMjLBVADzxaCEA7msoD6AhLLaszJcPsAAgEgCQoAWb0kK29qJoQICga5D6AhhHDUCAhHpJN9KZEM5pA+n/mDeBKAG3gQFImHFZ8xhAIBWAsMABG4yX7UTQ1wsfgAPbKd+1E0IEBQNch9AQwAsjKB8v/ydABgQEI9ApvoTGACASANDgAZrc52omhAIGuQ64X/wAAZrx32omhAEGuQ64WPwAB0bFGTINdKjjDTB8gB1CHQ0wUx+kAx+gAx02ox1DHUMNDTHzCCEMoGGF+6ljEByVj7AJVsEgL7AOLoMABu0gf6ANTUIvkABcjKBxXL/8nQd3SAGMjLBcsCIs8WUAX6AhTLaxLMzMlz+wDIQBSBAQj0UfKnAgBwgQEI1xj6ANM/yFQgR4EBCPRR8qeCEG5vdGVwdIAYyMsFywJQBs8WUAT6AhTLahLLH8s/yXP7AAIAlI42gQEI1xj6ANM/MFIkgQEI9Fnyp4IQZHN0cnB0gBjIywXLAlAFzxZQA/oCE8tqyx8Syz/Jc/sAkTDiA6TIyx8Syx/L//QAye1U
`

	_ = `
te6ccgECEwEAArUAART/APSkE/S88sgLAQIBIAIDAgFIBAUEvvKDCNcYINMf0x/THwL4I7vyZO1E0NMf0x/T//QE0VFDuvKhUVG68qIF+QFUEGT5EPKj+AAkpMjLH1JAyx9SMMv/UhD0AMntVPgPAdMHIcAA4wIhwAHjACHAAuMAAcADDxAREgFk0DLQ0wMBcbCRW+Ah10nBIJFb4AHTHyGCEHNi0Jy9kl8D4AL6QDABghBzYtCcupFb4w0GAgEgBwgA9AHTP/oA+kDUMND6QDBwVHAAIMjLH8sBywHLAMsAyXAgghDjoNSCyMsfUATPFhPLACL6AhLLAMzJcXCCEA+KfqXIyx8Wyz9QBPoCWM8W+CjPFhPLAIIQC+vCAPoCywDMyXGAGMjLBVADzxaCEA7msoD6AhLLaszJcPsAAgEgCQoAWb0kK29qJoQICga5D6AhhHDUCAhHpJN9KZEM5pA+n/mDeBKAG3gQFImHFZ8xhAIBWAsMABG4yX7UTQ1wsfgAPbKd+1E0IEBQNch9AQwAsjKB8v/ydABgQEI9ApvoTGACASANDgAZrc52omhAIGuQ64X/wAAZrx32omhAEGuQ64WPwAB0bFGTINdKjjDTB8gB1CHQ0wUx+kAx+gAx02ox1DHUMNDTHzCCEMoGGF+6ljEByVj7AJVsEgL7AOLoMABu0gf6ANTUIvkABcjKBxXL/8nQd3SAGMjLBcsCIs8WUAX6AhTLaxLMzMlz+wDIQBSBAQj0UfKnAgBwgQEI1xj6ANM/yFQgR4EBCPRR8qeCEG5vdGVwdIAYyMsFywJQBs8WUAT6AhTLahLLH8s/yXP7AAIAlI42gQEI1xj6ANM/MFIkgQEI9Fnyp4IQZHN0cnB0gBjIywXLAlAFzxZQA/oCE8tqyx8Syz/Jc/sAkTDiA6TIyx8Syx/L//QAye1U
`

	_ = `
te6ccgECEwEAAssAART/APSkE/S88sgLAQIBIAIDAgFIBAUE5vKDCNcYINMf0x/THwL4I7vyZO1E0NMf0x/T//QE0VFDuvKhUVG68qIF+QFUEGT5EPKj+AAkpMjLH1JAyx9SMMv/UhD0AMntVPgPAdMHIcAAn2xRkyDXSpbTB9QC+wDoMOAhwAHjACHAAuMAIcAD4wABwAUPEBESAWTQMtDTAwFxsJFb4CHXScEgkVvgAdMfIYIQc2LQnL2SXwPgAvpAMAGCEHNi0Jy6kVvjDQYCASAHCAD0AdM/+gD6QNQw0PpAMHBUcAAgyMsfywHLAcsAywDJcCCCEOOg1ILIyx9QBM8WE8sAIvoCEssAzMlxcIIQD4p+pcjLHxbLP1AE+gJYzxb4KM8WE8sAghAL68IA+gLLAMzJcYAYyMsFUAPPFoIQDuaygPoCEstqzMlw+wACASAJCgBZvSQrb2omhAgKBrkPoCGEcNQICEekk30pkQzmkD6f+YN4EoAbeBAUiYcVnzGEAgFYCwwAEbjJftRNDXCx+AA9sp37UTQgQFA1yH0BDACyMoHy//J0AGBAQj0Cm+hMYAIBIA0OABmtznaiaEAga5Drhf/AABmvHfaiaEAQa5DrhY/AAG7SB/oA1NQi+QAFyMoHFcv/ydB3dIAYyMsFywIizxZQBfoCFMtrEszMyXP7AMhAFIEBCPRR8qcCAHCBAQjXGPoA0z/IVCBHgQEI9FHyp4IQbm90ZXB0gBjIywXLAlAGzxZQBPoCFMtqEssfyz/Jc/sAAgBsgQEI1xj6ANM/UjaBAQj0WfKnghBkc3RycHSAGMjLBcsCUAbPFlAE+gIUy2oSyx/LP8lz+wACAKCOPJMg10qONNMH1CHQyAHTBTH6QDH6ADHTajHUMdQw0NMfMIIQygYYX7qWbBLJIvsAlTBREvsA4tQC+wDoMJEw4gOkyMsfEssfy//0AMntVA==
`
	_ = `
te6ccgECEwEAAsIAART/APSkE/S88sgLAQIBIAIDAgFIBAUE5vKDCNcYINMf0x/THwL4I7vyZO1E0NMf0x/T//QE0VFDuvKhUVG68qIF+QFUEGT5EPKj+AAkpMjLH1JAyx9SMMv/UhD0AMntVPgPAdMHIcAAn2xRkyDXSpbTB9QC+wDoMOAhwAHjACHAAuMAIcAD4wABwAUPEBESAWTQMtDTAwFxsJFb4CHXScEgkVvgAdMfIYIQc2LQnL2SXwPgAvpAMAGCEHNi0Jy6kVvjDQYCASAHCAD0AdM/+gD6QNQw0PpAMHBUcAAgyMsfywHLAcsAywDJcCCCEOOg1ILIyx9QBM8WE8sAIvoCEssAzMlxcIIQD4p+pcjLHxbLP1AE+gJYzxb4KM8WE8sAghAL68IA+gLLAMzJcYAYyMsFUAPPFoIQDuaygPoCEstqzMlw+wACASAJCgBZvSQrb2omhAgKBrkPoCGEcNQICEekk30pkQzmkD6f+YN4EoAbeBAUiYcVnzGEAgFYCwwAEbjJftRNDXCx+AA9sp37UTQgQFA1yH0BDACyMoHy//J0AGBAQj0Cm+hMYAIBIA0OABmtznaiaEAga5Drhf/AABmvHfaiaEAQa5DrhY/AAG7SB/oA1NQi+QAFyMoHFcv/ydB3dIAYyMsFywIizxZQBfoCFMtrEszMyXP7AMhAFIEBCPRR8qcCAHCBAQjXGPoA0z/IVCBHgQEI9FHyp4IQbm90ZXB0gBjIywXLAlAGzxZQBPoCFMtqEssfyz/Jc/sAAgBsgQEI1xj6ANM/UjaBAQj0WfKnghBkc3RycHSAGMjLBcsCUAbPFlAE+gIUy2oSyx/LP8lz+wACAI6OM5Mg10qOK9MH1AHQyAHTBTH6QDH6ADHTajHUMdQw0NMfMIIQygYYX7qUyVj7AJIwMeLoMJEw4gOkyMsfEssfy//0AMntVA==
`

	_ = `
te6ccgECEwEAAscAART/APSkE/S88sgLAQIBIAIDAgFIBAUE5vKDCNcYINMf0x/THwL4I7vyZO1E0NMf0x/T//QE0VFDuvKhUVG68qIF+QFUEGT5EPKj+AAkpMjLH1JAyx9SMMv/UhD0AMntVPgPAdMHIcAAn2xRkyDXSpbTB9QC+wDoMOAhwAHjACHAAuMAIcAD4wABwAUPEBESAWTQMtDTAwFxsJFb4CHXScEgkVvgAdMfIYIQc2LQnL2SXwPgAvpAMAGCEHNi0Jy6kVvjDQYCASAHCAD0AdM/+gD6QNQw0PpAMHBUcAAgyMsfywHLAcsAywDJcCCCEOOg1ILIyx9QBM8WE8sAIvoCEssAzMlxcIIQD4p+pcjLHxbLP1AE+gJYzxb4KM8WE8sAghAL68IA+gLLAMzJcYAYyMsFUAPPFoIQDuaygPoCEstqzMlw+wACASAJCgBZvSQrb2omhAgKBrkPoCGEcNQICEekk30pkQzmkD6f+YN4EoAbeBAUiYcVnzGEAgFYCwwAEbjJftRNDXCx+AA9sp37UTQgQFA1yH0BDACyMoHy//J0AGBAQj0Cm+hMYAIBIA0OABmtznaiaEAga5Drhf/AABmvHfaiaEAQa5DrhY/AAG7SB/oA1NQi+QAFyMoHFcv/ydB3dIAYyMsFywIizxZQBfoCFMtrEszMyXP7AMhAFIEBCPRR8qcCAHCBAQjXGPoA0z/IVCBHgQEI9FHyp4IQbm90ZXB0gBjIywXLAlAGzxZQBPoCFMtqEssfyz/Jc/sAAgBsgQEI1xj6ANM/UjaBAQj0WfKnghBkc3RycHSAGMjLBcsCUAbPFlAE+gIUy2oSyx/LP8lz+wACAJiOOJMg10qOMNMH1AHQ02rTBQHIywUB+kBZzxYB+gBZ+gISy2oB0x8wghDKBhhfupTJWPsAkjAx4ugwkTDiA6TIyx8Syx/L//QAye1U
`

	_ = `
te6ccgECEwEAAsUAART/APSkE/S88sgLAQIBIAIDAgFIBAUE5vKDCNcYINMf0x/THwL4I7vyZO1E0NMf0x/T//QE0VFDuvKhUVG68qIF+QFUEGT5EPKj+AAkpMjLH1JAyx9SMMv/UhD0AMntVPgPAdMHIcAAn2xRkyDXSpbTB9QC+wDoMOAhwAHjACHAAuMAIcAD4wABwAUPEBESAWTQMtDTAwFxsJFb4CHXScEgkVvgAdMfIYIQc2LQnL2SXwPgAvpAMAGCEHNi0Jy6kVvjDQYCASAHCAD0AdM/+gD6QNQw0PpAMHBUcAAgyMsfywHLAcsAywDJcCCCEOOg1ILIyx9QBM8WE8sAIvoCEssAzMlxcIIQD4p+pcjLHxbLP1AE+gJYzxb4KM8WE8sAghAL68IA+gLLAMzJcYAYyMsFUAPPFoIQDuaygPoCEstqzMlw+wACASAJCgBZvSQrb2omhAgKBrkPoCGEcNQICEekk30pkQzmkD6f+YN4EoAbeBAUiYcVnzGEAgFYCwwAEbjJftRNDXCx+AA9sp37UTQgQFA1yH0BDACyMoHy//J0AGBAQj0Cm+hMYAIBIA0OABmtznaiaEAga5Drhf/AABmvHfaiaEAQa5DrhY/AAG7SB/oA1NQi+QAFyMoHFcv/ydB3dIAYyMsFywIizxZQBfoCFMtrEszMyXP7AMhAFIEBCPRR8qcCAHCBAQjXGPoA0z/IVCBHgQEI9FHyp4IQbm90ZXB0gBjIywXLAlAGzxZQBPoCFMtqEssfyz/Jc/sAAgBsgQEI1xj6ANM/UjaBAQj0WfKnghBkc3RycHSAGMjLBcsCUAbPFlAE+gIUy2oSyx/LP8lz+wACAJSONpMg10qOLtMH1AHQ0wX6QPoA02oEyMsFUAPPFgH6AstqAdMfMIIQygYYX7qUyVj7AJIwMeLoMJEw4gOkyMsfEssfy//0AMntVA==
`

	_ = `
te6ccgECEwEAAtMAART/APSkE/S88sgLAQIBIAIDAgFIBAUE5vKDCNcYINMf0x/THwL4I7vyZO1E0NMf0x/T//QE0VFDuvKhUVG68qIF+QFUEGT5EPKj+AAkpMjLH1JAyx9SMMv/UhD0AMntVPgPAdMHIcAAn2xRkyDXSpbTB9QC+wDoMOAhwAHjACHAAuMAIcAD4wABwAUPEBESAWTQMtDTAwFxsJFb4CHXScEgkVvgAdMfIYIQc2LQnL2SXwPgAvpAMAGCEHNi0Jy6kVvjDQYCASAHCAD0AdM/+gD6QNQw0PpAMHBUcAAgyMsfywHLAcsAywDJcCCCEOOg1ILIyx9QBM8WE8sAIvoCEssAzMlxcIIQD4p+pcjLHxbLP1AE+gJYzxb4KM8WE8sAghAL68IA+gLLAMzJcYAYyMsFUAPPFoIQDuaygPoCEstqzMlw+wACASAJCgBZvSQrb2omhAgKBrkPoCGEcNQICEekk30pkQzmkD6f+YN4EoAbeBAUiYcVnzGEAgFYCwwAEbjJftRNDXCx+AA9sp37UTQgQFA1yH0BDACyMoHy//J0AGBAQj0Cm+hMYAIBIA0OABmtznaiaEAga5Drhf/AABmvHfaiaEAQa5DrhY/AAG7SB/oA1NQi+QAFyMoHFcv/ydB3dIAYyMsFywIizxZQBfoCFMtrEszMyXP7AMhAFIEBCPRR8qcCAHCBAQjXGPoA0z/IVCBHgQEI9FHyp4IQbm90ZXB0gBjIywXLAlAGzxZQBPoCFMtqEssfyz/Jc/sAAgBsgQEI1xj6ANM/UjaBAQj0WfKnghBkc3RycHSAGMjLBcsCUAbPFlAE+gIUy2oSyx/LP8lz+wACALCORJMg10qOPNMH1AHQ0wX6QPoA02rTHwGCEMoGGF+6jh2CEMoGGF0FyMsFUATPFlj6AstqEssfAc8WyVj7AJQQVl8G4ugwkTDiA6TIyx8Syx/L//QAye1U
`

	// 这里有钱
	_ = `
te6ccgECEwEAAtMAART/APSkE/S88sgLAQIBIAIDAgFIBAUE5vKDCNcYINMf0x/THwL4I7vyZO1E0NMf0x/T//QE0VFDuvKhUVG68qIF+QFUEGT5EPKj+AAkpMjLH1JAyx9SMMv/UhD0AMntVPgPAdMHIcAAn2xRkyDXSpbTB9QC+wDoMOAhwAHjACHAAuMAIcAD4wABwAUPEBESAWTQMtDTAwFxsJFb4CHXScEgkVvgAdMfIYIQc2LQnL2SXwPgAvpAMAGCEHNi0Jy6kVvjDQYCASAHCAD0AdM/+gD6QNQw0PpAMHBUcAAgyMsfywHLAcsAywDJcCCCEOOg1ILIyx9QBM8WE8sAIvoCEssAzMlxcIIQD4p+pcjLHxbLP1AE+gJYzxb4KM8WE8sAghAL68IA+gLLAMzJcYAYyMsFUAPPFoIQDuaygPoCEstqzMlw+wACASAJCgBZvSQrb2omhAgKBrkPoCGEcNQICEekk30pkQzmkD6f+YN4EoAbeBAUiYcVnzGEAgFYCwwAEbjJftRNDXCx+AA9sp37UTQgQFA1yH0BDACyMoHy//J0AGBAQj0Cm+hMYAIBIA0OABmtznaiaEAga5Drhf/AABmvHfaiaEAQa5DrhY/AAG7SB/oA1NQi+QAFyMoHFcv/ydB3dIAYyMsFywIizxZQBfoCFMtrEszMyXP7AMhAFIEBCPRR8qcCAHCBAQjXGPoA0z/IVCBHgQEI9FHyp4IQbm90ZXB0gBjIywXLAlAGzxZQBPoCFMtqEssfyz/Jc/sAAgBsgQEI1xj6ANM/UjaBAQj0WfKnghBkc3RycHSAGMjLBcsCUAbPFlAE+gIUy2oSyx/LP8lz+wACALCORJMg10qOPNMH1AHQ0wX6QPoA02rTHwGCEMoGGF+6jh2CEOoGGF0FyMsFUATPFlj6AstqEssfAc8WyVj7AJQQVl8G4ugwkTDiA6TIyx8Syx/L//QAye1U
`

	_ = `te6ccgECEwEAAtQAART/APSkE/S88sgLAQIBIAIDAgFIBAUE5vKDCNcYINMf0x/THwL4I7vyZO1E0NMf0x/T//QE0VFDuvKhUVG68qIF+QFUEGT5EPKj+AAkpMjLH1JAyx9SMMv/UhD0AMntVPgPAdMHIcAAn2xRkyDXSpbTB9QC+wDoMOAhwAHjACHAAuMAIcAD4wABwAUPEBESAWTQMtDTAwFxsJFb4CHXScEgkVvgAdMfIYIQc2LQnL2SXwPgAvpAMAGCEHNi0Jy6kVvjDQYCASAHCAD0AdM/+gD6QNQw0PpAMHBUcAAgyMsfywHLAcsAywDJcCCCEOOg1ILIyx9QBM8WE8sAIvoCEssAzMlxcIIQD4p+pcjLHxbLP1AE+gJYzxb4KM8WE8sAghAL68IA+gLLAMzJcYAYyMsFUAPPFoIQDuaygPoCEstqzMlw+wACASAJCgBZvSQrb2omhAgKBrkPoCGEcNQICEekk30pkQzmkD6f+YN4EoAbeBAUiYcVnzGEAgFYCwwAEbjJftRNDXCx+AA9sp37UTQgQFA1yH0BDACyMoHy//J0AGBAQj0Cm+hMYAIBIA0OABmtznaiaEAga5Drhf/AABmvHfaiaEAQa5DrhY/AAG7SB/oA1NQi+QAFyMoHFcv/ydB3dIAYyMsFywIizxZQBfoCFMtrEszMyXP7AMhAFIEBCPRR8qcCAHCBAQjXGPoA0z/IVCBHgQEI9FHyp4IQbm90ZXB0gBjIywXLAlAGzxZQBPoCFMtqEssfyz/Jc/sAAgBsgQEI1xj6ANM/UjaBAQj0WfKnghBkc3RycHSAGMjLBcsCUAbPFlAE+gIUy2oSyx/LP8lz+wACALKORZMg10qOPdMH1CHQ0wX6QPoA02rTHwGCEMoGGF+6jh02ghDqBhhdBMjLBVADzxYB+gLLassfWM8WyVj7AJVfBQL7AOLoMJEw4gOkyMsfEssfy//0AMntVA==`

	_ = `te6ccgECEwEAAuMAART/APSkE/S88sgLAQIBIAIDAgFIBAUE5vKDCNcYINMf0x/THwL4I7vyZO1E0NMf0x/T//QE0VFDuvKhUVG68qIF+QFUEGT5EPKj+AAkpMjLH1JAyx9SMMv/UhD0AMntVPgPAdMHIcAAn2xRkyDXSpbTB9QC+wDoMOAhwAHjACHAAuMAIcAD4wABwAUPEBESAWTQMtDTAwFxsJFb4CHXScEgkVvgAdMfIYIQc2LQnL2SXwPgAvpAMAGCEHNi0Jy6kVvjDQYCASAHCAD0AdM/+gD6QNQw0PpAMHBUcAAgyMsfywHLAcsAywDJcCCCEOOg1ILIyx9QBM8WE8sAIvoCEssAzMlxcIIQD4p+pcjLHxbLP1AE+gJYzxb4KM8WE8sAghAL68IA+gLLAMzJcYAYyMsFUAPPFoIQDuaygPoCEstqzMlw+wACASAJCgBZvSQrb2omhAgKBrkPoCGEcNQICEekk30pkQzmkD6f+YN4EoAbeBAUiYcVnzGEAgFYCwwAEbjJftRNDXCx+AA9sp37UTQgQFA1yH0BDACyMoHy//J0AGBAQj0Cm+hMYAIBIA0OABmtznaiaEAga5Drhf/AABmvHfaiaEAQa5DrhY/AAG7SB/oA1NQi+QAFyMoHFcv/ydB3dIAYyMsFywIizxZQBfoCFMtrEszMyXP7AMhAFIEBCPRR8qcCAHCBAQjXGPoA0z/IVCBHgQEI9FHyp4IQbm90ZXB0gBjIywXLAlAGzxZQBPoCFMtqEssfyz/Jc/sAAgBsgQEI1xj6ANM/UjaBAQj0WfKnghBkc3RycHSAGMjLBcsCUAbPFlAE+gIUy2oSyx/LP8lz+wACANCOVJMg10qOTNMH1CHQ0wX6QPoA02og10nCH44xNgXTHwGCEMoGGF+6jh2CEOoGGF0EyMsFUAPPFgH6AhTLassfWM8WyVj7AJQQRl8G4pVfBQL7AOLoMJEw4gOkyMsfEssfy//0AMntVA==`

	_ = `te6ccgECEwEAAuMAART/APSkE/S88sgLAQIBIAIDAgFIBAUE5vKDCNcYINMf0x/THwL4I7vyZO1E0NMf0x/T//QE0VFDuvKhUVG68qIF+QFUEGT5EPKj+AAkpMjLH1JAyx9SMMv/UhD0AMntVPgPAdMHIcAAn2xRkyDXSpbTB9QC+wDoMOAhwAHjACHAAuMAIcAD4wABwAUPEBESAWTQMtDTAwFxsJFb4CHXScEgkVvgAdMfIYIQc2LQnL2SXwPgAvpAMAGCEHNi0Jy6kVvjDQYCASAHCAD0AdM/+gD6QNQw0PpAMHBUcAAgyMsfywHLAcsAywDJcCCCEOOg1ILIyx9QBM8WE8sAIvoCEssAzMlxcIIQD4p+pcjLHxbLP1AE+gJYzxb4KM8WE8sAghAL68IA+gLLAMzJcYAYyMsFUAPPFoIQDuaygPoCEstqzMlw+wACASAJCgBZvSQrb2omhAgKBrkPoCGEcNQICEekk30pkQzmkD6f+YN4EoAbeBAUiYcVnzGEAgFYCwwAEbjJftRNDXCx+AA9sp37UTQgQFA1yH0BDACyMoHy//J0AGBAQj0Cm+hMYAIBIA0OABmtznaiaEAga5Drhf/AABmvHfaiaEAQa5DrhY/AAG7SB/oA1NQi+QAFyMoHFcv/ydB3dIAYyMsFywIizxZQBfoCFMtrEszMyXP7AMhAFIEBCPRR8qcCAHCBAQjXGPoA0z/IVCBHgQEI9FHyp4IQbm90ZXB0gBjIywXLAlAGzxZQBPoCFMtqEssfyz/Jc/sAAgBsgQEI1xj6ANM/UjaBAQj0WfKnghBkc3RycHSAGMjLBcsCUAbPFlAE+gIUy2oSyx/LP8lz+wACANCOVJMg10qOTNMH1CHQ0wX6QPoA02og10nCH44xNgXTHwGCEMoGGF+6jh2CEOoGGF0EyMsFUAPPFgH6AhTLassfWM8WyVj7AJQQRl8G4pVfBQL7AOLoMJEw4gOkyMsfEssfy//0AMntVA==`
	_ = `te6ccgECEwEAAuMAART/APSkE/S88sgLAQIBIAIDAgFIBAUE5vKDCNcYINMf0x/THwL4I7vyZO1E0NMf0x/T//QE0VFDuvKhUVG68qIF+QFUEGT5EPKj+AAkpMjLH1JAyx9SMMv/UhD0AMntVPgPAdMHIcAAn2xRkyDXSpbTB9QC+wDoMOAhwAHjACHAAuMAIcAD4wABwAUPEBESAWTQMtDTAwFxsJFb4CHXScEgkVvgAdMfIYIQc2LQnL2SXwPgAvpAMAGCEHNi0Jy6kVvjDQYCASAHCAD0AdM/+gD6QNQw0PpAMHBUcAAgyMsfywHLAcsAywDJcCCCEOOg1ILIyx9QBM8WE8sAIvoCEssAzMlxcIIQD4p+pcjLHxbLP1AE+gJYzxb4KM8WE8sAghAL68IA+gLLAMzJcYAYyMsFUAPPFoIQDuaygPoCEstqzMlw+wACASAJCgBZvSQrb2omhAgKBrkPoCGEcNQICEekk30pkQzmkD6f+YN4EoAbeBAUiYcVnzGEAgFYCwwAEbjJftRNDXCx+AA9sp37UTQgQFA1yH0BDACyMoHy//J0AGBAQj0Cm+hMYAIBIA0OABmtznaiaEAga5Drhf/AABmvHfaiaEAQa5DrhY/AAG7SB/oA1NQi+QAFyMoHFcv/ydB3dIAYyMsFywIizxZQBfoCFMtrEszMyXP7AMhAFIEBCPRR8qcCAHCBAQjXGPoA0z/IVCBHgQEI9FHyp4IQbm90ZXB0gBjIywXLAlAGzxZQBPoCFMtqEssfyz/Jc/sAAgBsgQEI1xj6ANM/UjaBAQj0WfKnghBkc3RycHSAGMjLBcsCUAbPFlAE+gIUy2oSyx/LP8lz+wACANCOVJMg10qOTNMH1CHQ0wX6QPoA02og10nCH44xNgXTHwGCEMoGGF+6jh2CEOoGGF0EyMsFUAPPFgH6AhTLassfWM8WyVj7AJQQRl8G4pVfBQL7AOLoMJEw4gOkyMsfEssfy//0AMntVA==`

	v35 := `te6ccgECEwEAAuMAART/APSkE/S88sgLAQIBIAIDAgFIBAUE5vKDCNcYINMf0x/THwL4I7vyZO1E0NMf0x/T//QE0VFDuvKhUVG68qIF+QFUEGT5EPKj+AAkpMjLH1JAyx9SMMv/UhD0AMntVPgPAdMHIcAAn2xRkyDXSpbTB9QC+wDoMOAhwAHjACHAAuMAIcAD4wABwAUPEBESAWTQMtDTAwFxsJFb4CHXScEgkVvgAdMfIYIQc2LQnL2SXwPgAvpAMAGCEHNi0Jy6kVvjDQYCASAHCAD0AdM/+gD6QNQw0PpAMHBUcAAgyMsfywHLAcsAywDJcCCCEOOg1ILIyx9QBM8WE8sAIvoCEssAzMlxcIIQD4p+pcjLHxbLP1AE+gJYzxb4KM8WE8sAghAL68IA+gLLAMzJcYAYyMsFUAPPFoIQDuaygPoCEstqzMlw+wACASAJCgBZvSQrb2omhAgKBrkPoCGEcNQICEekk30pkQzmkD6f+YN4EoAbeBAUiYcVnzGEAgFYCwwAEbjJftRNDXCx+AA9sp37UTQgQFA1yH0BDACyMoHy//J0AGBAQj0Cm+hMYAIBIA0OABmtznaiaEAga5Drhf/AABmvHfaiaEAQa5DrhY/AAG7SB/oA1NQi+QAFyMoHFcv/ydB3dIAYyMsFywIizxZQBfoCFMtrEszMyXP7AMhAFIEBCPRR8qcCAHCBAQjXGPoA0z/IVCBHgQEI9FHyp4IQbm90ZXB0gBjIywXLAlAGzxZQBPoCFMtqEssfyz/Jc/sAAgBsgQEI1xj6ANM/UjaBAQj0WfKnghBkc3RycHSAGMjLBcsCUAbPFlAE+gIUy2oSyx/LP8lz+wACANCOVJMg10qOTNMH1CHQ0wX6QPoA02og10nCH44xNgXTHwGCEMoGGF+6jh2CEOoGGF0EyMsFUAPPFgH6AhTLassfWM8WyVj7AJQQRl8G4pVfBQL7AOLoMJEw4gOkyMsfEssfy//0AMntVA==`

	bytes, err := base64.StdEncoding.DecodeString(v35)
	if err != nil {
		panic(err)
	}

	code, err := cell.FromBOC(bytes)
	if err != nil {
		panic(err)
	}

	return code
}

func getGCode() *cell.Cell {
	_ = `te6ccgECEwEAAokAART/APSkE/S88sgLAQIBIAIDAgFIBAUE+PKDCNcYINMf0x/THwL4I7vyZO1E0NMf0x/T//QE0VFDuvKhUVG68qIF+QFUEGT5EPKj+AAkpMjLH1JAyx9SMMv/UhD0AMntVPgPAdMHIcAAn2xRkyDXSpbTB9QC+wDoMOAhwAHjACHAAuMAAcADkTDjDQOkyMsfEssfy/8PEBESAWTQMtDTAwFxsJFb4CHXScEgkVvgAdMfIYIQc2LQnL2SXwPgAvpAMAGCEHNi0Jy6kVvjDQYCASAHCAD0AdM/+gD6QNQw0PpAMHBUcAAgyMsfywHLAcsAywDJcCCCEOOg1ILIyx9QBM8WE8sAIvoCEssAzMlxcIIQD4p+pcjLHxbLP1AE+gJYzxb4KM8WE8sAghAL68IA+gLLAMzJcYAYyMsFUAPPFoIQDuaygPoCEstqzMlw+wACASAJCgBZvSQrb2omhAgKBrkPoCGEcNQICEekk30pkQzmkD6f+YN4EoAbeBAUiYcVnzGEAgFYCwwAEbjJftRNDXCx+AA9sp37UTQgQFA1yH0BDACyMoHy//J0AGBAQj0Cm+hMYAIBIA0OABmtznaiaEAga5Drhf/AABmvHfaiaEAQa5DrhY/AAG7SB/oA1NQi+QAFyMoHFcv/ydB3dIAYyMsFywIizxZQBfoCFMtrEszMyXP7AMhAFIEBCPRR8qcCAHCBAQjXGPoA0z/IVCBHgQEI9FHyp4IQbm90ZXB0gBjIywXLAlAGzxZQBPoCFMtqEssfyz/Jc/sAAgBsgQEI1xj6ANM/MFIkgQEI9Fnyp4IQZHN0cnB0gBjIywXLAlAFzxZQA/oCE8tqyx8Syz/Jc/sAAAr0AMntVA==
`

	_ = `te6ccgECFAEAAvcAART/APSkE/S88sgLAQIBIAIDAgFIBAUE+PKDCNcYINMf0x/THwL4I7vyZO1E0NMf0x/T//QE0VFDuvKhUVG68qIF+QFUEGT5EPKj+AAkpMjLH1JAyx9SMMv/UhD0AMntVPgPAdMHIcAAn2xRkyDXSpbTB9QC+wDoMOAhwAHjACHAAuMAAcADkTDjDQOkyMsfEssfy/8QERITAuTQMtDTAwFxsJFb4CHXScEgkVvgAdMfIYIQc2LQnL0ighDVMnbbvbAighBHT4bPvbCSXwPgAvpAMCGCEHNi0Jy6kTDjDSCCENUydtu6jhdwIIAQyMsFIfoCy2pSIMsfyz/Jgwb7AN4gghBHT4bPupFb4w0GBwIBIAgJAPoC0z/6APpA1AHQ+kAwcFRwACDIyx/LAcsBywDLAMlwIIIQ46DUgsjLH1AEzxYTywAi+gISywDMyXFwghAPin6lyMsfF8s/UAX6AlADzxb4KM8WFMsAghAL68IA+gISywDMyXGAGMjLBVAFzxaCEA7msoD6AhTLahPMyXD7AABQAdM/MdQw0PpAMfpAMHAggBDIywVQA88WIfoCEstqEssfyz/Jgwb7AAIBIAoLAFm9JCtvaiaECAoGuQ+gIYRw1AgIR6STfSmRDOaQPp/5g3gSgBt4EBSJhxWfMYQCAVgMDQARuMl+1E0NcLH4AD2ynftRNCBAUDXIfQEMALIygfL/8nQAYEBCPQKb6ExgAgEgDg8AGa3OdqJoQCBrkOuF/8AAGa8d9qJoQBBrkOuFj8AAbtIH+gDU1CL5AAXIygcVy//J0Hd0gBjIywXLAiLPFlAF+gIUy2sSzMzJc/sAyEAUgQEI9FHypwIAcIEBCNcY+gDTP8hUIEeBAQj0UfKnghBub3RlcHSAGMjLBcsCUAbPFlAE+gIUy2oSyx/LP8lz+wACAGyBAQjXGPoA0z8wUiSBAQj0WfKnghBkc3RycHSAGMjLBcsCUAXPFlAD+gITy2rLHxLLP8lz+wAACvQAye1U`

	_ = `
te6ccgECFAEAAvkAART/APSkE/S88sgLAQIBIAIDAgFIBAUE+PKDCNcYINMf0x/THwL4I7vyZO1E0NMf0x/T//QE0VFDuvKhUVG68qIF+QFUEGT5EPKj+AAkpMjLH1JAyx9SMMv/UhD0AMntVPgPAdMHIcAAn2xRkyDXSpbTB9QC+wDoMOAhwAHjACHAAuMAAcADkTDjDQOkyMsfEssfy/8QERITAujQMtDTAwFxsJFb4CHXScEgkVvgAdMfIYIQc2LQnL0ighDVMnbbvbAighBHT4bPvbCSXwPgAvpAMCGCEHNi0Jy6kTDjDSCCENUydtu6jhn4J28iMHCAEMjLBVj6AstqUhDLH8mDBvsA3iCCEEdPhs+6kVvjDQYHAgEgCAkA+gLTP/oA+kDUAdD6QDBwVHAAIMjLH8sBywHLAMsAyXAgghDjoNSCyMsfUATPFhPLACL6AhLLAMzJcXCCEA+KfqXIyx8Xyz9QBfoCUAPPFvgozxYUywCCEAvrwgD6AhLLAMzJcYAYyMsFUAXPFoIQDuaygPoCFMtqE8zJcPsAAFAB0z8x1DDQ+kAx+kAwcCCAEMjLBVADzxYh+gISy2oSyx/LP8mDBvsAAgEgCgsAWb0kK29qJoQICga5D6AhhHDUCAhHpJN9KZEM5pA+n/mDeBKAG3gQFImHFZ8xhAIBWAwNABG4yX7UTQ1wsfgAPbKd+1E0IEBQNch9AQwAsjKB8v/ydABgQEI9ApvoTGACASAODwAZrc52omhAIGuQ64X/wAAZrx32omhAEGuQ64WPwABu0gf6ANTUIvkABcjKBxXL/8nQd3SAGMjLBcsCIs8WUAX6AhTLaxLMzMlz+wDIQBSBAQj0UfKnAgBwgQEI1xj6ANM/yFQgR4EBCPRR8qeCEG5vdGVwdIAYyMsFywJQBs8WUAT6AhTLahLLH8s/yXP7AAIAbIEBCNcY+gDTPzBSJIEBCPRZ8qeCEGRzdHJwdIAYyMsFywJQBc8WUAP6AhPLassfEss/yXP7AAAK9ADJ7VQ=`

	_ = `te6ccgECFAEAAwoAART/APSkE/S88sgLAQIBIAIDAgFIBAUE+PKDCNcYINMf0x/TH/pAMQL4I7vyZO1E0NMf0x/T//QE+kDRUVS68qFRYrryogb5AVQQdvkQ8qP4ACCkyMsfUmDLH1JQy/9SQPQAIs8Wye1U+A8C0wchwACfbGGTINdKltMH1AL7AOgw4CHAAeMAIcAC4wABwAORMOMNAaQQERITAvjQMtDTAwFxsJFb4CHXScEgkVvgAdMfIYIQc2LQnL0ighDVMnbbvbAighBHT4bPvbCSXwPgAvpAMCGCEHNi0Jy6kjAx4w3tRNCBAUDXIfQEMfpAMCGCENUydtu6jh34J28iMHAggBDIywUkzxZQA/oCEstqyx/Jgwb7AN4hBgcCASAICQD2AtM/+gD6QNQw0PpAMHBUcAAgyMsfywHLAcsAywDJcCCCEOOg1ILIyx9QBM8WE8sAIvoCEssAzMlxcIIQD4p+pcjLHxbLP1AE+gJYzxb4KM8WE8sAghAL68IA+gLLAMzJcYAYyMsFUATPFoIQDuaygPoCE8tqEszJcPsAAE6CEEdPhs+6jhtwIIAQyMsFUAPPFnH6AhLLahLLH8s/yYMG+wCRW+ICASAKCwBZvSQrb2omhAgKBrkPoCGEcNQICEekk30pkQzmkD6f+YN4EoAbeBAUiYcVnzGEAgFYDA0AEbjJftRNDXCx+AA9sp37UTQgQFA1yH0BDACyMoHy//J0AGBAQj0Cm+hMYAIBIA4PABmtznaiaEAga5Drhf/AABmvHfaiaEAQa5DrhY/AAG7SB/oA1NQi+QAFyMoHFcv/ydB3dIAYyMsFywIizxZQBfoCFMtrEszMyXP7AMhAFoEBCPRR8qcEAHCBAQjXGPoA0z/IVCBJgQEI9FHyp4IQbm90ZXB0gBjIywXLAlAGzxZQBPoCFMtqEssfyz/Jc/sABABsgQEI1xj6ANM/MFImgQEI9Fnyp4IQZHN0cnB0gBjIywXLAlAFzxZQA/oCE8tqyx8Uyz/Jc/sAACLIyx8Uyx8Sy//0AAHPFsntVA==
`

	v05 := `te6ccgECFgEAAzwAART/APSkE/S88sgLAQIBIAIDAgFIBAUE+PKDCNcYINMf0x/TH/pAMQL4I7vyZO1E0NMf0x/T//QE+kDRUVS68qFRYrryogb5AVQQdvkQ8qP4ACCkyMsfUmDLH1JQy/9SQPQAIs8Wye1U+A8C0wchwACfbGGTINdKltMH1AL7AOgw4CHAAeMAIcAC4wABwAORMOMNAaQSExQVBPDQMtDTAwFxsJFb4CHXScEgkVvgAdMfIYIQc2LQnL0ighDVMnbbvbAighBHT4bPvbAighBHT4bNvbCSXwPgAvpAMCGCEHNi0Jy6kjAx4w3tRNCBAUDXIfQEMfpAMCGCENUydtu64wAhghBHT4bPuuMAIYIQR0+GzboGBwgJAgEgCgsA9gLTP/oA+kDUMND6QDBwVHAAIMjLH8sBywHLAMsAyXAgghDjoNSCyMsfUATPFhPLACL6AhLLAMzJcXCCEA+KfqXIyx8Wyz9QBPoCWM8W+CjPFhPLAIIQC+vCAPoCywDMyXGAGMjLBVAEzxaCEA7msoD6AhPLahLMyXD7AAA6+CdvIjBwIIAQyMsFJM8WUAP6AhLLassfyYMG+wAANHAggBDIywUjzxZx+gLLalIwyx/LP8mDBvsAAECOG3AggBDIywVQA88WcfoCEstqEssfyz/Jgwb7AJFb4gIBIAwNAFm9JCtvaiaECAoGuQ+gIYRw1AgIR6STfSmRDOaQPp/5g3gSgBt4EBSJhxWfMYQCAVgODwARuMl+1E0NcLH4AD2ynftRNCBAUDXIfQEMALIygfL/8nQAYEBCPQKb6ExgAgEgEBEAGa3OdqJoQCBrkOuF/8AAGa8d9qJoQBBrkOuFj8AAbtIH+gDU1CL5AAXIygcVy//J0Hd0gBjIywXLAiLPFlAF+gIUy2sSzMzJc/sAyEAWgQEI9FHypwQAcIEBCNcY+gDTP8hUIEmBAQj0UfKnghBub3RlcHSAGMjLBcsCUAbPFlAE+gIUy2oSyx/LP8lz+wAEAGyBAQjXGPoA0z8wUiaBAQj0WfKnghBkc3RycHSAGMjLBcsCUAXPFlAD+gITy2rLHxTLP8lz+wAAIsjLHxTLHxLL//QAAc8Wye1U`

	bytes, err := base64.StdEncoding.DecodeString(v05)
	if err != nil {
		panic(err)
	}

	code, err := cell.FromBOC(bytes)
	if err != nil {
		panic(err)
	}

	return code
}

func getV4Code() *cell.Cell {
	code := `B5EE9C72410214010002D4000114FF00F4A413F4BCF2C80B010201200203020148040504F8F28308D71820D31FD31FD31F02F823BBF264ED44D0D31FD31FD3FFF404D15143BAF2A15151BAF2A205F901541064F910F2A3F80024A4C8CB1F5240CB1F5230CBFF5210F400C9ED54F80F01D30721C0009F6C519320D74A96D307D402FB00E830E021C001E30021C002E30001C0039130E30D03A4C8CB1F12CB1FCBFF1011121302E6D001D0D3032171B0925F04E022D749C120925F04E002D31F218210706C7567BD22821064737472BDB0925F05E003FA403020FA4401C8CA07CBFFC9D0ED44D0810140D721F404305C810108F40A6FA131B3925F07E005D33FC8258210706C7567BA923830E30D03821064737472BA925F06E30D06070201200809007801FA00F40430F8276F2230500AA121BEF2E0508210706C7567831EB17080185004CB0526CF1658FA0219F400CB6917CB1F5260CB3F20C98040FB0006008A5004810108F45930ED44D0810140D720C801CF16F400C9ED540172B08E23821064737472831EB17080185005CB055003CF1623FA0213CB6ACB1FCB3FC98040FB00925F03E20201200A0B0059BD242B6F6A2684080A06B90FA0218470D4080847A4937D29910CE6903E9FF9837812801B7810148987159F31840201580C0D0011B8C97ED44D0D70B1F8003DB29DFB513420405035C87D010C00B23281F2FFF274006040423D029BE84C600201200E0F0019ADCE76A26840206B90EB85FFC00019AF1DF6A26840106B90EB858FC0006ED207FA00D4D422F90005C8CA0715CBFFC9D077748018C8CB05CB0222CF165005FA0214CB6B12CCCCC973FB00C84014810108F451F2A7020070810108D718FA00D33FC8542047810108F451F2A782106E6F746570748018C8CB05CB025006CF165004FA0214CB6A12CB1FCB3FC973FB0002006C810108D718FA00D33F305224810108F459F2A782106473747270748018C8CB05CB025005CF165003FA0213CB6ACB1F12CB3FC973FB00000AF400C9ED54696225E5`

	boc, err := hex.DecodeString(code)
	if err != nil {
		panic(err)
	}

	cell, err := cell.FromBOC(boc)
	if err != nil {
		panic(err)
	}

	return cell
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

func getGData(publicKey ed25519.PublicKey, botAddr *address.Address) *cell.Cell {
	dataCell := cell.BeginCell().
		MustStoreUInt(0, 32).           // Seqno
		MustStoreUInt(SubwalletID, 32). // Subwallet ID
		MustStoreSlice(publicKey, 256). // Public Key
		MustStoreDict(nil).
		MustStoreAddr(botAddr).
		EndCell()

	return dataCell
}
