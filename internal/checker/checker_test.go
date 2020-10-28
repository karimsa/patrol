package checker

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/karimsa/patrol/internal/history"
)

func TestBooleanChecks(t *testing.T) {
	checker := New(&Checker{
		Group:    "staging",
		Name:     "Network is up",
		Type:     "boolean",
		Interval: 1 * time.Minute,
		Cmd:      "ping -c3 8.8.8.8",
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
		Cmd:      "ping -c3 8.8.8.8",
		History:  historyFile,
	})
	go checker.Run()
	<-time.After(1 * time.Second)
	checker.Close()

	items := historyFile.GetGroupItems("staging")
	if len(items) != 1 {
		t.Error(fmt.Errorf("Bad result for history: %#v", items))
		return
	}

	historyFile.Close()
}
