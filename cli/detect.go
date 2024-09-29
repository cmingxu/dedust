package cli

import (

	"github.com/cmingxu/dedust/detector"
	"github.com/cmingxu/dedust/utils"

	cli2 "github.com/urfave/cli/v2"
)

func detect(c *cli2.Context) error {
	d, err := detector.NewDetector(utils.ConstructDSN(c))
	if err != nil {
		return err
	}

	return d.Run()
}

