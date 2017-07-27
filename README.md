# Crane
Lift containers with ease - [www.craneup.tech](https://www.craneup.tech)


## Overview
Crane is a Docker orchestration tool similar to Docker Compose with extra
features and (arguably) smarter behaviour. It works by reading in some
configuration (JSON or YAML) which describes how to run containers. Crane is
ideally suited for development environments or continuous integration.

Next to the free version (Linux, Mac, Windows) there is a pro version for Mac which offers **ultra-fast bind-mounts via Unison**!

## Features

* Extensive support of Docker run flags
* Intuitive configuration format with 1:1 mapping to Docker run flags
* `docker-compose` compatible
* Grouping of containers
* Excluding / limiting containers via CLI
* Smart detach / attach behaviour
* Verbose output which shows exact Docker commands
* Hooks
* Container / network / volume prefix
* Tag override
* Generate other files from configuration via templates
* ... and much more!

## Installation, Documentation & Usage

Please see [www.craneup.tech](https://www.craneup.tech).

## Copyright & Licensing
Copyright © 2013-2017 Michael Sauter. See the LICENSE file for details.

---

[![GoDoc](https://godoc.org/github.com/michaelsauter/crane?status.png)](https://godoc.org/github.com/michaelsauter/crane)
[![Build Status](https://travis-ci.org/michaelsauter/crane.svg?branch=master)](https://travis-ci.org/michaelsauter/crane)
