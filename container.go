package main

import (
	"github.com/michaelsauter/crane/print"
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"
	"text/tabwriter"
)

type Container struct {
	Id         string
	Name       string `json:"name" yaml:"name"`
	Dockerfile string `json:"dockerfile" yaml:"dockerfile"`
	Image      string `json:"image" yaml:"image"`
	Run        RunParameters
}

type RunParameters struct {
	Cidfile     string      `json:"cidfile" yaml:"cidfile"`
	CpuShares   int         `json:"cpu-shares" yaml:"cpu-shares"`
	Detach      bool        `json:"detach" yaml:"detach"`
	Dns         []string    `json:"dns" yaml:"dns"`
	Entrypoint  string      `json:"entrypoint" yaml:"entrypoint"`
	Env         []string    `json:"env" yaml:"env"`
	Expose      []string    `json:"expose" yaml:"expose"`
	Host        string      `json:"host" yaml:"host"`
	Interactive bool        `json:"interactive" yaml:"interactive"`
	Link        []string    `json:"link" yaml:"link"`
	LxcConf     []string    `json:"lxc-conf" yaml:"lxc-conf"`
	Memory      string      `json:"memory" yaml:"memory"`
	Privileged  bool        `json:"privileged" yaml:"privileged"`
	Publish     []string    `json:"publish" yaml:"publish"`
	PublishAll  bool        `json:"publish-all" yaml:"publish-all"`
	Rm          bool        `json:"rm" yaml:"rm"`
	Tty         bool        `json:"tty" yaml:"tty"`
	User        string      `json:"user" yaml:"user"`
	Volume      []string    `json:"volume" yaml:"volume"`
	VolumesFrom []string    `json:"volumes-from" yaml:"volumes-from"`
	Workdir     string      `json:"workdir" yaml:"workdir"`
	Command     interface{} `json:"cmd" yaml:"cmd"`
}

