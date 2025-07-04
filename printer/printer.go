package printer

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"crypto/tls"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"math/rand"
	"net"
	"net/http"
	"net/http/httptrace"
	"os"
	"sync"
	"time"

	"github.com/cmingxu/dedust/bot"
	"github.com/cmingxu/dedust/model"
	"github.com/cmingxu/dedust/utils"
	"github.com/gorilla/websocket"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"
	"github.com/samber/lo"
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/liteclient"
	"github.com/xssnick/tonutils-go/tlb"
	"github.com/xssnick/tonutils-go/ton"
	"github.com/xssnick/tonutils-go/ton/wallet"
)

const TONCENTER_API_KEY = "00a6aa0b2060bfb6ad22b8fa3e34e856f132f995b0074310935d2da67781e9b2"
const TOKEN = "AEETAB4AU6BMELIAAAADMMZHBQOIVYFMRL7QZ77HCXATNHS5PF6CIJQJNAQRLC4OG73V2VQ"

var (
	ShardId = tlb.ShardID(0xe000000000000000)
)

var (
	DCM_ADDR = address.MustParseAddr("UQCwSxqefElovEPlpZ8bIEL_KXqWuqoOhwb65uYjos9bCDcM")
)

const (
	SendModeNormal   = int64(0)
	SendModeObsolate = int64(5)
)

type Printer struct {
	working    bool
	conn       *websocket.Conn
	wsEndpoint string
	tonConfig  string

	seqno   uint64
	balance tlb.Coins

	botPrivateKey ed25519.PrivateKey
	addr          *address.Address
	client        ton.APIClientWrapped
	ctx           context.Context

	pool *liteclient.ConnectionPool
	out  *os.File

	sendCnt uint32

	useTonAPI           bool
	useTonAPIBlockchain bool
	useTonCenter        bool
	useTonCenterV3      bool
	useANDL             bool
	enableTracing       bool

	upperlimit *big.Int
	lowerlimit *big.Int

	httpClt   *http.Client
	httpTrace *httptrace.ClientTrace

	db *sqlx.DB

	// v4 wallet mode 0 or 5
	v4walletMode int64

	skippedPoolAddress     []string
	skippedVictimAccountId []string
}

