package model

import (
	"bytes"
	"encoding/json"
	"time"

	"github.com/jmoiron/sqlx"
)

const TradeDDL = `
CREATE TABLE IF NOT EXISTS trades (
id BIGINT PRIMARY KEY AUTO_INCREMENT,
hash VARCHAR(64) NOT NULL,
walletType VARCHAR(16) NOT NULL,
swapType VARCHAR(16) NOT NULL,
tradeType VARCHAR(16) NOT NULL,
address VARCHAR(128) NOT NULL,
poolAddr VARCHAR(128) NOT NULL,
amountIn VARCHAR(128) NOT NULL,
boc TEXT NOT NULL,
amount VARCHAR(128) NOT NULL,
tokenAmount VARCHAR(128) NOT NULL,
limits TEXT NOT NULL,
recipient VARCHAR(128) NOT NULL,
referrer VARCHAR(128) NOT NULL,
fullfillBOC TEXT NOT NULL,
rejectBOC TEXT NOT NULL,
lastestReserve0 VARCHAR(256) NOT NULL,
lastestReserve1 VARCHAR(256) NOT NULL,
latestPoolLt VARCHAR(64) NOT NULL,
createdAt timestamp DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);
`

const HashTradeAddIndex = `
CREATE INDEX hash_index ON trades (hash);
`
const AddressTradeAddIndex = `
CREATE INDEX address_index ON trades (address);
`

const PoolAddrTradeIndex = `
CREATE INDEX pool_addr_index ON trades (poolAddr);
`

type WalletType string

const (
	WalletTypeV4R2 WalletType = "v4r2" // and version before
	WalletTypeV5R1 WalletType = "v5r1"
	WalletTypeV3   WalletType = "v3"
	WalletTypeBot  WalletType = "bot"
)

type SwapType string

const (
	SwapTypeNative SwapType = "native"
	SwapTypeJetton SwapType = "jetton"
)

type TradeType string

const (
	TradeTypeSell TradeType = "sell"
	TradeTypeBuy  TradeType = "buy"
)

type Trade struct {
	Id             uint64     `json:"id" db:"id"`
	Hash           string     `json:"hash" db:"hash"`
	WalletType     WalletType `json:"walletType" db:"walletType"`
	SwapType       SwapType   `json:"swapType" db:"swapType"`
	TradeType      TradeType  `json:"tradeType" db:"tradeType"`
	Address        string     `json:"address" db:"address"`
	PoolAddr       string     `json:"poolAddr" db:"poolAddr"`
	AmountIn       string     `json:"amountIn" db:"amountIn"`
	Boc            string     `json:"boc" db:"boc"`
	Amount         string     `json:"amount" db:"amount"`
	TokenAmount    string     `json:"tokenAmount" db:"tokenAmount"`
	Limit          string     `json:"limit" db:"limits"`
	Recipient      string     `json:"recipient" db:"recipient"`
	Referrer       string     `json:"referrer" db:"referrer"`
	FullfillBOC    string     `json:"fullfillBOC" db:"fullfillBOC"`
	RejectBOC      string     `json:"rejectBOC" db:"rejectBOC"`
	LatestReserve0 string     `json:"latestReserve0" db:"lastestReserve0"`
	LatestReserve1 string     `json:"latestReserve1" db:"lastestReserve1"`
	LatestPoolLt   uint64     `json:"latestPool" db:"latestPoolLt"`

	CreatedAt *time.Time `json:"createdAt" db:"createdAt"`

	PoolUpdateAt       time.Time `json:"-" db:"-"`
	HasNextStep        bool      `json:"-" db:"-"`
	HasMultipleActions bool      `json:"-" db:"-"`
	FirstSeen          time.Time `json:"-" db:"-"`
}

func CreateTradeTableIfNotExists(db *sqlx.DB) error {
	_, err := db.Exec(TradeDDL)
	if err != nil {
		return err
	}

	_, err = db.Exec(HashTradeAddIndex)

	_, err = db.Exec(AddressTradeAddIndex)

	_, err = db.Exec(PoolAddrTradeIndex)
	return err
}

func (t *Trade) SaveToDB(db *sqlx.DB) error {
	_, err := db.NamedExec(`
INSERT INTO trades (hash, walletType, swapType, tradeType, address, poolAddr, amountIn, boc, amount, tokenAmount, limits, recipient, referrer, fullfillBOC, rejectBOC, lastestReserve0, lastestReserve1, latestPoolLt) values(:hash, :walletType, :swapType, :tradeType, :address, :poolAddr, :amountIn, :boc, :amount, :tokenAmount, :limits, :recipient, :referrer, :fullfillBOC, :rejectBOC, :lastestReserve0, :lastestReserve1, :latestPoolLt)`, t)
	return err
}

func (t *Trade) Dump() string {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(t); err != nil {
		return ""
	}
	return buf.String()
}
