package cli

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/cmingxu/dedust/utils"
	cli2 "github.com/urfave/cli/v2"
	"github.com/xssnick/tonutils-go/liteclient"
)

func LiteserverIps(c *cli2.Context) error {
	config, err := liteclient.GetConfigFromFile(c.String("ton-config"))
	if err != nil {
		return err
	}

	for _, ls := range config.Liteservers {
		fmt.Fprintf(os.Stdout, "%s:%d\n", intToIp(uint64(ls.IP)), ls.Port)

		// if err := ipinfoio(intToIp(uint64(ls.IP))); err != nil {
		// 	return err
		// }
	}

	ctx := context.WithValue(context.Background(), "config", "")
	pool, _, err := utils.GetConnectionPool(c.String("ton-config"))

	client := utils.GetAPIClient(pool)

	time.Sleep(10 * time.Second)

	i := 0
	for i <= len(config.Liteservers) {
		ctx, err = pool.StickyContextNextNodeBalanced(ctx)
		if err != nil {
			return err
		}
		id := ctx.Value("_ton_node_sticky").(uint32)

		fmt.Printf("Node ID: %d, IP: %s\n", id, intToIp(uint64(id)))
		utils.Timeit("", func() {
			t, err := client.GetTime(ctx)
			if err != nil {
				fmt.Println(err)
				return
			}

			fmt.Printf("Time: %d\n", t)
		})

		i++
	}

	return nil
}

func intToIp(ip uint64) string {
	return fmt.Sprintf("%d.%d.%d.%d", byte(ip), byte(ip>>8), byte(ip>>16), byte(ip>>24))
}

func ipinfoio(ip string) error {
	out, err := exec.Command("curl", "-s", "ipinfo.io/"+ip).Output()
	if err != nil {
		return err
	}

	fmt.Println("IP info:", string(out))

	return nil
}
