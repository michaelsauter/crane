# Crane
Lift containers with ease

## Overview
Crane is a little tool to orchestrate Docker containers. It works by reading in a `Cranefile` (a JSON file) which describes how to obtain container images and how to run them. This simplifies setting up a development environemt a lot as you don't have to bring up every container manually, remembering all the arguments you need to pass. By storing the `Cranefile` next to the data and the app(s) in a repository, you can easily share the whole development environment.

## Installation
Dowload `crane` and put it in your path. As docker needs to be executed with `sudo`, I recommend placing it e.g. in `/usr/local/bin`.

## Usage
Crane is a very light wrapper around the Docker commands. This means that e.g. `run`, `rm`, `kill`, `start`, `stop` just call the corresponding Docker commands, but for all defined containers. The only exception is `provision` which either calls Docker's `build` or `pull`, depending on whether a Dockerfile is specified. For all available commands, see `crane help`. There is only one flag, `--verbose`, which will print more output.

## Cranefile
A `Cranefile` defines the containers. The structure is very simple. Every container can have multiple nested containers which are automatically linked to their parents. Every container consists of:

* `name` (required): Name of the container
* `image` (required): Name of the image to build/pull
* `dockerfile` (optional): Relative path to the Dockerfile
* `parameters` (optional): Available are `v` (volumes), `p` (ports), `e` (environment variables), `t` (TTY), `i` (interactive) and `cmd` (command). They all map to the arguments of Docker's `run`.
* `dependencies` (optional): Hash where the key is the alias used for the container link and the value is a container

## Example
For demonstration purposes, we'll bring up a PHP app (served by Apache) that depends both on a MySQL database and a Memcached server. The source code is available at http://github.com/michaelsauter/crane-example. Here's what the Cranefile looks like:

```
{
	"name": "webapp",
	"dockerfile": "apache",
	"image": "apache",
	"parameters": {
		"v": ["apache/www:/var/www"],
		"p": ["80:80"]
	},
	"dependencies": {
		"db": {
			"name": "c-mysql",
			"dockerfile": "mysql",
			"image": "mysql"
		},
		"cache": {
			"name": "c-memcached",
			"dockerfile": "memcached",
			"image": "memcached"
		}
	}
}
```
If you have Docker installed, you can just clone that repository and bring up the environment right now.
In the folder where the Cranefile is, type:

```
sudo crane provision --verbose
sudo crane run --verbose
```

This will bring up the webapp container, which has the MySQL and Memcached automatically linked. Open `http://localhost` and you should be greated with "Hello World".