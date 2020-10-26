package checker

import (
	"sync"
	"github.com/karimsa/patrol/internal/history"
)

type Suite struct {
	History *history.File
	Checkers []*Checker
}

func (s Suite) Start() {
	for _, checker := range s.Checkers {
		go checker.Run()
	}
}

func (s Suite) Stop() {
	for _, checker := range s.Checkers {
		checker.Close()
	}
}
