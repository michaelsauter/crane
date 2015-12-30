# Changelog

## 2.4.0 (2015-12-30)

* Fix a few minor configuration parsing problems when using maps.
  _@bjaglin_

* Allow usage of environment variables for _all_ flags' defaults
  Default for flag `foo-bar` can be provided via the environment variable
  `CRANE_FOO_BAR`.
  _@bjaglin_

* Gracefully ignore excluded containers in IPC/net dependencies.
  _@bjaglin_

* Support several references for exclusion
  `--exclude` can now be repeated on the CLI, and several values can be passed
  via `CRANE_EXCLUDE` using newline as a value separator.
  _@bjaglin_

* Add escape sequence for `$` in configuration
  `$$` now expands to `$` in all string configuration values.
  _@bjaglin_

* Add support for Docker networks
  They can be configured via a top-level `networks` setting, and used from
  containers via e.g. `net: foo`.
  _@michaelsauter_

* Add non-Docker dependency management
  Containers learned a new top-level `requires` setting, which works exactly
  like links (minus the alias support) and can be used to reach dependent
  containers when relying on Docker networks.
  _@michaelsauter_

* Add support for Docker volumes
  They can be configured via a top-level `volumes` setting, and used from
  containers via e.g. `volume: ["foo:/path"]`.
  _@michaelsauter_

## 2.3.1 (2015-12-12)

* Fix compatibility with engine < 1.8.0 and wrongly-forced option
  memory-swappinness on engine >= 1.8.0
  _@bjaglin_

## 2.3.0 (2015-12-09)

* When calculating affected containers, only existing ones are taken into
  account now
  _@bjaglin_

* Add flags `--ipc`, `--memory-reservation`, `--dns-opt`, `--stop-signal`,
  `--kernel-memory`, `--group-add`, `--memory-swappiness`
  _@bjaglin_

* Fix regression in `--net=container:foo`
  _@bjaglin_

* Fix removing containers when one or more of the targets is marked as unique
  but no container exists for it
  _@michaelsauter_

## 2.2.0 (2015-11-10)

* Add new `generate` command, which passes the targeted portion of the
  configuration through a given template to produce some output.

