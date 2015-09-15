# Changelog

## 2.0.0 (unreleased)

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
  you want to recreate `database` and all contaienrs that depend on it,
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