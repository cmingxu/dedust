package cli

import (
	"fmt"
	"os"

	"github.com/cmingxu/dedust/model"
	"github.com/cmingxu/dedust/utils"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	cli2 "github.com/urfave/cli/v2"
)

func bootstrap(c *cli2.Context) error {
	db, err := sqlx.Connect("mysql", utils.ConstructDSN(c))
	if err != nil {
		return err
	}
	defer db.Close()

	if err := model.CreateTableIfNotExists(db); err != nil {
		fmt.Fprintf(os.Stderr, "%+v", err)
	}

	if err := model.CreateTradeTableIfNotExists(db); err != nil {
		fmt.Fprintf(os.Stderr, "%+v", err)
	}

	if err := model.CreateBundleTableIfNotExists(db); err != nil {
		fmt.Fprintf(os.Stderr, "%+v", err)
	}

	return nil
}
