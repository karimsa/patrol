package patrol

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"text/template"
	"time"
	_ "embed"

	"github.com/andanhm/go-prettytime"
	"github.com/wcharczuk/go-chart"
	"github.com/wcharczuk/go-chart/drawing"

	"github.com/karimsa/patrol/internal/history"
	"github.com/karimsa/patrol/internal/logger"
)

type chartResult struct {
	SVG           string
	Min, Max, Avg float64
	Error         string
}

//go:embed dist/index.html
var indexHTML string

//go:embed dist/styles.css
var stylesCSS string

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
	seriesStyle = chart.Style{
		Show:        true,
		FontColor:   drawing.ColorBlack,
		StrokeWidth: 3,
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
				if a > b {
					r := make([]int, a-b+1)
					for i := 0; a >= b; i++ {
						r[i] = a
						a--
					}
					return r
				}

				r := make([]int, b-a+1)
				for i := 0; a <= b; i++ {
					r[i] = a
					a++
				}
				return r
			},
			"since": prettytime.Format,
			"fmtNum": func(n float64) string {
				parts := strings.Split(fmt.Sprintf("%.2f", n), ".")
				for i := len(parts[0]) - 3; i > 0; i -= 3 {
					parts[0] = parts[0][0:i] + ", " + parts[0][i:]
				}
				return parts[0] + "." + parts[1]
			},
			"chart": func(items []history.Item) chartResult {
				if len(items) < 1 {
					return chartResult{Error: "Data pending"}
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
						Style:   seriesStyle,
					},
				}

				// go-chart cannot draw constant functions
				if res.Min == res.Max {
					c.YAxis.Range = &chart.ContinuousRange{
						Min: res.Min - 1,
						Max: res.Max + 1,
					}
				}
				if len(items) == 1 {
					c.XAxis.Range = &chart.ContinuousRange{
						Min: float64(items[0].CreatedAt.UnixNano() - int64(24*time.Hour)),
						Max: float64(items[0].CreatedAt.UnixNano() + int64(24*time.Hour)),
					}
				} else if len(items) == 0 {
					c.XAxis.Range = &chart.ContinuousRange{
						Min: 0,
						Max: 1,
					}
				}

				buffer := bytes.Buffer{}
				if err := c.Render(chart.SVG, &buffer); err != nil {
					res.Error = fmt.Sprintf("Failed to render graph: %s", err)
				} else {
					res.SVG = base64.StdEncoding.EncodeToString(buffer.Bytes())
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
		Debug           bool
	}{
		Name:            p.name,
		Groups:          p.History.GetData(),
		NumServicesDown: 0,
		NumServices:     0,
		LatestCreatedAt: time.Unix(0, 0),
		GroupFilter:     query.Get("group"),
		StatusFilter:    query.Get("status"),
		Debug:           p.logLevel == logger.LevelDebug,
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
		p.logger.Warnf("Failed to execute template: %s", err)
		res.WriteHeader(500)
		res.Write([]byte(err.Error()))
	}
}
