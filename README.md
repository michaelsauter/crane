# Crane
Lift containers with ease

## Overview
Crane is a little tool to orchestrate Docker containers. It works by reading in a `Cranefile` (a JSON file) which describes how to obtain container images and how to run them. This simplifies setting up a development environemt a lot as you don't have to bring up every container manually, remembering all the arguments you need to pass. By storing the `Cranefile` next to the data and the app(s) in a repository, you can easily share the whole development environment.

## Installation
Dowload [the latest release](https://github.com/michaelsauter/crane/releases/download/v0.4.0/crane) of `crane` and put it in your path, e.g. in `/usr/local/bin`.

Of course, you will need to have Docker (> 0.8) installed on your system. If you are on OS X, I recommend using [docker-osx](https://github.com/noplay/docker-osx).

## Usage
Crane is a very light wrapper around the Docker commands. This means that e.g. `run`, `rm`, `kill`, `start`, `stop` just call the corresponding Docker commands, but for all defined containers. There are only two exception: `provision` which either calls Docker's `build` or `pull` (depending on whether a Dockerfile is specified) and `lift`, which will build and run the containers in one go.
You can get more information about what's happening behind the scenes by using `--verbose`.
Some commands have a `--force` flag, which will save you intermediate steps, such as stopping the containers before removing them, or rebuilding images even though they exist already.
For all available commands and details on usage, just type `crane`.

## Cranefile
A `Cranefile` defines the containers, one after the other. One caveat: if a container depends on another one, it must appear before that container in the file.
Every container consists of:

* `name` (string, required): Name of the container
* `image` (string, required): Name of the image to build/pull
* `dockerfile` (string, optional): Relative path to the Dockerfile
* `run` (object, optional): Parameters mapped to Docker's `run`.
  * `cidfile` (string)
  * `cpu-shares` (integer)
  * `detach` (boolean) `sudo docker attach <container name>` will work as normal.
  * `dns` (array)
  * `env` (array)
  * `expose` (array) Ports to expose to linked containers.
  * `host` (string)
  * `interactive` (boolean)
  * `link` (array) Link containers.
  * `lxc-conf` (array)
  * `memory` (string)
  * `privileged` (boolean)
  * `publish` (array) Map network ports to the container.
  * `publish-all` (boolean)
  * `rm` (boolean)
  * `tty` (boolean)
  * `user` (string)
  * `volume` (array) In contrast to plain Docker, the host path can be relative.
  * `volumes-from` (array) Mount volumes from other containers
  * `workdir` (string)
  * `cmd` (string) Command to append to `docker run` (overwriting `CMD`).

See the [Docker documentation](http://docs.docker.io/en/latest/reference/commandline/cli/#run) for more details about the parameters.

## Example
For demonstration purposes, we'll bring up a PHP app (served by Apache) that depends both on a MySQL database and a Memcached server. The source code is available at http://github.com/michaelsauter/crane-example. Here's what the Cranefile looks like:

```
[
  {
    "name": "crane_apache",
    "dockerfile": "apache",
    "image": "icrane_apache",
    "run": {
      "volumes-from": ["crane_app"],
      "publish": ["80:80"],
      "link": ["crane_mysql:db", "crane_memcached:cache"],
      "detach": true
    }
  },
  {
    "name": "crane_app",
    "dockerfile": "app",
    "image": "icrane_app",
    "run": {
      "volume": ["app/www:/srv/www:rw"],
      "detach": true
    }
  },
  {
    "name": "crane_mysql",
    "dockerfile": "mysql",
    "image": "icrane_mysql",
    "run": {
      "detach": true
    }
  },
  {
    "name": "crane_memcached",
    "dockerfile": "memcached",
    "image": "icrane_memcached",
    "run": {
      "detach": true
    }
  }
]
```
If you have Docker installed, you can just clone that repository and bring up the environment right now.
In the folder where the Cranefile is, type:

```
[sudo] crane lift
```

This will bring up the containers. The container running Apache has the MySQL and Memcached containers automatically linked. Open `http://localhost` and you should be greated with "Hello World".
