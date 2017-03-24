# Crane
Lift containers with ease - [www.craneup.tech](https://www.craneup.tech)


## Overview
Crane is a Docker orchestration tool similar to Docker Compose with extra
features. It works by reading in some configuration (JSON or YAML) which
describes how to obtain images and how to run containers. Crane is ideally
suited for setting up a development environment or continuous integration.

There's a free version for Linux and a pro version for macOS which offers
**ultra-fast bind-mounts via Unison**!

## Features

* Extensive support of Docker run flags
* Intuitive configuration format with 1:1 mapping to Docker run flags
* `docker-compose` compatible
* Fast bind-mounts via Unison on macOS
* Grouping of containers
* Excluding / limiting containers via CLI
* Smart detach / attach behaviour
* Verbose output which shows exact Docker commands
* Hooks
* Container / network / volume prefix
* Tag override
* Generate other files from configuration via templates
* ... and much more!

## Documentation & Usage

Please see [www.craneup.tech](https://www.craneup.tech) for more information
and documentation.

## Copyright & Licensing
Copyright © 2013-2017 Michael Sauter. See the LICENSE file for details.

---

[![GoDoc](https://godoc.org/github.com/michaelsauter/crane?status.png)](https://godoc.org/github.com/michaelsauter/crane)
[![Build Status](https://travis-ci.org/michaelsauter/crane.svg?branch=master)](https://travis-ci.org/michaelsauter/crane)
