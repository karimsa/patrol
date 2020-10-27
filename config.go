package patrol

import (
	"fmt"
	"io/ioutil"
	"time"

	"github.com/karimsa/patrol/internal/checker"
	"github.com/karimsa/patrol/internal/history"
	"gopkg.in/yaml.v2"
)

type notificationsRaw struct {
	OnFailure []struct {
		Type    string
		Options interface{}
	} `yaml:"on_failure"`
	OnSuccess []struct {
		Type    string
		Options interface{}
	} `yaml:"on_success"`
}

type configRaw struct {
	Name     string
	Port     int
	DB       string `yaml:"db"`
	Services map[string]struct {
		Checks []struct {
			Name       string
			Interval   time.Duration
			Timeout    time.Duration
			Cmd        string
			Type       string
			MetricUnit string `yaml:"unit"`
		}
		Notifications notificationsRaw
	}
	Notifications notificationsRaw
}

func FromConfigFile(filePath string, historyOptions *history.NewOptions) (patrol *Patrol, err error) {
	buffer, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	return FromConfig(buffer, historyOptions)
}

func FromConfig(data []byte, historyOptions *history.NewOptions) (patrol *Patrol, err error) {
	raw := configRaw{}
	err = yaml.UnmarshalStrict(data, &raw)
	if err != nil {
		return
	}

	patrolOpts := CreatePatrolOptions{
		Name: "Statuspage",
		Port: 80,
	}

	if raw.Name == "" {
		patrolOpts.Name = raw.Name
	}

	if raw.DB == "" {
		err = fmt.Errorf("'db' propery must be specified in config file")
		return
	}

	if historyOptions == nil {
		patrolOpts.History.File = raw.DB
	} else {
		patrolOpts.History = *historyOptions
	}

	// Just a random guess for size, estimating about 5 checks for
	// each defined service
	patrolOpts.Checkers = make([]*checker.Checker, 0, len(raw.Services)*5)

	// Needs to be created here, so history file is opened
	patrol, err = New(patrolOpts)
	if err != nil {
		return
	}

	for group, groupConfig := range raw.Services {
		for idx, checkConfig := range groupConfig.Checks {
			c := checker.New(&checker.Checker{
				Group:      group,
				Name:       checkConfig.Name,
				Type:       checkConfig.Type,
				Cmd:        checkConfig.Cmd,
				MetricUnit: checkConfig.MetricUnit,
				Interval:   checkConfig.Interval,
				CmdTimeout: checkConfig.Timeout,
				History:    patrol.history,
			})
			if c.Type == "" {
				c.Type = "boolean"
			}
			if c.Name == "" {
				err = fmt.Errorf("%d-th check missing name in %s", idx, group)
				return
			}
			if c.Cmd == "" {
				err = fmt.Errorf("%d-th check missing cmd in %s", idx, group)
				return
			}
			if c.Type == "metric" && c.MetricUnit == "" {
				err = fmt.Errorf("%d-th check is of type metric but is missing unit in %s", idx, group)
				return
			}

			patrolOpts.Checkers = append(patrolOpts.Checkers, c)
		}
	}

	return
}
