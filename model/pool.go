package model

import (
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/cmingxu/dedust/utils"
	"github.com/cmingxu/dedust/wallet"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"
	"github.com/tidwall/gjson"
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/liteclient"
	"github.com/xssnick/tonutils-go/ton"
	"github.com/xssnick/tonutils-go/ton/jetton"
	"github.com/xssnick/tonutils-go/tvm/cell"
)

var (
	ValidJettonWalletHash = []string{
		"vrBoPr64kn/p/I7AoYvH3ReJlomCWhIeq0bFo6hg0M4=",
		"iUaPAseOVwgC45l5yFFvw43wfqdqSDV+BTbyuns+43s=",
		"p2DWKdU0PnbQRQF9ncIW/IoweoN3gV/rKwpcSQ5zNIY=",
		"jSjqQht36AX+pSrPM1KWSZ8Drsjp/SHdtfJWSqZcSN4=",
	}

	WalletCodeBOCs = map[string]string{
		"vrBoPr64kn/p/I7AoYvH3ReJlomCWhIeq0bFo6hg0M4=": "b5ee9c7201021101000323000114ff00f4a413f4bcf2c80b0102016202030202cc0405001ba0f605da89a1f401f481f481a8610201d40607020120080900c30831c02497c138007434c0c05c6c2544d7c0fc03383e903e900c7e800c5c75c87e800c7e800c1cea6d0000b4c7e08403e29fa954882ea54c4d167c0278208405e3514654882ea58c511100fc02b80d60841657c1ef2ea4d67c02f817c12103fcbc2000113e910c1c2ebcb853600201200a0b0083d40106b90f6a2687d007d207d206a1802698fc1080bc6a28ca9105d41083deecbef09dd0958f97162e99f98fd001809d02811e428027d012c678b00e78b6664f6aa401f1503d33ffa00fa4021f001ed44d0fa00fa40fa40d4305136a1522ac705f2e2c128c2fff2e2c254344270542013541403c85004fa0258cf1601cf16ccc922c8cb0112f400f400cb00c920f9007074c8cb02ca07cbffc9d004fa40f40431fa0020d749c200f2e2c4778018c8cb055008cf1670fa0217cb6b13cc80c0201200d0e009e8210178d4519c8cb1f19cb3f5007fa0222cf165006cf1625fa025003cf16c95005cc2391729171e25008a813a08209c9c380a014bcf2e2c504c98040fb001023c85004fa0258cf1601cf16ccc9ed5402f73b51343e803e903e90350c0234cffe80145468017e903e9014d6f1c1551cdb5c150804d50500f214013e809633c58073c5b33248b232c044bd003d0032c0327e401c1d3232c0b281f2fff274140371c1472c7cb8b0c2be80146a2860822625a019ad822860822625a028062849e5c412440e0dd7c138c34975c2c0600f1000d73b51343e803e903e90350c01f4cffe803e900c145468549271c17cb8b049f0bffcb8b08160824c4b402805af3cb8b0e0841ef765f7b232c7c572cfd400fe8088b3c58073c5b25c60063232c14933c59c3e80b2dab33260103ec01004f214013e809633c58073c5b3327b552000705279a018a182107362d09cc8cb1f5230cb3f58fa025007cf165007cf16c9718010c8cb0524cf165006fa0215cb6a14ccc971fb0010241023007cc30023c200b08e218210d53276db708010c8cb055008cf165004fa0216cb6a12cb1f12cb3fc972fb0093356c21e203c85004fa0258cf1601cf16ccc9ed54",
		"iUaPAseOVwgC45l5yFFvw43wfqdqSDV+BTbyuns+43s=": "b5ee9c7201021101000323000114ff00f4a413f4bcf2c80b0102016202030202cc0405001ba0f605da89a1f401f481f481a8610201d40607020120080900c30831c02497c138007434c0c05c6c2544d7c0fc03383e903e900c7e800c5c75c87e800c7e800c1cea6d0000b4c7e08403e29fa954882ea54c4d167c0278208405e3514654882ea58c511100fc02b80d60841657c1ef2ea4d67c02f817c12103fcbc2000113e910c1c2ebcb853600201200a0b0083d40106b90f6a2687d007d207d206a1802698fc1080bc6a28ca9105d41083deecbef09dd0958f97162e99f98fd001809d02811e428027d012c678b00e78b6664f6aa401f1503d33ffa00fa4021f001ed44d0fa00fa40fa40d4305136a1522ac705f2e2c128c2fff2e2c254344270542013541403c85004fa0258cf1601cf16ccc922c8cb0112f400f400cb00c920f9007074c8cb02ca07cbffc9d004fa40f40431fa0020d749c200f2e2c4778018c8cb055008cf1670fa0217cb6b13cc80c0201200d0e009e8210178d4519c8cb1f19cb3f5007fa0222cf165006cf1625fa025003cf16c95005cc2391729171e25008a813a08209c9c380a014bcf2e2c504c98040fb001023c85004fa0258cf1601cf16ccc9ed5402f73b51343e803e903e90350c0234cffe80145468017e903e9014d6f1c1551cdb5c150804d50500f214013e809633c58073c5b33248b232c044bd003d0032c0327e401c1d3232c0b281f2fff274140371c1472c7cb8b0c2be80146a2860822625a019ad822860822625a028062849e5c412440e0dd7c138c34975c2c0600f1000d73b51343e803e903e90350c01f4cffe803e900c145468549271c17cb8b049f0bffcb8b08160824c4b402805af3cb8b0e0841ef765f7b232c7c572cfd400fe8088b3c58073c5b25c60063232c14933c59c3e80b2dab33260103ec01004f214013e809633c58073c5b3327b552000705279a018a182107362d09cc8cb1f5230cb3f58fa025007cf165007cf16c9718010c8cb0524cf165006fa0215cb6a14ccc971fb0010241023007cc30023c200b08e218210d53276db708010c8cb055008cf165004fa0216cb6a12cb1f12cb3fc972fb0093356c21e203c85004fa0258cf1601cf16ccc9ed54",
		"p2DWKdU0PnbQRQF9ncIW/IoweoN3gV/rKwpcSQ5zNIY=": "b5ee9c7201021201000334000114ff00f4a413f4bcf2c80b0102016202030202cc0405001ba0f605da89a1f401f481f481a8610201d40607020148080900c30831c02497c138007434c0c05c6c2544d7c0fc02f83e903e900c7e800c5c75c87e800c7e800c1cea6d0000b4c7e08403e29fa954882ea54c4d167c0238208405e3514654882ea58c511100fc02780d60841657c1ef2ea4d67c02b817c12103fcbc2000113e910c1c2ebcb853600201200a0b020120101101f100f4cffe803e90087c007b51343e803e903e90350c144da8548ab1c17cb8b04a30bffcb8b0950d109c150804d50500f214013e809633c58073c5b33248b232c044bd003d0032c032483e401c1d3232c0b281f2fff274013e903d010c7e800835d270803cb8b11de0063232c1540233c59c3e8085f2dac4f3200c03f73b51343e803e903e90350c0234cffe80145468017e903e9014d6f1c1551cdb5c150804d50500f214013e809633c58073c5b33248b232c044bd003d0032c0327e401c1d3232c0b281f2fff274140371c1472c7cb8b0c2be80146a2860822625a020822625a004ad8228608239387028062849f8c3c975c2c070c008e00d0e0f00ae8210178d4519c8cb1f19cb3f5007fa0222cf165006cf1625fa025003cf16c95005cc2391729171e25008a813a08208e4e1c0aa008208989680a0a014bcf2e2c504c98040fb001023c85004fa0258cf1601cf16ccc9ed5400705279a018a182107362d09cc8cb1f5230cb3f58fa025007cf165007cf16c9718010c8cb0524cf165006fa0215cb6a14ccc971fb0010241023000e10491038375f040076c200b08e218210d53276db708010c8cb055008cf165004fa0216cb6a12cb1f12cb3fc972fb0093356c21e203c85004fa0258cf1601cf16ccc9ed5400db3b51343e803e903e90350c01f4cffe803e900c145468549271c17cb8b049f0bffcb8b0a0823938702a8005a805af3cb8b0e0841ef765f7b232c7c572cfd400fe8088b3c58073c5b25c60063232c14933c59c3e80b2dab33260103ec01004f214013e809633c58073c5b3327b55200083200835c87b51343e803e903e90350c0134c7e08405e3514654882ea0841ef765f784ee84ac7cb8b174cfcc7e800c04e81408f214013e809633c58073c5b3327b5520",
		"jSjqQht36AX+pSrPM1KWSZ8Drsjp/SHdtfJWSqZcSN4=": "b5ee9c7201010101002300084202ba2918c8947e9b25af9ac1b883357754173e5812f807a3d6e642a14709595395",
	}
)

