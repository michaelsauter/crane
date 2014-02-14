package main

import (
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"
)

type Container struct {
	Id         string
	Name       string `json:"name"`
	Dockerfile string `json:"dockerfile"`
	Image      string `json:"image"`
	Run        RunParameters
}

type RunParameters struct {
	Cidfile     string   `json:"cidfile"`
	CpuShares   int      `json:"cpu-shares"`
	Detach      bool     `json:"detach"`
	Dns         []string `json:"dns"`
	Entrypoint  string   `json:"entrypoint"`
	Env         []string `json:"env"`
	Expose      []string `json:"expose"`
	Host        string   `json:"host"`
	Interactive bool     `json:"interactive"`
	Link        []string `json:"link"`
	LxcConf     []string `json:"lxc-conf"`
	Memory      string   `json:"memory"`
	Privileged  bool     `json:"privileged"`
	Publish     []string `json:"publish"`
	PublishAll  bool     `json:"publish-all"`
	Rm          bool     `json:"rm"`
	Tty         bool     `json:"tty"`
	User        string   `json:"user"`
	Volume      []string `json:"volume"`
	VolumesFrom []string `json:"volumes-from"`
	Workdir     string   `json:"workdir"`
	Command     string   `json:"cmd"`
}

func (container *Container) getId() (id string, err error) {
	if len(container.Id) > 0 {
		id = container.Id
	} else {
		// Inspect container, extracting the ID.
		// This will return gibberish if no container is found.
		args := []string{"inspect", "-format={{.ID}}", container.Name}
		output, outErr := commandOutput("docker", args)
		if err == nil {
			id = output
			container.Id = output
		} else {
			err = outErr
		}
	}
	return
}

func (container *Container) exists() bool {
	// `ps -a` returns all existant containers
	id, err := container.getId()
	if err != nil || len(id) == 0 {
		return false
	}
	dockerCmd := []string{"docker", "ps", "-q", "-a", "-notrunc"}
	grepCmd := []string{"grep", "-wF", id}
	output, err := pipedCommandOutput(dockerCmd, grepCmd)
	if err != nil {
		return false
	}
	result := string(output)
	if len(result) > 0 {
		return true
	} else {
		return false
	}
}

func (container *Container) running() bool {
	// `ps` returns all running containers
	id, err := container.getId()
	if err != nil || len(id) == 0 {
		return false
	}
	dockerCmd := []string{"docker", "ps", "-q", "-notrunc"}
	grepCmd := []string{"grep", "-wF", id}
	output, err := pipedCommandOutput(dockerCmd, grepCmd)
	if err != nil {
		return false
	}
	result := string(output)
	if len(result) > 0 {
		return true
	} else {
		return false
	}
}

func (container *Container) imageExists() bool {
	dockerCmd := []string{"docker", "images", "-notrunc"}
	grepCmd := []string{"grep", "-wF", container.Image}
	output, err := pipedCommandOutput(dockerCmd, grepCmd)
	if err != nil {
		return false
	}
	result := string(output)
	if len(result) > 0 {
		return true
	} else {
		return false
	}
}

// Pull image for container
func (container *Container) pullImage() {
	args := []string{"pull", container.Image}
	executeCommand("docker", args)
}

// Build image for container
func (container *Container) buildImage() {
	args := []string{"build", "-rm", "-t=" + container.Image, container.Dockerfile}
	executeCommand("docker", args)
}

func (container Container) provision(force bool) {
	if force || !container.imageExists() {
		if len(container.Dockerfile) > 0 {
			container.buildImage()
		} else {
			container.pullImage()
		}
	}
}

// Run or start container
func (container Container) runOrStart() {
	if container.exists() {
		container.start()
	} else {
		container.run()
	}
}

