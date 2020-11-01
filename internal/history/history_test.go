package history

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/karimsa/patrol/internal/logger"
)

func TestMultiOpen(t *testing.T) {
	os.Remove("./history-test.db")

	// 1st open/create
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

	for i := 0; i < 100; i++ {
		if err := history.Append(Item{
			Group:     "staging",
			Name:      "Website is up",
			Type:      "metric",
			Output:    []byte(fmt.Sprintf("%d-th", i)),
			CreatedAt: time.Now(),
		}); err != nil {
			panic(err)
		}
		<-time.After(1 * time.Millisecond)
	}

	var runAsserts = func() {
		items := history.GetGroupItems("staging", "Website is up")
		if len(items) != 90 {
			itemsStr := ""
			for _, item := range items {
				itemsStr += fmt.Sprintf("\t-> %s\n", item)
			}
			t.Error(fmt.Errorf("Failed to store/retrieve items:\n\nItems: length = %d\n\t%s\n\nHistory:\n\t%#v\n", len(items), itemsStr, history))
			return
		}

		order := make([]string, 0, len(items))
		for _, item := range items {
			order = append(order, string(item.Output))
		}
		if fmt.Sprintf("%#v", order) != `[]string{"99-th", "98-th", "97-th", "96-th", "95-th", "94-th", "93-th", "92-th", "91-th", "90-th", "89-th", "88-th", "87-th", "86-th", "85-th", "84-th", "83-th", "82-th", "81-th", "80-th", "79-th", "78-th", "77-th", "76-th", "75-th", "74-th", "73-th", "72-th", "71-th", "70-th", "69-th", "68-th", "67-th", "66-th", "65-th", "64-th", "63-th", "62-th", "61-th", "60-th", "59-th", "58-th", "57-th", "56-th", "55-th", "54-th", "53-th", "52-th", "51-th", "50-th", "49-th", "48-th", "47-th", "46-th", "45-th", "44-th", "43-th", "42-th", "41-th", "40-th", "39-th", "38-th", "37-th", "36-th", "35-th", "34-th", "33-th", "32-th", "31-th", "30-th", "29-th", "28-th", "27-th", "26-th", "25-th", "24-th", "23-th", "22-th", "21-th", "20-th", "19-th", "18-th", "17-th", "16-th", "15-th", "14-th", "13-th", "12-th", "11-th", "10-th"}` {
			panic(fmt.Errorf("Incorrectly ordered results:\n\n%#v\n\n%#v\n", order, items))
		}

		history.Close()
	}

	// 1st open/create
	runAsserts()

	// 2nd open
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
	runAsserts()

	// 3rd open
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
	runAsserts()
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

	items := history.GetGroupItems("staging", "Website is up")
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

func TestAutoCompact(t *testing.T) {
	dbFile := "./history-test-autocompact.db"
	os.Remove(dbFile)
	history, err := New(
		NewOptions{
			File:       dbFile,
			MaxEntries: 10,
			LogLevel:   logger.LevelDebug,
			Compact: CompactOptions{
				// To verify that compaction doesn't happen unnecessarily
				MaxWrites: 100000,
			},
		},
	)
	if err != nil {
		t.Error(err)
		return
	}
	for i := 0; i < 100; i++ {
		if err := history.Append(Item{
			Group:     "staging",
			Name:      "Website is up",
			Type:      "boolean",
			Output:    []byte("1st"),
			CreatedAt: time.Unix(0, (time.Now().UnixNano())-int64(24*time.Hour)),
		}); err != nil {
			t.Error(err)
			return
		}
	}
	for i := 0; i < 100; i++ {
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
	}
	if data, err := ioutil.ReadFile(dbFile); err != nil {
		t.Error(err)
		return
	} else if lines := len(strings.Split(strings.TrimSpace(string(data)), "\n")); lines < 200 {
		t.Error(fmt.Errorf("Compaction happened too early: Only %d lines exist in the history file\n%s", lines, data))
		return
	}

	history.Compact()
	if data, err := ioutil.ReadFile(dbFile); err != nil {
		t.Error(err)
		return
	} else if lines := len(strings.Split(strings.TrimSpace(string(data)), "\n")); lines > 2 {
		t.Error(fmt.Errorf("Compaction failed: %d lines in the history file (should be 2)\n%s", lines, data))
		return
	}

	history.Close()

	// With compaction options
	history, err = New(
		NewOptions{
			File:       dbFile,
			MaxEntries: 10,
			Compact: CompactOptions{
				MaxWrites: 10,
			},
			LogLevel: logger.LevelDebug,
		},
	)
	if err != nil {
		t.Error(err)
		return
	}
	for i := 0; i < 100; i++ {
		if err := history.Append(Item{
			Group:     "staging",
			Name:      "Website is up",
			Type:      "boolean",
			Output:    []byte{},
			CreatedAt: time.Now(),
		}); err != nil {
			t.Error(err)
			return
		}
	}
	history.Close()
	if data, err := ioutil.ReadFile(dbFile); err != nil {
		t.Error(err)
		return
	} else if lines := len(strings.Split(strings.TrimSpace(string(data)), "\n")); lines > 2 {
		t.Error(fmt.Errorf("Compaction failed: %d lines in the history file (should be 2)\n%s", lines, data))
		return
	}
}
