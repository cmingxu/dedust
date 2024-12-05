package cli

import (
	"context"
	"crypto/ed25519"
	"database/sql"
	"encoding/hex"
	"fmt"
	"net/http"
	"time"

	"github.com/cmingxu/dedust/bot"
	"github.com/cmingxu/dedust/model"
	"github.com/cmingxu/dedust/utils"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"
	cli2 "github.com/urfave/cli/v2"
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/tvm/cell"
)

const DedustPoolAPI = "https://api.dedust.io/v2/pools"

var (
	syncPoolCmd = &cli2.Command{
		Name: "sync-pool",
		Flags: []cli2.Flag{
			&host,
			&port,
			&user,
			&password,
			&database,
			&tonConfig,
			&walletSeed,
			&generateG,
		},
		Action: func(c *cli2.Context) error {
			if err := utils.SetupLogger(c.String("loglevel")); err != nil {
				return err
			}
			return syncPool(c)
		},
	}
)

func syncPool(c *cli2.Context) error {
	fmt.Println("Syncing pool...")

	botWalletSeeds := MustLoadSeeds(c.String("wallet-seed"))
	pk := pkFromSeed(botWalletSeeds)
	botAddr := bot.WalletAddress(pk.Public().(ed25519.PublicKey), nil, bot.Bot)

	db, err := sqlx.Connect("mysql", utils.ConstructDSN(c))
	if err != nil {
		return err
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	body, err := utils.Request(ctx, http.MethodGet, DedustPoolAPI, nil)
	if err != nil {
		return err
	}

	pools, err := model.LoadPoolsFromJSON(body)
	if err != nil {
		return err
	}

	connPool, ctx, err := utils.GetConnectionPool(c.String("ton-config"))
	if err != nil {
		return err
	}
	client := utils.GetAPIClient(connPool)
	masterBlock, err := client.GetMasterchainInfo(ctx)
	if err != nil {
		return err
	}

	for i, pool := range pools {
		ok, err := pool.ExistsInDB(db)
		if err != nil {
			log.Err(err).Msg("failed to check if pool exists in db")
			continue
		}

		if ok {
			log.Debug().Str("pool", pool.Address).Msgf("#%d pool already exists", i)
			continue
		}

		err = pool.FetchAssetMasterCode(ctx, connPool, masterBlock)
		if err != nil {
			log.Err(err).Msg("failed to fetch asset code")
			continue
		}

		err = pool.FetchVaultAddress(ctx, client, masterBlock)
		if err != nil {
			log.Err(err).Msg("failed to fetch vault address")
			continue
		}

		err = pool.FetchAssetWalletCode(ctx, client, masterBlock)
		if err != nil {
			log.Err(err).Msg("failed to fetch asset wallet code")
			continue
		}

		if err = pool.GenerateVault1JettonWalletAddress(); err != nil {
			log.Err(err).Msg("failed to generate vault1 jetton wallet address")
			continue
		}

		if c.Bool("generate-g") {
			vault1Addr := address.MustParseAddr(pool.Asset1Vault)
			asset1JettonMasterAddr := address.MustParseAddr(pool.Asset1Address)
			vault1JettonWalletAddr := address.MustParseAddr(pool.Asset1VaultJettonWalletAddress.String)
			code, ok := model.WalletCodeBOCs[pool.Asset1TokenWalletCode]
			if !ok {
				return fmt.Errorf("asset1 token wallet code not found")
			}

			content, _ := hex.DecodeString(code)
			codeCell, _ := cell.FromBOC(content)

			pk, gAddr, err := bot.BuildGBestFitInShard(
				botAddr,
				vault1Addr,
				asset1JettonMasterAddr,
				vault1JettonWalletAddr,
				codeCell,
			)

			if err != nil {
				log.Err(err).Msg("failed to build gbestfitinshard")
				continue
			}

			pool.PrivateKeyOfG = sql.NullString{String: hex.EncodeToString(pk), Valid: true}
			pool.GAddr = sql.NullString{String: gAddr.String(), Valid: true}
		}

		err = pool.SaveToDB(db)
		if err != nil {
			log.Err(err).Msg("failed to save pool to db")
			return err
		}

		log.Debug().Str("pool", pool.Address).Msgf("#%d pool saved", i)
	}

	return nil
}
