package patrol

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/karimsa/patrol/internal/checker"
	"github.com/karimsa/patrol/internal/history"
)

func TestServer(t *testing.T) {
	os.Remove("server-test.db")
	historyFile, err := history.New(history.NewOptions{
		File: "server-test.db",
	})
	if err != nil {
		t.Error(err)
		return
	}

	p, err := New(CreatePatrolOptions{
		Port: 8081,
		Checkers: []*checker.Checker{
			checker.New(&checker.Checker{
				Group:    "foo",
				Name:     "bar",
				Cmd:      "ping -c1 localhost",
				History:  historyFile,
				Interval: 1 * time.Minute,
			}),
			checker.New(&checker.Checker{
				Group:    "foo",
				Name:     "unbar",
				Cmd:      "ping -c1 localhost",
				History:  historyFile,
				Interval: 1 * time.Minute,
			}),
		},
	}, historyFile)
	if err != nil {
		t.Error(err)
		return
	}

	p.Start()
	<-time.After(1 * time.Second)

	res, err := http.Get("http://localhost:8081")
	if err != nil {
		t.Error(err)
		return
	}
	if res.StatusCode != 200 {
		t.Error(fmt.Errorf("Server returned non-200 status: %#v", res))
		return
	}
	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Printf("%s\n", data)

	p.Close()
}
