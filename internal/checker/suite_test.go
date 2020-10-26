package checker

import (
	"testing"
	"os"
	"github.com/karimsa/patrol/internal/history"
	"time"
	"fmt"
)

func TestSuite(t *testing.T) {
	os.Remove("history-checker.db")
	historyFile, err := history.New(history.NewOptions{
		File: "history-checker.db",
	})
	if err != nil {
		t.Error(err)
		return
	}

	suite := Suite{
		History: historyFile,
		Checkers: []*Checker{
			New(&Checker{
				Group:    "staging",
				Name:     "Network is up",
				Type:     "boolean",
				Interval: 1 * time.Minute,
				Cmd:      "ping -c1 8.8.8.8",
				History:  historyFile,
			}),
			New(&Checker{
				Group:    "staging",
				Name:     "Web is up",
				Type:     "boolean",
				Interval: 1 * time.Minute,
				Cmd:      "curl -fsSL https://staging.hirefast.ca",
				History:  historyFile,
			}),
		},
	}
	suite.Start()
	<-time.After(3 * time.Second)
	suite.Stop()

	items := historyFile.GetGroupItems("staging")
	if len(items) != 2 {
		t.Error(fmt.Errorf("Bad result for history: %#v", items))
		return
	}

	historyFile.Close()
}
