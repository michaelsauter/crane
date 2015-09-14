# Changelog

## 2.0.0 (unreleased)

* `start` behaves like `run` did in 1.x, `run` and `lift` behave like their `--recreate` counterparts in 1.x
* One target only. This simplification was needed for ad-hoc commands (see below)
* Automatic resolution and handling of dependencies when creating and or starting containers
* Extension of target to affected containers and/or dependencies (replaces cascade flags `--cascade-affected` and `--cascade-dependencies`)
* Ad-hoc cmd for `lift` / `run` / `create`
* Uniquely named containers (allowing multiple containers of the same configuration with different commands at the same time)
* `--exclude` groups/containers (can be used in most cases instead of the removed `--ignore-missing` flag)
* Trigger `post-start` hooks after `start` event is sent
* `crane exec`, which starts container first if necessary
* Build hooks (`pre-build` and `post-build`) _@t-suwa_
* Replace `dockerfile` key with a `build` map (with a `context` key, which is the equivalent of `dockerfile`)
* Add `--prefix` option, which adds a prefix to each container name in the target (can also be set via `CRANE_PREFIX`) _@tmc_
* Allow configuration to be specified via `CRANE_CONFIG`, too

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