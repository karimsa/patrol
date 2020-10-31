package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/karimsa/patrol"
	"github.com/urfave/cli/v2"
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
		p, config, err := patrol.FromConfigFile(ctx.String("config"), nil)
		if err != nil {
			return err
		}

		cs, err := json.MarshalIndent(config, "", "\t")
		if err != nil {
			return err
		}

		log.Printf("Config: %s\n", cs)
		p.Start()

		sigInt := make(chan os.Signal, 1)
		signal.Notify(sigInt, os.Interrupt)
		<-sigInt

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
		&cli.BoolFlag{
			Name:  "no-compact",
			Usage: "If specified, compaction is skipped.",
			Value: false,
		},
	},
	Action: func(ctx *cli.Context) error {
		p, config, err := patrol.FromConfigFile(ctx.String("config"), nil)
		if err != nil {
			return err
		}

		cs, err := json.MarshalIndent(config, "", "\t")
		if err != nil {
			return err
		}

		log.Printf("Config: %s\n", cs)
		if !ctx.Bool("no-compact") {
			p.Compact()
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
