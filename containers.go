package main

import (
	"fmt"
	"os"
	"text/tabwriter"
)

type Containers []Container

func containerInList(container Container, list []string) bool {
	for _, listItem := range list {
		if os.ExpandEnv(listItem) == container.Name() {
			return true
		}
	}
	return false
}

func (containers Containers) filter(list []string) []Container {
	var filtered []Container
	for i := 0; i < len(containers); i++ {
		if containerInList(containers[i], list) {
			filtered = append(filtered, containers[i])
		}
	}
	return filtered
}

func (containers Containers) reversed() []Container {
	var reversed []Container
	for i := len(containers) - 1; i >= 0; i-- {
		reversed = append(reversed, containers[i])
	}
	return reversed
}

// Lift containers (provision + run).
// When rebuild set, this will rebuild all images
// and recreate all containers.
func (containers Containers) lift(nocache bool, kill bool, rebuild bool) {
	if rebuild {
		containers.provision(nocache, rebuild)
	}
	containers.runOrStart(kill)
}

// Provision containers.
func (containers Containers) provision(nocache bool, rebuild bool) {
	for _, container := range containers.reversed() {
		container.provision(nocache, rebuild)
	}
}

// Run containers.
// When killed, removes existing containers first.
func (containers Containers) run(kill bool, rebuild bool, nocache bool) {
	if kill {
		containers.rm(kill)
	}
	if rebuild {
		containers.provision(rebuild, nocache)
	}
	for _, container := range containers.reversed() {
		container.run()
	}
}

// Run or start containers.
// When Killed, removes existing containers first.
func (containers Containers) runOrStart(kill bool) {
	if kill {
		containers.rm(kill)
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
// When kill specified it will kill the container and then remove it.
func (containers Containers) rm(kill bool) {
	if kill {
		containers.kill()
	} else {
		containers.stop()
	}
	for _, container := range containers {
		container.rm()
	}
}

// Push containers.
func (containers Containers) push() {
	for _, container := range containers {
		container.push()
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
