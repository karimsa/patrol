<h1 align="center">
  <img src=".github/logo.png" alt="Patrol" />
</h1>

<p align="center">Host your own status pages.</p>

<p align="center">
  <a href="https://circleci.com/gh/karimsa/patrol">
    <img src="https://circleci.com/gh/karimsa/patrol.svg?style=svg" alt="CircleCI" />
  </a>
</p>

 - [Usage](#usage)
 - [Docker tags](#docker-tags)
 - [Creating a service](#creating-a-service)
 - [Creating health checks](#creating-health-checks)
	- [Health check images](#health-check-images)
	- [Health check options](#health-check-options)
 - [Badges](#badges)
 - [License](#license)

## Usage

The purpose of `patrol` is to be able to self-host an automated status page that gives you an overview of
your operations. The idea is different from something like Atlassian's Statuspage, since that is more for
communicating your operation status to external stakeholders while `patrol` is more for just monitoring.

To run `patrol` on your own, you simply need access to a machine with `docker` installed. To start, you should write your own configuration file, to something like [this](example.yml).

You can then run `patrol` via docker:

```shell
$ ls
patrol.yml
$ docker run \
        -d \
        --name patrol \
        --restart=on-failure \
        -v "$PWD:/config" \
        -v /var/run/docker.sock:/var/run/docker.sock:ro \
        -p 80:8080 \
        --log-driver json-file \
        --log-opt max-size=100m \
        karimsa/patrol:latest \
        	--config /config/hirefast.yml
```

This will start patrol on port `80` with the web interface. It will also give patrol access to your host machine's docker daemon so that it can spin up additional containers to run checks.

*Note: limiting the maximum log size for patrol is crucial, since patrol logs every time checks are run.*

## Docker tags

There are two tags that are published to the docker repo for this project:

 - `latest`: As per docker convention, this is the latest stable release of patrol.
 - `unstable`: This is the latest copy of the image from `master` - if you like to live life on the edge.

## Creating a service

Services in patrol are simply a collection of health checks. For now, they are mostly a visual grouping - checks belonging to the same service will be grouped together on the status page. To create a new service, you simply need to add a new key-value pair to the `services` key of the configuration.

For example, a simple service assigned to 'google.ca' could have the configuration:

```yaml
services:
  google.ca:
    - name: Delivers homepage
      cmd: 'curl -fsSL https://www.google.ca/'
```

## Creating health checks

Health checks are the core of patrol. Each health check is a simple shell script that tests the availability of a given feature in a service. If the script executes successfully, the health check is considered to be passed. If the script exits with a non-zero exit code, the health check is considered to be failed.

A simple health check might test an HTTP server's ability to deliver content by simplying executing a `curl` request. In the example above, `curl -fsSL https://www.google.ca/` is used to simply hit the google homepage at `www.google.ca` and will fail if the content is not delivered.

Since health checks can be any shell script, it is not necessary that you only use one command. For instance, if you are testing the availability of a specific page (let's say login) and your SPA might have a 404 but the HTTP server might not return a 404, you can use `grep` to verify that the right content was received instead of 'any content was received'.

For example:

```yaml
services:
  My App:
    - name: Delivers login
      cmd: 'curl -fsSL https://myapp.com/login | grep MyApp'
```

Since the errors are carried forward in the pipe (patrol enables the `-e` flag by default when executing your scripts), and `grep` fails when it cannot find the query, this check is complete.

As you can see, layering multiple checks might help you diagnose *where* the issue is when there is an issue. In this case, having both the `Delivers homepage` and `Delivers login` checks might tell you that if the first succeeds and the second fails, there is most likely a content delivery issue as opposed to an infrastructure issue.

**Note:** Since the exit code of the health check is used to determine whether the service is running or not, it is important that your command is setup to only fail if the service is failing. In the example of a `curl` request, you must specify the `-f, --fail` flag to ensure that curl exits with a non-zero exit code if the web server does not respond with a 2XX/3XX response.

### Health check images

Patrol runs each health check in its own docker container which are by default configured to be siblings of the patrol container itself. Though this might be useful for isolation, its main use case is to make it really easy to encapsulate your command into the right virtual environment.

For instance, let's say that you need to create a health check against a Redis deployment. You might be able to use general network tools like `ping` and `telnet` to test its uptime, but ideally, you should use the `redis-cli` to send commands.

To do this, you can simply specify the docker image that should be used when executing your health check. If at the time of running the health check the image is not available, patrol will pull the image from the docker registry.

Here's an example for redis:

```yaml
services:
  Redis:
    - name: Responds to PING
      image: redis:5
      cmd: 'redis-cli -h myredis.com -a myauth PING'
    - name: Job stream exists
      image: redis:5
      cmd: 'redis-cli -h myredis.com -a myauth XINFO STREAM mystream'
```

Currently, patrol does not support encrypting your sensitive information when pushing to git, so please be weary.

### Health check options

 - **name** (required): a string specifying the name to give this health check. If this name is changed, the entire history for the health check will be reset.
 - **cmd** (required; string/array):
	- If this is a string, it must be a command which can be passed to the shell via `/bin/sh -c 'cmd'`.
	- If this is an array, it must have all string elements and the contents will be concatenated with a ';' in between and then passed to the shell.
 - **image** (default: `karimsa/patrol`): this is the Docker image within which you want to run the command. It defaults to the patrol image which contains `curl`, `ping`, and `jq`.
 - **type** (optional; values: 'metric'): if specified as 'metric', the health check will considered to be of type metric and therefore its stdout will be parsed as a numeric value.
 - **unit** (optional): if specified and type is metric, will be used when displaying the metric chart on the status page.
 - **historySize** (default: 80): this is the maximum number of data points to persist. If the history of a service ever exceeds this amount, the oldest data points will be discarded from storage. This helps keep the storage of patrol low - especially since the underlying db driver is bad at handling large amounts of data.

## Badges

To embed live badges in your README or elsewhere, you can simply add `/badge` to the end of your patrol URL.
For instance, if your public URL is `https://status.myapp.com`, your badge will be available at: `https://status.myapp.com/badge`.

You can also pass a custom `style` query param which is passed directly to [shields.io](https://shields.io). For example, `https://status.myapp.com/badge?style=flat`.

## License

Licensed under MIT license.

Copyright &copy; 2019-present Karim Alibhai.

Badge in logo created by **Artdabana@Design** from the Noun Project.
