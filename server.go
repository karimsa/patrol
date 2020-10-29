package patrol

import (
	"log"
	"net/http"
	"net/url"
	"text/template"
	"time"

	"github.com/andanhm/go-prettytime"
	"github.com/karimsa/patrol/internal/history"
)

//go:generate ./scripts/build-css.sh

var (
	pageView = template.Must(
		template.New("index").Funcs(template.FuncMap{
			"mul": func(a, b int) int {
				return a * b
			},
			"sub": func(a, b int) int {
				return a - b
			},
			"plus": func(a, b int) int {
				return a + b
			},
			"nums": func(a, b int) []int {
				r := make([]int, b-a)
				for i := a; i < b; i++ {
					r[i-a] = i
				}
				return r
			},
			"since": prettytime.Format,
		}).Parse(indexHTML),
	)
)

func init() {
	template.Must(pageView.New("styles.css").Parse(stylesCSS))
}

func (p *Patrol) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	query, err := url.ParseQuery(req.URL.RawQuery)
	if err != nil {
		log.Printf("warn: Query parsing failed: %s", err)
	}

	data := struct {
		Name            string
		Groups          map[string]map[string][]history.Item
		NumServicesDown int
		LatestCreatedAt time.Time
		GroupFilter     string
		StatusFilter    string
	}{
		Name:            p.name,
		Groups:          p.history.GetData(),
		NumServicesDown: 0,
		LatestCreatedAt: time.Unix(0, 0),
		GroupFilter:     query.Get("group"),
		StatusFilter:    query.Get("status"),
	}

	for _, group := range data.Groups {
		for _, items := range group {
			if items[0].Status == "unhealthy" {
				data.NumServicesDown++
			}
			if data.LatestCreatedAt.Before(items[0].CreatedAt) {
				data.LatestCreatedAt = items[0].CreatedAt
			}
		}
	}

	if err := pageView.Execute(res, data); err != nil {
		res.WriteHeader(500)
		res.Write([]byte(err.Error()))
	}
}
