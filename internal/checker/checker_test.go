package checker

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/karimsa/patrol/internal/history"
	"github.com/karimsa/patrol/internal/logger"
)

func TestBooleanChecks(t *testing.T) {
	checker := New(&Checker{
		Group:    "staging",
		Name:     "Network is up",
		Type:     "boolean",
		Interval: 1 * time.Minute,
		Cmd:      "ping -c3 localhost",
	})

	item := checker.Check()
	if item.Group != "staging" || item.Name != "Network is up" || item.Type != "boolean" || item.Status != "healthy" {
		t.Error(fmt.Errorf("Unexpected result from check: %s", item))
		return
	}
}

func TestRunLoop(t *testing.T) {
	os.Remove("history-checker.db")
	historyFile, err := history.New(history.NewOptions{
		File: "history-checker.db",
	})
	if err != nil {
		t.Error(err)
		return
	}

	checker := New(&Checker{
		Group:    "staging",
		Name:     "Network is up",
		Type:     "boolean",
		Interval: 1 * time.Minute,
		Cmd:      "ping -c3 localhost",
		History:  historyFile,
	})
	checker.Start()

	var items []history.Item
	for i := 0; i < 10 && len(items) == 0; i++ {
		items = historyFile.GetItems(checker)
		time.Sleep(1 * time.Second)
	}
	if len(items) != 1 {
		t.Error(fmt.Errorf("Bad result for history: %#v", items))
		return
	}

	checker.Close()
	historyFile.Close()
}

func TestRetries(t *testing.T) {
	fd, err := ioutil.TempFile(os.TempDir(), "*")
	if err != nil {
		t.Error(err)
		return
	}
	fd.Close()

	historyFile, err := history.New(history.NewOptions{
		File: "history-retries.db",
	})
	if err != nil {
		t.Error(err)
		return
	}

	checker := New(&Checker{
		Group:         "file writer",
		Name:          "write hello",
		Type:          "boolean",
		Interval:      1 * time.Minute,
		Cmd:           fmt.Sprintf("echo hello world >> %s; exit 1", fd.Name()),
		RetryInterval: 1 * time.Nanosecond,
		History:       historyFile,
	})
	checker.SetLogLevel(logger.LevelDebug)

	var items []history.Item
	for i := 0; i < 10 && len(items) == 0; i++ {
		items = historyFile.GetItems(checker)
		time.Sleep(1 * time.Second)
	}
	if len(items) != 1 {
		t.Error(fmt.Errorf("Bad result for history: %#v", items))
		return
	}
	checker.Check()

	data, err := ioutil.ReadFile(fd.Name())
	if err != nil {
		t.Error(err)
		return
	}

	if lines := strings.Split(strings.TrimSpace(string(data)), "\n"); len(lines) != checker.MaxRetries {
		t.Error(fmt.Errorf("Check not retried enough times: %#v", lines))
		return
	}

	historyFile.Close()
}
