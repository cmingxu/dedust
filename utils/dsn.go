package utils

import (
	"fmt"

	cli2 "github.com/urfave/cli/v2"
)

func ConstructDSN(ctx *cli2.Context) string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true", ctx.String("user"), ctx.String("password"), ctx.String("host"), ctx.Int("port"), ctx.String("database"))
}
