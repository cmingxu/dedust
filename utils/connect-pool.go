package utils

import (
	"context"
	"os"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/xssnick/tonutils-go/liteclient"
	"github.com/xssnick/tonutils-go/ton"
)

var (
	pool *liteclient.ConnectionPool
	ctx  context.Context
)

func GetConnectionPool(urlOrFile string) (pool *liteclient.ConnectionPool,
	ctx context.Context,
	err error) {
	if pool != nil {
		return pool, ctx, nil
	}

	pool = liteclient.NewConnectionPool()

	var connectErr error
	if _, err := os.Stat(urlOrFile); err != nil && os.IsNotExist(err) {
		connectErr = pool.AddConnectionsFromConfigUrl(context.Background(), urlOrFile)
	} else {
		connectErr = pool.AddConnectionsFromConfigFile(urlOrFile)
	}

	if connectErr != nil {
		return nil, nil, connectErr
	}

	// bound all requests to single ton node

	go func() {
		ticker := time.NewTicker(50 * time.Millisecond)
		defer ticker.Stop()

		client := ton.NewAPIClient(pool, ton.ProofCheckPolicyUnsafe).WithRetry()
		for range ticker.C {
			_, err := client.GetTime(context.Background())
			if err != nil {
				log.Error().Err(err).Msg("*** failed to get time from node ***")
			}
		}
	}()

	return pool, context.Background(), nil
}

func GetAPIClient(pool *liteclient.ConnectionPool) ton.APIClientWrapped {
	return ton.NewAPIClient(pool, ton.ProofCheckPolicyUnsafe).WithRetry()
}

func GetAPIClientWithTimeout(pool *liteclient.ConnectionPool, timeout time.Duration) ton.APIClientWrapped {
	return ton.NewAPIClient(pool, ton.ProofCheckPolicyUnsafe).WithRetry().WithTimeout(timeout)
}
