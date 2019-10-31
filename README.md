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

To run `patrol` on your own, you simply need access to a machine with `docker` installed. To start, you should write your own configuration file, to something like [this](example.yml).

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

## Badges

To embed live badges in your README or elsewhere, you can simply add `/badge` to the end of your patrol URL.
For instance, if your public URL is `https://status.myapp.com`, your badge will be available at: `https://status.myapp.com/badge`.

You can also pass a custom `style` query param which is passed directly to [shields.io](https://shields.io). For example, `https://status.myapp.com/badge?style=flat`.

## License

Licensed under MIT license.

Copyright &copy; 2019-present Karim Alibhai.
