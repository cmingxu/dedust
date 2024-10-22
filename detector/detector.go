package detector

import (
	"context"
	"io"
	"sync"
	"time"

	"github.com/cmingxu/dedust/model"
	"github.com/cmingxu/dedust/utils"

	"github.com/jmoiron/sqlx"
	"github.com/patrickmn/go-cache"
	"github.com/rs/zerolog/log"
	"github.com/xssnick/tonutils-go/liteclient"
	"github.com/xssnick/tonutils-go/tlb"
	"github.com/xssnick/tonutils-go/ton"
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

	// api client
	apiClient ton.APIClientWrapped
	connPool  *liteclient.ConnectionPool
	apiCtx    context.Context

	// cache chance 防止重复的交易产生的机会信号
	chanceCache *cache.Cache

	// selling cache 防止最新的 selling 对 reserve 产生影响
	sellingCache *cache.Cache

	// cooldown cache 防止在相同 pool 上的过于频繁的信号
	cooldownCache *cache.Cache

	out io.Writer
}

func NewDetector(dsn string, tonConfig string, out io.Writer) (*Detector, error) {
	var err error
	detector := &Detector{
		db: nil,

		poolLock:       sync.RWMutex{},
		poolMap:        make(map[string]*model.Pool),
		poolRenewTimer: time.NewTimer(PoolRenewInterval),

		stopChan: make(chan struct{}),
		stopOnce: sync.Once{},

		out: out,
	}

	detector.db, err = sqlx.Connect("mysql", dsn)
	if err != nil {
		return nil, err
	}

	detector.connPool, detector.apiCtx, err = utils.GetConnectionPool(tonConfig)
	if err != nil {
		return nil, err
	}

	detector.apiClient = utils.GetAPIClient(detector.connPool)

	// a cache expire at 5s and purge at 10s
	detector.chanceCache = cache.New(5*time.Second, 1*time.Second)

	// 会将某个 pool 最近 45 的 sell 缓存起来
	detector.sellingCache = cache.New(45*time.Second, 1*time.Second)

	// 如果多次出现 chance，则该 pool 会被冷却 90s
	detector.cooldownCache = cache.New(90*time.Second, 1*time.Second)

	return detector, nil
}

func (d *Detector) Run(preUpdate bool) error {
	if err, _ := d.renewPoolsFromDB(d.db); err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())

	bundleChanceCh := make(chan *model.BundleChance, 10)
	mpResponseCh := make(chan *MPResponse, 10)

	poolsRenewedCh := make(chan struct{}, 1)
	poolsRenewedCh <- struct{}{}

	go func() {
		if err := d.PoolReserveConsumer(ctx); err != nil {
			log.Error().Err(err).Msg("failed to subscribe to pool reserve")
		}
	}()

	if preUpdate {
		if err := d.PoolReserveUpdater(ctx); err != nil {
			log.Error().Err(err).Msg("failed to update pool reserve")
		}
	}

	// fetching pool information from DB and update those in memory
	go d.PerodicallyRenewPoolsFromDB(ctx, poolsRenewedCh)

	go func() {
		// subscribe to mempool, should reconnect websocket upon poolsRenewedCh signal
		if err := d.SubscribeTradeSignalFromTonAPIMemPool(ctx, poolsRenewedCh, mpResponseCh); err != nil {
			log.Error().Err(err).Msg("failed to subscribe to mempool")
		}
	}()

	go func() {
		if err := d.RunWSServer(ctx, bundleChanceCh); err != nil {
			log.Error().Err(err).Msg("failed to run websocket server")
		}
	}()

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
			log.Warn().Msgf("main pool %s not found", mpResponse.InvolvedAccounts)
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

		// trade is return even err is not nil
		trade, err := d.parseTrade(pool, outMessage.AsExternalIn())
		if chance, err := d.BuildBundleChance(pool, trade); err == nil {
			log.Debug().Msgf("BundleChance %+v", chance)

			if _, found := d.cooldownCache.Get(pool.Address); found {
				log.Debug().Msg("cooldown cache hit")
				continue
			}

			d.cooldownCache.Set(pool.Address, struct{}{}, cache.DefaultExpiration)
			bundleChanceCh <- chance
		} else {
			log.Debug().Err(err).Msg("failed to build bundle chance")
		}

		if err != nil {
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
