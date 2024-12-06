package detector

import (
	"crypto/tls"
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

const TonAPIMemPoolEndpoint = "wss://tonapi.io/v2/websocket?token=%s"

type MPRequest struct {
	Id      int      `json:"id"`
	Method  string   `json:"method"`
	JsonRPC string   `json:"jsonrpc"`
	Params  []string `json:"params"`
}

type WebsocketMPResponse struct {
	Boc              string   `json:"boc"`
	InvolvedAccounts []string `json:"involved_accounts"`
}

func (w WebsocketMPResponse) String() string {
	return fmt.Sprintf("boc: %s, involved_accounts: %s", w.Boc, w.InvolvedAccounts)
}

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

	dialer := websocket.DefaultDialer
	dialer.TLSClientConfig = &tls.Config{
		InsecureSkipVerify: true,
	}
	conn, _, err := dialer.Dial(fmt.Sprintf(TonAPIMemPoolEndpoint, TOKEN), nil)

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
		mpResponse := WebsocketMPResponse{
			Boc: result.Get("params.boc").String(),
		}
		result.Get("params.involved_accounts").ForEach(func(_, value gjson.Result) bool {
			mpResponse.InvolvedAccounts = append(mpResponse.InvolvedAccounts, value.String())
			return true
		})

		log.Debug().Msgf("received mempool response %+s", mpResponse.String())
	}
}
