package cli

import (
	"github.com/cmingxu/dedust/detector"
	"github.com/cmingxu/dedust/utils"

	cli2 "github.com/urfave/cli/v2"
)

var memPoolCmd = &cli2.Command{
	Name:        "mem-pool",
	Description: "to check tonapi mempool",
	Flags: []cli2.Flag{
		&host,
		&port,
		&user,
		&password,
		&database,
	},
	Action: func(c *cli2.Context) error {
		if err := utils.SetupLogger(c.String("loglevel")); err != nil {
			return err
		}
		return memPool(c)
	},
}

func memPool(c *cli2.Context) error {
	return detector.MemPoolCheck(utils.ConstructDSN(c))
}
