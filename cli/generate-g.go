package cli

import (
	"crypto/ed25519"
	"database/sql"
	"sync"
	"encoding/hex"
	"fmt"

	"github.com/cmingxu/dedust/model"
	"github.com/cmingxu/dedust/utils"
	"github.com/cmingxu/dedust/wallet"
	"github.com/cmingxu/dedust/bot"
	"github.com/jmoiron/sqlx"
	"github.com/samber/lo"
	"github.com/rs/zerolog/log"
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/tvm/cell"
	cli2 "github.com/urfave/cli/v2"
)

func GenGForPools(c *cli2.Context) error {
	db, err := sqlx.Connect("mysql", utils.ConstructDSN(c))
	if err != nil {
		return err
	}
	defer db.Close()

	pools, err := model.LoadPoolsFromDB(db, true)
	if err != nil {
		return err
	}

	botWalletSeeds := MustLoadSeeds(c.String("bot-wallet-seed"))
	pk := pkFromSeed(botWalletSeeds)
	botAddr := bot.BotAddress(pk.Public().(ed25519.PublicKey))

	num := 50
	wg := sync.WaitGroup{}
	poolsGroup := lo.Chunk(pools, num)

	for i, group := range poolsGroup {
		wg.Add(len(group))
		for x, pool := range group {
			go func(i int, pool *model.Pool) {
				defer wg.Done()
				fmt.Println("====================================  ", i)
				//. fmt.Println("address", pool.Address)
				//. fmt.Println("vault", pool.Asset1Vault)
				//. fmt.Println("master", pool.Asset1Address)
				//. fmt.Println("wallet code hash", pool.Asset1TokenWalletCode)

				code, _ := model.WalletCodeBOCs[pool.Asset1TokenWalletCode]
				content, _ := hex.DecodeString(code)
				codeCell, _ := cell.FromBOC(content)
				ownerAddr := address.MustParseAddr(pool.Asset1Vault)
				jettonMasterAddr := address.MustParseAddr(pool.Asset1Address)

				jettonWalletCell := wallet.CalculateUserJettonWalletAddress(
					ownerAddr,
					jettonMasterAddr,
					codeCell,
				)

				// fmt.Println("jetton wallet", utils.CellToAddress(jettonWalletCell).String())

				pool.Asset1VaultJettonWalletAddress.String = utils.CellToAddress(jettonWalletCell).String()

				vault1Addr := address.MustParseAddr(pool.Asset1Vault)
				asset1JettonMasterAddr := address.MustParseAddr(pool.Asset1Address)
				vault1JettonWalletAddr := address.MustParseAddr(pool.Asset1VaultJettonWalletAddress.String)
				code, ok := model.WalletCodeBOCs[pool.Asset1TokenWalletCode]
				if !ok {
					fmt.Println("code not found")
					log.Error().Str("code", pool.Asset1TokenWalletCode).Msg("code not found")
					return
				}

				pk, gAddr, err := bot.BuildGBestFitInShard(
					botAddr,
					vault1Addr,
					asset1JettonMasterAddr,
					vault1JettonWalletAddr,
					codeCell,
				)
				if err != nil {
					fmt.Println("failed to generate private key for G")
					log.Error().Err(err).Msg("failed to generate private key for G")
					return
				}

				pool.GAddr = sql.NullString{String: gAddr.String(), Valid: true}
				pool.PrivateKeyOfG = sql.NullString{String: hex.EncodeToString(pk), Valid: true}

				 // fmt.Println("GAddr", gAddr.String())
				fmt.Println("GPrivateKey", hex.EncodeToString(pk))

				if err = pool.UpdateG(db); err != nil {
					fmt.Println("failed to save pool to db")
					log.Error().Err(err).Msg("failed to save pool to db")
					return
				}

			}(i * 50 + x, pool)
		}

		wg.Wait()
	}

	return nil
}
