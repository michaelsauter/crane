package crane

import (
	"fmt"
	"github.com/michaelsauter/crane/print"
	"os"
	"path"
	"strconv"
	"strings"
)

type Container struct {
	id            string
	RawName       string
	RawDockerfile string `json:"dockerfile" yaml:"dockerfile"`
	RawImage      string `json:"image" yaml:"image"`
	Run           RunParameters
	Rm            RmParameters
}

type RunParameters struct {
	RawCidfile     string      `json:"cidfile" yaml:"cidfile"`
	CpuShares      int         `json:"cpu-shares" yaml:"cpu-shares"`
	Detach         bool        `json:"detach" yaml:"detach"`
	RawDns         []string    `json:"dns" yaml:"dns"`
	RawEntrypoint  string      `json:"entrypoint" yaml:"entrypoint"`
	RawEnv         []string    `json:"env" yaml:"env"`
	RawEnvFile     string      `json:"env-file" yaml:"env-file"`
	RawExpose      []string    `json:"expose" yaml:"expose"`
	RawHostname    string      `json:"hostname" yaml:"hostname"`
	Interactive    bool        `json:"interactive" yaml:"interactive"`
	RawLink        []string    `json:"link" yaml:"link"`
	RawLxcConf     []string    `json:"lxc-conf" yaml:"lxc-conf"`
	RawMemory      string      `json:"memory" yaml:"memory"`
	RawNet         string      `json:"net" yaml:"net"`
	Privileged     bool        `json:"privileged" yaml:"privileged"`
	RawPublish     []string    `json:"publish" yaml:"publish"`
	PublishAll     bool        `json:"publish-all" yaml:"publish-all"`
	Rm             bool        `json:"rm" yaml:"rm"`
	Tty            bool        `json:"tty" yaml:"tty"`
	RawUser        string      `json:"user" yaml:"user"`
	RawVolume      []string    `json:"volume" yaml:"volume"`
	RawVolumesFrom []string    `json:"volumes-from" yaml:"volumes-from"`
	RawWorkdir     string      `json:"workdir" yaml:"workdir"`
	RawCmd         interface{} `json:"cmd" yaml:"cmd"`
}

type RmParameters struct {
	Volumes bool `json:"volumes" yaml:"volumes"`
}

func (container *Container) Dependencies() *Dependencies {
	var linkParts []string
	dependencies := &Dependencies{all: []string{}, link: []string{}, volumesFrom: []string{}, net: ""}
	for _, link := range container.Run.Link() {
		linkParts = strings.Split(link, ":")
		dependencies.all = append(dependencies.all, linkParts[0])
		dependencies.link = append(dependencies.link, linkParts[0])
	}
	for _, volumeFrom := range container.Run.VolumesFrom() {
		if !dependencies.includes(volumeFrom) {
			dependencies.all = append(dependencies.all, volumeFrom)
			dependencies.volumesFrom = append(dependencies.volumesFrom, volumeFrom)
		}
	}
	if netParts := strings.Split(container.Run.Net(), ":"); len(netParts) == 2 && netParts[0] == "container" {
		// We'll just assume here that the reference is a name, and not an id, even
		// though docker supports it, since we have no bullet-proof way to tell:
		// heuristics to detect whether it's an id could bring false positives, and
		// a lookup in the list of container names could bring false negatives
		dependencies.net = netParts[1]
		if !dependencies.includes(dependencies.net) {
			dependencies.all = append(dependencies.all, dependencies.net)
		}
	}
	return dependencies
}

func (container *Container) Name() string {
	return os.ExpandEnv(container.RawName)
}

func (container *Container) Dockerfile() string {
	return os.ExpandEnv(container.RawDockerfile)
}

func (container *Container) Image() string {
	return os.ExpandEnv(container.RawImage)
}

func (r *RunParameters) Cidfile() string {
	return os.ExpandEnv(r.RawCidfile)
}

func (r *RunParameters) Dns() []string {
	var dns []string
	for _, rawDns := range r.RawDns {
		dns = append(dns, os.ExpandEnv(rawDns))
	}
	return dns
}

func (r *RunParameters) Entrypoint() string {
	return os.ExpandEnv(r.RawEntrypoint)
}

func (r *RunParameters) Env() []string {
	var env []string
	for _, rawEnv := range r.RawEnv {
		env = append(env, os.ExpandEnv(rawEnv))
	}
	return env
}

func (r *RunParameters) EnvFile() string {
	return os.ExpandEnv(r.RawEnvFile)
}

func (r *RunParameters) Expose() []string {
	var expose []string
	for _, rawExpose := range r.RawExpose {
		expose = append(expose, os.ExpandEnv(rawExpose))
	}
	return expose
}

func (r *RunParameters) Hostname() string {
	return os.ExpandEnv(r.RawHostname)
}

func (r *RunParameters) Link() []string {
	var link []string
	for _, rawLink := range r.RawLink {
		link = append(link, os.ExpandEnv(rawLink))
	}
	return link
}

