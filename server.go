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
		}).Parse(`
			<!doctype html>
			<html lang="en-US">
				<head>
					<meta charset="UTF-8">
					<title>{{.Name}}</title>
					<meta name="viewport" content="width=device-width, initial-scale=1.0">
					<link href="https://unpkg.com/tailwindcss@1.9.6/dist/tailwind.min.css" rel="stylesheet">
					<script src="https://cdnjs.cloudflare.com/ajax/libs/turbolinks/5.2.0/turbolinks.js"></script>
					<script>
						Turbolinks.Visit.prototype.performScroll = function(){}
						Turbolinks.BrowserAdapter.prototype.reload = function(){}
						function render() { Turbolinks.visit(location.href, { action: 'replace' }) }
						setInterval(render, 5 * 1000)
						window.addEventListener('focus', render)
					</script>
				</head>
				<body class="bg-gray-300">
					<header class="bg-gray-800 py-12">
						<div class="container lg:px-20 mx-auto">
							<h1 class="text-2xl font-bold text-white mb-4">MyApp Status</h1>
							<div class="{{if (eq .NumServicesDown 0)}}bg-green-700{{else}}bg-red-800{{end}} shadow-sm p-5 rounded mb-4 flex items-center justify-between">
								{{if (eq .NumServicesDown 0)}}
									<p class="font-semibold text-xl text-white">All systems operational</p>
								{{else}}
									<p class="font-semibold text-xl text-white">{{.NumServicesDown}} Systems are down</p>
								{{end}}
								<span class="text-white text-sm">Last updated: {{since .LatestCreatedAt}}</span>
							</div>
						</div>
					</header>

					<main class="container mx-auto lg:px-20 py-12">
						{{$data := .}}
						{{range $groupName, $group := $data.Groups}}
							{{if eq $groupName (or $data.GroupFilter $groupName)}}
							<div class="mb-12">
									<div class="mb-4 flex items-center">
										<h2 class="font-bold text-2xl inline-block">{{$groupName}}</h2>
										{{if eq $data.GroupFilter ""}}
											<a href="/?group={{$groupName}}" class="bg-blue-600 px-2 py-1 rounded text-white shadow-sm text-sm ml-4">Focus</a>
										{{else}}
											<a href="/" class="bg-indigo-600 px-2 py-1 rounded text-white shadow-sm text-sm ml-4">Unfocus</a>
										{{end}}
									</div>
								{{range $checkName, $items := $group}}
										{{$latestItem := index $items 0}}
									<div class="bg-white shadow-sm p-5 rounded mb-12">
										<div class="mb-4 flex items-center justify-between">
											<h3 class="font-semibold">{{$checkName}}</h3>
											<div class="flex items-center">
													{{if eq $latestItem.Status "healthy"}}
													<span class="font-semibold text-green-700">Healthy</span>
													{{else if eq $latestItem.Status "unhealthy"}}
													<span class="font-semibold text-red-800">Unhealthy</span>
												{{else}}
													<span class="font-semibold text-orange-600">Recovered</span>
												{{end}}

													<span class="text-gray-700 text-xs ml-4">{{ since $latestItem.CreatedAt }}</span>
											</div>
										</div>

										<div>
											<svg class="mx-auto" viewBox="0 0 318 10">
												{{range $_, $idx := nums 0 (sub 80 (len $items))}}
													<rect
														height="10"
														width="2"
														x="{{ mul $idx 4 }}"
														y="0"
														fill="#d9dbde"
													/>
												{{end}}
												{{range $idx, $item := $items}}
													<rect
														height="10"
														width="2"
														x="{{ mul (plus $idx 79) 4 }}"
														y="0"
														fill="{{if eq $item.Status "healthy"}}#38a169{{else}}#9b2c2c{{end}}"
													/>
												{{end}}
											</svg>
										</div>
									</div>
								{{end}}
							</div>
						{{end}}
						{{end}}
					</main>
				</body>
			</html>
		`),
	)
)

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
	}{
		Name:            p.name,
		Groups:          p.history.GetData(),
		NumServicesDown: 0,
		LatestCreatedAt: time.Unix(0, 0),
		GroupFilter:     query.Get("group"),
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
