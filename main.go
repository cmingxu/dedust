package main

import (
	"os"

	"github.com/cmingxu/dedust/cli"
)

func main() {
	os.Exit(cli.Run(os.Args))

}
