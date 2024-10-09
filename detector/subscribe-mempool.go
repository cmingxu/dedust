package detector

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/samber/lo"
	"github.com/tidwall/gjson"
)

// ton API websocket request support up to 1000 accountsId passingin
const AccountsLenOfEachWebSocketRequest = 950
const TOKEN = "AEETAB4AU6BMELIAAAADMMZHBQOIVYFMRL7QZ77HCXATNHS5PF6CIJQJNAQRLC4OG73V2VQ"
const TonAPIMemPoolEndpoint = "wss://tonapi.io/v2/websocket?token=%s"

type MPRequest struct {
	Id      int      `json:"id"`
	Method  string   `json:"method"`
	JsonRPC string   `json:"jsonrpc"`
	Params  []string `json:"params"`
}

type MPResponse struct {
	Boc              string   `json:"boc"`
	InvolvedAccounts []string `json:"involved_accounts"`
}

func (r *MPResponse) String() string {
	hash := sha1.New()
	hash.Write([]byte(r.Boc))
	bocHash := hash.Sum(nil)
	return fmt.Sprintf("boc: %s, involved accounts len: %d", hex.EncodeToString(bocHash), len(r.InvolvedAccounts))
}

func (d *Detector) SubscribeTradeSignalFromTonAPIMemPool(ctx context.Context,
	poolsRenewedCh chan struct{},
	mpResponseCh chan *MPResponse) error {
	subCtx, subCancel := context.WithCancel(context.Background())
	for {
		select {
		case <-ctx.Done():
			subCancel()
			return nil

		case <-poolsRenewedCh:
			log.Debug().Msg("renew pool from db")
			subCancel()
			subCtx, subCancel = context.WithCancel(context.Background())

			for i, accounts := range lo.Chunk(lo.Keys(d.poolMap), AccountsLenOfEachWebSocketRequest) {
				go func() {
				AGAIN:
					log.Debug().Msgf("subscribe to mempool with %d accounts", len(accounts))
					if err := d.subscribe(subCtx, i, accounts, mpResponseCh); err != nil {
						log.Error().Err(err).Msg("failed to subscribe")
						time.Sleep(1 * time.Second)
						goto AGAIN
					}
				}()
			}
		}
	}
}

func (d *Detector) subscribe(ctx context.Context,
	serialNo int,
	accounts []string,
	mpResponseCh chan *MPResponse) error {
	log.Debug().Msgf("#%d goroutine subscribe %d accounts", serialNo, len(accounts))
	defer func() {
		log.Debug().Msgf("#%d goroutine subscribe %d accounts done", serialNo, len(accounts))
	}()

	params := MPRequest{
		Id:      1,
		Method:  "subscribe_mempool",
		JsonRPC: "2.0",
		Params:  []string{},
	}

	params.Params = append(params.Params, fmt.Sprintf(
		"accounts=%s", strings.Join(accounts, ","),
	))

	conn, _, err := websocket.DefaultDialer.Dial(fmt.Sprintf(TonAPIMemPoolEndpoint, TOKEN), nil)
	if err != nil {
		return errors.Wrap(err, "failed to dial")
	}
	defer conn.Close()

	if err := conn.WriteJSON(params); err != nil {
		return errors.Wrap(err, "failed to write json")
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
		}
		_, message, err := conn.ReadMessage()
		if err != nil {
			return err
		}

		result := gjson.Parse(string(message))
		mpResponse := MPResponse{
			Boc: result.Get("params.boc").String(),
		}
		result.Get("params.involved_accounts").ForEach(func(_, value gjson.Result) bool {
			mpResponse.InvolvedAccounts = append(mpResponse.InvolvedAccounts, value.String())
			return true
		})

		log.Debug().Msgf("received mempool response %+s", mpResponse.String())
		mpResponseCh <- &mpResponse
	}
}
