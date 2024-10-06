package model

import (
	"context"
	"fmt"
	"time"

	"github.com/cmingxu/dedust/utils"
	"github.com/jmoiron/sqlx"
	"github.com/tidwall/gjson"
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/tlb"
	"github.com/xssnick/tonutils-go/ton"
)

var (
	ValidJettonWalletHash = []string{
		"vrBoPr64kn/p/I7AoYvH3ReJlomCWhIeq0bFo6hg0M4=",
		"iUaPAseOVwgC45l5yFFvw43wfqdqSDV+BTbyuns+43s=",
		"eqG3vmgENk7aC/453lGY0t2R9O7/XrVy7gSz6mqogdk=",
		"p2DWKdU0PnbQRQF9ncIW/IoweoN3gV/rKwpcSQ5zNIY=",
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
			asset1Vault varchar(255),
			createdAt timestamp DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
		);
`

var (
	ReseveLenForThousands   = 9 + 3
	ReseveLenForHundrends   = 9 + 2
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

	Asset0Code            string `json:"asset0Code" db:"asset0Code"`
	Asset0TokenWalletCode string `json:"asset0TokenWalletCode" db:"asset0TokenWalletCode"`
	Asset1Code            string `json:"asset1Code" db:"asset1Code"`
	Asset1TokenWalletCode string `json:"asset1TokenWalletCode" db:"asset1TokenWalletCode"`

	Asset0Vault string `json:"asset0Vault" db:"asset0Vault"`
	Asset1Vault string `json:"asset1Vault" db:"asset1Vault"`

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
		statement := fmt.Sprintf("SELECT * FROM pools WHERE asset0Type = 'native' and LENGTH(asset0Reserve) >= %d AND asset1TokenWalletCode in (?)", ReseveLenForHundrends)
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

func (p *Pool) FetchAssetCode() error {
	var err error

	if len(p.Asset0Address) != 0 {
		p.Asset0Code, err = fetchJettonMasterCodeHash(p.Asset0Address)
		if err != nil {
			return err
		}
	}

	if len(p.Asset1Address) != 0 {
		p.Asset1Code, err = fetchJettonMasterCodeHash(p.Asset1Address)
		if err != nil {
			return err
		}
	}

	if len(p.Asset0Address) != 0 {
		p.Asset0TokenWalletCode, err = fetchJettonWalletCodeHash(p.Asset0Address)
		if err != nil {
			return err
		}
	}

	if len(p.Asset1Address) != 0 {
		p.Asset1TokenWalletCode, err = fetchJettonWalletCodeHash(p.Asset1Address)
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *Pool) FetchVaultAddress(ctx context.Context, client ton.APIClientWrapped,
	masterBlock *ton.BlockIDExt) error {

	type Req struct {
		_    tlb.Magic        `tlb:"$0001"`
		Addr *address.Address `tlb:"addr"`
	}
	if len(p.Asset0Address) == 0 {
		return nil
	}

	req := Req{
		Addr: address.MustParseAddr(p.Asset0Address),
	}

	c, _ := tlb.ToCell(req)
	stack, err := client.RunGetMethod(ctx, masterBlock,
		DedustFactory,
		"get_vault_address", c.BeginParse())
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
	fmt.Println("vault0 addess", p.Asset0Vault)

	req1 := Req{
		Addr: address.MustParseAddr(p.Asset1Address),
	}

	c, _ = tlb.ToCell(req1)
	stack1, err := client.RunGetMethod(ctx, masterBlock,
		DedustFactory,
		"get_vault_address", c.BeginParse())
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
	fmt.Println("vault1 addess", p.Asset1Vault)

	return nil
}

func fetchJettonMasterCodeHash(accountId string) (string, error) {
	url := fmt.Sprintf("http://49.12.81.26:8080/api/v0/accounts?address=%s&latest=true", accountId)
	antonUrl := fmt.Sprintf("https://anton.tools/api/v0/accounts?address=%s&latest=true", accountId)
	var resp []byte
	var err error
	resp, err = utils.Request(context.Background(), "GET", url, nil)
	if err == nil && len(resp) > 100 {
		goto GOT
	}

	resp, err = utils.Request(context.Background(), "GET", antonUrl, nil)
	if err != nil {
		return "", err
	}

GOT:

	return gjson.Get(string(resp), "results.0.code_hash").String(), nil
}

func fetchJettonWalletCodeHash(accountId string) (string, error) {
	url := fmt.Sprintf("http://49.12.81.26:8080/api/v0/accounts?interface=jetton_wallet&minter_address=%s&limit=1", accountId)
	antonUrl := fmt.Sprintf("https://anton.tools/api/v0/accounts?interface=jetton_wallet&minter_address=%s&limit=1", accountId)
	var resp []byte
	var err error
	resp, err = utils.Request(context.Background(), "GET", url, nil)
	if err == nil && len(resp) > 100 {
		goto GOT
	}

	resp, err = utils.Request(context.Background(), "GET", antonUrl, nil)
	if err != nil {
		return "", err
	}

GOT:
	return gjson.Get(string(resp), "results.0.code_hash").String(), nil
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
	row, err := db.Query("SELECT * FROM pools WHERE address = ?", p.Address)
	if err != nil {
		return err
	}

	if row.Next() {
		row.Close()
		return nil
	}
	row.Close()

	row1, err := db.NamedQuery("INSERT INTO pools (address, lt, totalSupply, type, tradeFee, asset0Address, asset1Address, asset0Type, asset1Type, asset0Name, asset1Name, asset0Symbol, asset1Symbol, asset0Decimal, asset1Decimal, lastPrice, asset0Image, asset1Image, asset0Reserve, asset1Reserve, asset0Code, asset1Code, asset0TokenWalletCode, asset1TokenWalletCode, asset0Vault, asset1Vault) VALUES (:address, :lt, :totalSupply, :type, :tradeFee, :asset0Address, :asset1Address, :asset0Type, :asset1Type, :asset0Name, :asset1Name, :asset0Symbol, :asset1Symbol, :asset0Decimal, :asset1Decimal, :lastPrice, :asset0Image, :asset1Image, :asset0Reserve, :asset1Reserve, :asset0Code, :asset1Code, :asset0TokenWalletCode, :asset1TokenWalletCode, :asset0Vault, :asset1Vault)", p)
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
