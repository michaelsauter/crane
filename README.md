# Crane
Lift containers with ease - [www.crane-orchestration.com](https://www.crane-orchestration.com?utm_source=github&utm_medium=web&utm_campaign=readme&utm_content=header)


## Overview
Crane is a Docker orchestration tool similar to Docker Compose with extra
features and (arguably) smarter behaviour. It works by reading in some
configuration (JSON or YAML) which describes how to run containers. Crane is
ideally suited for development environments or continuous integration.

## Features

* Extensive support of Docker run flags
* Simple configuration with 1:1 mapping to Docker run flags
* `docker-compose` compatible
* **ultra-fast bind-mounts via Unison on Mac** in the PRO version
* Shortcut commands
* Grouping of containers
* Excluding / limiting containers via CLI
* Smart detach / attach behaviour
* Verbose output which shows exact Docker commands
* Hooks
* ... and much more!

## Documentation & Usage

Please see [www.crane-orchestration.com](https://www.crane-orchestration.com?utm_source=github&utm_medium=web&utm_campaign=readme&utm_content=docs).

## Installation

The latest release is 3.4.2 and requires Docker >= 1.13.
Please have a look at the [changelog](https://github.com/michaelsauter/crane/blob/master/CHANGELOG.md) when upgrading.

The free version can be installed via:

```
bash -c "`curl -sL https://raw.githubusercontent.com/michaelsauter/crane/v3.4.2/download.sh`" && \
mv crane /usr/local/bin/crane
```

**If you are on Mac, check out the PRO version which seamlessly integrates
[ultra-fast bind-mounts via Unison](https://www.crane-orchestration.com?utm_source=github&utm_medium=web&utm_campaign=readme&utm_content=pro)**.

---

Copyright © 2013-2018 Michael Sauter. See the LICENSE file for details.

---

[![GoDoc](https://godoc.org/github.com/michaelsauter/crane?status.png)](https://godoc.org/github.com/michaelsauter/crane)
[![Build Status](https://travis-ci.org/michaelsauter/crane.svg?branch=master)](https://travis-ci.org/michaelsauter/crane)
