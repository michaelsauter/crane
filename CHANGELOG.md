# Changelog

## Unreleased

* A prefix can be specified in the configuration now. When no prefix is set,
  the default (the folder name) is used. When `prefix: false` is configured,
  by default no prefix is used (this was Crane's 2.x behaviour). An explicit
  prefix can be set via `prefix: "foo"`.
* Support `--health-cmd`, `--health-interval`, `--health-retries`,
  `--health-timeout` and `--no-healthcheck`. Configuration via docker-compose's
  `healtcheck` object is also possible (though only string is supported for
  `test` at the moment).

## 3.0.0 beta (2017-03-13)

As this is a major release, lots of things have changed. Please review the
following list very carefully and adjust as needed.

CLI interface changes:

* `lift` has been renamed to `up`.
* `rm` doesn't have the short flag `-f` anymore, but `--force` still works
* `rm` learned `--volumes` to removed associated volumes, too
* `exec` adds `--interactive` and `--tty` by default, and supports specifying
  `--privileged` and `--user USER`
* The short-hand for `--exclude` is `-x` now
* Dynamic targets (`+affected`/`+dependencies`) have been dropped, but it is
  still possible to extend a target to all dependencies by using
  `--extend`/`-e`. `+affected` was mainly useful with legacy links and should
  not be needed with the "new" networks. If you used it regularly, please open
  an issue.
* `lift`/`run` by default attach to the target, and detach from dependencies. If
  the target is configured to detach (either via file or `--detach`/`-d`)
  however, then `lift`/`run` detach from the target, too.
* The configuration flag `--config`/`-c` can be repeated now. The specified
  configurations are merged together, last one wins. Multiple configuration
  files can also be set via `CRANE_CONFIG` by separating the files via colons.
  Consequently, the `--override` flag has been dropped.

Configuration changes:

* The parameters for `stop`, `start` and `exec` have been removed. The
  corresponding CLI commands have learned some of the options to mitigate this.
* The top-level configuration key `containers` has been renamed to `services` to
  align with docker-compose.
* All `run ` parameters have been moved up one level to align with
  docker-compose.
* `cmd` has to be specified as `command` now.
* `image` is optional now. If it is not given, the service name will be used as
  the image name.
* `stop-timeout`, `sysctl` and `userns` can be configured now.
* Most configuration keys have an alternative name now, e.g. `publish` can now
  be specified as `ports`. The alternative name is the one that docker-compose
  uses. This allows Crane to read docker-compose configuration files, with some
  minor exceptions.
* `net` also allows the form `service:<container>` now.

Behaviour changes:

* Crane reads `docker-compose.yml`, `docker-compose.override.yml`, `crane.yml`,
  `crane.override.yml` by default now (in this order).
* To align with docker-compose, Crane sets a default prefix now, which is the
  name of the directory where the configuration files are in. You can remove
  the prefix by passing `--prefix=""`/`-p ""` or setting `CRANE_PREFIX=""`.
* If neither `net`, `network_mode` or `networks` is configured, the default
  network mode is no longer `bridge` but a `default` network (to align with
  docker-compose).
* The integration of Unison on macOS has been extracted into a paid pro version,
  see [www.craneup.tech](https://www.craneup.tech). There is still a free
  version for macOS without this feature. If you would like to support Crane,
  you can also buy a pro version for Linux, but right now there is no difference
  in functionality.


## 2.11.0 (2016-11-14)

* Allow to override configuration by another configuration file. By default,
  Crane is looking for e.g. `crane.override.yml` if the config file in use is
  `crane.yml`. This can be customized via `--override`/`CRANE_OVERRIDE`.

  _@michaelsauter_

## 2.10.3 (2016-10-27)

* Add `--debug` flag to `crane mac-sync start` which turns on verbose logging
  and runs the sync in the foreground.

  _@michaelsauter_

* Add image, flags, uid and gid to md5 hash identifying the Mac sync server.

  _@michaelsauter_

* Check Unison client as well when determining whether a Mac sync is running.

  _@michaelsauter_

## 2.10.2 (2016-10-13)

* Add `autostart` option to Mac syncs.

  _@michaelsauter_

## 2.10.1 (2016-10-10)

* Allow to pass relative host directories to `mac-sync start/stop`.

  _@michaelsauter_

* Print error message when volume passed to `mac-sync start/stop` is not
  configured.

  _@michaelsauter_

## 2.10.0 (2016-10-08)

* Optional Unison sync for Docker for Mac to improve performance.
  For more information, see [Docker for Mac with Unison sync](https://github.com/michaelsauter/crane#docker-for-mac-with-unison-sync).

  _@michaelsauter_

* Built with Go 1.7.1, which might solve potential issues with macOS Sierra.

  _@michaelsauter_

## 2.9.1 (2016-08-15)

* Fix tag override for images with no tag specified in the config.

  _@michaelsauter_

## 2.9.0 (2016-05-16)

* Add `--dry-run` option to see what a certain command would do without actually
  doing anything.

  _@michaelsauter_

* Add `--subnet` support for networks

  _@michaelsauter_

* Document `--only` flag

  _@michaelsauter_

## 2.8.2 (2016-05-04)

* Add more error output when hook execution fails

  _@michaelsauter_

* Fix usage of CLI long flags prefixed by `--no-`

  _@bjaglin_

* Add support for `--no-stream` to `crane stats` (Docker >= 1.7)

  _@bjaglin_

## 2.8.1 (2016-04-21)

* Fix broken target when a default group is specified. The regression was
  introduced with #275, and reported in #285.

  _@michaelsauter_

## 2.8.0 (2016-04-10)

* Implicit ad-hoc containers. The pre-defined `unique` key is removed in favour
  of treating every container as unique when a command is passed via the CLI.
  The container name is suffixed with a timestamp, and the following changes are
  made to the configuration:

  * `publish`, `publish-all`, `ip`, `ip6` and `detach` are disabled
  * `rm` is enabled

  Note that ad-hoc containers are not targeted by Crane in any way, e.g. when
  running `crane rm`.

  _@michaelsauter_

* Add `--only` flag to restrict command to a container or group. This can be
  used for example to start one container without its dependencies.

  _@michaelsauter_

* Rename short flag `--output` for `generate` command from `-o` to `-O` since
  `-o` is the short flag for `--only` now. Since the generate feature is
  marked as experimental, this change is done in a minor version.

  _@michaelsauter_

## 2.7.0 (2016-03-20)

* Refactor `exclude` behaviour to make it easier to support `only` in the future

  _@michaelsauter_

* Add missing flags `--ip` and `--ip6` from Docker 1.10 to Crane

  _@dreamcat4_

* Better handling of incorrect syntax for `env` and `label` configuration values

  _@bjaglin_

* `docker run/exec/start` flags catch-up in Crane configuration.

  _@bjaglin_

* Add new `build-arg` key to the `build` map. Equivalent of `docker build --build-arg KEY=VALUE`

  Example:
  ```
  containers:
    foo:
      image: foo
      build:
        context: "."
        build-arg:
          - KEY=VALUE
  ```
  Requires Docker 1.9+

## 2.6.0 (2016-02-25)

* Do not require link containers when requires is set. This is a breaking change
  if you defined both `requires` and `run/link`, and relied on Crane resolving
  the containers defined only in `run/link`. Now, all entries in `run/link` are
  treated as aliases only, and Crane does not handle them specifically. This
  allows to have optional dependencies when using Docker 1.9+ networks.

  _@michaelsauter_

* Update dependencies. While this should generally not have any side-effects, it
  is likely that coloring support was broken on Windows earlier and might be
  fixed now.

  _@michaelsauter_

## 2.5.2 (2016-02-19)

* Limit the number of Docker calls when cascading commands to affected containers.

  _@bjaglin_

## 2.5.1 (2016-02-03)

* Bugfix: Wait for post-start hook to complete. #258

  Otherwise the hook runs in the background and does not block.
  _@jesper_

* Bugfix: De-duplicate required volumes and networks in verbose output. #253

  _@michaelsauter_


## 2.5.0 (2016-01-09)

* Expose level of concurrency when provisioning

  `lift` and `provision` commands can now be sped-up by passing
  a custom parallelism level via `--parallel`/`-l`.
  _@bjaglin_

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
