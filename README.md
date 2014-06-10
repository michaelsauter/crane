# Crane
Lift containers with ease

## Overview
Crane is a little tool to orchestrate Docker containers. It works by reading in some configuration (JSON or YAML) which describes how to obtain images and how to run the containers. This simplifies setting up a development environemt a lot as you don't have to bring up every container manually, remembering all the arguments you need to pass. By storing the configuration next to the data and the app(s) in a repository, you can easily share the whole environment.

## Installation
The latest release can be installed via:

```
bash -c "`curl -sL https://raw.githubusercontent.com/michaelsauter/crane/master/download.sh`" && sudo mv crane /usr/local/bin/crane
```
You can also build Crane yourself by using the Go toolchain (`go get` and `go install`).

Of course, you will need to have Docker (>= 0.8) installed on your system. I generally recommend to do this on Ubuntu, but if you are on OS X, you can also try [docker-osx](https://github.com/noplay/docker-osx). [boot2docker](https://github.com/boot2docker/boot2docker) is nice, but unfortunately, it does not support bind-mounting volumes yet.

## Usage
Crane is a very light wrapper around the Docker commands. This means that e.g. `run`, `rm`, `kill`, `start`, `stop` just call the corresponding Docker commands, but for all defined containers. Additionally, there are a few special commands:

* `provision` either calls Docker's `build` or `pull` (depending on whether a Dockerfile is specified)
* `lift` will build and run the containers in one go
* `status` will display information about the state of the containers

You can get more information about what's happening behind the scenes by using `--verbose`.
Some commands have a `--force` flag, which will save you intermediate steps, such as stopping the containers before removing them, or rebuilding images when they exist already. When you use `--force` to remove containers first, you can also use `--kill` if you're impatient.
For all available commands and details on usage, just type `crane`.

## crane.json / crane.yaml
The configuration defines an array of containers in either JSON or YAML. By default, the configuration is expected in the current directory (`crane.json` or `crane.yaml`/`crane.yml`), but it can also be specified via `--config`. If a container depends on another one, it must appear before that container in the configuration file.
Every container consists of:

* `name` (string, required): Name of the container
* `image` (string, required): Name of the image to build/pull
* `dockerfile` (string, optional): Relative path to the Dockerfile
* `run` (object, optional): Parameters mapped to Docker's `run`.
	* `cidfile` (string)
	* `cpu-shares` (integer)
	* `detach` (boolean) `sudo docker attach <container name>` will work as normal.
	* `dns` (array)
	* `env` (array)
	* `env-file` (string)
	* `expose` (array) Ports to expose to linked containers.
	* `hostname` (string)
	* `interactive` (boolean)
	* `link` (array) Link containers.
	* `lxc-conf` (array)
	* `memory` (string)
	* `privileged` (boolean)
	* `publish` (array) Map network ports to the container.
	* `publish-all` (boolean)
	* `rm` (boolean)
	* `tty` (boolean)
	* `user` (string)
	* `volume` (array) In contrast to plain Docker, the host path can be relative.
	* `volumes-from` (array) Mount volumes from other containers
	* `workdir` (string)
	* `cmd` (array/string) Command to append to `docker run` (overwriting `CMD`).

See the [Docker documentation](http://docs.docker.io/en/latest/reference/commandline/cli/#run) for more details about the parameters.

## Example
For demonstration purposes, we'll bring up a PHP app (served by Apache) that depends both on a MySQL database and a Memcached server. The source code is available at http://github.com/michaelsauter/crane-example. Here's what the `crane.json` looks like:

```
{
	"containers": [
		{
			"name": "crane_apache",
			"dockerfile": "apache",
			"image": "icrane_apache",
			"run": {
				"volumes-from": ["crane_app"],
				"publish": ["80:80"],
				"link": ["crane_mysql:db", "crane_memcached:cache"],
				"detach": true
			}
		},
		{
			"name": "crane_app",
			"dockerfile": "app",
			"image": "icrane_app",
			"run": {
				"volume": ["app/www:/srv/www:rw"],
				"detach": true
			}
		},
		{
			"name": "crane_mysql",
			"dockerfile": "mysql",
			"image": "icrane_mysql",
			"run": {
				"detach": true
			}
		},
		{
			"name": "crane_memcached",
			"dockerfile": "memcached",
			"image": "icrane_memcached",
			"run": {
				"detach": true
			}
		}
	]
}
```
If you have Docker installed, you can just clone that repository and bring up the environment right now.
In the folder where the `crane.json` is, type:

```
[sudo] crane lift
```

This will bring up the containers. The container running Apache has the MySQL and Memcached containers automatically linked. Open `http://localhost` and you should be greeted with "Hello World".

If you want to use YAML instead of JSON, here's what a simple configuration looks like:

```
containers:
	- name: pry
		image: d11wtq/ruby
		run:
			interactive: true
			tty: true
			cmd: pry

```

## Advanced Usage
Next to containers, you can also specify groups, and then execute Crane commands that only target those groups. If you do not specify `--target`, the command will apply to all containers. However, you can override the default by specifying a `default` group. Also, every container can be targeted by using the name of the container as an argument to `--target`. Groups of containers can be specifiec like this (YAML shown):

```
groups:
	databases: ["database1", "database2"]
	development: ["container1", "container2"]

```

This could be used like so: `crane provision --target="container1"` or `crane run --target="databases"`.

## Other Crane-backed environments
* [Silex + Nginx/php-fpm + MySQL](https://github.com/michaelsauter/silex-crane-env)
* [Symfony2 + Apache + MySQL](https://github.com/michaelsauter/symfony2-crane-env)
* [Sinatra + PostrgeSQL](https://github.com/michaelsauter/sinatra-crane-env)