const PoolCreationDDL = `
		CREATE TABLE IF NOT EXISTS pools (
			id int PRIMARY KEY AUTO_INCREMENT,
			address varchar(255),
			lt BIGINT,
			totalSupply varchar(255),
			type varchar(255),
			tradeFee Decimal(5, 2),
			asset0Address VARCHAR(255),
			asset1Address VARCHAR(255),
			asset0Type VARCHAR(255),
			asset1Type VARCHAR(255),
			asset0Name VARCHAR(255),
			asset1Name VARCHAR(255),
			asset0Symbol VARCHAR(255),
			asset1Symbol VARCHAR(255),
			asset0Decimal INTEGER,
			asset1Decimal INTEGER,
			lastPrice BIGINT,
			asset0Image VARCHAR(255),
			asset1Image VARCHAR(255),
			asset0Reserve varchar(255),
			asset1Reserve varchar(255),
			asset0Vault varchar(255),
			asset0Code text,
			asset0TokenWalletCode text,

			asset1Code text,
			asset1TokenWalletCode text,
			jettonWalletCode text,
			asset1Vault varchar(255),
			asset1VaultJettonWalletAddress varchar(255),
			privateKeyOfG text,
			gAddr varchar(255),
			createdAt timestamp DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
		);
`

var (
	ReseveLenForTen         = 9 + 1
	ReseveLenForHundrends   = 9 + 2
	ReseveLenForThousands   = 9 + 3
	ReseveLenFor10Thousands = 9 + 4
)

