package main

import (
	"encoding/json"
	"io/ioutil"
	"strings"
	"path"
	"os"
)

type Container struct {
	Name         string
	Dockerfile   string
	Image        string
	Parameters   Parameters
	Dependencies map[string]Container
}

type Parameters struct {
	Ports        []string `json:"p"`
	Volumes      []string `json:"v"`
	Environments []string `json:"e"`
	Tty          bool     `json:"t"`
	Interactive  bool     `json:"i"`
	Command      string   `json:"cmd"`
}

func readCranefile(filename string) Container {
	file, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(err)
	}

	var container Container
	err = json.Unmarshal(file, &container)
	if err != nil {
		panic(err)
	}
	return container
}

func (container *Container) provision() {
	if len(container.Dockerfile) > 0 {
		container.build()
	} else {
		container.pull()
	}
}

// Pull container and provision dependent containers
func (container *Container) pull() {
	// Provision the dependencies
	for _, dependency := range container.Dependencies {
		dependency.provision()
	}
	// Pull image for container
	args := []string{"pull", container.Image}
	executeCommand("docker", args)
}

// Build container and provision dependent containers
func (container *Container) build() {
	// Provision the dependencies
	for _, dependency := range container.Dependencies {
		dependency.provision()
	}
	// Build image for container
	args := []string{"build", "-rm", "-t=" + container.Image, container.Dockerfile}
	executeCommand("docker", args)
}

// Run container and dependent containers
func (container *Container) run() {
	var links map[string]string
	links = make(map[string]string)
	// Run the dependencies
	for key, dependency := range container.Dependencies {
		dependency.run()
		links[key] = dependency.Name
	}
	// Run this container
	// Assemble command arguments
	args := []string{"run", "-d"}
	// Volumes
	for _, volume := range container.Parameters.Volumes {
		paths := strings.Split(volume, ":")
		if !path.IsAbs(paths[0]) {
			cwd, _ := os.Getwd()
			paths[0] = cwd + "/" + paths[0]
		}
		args = append(args, "-v", strings.Join(paths, ":"))
	}
	// Ports
	for _, port := range container.Parameters.Ports {
		args = append(args, "-p", port)
	}
	// Environment Variables
	for _, environment := range container.Parameters.Environments {
		args = append(args, "-e", environment)
	}
	// Interactive
	if container.Parameters.Interactive {
		args = append(args, "-i")
	}
	// Tty
	if container.Parameters.Tty {
		args = append(args, "-t")
	}
	// Links
	for key, name := range links {
		args = append(args, "-link", name+":"+key)
	}
	// Name
	args = append(args, "-name", container.Name)
	// Image
	args = append(args, container.Image)
	// Command
	if len(container.Parameters.Command) > 0 {
		args = append(args, container.Parameters.Command)
	}
	// Execute command
	executeCommand("docker", args)
}

// Kill container and dependent containers
func (container *Container) kill() {
	// Kill this container
	args := []string{"kill", container.Name}
	executeCommand("docker", args)
	// Kill the dependencies
	for _, dependency := range container.Dependencies {
		dependency.kill()
	}
}

// Start container and dependent containers
func (container *Container) start() {
	// Start the dependencies
	for _, dependency := range container.Dependencies {
		dependency.start()
	}
	// Start this container
	args := []string{"start", container.Name}
	executeCommand("docker", args)
}

// Stop container and dependent containers
func (container *Container) stop() {
	// Stop this container
	args := []string{"stop", container.Name}
	executeCommand("docker", args)
	// Stop the dependencies
	for _, dependency := range container.Dependencies {
		dependency.stop()
	}
}

// Remove container and dependent containers
func (container *Container) rm() {
	// Remove this container
	args := []string{"rm", container.Name}
	executeCommand("docker", args)
	// Remove the dependencies
	for _, dependency := range container.Dependencies {
		dependency.rm()
	}
}
