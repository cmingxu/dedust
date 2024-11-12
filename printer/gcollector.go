package printer

import (
	"context"
	"crypto/ed25519"
	"time"

	"github.com/cmingxu/dedust/bot"
	"github.com/cmingxu/dedust/model"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"
	"github.com/samber/lo"
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/tlb"
	"github.com/xssnick/tonutils-go/ton"
	"github.com/xssnick/tonutils-go/ton/wallet"
	"github.com/xssnick/tonutils-go/tvm/cell"
)

type GCollector struct {
	db     *sqlx.DB
	ctx    context.Context
	client ton.APIClientWrapped
	botPk  ed25519.PrivateKey
}

func NewGCollector(ctx context.Context,
	client ton.APIClientWrapped,
	db *sqlx.DB,
	botPk ed25519.PrivateKey,
	destAddr *address.Address,
) *GCollector {
	c := &GCollector{
		ctx:    ctx,
		client: client,
		db:     db,
		botPk:  botPk,
	}

	return c

}

func (c *GCollector) Run() error {
	ticker := time.NewTicker(120 * time.Second)
	if err := c.collect(); err != nil {
		return err
	}

	for {
		select {
		case <-c.ctx.Done():
			return nil
		case <-ticker.C:
			if err := c.collect(); err != nil {
				return err
			}
		}
	}
}

func (c *GCollector) collect() error {
	log.Info().Msg("collecting now")

	botAddr := bot.WalletAddress(c.botPk.Public().(ed25519.PublicKey), nil, bot.Bot)

	bundles := []model.Bundle{}
	if err := c.db.Select(&bundles, "SELECT * FROM bundles WHERE withdraw = ? AND createdAt < ?", false, time.Now().Add(-time.Second*300)); err != nil {
		return err
	}

	bundlesLatest := []model.Bundle{}
	if err := c.db.Select(&bundlesLatest, "SELECT * FROM bundles WHERE withdraw = ? AND createdAt >= ?", false, time.Now().Add(-time.Second*300)); err != nil {
		return err
	}

	for _, bundle := range bundles {
		_, found := lo.Find(bundlesLatest, func(b model.Bundle) bool {
			return b.Address == bundle.Address
		})

		if found {
			continue
		}

		log.Info().Msgf("processing bundle %s", bundle.Address)
		gAddr := address.MustParseAddr(bundle.Address)

		masterBlock, err := c.client.GetMasterchainInfo(c.ctx)
		if err != nil {
			log.Error().Err(err).Msg("failed to get masterchain info")
			continue
		}

		seqno, err := bot.GetSeqno(c.ctx, c.client, masterBlock, botAddr)
		if err != nil {
			log.Error().Err(err).Msg("failed to get seqno")
			continue
		}

		acc, err := c.client.GetAccount(c.ctx, masterBlock, gAddr)
		if err != nil {
			log.Error().Err(err).Msg("failed to get account")
			continue
		}

		if acc.State == nil {
			log.Info().Msg("account not found")
			continue
		}

		if acc.State.Balance.Nano().Cmp(tlb.MustFromTON("0.01").Nano()) > 0 {
			nbot := bot.NewWallet(c.ctx, c.client, bot.Bot, c.botPk, nil, seqno)
			msgBody := cell.BeginCell().
				MustStoreUInt(0x474f86cd, 32).
				EndCell()

			msg := &wallet.Message{
				Mode: wallet.PayGasSeparately + wallet.IgnoreErrors,
				InternalMessage: &tlb.InternalMessage{
					IHRDisabled: true,
					Bounce:      false,
					DstAddr:     gAddr,
					Amount:      tlb.MustFromTON("0.1"),
					Body:        msgBody,
				},
			}

			if err = nbot.Send(c.ctx, 0, msg, true); err != nil {
				log.Error().Err(err).Msg("failed to send")
				continue
			}
		}

		if _, err := c.db.Exec("UPDATE bundles SET withdraw = ? WHERE address = ?", true, bundle.Address); err != nil {
			log.Error().Err(err).Msg("failed to update bundle")
			continue
		}
	}

	return nil
}