func (r *RunParameters) LxcConf() []string {
	var lxcConf []string
	for _, rawLxcConf := range r.RawLxcConf {
		lxcConf = append(lxcConf, os.ExpandEnv(rawLxcConf))
	}
	return lxcConf
}

func (r *RunParameters) Memory() string {
	return os.ExpandEnv(r.RawMemory)
}

func (r *RunParameters) Net() string {
	// Default to bridge
	if len(r.RawNet) == 0 {
		return "bridge"
	} else {
		return os.ExpandEnv(r.RawNet)
	}
}

func (r *RunParameters) Publish() []string {
	var publish []string
	for _, rawPublish := range r.RawPublish {
		publish = append(publish, os.ExpandEnv(rawPublish))
	}
	return publish
}

func (r *RunParameters) User() string {
	return os.ExpandEnv(r.RawUser)
}

func (r *RunParameters) Volume() []string {
	var volumes []string
	for _, rawVolume := range r.RawVolume {
		volume := os.ExpandEnv(rawVolume)
		paths := strings.Split(volume, ":")
		if !path.IsAbs(paths[0]) {
			cwd, _ := os.Getwd()
			paths[0] = cwd + "/" + paths[0]
		}
		volumes = append(volumes, strings.Join(paths, ":"))
	}
	return volumes
}

func (r *RunParameters) VolumesFrom() []string {
	var volumesFrom []string
	for _, rawVolumesFrom := range r.RawVolumesFrom {
		volumesFrom = append(volumesFrom, os.ExpandEnv(rawVolumesFrom))
	}
	return volumesFrom
}

func (r *RunParameters) Workdir() string {
	return os.ExpandEnv(r.RawWorkdir)
}

func (r *RunParameters) Cmd() []string {
	var cmd []string
	if r.RawCmd != nil {
		switch rawCmd := r.RawCmd.(type) {
		case string:
			if len(rawCmd) > 0 {
				cmd = append(cmd, os.ExpandEnv(rawCmd))
			}
		case []interface{}:
			cmds := make([]string, len(rawCmd))
			for i, v := range rawCmd {
				cmds[i] = os.ExpandEnv(v.(string))
			}
			cmd = append(cmd, cmds...)
		default:
			print.Errorf("cmd is of unknown type!")
		}
	}
	return cmd
}

func (container *Container) Id() (id string, err error) {
	if len(container.id) > 0 {
		id = container.id
	} else {
		// Inspect container, extracting the ID.
		// This will return gibberish if no container is found.
		args := []string{"inspect", "--format={{.Id}}", container.Name()}
		output, outErr := commandOutput("docker", args)
		if outErr == nil {
			id = output
			container.id = output
		} else {
			err = outErr
		}
	}
	return
}

