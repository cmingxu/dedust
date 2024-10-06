package detector

import (
	"context"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/xssnick/tonutils-go/address"
)

// update pool reserves from get method
func (d *Detector) PoolReserveUpdater(ctx context.Context) error {
	poolIds := make([]string, 0)
	d.poolLock.RLock()
	for k := range d.poolMap {
		poolIds = append(poolIds, k)
	}
	d.poolLock.RUnlock()

	updaterLogger := log.With().Str("module", "pool-reserve-updater").Logger()

	for index, poolId := range poolIds {
		log.Debug().Msgf("updating pool %s (%d/%d)", poolId, index+1, len(poolIds))

		addr, err := address.ParseRawAddr(poolId)
		if err != nil {
			updaterLogger.Error().Err(err).Msgf("failed to parse pool id %s", poolId)
			continue
		}

		block, err := d.apiClient.GetMasterchainInfo(d.apiCtx)
		if err != nil {
			updaterLogger.Error().Err(err).Msg("failed to get masterchain info")
			continue
		}
		stack, err := d.apiClient.RunGetMethod(d.apiCtx, block, addr, "get_reserves")
		if err != nil {
			updaterLogger.Error().Err(err).Msgf("failed to get reserves for pool %s", poolId)
			continue
		}

		d.poolLock.RLock()
		pool, ok := d.poolMap[poolId]
		if !ok {
			d.poolLock.RUnlock()
			updaterLogger.Warn().Msgf("pool %s not found", poolId)
			continue
		}
		d.poolLock.RUnlock()

		reserve0, err := stack.Int(0)
		if err != nil {
			updaterLogger.Error().Err(err).Msg("failed to get slice")
			continue
		}
		pool.Asset0Reserve = reserve0.String()

		reserve1, err := stack.Int(1)
		if err != nil {
			updaterLogger.Error().Err(err).Msg("failed to get slice")
			continue
		}
		pool.Asset1Reserve = reserve1.String()

		pool.UpdatedAt = time.Now()

		if err := pool.UpdateReserves(d.db); err != nil {
			updaterLogger.Error().Err(err).Msg("failed to update pool")
			continue
		}

		updaterLogger.Debug().Msgf("pool %s updated new reserves %s <=> %s",
			poolId, pool.Asset0Reserve, pool.Asset1Reserve)
	}

	return nil
}
