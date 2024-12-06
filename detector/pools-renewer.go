package detector

import (
	"context"
	"time"

	"github.com/cmingxu/dedust/model"
	"github.com/cmingxu/dedust/utils"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"
	"github.com/xssnick/tonutils-go/address"
)

// pool renew interval 1 hour for now
const PoolRenewInterval = 60 * 60 * time.Second

// filter pool
const OutstandingPoolOnly = true

func (d *Detector) PerodicallyRenewPoolsFromDB(ctx context.Context) {
	for {
		select {
		case <-d.poolRenewTimer.C:
			log.Debug().Msg("renew pool from db")
			if err := d.renewPoolsFromDB(d.db); err != nil {
				log.Error().Err(err).Msg("renew pool from db failed")
			}

			log.Debug().Msgf("total pool count is %d now", len(d.poolMap))
		case <-ctx.Done():
			return
		}
	}
}

func (d *Detector) renewPoolsFromDB(db *sqlx.DB) (err error) {
	d.poolLock.Lock()
	defer d.poolLock.Unlock()
	defer d.poolRenewTimer.Reset(PoolRenewInterval)

	pools, err := model.LoadPoolsFromDB(d.db, OutstandingPoolOnly)
	if err != nil {
		return err
	}

	for _, p := range pools {
		if _, ok := d.poolMap[utils.RawAddr(address.MustParseAddr(p.Address))]; !ok {
			d.poolMap[utils.RawAddr(address.MustParseAddr(p.Address))] = p
		}
	}

	return nil
}
