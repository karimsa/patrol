package history

import (
	"fmt"
	"os"
	"sync"
	"testing"
	"time"
)

func TestDoubleOpen(t *testing.T) {
	os.Remove("./history-test.db")

	history, err := New(
		NewOptions{
			File:                "./history-test.db",
			MaxEntries:          90,
			MaxConcurrentWrites: 100,
		},
	)
	if err != nil {
		t.Error(err)
		return
	}

	wg := &sync.WaitGroup{}
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(i int, n time.Time) {
			defer wg.Done()
			if err := history.Append(Item{
				Group:     "staging",
				Name:      "Website is up",
				Type:      "metric",
				Output:    []byte(fmt.Sprintf("%d-th", i)),
				CreatedAt: n,
			}); err != nil {
				panic(err)
			}
		}(i, time.Now())
	}
	wg.Wait()

	items := history.GetGroupItems("staging")
	if len(items) != 10 {
		t.Error(fmt.Errorf("Failed to store/retrieve items:\n\nItems:\n\t%#v\n\nHistory:\n\t%#v\n", items, history))
		return
	}

	order := make([]string, 0, len(items))
	for _, item := range items {
		order = append(order, string(item.Output))
	}
	if fmt.Sprintf("%#v", order) != `[]string{"9-th", "8-th", "7-th", "6-th", "5-th", "4-th", "3-th", "2-th", "1-th", "0-th"}` {
		t.Error(fmt.Errorf("Incorrectly ordered results:\n\n%#v\n\n%#v\n", order, items))
		return
	}

	history.Close()

	history, err = New(
		NewOptions{
			File:                "./history-test.db",
			MaxEntries:          90,
			MaxConcurrentWrites: 100,
		},
	)
	if err != nil {
		t.Error(err)
		return
	}

	items = history.GetGroupItems("staging")
	if len(items) != 10 {
		t.Error(fmt.Errorf("Failed to store/retrieve items:\n\nItems:\n\t%#v\n\nHistory:\n\t%#v\n", items, history))
		return
	}

	order = make([]string, 0, len(items))
	for _, item := range items {
		order = append(order, string(item.Output))
	}
	if fmt.Sprintf("%#v", order) != `[]string{"9-th", "8-th", "7-th", "6-th", "5-th", "4-th", "3-th", "2-th", "1-th", "0-th"}` {
		t.Error(fmt.Errorf("Incorrectly ordered results:\n\nOrder:\n\t%#v\n\n\nItems:\n\t%#v\n\nHistory:\n\t%#v\n", order, items, history))
		return
	}

	history.Close()
}

func TestUpserts(t *testing.T) {
	os.Remove("./history-test-upsert.db")
	history, err := New(
		NewOptions{
			File:                "./history-test-upsert.db",
			MaxEntries:          90,
			MaxConcurrentWrites: 100,
		},
	)
	if err != nil {
		t.Error(err)
		return
	}

	if err := history.Append(Item{
		Group:     "staging",
		Name:      "Website is up",
		Type:      "boolean",
		Output:    []byte("1st"),
		CreatedAt: time.Now(),
	}); err != nil {
		t.Error(err)
		return
	}
	if err := history.Append(Item{
		Group:     "staging",
		Name:      "Website is up",
		Type:      "boolean",
		Output:    []byte("2nd"),
		CreatedAt: time.Now(),
	}); err != nil {
		t.Error(err)
		return
	}

	items := history.GetGroupItems("staging")
	if len(items) != 1 {
		t.Error(fmt.Errorf("Failed to store/retrieve items:\n\nItems:\n\t%#v\n\nHistory:\n\t%#v\n", items, history))
		return
	}
	if string(items[0].Output) != "2nd" {
		t.Error(fmt.Errorf("Wrong record stored in upsert:\n\nItems:\n\t%#v\n\nHistory:\n\t%#v\n", items, history))
		return
	}

	history.Close()
}
