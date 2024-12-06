package detector

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/r3labs/sse"
	"github.com/rs/zerolog/log"
)

// ton API websocket request support up to 1000 accountsId passingin
const AccountsLenOfEachWebSocketRequest = 950
const TOKEN = "AEETAB4AU6BMELIAAAADMMZHBQOIVYFMRL7QZ77HCXATNHS5PF6CIJQJNAQRLC4OG73V2VQ"

const TonAPIMemPoolSSE = "https://tonapi.io/v2/sse/mempool?token=%s"

// const TonAPIMemPoolSSE = "https://116.202.150.118/v2/sse?token=%s"
type MPResponse struct {
	Boc string `json:"boc"`
}

func (r *MPResponse) ShortString() string {
	hash := sha1.New()
	hash.Write([]byte(r.Boc))
	bocHash := hash.Sum(nil)
	return fmt.Sprintf("boc: %s", hex.EncodeToString(bocHash))
}

func (r *MPResponse) String() string {
	hash := sha1.New()
	hash.Write([]byte(r.Boc))
	bocHash := hash.Sum(nil)
	return fmt.Sprintf("boc: %s", hex.EncodeToString(bocHash))
}

func (d *Detector) SubscribeTradeSignalFromTonAPIMemPool(mpResponseCh chan *MPResponse) error {
AGIAN:
	client := sse.NewClient(fmt.Sprintf(TonAPIMemPoolSSE, TOKEN))
	err := client.Subscribe("messages", func(msg *sse.Event) {
		var boc struct {
			Boc string `json:"boc"`
		}

		if err := json.Unmarshal(msg.Data, &boc); err != nil {
			log.Error().Err(err).Msg("failed to unmarshal boc")
			return
		}

		mpResponseCh <- &MPResponse{
			Boc: boc.Boc,
		}
	})

	if err != nil {
		time.Sleep(time.Second * 1)
		goto AGIAN
	}

	return nil
}
