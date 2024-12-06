package cli

import (
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/cmingxu/dedust/utils"
	"github.com/gorilla/websocket"
	"github.com/r3labs/sse"
	"github.com/rs/zerolog/log"
	"github.com/xssnick/tonutils-go/tlb"
	"github.com/xssnick/tonutils-go/tvm/cell"

	cli2 "github.com/urfave/cli/v2"
)

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

const TOKEN = "AEETAB4AU6BMELIAAAADMMZHBQOIVYFMRL7QZ77HCXATNHS5PF6CIJQJNAQRLC4OG73V2VQ"

const TonAPIMemPoolSSE = "https://tonapi.io/v2/sse/mempool?token=%s"

// const TonAPIMemPoolSSE = "https://116.202.150.118/v2/sse/mempool?token=%s"

var memPoolAllCmd = &cli2.Command{
	Name:        "mem-pool-all",
	Description: "to subscribe mem pool all accounts",
	Flags:       []cli2.Flag{},
	Action: func(c *cli2.Context) error {
		if err := utils.SetupLogger(c.String("loglevel")); err != nil {
			return err
		}
		return memPoolAllAccount(c)
	},
}

func memPoolAllAccount(c *cli2.Context) error {
	dialer := websocket.DefaultDialer
	dialer.TLSClientConfig = &tls.Config{
		InsecureSkipVerify: true,
	}
	client := sse.NewClient(fmt.Sprintf(TonAPIMemPoolSSE, TOKEN))
	client.Subscribe("messages", func(msg *sse.Event) {
		log.Debug().Msgf("====================")
		log.Debug().Msgf("received mempool response %+s", string(msg.Data))

		var boc struct {
			Boc string `json:"boc"`
		}

		if err := json.Unmarshal(msg.Data, &boc); err != nil {
			log.Error().Err(err).Msg("failed to unmarshal boc")
			return
		}

		anyMessage, err := externalMessageFromBOC(boc.Boc)
		if err != nil {
			log.Error().Err(err).Msg("failed to parse message from boc")
			return
		}

		log.Debug().Msgf("message type(%s)", anyMessage.MsgType)
		if anyMessage.MsgType == tlb.MsgTypeExternalIn {
			external, ok := anyMessage.Msg.(*tlb.ExternalMessage)
			if !ok {
				log.Error().Msg("failed to convert to external message")
				return
			}

			log.Debug().Msgf("external dst: %s", external.DstAddr)
			log.Debug().Msgf("external src: %s", external.SrcAddr)
			log.Debug().Msgf("external body: %s", external.Body.Dump())
		}
	})

	return nil
}

func externalMessageFromBOC(boc string) (*tlb.Message, error) {
	var msg tlb.Message
	rawBoc, err := base64.StdEncoding.DecodeString(boc)
	if err != nil {
		return nil, err
	}

	c, err := cell.FromBOC(rawBoc)
	if err != nil {
		return nil, err
	}

	if err := tlb.LoadFromCell(&msg, c.BeginParse()); err != nil {
		return nil, err
	}

	return &msg, nil
}