// Run container
func (container Container) run() {
	if container.exists() {
		fmt.Printf(" ! %s does already exist. Use --force to recreate.\n", container.Name)
	} else {
		// Assemble command arguments
		args := []string{"run"}
		// Cidfile
		if len(container.Run.Cidfile) > 0 {
			args = append(args, "--cidfile", container.Run.Cidfile)
		}
		// CPU shares
		if container.Run.CpuShares > 0 {
			args = append(args, "--cpu-shares", strconv.Itoa(container.Run.CpuShares))
		}
		// Detach
		if container.Run.Detach {
			args = append(args, "--detach")
		}
		// Dns
		for _, dns := range container.Run.Dns {
			args = append(args, "--dns", dns)
		}
		// Entrypoint
		if len(container.Run.Entrypoint) > 0 {
			args = append(args, "--workdir", container.Run.Entrypoint)
		}
		// Env
		for _, env := range container.Run.Env {
			args = append(args, "--env", env)
		}
		// Expose
		for _, expose := range container.Run.Expose {
			args = append(args, "--expose", expose)
		}
		// Host
		if len(container.Run.Host) > 0 {
			args = append(args, "--host", container.Run.Host)
		}
		// Interactive
		if container.Run.Interactive {
			args = append(args, "--interactive")
		}
		// Link
		for _, link := range container.Run.Link {
			args = append(args, "--link", link)
		}
		// LxcConf
		for _, lxcConf := range container.Run.LxcConf {
			args = append(args, "--lxc-conf", lxcConf)
		}
		// Memory
		if len(container.Run.Memory) > 0 {
			args = append(args, "--memory", container.Run.Memory)
		}
		// Privileged
		if container.Run.Privileged {
			args = append(args, "--privileged")
		}
		// Publish
		for _, port := range container.Run.Publish {
			args = append(args, "--publish", port)
		}
		// PublishAll
		if container.Run.PublishAll {
			args = append(args, "--publish-all")
		}
		// Rm
		if container.Run.Rm {
			args = append(args, "--rm")
		}
		// Tty
		if container.Run.Tty {
			args = append(args, "--tty")
		}
		// User
		if len(container.Run.User) > 0 {
			args = append(args, "--user", container.Run.User)
		}
		// Volumes
		for _, volume := range container.Run.Volume {
			paths := strings.Split(volume, ":")
			if !path.IsAbs(paths[0]) {
				cwd, _ := os.Getwd()
				paths[0] = cwd + "/" + paths[0]
			}
			args = append(args, "--volume", strings.Join(paths, ":"))
		}
		// VolumesFrom
		for _, volumeFrom := range container.Run.VolumesFrom {
			args = append(args, "--volumes-from", volumeFrom)
		}
		// Workdir
		if len(container.Run.Workdir) > 0 {
			args = append(args, "--workdir", container.Run.Workdir)
		}

		// Name
		args = append(args, "--name", container.Name)
		// Image
		args = append(args, container.Image)
		// Command
		if len(container.Run.Command) > 0 {
			args = append(args, container.Run.Command)
		}
		// Execute command
		executeCommand("docker", args)
	}
}

// Start container
func (container Container) start() {
	if container.exists() {
		if container.running() {
			fmt.Printf(" i %s skipped as container is already running.\n", container.Name)
		} else {
			args := []string{"start", container.Name}
			executeCommand("docker", args)
		}
	} else {
		fmt.Printf(" ! %s does not exist.\n", container.Name)
	}
}

// Kill container
func (container Container) kill() {
	if container.running() {
		args := []string{"kill", container.Name}
		executeCommand("docker", args)
	} else {
		fmt.Printf(" i %s skipped as container is not running.\n", container.Name)
	}
}

// Stop container
func (container Container) stop() {
	if container.running() {
		args := []string{"stop", container.Name}
		executeCommand("docker", args)
	} else {
		fmt.Printf(" i %s skipped as container is not running.\n", container.Name)
	}
}

// Remove container
func (container Container) rm() {
	if container.exists() {
		if container.running() {
			fmt.Printf(" ! Can't remove running container %s.\n", container.Name)
		} else {
			args := []string{"rm", container.Name}
			executeCommand("docker", args)
		}
	} else {
		fmt.Printf(" i %s skipped as container does not exist.\n", container.Name)
	}
}