func NewPrinter(
	tonConfig string,
	addr *address.Address,
	botprivateKey ed25519.PrivateKey,
	wsEndpoint string,
	outPath string,
	sendCnt uint32,
	useTonAPI bool,
	useTonAPIBlockchain bool,
	useTonCenter bool,
	useTonCenterV3 bool,
	useANDL bool,
	enableTracing bool,
	limit string,
	lowerlimit string,
	walletMode int64,
	mysql string,
	skippedPoolAddress []string,
	skippedVictimAccountId []string,
) (*Printer, error) {
	p := &Printer{
		tonConfig:              tonConfig,
		wsEndpoint:             wsEndpoint,
		botPrivateKey:          botprivateKey,
		addr:                   addr,
		sendCnt:                sendCnt,
		useTonAPI:              useTonAPI,
		useTonAPIBlockchain:    useTonAPIBlockchain,
		useTonCenter:           useTonCenter,
		useTonCenterV3:         useTonCenterV3,
		useANDL:                useANDL,
		enableTracing:          enableTracing,
		v4walletMode:           walletMode,
		skippedPoolAddress:     skippedPoolAddress,
		skippedVictimAccountId: skippedVictimAccountId,
	}

	var err error
	p.pool, p.ctx, err = utils.GetConnectionPool(tonConfig)
	if err != nil {
		return nil, err
	}

	p.client = utils.GetAPIClient(p.pool)

	l, err := tlb.FromTON(limit)
	if err != nil {
		return nil, err
	}

	p.upperlimit = l.Nano()

	l, err = tlb.FromTON(lowerlimit)
	if err != nil {
		return nil, err
	}

	p.lowerlimit = l.Nano()

	p.httpClt = &http.Client{
		Timeout: 5 * time.Second,
	}

	p.out, err = os.OpenFile(outPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}

	p.working = false

	p.db, err = sqlx.Connect("mysql", mysql)
	if err != nil {
		return nil, err
	}
	dialer := &net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 2 * time.Second,
		DualStack: true,
		KeepAliveConfig: net.KeepAliveConfig{
			Enable: true,
		},
	}

	t := http.DefaultTransport.(*http.Transport).Clone()

	t.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
		log.Debug().Msgf("address original = %s", addr)
		if addr == "tonapi.io:443" {
			addr = "116.202.150.118:443"
		}

		if addr == "toncenter.com:443" {
			addr = "104.26.0.179:443"
		}
		return dialer.DialContext(ctx, network, addr)
	}

	t.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	t.MaxConnsPerHost = 100
	t.MaxIdleConns = 100
	t.MaxIdleConnsPerHost = 100
	t.IdleConnTimeout = 0

	p.httpClt.Transport = t

	p.httpTrace = &httptrace.ClientTrace{
		GetConn:      func(hostPort string) { log.Debug().Msgf("starting to create conn ", hostPort) },
		GotConn:      func(info httptrace.GotConnInfo) { log.Debug().Msgf("connection established %+v", info) },
		DNSStart:     func(info httptrace.DNSStartInfo) { log.Debug().Msgf("starting to look up dns %+v", info) },
		DNSDone:      func(info httptrace.DNSDoneInfo) { log.Debug().Msgf("done looking up dns", info) },
		ConnectStart: func(network, addr string) { log.Debug().Msgf("starting tcp connection %s, %s", network, addr) },
		ConnectDone: func(network, addr string, err error) {
			log.Debug().Msgf("tcp connection created %s %s %+v", network, addr, err)
		},
		WroteHeaders:         func() { log.Debug().Msgf("wrote headers") },
		WroteRequest:         func(info httptrace.WroteRequestInfo) { log.Debug().Msgf("wrote request %+v", info) },
		GotFirstResponseByte: func() { log.Debug().Msgf("got first response byte") },
	}

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

	reconnectPoolTicker := time.NewTicker(1000000000 * time.Second)
	defer reconnectPoolTicker.Stop()

	timer := time.NewTimer(6000000 * time.Second)
	for {
		select {
		case chance := <-chanceCh:
			log.Debug().Msgf("chance:  %+s", chance.ShortDump())
			if p.MeetRequirement(&chance) {
				poolAddr := address.MustParseAddr(chance.PoolAddress)
				botInAmount := stringToBigInt(chance.BotIn)
				nextLimit := botInAmount

				limitBN := stringToBigInt(chance.BotJettonOut)
				// set limit as 99.99 of expected
				limitBN9999 := new(big.Int).Div(new(big.Int).Mul(limitBN, big.NewInt(9999)), big.NewInt(10000))
				limit := limitBN9999

				// 如果 25s 内没有购买成功，放弃， 当前网络比较慢
				deadline := time.Now().Add(25 * time.Second)

				log.Debug().Msgf("Let's BUN it [pool %s, botIn %s, limit %s, nextLimit %s, deadline: %s]",
					poolAddr.String(), botInAmount.String(), limit.String(), nextLimit.String(),
					deadline.Format(time.RFC3339Nano))

				nbot := bot.NewWallet(p.ctx, p.client, bot.Bot, p.botPrivateKey, nil, p.seqno)

				pk, err := hex.DecodeString(chance.PrivateKeyOfG)
				if err != nil {
					log.Error().Err(err).Msg("failed to decode private key")
					continue
				}

				g, gAddr, _ := bot.BuildG(p.addr,
					pk,
					tlb.MustFromTON("0.3"))

				// msg := nbot.BuildBundle(
				// 	poolAddr,
				// 	botInAmount,
				// 	limit,
				// 	nextLimit,
				// 	uint64(deadline.Unix()),
				// 	gAddr,
				// 	p.v4walletMode,
				// )

				msgAdnl := []*wallet.Message{g}
				msgToncenter := []*wallet.Message{g}
				msgToncenterV3 := []*wallet.Message{g}
				msgTonApi := []*wallet.Message{g}
				msgTonApiBlockchain := []*wallet.Message{g}

				if p.enableTracing {
					c1, _ := p.BuildAuxTransfer(nbot, tlb.MustFromTON("0.000001"), "adnl")
					msgAdnl = append(msgAdnl, c1)
					c2, _ := p.BuildAuxTransfer(nbot, tlb.MustFromTON("0.000001"), "tc")
					msgToncenter = append(msgToncenter, c2)
					c3, _ := p.BuildAuxTransfer(nbot, tlb.MustFromTON("0.000001"), "tcv3")
					msgToncenterV3 = append(msgToncenterV3, c3)
					c4, _ := p.BuildAuxTransfer(nbot, tlb.MustFromTON("0.000001"), "ta")
					msgTonApi = append(msgTonApi, c4)
					c5, _ := p.BuildAuxTransfer(nbot, tlb.MustFromTON("0.000001"), "tab")
					msgTonApiBlockchain = append(msgTonApiBlockchain, c5)
				}

				httpSendCnt := 20
				if p.useTonCenterV3 {
					i := 0
					for i < httpSendCnt {
						go func() {
							msg := nbot.BuildBundle(
								poolAddr,
								botInAmount,
								limit,
								nextLimit,
								uint64(deadline.Unix())+uint64(rand.Int63n(1000)),
								gAddr,
								p.v4walletMode,
							)

							msgList := []*wallet.Message{msg}
							msgList = append(msgList, msgToncenterV3...)

							err := utils.TimeitReturnError("sending with TONCENTER v3", func() error {
								return p.SendWithTONCenterV3(&chance, nbot, msgList)
							})

							if err != nil {
								log.Error().Err(err).Msg("failed to send")
							}
						}()
						i++
					}
				}

				if p.useTonCenter {
					i := 0
					for i < httpSendCnt {
						go func() {

							msg := nbot.BuildBundle(
								poolAddr,
								botInAmount,
								limit,
								nextLimit,
								uint64(deadline.Unix())+uint64(rand.Int63n(1000)),
								gAddr,
								p.v4walletMode,
							)

							msgList := []*wallet.Message{msg}
							msgList = append(msgList, msgToncenter...)

							err := utils.TimeitReturnError("sending with TONCENTER", func() error {
								return p.SendWithTONCenter(&chance, nbot, msgList)
							})

							if err != nil {
								log.Error().Err(err).Msg("failed to send")
							}
						}()

						i++
					}
				}

				if p.useANDL {
					go func() {
						log.Debug().Msgf("sending with ANDL %d", p.sendCnt)
						ctxSlices := make([]context.Context, 0)

						i := 0
						nodeCtx := context.WithValue(context.Background(), "foo", struct{}{})
						for i < int(p.sendCnt) {
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
								err := utils.TimeitReturnError(fmt.Sprintf("send with andl %d [%d]", nid, i), func() error {
									count := 0
									for count < 20 {
										msg := nbot.BuildBundle(
											poolAddr,
											botInAmount,
											limit,
											nextLimit,
											uint64(deadline.Unix())+uint64(rand.Int63n(5000)),
											gAddr,
											p.v4walletMode,
										)
										msgList := []*wallet.Message{msg}
										msgList = append(msgList, msgAdnl...)
										err = nbot.SendMany(c, p.v4walletMode, msgList, false)
										if err != nil {
											log.Error().Err(err).Msgf("failed to send bundle andl")
											break
										}
										count++
									}
									return nil
								})

								if err != nil {
									log.Error().Err(err).Msgf("failed to send bundle %d", i)
								}
							}(i, c)
						}
						wg.Wait()
						p.working = true

					}()
				}

				if p.useTonAPI {
					i := 0
					for i < httpSendCnt {
						go func() {
							msg := nbot.BuildBundle(
								poolAddr,
								botInAmount,
								limit,
								nextLimit,
								uint64(deadline.Unix())+uint64(rand.Int63n(1000)),
								gAddr,
								p.v4walletMode,
							)
							msgList := []*wallet.Message{msg}
							msgList = append(msgList, msgTonApi...)

							err := utils.TimeitReturnError("sending with TONAPI", func() error {
								return p.SendWithTONAPI(&chance, nbot, msgList)
							})

							if err != nil {
								log.Error().Err(err).Msg("failed to send")
							}
						}()

						i++
					}
				}

				if p.useTonAPIBlockchain {
					i := 0
					for i < httpSendCnt {
						go func() {

							msg := nbot.BuildBundle(
								poolAddr,
								botInAmount,
								limit,
								nextLimit,
								uint64(deadline.Unix())+uint64(rand.Int63n(1000)),
								gAddr,
								p.v4walletMode,
							)
							msgList := []*wallet.Message{msg}
							msgList = append(msgList, msgTonApiBlockchain...)

							err := utils.TimeitReturnError("sending with TONAPI blockchain", func() error {
								return p.SendWithTONAPIBlockchain(&chance, nbot, msgList)
							})

							if err != nil {
								log.Error().Err(err).Msg("failed to send")
							}
						}()
						i++
					}
				}

				if err := chance.DumpToIO(p.out); err != nil {
					log.Error().Err(err).Msg("failed to write to file")
				}

				if err != nil {
					log.Error().Err(err).Msg("failed to send")
				}

				if err = p.SaveGActionInDB(chance.PrivateKeyOfG, gAddr.String()); err != nil {
					log.Error().Err(err).Msg("failed to save g action")
				}

				timer.Reset(60 * time.Second)
			}

		case <-timer.C:
			log.Debug().Msg("timer due")
			timer.Reset(6000000 * time.Second)
			if p.working {
				p.working = false
			}

		case <-reconnectPoolTicker.C:
			log.Debug().Msg("reconnect pool now")
			if !p.working {
				p.pool, p.ctx, err = utils.Reconnect(p.tonConfig)
				p.client = utils.GetAPIClient(p.pool)
			}
		case <-ticker.C:
			fmt.Printf("[+] T: %s B: %s, S: %d, W: %t\n",
				time.Now().Format(time.Kitchen),
				p.balance.String(), p.seqno, p.working)
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
	if len(chance.PrivateKeyOfG) == 0 {
		log.Debug().Msg("[-] SKIP, no private key for G")
		return false
	}

	chanceAddr := address.MustParseAddr(chance.VictimAccountId)
	if chanceAddr.String() == p.addr.String() {
		log.Debug().Msg("[-] SKIP, victim is me")
		return false
	}

	if _, found := lo.Find(p.skippedPoolAddress, func(s string) bool {
		return address.MustParseAddr(s).String() == address.MustParseAddr(chance.PoolAddress).String()
	}); found {
		log.Debug().Msg("[-] SKIP, pool is in skipped list")
		return false
	}

	if _, found := lo.Find(p.skippedVictimAccountId, func(s string) bool {
		return address.MustParseAddr(s).String() == address.MustParseAddr(chance.VictimAccountId).String()
	}); found {
		log.Debug().Msg("[-] SKIP, victim is in skipped list")
		return false
	}

	in := stringToBigInt(chance.BotIn)
	// 如果余额不足
	if p.balance.Nano().Cmp(in) < 0 {
		log.Debug().Msg("[-] SKIP, balance not enough")
		return false
	}

	if in.Cmp(p.upperlimit) > 0 {
		log.Debug().Msg("[-] SKIP, in amount too big")
		return false
	}

	if in.Cmp(p.lowerlimit) < 0 {
		log.Debug().Msg("[-] SKIP, in amount too small")
		return false
	}

	profit := stringToBigInt(chance.Profit)
	if profit.Cmp(tlb.MustFromTON("0.3").Nano()) < 0 {
		return false
	}

	if profit.Cmp(tlb.MustFromTON("1").Nano()) > 0 {
		return true
	}

	if st(in, "20") && bt(profit, "0.3") {
		return true
	}

	if st(in, "50") && bt(profit, "0.4") {
		return true
	}

	if st(in, "100") && bt(profit, "1") {
		return true
	}

	if st(in, "150") && bt(profit, "1.1") {
		return true
	}

	return false
}

