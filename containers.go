package main

import (
	"encoding/json"
	"io/ioutil"
)

type Containers []Container

func readCranefile(filename string) Containers {
	file, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(err)
	}

	var containers Containers
	err = json.Unmarshal(file, &containers)
	if err != nil {
		panic(err)
	}
	return containers
}

func (containers Containers) reversed() []Container {
	var reversed []Container
	for i := len(containers) - 1; i >= 0; i-- {
		reversed = append(reversed, containers[i])
	}
	return reversed
}

// Lift containers (provision + run).
// When forced, this will rebuild all images
// and recreate all containers.
func (containers Containers) lift(force bool) {
	containers.provision(force)
	containers.runOrStart(force)
}

// Provision containers.
// When forced, this will rebuild all images.
func (containers Containers) provision(force bool) {
	for _, container := range containers.reversed() {
		container.provision(force)
	}
}

// Run containers.
// When forced, removes existing containers first.
func (containers Containers) run(force bool) {
	if force {
		containers.rm(force)
	}
	for _, container := range containers.reversed() {
		container.run()
	}
}

// Run or start containers.
// When forced, removes existing containers first.
func (containers Containers) runOrStart(force bool) {
	if force {
		containers.rm(force)
	}
	for _, container := range containers.reversed() {
		container.runOrStart()
	}
}

// Start containers.
func (containers Containers) start() {
	for _, container := range containers.reversed() {
		container.start()
	}
}

// Kill containers.
func (containers Containers) kill() {
	for _, container := range containers {
		container.kill()
	}
}

// Stop containers.
func (containers Containers) stop() {
	for _, container := range containers {
		container.stop()
	}
}

// Remove containers.
// When forced, stops existing containers first.
func (containers Containers) rm(force bool) {
	if force {
		containers.stop()
	}
	for _, container := range containers {
		container.rm()
	}
}
