# Crane
Lift containers with ease - [michaelsauter.github.io/crane](https://michaelsauter.github.io/crane/)


## Overview
Crane is a Docker orchestration tool similar to Docker Compose with extra
features and (arguably) smarter behaviour. It works by reading in some
configuration (JSON or YAML) which describes how to run containers. Crane is
ideally suited for development environments or continuous integration.

## Features

* Extensive support of Docker run flags
* Simple configuration with 1:1 mapping to Docker run flags
* `docker-compose` compatible
* **ultra-fast bind-mounts via Unison on Mac**
* Shortcut commands
* Flexible ways to target containers (through groups and CLI flags to exclude/limit)
* Smart detach / attach behaviour
* Verbose output which shows exact Docker commands
* Hooks
* ... and much more!

## Documentation & Usage

Please see [michaelsauter.github.io/crane/docs.html](https://michaelsauter.github.io/crane/docs.html).

## Installation

The latest release is 3.6.0 and requires Docker >= 1.13.
Please have a look at the [changelog](https://github.com/michaelsauter/crane/blob/master/CHANGELOG.md) when upgrading.

```
bash -c "`curl -sL https://raw.githubusercontent.com/michaelsauter/crane/v3.6.0/download.sh`" && \
mv crane /usr/local/bin/crane
```

---

Copyright © 2013-2020 Michael Sauter. See the LICENSE file for details.

---

[![GoDoc](https://godoc.org/github.com/michaelsauter/crane?status.png)](https://godoc.org/github.com/michaelsauter/crane)
[![Build Status](https://travis-ci.org/michaelsauter/crane.svg?branch=master)](https://travis-ci.org/michaelsauter/crane)