func bt(a *big.Int, b string) bool {
	bi := tlb.MustFromTON(b)
	return a.Cmp(bi.Nano()) > 0
}

func st(a *big.Int, b string) bool {
	bi := tlb.MustFromTON(b)
	return a.Cmp(bi.Nano()) < 0
}

func stringToBigInt(s string) *big.Int {
	i := new(big.Int)
	i.SetString(s, 10)
	return i
}

func (p *Printer) SendWithANDL(
	chance *model.BundleChance,
	nbot *bot.Wallet,
	msgs []*wallet.Message,
) error {
	var err error
	ctxSlices := make([]context.Context, 0)

	i := 0
	nodeCtx := context.WithValue(context.Background(), "foo", struct{}{})
	for i < int(p.sendCnt) {
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
			err := utils.TimeitReturnError(fmt.Sprintf("send with andl %d [%d]", nid, i), func() error {
				count := 0
				for count < 10 {
					nbot.SendMany(c, p.v4walletMode, msgs, false)
					count++
				}
				return nil
			})

			if err != nil {
				log.Error().Err(err).Msgf("failed to send bundle %d", i)
			}
		}(i, c)
	}
	wg.Wait()
	p.working = true

	return nil
}

func (p *Printer) SendWithTONAPIBlockchain(chance *model.BundleChance,
	nbot *bot.Wallet,
	msgs []*wallet.Message) error {
	c := context.Background()

	externalMsg, err := nbot.BuildExternalMessageForMany(c,
		p.v4walletMode,
		msgs)
	if err != nil {
		return err
	}

	cell, err := tlb.ToCell(externalMsg)
	if err != nil {
		return err
	}

	var body struct {
		Boc string `json:"boc"`
	}
	body.Boc = base64.StdEncoding.EncodeToString(cell.ToBOC())

	buf := bytes.NewBuffer(nil)
	if err := json.NewEncoder(buf).Encode(body); err != nil {
		return err
	}

	req, err := http.NewRequest("POST", "https://tonapi.io/v2/blockchain/message", buf)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", TOKEN))
	clientTraceCtx := httptrace.WithClientTrace(req.Context(), p.httpTrace)
	resp, err := p.httpClt.Do(req.WithContext(clientTraceCtx))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	log.Debug().Msgf("tonapi blockchain resp status: %d", resp.StatusCode)
	if resp.StatusCode != http.StatusOK {
		content, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		log.Debug().Msgf("toncenter v3 resp content: %s", content)
	}

	return nil
}

