package detector

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/cmingxu/dedust/model"
	"github.com/cmingxu/dedust/utils"
	"github.com/r3labs/sse"
	"github.com/rs/zerolog/log"
	"github.com/xssnick/tonutils-go/address"
)

const DedustPayoutFromPoolURL = "https://tonapi.io/v2/sse/accounts/transactions?accounts=ALL&operations=0x61ee542d&token=%s"

type DedustPayOutMessage struct {
	AccountId string `json:"account_id"`
	Lt        uint64 `json:"lt"`
	TxHash    string `json:"tx_hash"`
}

func (d *Detector) PoolReserveConsumer(ctx context.Context) error {
	client := sse.NewClient(fmt.Sprintf(DedustPayoutFromPoolURL, TOKEN))

	for {
		err := client.Subscribe("messages", func(msg *sse.Event) {
			if len(msg.Data) == 0 {
				return
			}

			var payout DedustPayOutMessage
			err := json.Unmarshal(msg.Data, &payout)
			if err != nil {
				log.Debug().Err(err).Msg("failed to unmarshal payout message")
				return
			}
			log.Debug().Msgf("SSE pool reserve accountId: %s LT: %d", payout.AccountId, payout.Lt)

			d.poolLock.RLock()
			pool, ok := d.poolMap[payout.AccountId]
			d.poolLock.RUnlock()

			if ok {
				go func() {
					if err := utils.TimeitReturnError(fmt.Sprintf("update pool reserve: %s %d",
						pool.Address, payout.Lt), func() error {
						return d.updatePoolReserve(ctx, pool, payout.AccountId, payout.Lt)
					}); err != nil {
						log.Error().Err(err).Msg("failed to update pool reserve")
					}
				}()
			} else {
				log.Warn().Msgf("ReserveConsumer - pool %s not found", payout.AccountId)
			}
		})

		if err != nil {
			log.Error().Err(err).Msg("failed to subscribe to dedust payout from pool")
		}
	}
}

func (d *Detector) updatePoolReserve(ctx context.Context, pool *model.Pool,
	accountId string, lt uint64) error {
	client := utils.GetAPIClientWithTimeout(d.connPool, time.Second*10)
	block, err := client.GetMasterchainInfo(ctx)
	if err != nil {
		return err
	}

	poolAddr := address.MustParseAddr(pool.Address)
	stack, err := client.RunGetMethod(ctx, block,
		poolAddr,
		"get_reserves")
	if err != nil {
		return err
	}

	r0, err := stack.Int(0)
	if err != nil {
		return err
	}
	pool.Asset0Reserve = r0.String()

	r1, err := stack.Int(1)
	if err != nil {
		return err
	}

	pool.Asset1Reserve = r1.String()

	return pool.SaveToDB(d.db)
}
