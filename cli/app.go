package cli

import (
	"fmt"
	"os"

	cli2 "github.com/urfave/cli/v2"
)

var (
	flagLogLevel = cli2.StringFlag{
		Name:  "loglevel",
		Value: "info",
		Usage: "Set the log level (debug, info, warn, error, fatal, panic)",
	}

	host = cli2.StringFlag{
		Name:  "host",
		Value: "localhost",
	}

	port = cli2.IntFlag{
		Name:  "port",
		Value: 8080,
	}

	user = cli2.StringFlag{
		Name:  "user",
		Value: "root",
	}

	password = cli2.StringFlag{
		Name:  "password",
		Value: "password",
	}
)

func Run(args []string) int {
	fmt.Println("Running CLI with args: ", args)

	app := &cli2.App{
		Name:  "dedust",
		Usage: "A CLI tool to dedust your code",
		Flags: []cli2.Flag{
			&flagLogLevel,
		},
		Commands: []*cli2.Command{
			{
				Name:  "info",
				Flags: []cli2.Flag{&host, &port, &user, &password},
				Action: func(c *cli2.Context) error {
					fmt.Printf("Info command: %s\n", dbURL(*c))
					return nil
				},
			},
		},
	}

	if err := app.Run(args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	return 0
}

func dbURL(ctx cli2.Context) string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/", ctx.String("user"), ctx.String("password"), ctx.String("host"), ctx.Int("port"))
}
