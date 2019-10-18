<h1 align="center">patrol</h1>
<p align="center">Host your own status pages.</p>

<p align="center">
  <a href="https://circleci.com/gh/karimsa/patrol">
    <img src="https://circleci.com/gh/karimsa/patrol.svg?style=svg" alt="CircleCI" />
  </a>
</p>

## Usage

The purpose of `patrol` is to be able to self-host an automated status page that gives you an overview of
your operations. The idea is different from something like Atlassian's Statuspage, since that is more for
communicating your operation status to external stakeholders while `patrol` is more for just monitoring.

To run `patrol` on your own, you simply need access to a machine with `docker` installed. To start, you
should write your own configuration file, to something of this effect:

```yaml
web:
  ## Sets the title at the top of the status page
  title: MyApp Status

## This is the directory, relative to the config file, where all data
## will be stored. This includes every historical check ever run.
dbDirectory: db

## Map consisting of services to display on your statuspage. Each service
## can have multiple checks. Each check runs in its own docker container, so
## you have the luxury of picking a docker image that already has the tools
## you need. If no image is given, it defaults to `byrnedo/alpine-curl` - which
## is an image built from alpine linux but with curl.
##
## All check commands are simply run using the default shell in the image (/bin/sh).
services:
  App:
    - name: API Status
      interval: 60s
      check: 'curl -fsSL https://app.myapp.com/api/v0/status'
    - name: Web delivers homepage
      interval: 60s
      check: 'curl -fsSL -o /dev/null https://app.myapp.ca/'
    - name: Web delivers login
      interval: 60s
      check: 'curl -fsSL -o /dev/null https://app.myapp.ca/login'
  Redis:
    - name: Responds to pings
      interval: 60s
      image: redis:5
      check: '! redis-cli -h redis.ca -n 0 -a pass ping | grep ERR'
  Mongo:
    - name: Users exist
      interval: 60s
      image: mongo:4.2
      check: 'test "`mongo "mongodb://user:pwd@mongo.com/myapp" --eval "db.users.estimatedDocumentCount()" | tail -n 1`" != "0"'

notifications:
  ## Array of notifications to emit when checks fail
  ## Currently, only a webhook is supported. All options are passed directly to `request` - but
  ## string bodies get interpolated with information about the check that triggered the notification.
  on_failure:
    - type: webhook
      options:
        method: post
        url: https://hooks.slack.com/services/MY_CUSTOM_WEBHOOK
        headers:
          'Content-Type': 'application/json'
        body: '{"text":"Service \"{{service}}\" is down (check \"{{check.name}}\" failed)."}'
  ## Array of notifications to emit when checks complete successfully occur
  ## Depending on your interval settings, this might be a hell of a lot.
  on_success:
    - type: webhook
      options:
        method: post
        url: https://hooks.slack.com/services/MY_CUSTOM_WEBHOOK
        headers:
          'Content-Type': 'application/json'
        body: '{"text":"Service \"{{service}}\" is up (check \"{{check.name}}\" completed)."}'
```

You can then run `patrol` via docker:

```shell
$ ls
patrol.yml
$ docker run \
	-v "$PWD:/config" \
	-v /var/run/docker.sock:/var/run/docker.sock \
	-p 80:8080 \
	karimsa/patrol:latest \
	--config /config/patrol.yml
```

This will start patrol on ports `80` and `8080`, where `80` will host the web interface and `8080` will host
the API server. It will also give patrol access to your host machine's docker daemon so that it can spin up
additional containers.

## License

Licensed under MIT license.

Copyright &copy; 2019-present Karim Alibhai.