const PoolAddIndex = `
ALTER TABLE pools ADD INDEX address_index(address)
	`

type PoolType string

const (
	Volatile PoolType = "volatile"
	Stable   PoolType = "stable"
)

type AssetType string

const (
	Native AssetType = "native"
	Jetton AssetType = "jetton"
)

type Pool struct {
	ID            int       `json:"id" db:"id"`
	Address       string    `json:"address" db:"address"`
	Lt            uint64    `json:"lt" db:"lt"`
	TotalSupply   string    `json:"totalSupply" db:"totalSupply"`
	Type          PoolType  `json:"type" db:"type"`
	TradeFee      float32   `json:"tradeFee" db:"tradeFee"`
	Asset0Address string    `json:"asset0Address" db:"asset0Address"`
	Asset1Address string    `json:"asset1Address" db:"asset1Address"`
	Asset0Type    AssetType `json:"asset0Type" db:"asset0Type"`
	Asset1Type    AssetType `json:"asset1Type" db:"asset1Type"`
	Asset0Name    string    `json:"asset0Name" db:"asset0Name"`
	Asset1Name    string    `json:"asset1Name" db:"asset1Name"`
	Asset0Symbol  string    `json:"asset0Symbol" db:"asset0Symbol"`
	Asset1Symbol  string    `json:"asset1Symbol" db:"asset1Symbol"`
	Asset0Decimal uint8     `json:"asset0Decimal" db:"asset0Decimal"`
	Asset1Decimal uint8     `json:"asset1Decimal" db:"asset1Decimal"`
	LastPrice     uint64    `json:"lastPrice" db:"lastPrice"`
	Asset0Image   string    `json:"asset0Image" db:"asset0Image"`
	Asset1Image   string    `json:"asset1Image" db:"asset1Image"`
	Asset0Reserve string    `json:"asset0Reserve" db:"asset0Reserve"`
	Asset1Reserve string    `json:"asset1Reserve" db:"asset1Reserve"`

	Asset0Code            string         `json:"asset0Code" db:"asset0Code"`
	Asset0TokenWalletCode string         `json:"asset0TokenWalletCode" db:"asset0TokenWalletCode"`
	Asset1Code            string         `json:"asset1Code" db:"asset1Code"`
	Asset1TokenWalletCode string         `json:"asset1TokenWalletCode" db:"asset1TokenWalletCode"`
	JettonWalletCode      sql.NullString `json:"jettonWalletCode" db:"jettonWalletCode"`

	Asset0Vault string `json:"asset0Vault" db:"asset0Vault"`
	Asset1Vault string `json:"asset1Vault" db:"asset1Vault"`

	// Vault 对应的 jetton wallet 地址
	Asset1VaultJettonWalletAddress sql.NullString `json:"asset1VaultJettonWalletAddress" db:"asset1VaultJettonWalletAddress"`
	PrivateKeyOfG                  sql.NullString `json:"privateKeyOfG" db:"privateKeyOfG"`
	GAddr                          sql.NullString `json:"gAddr" db:"gAddr"`

	CreatedAt *time.Time `json:"createdAt" db:"createdAt"`

	UpdatedAt time.Time `json:"-" db:"-"`

	Batch int32 `json:"-" db:"-"`
}

