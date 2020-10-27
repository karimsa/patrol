package patrol

import (
	"fmt"

	"github.com/karimsa/patrol/internal/checker"
	"github.com/karimsa/patrol/internal/history"
	"github.com/karimsa/patrol/internal/logger"
)

type Patrol struct {
	port     int
	name     string
	history  *history.File
	checkers []*checker.Checker
}

type CreatePatrolOptions struct {
	Port     uint32
	Name     string
	History  history.NewOptions
	Checkers []*checker.Checker
	LogLevel logger.LogLevel
}

func New(options CreatePatrolOptions, historyFile *history.File) (*Patrol, error) {
	if historyFile == nil {
		var err error
		options.History.LogLevel = options.LogLevel
		historyFile, err = history.New(options.History)
		if err != nil {
			return nil, err
		}
	}

	p := &Patrol{
		port:     int(options.Port),
		name:     options.Name,
		history:  historyFile,
		checkers: options.Checkers,
	}
	p.SetLogLevel(options.LogLevel)
	return p, nil
}

func (p *Patrol) SetLogLevel(level logger.LogLevel) {
	p.history.SetLogLevel(level)
	for _, checker := range p.checkers {
		checker.SetLogLevel(level)
	}
}

func (p *Patrol) Start() {
	if p.checkers == nil || len(p.checkers) == 0 {
		panic(fmt.Errorf("Cannot start patrol with zero checkers"))
	}

	for _, checker := range p.checkers {
		go checker.Run()
	}
}

func (p *Patrol) Stop() {
	for _, checker := range p.checkers {
		checker.Close()
	}
}

func (p *Patrol) Close() {
	p.Stop()
	p.history.Close()
}