func (container *Container) exists() bool {
	// `ps -a` returns all existant containers
	id, err := container.Id()
	if err != nil || len(id) == 0 {
		return false
	}
	dockerCmd := []string{"docker", "ps", "--quiet", "--all", "--no-trunc"}
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
	id, err := container.Id()
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

func (container *Container) paused() bool {
	args := []string{"inspect", "--format={{.State.Paused}}", container.Name()}
	output, err := commandOutput("docker", args)
	if err != nil {
		return false
	}
	paused, err := strconv.ParseBool(output)
	if err != nil {
		return false
	}
	return paused
}

func (container *Container) imageExists() bool {
	dockerCmd := []string{"docker", "images", "--no-trunc"}
	grepCmd := []string{"grep", "-wF", container.Image()}
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

func (container *Container) status() []string {
	fields := []string{container.Name(), container.Image(), "-", "-", "-", "-", "-"}
	args := []string{"inspect", "--format={{.Id}}\t{{.Image}}\t{{if .NetworkSettings.IPAddress}}{{.NetworkSettings.IPAddress}}{{else}}-{{end}}\t{{range $k,$v := $.NetworkSettings.Ports}}{{$k}},{{else}}-{{end}}\t{{.State.Running}}", container.Name()}
	output, err := commandOutput("docker", args)
	if err == nil {
		copy(fields[2:], strings.Split(output, "\t"))
		// we asked for the image id the container was created from
		fields[3] = strconv.FormatBool(imageIdFromTag(fields[1]) == fields[3])
	}
	return fields
}

// Pull image for container
func (container *Container) pullImage() {
	fmt.Printf("Pulling image %s ... ", container.Image())
	args := []string{"pull", container.Image()}
	executeCommand("docker", args)
}

// Build image for container
func (container *Container) buildImage(nocache bool) {
	fmt.Printf("Building image %s ... ", container.Image())
	args := []string{"build"}
	if nocache {
		args = append(args, "--no-cache")
	}
	args = append(args, "--rm", "--tag="+container.Image(), container.Dockerfile())
	executeCommand("docker", args)
}

func (container Container) provision(nocache bool) {
	if len(container.Dockerfile()) > 0 {
		container.buildImage(nocache)
	} else {
		container.pullImage()
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

// Provision or skip container
func (container Container) provisionOrSkip(update bool, nocache bool) {
	if update || !container.imageExists() {
		container.provision(nocache)
	}
}

// Run container
func (container Container) run() {
	if container.exists() {
		print.Noticef("Container %s does already exist. Use --recreate to recreate.\n", container.Name())
		if !container.running() {
			container.start()
		}
	} else {
		fmt.Printf("Running container %s ... ", container.Name())
		// Assemble command arguments
		args := []string{"run"}
		// Cidfile
		if len(container.Run.Cidfile()) > 0 {
			args = append(args, "--cidfile", container.Run.Cidfile())
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
		for _, dns := range container.Run.Dns() {
			args = append(args, "--dns", dns)
		}
		// Entrypoint
		if len(container.Run.Entrypoint()) > 0 {
			args = append(args, "--entrypoint", container.Run.Entrypoint())
		}
		// Env
		for _, env := range container.Run.Env() {
			args = append(args, "--env", env)
		}
		// Env file
		if len(container.Run.EnvFile()) > 0 {
			args = append(args, "--env-file", container.Run.EnvFile())
		}
		// Expose
		for _, expose := range container.Run.Expose() {
			args = append(args, "--expose", expose)
		}
		// Host
		if len(container.Run.Hostname()) > 0 {
			args = append(args, "--hostname", container.Run.Hostname())
		}
		// Interactive
		if container.Run.Interactive {
			args = append(args, "--interactive")
		}
		// Link
		for _, link := range container.Run.Link() {
			args = append(args, "--link", link)
		}
		// LxcConf
		for _, lxcConf := range container.Run.LxcConf() {
			args = append(args, "--lxc-conf", lxcConf)
		}
		// Memory
		if len(container.Run.Memory()) > 0 {
			args = append(args, "--memory", container.Run.Memory())
		}
		// Net
		if container.Run.Net() != "bridge" {
			args = append(args, "--net", container.Run.Net())
		}
		// Privileged
		if container.Run.Privileged {
			args = append(args, "--privileged")
		}
		// Publish
		for _, port := range container.Run.Publish() {
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
		if len(container.Run.User()) > 0 {
			args = append(args, "--user", container.Run.User())
		}
		// Volumes
		for _, volume := range container.Run.Volume() {
			args = append(args, "--volume", volume)
		}
		// VolumesFrom
		for _, volumeFrom := range container.Run.VolumesFrom() {
			args = append(args, "--volumes-from", volumeFrom)
		}
		// Workdir
		if len(container.Run.Workdir()) > 0 {
			args = append(args, "--workdir", container.Run.Workdir())
		}
		// Name
		args = append(args, "--name", container.Name())
		// Image
		args = append(args, container.Image())
		// Command
		args = append(args, container.Run.Cmd()...)
		// Execute command
		executeCommand("docker", args)
	}
}

// Start container
func (container Container) start() {
	if container.exists() {
		if !container.running() {
			fmt.Printf("Starting container %s ... ", container.Name())
			args := []string{"start", container.Name()}
			executeCommand("docker", args)
		}
	} else {
		print.Errorf("Container %s does not exist.\n", container.Name())
	}
}

// Kill container
func (container Container) kill() {
	if container.running() {
		fmt.Printf("Killing container %s ... ", container.Name())
		args := []string{"kill", container.Name()}
		executeCommand("docker", args)
	}
}

// Stop container
func (container Container) stop() {
	if container.running() {
		fmt.Printf("Stopping container %s ... ", container.Name())
		args := []string{"stop", container.Name()}
		executeCommand("docker", args)
	}
}

// Pause container
func (container Container) pause() {
	if container.running() {
		if container.paused() {
			print.Noticef("Container %s is already paused.\n", container.Name())
		} else {
			fmt.Printf("Pausing container %s ... ", container.Name())
			args := []string{"pause", container.Name()}
			executeCommand("docker", args)
		}
	} else {
		print.Noticef("Container %s is not running.\n", container.Name())
	}
}

// Unpause container
func (container Container) unpause() {
	if container.paused() {
		fmt.Printf("Unpausing container %s ... ", container.Name())
		args := []string{"unpause", container.Name()}
		executeCommand("docker", args)
	}
}

// Remove container
func (container Container) rm() {
	if container.exists() {
		if container.running() {
			print.Errorf("Container %s is running and cannot be removed.\n", container.Name())
		} else {
			args := []string{"rm"}
			if container.Rm.Volumes {
				fmt.Printf("Removing container %s and its volumes ... ", container.Name())
				args = append(args, "--volumes")
			} else {
				fmt.Printf("Removing container %s ... ", container.Name())
			}
			args = append(args, container.Name())
			executeCommand("docker", args)
		}
	}
}

// Push container
func (container Container) push() {
	if len(container.Image()) > 0 {
		fmt.Printf("Pushing image %s ... ", container.Image())
		args := []string{"push", container.Image()}
		executeCommand("docker", args)
	} else {
		print.Noticef("Skipping %s as it does not have an image name.\n", container.Name())
	}
}

// Return the image id of a tag, or an empty string if it doesn't exist
func imageIdFromTag(tag string) string {
	args := []string{"inspect", "--format={{.Id}}", tag}
	output, err := commandOutput("docker", args)
	if err != nil {
		return ""
	}
	return string(output)
}