var (
	DedustFactory = address.MustParseAddr("EQBfBWT7X2BHg9tXAxzhz2aKiNTU1tpt5NsiK0uSDW_YAJ67")
)

func CreateTableIfNotExists(db *sqlx.DB) error {
	_, err := db.Exec(PoolCreationDDL)
	if err != nil {
		return err
	}

	_, err = db.Exec(PoolAddIndex)

	return err
}

func LoadPoolsFromJSON(data []byte) ([]*Pool, error) {
	pools := make([]*Pool, 0)
	result := gjson.Parse(string(data))

	result.ForEach(func(key, value gjson.Result) bool {
		pool := Pool{}

		pool.Address = value.Get("address").String()
		pool.Lt = value.Get("lt").Uint()
		pool.TotalSupply = value.Get("totalSupply").String()
		pool.Type = PoolType(value.Get("type").String())
		pool.TradeFee = float32(value.Get("tradeFee").Float())
		pool.Asset0Address = value.Get("assets.0.address").String()
		pool.Asset0Type = AssetType(value.Get("assets.0.type").String())
		pool.Asset0Name = value.Get("assets.0.metadata.name").String()
		pool.Asset0Symbol = value.Get("assets.0.metadata.symbol").String()
		pool.Asset0Image = value.Get("assets.0.metadata.image").String()
		pool.Asset0Decimal = uint8(value.Get("assets.0.metadata.decimals").Uint())
		pool.Asset0Reserve = value.Get("reserves.0").String()

		pool.Asset1Address = value.Get("assets.1.address").String()
		pool.Asset1Type = AssetType(value.Get("assets.1.type").String())
		pool.Asset1Name = value.Get("assets.1.metadata.name").String()
		pool.Asset1Symbol = value.Get("assets.1.metadata.symbol").String()
		pool.Asset1Image = value.Get("assets.1.metadata.image").String()
		pool.Asset1Decimal = uint8(value.Get("assets.1.metadata.decimals").Uint())
		pool.Asset1Reserve = value.Get("reserves.1").String()

		pools = append(pools, &pool)
		return true
	})

	return pools, nil
}

func LoadPoolsFromDB(db *sqlx.DB, outstandingOnly bool) ([]*Pool, error) {
	pools := make([]*Pool, 0)
	if outstandingOnly {
		statement := fmt.Sprintf("SELECT * FROM pools WHERE asset0Type = 'native' and LENGTH(asset0Reserve) >= %d AND asset1TokenWalletCode in (?)", ReseveLenForTen)
		query, args, err := sqlx.In(statement, ValidJettonWalletHash)
		if err != nil {
			return nil, err
		}

		query = db.Rebind(query)
		err = db.Select(&pools, query, args...)
		return pools, err
	}

	err := db.Select(&pools, "SELECT * FROM pools")
	return pools, err
}

func (p *Pool) FetchAssetMasterCode(ctx context.Context,
	conn *liteclient.ConnectionPool,
	masterBlock *ton.BlockIDExt,
) error {

	client := utils.GetAPIClient(conn)

	addr0, err := address.ParseAddr(p.Asset0Address)
	if err == nil {
		acc, err := client.GetAccount(ctx, masterBlock, addr0)
		if err != nil {
			return err
		}

		if acc.Code == nil {
			return fmt.Errorf("asset0 code is nil")
		}

		p.Asset0Code = base64.StdEncoding.EncodeToString(acc.Code.Hash())
	}

	acc, err := client.GetAccount(ctx, masterBlock, address.MustParseAddr(p.Asset1Address))
	if err != nil {
		return err
	}

	if acc.Code == nil {
		return fmt.Errorf("asset1 code is nil")
	}

	p.Asset1Code = base64.StdEncoding.EncodeToString(acc.Code.Hash())

	return nil
}

