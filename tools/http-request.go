package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"
)

func main2() {
	dialer := &net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
		DualStack: true,
	}
	http.DefaultTransport.(*http.Transport).DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
		fmt.Println("address original =", addr)
		if addr == "toncenter.com:443" {
			addr = "104.26.0.179:443"
			fmt.Println("address modified =", addr)
		}
		return dialer.DialContext(ctx, network, addr)
	}

	req, err := http.NewRequest("GET", "https://toncenter.com/api/v2/getMasterchainInfo", nil)
	if err != nil {
		log.Fatal(err)
	}

	req.Header.Set("User-Agent", "tonup")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	log.Println(resp.Header, err)

}
