<p align="center">
  <img src=".github/preview.png" alt="Preview image" />
</p>

<p align="center">Host your own uptime monitoring status pages.</p>

<p align="center">
  <a href="https://circleci.com/gh/karimsa/patrol">
    <img src="https://circleci.com/gh/karimsa/patrol.svg?style=svg" alt="CircleCI" />
  </a>
</p>

 - [TLDR](#tldr)
	- [Installing natively](#installing-natively)
	- [Running with docker](#running-with-docker)
 - [Usage](#usage)
 - [Docker tags](#docker-tags)
 - [Creating a service](#creating-a-service)
 - [Creating health checks](#creating-health-checks)
	- [Health check images](#health-check-images)
	- [Health check options](#health-check-options)
 - [Managing secrets](#managing-secrets)
 - [License](#license)

## TL;DR

  1. Create a config file like [this one](https://github.com/karimsa/patrol/tree/master/example.yml).

### Installing natively

```shell
$ curl -sf https://gobinaries.com/karimsa/patrol/cmd/patrol | sh
```

### Running with docker

```shell
$ docker run \
		-d \
		--name patrol \
		--restart=on-failure \
		-v "$PWD:/data" \
		-p 80:8080 \
		--log-driver json-file \
		--log-opt max-size=100m \
		ghcr.io/karimsa/patrol:latest \
			run \
			--config /config/patrol.yml
```

## Usage

The purpose of `patrol` is to be able to self-host an automated status page that gives you an overview of
your operations. The idea is different from something like Atlassian's Statuspage, since that is more for
communicating your operation status to external stakeholders while `patrol` is more for just monitoring.

To run `patrol` on your own, you simply need access to a machine with `docker` installed. To start, you should write your own configuration file, to something like [this](example.yml).

You can then run `patrol` via docker:

```shell
$ ls
patrol.yml
$ patrol run --config /config/patrol.yml
```

This will start patrol on port `80` with the web interface. It will also give patrol access to your host machine's docker daemon so that it can spin up additional containers to run checks.

*Note: limiting the maximum log size for patrol is crucial, since patrol logs every time checks are run.*

## Docker tags

The docker image is hosted at `ghcr.io/karimsa/patrol`.

There are two tags that are published to the docker repo for this project:

 - `latest`: As per docker convention, this is the latest stable release of patrol.
 - `unstable`: This is the latest copy of the image from `master` - if you like to live life on the edge.

## Creating a service

Services in patrol are simply a collection of health checks. For now, they are mostly a visual grouping - checks belonging to the same service will be grouped together on the status page. To create a new service, you simply need to add a new key-value pair to the `services` key of the configuration.

For example, a simple service assigned to 'google.ca' could have the configuration:

```yaml
services:
  google.ca:
	checks:
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
	checks:
    - name: Delivers login
      cmd: 'curl -fsSL https://myapp.com/login | grep MyApp'
```

**Default shell options**:

 * Patrol first tries to use the shell specified by `$SHELL` - in the provided docker image, this env var is undefined.
 * If no shell is found, patrol tries to rely on `/bin/sh`. If `/bin/sh` is symlinked to `dash` (on ubuntu systems), it will rely on `/bin/bash`.
 * The shell is started with `-e` and `-o pipefail`. This means that any failing commands in your script will result in the health check failing, and any errors in piped operations resulting in failure of the entire piped command.

Since in the above example the errors are carried forward in the pipe, and `grep` fails when it cannot find the query, this check is complete.

As you can see, layering multiple checks might help you diagnose *where* the issue is when there is an issue. In this case, having both the `Delivers homepage` and `Delivers login` checks might tell you that if the first succeeds and the second fails, there is most likely a content delivery issue as opposed to an infrastructure issue.

**Note:** Since the exit code of the health check is used to determine whether the service is running or not, it is important that your command is setup to only fail if the service is failing. In the example of a `curl` request, you must specify the `-f, --fail` flag to ensure that curl exits with a non-zero exit code if the web server does not respond with a 2XX/3XX response.

### Health check options

 - **name** (required): a string specifying the name to give this health check. If this name is changed, the entire history for the health check will be reset.
 - **cmd** (required; string/array):
	- If this is a string, it must be a command which can be passed to the shell via `/bin/sh -c 'cmd'`.
	- If this is an array, it must have all string elements and the contents will be concatenated with a ';' in between and then passed to the shell.
 - **type** ('boolean' or 'metric', defaults to boolean): if specified as 'metric', the stdout of the check's command will be parsed as a numeric value.
 - **unit** (required if type is 'metric'): if type is metric, this will be used when displaying the metric chart on the status page.

## Managing Secrets

There are two ways to manage secrets for patrol config files.

### Using environment variables

The first is to store secure values inside of environment variables. Since patrol passes its own environment variables down to the child process, any environment variables that are passed to the patrol process (via docker or otherwise) are made available to the commands.

Example:

```shell
$ cat > patrol.yml << EOF
services:
	env test:
		checks:
		- name: testing environment vars
		  cmd: 'test "$SECRET" = "hello"'
EOF
$ patrol run --config patrol.yml
# the check will always fail here
$ SECRET=hello patrol run --config patrol.yml
# the check will always pass
```

### Encrypting your config file

You can use `openssl` or [`secrets`](https://github.com/karimsa/secrets) to encrypt specific keys or the entire config file. When deploying your statuspage, remember to decrypt the keys or file so that patrol can access the raw values.

## License

Licensed under MIT license.

Copyright &copy; 2019-present Karim Alibhai.

Badge in logo created by **Artdabana@Design** from the Noun Project.
