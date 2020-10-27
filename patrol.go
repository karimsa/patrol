package patrol

import (
	"github.com/karimsa/patrol/internal/checker"
	"github.com/karimsa/patrol/internal/history"
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
}

func New(options CreatePatrolOptions) (*Patrol, error) {
	historyFile, err := history.New(options.History)
	if err != nil {
		return nil, err
	}

	return &Patrol{
		port:     int(options.Port),
		name:     options.Name,
		history:  historyFile,
		checkers: options.Checkers,
	}, nil
}

func (p *Patrol) Start() {
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
