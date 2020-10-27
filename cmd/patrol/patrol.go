package main

import (
	"fmt"
	"github.com/karimsa/patrol"
	"github.com/urfave/cli/v2"
	"os"
)

var (
	configFlag = &cli.PathFlag{
		Name:      "config",
		Usage:     "Path to config file",
		TakesFile: true,
		Required:  true,
	}
)

var cmdRun = &cli.Command{
	Name:  "run",
	Usage: "Run statuspage using given configuration file.",
	Flags: []cli.Flag{
		configFlag,
	},
	Action: func(ctx *cli.Context) error {
		p, err := patrol.FromConfigFile(ctx.String("config"), nil)
		if err != nil {
			return err
		}
		// ...
		p.Close()
		return nil
	},
}

var cmdCheckConfig = &cli.Command{
	Name:    "check-config",
	Aliases: []string{"c"},
	Usage:   "Validate statuspage configuration file. Data file will be created if it does not exist and will be compacted if it already exists.",
	Flags: []cli.Flag{
		configFlag,
	},
	Action: func(ctx *cli.Context) error {
		p, err := patrol.FromConfigFile(ctx.String("config"), nil)
		if err != nil {
			return err
		}
		p.Close()
		return nil
	},
}

func main() {
	app := &cli.App{
		Name:  "patrol",
		Usage: "Host your own statuspages.",
		Commands: []*cli.Command{
			cmdRun,
			cmdCheckConfig,
		},
		Authors: []*cli.Author{
			&cli.Author{
				Name:  "Karim Alibhai",
				Email: "karim@alibhai.co",
			},
		},
		Copyright: "(C) 2020-present Karim Alibhai",
	}
	err := app.Run(os.Args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}
