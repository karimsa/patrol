package checker

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strconv"
	"sync"
	"time"

	"github.com/karimsa/patrol/internal/history"
	"github.com/karimsa/patrol/internal/logger"
)

var (
	SHELL = os.Getenv("SHELL")
)

func init() {
	log.Printf("Initializing with SHELL = %s", SHELL)
}

type Checker struct {
	Group      string
	Name       string
	Type       string
	Cmd        string
	MetricUnit string
	Interval   time.Duration
	CmdTimeout time.Duration
	History    *history.File

	logger   logger.Logger
	doneChan chan bool
	wg       *sync.WaitGroup
}

func New(c *Checker) *Checker {
	if c.CmdTimeout.Milliseconds() == 0 {
		c.CmdTimeout = 1 * time.Minute
	}
	c.doneChan = make(chan bool, 1)
	c.wg = &sync.WaitGroup{}
	c.SetLogLevel(logger.LevelInfo)
	c.History.AddChecker(c)
	return c
}

func (c *Checker) GetGroup() string {
	return c.Group
}

func (c *Checker) GetName() string {
	return c.Name
}

func (c *Checker) SetLogLevel(level logger.LogLevel) {
	c.logger = logger.New(
		level,
		fmt.Sprintf("%s:%s:", c.Group, c.Name),
	)
}

func (c *Checker) Check() history.Item {
	c.logger.Debugf("Checking status")

	stdout := bytes.Buffer{}
	stderr := bytes.Buffer{}
	combinedOutput := bytes.Buffer{}

	ctx, cancel := context.WithTimeout(
		context.TODO(),
		c.CmdTimeout,
	)
	cmd := exec.CommandContext(
		ctx,
		SHELL,
		"-o",
		"pipefail",
		"-ec",
		c.Cmd,
	)
	cmd.Stdin = os.Stdin
	cmd.Stdout = io.MultiWriter(&stdout, &combinedOutput)
	cmd.Stderr = io.MultiWriter(&stderr, &combinedOutput)

	cmdStart := time.Now()
	err := cmd.Run()
	cancel()

	item := history.Item{
		Group:      c.Group,
		Name:       c.Name,
		Type:       c.Type,
		Output:     combinedOutput.Bytes(),
		CreatedAt:  time.Now(),
		Duration:   time.Since(cmdStart),
		Metric:     0,
		MetricUnit: c.MetricUnit,
		Status:     "",
		Error:      "",
	}

	if exitErr, ok := err.(*exec.ExitError); err != nil && ok {
		item.Status = "unhealthy"
		item.Error = fmt.Sprintf("Process exited with status %d", exitErr.ExitCode())
	} else if err != nil {
		item.Status = "unhealthy"
		item.Error = fmt.Sprintf("Failed to run: #%v", err)
	} else {
		item.Status = "healthy"

		if c.Type == "metric" {
			n, err := strconv.ParseInt(string(stdout.Bytes()), 10, 64)
			if err == nil {
				item.Metric = n
			} else {
				item.Error = fmt.Sprintf("Failed to parse metric from output: %s", err)
			}
		}
	}

	c.logger.Infof("Check completed: %s", item)
	return item
}

func (c *Checker) Run() {
	c.wg.Add(1)
	defer c.wg.Done()

	for {
		item := c.Check()
		if err := c.History.Append(item); err != nil {
			panic(err)
		}

		c.logger.Infof("Waiting %s before checking again", c.Interval)
		select {
		case <-time.After(c.Interval):
		case <-c.doneChan:
			return
		}
	}
}

func (c *Checker) Close() {
	c.doneChan <- true
	c.wg.Wait()
}
