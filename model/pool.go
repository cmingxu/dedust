package model

import (
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/tidwall/gjson"
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
	ID            int        `json:"id" db:"id"`
	Address       string     `json:"address" db:"address"`
	Lt            uint64     `json:"lt" db:"lt"`
	TotalSupply   string     `json:"totalSupply" db:"totalSupply"`
	Type          PoolType   `json:"type" db:"type"`
	TradeFee      float32    `json:"tradeFee" db:"tradeFee"`
	Asset0Address string     `json:"asset0Address" db:"asset0Address"`
	Asset1Address string     `json:"asset1Address" db:"asset1Address"`
	Asset0Type    AssetType  `json:"asset0Type" db:"asset0Type"`
	Asset1Type    AssetType  `json:"asset1Type" db:"asset1Type"`
	Asset0Name    string     `json:"asset0Name" db:"asset0Name"`
	Asset1Name    string     `json:"asset1Name" db:"asset1Name"`
	Asset0Symbol  string     `json:"asset0Symbol" db:"asset0Symbol"`
	Asset1Symbol  string     `json:"asset1Symbol" db:"asset1Symbol"`
	Asset0Decimal uint8      `json:"asset0Decimal" db:"asset0Decimal"`
	Asset1Decimal uint8      `json:"asset1Decimal" db:"asset1Decimal"`
	LastPrice     uint64     `json:"lastPrice" db:"lastPrice"`
	Asset0Image   string     `json:"asset0Image" db:"asset0Image"`
	Asset1Image   string     `json:"asset1Image" db:"asset1Image"`
	Asset0Reseve  string     `json:"asset0Reserve" db:"asset0Reserve"`
	Asset1Reseve  string     `json:"asset1Reserve" db:"asset1Reserve"`
	CreatedAt     *time.Time `json:"createdAt" db:"createdAt"`

	Batch int32 `json:"-" db:"-"`
}

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
		pool.Asset0Reseve = value.Get("reserves.0").String()

		pool.Asset1Address = value.Get("assets.1.address").String()
		pool.Asset1Type = AssetType(value.Get("assets.1.type").String())
		pool.Asset1Name = value.Get("assets.1.metadata.name").String()
		pool.Asset1Symbol = value.Get("assets.1.metadata.symbol").String()
		pool.Asset1Image = value.Get("assets.1.metadata.image").String()
		pool.Asset1Decimal = uint8(value.Get("assets.1.metadata.decimals").Uint())
		pool.Asset1Reseve = value.Get("reserves.1").String()

		pools = append(pools, &pool)
		return true
	})

	return pools, nil
}

func LoadPoolsFromDB(db *sqlx.DB, outstandingOnly bool) ([]*Pool, error) {
	pools := make([]*Pool, 0)
	if outstandingOnly {
		err := db.Select(&pools, fmt.Sprintf("SELECT * FROM pools WHERE asset0Type = 'native' and LENGTH(asset0Reserve) >= %d", ReseveLenForHundrends))
		return pools, err
	}

	err := db.Select(&pools, "SELECT * FROM pools")
	return pools, err
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

	row1, err := db.NamedQuery("INSERT INTO pools (address, lt, totalSupply, type, tradeFee, asset0Address, asset1Address, asset0Type, asset1Type, asset0Name, asset1Name, asset0Symbol, asset1Symbol, asset0Decimal, asset1Decimal, lastPrice, asset0Image, asset1Image, asset0Reserve, asset1Reserve) VALUES (:address, :lt, :totalSupply, :type, :tradeFee, :asset0Address, :asset1Address, :asset0Type, :asset1Type, :asset0Name, :asset1Name, :asset0Symbol, :asset1Symbol, :asset0Decimal, :asset1Decimal, :lastPrice, :asset0Image, :asset1Image, :asset0Reserve, :asset1Reserve)", p)
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
