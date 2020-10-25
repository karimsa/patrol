package checker

import (
	"fmt"
	"testing"
	"time"
)

func TestBooleanChecks(t *testing.T) {
	checker := New(&Checker{
		Group:    "staging",
		Name:     "Network is up",
		Type:     "boolean",
		Interval: 1 * time.Minute,
		Cmd:      "ping -c1 8.8.8.8",
	})

	item := checker.Check()
	if item.Group != "staging" || item.Name != "Network is up" || item.Type != "boolean" || item.Status != "healthy" {
		t.Error(fmt.Errorf("Unexpected result from check: %s", item))
		return
	}
}
