package cli

import (
	"context"
	"fmt"
	"time"

	"github.com/cmingxu/dedust/model"
	"github.com/cmingxu/dedust/utils"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"
	cli2 "github.com/urfave/cli/v2"
	"github.com/xssnick/tonutils-go/address"
)

var (
	checkPoolReserveCmd = &cli2.Command{
		Name: "check-pool-reserve",
		Flags: []cli2.Flag{
			&host,
			&port,
			&user,
			&password,
			&database,
			&tonConfig,
		},
		Action: func(c *cli2.Context) error {
			if err := utils.SetupLogger(c.String("loglevel")); err != nil {
				return err
			}
			return checkPool(c)
		},
	}
)

func checkPool(c *cli2.Context) error {
	fmt.Println("Checking pool...")

	db, err := sqlx.Connect("mysql", utils.ConstructDSN(c))
	if err != nil {
		return err
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	connPool, ctx, err := utils.GetConnectionPool(c.String("ton-config"))
	if err != nil {
		return err
	}
	client := utils.GetAPIClient(connPool)
	masterBlock, err := client.GetMasterchainInfo(ctx)
	if err != nil {
		return err
	}

	pools, err := model.LoadPoolsFromDB(db, true)
	if err != nil {
		return err
	}

	changedCnt := 0
	unchangedCnt := 0
	for i, pool := range pools {
		fmt.Println("==================")
		fmt.Printf("Checking pool %d/%d: %s\n", i+1, len(pools), pool.Address)
		if i%30 == 0 {
			masterBlock, err = client.GetMasterchainInfo(ctx)
			if err != nil {
				log.Error().Err(err).Msg("failed to get masterchain info")
				continue
			}
		}

		addr := address.MustParseAddr(pool.Address)
		stack, err := client.RunGetMethod(ctx, masterBlock, addr, "get_reserves")
		if err != nil {
			log.Error().Err(err).Msgf("failed to get reserves %s", pool.Address)
			continue
		}

		reserve0, err := stack.Int(0)
		if err != nil {
			log.Error().Err(err).Msg("failed to get slice #0")
			continue
		}

		reserve1, err := stack.Int(1)
		if err != nil {
			log.Error().Err(err).Msg("failed to get slice #1")
			continue
		}

		if pool.Asset0Reserve != reserve0.String() || pool.Asset1Reserve != reserve1.String() {
			changedCnt++
			fmt.Printf("Pool %s reserve changed: %s %s -> %s %s\n",
				pool.Address, pool.Asset0Reserve, pool.Asset1Reserve, reserve0.String(), reserve1.String())
		} else {
			unchangedCnt++
			fmt.Printf("Pool %s reserve unchanged: %s %s\n",
				pool.Address, pool.Asset0Reserve, pool.Asset0Reserve)
		}
	}

	fmt.Printf("Total %d pools, %d changed, %d unchanged\n", len(pools), changedCnt, unchangedCnt)

	return nil
}