func (p *Printer) SendWithTONAPI(chance *model.BundleChance,
	nbot *bot.Wallet,
	msgs []*wallet.Message) error {
	c := context.Background()

	externalMsg, err := nbot.BuildExternalMessageForMany(c,
		p.v4walletMode,
		msgs)
	if err != nil {
		return err
	}

	cell, err := tlb.ToCell(externalMsg)
	if err != nil {
		return err
	}

	var body struct {
		Body string `json:"body"`
	}
	body.Body = base64.StdEncoding.EncodeToString(cell.ToBOC())

	buf := bytes.NewBuffer(nil)
	if err := json.NewEncoder(buf).Encode(body); err != nil {
		return err
	}

	req, err := http.NewRequest("POST", "https://tonapi.io/v2/liteserver/send_message", buf)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", TOKEN))
	clientTraceCtx := httptrace.WithClientTrace(req.Context(), p.httpTrace)
	resp, err := p.httpClt.Do(req.WithContext(clientTraceCtx))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		content, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		log.Debug().Msgf("toncenter v3 resp content: %s", content)
	}
	log.Debug().Msgf("tonapi resp status: %d", resp.StatusCode)
	return nil
}

func (p *Printer) SendWithTONCenter(chance *model.BundleChance,
	nbot *bot.Wallet,
	msgs []*wallet.Message) error {
	c := context.Background()

	externalMsg, err := nbot.BuildExternalMessageForMany(c,
		p.v4walletMode,
		msgs,
	)
	if err != nil {
		return err
	}

	cell, err := tlb.ToCell(externalMsg)
	if err != nil {
		return err
	}

	var body struct {
		Body string `json:"boc"`
	}
	body.Body = base64.StdEncoding.EncodeToString(cell.ToBOC())

	buf := bytes.NewBuffer(nil)
	if err := json.NewEncoder(buf).Encode(body); err != nil {
		log.Error().Err(err).Msg("failed to encode body")
	}

	req, err := http.NewRequest("POST", "https://toncenter.com/api/v2/sendBoc", buf)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-Api-Key", TONCENTER_API_KEY)
	clientTraceCtx := httptrace.WithClientTrace(req.Context(), p.httpTrace)
	resp, err := p.httpClt.Do(req.WithContext(clientTraceCtx))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	log.Debug().Msgf("toncenter resp status: %d", resp.StatusCode)
	if resp.StatusCode != http.StatusOK {
		content, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		log.Debug().Msgf("toncenter v3 resp content: %s", content)
	}
	return nil
}