func (container *Container) getId() (id string, err error) {
	if len(container.Id) > 0 {
		id = container.Id
	} else {
		// Inspect container, extracting the ID.
		// This will return gibberish if no container is found.
		args := []string{"inspect", "--format={{.ID}}", container.Name}
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
	dockerCmd := []string{"docker", "ps", "--quit", "--all", "--no-trunc"}
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
	dockerCmd := []string{"docker", "ps", "--quiet", "--no-trunc"}
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
	dockerCmd := []string{"docker", "images", "--no-trunc"}
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

func (container *Container) status(w *tabwriter.Writer) {
	args := []string{"inspect", "--format={{.State.Running}}\t{{.ID}}\t{{if .NetworkSettings.IPAddress}}{{.NetworkSettings.IPAddress}}{{else}}-{{end}}\t{{range $k,$v := $.NetworkSettings.Ports}}{{$k}},{{end}}", container.Name}
	output, err := commandOutput("docker", args)
	if err != nil {
		fmt.Fprintf(w, "%s\tError:%v\t%v\n", container.Name, err, output)
		return
	}
	fmt.Fprintf(w, "%s\t%s\n", container.Name, output)
}

// Pull image for container
func (container *Container) pullImage() {
	fmt.Printf("Pulling image %s ... ", container.Image)
	args := []string{"pull", container.Image}
	executeCommand("docker", args)
}

// Build image for container
func (container *Container) buildImage() {
	fmt.Printf("Building image %s ... ", container.Image)
	args := []string{"build", "--rm", "--tag=" + container.Image, os.ExpandEnv(container.Dockerfile)}
	executeCommand("docker", args)
}

func (container Container) provision(force bool) {
	if force || !container.imageExists() {
		if len(container.Dockerfile) > 0 {
			container.buildImage()
		} else {
			container.pullImage()
		}
	} else {
		print.Notice("Image %s does already exist. Use --force to recreate.\n", container.Image)
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
		print.Notice("Container %s does already exist. Use --force to recreate.\n", container.Name)
		if !container.running() {
			container.start()
		}
	} else {
		fmt.Printf("Running container %s ... ", container.Name)
		// Assemble command arguments
		args := []string{"run"}
		// Cidfile
		if len(container.Run.Cidfile) > 0 {
			args = append(args, "--cidfile", os.ExpandEnv(container.Run.Cidfile))
		}
		// CPU shares
		if container.Run.CpuShares > 0 {
			args = append(args, "--cpu-shares", os.ExpandEnv(strconv.Itoa(container.Run.CpuShares)))
		}
		// Detach
		if container.Run.Detach {
			args = append(args, "--detach")
		}
		// Dns
		for _, dns := range container.Run.Dns {
			args = append(args, "--dns", os.ExpandEnv(dns))
		}
		// Entrypoint
		if len(container.Run.Entrypoint) > 0 {
			args = append(args, "--entrypoint", os.ExpandEnv(container.Run.Entrypoint))
		}
		// Env
		for _, env := range container.Run.Env {
			args = append(args, "--env", os.ExpandEnv(env))
		}
		// Expose
		for _, expose := range container.Run.Expose {
			args = append(args, "--expose", os.ExpandEnv(expose))
		}
		// Host
		if len(container.Run.Host) > 0 {
			args = append(args, "--hostname", os.ExpandEnv(container.Run.Host))
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
			args = append(args, "--lxc-conf", os.ExpandEnv(lxcConf))
		}
		// Memory
		if len(container.Run.Memory) > 0 {
			args = append(args, "--memory", os.ExpandEnv(container.Run.Memory))
		}
		// Privileged
		if container.Run.Privileged {
			args = append(args, "--privileged")
		}
		// Publish
		for _, port := range container.Run.Publish {
			args = append(args, "--publish", os.ExpandEnv(port))
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
			args = append(args, "--user", os.ExpandEnv(container.Run.User))
		}
		// Volumes
		for _, volume := range container.Run.Volume {
			paths := strings.Split(volume, ":")
			if !path.IsAbs(paths[0]) {
				cwd, _ := os.Getwd()
				paths[0] = cwd + "/" + paths[0]
			}
			args = append(args, "--volume", os.ExpandEnv(strings.Join(paths, ":")))
		}
		// VolumesFrom
		for _, volumeFrom := range container.Run.VolumesFrom {
			args = append(args, "--volumes-from", os.ExpandEnv(volumeFrom))
		}
		// Workdir
		if len(container.Run.Workdir) > 0 {
			args = append(args, "--workdir", os.ExpandEnv(container.Run.Workdir))
		}

		// Name
		args = append(args, "--name", container.Name)
		// Image
		args = append(args, container.Image)
		// Command
		if container.Run.Command != nil {
			switch cmd := container.Run.Command.(type) {
			case string:
				if len(cmd) > 0 {
					args = append(args, cmd)
				}
			case []interface{}:
				cmds := make([]string, len(cmd))
				for i, v := range cmd {
					cmds[i] = v.(string)
				}
				args = append(args, cmds...)
			default:
				print.Error("cmd is of unknown type!")
			}
		}
		// Execute command
		executeCommand("docker", args)
	}
}

// Start container
func (container Container) start() {
	if container.exists() {
		if !container.running() {
			fmt.Printf("Starting container %s ... ", container.Name)
			args := []string{"start", container.Name}
			executeCommand("docker", args)
		}
	} else {
		print.Error("Container %s does not exist.\n", container.Name)
	}
}

// Kill container
func (container Container) kill() {
	if container.running() {
		fmt.Printf("Killing container %s ... ", container.Name)
		args := []string{"kill", container.Name}
		executeCommand("docker", args)
	}
}

// Stop container
func (container Container) stop() {
	if container.running() {
		fmt.Printf("Stopping container %s ... ", container.Name)
		args := []string{"stop", container.Name}
		executeCommand("docker", args)
	}
}

// Remove container
func (container Container) rm() {
	if container.exists() {
		if container.running() {
			print.Error("Container %s is running and cannot be removed.\n", container.Name)
		} else {
			fmt.Printf("Removing container %s ... ", container.Name)
			args := []string{"rm", container.Name}
			executeCommand("docker", args)
		}
	}
}
