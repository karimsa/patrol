package patrol

import (
	"os"
	"testing"
)

const configStr = `
db: config-test.db
name: MyApp Status
port: 80
services:
  API:
    checks:
      - name: API Status
        interval: 60s
        cmd: 'curl -fsSL https://app.myapp.com/api/v0/status'
    notifications:
      on_failure:
        - type: webhook
          options:
            method: delete
            url: https://api.heroku.com/apps/MY_HEROKU_APP/dynos
            headers:
              Authorization: 'Bearer heroku-token'
              Accept: 'application/vnd.heroku+json; version=3'
  Web:
    checks:
    - name: Web delivers homepage
      interval: 60s
      cmd: 'curl -fsSL -o /dev/null https://app.myapp.ca/'
    - name: Web delivers login
      interval: 60s
      cmd: 'curl -fsSL -o /dev/null https://app.myapp.ca/login'
    - name: Homepage latency
      type: metric
      unit: ms
      interval: 60s
      cmd: 'curl -fsSL -w "%{time_total}" -o /dev/null https://google.ca'
  Redis:
    checks:
    - name: Responds to pings
      interval: 60s
      cmd: '! redis-cli -h redis.ca -n 0 -a pass ping | grep ERR'
  Mongo:
    checks:
    - name: Users exist
      interval: 60s
      cmd: 'echo doing stuff'
notifications:
  on_failure:
    - type: webhook
      options:
        method: post
        url: https://hooks.slack.com/services/MY_CUSTOM_WEBHOOK
        headers:
          'Content-Type': 'application/json'
        body: '{"text":"Service \"{{service}}\" is down (check \"{{check.name}}\" failed)."}'
  on_success:
    - type: webhook
      options:
        method: post
        url: https://hooks.slack.com/services/MY_CUSTOM_WEBHOOK
        headers:
          'Content-Type': 'application/json'
        body: '{"text":"Service \"{{service}}\" is up (check \"{{check.name}}\" completed)."}'
`

func TestConfigValidate(t *testing.T) {
	os.Remove("config-test.db")
	if _, _, err := FromConfig([]byte(configStr), nil); err != nil {
		t.Error(err)
		return
	}
}
