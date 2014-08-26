package crane

import (
	"fmt"
	"github.com/michaelsauter/crane/print"
	"os"
	"path"
	"strconv"
	"strings"
)

type Container interface {
	Name() string
	Dockerfile() string
	Image() string
	Id() string
	Dependencies() *Dependencies
	Exists() bool
	Running() bool
	Paused() bool
	ImageExists() bool
	Status() []string
	Provision(nocache bool)
	ProvisionOrSkip(update bool, nocache bool)
	Run()
	Start()
	RunOrStart()
	Kill()
	Stop()
	Pause()
	Unpause()
	Rm(force bool)
	Push()
}

type container struct {
	id            string
	RawName       string
	RawDockerfile string          `json:"dockerfile" yaml:"dockerfile"`
	RawImage      string          `json:"image" yaml:"image"`
	RunParams     RunParameters   `json:"run" yaml:"run"`
	RmParams      RmParameters    `json:"rm" yaml:"rm"`
	StartParams   StartParameters `json:"start" yaml:"start"`
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

type StartParameters struct {
	Attach      bool `json:"attach" yaml:"attach"`
	Interactive bool `json:"interactive" yaml:"interactive"`
}

func (c *container) Dependencies() *Dependencies {
	var linkParts []string
	dependencies := &Dependencies{all: []string{}, link: []string{}, volumesFrom: []string{}, net: ""}
	for _, link := range c.RunParams.Link() {
		linkParts = strings.Split(link, ":")
		dependencies.all = append(dependencies.all, linkParts[0])
		dependencies.link = append(dependencies.link, linkParts[0])
	}
	for _, volumeFrom := range c.RunParams.VolumesFrom() {
		if !dependencies.includes(volumeFrom) {
			dependencies.all = append(dependencies.all, volumeFrom)
			dependencies.volumesFrom = append(dependencies.volumesFrom, volumeFrom)
		}
	}
	if netParts := strings.Split(c.RunParams.Net(), ":"); len(netParts) == 2 && netParts[0] == "container" {
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

func (c *container) Name() string {
	return os.ExpandEnv(c.RawName)
}

func (c *container) Dockerfile() string {
	return os.ExpandEnv(c.RawDockerfile)
}

func (c *container) Image() string {
	return os.ExpandEnv(c.RawImage)
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

func (c *container) Id() string {
	if len(c.id) == 0 {
		// `docker inspect` works both for image or containers, make sure this is a
		// container payload we get back, otherwise we might end up getting the Id
		// of the image of the same name
		c.id = inspectString(c.Name(), "{{if .State}}{{.Id}}{{else}}{{end}}")
	}
	return c.id
}

func (c *container) Exists() bool {
	return c.Id() != ""
}

func (c *container) Running() bool {
	if !c.Exists() {
		return false
	}
	return inspectBool(c.Id(), "{{.State.Running}}")
}

func (c *container) Paused() bool {
	if !c.Exists() {
		return false
	}
	return inspectBool(c.Id(), "{{.State.Paused}}")
}

func (c *container) ImageExists() bool {
	dockerCmd := []string{"docker", "images", "--no-trunc"}
	grepCmd := []string{"grep", "-wF", c.Image()}
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

func (c *container) Status() []string {
	fields := []string{c.Name(), c.Image(), "-", "-", "-", "-", "-"}
	output := inspectString(c.Id(), "{{.Id}}\t{{.Image}}\t{{if .NetworkSettings.IPAddress}}{{.NetworkSettings.IPAddress}}{{else}}-{{end}}\t{{range $k,$v := $.NetworkSettings.Ports}}{{$k}},{{else}}-{{end}}\t{{.State.Running}}")
	if output != "" {
		copy(fields[2:], strings.Split(output, "\t"))
		// we asked for the image id the container was created from
		fields[3] = strconv.FormatBool(imageIdFromTag(fields[1]) == fields[3])
	}
	return fields
}

func (c *container) Provision(nocache bool) {
	if len(c.Dockerfile()) > 0 {
		c.buildImage(nocache)
	} else {
		c.pullImage()
	}
}

// Run or start container
func (c *container) RunOrStart() {
	if c.Exists() {
		c.Start()
	} else {
		c.Run()
	}
}

// Provision or skip container
func (c *container) ProvisionOrSkip(update bool, nocache bool) {
	if update || !c.ImageExists() {
		c.Provision(nocache)
	}
}

// Run container
func (c *container) Run() {
	if c.Exists() {
		print.Noticef("Container %s does already exist. Use --recreate to recreate.\n", c.Name())
		if !c.Running() {
			c.Start()
		}
	} else {
		fmt.Printf("Running container %s ... ", c.Name())
		// Assemble command arguments
		args := []string{"run"}
		// Cidfile
		if len(c.RunParams.Cidfile()) > 0 {
			args = append(args, "--cidfile", c.RunParams.Cidfile())
		}
		// CPU shares
		if c.RunParams.CpuShares > 0 {
			args = append(args, "--cpu-shares", strconv.Itoa(c.RunParams.CpuShares))
		}
		// Detach
		if c.RunParams.Detach {
			args = append(args, "--detach")
		}
		// Dns
		for _, dns := range c.RunParams.Dns() {
			args = append(args, "--dns", dns)
		}
		// Entrypoint
		if len(c.RunParams.Entrypoint()) > 0 {
			args = append(args, "--entrypoint", c.RunParams.Entrypoint())
		}
		// Env
		for _, env := range c.RunParams.Env() {
			args = append(args, "--env", env)
		}
		// Env file
		if len(c.RunParams.EnvFile()) > 0 {
			args = append(args, "--env-file", c.RunParams.EnvFile())
		}
		// Expose
		for _, expose := range c.RunParams.Expose() {
			args = append(args, "--expose", expose)
		}
		// Host
		if len(c.RunParams.Hostname()) > 0 {
			args = append(args, "--hostname", c.RunParams.Hostname())
		}
		// Interactive
		if c.RunParams.Interactive {
			args = append(args, "--interactive")
		}
		// Link
		for _, link := range c.RunParams.Link() {
			args = append(args, "--link", link)
		}
		// LxcConf
		for _, lxcConf := range c.RunParams.LxcConf() {
			args = append(args, "--lxc-conf", lxcConf)
		}
		// Memory
		if len(c.RunParams.Memory()) > 0 {
			args = append(args, "--memory", c.RunParams.Memory())
		}
		// Net
		if c.RunParams.Net() != "bridge" {
			args = append(args, "--net", c.RunParams.Net())
		}
		// Privileged
		if c.RunParams.Privileged {
			args = append(args, "--privileged")
		}
		// Publish
		for _, port := range c.RunParams.Publish() {
			args = append(args, "--publish", port)
		}
		// PublishAll
		if c.RunParams.PublishAll {
			args = append(args, "--publish-all")
		}
		// Rm
		if c.RunParams.Rm {
			args = append(args, "--rm")
		}
		// Tty
		if c.RunParams.Tty {
			args = append(args, "--tty")
		}
		// User
		if len(c.RunParams.User()) > 0 {
			args = append(args, "--user", c.RunParams.User())
		}
		// Volumes
		for _, volume := range c.RunParams.Volume() {
			args = append(args, "--volume", volume)
		}
		// VolumesFrom
		for _, volumeFrom := range c.RunParams.VolumesFrom() {
			args = append(args, "--volumes-from", volumeFrom)
		}
		// Workdir
		if len(c.RunParams.Workdir()) > 0 {
			args = append(args, "--workdir", c.RunParams.Workdir())
		}
		// Name
		args = append(args, "--name", c.Name())
		// Image
		args = append(args, c.Image())
		// Command
		args = append(args, c.RunParams.Cmd()...)
		// Execute command
		executeCommand("docker", args)
	}
}

// Start container
func (c *container) Start() {
	if c.Exists() {
		if !c.Running() {
			fmt.Printf("Starting container %s ... ", c.Name())
			args := []string{"start"}
			if c.StartParams.Attach {
				args = append(args, "--attach")
			}
			if c.StartParams.Interactive {
				args = append(args, "--interactive")
			}
			args = append(args, c.Name())
			executeCommand("docker", args)
		}
	} else {
		print.Errorf("Container %s does not exist.\n", c.Name())
	}
}

// Kill container
func (c *container) Kill() {
	if c.Running() {
		fmt.Printf("Killing container %s ... ", c.Name())
		args := []string{"kill", c.Name()}
		executeCommand("docker", args)
	}
}

// Stop container
func (c *container) Stop() {
	if c.Running() {
		fmt.Printf("Stopping container %s ... ", c.Name())
		args := []string{"stop", c.Name()}
		executeCommand("docker", args)
	}
}

// Pause container
func (c *container) Pause() {
	if c.Running() {
		if c.Paused() {
			print.Noticef("Container %s is already paused.\n", c.Name())
		} else {
			fmt.Printf("Pausing container %s ... ", c.Name())
			args := []string{"pause", c.Name()}
			executeCommand("docker", args)
		}
	} else {
		print.Noticef("Container %s is not running.\n", c.Name())
	}
}

// Unpause container
func (c *container) Unpause() {
	if c.Paused() {
		fmt.Printf("Unpausing container %s ... ", c.Name())
		args := []string{"unpause", c.Name()}
		executeCommand("docker", args)
	}
}

// Remove container
func (c *container) Rm(force bool) {
	if c.Exists() {
		if !force && c.Running() {
			print.Errorf("Container %s is running and cannot be removed. Use --force to remove anyway.\n", c.Name())
		} else {
			args := []string{"rm"}
			if force {
				args = append(args, "--force")
			}
			if c.RmParams.Volumes {
				fmt.Printf("Removing container %s and its volumes ... ", c.Name())
				args = append(args, "--volumes")
			} else {
				fmt.Printf("Removing container %s ... ", c.Name())
			}
			args = append(args, c.Name())
			executeCommand("docker", args)
			c.id = ""
		}
	}
}

// Push container
func (c *container) Push() {
	if len(c.Image()) > 0 {
		fmt.Printf("Pushing image %s ... ", c.Image())
		args := []string{"push", c.Image()}
		executeCommand("docker", args)
	} else {
		print.Noticef("Skipping %s as it does not have an image name.\n", c.Name())
	}
}

// Pull image for container
func (c *container) pullImage() {
	fmt.Printf("Pulling image %s ... ", c.Image())
	args := []string{"pull", c.Image()}
	executeCommand("docker", args)
}

// Build image for container
func (c *container) buildImage(nocache bool) {
	fmt.Printf("Building image %s ... ", c.Image())
	args := []string{"build"}
	if nocache {
		args = append(args, "--no-cache")
	}
	args = append(args, "--rm", "--tag="+c.Image(), c.Dockerfile())
	executeCommand("docker", args)
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

// Attempt to parse the value referenced by the go template
// for the `docker inspect` as a boolean, fallbacking to
// false on error
func inspectBool(container string, format string) bool {
	output := inspectString(container, format)
	flag, _ := strconv.ParseBool(output)
	return flag
}

// Returns the value referenced by the go template for
// the `docker inspect` as a string, fallbacking to
// an empty string on error
func inspectString(container string, format string) string {
	args := []string{"inspect", "--format=" + format, container}
	output, err := commandOutput("docker", args)
	if err != nil {
		return ""
	}
	return output
}
