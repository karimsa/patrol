package patrol

import (
	"net/http"
	"text/template"

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
						function render() { Turbolinks.visit(location.href, { action: 'replace' }) }
						setInterval(render, 5 * 1000)
						window.addEventListener('focus', render)
					</script>
				</head>
				<body class="bg-gray-300">
					<header class="bg-gray-800 py-5">
						<div class="container lg:px-20 mx-auto">
							<h1 class="text-2xl font-bold text-white mb-4">MyApp Status</h1>
							<div class="bg-green-600 shadow-sm p-5 rounded mb-4 flex items-center justify-between">
								<p class="font-bold text-xl text-white">All Systems Operational</p>
								<span class="text-white text-sm">Last updated: a minute ago</span>
							</div>
						</div>
					</header>

					<main class="container mx-auto lg:px-20 py-5">
						{{range $groupName, $group := .Groups}}
							<div class="mb-12">
								<h4 class="font-bold mb-4 text-2xl">{{$groupName}}</h4>
								{{range $checkName, $items := $group}}
									<div class="bg-white shadow-sm p-5 rounded">
										<div class="mb-3 flex items-center justify-between">
											<p class="font-bold">{{$checkName}}</p>
											<div class="flex items-center">
												<span class="font-bold text-green-600">Healthy</span>
												<span class="text-gray-500 text-xs ml-4">a minute ago</span>
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
														fill="#38a169"
													/>
												{{end}}
											</svg>
										</div>
									</div>
								{{end}}
							</div>
						{{end}}
					</main>
				</body>
			</html>
		`),
	)
)

func (p *Patrol) ServeHTTP(res http.ResponseWriter, _ *http.Request) {
	data := struct {
		Name   string
		Groups map[string]map[string][]history.Item
	}{
		Name:   p.name,
		Groups: p.history.GetData(),
	}
	if err := pageView.Execute(res, data); err != nil {
		res.WriteHeader(500)
		res.Write([]byte(err.Error()))
	}
}
