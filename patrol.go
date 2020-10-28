package patrol

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/karimsa/patrol/internal/checker"
	"github.com/karimsa/patrol/internal/history"
	"github.com/karimsa/patrol/internal/logger"
)

type Patrol struct {
	name     string
	port int
	https    *PatrolHttpsOptions
	history  *history.File
	checkers []*checker.Checker
	server   *http.Server
}

type PatrolHttpsOptions struct {
	Cert, Key string
	Port uint32
}

type CreatePatrolOptions struct {
	Port     uint32
	HTTPS    *PatrolHttpsOptions
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
		name:     options.Name,
		port: int(options.Port),
		https: options.HTTPS,
		history:  historyFile,
		checkers: options.Checkers,
		server: &http.Server{},
	}
	if p.name == "" {
		p.name = "Statuspage"
	}
	p.server.Handler = p
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

	go func() {
		var err error
		if p.https == nil {
			p.server.Addr = fmt.Sprintf(":%d", p.port)
			err = p.server.ListenAndServe()
		} else {
			go func() {
				err := http.ListenAndServe(fmt.Sprintf(":%d", p.port), http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
					headers := res.Header()
					headers["Location"] = []string{
						fmt.Sprintf("https://%s:%d", strings.Split(req.Host, ":")[0], p.https.Port),
					}
					res.WriteHeader(http.StatusTemporaryRedirect)
				}))
				if err != nil && err != http.ErrServerClosed {
					panic(err)
				}
			}()

			p.server.Addr = fmt.Sprintf(":%d", p.https.Port)
			err = p.server.ListenAndServeTLS(p.https.Cert, p.https.Key)
		}

		if err != nil && err != http.ErrServerClosed {
			panic(err)
		}
	}()
}

func (p *Patrol) Stop() {
	for _, checker := range p.checkers {
		checker.Close()
	}

	ctx, cancel := context.WithTimeout(context.TODO(), 5*time.Second)
	defer cancel()
	if err := p.server.Shutdown(ctx); err != nil {
		panic(err)
	}
}

func (p *Patrol) Close() {
	p.Stop()
	p.history.Close()
}
