package main

import (
	"bytes"
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
	if _, err := os.Stat("crane.json"); err == nil {
		return readCranefile("crane.json")
	}
	if _, err := os.Stat("Cranefile"); err == nil {
		printNotice("Using a Cranefile is deprecated. Please use crane.json instead.\n")
		return readCranefile("Cranefile")
	}
	panic("No crane.json found!")
}

func readCranefile(filename string) Containers {
	file, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(err)
	}

	return unmarshal(file)
}

// Thanks to https://github.com/markpeek/packer/commit/5bf33a0e91b2318a40c42e9bf855dcc8dd4cdec5
func displaySyntaxError(js []byte, syntaxError error) (err error) {
	syntax, ok := syntaxError.(*json.SyntaxError)
	if !ok {
		err = syntaxError
		return
	}
	newline := []byte{'\x0a'}
	space := []byte{' '}

	start, end := bytes.LastIndex(js[:syntax.Offset], newline)+1, len(js)
	if idx := bytes.Index(js[start:], newline); idx >= 0 {
		end = start + idx
	}

	line, pos := bytes.Count(js[:start], newline)+1, int(syntax.Offset)-start-1

	err = fmt.Errorf("\nError in line %d: %s \n%s\n%s^", line, syntaxError, js[start:end], bytes.Repeat(space, pos))
	return
}

func unmarshal(data []byte) Containers {
	var containers Containers
	err := json.Unmarshal(data, &containers)
	if err != nil {
		err = displaySyntaxError(data, err)
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
