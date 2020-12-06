package patrol

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/NYTimes/gziphandler"
	"github.com/karimsa/patrol/internal/checker"
	"github.com/karimsa/patrol/internal/history"
	"github.com/karimsa/patrol/internal/logger"
)

// Options used to setup patrol's HTTP server.
type PatrolHttpsOptions struct {
	// Paths to SSL certificate and key files - cannot be zero value.
	Cert, Key string

	// This port is used to run the HTTPS server. Zero value is invalid
	// for port.
	Port uint32
}

// Patrol instance to manage a set of checkers, a history file, and run
// a web server to serve the web interface. Currently, instances cannot
// be created directly. You must use: 'New', 'FromConfig', or 'FromConfigFile'.
type Patrol struct {
	History *history.File

	name                string
	port                int
	https               *PatrolHttpsOptions
	checkers            []*checker.Checker
	server              *http.Server
	logger              logger.Logger
	logLevel            logger.LogLevel
	groupEventHandlers  map[string]EventHandlers
	globalEventHandlers EventHandlers
}

// Map that goes from item status values to a list of notification objects
type EventHandlers map[string][]*singleNotificationConfig

// Options for creating a new patrol instance.
type CreatePatrolOptions struct {
	// Port at which to listen for HTTP requests. If HTTPS
	// options are specified, this port simply acts as an
	// HTTP to HTTPS redirect server.
	Port uint32

	// HTTPS options to listen on HTTPS as well as HTTP.
	// Zero value indicates no HTTPS server.
	HTTPS *PatrolHttpsOptions

	// Name is used to render the web interface. It is used
	// as the page's <title> and the heading at the top of
	// the page.
	Name string

	// History options are used to open and create a new history
	// file. If a history file is specified to the constructor, this
	// struct is ignored.
	History history.NewOptions

	// Set of checkers that should be managed by the patrol instance.
	// This slice cannot be nil, but it can be empty.
	Checkers []*checker.Checker

	// Minimum level of logs that should be printed. This value is forced
	// onto the 'history.File' and 'checker.Checker' objects that are
	// managed by this patrol instance.
	LogLevel logger.LogLevel

	// Event handlers by group
	GroupEventHandlers map[string]EventHandlers

	// Event handlers for all changes
	GlobalEventHandlers EventHandlers
}

func New(options CreatePatrolOptions, historyFile *history.File) (*Patrol, error) {
	if historyFile == nil {
		groups := make(map[string]map[string]bool, len(options.Checkers))
		for _, checker := range options.Checkers {
			if _, ok := groups[checker.Group]; !ok {
				groups[checker.Group] = make(map[string]bool, len(options.Checkers))
			}
			groups[checker.Group][checker.Name] = true
		}

		var err error
		options.History.LogLevel = options.LogLevel
		options.History.Groups = groups
		historyFile, err = history.New(options.History)
		if err != nil {
			return nil, err
		}
	}

	p := &Patrol{
		name:                options.Name,
		port:                int(options.Port),
		https:               options.HTTPS,
		checkers:            options.Checkers,
		server:              &http.Server{},
		logLevel:            options.LogLevel,
		logger:              logger.New(options.LogLevel, ""),
		groupEventHandlers:  options.GroupEventHandlers,
		globalEventHandlers: options.GlobalEventHandlers,

		History: historyFile,
	}
	p.server.Handler = gziphandler.GzipHandler(p)
	if p.name == "" {
		p.name = "Statuspage"
	}
	p.SetLogLevel(options.LogLevel)
	return p, nil
}

func (p *Patrol) String() string {
	hStr := strings.Split(p.History.String(), "\n")
	for i := 1; i < len(hStr); i++ {
		hStr[i] = "\t" + hStr[i]
	}

	return strings.Join([]string{
		fmt.Sprintf("Patrol{"),
		fmt.Sprintf("\tname: %s,", p.name),
		fmt.Sprintf("\tport: %d,", p.port),
		fmt.Sprintf("\thttps: %#v,", p.https),
		fmt.Sprintf("\tcheckers: %d checkers,", len(p.checkers)),
		fmt.Sprintf("\tlogLevel: %d,", p.logLevel),
		fmt.Sprintf("\tHistory: %s,", strings.Join(hStr, "\n")),
		fmt.Sprintf("}"),
	}, "\n")
}

func (p *Patrol) SetLogLevel(level logger.LogLevel) {
	p.logLevel = level
	p.logger = logger.New(level, "")
	p.History.SetLogLevel(level)
	for _, checker := range p.checkers {
		checker.SetLogLevel(level)
	}
}

func (p *Patrol) OnCheckerStatus(status, group, checker string) {
	p.logger.Debugf("status changed: %s, %s, %s", status, group, checker)

	if p.globalEventHandlers != nil {
		if handlers, ok := p.globalEventHandlers[status]; ok {
			p.logger.Debugf("Sending global notification for %s status of %s", status, group)
			for _, n := range handlers {
				n.Run()
			}
		}
	}
	if groupHandlers, ok := p.groupEventHandlers[group]; ok {
		if handlers, ok := groupHandlers[status]; ok {
			p.logger.Debugf("Sending global notification for %s status of %s", status, group)
			for _, n := range handlers {
				n.Run()
			}
		}
	}
}

func (p *Patrol) Start() {
	if p.checkers == nil || len(p.checkers) == 0 {
		panic(fmt.Errorf("Cannot start patrol with zero checkers"))
	}

	for _, checker := range p.checkers {
		checker.Start(p)
	}

	go func() {
		var err error
		if p.https == nil {
			p.server.Addr = fmt.Sprintf(":%d", p.port)
			err = p.server.ListenAndServe()
		} else {
			go func() {
				err := http.ListenAndServe(fmt.Sprintf(":%d", p.port), http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
					http.Redirect(
						res,
						req,
						fmt.Sprintf("https://%s:%d", strings.Split(req.Host, ":")[0], p.https.Port),
						http.StatusTemporaryRedirect,
					)
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
	p.logger.Infof("Waiting for graceful shutdown")
	p.Stop()
	p.History.Close()
}
