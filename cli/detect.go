package cli

import (
	"os"

	"github.com/cmingxu/dedust/detector"
	"github.com/cmingxu/dedust/utils"

	cli2 "github.com/urfave/cli/v2"
)

func detect(c *cli2.Context) error {
	outFile, err := os.OpenFile(c.String("detect-output"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer outFile.Close()

	d, err := detector.NewDetector(utils.ConstructDSN(c), c.String("ton-config"),
		outFile)
	if err != nil {
		return err
	}

	return d.Run(c.Bool("pre-update-reserve"))
}
