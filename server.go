package patrol

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"text/template"
	"time"

	"github.com/andanhm/go-prettytime"
	"github.com/karimsa/patrol/internal/history"
	"github.com/wcharczuk/go-chart"
	"github.com/wcharczuk/go-chart/drawing"
)

//go:generate ./scripts/build-css.sh

type chartResult struct {
	SVG           string
	Min, Max, Avg float64
}

var (
	gridMajorStyle = chart.Style{
		Show:        true,
		StrokeWidth: 1.5,
		StrokeColor: drawing.Color{
			A: 90,
		},
	}
	gridMinorStyle = chart.Style{
		Show:        true,
		StrokeWidth: 1.0,
		StrokeColor: drawing.Color{
			A: 50,
		},
	}
	metricChartDefaults = chart.Chart{
		XAxis: chart.XAxis{
			Style:          chart.StyleShow(),
			GridMajorStyle: gridMajorStyle,
			GridMinorStyle: gridMinorStyle,
		},
		YAxis: chart.YAxis{
			Style:          chart.StyleShow(),
			GridMajorStyle: gridMajorStyle,
			GridMinorStyle: gridMinorStyle,
		},
	}
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
			"chart": func(items []history.Item) chartResult {
				if len(items) < 2 {
					return chartResult{SVG: "Data pending"}
				}

				res := chartResult{
					Min: items[0].Metric,
					Max: items[0].Metric,
					Avg: 0,
				}
				xValues := make([]time.Time, len(items))
				yValues := make([]float64, len(items))
				for i, item := range items {
					xValues[i] = item.CreatedAt
					yValues[i] = item.Metric

					if res.Min > item.Metric {
						res.Min = item.Metric
					}
					if res.Max < item.Metric {
						res.Max = item.Metric
					}
					res.Avg += item.Metric
				}
				res.Avg /= float64(len(items))

				c := metricChartDefaults
				c.Series = []chart.Series{
					chart.TimeSeries{
						XValues: xValues,
						YValues: yValues,
						Style: chart.Style{
							Show:      true,
							FontColor: drawing.ColorBlack,
						},
					},
				}

				// go-chart cannot draw constant functions
				if res.Min == res.Max {
					c.YAxis.Range = &chart.ContinuousRange{
						Min: res.Min,
						Max: res.Max * 1.001,
					}
				}

				buffer := bytes.Buffer{}
				if err := c.Render(chart.SVG, &buffer); err != nil {
					res.SVG = fmt.Sprintf("Failed to render graph: %s", err)
				} else {
					res.SVG = string(buffer.Bytes())
				}
				return res
			},
		}).Parse(indexHTML),
	)
)

func init() {
	template.Must(pageView.New("styles.css").Parse(stylesCSS))
}

func (p *Patrol) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	defer func() {
		if err := recover(); err != nil {
			log.Printf("Error while serving http: %s", err)
		}
	}()

	query, err := url.ParseQuery(req.URL.RawQuery)
	if err != nil {
		log.Printf("warn: Query parsing failed: %s", err)
	}

	data := struct {
		Name            string
		Groups          map[string]map[string][]history.Item
		NumServicesDown int
		NumServices     int
		LatestCreatedAt time.Time
		GroupFilter     string
		StatusFilter    string
	}{
		Name:            p.name,
		Groups:          p.history.GetData(),
		NumServicesDown: 0,
		NumServices:     0,
		LatestCreatedAt: time.Unix(0, 0),
		GroupFilter:     query.Get("group"),
		StatusFilter:    query.Get("status"),
	}

	for _, group := range data.Groups {
		for _, items := range group {
			if len(items) > 0 {
				if items[0].Status == "unhealthy" {
					data.NumServicesDown++
				}
				if data.LatestCreatedAt.Before(items[0].CreatedAt) {
					data.LatestCreatedAt = items[0].CreatedAt
				}
				data.NumServices++
			}
		}
	}

	if err := pageView.Execute(res, data); err != nil {
		res.WriteHeader(500)
		res.Write([]byte(err.Error()))
	}
}
