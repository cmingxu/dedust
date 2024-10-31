package model

import (
	"time"

	"github.com/jmoiron/sqlx"
)

const BundleCreationDDL = `
		CREATE TABLE IF NOT EXISTS bundles (
			id int PRIMARY KEY AUTO_INCREMENT,
			address varchar(255),
			privateKey text,
			withdraw bool DEFAULT false,
			createdAt timestamp DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
		);
`

type Bundle struct {
	ID         int        `json:"id" db:"id"`
	Address    string     `json:"address" db:"address"`
	PrivateKey string     `json:"privateKey" db:"privateKey"`
	Withdraw   bool       `json:"withdraw" db:"withdraw"`
	CreatedAt  *time.Time `json:"createdAt" db:"createdAt"`
}

func CreateBundleTableIfNotExists(db *sqlx.DB) error {
	_, err := db.Exec(BundleCreationDDL)
	if err != nil {
		return err
	}

	return err
}

func (b *Bundle) SaveToDB(db *sqlx.DB) error {
	row1, err := db.NamedQuery("INSERT INTO bundles(address, privateKey) VALUES (:address, :privateKey) ", b)
	if err != nil {
		return err
	}

	return row1.Close()
}

func (b *Bundle) UpdateWithdraw(db *sqlx.DB) error {
	row, err := db.NamedQuery("UPDATE bundles SET withdraw = :withdraw  WHERE address = :address", b)
	if err != nil {
		return err
	}

	return row.Close()
}
