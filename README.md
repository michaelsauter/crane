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

| Command | Maps to | Explanation and Options |
| ------------- | ----------- | ---------|
| create      | create           | Creates containers if they don't exist yet. Use `--recreate` to remove containers first. |
| run         | run              | Runs containers if they don't exist yet, otherwise just starts them. Use `--recreate` to remove containers first. |
| rm          | rm               | Removes containers if they exist. Use `--kill` to kill containers first. |
| start       | start            | Starts containers unless they are running already. |
| stop        | stop             | Stops containers if they are running. |
| kill        | kill             | Kills containers if they are running. |
| pause       | pause            |  |
| unpause     | unpause          |  |
| logs        | logs             | Logs of containers are multiplexed. Use `--follow` to follow log output. |
| stats       | stats            | Needs Docker >= 1.5 |
| provision   | pull/build       | Calls Docker's `pull` if no Dockerfile is specified. Otherwise it builds the image, optionally with disabled cache by passing `--no-cache`. |
| pull        | pull             | |
| push        | push             | |
| lift        | pull/build + run | Provisions and runs containers in one go. Use `--recreate` to remove containers first, `--no-cache` to disable build cache. |
| status      | -                | Displays information similar to `docker ps` for the given target. |
| graph       | -                | Dumps the relations between containers as a dependency graph, using the DOT format. See built-in help for more information about style conventions used in that representation. |

You can get more information about what's happening behind the scenes for all commands by using `--verbose`. All options have a short version as well, e.g. `lift -rn`.

## crane.json / crane.yaml
The configuration defines a map of containers in either JSON or YAML. By default, the configuration is expected in a file named `crane.json` or `crane.yaml`/`crane.yml`, or a file given via `--config`. Those files are searched for in the current directory, then recursively in the parent directory. Dependencies between containers are automatically detected and resolved.
The map of containers consists of the name of the container mapped to the container configuration, which consists of:

* `image` (string, required): Name of the image to build/pull
* `dockerfile` (string, optional): Relative path to the Dockerfile
* `run` (object, optional): Parameters mapped to Docker's `run` & `create`.
	* `add-host` (array) Add custom host-to-IP mappings.
	* `cap-add` (array) Add Linux capabilities.
	* `cap-drop` (array) Drop Linux capabilities.
	* `cgroup-parent` (string) Need Docker >= 1.6
	* `cidfile` (string)
	* `cpuset` (integer)
	* `cpu-shares` (integer)
	* `detach` (boolean) `sudo docker attach <container name>` will work as normal.
	* `device` (array) Add host devices.
	* `dns` (array)
	* `entrypoint` (string)
	* `env` (array/mapping) Can be declared as a string array with `"key[=value]"` items or a string-to-string mapping where each `key: value` will be translated to the corresponding `"key=value"` string.
	* `env-file` (array)
	* `expose` (array) Ports to expose to linked containers.
	* `hostname` (string)
	* `interactive` (boolean)
	* `label` (array/mapping) Can be declared as a string array with `"key[=value]"` items or a string-to-string mapping where each `key: value` will be translated to the corresponding `"key=value"` string.. Need Docker >= 1.6
	* `label-file` (array) Need Docker >= 1.6
	* `link` (array) Link containers.
	* `log-driver` (string) Need Docker >= 1.6
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
	* `ulimit` (array) Need Docker >= 1.6
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

Note that basic environment variable expansion (`${FOO}`, `$FOO`) is supported throughout the configuration, but advanced shell features such as command substitution (`$(cat foo)`, `` `cat foo` ``) or advanced expansions (`sp{el,il,al}l`, `foo*`, `~/project`, `$((A * B))`, `${PARAMETER#PATTERN}`) are *not* as the Docker CLI is called directly.

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
Note you can also declare the `env` parameter as:

```
env:
  MYSQL_ROOT_PASSWORD: mysecretpassword
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

### Cascading commands

When using targets, it is also possible to cascade the commands to related containers. There are 2 different flags, `--cascade-affected` and `--cascade-dependencies`. In our example configuration above, when targeting the `mysql` container, the `apache` container would be considered to be "affected". When targeting the `apache` container, the `mysql` container would be considered as a "dependency". Both flags take a string argument, which specifies which type of cascading is desired, options are `volumesFrom`, `link`, `net` and `all`.

### Hooks

In order to run certain commands before or after key lifecycle events of containers, hooks can declared in the configuration. They are run synchronously on the host where Crane is installed, outside containers, via an `exec` call. They may interrupt the flow by returning a non-zero status. If shell features more advanced than basic variable expansion is required, you should explicitly spawn a shell to run the command in (`sh -c 'ls *'`).

Hooks are declared at the top level of the configuration, under the `hooks` key. See YAML example below:

```
containers:
  service1:
    image: busybox
    run:
      detach: true
      cmd: ["sleep", "50"]
  service2:
    image: busybox
    run:
      detach: true
      cmd: ["sleep", "50"]
  service3:
    image: busybox
    run:
      detach: true
      cmd: ["sleep", "50"]
groups:
  foo:
    - service1
    - service2
  bar:
    - service2
    - service3
hooks:
  foo:
    post-start: echo container from foo started
  bar:
    post-stop: echo container from bar stopped
  service3:
    post-stop: echo container service3 stopped
```

Hooks can be defined on a group level (`foo`, `bar`) so that they apply to all containers within that group, or directly on a container (`service3`). At most one hook can be registered per container and per event. When more than one hook is found for a given container and a given event, the following rules apply:
* Container-defined hooks have priority over group-defined ones, so in the example above, only "container service3 stopped" will be echoed when stopping `service3`.
* A fatal error will be raised at startup when 2 group-inherited hooks conflict. This is not the case in the previous example; even though `foo` and `bar` both contain `service2`, the hooks they declare are disjoint.

The following hooks are currently available:
* `pre-start`: Executed before starting or running a container
* `post-start`: Executed after starting or running a container
* `pre-stop`: Executed before stopping a container
* `post-stop`: Executed after stopping a container

### YAML advanced usage
YAML gives you some advanced features like [alias](http://www.yaml.org/spec/1.2/spec.html#id2786196) and [merge](http://yaml.org/type/merge.html). They allow you to easily avoid duplicated code in your `crane.yml` file. As a example, imagine you need to define 2 different containers: `web` and `admin`. They share almost the same configuration but the `cmd` declaration. And imagine you also need 2 instances for each one for using with a node balancer. Then you can declare them as simply:

```
containers:
  web1: &web
    image: my-web-app
    run: &web-run
      link: ["db:db"]
      ...
      cmd: web
  web2: *web

  admin1: &admin { <<: *web, run: { <<: *web-run , cmd: admin }}
  admin2: *admin
```

As a summary, `&anchor` declares the anchor property, `*alias` is the alias indicator to simply copy the mapping it references, and `<<: *merge` includes all the mapping but let you override some keys.

## Some Crane-backed sample environments
* [Silex + Nginx/php-fpm + MySQL](https://github.com/michaelsauter/silex-crane-env)
* [Symfony2 + Apache + MySQL](https://github.com/michaelsauter/symfony2-crane-env)
* [Sinatra + PostgreSQL](https://github.com/michaelsauter/sinatra-crane-env)

## Copyright & Licensing
Copyright © 2013-2015 Michael Sauter. See the LICENSE file for details.

---

[![GoDoc](https://godoc.org/github.com/michaelsauter/crane?status.png)](https://godoc.org/github.com/michaelsauter/crane)
[![Build Status](https://travis-ci.org/michaelsauter/crane.svg?branch=master)](https://travis-ci.org/michaelsauter/crane)