* Remove `graph` command. The same output can be achieved with the new
  `generate` command specifying this
  [DOT template](https://raw.githubusercontent.com/michaelsauter/crane-templates/7171e8b1ef6c80666ed2da8bdcbc8849aaef2d2a/dot.tmpl).

* Unique containers can now be addressed by Crane later on, e.g. given a
  unique container `foo`, `crane kill foo`  will kill all instances of `foo`.
  All other commands that did not work with unique containers previously, e.g.
  `status` or `logs` will take unique containers into account now as well.

* Fix broken `stats` and `logs` commands if a prefix was given.

* Add `--tag` global flag, which overrides image tag part temporarily.
  A typical use of `--tag` flag is to synchronize image tags with the tag of
  VCSs. You can also set the tag via `CRANE_TAG` environment variable.
  _@t-suwa_

  Example:
  If you specify a `--tag rc-2`, you will get these results:

  |original image name|overridden image name|
  |-------------------|---------------------|
  |nginx              |nginx:rc-2           |
  |nginx:1.9          |nginx:rc-2           |
  |repo/nginx         |repo/nginx:rc-2      |
  |host:5000/nginx    |host:5000/nginx:rc-2 |

* [Internal] Introduce new `ContainerInfo` interface which is a subset of
  `Container`. At the same time, clean up the `Container` interface to include
  only the externally used methods.


## 2.1.0 (2015-10-15)

* Add new `file` key to the `build` map. Equivalent of `docker build --file=<file>`

  Example:
  ```
  containers:
    foo:
      image: foo
      build:
        context: "."
        file: other_dockerfile.dkr
  ```
  _@dreamcat4_

* Add support for `--dns-search` flag
  _@scornelissen85_

* Fix broken `net` flag


## 2.0.1 (2015-09-16)

* Fixes messed up output for `crane status` using Docker 1.8.


## 2.0.0 (2015-09-15)

* `start` behaves like `run` did in 1.x, `run` and `lift` behave like their
  `--recreate` counterparts in 1.x. The flag `--recreate` was removed
  consequently.

* Remove multiple target handling. Commands can only be applied to a single
  target now, which can be either a group or a container. This simplification
  was needed for ad-hoc commands (see below).

* Automatic resolution and handling of dependencies when creating and or
  starting containers. This is different to 1.x, in which the target needed to
  include all containers required for the command to succeed.

  Example:
  Given a configuration with container `web` depending on container `database`,
  in 1.x you could not use `crane run web`, since the linked `database` was not
  included in the target. In 2.x however, Crane detects that a running
  `database` is required for `web` to run, which means it will bring `database`
  into a running state (by executing the equivalent of `crane start database`
  before dealing with `web`). Note that e.g. `crane kill web` will only kill
  `web` and do nothing to `database`, since a normal `docker kill web` will
  succeed just fine.

* Extension of target to affected containers and/or dependencies. This replaces
  the cascade flags `--cascade-affected` and `--cascade-dependencies`.

  Example:
  Given a configuration with container `web` depending on container `database`,
  if you wanted to make sure both `web` and `database` are recreated, you can
  use the target `web` and extend it with `dependencies` like so:
  `crane run web+dependencies`. `affected` works the other way around, so if
  you want to recreate `database` and all containers that depend on it,
  execute `crane run database+affected`. You may also use the shortcuts `d` and
  `a`.

* Ad-hoc cmd for `lift` / `run` / `create`.

  Example:
  If you have a multi-purpose `ruby` container in your configuration, you can
  now run multiple commands, e.g. `crane run ruby pry` and `crane run ruby irb`.

* Uniquely named containers. This configuration option (`unique: true`) allows
  multiple instances of one container configuration at the same time. Crane will
  append a timestamp (with millisecond precision) to the container name to make
  it unique. It is advised to use this together with `run: {rm: true}` since
  Crane is not able to e.g. kill unique containers.

* Remove `--ignore-missing` flag since it does not work well with the new
  dependency resolution. As an alternative, a new `--exclude` flag has been introduced, which allows to exclude a group or a container.

  Example:
  Given a configuration with container `web` depending on container `database`,
  you can just run `web` by itself using `crane --exclude database run web`.

* Trigger `post-start` hooks after `start` event is sent. Previously, it was
  executed after control was handed back to Crane, which could be very late in
  case the container attached to STDIN.

* `crane exec`, which maps to `docker exec`, but additionaly starts the targeted
  container(s) first if necessary. The options `interactive` and `tty` can be
  configured underneath the `exec` key in the configuration.

* Build hooks (`pre-build` and `post-build`)
  _@t-suwa_

* Remove `dockerfile` key. Instead, there is a `build` map now with a `context`
  key (which is the equivalent of `dockerfile`).

  Example:
  ```
  containers:
    foo:
      image: foo
      build:
        context: "."
  ```
  In the future, the `build` map might be extended with further options used by
  `docker build`. The map was introduced now so those changes can be made in a
  backwards-compatible way.

* Add `--prefix` option, which adds a prefix to each container name in the
  target.

  Example:
  `crane --prefix="foo_" run web` will run `foo_web`.

* Allow configuration to be specified via `CRANE_CONFIG`, and prefix via
  `CRANE_PREFIX`.
  _@tmc_

Required Docker version: >= 1.6


## 1.5.1 (2015-07-29)

* Allow Crane config in filesystem root _@bjaglin_
* Support `--cpu-period`, `--cpu-quota`, `--oom-kill-disable`, `--uts` and `--blkio-weight` _@bjaglin_
* Support `--since` flag for logs command _@bjaglin_

Required Docker version: >= 1.3


## 1.5.0 (2015-07-07)

* Support for `--log-opt`
* Execute commands in directory of config
* Don't provision the same image twice
* Speed up image check

Required Docker version: >= 1.3

Thanks a lot to @bjaglin and @jesper for the contributions!


## 1.4.0 (2015-06-20)

* Allow Crane to be used inside sub-directories (similar to Git binary)
* Execute stop hooks also when a running container is killed or removed

Both changes are potentially breaking, so please check if you're affected.

Required Docker version: >= 1.3


## 1.3.1 (2015-05-29)

* Only pull images for which no Dockerfile is specified

Required Docker version: >= 1.3


## 1.3.0 (2015-05-28)

* Add support for `ulimit`, `log-driver`, `label`, `label-file` and `cgroup-parent` options
* Add `crane pull` mapping to `docker pull`
* Add Windows exe (which may or may not work properly)
* Corrected dependency handling for `--volumes-from` when suffixes such as ro are used

Required Docker version: >= 1.3

Thanks a lot to @bjaglin for the contributions!


## 1.2.0 (2015-05-05)

* Container hooks
* `--ignore-missing` flag
* Support for multiple links to the same container
* `env` declaration as mapping
* Improved docs

Required Docker version: >= 1.3

Thanks a lot to @bjaglin, @mishak87 and @adrianhurt for the contributions!


## 1.1.1 (2015-02-15)

* `sig-proxy` option applied correctly

Required Docker version: >= 1.3

Thanks a lot to @bjaglin for the contribution!


## 1.1.0 (2015-02-09)

* Add support for `mac-address`, `pid`, `read-only`, `memory-swap`, `security-opt` and `sig-proxy` options
* `crane stats`, mapping to `docker stats`
* Raise error if group reference is not a container
* Raise error if YAML is invalid
* Dump errors and verbose output to `STDERR`

Required Docker version: >= 1.3

Thanks a lot to @lefeverd and @bjaglin for the contributions!


## 1.0.0 (2014-11-27)

* Added `logs` subcommand mapping to `docker logs` (thanks @bjaglin)

Required Docker version: >= 1.3


## pre 1.0

Please see the [releases](https://github.com/michaelsauter/crane/releases).
