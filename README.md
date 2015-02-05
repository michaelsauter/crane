# Crane
Lift containers with ease

## Overview
Crane is a tool to orchestrate Docker containers. It works by reading in some configuration (JSON or YAML) which describes how to obtain images and how to run containers. This simplifies setting up a development environment a lot as you don't have to bring up every container manually, remembering all the arguments you need to pass. By storing the configuration next to the data and the app(s) in a repository, you can easily share the whole environment.

## Installation
The latest release can be installed via:

```
bash -c "`curl -sL https://raw.githubusercontent.com/michaelsauter/crane/master/download.sh`" && sudo mv crane /usr/local/bin/crane
```
You can also build Crane yourself by using the Go toolchain (`go get` and `go install`). Please have a look at the [release notes](https://github.com/michaelsauter/crane/releases) for the changelog if you're upgrading.

Of course, you will need to have Docker (>= 1.3) installed.

## Usage
Crane is a very light wrapper around the Docker CLI. This means that most commands just call the corresponding Docker command, but for all targeted containers. Additionally, there are a few special commands.

### `create`
Maps to `docker create`. Containers can be recreated by passing `--recreate`.

### `run`
Maps to `docker run`.  If a container already exists, it is just started. However, containers can be recreated by passing `--recreate`.

### `rm`
Maps to `docker rm`. Running containers can be killed first with `--kill`.

### `kill`
Maps to `docker kill`.

### `start`
Maps to `docker start`.

### `stop`
Maps to `docker stop`.

### `pause`
Maps to `docker pause`.

### `unpause`
Maps to `docker unpause`.

### `provision`
Either calls Docker's `build` or `pull`, depending on whether a Dockerfile is specified. The Docker cache can be disabled by passing `--no-cache`.

### `push`
Maps to `docker push`.

### `lift`
Will provision and run the containers in one go. By default, it does as little as possible to get the containers running. This means it only provisions images if necessary and just starts containers if they already exist. To update the images and recreate the containers, pass `--recreate` (and optionally `--no-cache`).

### `logs`
Maps to `docker logs`, multiplexing the logs of the targeted containers chronologically.

### `status`
Displays information about the state of the containers.

### `stats`
Maps to `docker stats`. Need Docker >= 1.5

### `graph`
Parses your config file and dumps the relations between containers as a dependency graph, using the DOT format. See built-in help for more information about style conventions used in that representation.

You can get more information about what's happening behind the scenes for all commands by using `--verbose`. All options have a short version as well, e.g. `lift -rn`.

## crane.json / crane.yaml
The configuration defines a map of containers in either JSON or YAML. By default, the configuration is expected in the current directory (`crane.json` or `crane.yaml`/`crane.yml`), but the location can also be specified via `--config`. Dependencies between containers are automatically detected and resolved.
The map of containers consists of the name of the container mapped to the container configuration, which consists of:

* `image` (string, required): Name of the image to build/pull
* `dockerfile` (string, optional): Relative path to the Dockerfile
* `run` (object, optional): Parameters mapped to Docker's `run` & `create`.
	* `add-host` (array) Add custom host-to-IP mappings.
	* `cap-add` (array) Add Linux capabilities.
	* `cap-drop` (array) Drop Linux capabilities.
	* `cidfile` (string)
	* `cpuset` (integer)
	* `cpu-shares` (integer)
	* `detach` (boolean) `sudo docker attach <container name>` will work as normal.
	* `device` (array) Add host devices.
	* `dns` (array)
	* `entrypoint` (string)
	* `env` (array)
	* `env-file` (array)
	* `expose` (array) Ports to expose to linked containers.
	* `hostname` (string)
	* `interactive` (boolean)
	* `link` (array) Link containers.
	* `lxc-conf` (array)
	* `mac-address` (string) Need Docker >= 1.4
	* `memory` (string)
	* `memory-swap` (string) Need Docker >= 1.5
	* `net` (string) The `container:id` syntax is not supported, use `container:name` if you want to reuse another container network stack.
	* `pid` (string) Need Docker >= 1.5
	* `privileged` (boolean)
	* `publish` (array) Map network ports to the container.
	* `publish-all` (boolean)
	* `read-only` (boolean) Need Docker >= 1.5
	* `restart` (string) Restart policy.
	* `rm` (boolean)
	* `tty` (boolean)
	* `user` (string)
	* `volume` (array) In contrast to plain Docker, the host path can be relative.
	* `volumes-from` (array) Mount volumes from other containers
	* `workdir` (string)
	* `cmd` (array/string) Command to append to `docker run` (overwriting `CMD`).
* `rm` (object, optional): Parameters mapped to Docker's `rm`.
	* `volumes` (boolean)
* `start` (object, optional): Parameters mapped to Docker's `start`.
	* `attach` (boolean)
	* `interactive` (boolean)

See the [Docker documentation](http://docs.docker.io/en/latest/reference/commandline/cli/#run) for more details about the parameters.

## Example
A typical `crane.yaml` looks like this:

```
containers:
	apache:
		image: some-apache-image:latest
		run:
			volumes-from: ["app"]
			publish: ["80:80"]
			link: ["mysql:db", "memcached:cache"]
			detach: true
	app:
		dockerfile: app
		image: michaelsauter/app
		run:
			volume: ["app/www:/srv/www:rw"]
			detach: true
	mysql:
		image: mysql
		run:
			env: ["MYSQL_ROOT_PASSWORD=mysecretpassword"]
			detach: true
	memcached:
		image: tutum/memcached
		run:
			detach: true
