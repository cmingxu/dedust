package detector

import (
	"fmt"
	"strings"

	"github.com/cmingxu/dedust/model"
	"github.com/cmingxu/dedust/utils"
	"github.com/gorilla/websocket"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/samber/lo"
	"github.com/tidwall/gjson"
	"github.com/xssnick/tonutils-go/address"
)

func MemPoolCheck(dsn string) error {
	db, err := sqlx.Connect("mysql", dsn)
	if err != nil {
		return err
	}

	poolMap := make(map[string]*model.Pool)
	pools, err := model.LoadPoolsFromDB(db, true)
	if err != nil {
		return err
	}

	for _, p := range pools {
		if _, ok := poolMap[utils.RawAddr(address.MustParseAddr(p.Address))]; !ok {
			poolMap[utils.RawAddr(address.MustParseAddr(p.Address))] = p
		}
	}

	accounts := lo.Keys(poolMap)

	if len(accounts) > 500 {
		accounts = accounts[:500]
	}

	log.Debug().Msgf("subscribe to mempool with %d accounts", len(accounts))
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

	log.Debug().Msgf("subscribing to mempool with %+v", params)
	if err := conn.WriteJSON(params); err != nil {
		return errors.Wrap(err, "failed to write json")
	}

	for {
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
	}
}