func (p *Pool) FetchVaultAddress(ctx context.Context,
	client ton.APIClientWrapped,
	masterBlock *ton.BlockIDExt) error {

	if len(p.Asset0Address) != 0 {
		addr0 := address.MustParseAddr(p.Asset0Address)

		c := cell.BeginCell().
			MustStoreUInt(1, 4).
			MustStoreInt(int64(addr0.Workchain()), 8).
			MustStoreBinarySnake(addr0.Data())

		stack, err := client.RunGetMethod(ctx, masterBlock,
			DedustFactory,
			"get_vault_address", c.EndCell().BeginParse())
		if err != nil {
			return err
		}

		vaultAddrCell, err := stack.Slice(0)
		if err != nil {
			return err
		}

		vaultAddr, err := vaultAddrCell.LoadAddr()
		if err != nil {
			return err
		}
		p.Asset0Vault = vaultAddr.String()
		log.Debug().Msgf("pool %s, asset0addr: %s, vault0 addr: %s",
			p.Address, p.Asset0Address, vaultAddr.String())
	}

	if len(p.Asset1Address) != 0 {
		addr1 := address.MustParseAddr(p.Asset1Address)
		c := cell.BeginCell().
			MustStoreUInt(1, 4).
			MustStoreInt(int64(addr1.Workchain()), 8).
			MustStoreBinarySnake(addr1.Data())

		stack1, err := client.RunGetMethod(ctx, masterBlock,
			DedustFactory,
			"get_vault_address", c.EndCell().BeginParse())
		if err != nil {
			return err
		}

		vault1AddrCell, err := stack1.Slice(0)
		if err != nil {
			return err
		}

		vault1Addr, err := vault1AddrCell.LoadAddr()
		if err != nil {
			return err
		}
		p.Asset1Vault = vault1Addr.String()
		log.Debug().Msgf("pool %s, asset1addr: %s, vault1 addr: %s",
			p.Address, p.Asset1Address, vault1Addr.String())
	}

	return nil
}

func (p *Pool) FetchAssetWalletCode(
	ctx context.Context,
	client ton.APIClientWrapped,
	masterBlock *ton.BlockIDExt,
) error {
	if len(p.Asset0Address) != 0 && len(p.Asset1Vault) != 0 {
		addr := address.MustParseAddr(p.Asset0Address)
		master := jetton.NewJettonMasterClient(client, addr)

		tokenWallet, err := master.GetJettonWallet(ctx, address.MustParseAddr(p.Asset0Vault))
		if err != nil {
			return err
		}

		acc, err := client.GetAccount(ctx, masterBlock, tokenWallet.Address())
		if err != nil {
			return err
		}

		if acc.Code == nil {
			log.Debug().Msgf("asset0 token wallet %s code is nil", tokenWallet.Address().String())
			goto NEXT
		}

		p.Asset0TokenWalletCode = base64.StdEncoding.EncodeToString(acc.Code.Hash())
		log.Debug().Msgf("pool %s, asset0addr: %s, token wallet %s code: %s",
			p.Address, p.Asset0Address, tokenWallet.Address().String(), p.Asset0TokenWalletCode)
	}

NEXT:

	if len(p.Asset1Address) != 0 && len(p.Asset1Vault) != 0 {
		addr := address.MustParseAddr(p.Asset1Address)
		master := jetton.NewJettonMasterClient(client, addr)

		tokenWallet, err := master.GetJettonWallet(ctx, address.MustParseAddr(p.Asset1Vault))
		if err != nil {
			return err
		}

		acc, err := client.GetAccount(ctx, masterBlock, tokenWallet.Address())
		if err != nil {
			return err
		}

		if acc.Code == nil {
			return fmt.Errorf("asset1 token wallet code is nil")
		}

		p.JettonWalletCode = sql.NullString{Valid: true, String: base64.StdEncoding.EncodeToString(acc.Code.ToBOC())}
		p.Asset1TokenWalletCode = base64.StdEncoding.EncodeToString(acc.Code.Hash())
		log.Debug().Msgf("pool %s, asset1addr: %s, token wallet %s code: %s",
			p.Address, p.Asset1Address, tokenWallet.Address().String(), p.Asset1TokenWalletCode)
	}

	return nil
}

