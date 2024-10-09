package printer

import (
	"context"
	"crypto/ed25519"
	"encoding/json"
	"math/big"
	"os"
	"sync"
	"time"

	"github.com/cmingxu/dedust/bot"
	"github.com/cmingxu/dedust/model"
	"github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/liteclient"
	"github.com/xssnick/tonutils-go/tlb"
	"github.com/xssnick/tonutils-go/ton"
)

type Printer struct {
	working    bool
	conn       *websocket.Conn
	wsEndpoint string

	seqno   uint64
	balance tlb.Coins

	botPrivateKey ed25519.PrivateKey
	addr          *address.Address
	client        ton.APIClientWrapped
	ctx           context.Context

	pool *liteclient.ConnectionPool
	out  *os.File
}

func NewPrinter(
	ctx context.Context,
	pool *liteclient.ConnectionPool,
	client ton.APIClientWrapped,
	addr *address.Address,
	botprivateKey ed25519.PrivateKey,
	wsEndpoint string,
	outPath string,
) (*Printer, error) {
	p := &Printer{
		wsEndpoint:    wsEndpoint,
		botPrivateKey: botprivateKey,
		addr:          addr,
		pool:          pool,
		client:        client,
		ctx:           ctx,
	}

	var err error
	p.out, err = os.OpenFile(outPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}

	p.working = false
	return p, nil
}

func (p *Printer) Run() error {
	var err error

	chanceCh := make(chan model.BundleChance, 10)
	go func() {
	RETRY:
		p.conn, _, err = websocket.DefaultDialer.Dial(p.wsEndpoint, nil)
		if err != nil {
			log.Error().Err(err).Msg("failed to connect to ws")
			time.Sleep(3 * time.Second)
			goto RETRY
		}

		defer p.conn.Close()
		for {
			messageType, message, err := p.conn.ReadMessage()
			if err != nil {
				goto RETRY
			}

			if messageType == websocket.PingMessage {
				err = p.conn.WriteMessage(websocket.PongMessage, nil)
				if err != nil {
					goto RETRY
				}
			}

			var chance model.BundleChance
			err = json.Unmarshal(message, &chance)
			if err != nil {
				continue
			}

			chanceCh <- chance
		}
	}()

	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	timer := time.NewTimer(6000000 * time.Second)
	for {
		select {
		case chance := <-chanceCh:
			log.Debug().Msgf("got chance info %+s", chance.Dump())
			if p.MeetRequirement(&chance) {
				log.Debug().Msg("meet requirement")

				poolAddr := address.MustParseAddr(chance.PoolAddress)
				botInAmount := tlb.MustFromNano(stringToBigInt(chance.BotIn), 9)
				nextLimit := botInAmount

				limitBN := stringToBigInt(chance.BotJettonOut)
				// set limit as 99.95 of expected
				limitBN9995 := new(big.Int).Div(new(big.Int).Mul(limitBN, big.NewInt(9995)), big.NewInt(10000))
				limit := tlb.MustFromNano(limitBN9995, 9)

				log.Debug().Msgf("pool %s, botIn %s, limit %s, nextLimit %s",
					poolAddr.String(), botInAmount.String(), limit.String(), nextLimit.String())

				nbot := bot.NewBotWallet(p.ctx, p.client, p.addr, p.botPrivateKey, p.seqno)
				msg := nbot.BuildBundle(
					poolAddr,
					botInAmount.Nano(),
					limit.Nano(),
					nextLimit.Nano(),
				)

				log.Debug().Msgf("built bundle %+v", msg)

				ctx := context.Background()
				err = nbot.Send(ctx, msg, false)
				if err != nil {
					log.Error().Err(err).Msg("failed to send bundle")
				}

				ctxSlices := make([]context.Context, 0)

				i := 0
				nodeCtx := context.WithValue(context.Background(), "foo", struct{}{})
				for i < 5 {
					nodeCtx, err = p.pool.StickyContextNextNodeBalanced(nodeCtx)
					if err != nil {
						log.Error().Err(err).Msg("failed to get next node")
						break
					}
					nodeId, _ := nodeCtx.Value("_ton_node_sticky").(uint32)
					ctxSlices = append(ctxSlices,
						context.WithValue(context.Background(), "_ton_node_sticky", nodeId))
					i++
				}

				var wg sync.WaitGroup
				wg.Add(len(ctxSlices))
				for i, c := range ctxSlices {
					go func(i int, c context.Context) {
						defer wg.Done()

						nid, _ := c.Value("_ton_node_sticky").(uint32)
						err = nbot.Send(c, msg, false)
						log.Debug().Msgf("sent bundle to node %d[%d]", nid, i)
						if err != nil {
							log.Error().Err(err).Msg("failed to send bundle")
						}
					}(i, c)
				}

				wg.Wait()

				p.working = true

				if err := chance.CSV(p.out); err != nil {
					log.Error().Err(err).Msg("failed to write to file")
				}

				timer.Reset(60 * time.Second)
			}

		case <-timer.C:
			log.Debug().Msg("timer due")
			timer.Reset(6000000 * time.Second)
			if p.working {
				p.working = false
			}
		case <-ticker.C:
			log.Debug().Msgf("dida dida, seqno: %d, working: %t", p.seqno, p.working)
			go func() {
				no, balance, err := p.getInfo()
				if err != nil {
					log.Error().Err(err).Msg("failed to get seqno")
				} else {
					p.seqno = no
					p.balance = balance
				}
			}()
		}
	}
}
func (p *Printer) getInfo() (uint64, tlb.Coins, error) {
	masterBlock, err := p.client.GetMasterchainInfo(p.ctx)

	if err != nil {
		return 0, tlb.Coins{}, err
	}
	stack, err := p.client.RunGetMethod(p.ctx, masterBlock, p.addr, "seqno")

	if err != nil {
		return 0, tlb.Coins{}, err
	}

	seqBI, err := stack.Int(0)
	if err != nil {
		return 0, tlb.Coins{}, err
	}

	acc, err := p.client.GetAccount(p.ctx, masterBlock, p.addr)
	if err != nil {
		return 0, tlb.Coins{}, err
	}

	return seqBI.Uint64(), acc.State.Balance, nil
}

func (p *Printer) MeetRequirement(chance *model.BundleChance) bool {
	in := stringToBigInt(chance.BotIn)
	if p.balance.Nano().Cmp(in) < 0 {
		return false
	}

	// ROI bigger then 1%
	roi := stringToBigInt(chance.Roi)
	roiBiggerThen1Percent := roi.Cmp(big.NewInt(100)) > 0

	// profit more then 0.5 TON
	profitWaterMark := tlb.MustFromTON("0.3")
	profit := stringToBigInt(chance.Profit)
	profitMoreThenWaterMark := profit.Cmp(profitWaterMark.Nano()) > 0

	return roiBiggerThen1Percent || profitMoreThenWaterMark
}

func stringToBigInt(s string) *big.Int {
	i := new(big.Int)
	i.SetString(s, 10)
	return i
}