func (p *Printer) SendWithTONCenterV3(chance *model.BundleChance,
	nbot *bot.Wallet,
	msgs []*wallet.Message) error {
	c := context.Background()

	externalMsg, err := nbot.BuildExternalMessageForMany(c,
		p.v4walletMode,
		msgs,
	)
	if err != nil {
		return err
	}

	cell, err := tlb.ToCell(externalMsg)
	if err != nil {
		return err
	}

	var body struct {
		Body string `json:"boc"`
	}
	body.Body = base64.StdEncoding.EncodeToString(cell.ToBOC())

	buf := bytes.NewBuffer(nil)
	if err := json.NewEncoder(buf).Encode(body); err != nil {
		log.Error().Err(err).Msg("failed to encode body")
	}

	req, err := http.NewRequest("POST", "https://toncenter.com/api/v3/message", buf)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-Api-Key", TONCENTER_API_KEY)
	clientTraceCtx := httptrace.WithClientTrace(req.Context(), p.httpTrace)
	resp, err := p.httpClt.Do(req.WithContext(clientTraceCtx))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	log.Debug().Msgf("toncenter v3 resp status: %d", resp.StatusCode)
	if resp.StatusCode != http.StatusOK {
		content, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		log.Debug().Msgf("toncenter v3 resp content: %s", content)
	}
	return nil
}

func (p *Printer) BuildAuxTransfer(bw *bot.Wallet, amount tlb.Coins, comment string) (*wallet.Message, error) {
	return bw.BuildTransfer(DCM_ADDR, amount, true, comment)
}

func (p *Printer) SaveGActionInDB(pk, addr string) error {
	bundle := model.Bundle{
		Address:    addr,
		PrivateKey: pk,
	}

	return bundle.SaveToDB(p.db)
}
