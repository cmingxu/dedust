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

func Request(ctx context.Context, method string, url string, body io.Reader) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return []byte{}, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "dedust-cli/1.0")
	client := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
		},
		Timeout: time.Second * 10,
	}

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
