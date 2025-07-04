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
	poolLock       sync.RWMutex
	poolMap        map[string]*model.Pool

	// stop signal
	stopChan chan struct{}
	stopOnce sync.Once

	// api client ton pool
	apiClient ton.APIClientWrapped
	connPool  *liteclient.ConnectionPool
	apiCtx    context.Context

	// cache chance 防止重复的交易产生的机会信号
	chanceCache *cache.Cache
	// selling cache 防止最新的 selling 对 reserve 产生影响
	sellingCache *cache.Cache
	// cooldown cache 防止在相同 pool 上的过于频繁的信号
	cooldownCache *cache.Cache
	// ton transfer cache
	tonTransferCache *cache.Cache

	// dump sink
	out io.Writer

	terminator tlb.Coins

	tonapiIP string
}

func NewDetector(dsn string, tonConfig string, out io.Writer,
	terminator tlb.Coins,
	tonapiIp string,
) (*Detector, error) {
	var err error
	detector := &Detector{
		db: nil,

		poolLock:       sync.RWMutex{},
		poolMap:        make(map[string]*model.Pool),
		poolRenewTimer: time.NewTimer(PoolRenewInterval),

		stopChan: make(chan struct{}),
		stopOnce: sync.Once{},

		out:        out,
		terminator: terminator,
		tonapiIP:   tonapiIp,
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
	detector.chanceCache = cache.New(30*time.Second, 1*time.Second)
	// 会将某个 pool 最近 45 的 sell 缓存起来
	detector.sellingCache = cache.New(45*time.Second, 1*time.Second)
	// 如果多次出现 chance，则该 pool 会被冷却 45s
	detector.cooldownCache = cache.New(30*time.Second, 1*time.Second)
	// 用于缓存 ton transfer
	detector.tonTransferCache = cache.New(30*time.Second, 1*time.Second)

	return detector, nil
}

func (d *Detector) Run(preUpdate bool) error {
	if err := d.renewPoolsFromDB(d.db); err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())

	bundleChanceCh := make(chan *model.BundleChance, 10)
	mpResponseCh := make(chan *MPResponse, 10)

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

	go func() {
		ticker := time.NewTicker(time.Second * 600)
		defer ticker.Stop()

		for {
			select {
			case <-d.stopChan:
				cancel()
				log.Info().Msg("detector stopped due to stopChan")
				return
			case <-ticker.C:
				if err := d.PoolReserveUpdater(ctx); err != nil {
					log.Error().Err(err).Msg("failed to update pool reserve")
				}
			}
		}
	}()

	// fetching pool information from DB and update those in memory
	go d.PerodicallyRenewPoolsFromDB(ctx)

	go func() {
		// subscribe to mempool, should reconnect websocket upon poolsRenewedCh signal
		if err := d.SubscribeTradeSignalFromTonAPIMemPool(mpResponseCh); err != nil {
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

		//	log.Debug().Msgf("received mempool response %+s", mpResponse.ShortString())

		go func(mp *MPResponse) {
			if len(mp.Boc) == 0 {
				log.Warn().Msg("empty BOC")
				return
			}

			outMessage, err := d.outerMessageFromBOC(mp.Boc)
			if err != nil {
				log.Warn().Err(err).Msg("failed to parse external message")
				return
			}

			if tlb.MsgTypeExternalIn != outMessage.MsgType {
				log.Warn().Msg("not external message")
				return
			}

			// trade is return even err is not nil
			pool, trade, err := d.parseTrade(outMessage.AsExternalIn(), mp.Boc)
			if err != nil {
				return
			}

			if pool == nil {
				return
			}

			if chance, err := d.BuildBundleChance(pool, trade); err == nil {
				log.Debug().Msgf("BundleChance %+v", chance)

				if _, found := d.cooldownCache.Get(pool.Address); found {
					log.Debug().Msg("cooldown cache hit")
				}

				d.cooldownCache.Set(pool.Address, struct{}{}, cache.DefaultExpiration)
				bundleChanceCh <- chance
			} else {
				log.Debug().Err(err).Msg("failed to build bundle chance")

			}

			if err != nil {
				log.Err(err).Msg("failed to handle external message")
			}
		}(<-mpResponseCh)
	}
}

func (d *Detector) Stop() {
	d.stopOnce.Do(func() {
		close(d.stopChan)
	})
}
