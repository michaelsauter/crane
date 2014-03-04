package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"text/tabwriter"
)

type Containers []Container

func getContainers(config string) Containers {
	if len(config) > 0 {
		return unmarshal([]byte(config))
	}
	return readCranefile("Cranefile")
}

func readCranefile(filename string) Containers {
	file, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(err)
	}

	return unmarshal(file)
}

func unmarshal(data []byte) Containers {
	var containers Containers
	err := json.Unmarshal(data, &containers)
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
func (containers Containers) lift(force bool, kill bool) {
	containers.provision(force)
	containers.runOrStart(force, kill)
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
func (containers Containers) run(force bool, kill bool) {
	if force {
		containers.rm(force, kill)
	}
	for _, container := range containers.reversed() {
		container.run()
	}
}

// Run or start containers.
// When forced, removes existing containers first.
func (containers Containers) runOrStart(force bool, kill bool) {
	if force {
		containers.rm(force, kill)
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
func (containers Containers) rm(force bool, kill bool) {
	if force {
		if kill {
			containers.kill()
		} else {
			containers.stop()
		}
	}
	for _, container := range containers {
		container.rm()
	}
}

// Status of containers.
func (containers Containers) status() {
	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 0, 8, 1, '\t', 0)
	fmt.Fprintln(w, "Name\tRunning\tID\tIP\tPorts")
	for _, container := range containers {
		container.status(w)
	}
	w.Flush()
}
