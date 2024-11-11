package cli

import (
	"os"

	"github.com/cmingxu/dedust/detector"
	"github.com/cmingxu/dedust/utils"
	"github.com/xssnick/tonutils-go/tlb"

	cli2 "github.com/urfave/cli/v2"
)

var (
	detectCmd = &cli2.Command{
		Name:        "detect",
		Description: "to detect and save dedust trading infomation",
		Flags: []cli2.Flag{
			&host,
			&port,
			&user,
			&password,
			&database,
			&tonConfig,
			&preUpdateReserve,
			&output,
			&terminiator,
		},
		Action: func(c *cli2.Context) error {
			if err := utils.SetupLogger(c.String("loglevel")); err != nil {
				return err
			}
			return detect(c)
		},
	}
)

func detect(c *cli2.Context) error {
	outFile, err := os.OpenFile(c.String("detect-output"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer outFile.Close()

	terminiator := tlb.MustFromTON(c.String("terminator"))

	d, err := detector.NewDetector(utils.ConstructDSN(c),
		c.String("ton-config"),
		outFile, terminiator)
	if err != nil {
		return err
	}

	return d.Run(c.Bool("pre-update-reserve"))
}