func (p *Pool) GenerateVault1JettonWalletAddress() error {
	if len(p.Asset1TokenWalletCode) == 0 {
		return fmt.Errorf("asset1 token wallet code is empty")
	}

	if len(p.Asset1Vault) == 0 {
		return fmt.Errorf("asset1 vault address is empty")
	}

	if len(p.Asset1Address) == 0 {
		return fmt.Errorf("asset1 address is empty")
	}

	vault1Addr := address.MustParseAddr(p.Asset1Vault)
	asset1JettonMasterAddr := address.MustParseAddr(p.Asset1Address)
	code, ok := WalletCodeBOCs[p.Asset1TokenWalletCode]
	if !ok {
		return fmt.Errorf("asset1 token wallet code not found")
	}

	content, _ := hex.DecodeString(code)
	codeCell, _ := cell.FromBOC(content)

	jettonWalletCell := wallet.CalculateUserJettonWalletAddress(
		vault1Addr,
		asset1JettonMasterAddr,
		codeCell,
	)

	p.Asset1VaultJettonWalletAddress = sql.NullString{
		String: utils.CellToAddress(jettonWalletCell).String(),
		Valid:  true,
	}

	return nil
}

func (p *Pool) ExistsInDB(db *sqlx.DB) (bool, error) {
	row, err := db.Query("SELECT * FROM pools WHERE address = ?", p.Address)
	if err != nil {
		return false, err
	}

	if row.Next() {
		row.Close()
		return true, nil
	}
	row.Close()

	return false, nil
}

func (p *Pool) SaveToDB(db *sqlx.DB) error {
	fmt.Println("saving pool", p.Address)
	row, err := db.Query("SELECT * FROM pools WHERE address = ?", p.Address)
	if err != nil {
		return err
	}

	if row.Next() {
		row.Close()
		return nil
	}
	row.Close()

	row1, err := db.NamedQuery("INSERT INTO pools (address, lt, totalSupply, type, tradeFee, asset0Address, asset1Address, asset0Type, asset1Type, asset0Name, asset1Name, asset0Symbol, asset1Symbol, asset0Decimal, asset1Decimal, lastPrice, asset0Image, asset1Image, asset0Reserve, asset1Reserve, asset0Code, asset1Code, asset0TokenWalletCode, asset1TokenWalletCode, asset0Vault, asset1Vault, jettonWalletCode, asset1VaultJettonWalletAddress, privateKeyOfG, gAddr) VALUES (:address, :lt, :totalSupply, :type, :tradeFee, :asset0Address, :asset1Address, :asset0Type, :asset1Type, :asset0Name, :asset1Name, :asset0Symbol, :asset1Symbol, :asset0Decimal, :asset1Decimal, :lastPrice, :asset0Image, :asset1Image, :asset0Reserve, :asset1Reserve, :asset0Code, :asset1Code, :asset0TokenWalletCode, :asset1TokenWalletCode, :asset0Vault, :asset1Vault, :jettonWalletCode, :asset1VaultJettonWalletAddress, :privateKeyOfG, :gAddr)", p)
	if err != nil {
		return err
	}

	return row1.Close()
}

func (p *Pool) UpdateReserves(db *sqlx.DB) error {
	row, err := db.NamedQuery("UPDATE pools SET asset0Reserve = :asset0Reserve, asset1Reserve = :asset1Reserve, lt = :lt WHERE address = :address", p)
	if err != nil {
		return err
	}

	return row.Close()
}

func (p *Pool) UpdateG(db *sqlx.DB) error {
	p.Asset1VaultJettonWalletAddress.Valid = true
	p.PrivateKeyOfG.Valid = true
	p.GAddr.Valid = true

	row, err := db.NamedQuery("UPDATE pools SET asset1VaultJettonWalletAddress = :asset1VaultJettonWalletAddress, privateKeyOfG = :privateKeyOfG, gAddr = :gAddr WHERE address = :address", p)
	if err != nil {
		return err
	}

	return row.Close()
}
