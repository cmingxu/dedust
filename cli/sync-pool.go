package cli

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/cmingxu/dedust/model"
	"github.com/cmingxu/dedust/utils"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"
	cli2 "github.com/urfave/cli/v2"
)

const DedustPoolAPI = "https://api.dedust.io/v2/pools"

func syncPool(c *cli2.Context) error {
	fmt.Println("Syncing pool...")

	db, err := sqlx.Connect("mysql", utils.ConstructDSN(c))
	if err != nil {
		return err
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	body, err := utils.Request(ctx, http.MethodGet, DedustPoolAPI, nil)
	if err != nil {
		return err
	}

	pools, err := model.LoadPoolsFromJSON(body)
	if err != nil {
		return err
	}

	for _, pool := range pools {
		err := pool.SaveToDB(db)
		if err != nil {
			log.Err(err).Msg("Failed to save pool to db")
			return err
		}

		log.Debug().Str("pool", pool.Address).Msg("Pool saved")
	}

	return nil
}
