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

		log.Printf("Config: %s", cs)
		log.Printf("Patrol: %s", p)
		if !ctx.Bool("no-compact") {
			p.History.Compact()
		}

		p.Close()
		return nil
	},
}

func sliceContains(list []string, str string) bool {
	if len(list) == 0 {
		return true
	}
	for _, elm := range list {
		if elm == str {
			return true
		}
	}
	return false
}

var cmdList = &cli.Command{
	Name:    "list",
	Aliases: []string{"ls"},
	Usage:   "List records from data file.",
	Flags: []cli.Flag{
		configFlag,
		&cli.StringSliceFlag{
			Name:  "group",
			Usage: "Filter by group name",
		},
		&cli.StringSliceFlag{
			Name:  "check",
			Usage: "Filter by check name",
		},
		&cli.StringSliceFlag{
			Name:  "type",
			Usage: "Filter by check type (boolean, metric)",
		},
		&cli.StringSliceFlag{
			Name:  "status",
			Usage: "Filter by status name",
		},
		&cli.IntFlag{
			Name:    "count",
			Aliases: []string{"c"},
			Usage:   "Max number of matches to print",
		},
	},
	Action: func(ctx *cli.Context) error {
		p, _, err := patrol.FromConfigFile(ctx.String("config"), nil)
		if err != nil {
			return err
		}

		groupFilter := ctx.StringSlice("group")
		checkFilter := ctx.StringSlice("check")
		typeFilter := ctx.StringSlice("type")
		statusFilter := ctx.StringSlice("status")
		maxMatches := ctx.Int("count")

		data := p.History.GetData()
		numMatches := 0
	outer:
		for groupName, group := range data {
			if sliceContains(groupFilter, groupName) {
				for checkName, items := range group {
					if sliceContains(checkFilter, checkName) {
						for _, item := range items {
							if sliceContains(typeFilter, item.Type) && sliceContains(statusFilter, item.Status) {
								fmt.Printf("-\n%s\n", item)
								numMatches++

								if numMatches >= maxMatches {
									break outer
								}
							}
						}
					}
				}
			}
		}
		fmt.Printf("-\n")

		p.Close()
		return nil
	},
}

func main() {
	app := &cli.App{
		Name:  "patrol",
		Usage: "Host your own statuspages.",
		Commands: []*cli.Command{
			cmdCheckConfig,
			cmdRun,
			cmdList,
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