```
All specified Docker containers can then be created and started with:

```
crane lift
```

This will bring up the containers. The container running Apache has the MySQL and Memcached containers automatically linked.

If you want to use JSON instead of YAML, here's what a simple configuration looks like:

```
{
	"containers": {
		"pry": {
			"image": "d11wtq/ruby",
			"run": {
				"interactive": true,
				"tty": true,
				"cmd": "pry"
			}
		}
	}
}
```

## Advanced Usage

### Groups
Next to containers, you can also specify groups, and then execute Crane commands that only target those groups. If you do not specify any target as non-option arguments, the command will apply to all containers. However, you can override this by specifying a `default` group. Also, every container can be targeted individually by using the name of the container in the non-option arguments. Note that any number of group or container references can be used as target, and that ordering doesn't matter since containers will be ordered according to the dependency graph. Groups of containers can be specified like this (YAML shown):

```
containers:
  database1:
    ../..
  database2:
    ../..
  service1:
    ../..
  service2:
    ../..
groups:
  default: ["service1", "database1"]
  databases: ["database1", "database2"]
  services: ["service1", "service2"]

```

This could be used like so: `crane provision service1`, `crane run -v databases` or `crane lift -r services database1`. `crane status` is an alias for `crane status default`, which in that example is an alias for `crane status service1 database1`.


### Cascading
When using targets, it is also possible to cascade the commands to related containers. There are 2 different flags, `--cascade-affected` and `--cascade-dependencies`. In our example configuration above, when targeting the `mysql` container, the `apache` container would be considered to be "affected". When targeting the `apache` container, the `mysql` container would be considered as a "dependency". Both flags take a string argument, which specifies which type of cascading is desired, options are `all`, `required`, `link`, `volumesFrom` and `net`.

### Optional dependencies
All dependencies are considered required, and thus enforced, unless their declaration is prefixed with a "?". Apart from allowing to get more control for selecting containers implicitly via cascading, when declaring a `link` (respectively a `volumesFrom`) optional, the dependency will be silently omitted if the target is not running (respectively not existing). In the following example, `foo` has an optional link dependency to `bar`, meaning that that link will only be created if `bar` already exists when starting `foo`, or if `bar` is part of the targeted containers list when starting `foo`, explictly or implicitly (via `--cascade-dependencies=all` for instance):
```
containers:
  foo:
    image: busybox
    run:
      detach: true
      cmd: ["sleep", "50"]
      link: ["?bar:bar"]
  bar:
    image: busybox
    run:
      detach: true
      cmd: ["sleep", "50"]
```

## Some Crane-backed sample environments
* [Silex + Nginx/php-fpm + MySQL](https://github.com/michaelsauter/silex-crane-env)
* [Symfony2 + Apache + MySQL](https://github.com/michaelsauter/symfony2-crane-env)
* [Sinatra + PostrgeSQL](https://github.com/michaelsauter/sinatra-crane-env)

## Copyright & Licensing
Copyright © 2013-2014 Michael Sauter. See the LICENSE file for details.

---

[![GoDoc](https://godoc.org/github.com/michaelsauter/crane?status.png)](https://godoc.org/github.com/michaelsauter/crane)
