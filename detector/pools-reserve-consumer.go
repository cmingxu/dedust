package detector

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/cmingxu/dedust/model"
	"github.com/cmingxu/dedust/utils"
	"github.com/r3labs/sse"
	"github.com/rs/zerolog/log"
	"github.com/tidwall/gjson"
)

const DedustPayoutFromPoolURL = "https://tonapi.io/v2/sse/accounts/transactions?accounts=ALL&operations=0x61ee542d&token=%s"

const AccountInfoURL = "http://49.12.81.26:8080/api/v0/accounts?address=%s&latest=true"

// const AccountInfoURL = "https://anton.tools/api/v0/accounts?address=%s&latest=true"

type DedustPayOutMessage struct {
	AccountId string `json:"account_id"`
	Lt        uint64 `json:"lt"`
	TxHash    string `json:"tx_hash"`
}

func (d *Detector) PoolReserveConsumer(ctx context.Context) error {
	client := sse.NewClient(fmt.Sprintf(DedustPayoutFromPoolURL, TOKEN))

	return client.Subscribe("messages", func(msg *sse.Event) {
		if len(msg.Data) == 0 {
			return
		}

		var payout DedustPayOutMessage
		err := json.Unmarshal(msg.Data, &payout)
		if err != nil {
			log.Debug().Err(err).Msg("failed to unmarshal payout message")
			return
		}

		//log.Debug().Msgf("tx-hash: %s", payout.TxHash)
		//log.Debug().Msgf("accountId: %s (%s)", payout.AccountId, address.MustParseRawAddr(payout.AccountId).String())
		//log.Debug().Msgf("lt: %d", payout.Lt)

		d.poolLock.RLock()
		defer d.poolLock.RUnlock()

		pool, ok := d.poolMap[payout.AccountId]
		if ok {
			go d.updatePoolReserve(ctx, pool, payout.AccountId, payout.Lt)
		} else {
			log.Warn().Msgf("pool %s not found", payout.AccountId)
		}
	})
}

func (d *Detector) updatePoolReserve(ctx context.Context, pool *model.Pool,
	accountId string, lt uint64) {
	i := 0
	for i < 40 {
		resp, err := utils.Request(ctx, http.MethodGet, fmt.Sprintf(AccountInfoURL, accountId), nil)
		if err != nil {
			log.Error().Err(err).Msg("failed to request transaction detail")
			return
		}

		if len(resp) < 100 {
			log.Error().Msgf("transaction detail too short: %s", string(resp))
			return
		}

		result := gjson.Parse(string(resp))
		anton_last_lt := uint64(result.Get("results.0.last_tx_lt").Int())

		if anton_last_lt >= lt {
			log.Debug().Msgf("lastest anton account info found at LT %d after %d tries", lt, i)
			pool.Asset0Reserve = result.Get("results.0.executed_get_methods.dedust_v2_pool.2.returns.0").String()
			pool.Asset1Reserve = result.Get("results.0.executed_get_methods.dedust_v2_pool.2.returns.1").String()
			pool.Lt = lt

			pool.UpdatedAt = time.Now()

			if err := pool.UpdateReserves(d.db); err != nil {
				log.Error().Err(err).Msg("failed to update pool")
				return
			}
			log.Debug().Msgf("pool %s updated new reserves %s <=> %s",
				pool.Address, pool.Asset0Reserve, pool.Asset1Reserve)

			break
		}

		time.Sleep(time.Second * 1)
		i += 1
	}
}
