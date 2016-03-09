# Crane
Lift containers with ease


## Overview
Crane is a tool to orchestrate Docker containers. It works by reading in some
configuration (JSON or YAML) which describes how to obtain images and how to run
containers. Crane is ideally suited for setting up a development environment or
continuous integration.

* [Installation](#installation)
* [Usage](#usage)
* [Configuration](#configuration)
* [Example](#example)
* [Advanced Usage](#advanced-usage)
  * [Groups and Targeting](#groups-and-targeting)
  * [Extending the target](#extending-the-target)
  * [Excluding containers](#excluding-containers)
  * [Networking](#networking)
  * [Volumes](#volumes)
  * [Hooks](#hooks)
  * [Parallism](#parallism)
  * [Container Prefixes](#container-prefixes)
  * [Override image tag](#override-image-tag)
  * [Unique Names](#unique-names)
  * [Generate command](#generate-command)
  * [YAML advanced usage](#yaml-advanced-usage)


## Installation
The latest release (2.6.0) can be installed via:

```
bash -c "`curl -sL https://raw.githubusercontent.com/michaelsauter/crane/v2.6.0/download.sh`" && sudo mv crane /usr/local/bin/crane
```

Older releases can be found on the
[releases](https://github.com/michaelsauter/crane/releases) page. You can also
build Crane yourself by using the standard Go toolchain.

Please have a look at the
[changelog](https://github.com/michaelsauter/crane/blob/v2.6.0/CHANGELOG.md)
when upgrading.

Of course, you will need to have Docker (>= 1.6) installed.


## Usage
Crane is a very light wrapper around the Docker CLI. This means that most
commands just call the corresponding Docker command, but for all targeted
containers. The basic format is `crane <command> <target>`, where `<command>`
corresponds to a Docker command, and `<target>` either is a single container or
a [group](#groups-and-targeting).

When executing commands, keep the following 2 rules in mind:

1. Crane will apply the command ONLY to the target
2. Crane will do with other containers in the configuration whatever it takes in
order for (1) to succeed

As an example, imagine you have a container `web` depending on container
`database`. When you execute `crane run web`, then Crane will start `database`
first, then run `web` (recreating `web` if it already exists). There are ways to
[dynamically extend the target](#extending-the-target) (so that `database` would
be recreated as well for example).

Following are a list of supported commands and possible options:

| Command | Maps to | Explanation and Options |
| ------------- | ----------- | ---------|
| create      | create           | Creates containers. Any existing containers are removed first. |
| run         | run              | Runs containers. Any existing containers are removed first. |
| rm          | rm               | Removes containers if they exist. Use `--force` to kill running containers first. |
| start       | start            | Starts containers if not they are not running yet. Non-existing containers are created first. |
| stop        | stop             | Stops containers if they are running. |
| kill        | kill             | Kills containers if they are running. |
| pause       | pause            |  |
| unpause     | unpause          |  |
| exec        | exec             | Starts container first if not running yet. |
| logs        | logs             | Logs of containers are multiplexed. Use `--follow` to follow log output. |
| stats       | stats            | |
| push        | push             | |
| pull        | pull             | |
| provision   | pull/build       | Calls Docker's `pull` if no Dockerfile is specified. Otherwise it builds the image, optionally with disabled cache by passing `--no-cache`. |
| lift        | pull/build + run | Provisions and runs containers in one go. Use `--no-cache` to disable build cache. |
| status      | -                | Displays information similar to `docker ps` for the given target. |
| generate    | -                | Passes the targeted portion of the config through given `--template` and outputs the result to STDOUT or given `--output` file. |

You can get more information about what's happening behind the scenes for all commands by using `--verbose`. Most options have a short version as well, e.g. `lift -n`. The CLI provides help for every command, e.g. `crane help run`.


## Configuration
The configuration defines a map of containers in either JSON or YAML. By default, the configuration is expected in a file named `crane.json` or `crane.yaml`/`crane.yml`. The file can also be set via `--config` or `CRANE_CONFIG`. If the given path is relative, Crane searches for the configuration in the current directory, then recursively in the parent directory. Dependencies between containers are automatically detected and resolved.
The map of containers consists of the name of the container mapped to the container configuration, with the following keys:

* `image` (string, required): Name of the image to build/pull
* `unique` (boolean, optional) `true` assigns a unique name to the container (experimental)
* `requires` (array) Container dependencies (experimental)
* `run` (object, optional): Parameters mapped to Docker's `run` & `create`.
	* `add-host` (array) Add custom host-to-IP mappings.
	* `blkio-weight` (integer) Need Docker >= 1.7
	* `cap-add` (array) Add Linux capabilities.
	* `cap-drop` (array) Drop Linux capabilities.
	* `cgroup-parent` (string)
	* `cidfile` (string)
	* `cpu-period` (integer) Need Docker >= 1.7
	* `cpu-quota` (integer) Need Docker >= 1.7
	* `cpuset` (integer)
	* `cpu-shares` (integer)
	* `detach` (boolean) `sudo docker attach <container name>` will work as normal.
	* `device` (array) Add host devices.
	* `dns` (array)
	* `dns-opt` (array) Need Docker >= 1.9
	* `dns-search` (array)
	* `entrypoint` (string)
	* `env` (array/mapping) Can be declared as a string array with `"key[=value]"` items or a string-to-string mapping where each `key: value` will be translated to the corresponding `"key=value"` string.
	* `env-file` (array)
	* `expose` (array) Ports to expose to linked containers.
	* `group-add` (array) Need Docker >= 1.8
	* `hostname` (string)
	* `interactive` (boolean)
	* `ipc` (string) The `container:id` syntax is not supported, use `container:name` if you want to reuse another container IPC.
	* `kernel-memory` (string) Need Docker >= 1.9
	* `label` (array/mapping) Can be declared as a string array with `"key[=value]"` items or a string-to-string mapping where each `key: value` will be translated to the corresponding `"key=value"` string.
	* `label-file` (array)
	* `link` (array) Link containers.
	* `log-driver` (string)
	* `log-opt` (array) Need Docker >= 1.7
	* `lxc-conf` (array)
	* `mac-address` (string)
	* `memory` (string)
	* `memory-reservation` (string) Need Docker >= 1.9
	* `memory-swap` (string)
	* `memory-swappiness` (int) Need Docker >= 1.8
	* `net` (string) The `container:id` syntax is not supported, use `container:name` if you want to reuse another container network stack.
	* `oom-kill-disable` (bool) Need Docker >= 1.7
	* `pid` (string)
	* `privileged` (boolean)
	* `publish` (array) Map network ports to the container.
	* `publish-all` (boolean)
	* `read-only` (boolean)
	* `restart` (string) Restart policy.
	* `security-opt` (array)
	* `stop-signal` (string)  Need Docker >= 1.9
	* `sig-proxy` (boolean) `true` by default
	* `rm` (boolean)
	* `tty` (boolean)
	* `ulimit` (array)
	* `user` (string)
	* `uts` (string) Need Docker >= 1.7
	* `volume` (array) In contrast to plain Docker, the host path can be relative.
	* `volumes-from` (array) Mount volumes from other containers
	* `workdir` (string)
	* `cmd` (array/string) Command to append to `docker run` (overwriting `CMD`).
* `rm` (object, optional): Parameters mapped to Docker's `rm`.
	* `volumes` (boolean)
* `start` (object, optional): Parameters mapped to Docker's `start`.
	* `attach` (boolean)
	* `interactive` (boolean)
* `build` (object, optional): Parameters mapped to Docker's `build`.
	* `context` (string)
	* `file` (string)
  * `build-arg` (array/mapping) Provide build arguments. Need Docker >= 1.9
* `exec` (object, optional): Parameters mapped to Docker's `exec`.
  * `interactive` (boolean)
  * `tty` (boolean)

Note that basic environment variable expansion (`${FOO}`, `$FOO`) is supported throughout the configuration, but advanced shell features such as command substitution (`$(cat foo)`, `` `cat foo` ``) or advanced expansions (`sp{el,il,al}l`, `foo*`, `~/project`, `$((A * B))`, `${PARAMETER#PATTERN}`) are *not* as the Docker CLI is called directly. Use `$$` for escaping a raw `$`.

See the [Docker documentation](http://docs.docker.com/reference/commandline/cli/) for more details about the parameters.


## Example
Taken from a basic
[Sinatra blog app](https://github.com/michaelsauter/sinatra-crane-env), a
typical `crane.yaml` looks like this:

```
containers:
  blog:
    build:
      context: image
    image: michaelsauter/sinatra-example
    run:
      publish: ["9292:9292"]
      volume: ["blog:/blog"]
      link: ["postgres:db"]
      env:
        - "POSTGRESQL_DB=default"
        - "POSTGRESQL_USER=default"
        - "POSTGRESQL_PASS=default"
      tty: true
      interactive: true
      cmd: "start-blog"
  postgres:
    image: d11wtq/postgres
    run:
      detach: true
```

The specified Docker containers can then be created and started in the correct
order with:

```
crane lift blog
```

If you want to use JSON instead of YAML, here's what a simple configuration
looks like:

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

### Groups and Targeting

Next to containers, you can also specify groups, and then execute Crane commands
against those groups. If you do not specify any target, the command will apply
to all containers. However, you can override this by specifying a `default`
group. Also, every container can be targeted individually by using the name of
the container. Groups of containers can be specified like this (YAML shown):

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

This could be used like so: `crane provision service1`, `crane run databases`
or `crane lift services`. `crane status` is an alias for `crane status default`
in this example. If `default` were not specified, then `crane lift` would start
`database1`, `database2`, `service1` and `service2`.


### Extending the target

It is also possible to extend the target to related containers. There are 2
different "dynamic" groups, `affected` and `dependencies` (both have a short
version `a` and `d`). In our example configuration above, when targeting the
`postgres` container, if it had been started earlier, the `blog` container
would be considered to be "affected". When targeting the `blog` container,
the `postgres` container would be considered as a "dependency". Therefore
`crane run postgres+affected` will recreate both `postgres` and `blog`.
Similarly, `crane run blog+dependencies` will recreate `blog` and `postgres`.
It is possible to combine `affected` and `dependencies`.


### Excluding containers

If you want to exclude a container or a whole group from a Crane command, you
can specify this with `--exclude <reference>` (or via `CRANE_EXCLUDE`). The
flag can be repeated to exclude several containers or groups (use a multi-line
environment variable value to pass several references via `CRANE_EXCLUDE`).

Excluded containers' declaration _and_ references in the configuration file
will be completely ignored, so their dependencies will also be excluded
(unless they are also required by other non-excluded containers).

This feature is experimental, which means it can be changed or even removed
in every minor version update.


### Networking
Docker networks are supported via the top-level config `networks`, which does not take additional parameters at this stage. As links are not strict dependencies for containers attached to a user-defined network (but simply aliases), `requires` can be used instead to indicate that a container must be started for another one to be functional. Networks are automatically created by Crane when necessary, and never cleaned up. When a [prefix](#container-prefixes) is used, it is also applied to the network.

```
containers:
  foo:
    requires: ["bar"]
    run:
      net: qux
  bar:
    run:
      net: qux
networks:
  qux:
```

This feature is experimental, which means it can be changed or even removed in
every minor version update.


### Volumes
Docker volumes are supported via the top-level config `volumes`, which does not
take additional parameters at this stage. Volumes are automatically created by
Crane when necessary, and never cleaned up. When a [prefix](#container-prefixes)
is used, it is also applied to the volume.

```
containers:
  foo:
    run:
      volume: ["bar:/path"]
volumes:
  bar:
```

This feature is experimental, which means it can be changed or even removed in
every minor version update.


### Hooks

In order to run certain commands before or after key lifecycle events of containers, hooks can be declared in the configuration. They are run synchronously on the host where Crane is installed, outside containers, via an `exec` call. They may interrupt the flow by returning a non-zero status. If shell features more advanced than basic variable expansion is required, you should explicitly spawn a shell to run the command in (`sh -c 'ls *'`).

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
* `pre-build`: Executed before building an image
* `post-build`: Executed after building an image
* `pre-start`: Executed before starting or running a container
* `post-start`: Executed after starting or running a container
* `pre-stop`: Executed before stopping, killing or removing a running container
* `post-stop`: Executed after stopping, killing or removing a running container

Every hook will have the name of the container for which this hook runs available as the environment variable `CRANE_HOOKED_CONTAINER`.


### Parallism
By default, Crane executes all commands sequentially. However, you might want
to increase the level of parallelism for network-heavy operations, in order to
cut down the overall run time. The `--parallel`/`-l` flag allows you to
specify the level of parallelism for commands where network can be a
bottleneck (namely `provision` and `lift`). Passing a value of 0 effectively
disable throttling, which means that all provisioning will be done in parallel.


### Container Prefixes
It is possible to prefix containers with a global `--prefix` flag, which is just
prepended to the container name. Remember that you will have to provide the same
prefix for subsequent calls if you want to address the same set of containers. A
common use case for this feature is to launch a set of containers
in parallel, e.g. for CI builds. Container prefixes can also be supplied by the
`CRANE_PREFIX` environment variable.


### Override image tag
By using a the `--tag` flag, it is possible to globally overrides image tags. If
you specify `--tag 2.0-rc2`, an image name `repo/app:1.0` is treated as
`repo/app:2.0-rc2`. The `CRANE_TAG` environment variable can also be used to
set the global tag.


### Unique names
If `unique` is set to true, Crane will add a timestamp to the container name
(e.g. `foo` will become `foo-unique-1447155694523`), making it possible to have
multiple containers based on the same Crane config. Consider setting `rm` to
`true` at the same time to avoid lots of exited containers. This feature is
experimental, which means it can be changed or even removed in every minor
version update.


### Generate command
The `generate` command can transform (part of) the configuration based on a
given template, making it easy to re-use the configuation with other tools.
`--template` is a required flag, which should point to a Go template. By
default, the output is printed to STDOUT. It can also be written to a file using
the `--output` flag. If the given filename contains `%s`, then multiple files
are written (one per container), substituting `%s` with the name of the
container. For each container, an object of type
[ContainerInfo](https://godoc.org/github.com/michaelsauter/crane/crane#ContainerInfo)
is passed to the template. If one file is generated for all targeted containers,
a list of containers is located under the key `Containers`.  This feature is
experimental, which means it can be changed or even removed in every minor
version update.

Example templates can be found at
[crane-templates](https://github.com/michaelsauter/crane-templates).


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


## Copyright & Licensing
Copyright © 2013-2016 Michael Sauter. See the LICENSE file for details.

---

[![GoDoc](https://godoc.org/github.com/michaelsauter/crane?status.png)](https://godoc.org/github.com/michaelsauter/crane)
[![Build Status](https://travis-ci.org/michaelsauter/crane.svg?branch=master)](https://travis-ci.org/michaelsauter/crane)
