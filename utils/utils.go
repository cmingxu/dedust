package utils

import (
	"context"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/tlb"
)

var (
	client = &http.Client{
		Transport: &http.Transport{
			Proxy:           http.ProxyFromEnvironment,
			MaxIdleConns:    200,
			MaxConnsPerHost: 200,
		},
		Timeout: time.Second * 15,
	}
)

func Request(ctx context.Context, method string, url string, body io.Reader) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return []byte{}, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "dedust-cli/1.0")
	resp, err := client.Do(req)
	if err != nil {
		return []byte{}, err
	}

	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return []byte{}, err
	}

	return data, nil
}

func RawAddr(addr *address.Address) string {
	return fmt.Sprintf("%d:%s", addr.Workchain(), hex.EncodeToString(addr.Data()))
}

func CoinsToFloatTON(t tlb.Coins) float32 {
	p, _ := strconv.ParseFloat(t.String(), 32)
	return float32(p)
}
