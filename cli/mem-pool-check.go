package cli

import (
	"github.com/cmingxu/dedust/detector"
	"github.com/cmingxu/dedust/utils"

	cli2 "github.com/urfave/cli/v2"
)

func memPoolCheck(c *cli2.Context) error {
	return detector.MemPoolCheck(utils.ConstructDSN(c))
}
