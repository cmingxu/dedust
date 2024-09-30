package detector

import (
	"context"
	"crypto/ed25519"
	"crypto/hmac"
	"crypto/sha512"
	"sync"
	"time"

	"github.com/cmingxu/dedust/model"
	"github.com/cmingxu/dedust/utils"
	"golang.org/x/crypto/pbkdf2"

	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"
	"github.com/xssnick/tonutils-go/tlb"
	"github.com/xssnick/tonutils-go/ton"
)

const (
	_Iterations   = 100000
	_Salt         = "TON default seed"
	_BasicSalt    = "TON seed version"
	_PasswordSalt = "TON fast seed version"
)

type Detector struct {
	db *sqlx.DB

	// pool renew interval
	poolRenewTimer *time.Timer

	// pool lock
	poolLock sync.RWMutex
	poolMap  map[string]*model.Pool

	stopChan chan struct{}
	stopOnce sync.Once

	//////////////////////////
	bought  bool
	client  ton.APIClientWrapped
	poolCtx context.Context
	pk      ed25519.PrivateKey
}

func NewDetector(dsn string) (*Detector, error) {
	var err error
	detector := &Detector{
		db: nil,

		poolLock:       sync.RWMutex{},
		poolMap:        make(map[string]*model.Pool),
		poolRenewTimer: time.NewTimer(PoolRenewInterval),

		stopChan: make(chan struct{}),
		stopOnce: sync.Once{},
	}

	detector.db, err = sqlx.Connect("mysql", dsn)
	if err != nil {
		return nil, err
	}

	// establish connection to the server
	connPool, ctx, err := utils.GetConnectionPool("https://ton.org/global-config.json")
	if err != nil {
		return nil, err
	}
	client := utils.GetAPIClientWithTimeout(connPool, time.Second*10)

	detector.poolCtx = ctx
	detector.client = client

	seeds := "select test beauty clay matrix radar call dust apple crash master normal salmon message annual wagon repair business wet office stumble spike treat pause"
	// calculate new PK for new wallet
	mac := hmac.New(sha512.New, []byte(seeds))
	mac.Write([]byte(""))
	hash := mac.Sum(nil)

	p := pbkdf2.Key(hash, []byte(_BasicSalt), _Iterations/256, 1, sha512.New)
	if p[0] != 0 {
		panic("invalid new wallet seed")
	}
	detector.pk = ed25519.NewKeyFromSeed(pbkdf2.Key(hash, []byte(_Salt), _Iterations, 32, sha512.New))
	detector.bought = false

	return detector, nil
}

func (d *Detector) Run() error {
	if err, _ := d.renewPoolsFromDB(d.db); err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())

	mpResponseCh := make(chan *MPResponse, 10)

	poolsRenewedCh := make(chan struct{}, 1)
	poolsRenewedCh <- struct{}{}
	// fetching pool information from DB and update those in memory
	go d.PerodicallyRenewPoolsFromDB(ctx, poolsRenewedCh)
	go func() {
		// subscribe to mempool, should reconnect websocket upon poolsRenewedCh signal
		if err := d.SubscribeTradeSignalFromTonAPIMemPool(ctx, poolsRenewedCh, mpResponseCh); err != nil {
			log.Error().Err(err).Msg("failed to subscribe to mempool")
		}
	}()

	//go func() {
	//	if err := d.PoolReserveUpdater(ctx); err != nil {
	//		log.Error().Err(err).Msg("failed to subscribe to pool reserve")
	//	}
	//}()

	for {
		select {
		case <-d.stopChan:
			cancel()
			log.Info().Msg("detector stopped due to stopChan")
			return nil
		default:
		}

		mpResponse := <-mpResponseCh
		log.Debug().Msgf("received mempool response %+s", mpResponse.String())

		var pool *model.Pool
		for _, account := range mpResponse.InvolvedAccounts {
			if p, ok := d.poolMap[account]; ok {
				pool = p
				break
			}
		}

		if pool == nil {
			log.Warn().Msgf("pool %s not found", mpResponse.InvolvedAccounts)
			continue
		}

		if len(mpResponse.Boc) == 0 {
			log.Warn().Msg("empty BOC")
			continue
		}

		outMessage, err := d.outerMessageFromBOC(mpResponse.Boc)
		if err != nil {
			log.Warn().Err(err).Msg("failed to parse message")
			continue
		}

		if tlb.MsgTypeExternalIn != outMessage.MsgType {
			log.Warn().Msg("not external message")
			continue
		}

		if err := d.handleExternalMsg(pool, outMessage.AsExternalIn()); err != nil {
			log.Err(err).Msg("failed to handle external message")
			continue
		}
	}
}

func (d *Detector) Stop() {
	d.stopOnce.Do(func() {
		close(d.stopChan)
	})
}
