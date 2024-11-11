package cli

import (
	"fmt"

	"github.com/cmingxu/dedust/model"
	"github.com/cmingxu/dedust/utils"
	"github.com/jmoiron/sqlx"
	cli2 "github.com/urfave/cli/v2"
)

var (
	infoCmd = &cli2.Command{
		Name: "info",
		Flags: []cli2.Flag{
			&host,
			&port,
			&user,
			&password,
			&database,
		},
		Action: func(c *cli2.Context) error {
			if err := utils.SetupLogger(c.String("loglevel")); err != nil {
				return err
			}

			return info(c)
		},
	}
)

func info(c *cli2.Context) error {
	db, err := sqlx.Connect("mysql", utils.ConstructDSN(c))
	if err != nil {
		return err
	}
	defer db.Close()

	pools, err := model.LoadPoolsFromDB(db, false)
	if err != nil {
		return err
	}

	outstandingPools, err := model.LoadPoolsFromDB(db, true)
	if err != nil {
		return err
	}

	fmt.Println("Pools:", len(pools), "Outstanding Pools:", len(outstandingPools))

	return nil
}
