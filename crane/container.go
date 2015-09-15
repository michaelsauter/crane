package crane

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/flynn/go-shlex"
	"io"
	"os"
	"path"
	"strconv"
	"strings"
)

type Container interface {
	Name() string
	ActualName() string
	BuildContext() string
	Image() string
	ImageWithTag() string
	Id() string
	Dependencies(excluded []string) *Dependencies
	Exists() bool
	Running() bool
	Unique() bool
	Paused() bool
	ImageExists() bool
	Status() []string
	Lift(cmds []string, nocache bool, excluded []string, configPath string)
	Provision(nocache bool)
	PullImage()
	Create(cmds []string, excluded []string, configPath string)
	Run(cmds []string, excluded []string, configPath string)
	Start(excluded []string, configPath string)
	Kill()
	Stop()
	Pause()
	Unpause()
	Exec(cmds []string, configPath string)
	Rm(force bool)
	Logs(follow bool, since string, tail string) (stdout, stderr io.Reader)
	Push()
	Hooks() Hooks
}

type container struct {
	id            string
	RawName       string
	RawUnique     bool            `json:"unique" yaml:"unique"`
	RawImage      string          `json:"image" yaml:"image"`
	BuildParams   BuildParameters `json:"build" yaml:"build"`
	RunParams     RunParameters   `json:"run" yaml:"run"`
	RmParams      RmParameters    `json:"rm" yaml:"rm"`
	StartParams   StartParameters `json:"start" yaml:"start"`
	ExecParams    ExecParameters  `json:"exec" yaml:"exec"`
	hooks         hooks
}

type BuildParameters struct {
	RawContext string      `json:"context" yaml:"context"`
}

type RunParameters struct {
	RawAddHost      []string    `json:"add-host" yaml:"add-host"`
	BlkioWeight     int         `json:"blkio-weight" yaml:"blkio-weight"`
	RawCapAdd       []string    `json:"cap-add" yaml:"cap-add"`
	RawCapDrop      []string    `json:"cap-drop" yaml:"cap-drop"`
	RawCgroupParent string      `json:"cgroup-parent" yaml:"cgroup-parent"`
	CpuPeriod       int         `json:"cpu-period" yaml:"cpu-period"`
	CpuQuota        int         `json:"cpu-quota" yaml:"cpu-quota"`
	RawCidfile      string      `json:"cidfile" yaml:"cidfile"`
	Cpuset          int         `json:"cpuset" yaml:"cpuset"`
	CpuShares       int         `json:"cpu-shares" yaml:"cpu-shares"`
	Detach          bool        `json:"detach" yaml:"detach"`
	RawDevice       []string    `json:"device" yaml:"device"`
	RawDns          []string    `json:"dns" yaml:"dns"`
	RawEntrypoint   string      `json:"entrypoint" yaml:"entrypoint"`
	RawEnv          interface{} `json:"env" yaml:"env"`
	RawEnvFile      []string    `json:"env-file" yaml:"env-file"`
	RawExpose       []string    `json:"expose" yaml:"expose"`
	RawHostname     string      `json:"hostname" yaml:"hostname"`
	Interactive     bool        `json:"interactive" yaml:"interactive"`
	RawLabel        interface{} `json:"label" yaml:"label"`
	RawLabelFile    []string    `json:"label-file" yaml:"label-file"`
	RawLink         []string    `json:"link" yaml:"link"`
	RawLogDriver    string      `json:"log-driver" yaml:"log-driver"`
	RawLogOpt       []string    `json:"log-opt" yaml:"log-opt"`
	RawLxcConf      []string    `json:"lxc-conf" yaml:"lxc-conf"`
	RawMacAddress   string      `json:"mac-address" yaml:"mac-address"`
	RawMemory       string      `json:"memory" yaml:"memory"`
	RawMemorySwap   string      `json:"memory-swap" yaml:"memory-swap"`
	RawNet          string      `json:"net" yaml:"net"`
	OomKillDisable  bool        `json:"oom-kill-disable" yaml:"oom-kill-disable"`
	RawPid          string      `json:"pid" yaml:"pid"`
	Privileged      bool        `json:"privileged" yaml:"privileged"`
	RawPublish      []string    `json:"publish" yaml:"publish"`
	PublishAll      bool        `json:"publish-all" yaml:"publish-all"`
	ReadOnly        bool        `json:"read-only" yaml:"read-only"`
	RawRestart      string      `json:"restart" yaml:"restart"`
	Rm              bool        `json:"rm" yaml:"rm"`
	RawSecurityOpt  []string    `json:"security-opt" yaml:"security-opt"`
	SigProxy        OptBool     `json:"sig-proxy" yaml:"sig-proxy"`
	Tty             bool        `json:"tty" yaml:"tty"`
	RawUlimit       []string    `json:"ulimit" yaml:"ulimit"`
	RawUser         string      `json:"user" yaml:"user"`
	RawUts          string      `json:"uts" yaml:"uts"`
	RawVolume       []string    `json:"volume" yaml:"volume"`
	RawVolumesFrom  []string    `json:"volumes-from" yaml:"volumes-from"`
	RawWorkdir      string      `json:"workdir" yaml:"workdir"`
	RawCmd          interface{} `json:"cmd" yaml:"cmd"`
}

type RmParameters struct {
	Volumes bool `json:"volumes" yaml:"volumes"`
}

type StartParameters struct {
	Attach      bool `json:"attach" yaml:"attach"`
	Interactive bool `json:"interactive" yaml:"interactive"`
}

type ExecParameters struct {
	Interactive bool `json:"interactive" yaml:"interactive"`
	Tty         bool `json:"tty" yaml:"tty"`
}

type OptBool struct {
	Defined bool
	Value   bool
}

func (o *OptBool) UnmarshalYAML(unmarshal func(interface{}) error) error {
	if err := unmarshal(&o.Value); err != nil {
		return err
	}
	o.Defined = true
	return nil
}

func (o *OptBool) UnmarshalJSON(b []byte) (err error) {
	if err := json.Unmarshal(b, &o.Value); err != nil {
		return err
	}
	o.Defined = true
	return
}

func (o *OptBool) Truthy() bool {
	return !o.Defined || o.Value
}

func (o *OptBool) Falsy() bool {
	return o.Defined && !o.Value
}

func (c *container) netContainer() (name string) {
	if netParts := strings.Split(c.RunParams.Net(), ":"); len(netParts) == 2 && netParts[0] == "container" {
		// We'll just assume here that the reference is a name, and not an id, even
		// though docker supports it, since we have no bullet-proof way to tell:
		// heuristics to detect whether it's an id could bring false positives, and
		// a lookup in the list of container names could bring false negatives
		name = netParts[1]
	}
	return
}

func (c *container) Dependencies(excluded []string) *Dependencies {
	dependencies := &Dependencies{}
	for _, link := range c.RunParams.Link() {
		linkName := strings.Split(link, ":")[0]
		if !includes(excluded, linkName) && !dependencies.includes(linkName) {
			dependencies.All = append(dependencies.All, linkName)
			dependencies.Link = append(dependencies.Link, linkName)
		}
	}
	for _, volumesFrom := range c.RunParams.VolumesFrom() {
		volumesFromName := strings.Split(volumesFrom, ":")[0]
		if !includes(excluded, volumesFromName) && !dependencies.includes(volumesFromName) {
			dependencies.All = append(dependencies.All, volumesFromName)
			dependencies.VolumesFrom = append(dependencies.VolumesFrom, volumesFromName)
		}
	}
	if dependencies.Net = c.netContainer(); dependencies.Net != "" {
		if !includes(excluded, dependencies.Net) && !dependencies.includes(dependencies.Net) {
			dependencies.All = append(dependencies.All, dependencies.Net)
		}
	}
	return dependencies
}

func (c *container) Name() string {
	return os.ExpandEnv(c.RawName)
}

func (c *container) ActualName() string {
	if c.Unique() {
		return c.prefixedName() + "-" + cfg.UniqueId()
	} else {
		return c.prefixedName()
	}
}

func (c *container) BuildContext() string {
	return os.ExpandEnv(c.BuildParams.RawContext)
}

func (c *container) Image() string {
	return os.ExpandEnv(c.RawImage)
}

func (c *container) Unique() bool {
	return c.RawUnique
}

func (c *container) ImageWithTag() string {
	imageParts := strings.Split(c.Image(), "/")
	lastImagePart := imageParts[len(imageParts)-1]
	if !strings.Contains(lastImagePart, ":") {
		return lastImagePart + ":latest"
	}
	return lastImagePart
}

func (r *RunParameters) AddHost() []string {
	var addHost []string
	for _, rawAddHost := range r.RawAddHost {
		addHost = append(addHost, os.ExpandEnv(rawAddHost))
	}
	return addHost
}

func (r *RunParameters) CapAdd() []string {
	var capAdd []string
	for _, rawCapAdd := range r.RawCapAdd {
		capAdd = append(capAdd, os.ExpandEnv(rawCapAdd))
	}
	return capAdd
}

func (r *RunParameters) CapDrop() []string {
	var capDrop []string
	for _, rawCapDrop := range r.RawCapDrop {
		capDrop = append(capDrop, os.ExpandEnv(rawCapDrop))
	}
	return capDrop
}

func (r *RunParameters) CgroupParent() string {
	return os.ExpandEnv(r.RawCgroupParent)
}

func (r *RunParameters) Cidfile() string {
	return os.ExpandEnv(r.RawCidfile)
}

func (r *RunParameters) Device() []string {
	var device []string
	for _, rawDevice := range r.RawDevice {
		device = append(device, os.ExpandEnv(rawDevice))
	}
	return device
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
	if r.RawEnv != nil {
		switch rawEnv := r.RawEnv.(type) {
		case []interface{}:
			for _, v := range rawEnv {
				env = append(env, os.ExpandEnv(v.(string)))
			}
		case map[interface{}]interface{}:
			for k, v := range rawEnv {
				env = append(env, os.ExpandEnv(k.(string))+"="+os.ExpandEnv(v.(string)))
			}
		default:
			printErrorf("env is of unknown type!")
		}
	}
	return env
}

func (r *RunParameters) EnvFile() []string {
	var envFile []string
	for _, rawEnvFile := range r.RawEnvFile {
		envFile = append(envFile, os.ExpandEnv(rawEnvFile))
	}
	return envFile
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

func (r *RunParameters) Label() []string {
	var label []string
	if r.RawLabel != nil {
		switch rawLabel := r.RawLabel.(type) {
		case []interface{}:
			for _, v := range rawLabel {
				label = append(label, os.ExpandEnv(v.(string)))
			}
		case map[interface{}]interface{}:
			for k, v := range rawLabel {
				label = append(label, os.ExpandEnv(k.(string))+"="+os.ExpandEnv(v.(string)))
			}
		default:
			printErrorf("label is of unknown type!")
		}
	}
	return label
}

func (r *RunParameters) LabelFile() []string {
	var labelFile []string
	for _, rawLabelFile := range r.RawLabelFile {
		labelFile = append(labelFile, os.ExpandEnv(rawLabelFile))
	}
	return labelFile
}

func (r *RunParameters) Link() []string {
	var link []string
	for _, rawLink := range r.RawLink {
		link = append(link, os.ExpandEnv(rawLink))
	}
	return link
}

func (r *RunParameters) LogDriver() string {
	return os.ExpandEnv(r.RawLogDriver)
}

func (r *RunParameters) LogOpt() []string {
	var opt []string
	for _, rawOpt := range r.RawLogOpt {
		opt = append(opt, os.ExpandEnv(rawOpt))
	}
	return opt
}

func (r *RunParameters) LxcConf() []string {
	var lxcConf []string
	for _, rawLxcConf := range r.RawLxcConf {
		lxcConf = append(lxcConf, os.ExpandEnv(rawLxcConf))
	}
	return lxcConf
}

func (r *RunParameters) MacAddress() string {
	return os.ExpandEnv(r.RawMacAddress)
}

func (r *RunParameters) Memory() string {
	return os.ExpandEnv(r.RawMemory)
}

func (r *RunParameters) MemorySwap() string {
	return os.ExpandEnv(r.RawMemorySwap)
}

func (r *RunParameters) Net() string {
	// Default to bridge
	if len(r.RawNet) == 0 {
		return "bridge"
	} else {
		return os.ExpandEnv(r.RawNet)
	}
}

func (r *RunParameters) Pid() string {
	return os.ExpandEnv(r.RawPid)
}

func (r *RunParameters) Publish() []string {
	var publish []string
	for _, rawPublish := range r.RawPublish {
		publish = append(publish, os.ExpandEnv(rawPublish))
	}
	return publish
}

func (r *RunParameters) Restart() string {
	return os.ExpandEnv(r.RawRestart)
}

func (r *RunParameters) SecurityOpt() []string {
	var securityOpt []string
	for _, rawSecurityOpt := range r.RawSecurityOpt {
		securityOpt = append(securityOpt, os.ExpandEnv(rawSecurityOpt))
	}
	return securityOpt
}

func (r *RunParameters) Ulimit() []string {
	var ulimit []string
	for _, rawUlimit := range r.RawUlimit {
		ulimit = append(ulimit, os.ExpandEnv(rawUlimit))
	}
	return ulimit
}

func (r *RunParameters) User() string {
	return os.ExpandEnv(r.RawUser)
}

func (r *RunParameters) Uts() string {
	return os.ExpandEnv(r.RawUts)
}

func (r *RunParameters) Volume(configPath string) []string {
	var volumes []string
	for _, rawVolume := range r.RawVolume {
		volume := os.ExpandEnv(rawVolume)
		paths := strings.Split(volume, ":")
		if !path.IsAbs(paths[0]) {
			paths[0] = configPath + "/" + paths[0]
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
				cmds, err := shlex.Split(os.ExpandEnv(rawCmd))
				if err != nil {
					printErrorf("Error when parsing cmd `%v`: %v. Proceeding with %q.", rawCmd, err, cmds)
				}
				cmd = append(cmd, cmds...)
			}
		case []interface{}:
			cmds := make([]string, len(rawCmd))
			for i, v := range rawCmd {
				cmds[i] = os.ExpandEnv(v.(string))
			}
			cmd = append(cmd, cmds...)
		default:
			printErrorf("cmd is of unknown type!")
		}
	}
	return cmd
}

func (c *container) Id() string {
	if len(c.id) == 0 && !c.Unique() {
		// `docker inspect` works for both image and containers, make sure this is a
		// container payload we get back, otherwise we might end up getting the Id
		// of the image of the same name.
		c.id = inspectString(c.ActualName(), "{{if .State}}{{.Id}}{{else}}{{end}}")
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
	_, err := commandOutput("docker", []string{"inspect", c.ImageWithTag()})
	return err == nil
}

func (c *container) Status() []string {
	if c.Unique() {
		fmt.Printf("Cannot show status of uniquely named container(s) %s.\n", c.prefixedName())
		return []string{c.prefixedName(), c.Image(), "-", "-", "-", "-", "-"}
	}
	fields := []string{c.ActualName(), c.Image(), "-", "-", "-", "-", "-"}
	output := inspectString(c.Id(), "{{.Id}}\t{{.Image}}\t{{if .NetworkSettings.IPAddress}}{{.NetworkSettings.IPAddress}}{{else}}-{{end}}\t{{range $k,$v := $.NetworkSettings.Ports}}{{$k}},{{else}}-{{end}}\t{{.State.Running}}")
	if output != "" {
		copy(fields[2:], strings.Split(output, "\t"))
		// We asked for the image id the container was created from
		fields[3] = strconv.FormatBool(imageIdFromTag(fields[1]) == fields[3])
	}
	return fields
}

func (c *container) Lift(cmds []string, nocache bool, excluded []string, configPath string) {
	c.Provision(nocache)
	c.Run(cmds, excluded, configPath)
}

func (c *container) Provision(nocache bool) {
	if len(c.BuildContext()) > 0 {
		c.buildImage(nocache)
	} else {
		c.PullImage()
	}
}

// Create container
func (c *container) Create(cmds []string, excluded []string, configPath string) {
	if !c.Unique() {
		c.Rm(true)
	}
	fmt.Printf("Creating container %s ... ", c.ActualName())

	args := append([]string{"create"}, c.createArgs(cmds, excluded, configPath)...)
	executeCommand("docker", args)
}

// Run container, or start it if already existing
func (c *container) Run(cmds []string, excluded []string, configPath string) {
	if !c.Unique() {
		c.Rm(true)
	}
	executeHook(c.Hooks().PreStart(), c.ActualName())
	fmt.Printf("Running container %s ... ", c.ActualName())

	args := []string{"run"}
	// Detach
	if c.RunParams.Detach {
		args = append(args, "--detach")
	}
	args = append(args, c.createArgs(cmds, excluded, configPath)...)
	c.executePostStartHook()
	executeCommand("docker", args)
}

func (c *container) executePostStartHook() {
	if len(c.Hooks().PostStart()) > 0 {
		cmd, cmdOut, _ := executeCommandBackground("docker", []string{"events", "--filter", "event=start", "--filter", "container=" + c.ActualName()})
		go func() {
			defer func() {
				handleRecoveredError(recover())
			}()
			r := bufio.NewReader(cmdOut)
			_, _, err := r.ReadLine()
			cmd.Process.Kill()
			if err != nil {
				printNoticef("Could not execute post-start hook for %s.", c.ActualName())
			} else {
				executeHook(c.Hooks().PostStart(), c.ActualName())
			}
		}()
	}
}

// Returns all the flags to be passed to `docker create`
func (c *container) createArgs(cmds []string, excluded []string, configPath string) []string {
	args := []string{}
	// AddHost
	for _, addHost := range c.RunParams.AddHost() {
		args = append(args, "--add-host", addHost)
	}
	// BlkioWeight
	if c.RunParams.BlkioWeight > 0 {
		args = append(args, "--blkio-weight", strconv.Itoa(c.RunParams.BlkioWeight))
	}
	// CapAdd
	for _, capAdd := range c.RunParams.CapAdd() {
		args = append(args, "--cap-add", capAdd)
	}
	// CapDrop
	for _, capDrop := range c.RunParams.CapDrop() {
		args = append(args, "--cap-drop", capDrop)
	}
	// CgroupParent
	if len(c.RunParams.CgroupParent()) > 0 {
		args = append(args, "--cgroup-parent", c.RunParams.CgroupParent())
	}
	// Cidfile
	if len(c.RunParams.Cidfile()) > 0 {
		args = append(args, "--cidfile", c.RunParams.Cidfile())
	}
	// CpuPeriod
	if c.RunParams.CpuPeriod > 0 {
		args = append(args, "--cpu-period", strconv.Itoa(c.RunParams.CpuPeriod))
	}
	// CpuQuota
	if c.RunParams.CpuQuota > 0 {
		args = append(args, "--cpu-quota", strconv.Itoa(c.RunParams.CpuQuota))
	}
	// CPU set
	if c.RunParams.Cpuset > 0 {
		args = append(args, "--cpuset", strconv.Itoa(c.RunParams.Cpuset))
	}
	// CPU shares
	if c.RunParams.CpuShares > 0 {
		args = append(args, "--cpu-shares", strconv.Itoa(c.RunParams.CpuShares))
	}
	// Device
	for _, device := range c.RunParams.Device() {
		args = append(args, "--device", device)
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
	for _, envFile := range c.RunParams.EnvFile() {
		args = append(args, "--env-file", envFile)
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
	// Label
	for _, label := range c.RunParams.Label() {
		args = append(args, "--label", label)
	}
	// LabelFile
	for _, labelFile := range c.RunParams.LabelFile() {
		args = append(args, "--label-file", labelFile)
	}
	// Link
	for _, link := range c.RunParams.Link() {
		linkParts := strings.Split(link, ":")
		linkName := linkParts[0]
		if !includes(excluded, linkName) {
			linkParts[0] = cfg.Container(linkName).ActualName()
			args = append(args, "--link", strings.Join(linkParts, ":"))
		}
	}
	// LogDriver
	if len(c.RunParams.LogDriver()) > 0 {
		args = append(args, "--log-driver", c.RunParams.LogDriver())
	}
	// LogOpt
	for _, opt := range c.RunParams.LogOpt() {
		args = append(args, "--log-opt", opt)
	}
	// LxcConf
	for _, lxcConf := range c.RunParams.LxcConf() {
		args = append(args, "--lxc-conf", lxcConf)
	}
	// Mac address
	if len(c.RunParams.MacAddress()) > 0 {
		args = append(args, "--mac-address", c.RunParams.MacAddress())
	}
	// Memory
	if len(c.RunParams.Memory()) > 0 {
		args = append(args, "--memory", c.RunParams.Memory())
	}
	// MemorySwap
	if len(c.RunParams.MemorySwap()) > 0 {
		args = append(args, "--memory-swap", c.RunParams.MemorySwap())
	}
	// Net
	if c.RunParams.Net() != "bridge" {
		if !includes(excluded, c.netContainer()) {
			args = append(args, "--net", cfg.Container(c.RunParams.Net()).ActualName())
		}
	}
	// OomKillDisable
	if c.RunParams.OomKillDisable {
		args = append(args, "--oom-kill-disable")
	}
	// PID
	if len(c.RunParams.Pid()) > 0 {
		args = append(args, "--pid", c.RunParams.Pid())
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
	// ReadOnly
	if c.RunParams.ReadOnly {
		args = append(args, "--read-only")
	}
	// Restart
	if len(c.RunParams.Restart()) > 0 {
		args = append(args, "--restart", c.RunParams.Restart())
	}
	// Rm
	if c.RunParams.Rm {
		args = append(args, "--rm")
	}
	// SecurityOpt
	for _, securityOpt := range c.RunParams.SecurityOpt() {
		args = append(args, "--security-opt", securityOpt)
	}
	// SigProxy
	if c.RunParams.SigProxy.Falsy() {
		args = append(args, "--sig-proxy=false")
	}
	// Tty
	if c.RunParams.Tty {
		args = append(args, "--tty")
	}
	// Ulimit
	for _, ulimit := range c.RunParams.Ulimit() {
		args = append(args, "--ulimit", ulimit)
	}
	// User
	if len(c.RunParams.User()) > 0 {
		args = append(args, "--user", c.RunParams.User())
	}
	// Uts
	if len(c.RunParams.Uts()) > 0 {
		args = append(args, "--uts", c.RunParams.Uts())
	}
	// Volumes
	for _, volume := range c.RunParams.Volume(configPath) {
		args = append(args, "--volume", volume)
	}
	// VolumesFrom
	for _, volumesFrom := range c.RunParams.VolumesFrom() {
		volumesFromParts := strings.Split(volumesFrom, ":")
		volumesFromName := volumesFromParts[0]
		if !includes(excluded, volumesFromName) {
			volumesFromParts[0] = cfg.Container(volumesFromName).ActualName()
			args = append(args, "--volumes-from", strings.Join(volumesFromParts, ":"))
		}
	}
	// Workdir
	if len(c.RunParams.Workdir()) > 0 {
		args = append(args, "--workdir", c.RunParams.Workdir())
	}
	// Name
	args = append(args, "--name", c.ActualName())
	// Image
	args = append(args, c.Image())
	// Command
	if len(cmds) > 0 {
		args = append(args, cmds...)
	} else {
		args = append(args, c.RunParams.Cmd()...)
	}
	return args
}

// Start container
func (c *container) Start(excluded []string, configPath string) {
	if c.Exists() && !c.Unique() {
		if !c.Running() {
			executeHook(c.Hooks().PreStart(), c.ActualName())
			fmt.Printf("Starting container %s ... ", c.ActualName())
			args := []string{"start"}
			if c.StartParams.Attach {
				args = append(args, "--attach")
			}
			if c.StartParams.Interactive {
				args = append(args, "--interactive")
			}
			args = append(args, c.ActualName())
			c.executePostStartHook()
			executeCommand("docker", args)
		}
	} else {
		c.Run([]string{}, excluded, configPath)
	}
}

// Kill container
func (c *container) Kill() {
	if c.Unique() {
		fmt.Printf("Cannot kill uniquely named container(s) %s.\n", c.prefixedName())
		return
	}
	if c.Running() {
		executeHook(c.Hooks().PreStop(), c.ActualName())
		fmt.Printf("Killing container %s ... ", c.ActualName())
		args := []string{"kill", c.ActualName()}
		executeCommand("docker", args)
		executeHook(c.Hooks().PostStop(), c.ActualName())
	}
}

// Stop container
func (c *container) Stop() {
	if c.Unique() {
		fmt.Printf("Cannot stop uniquely named container(s) %s.\n", c.prefixedName())
		return
	}
	if c.Running() {
		executeHook(c.Hooks().PreStop(), c.ActualName())
		fmt.Printf("Stopping container %s ... ", c.ActualName())
		args := []string{"stop", c.ActualName()}
		executeCommand("docker", args)
		executeHook(c.Hooks().PostStop(), c.ActualName())
	}
}

// Pause container
func (c *container) Pause() {
	if c.Unique() {
		fmt.Printf("Cannot pause uniquely named container(s) %s.\n", c.prefixedName())
		return
	}
	if c.Running() {
		if c.Paused() {
			printNoticef("Container %s is already paused.\n", c.ActualName())
		} else {
			fmt.Printf("Pausing container %s ... ", c.ActualName())
			args := []string{"pause", c.ActualName()}
			executeCommand("docker", args)
		}
	} else {
		printNoticef("Container %s is not running.\n", c.ActualName())
	}
}

// Unpause container
func (c *container) Unpause() {
	if c.Unique() {
		fmt.Printf("Cannot unpause uniquely named container(s) %s.\n", c.prefixedName())
		return
	}
	if c.Paused() {
		fmt.Printf("Unpausing container %s ... ", c.ActualName())
		args := []string{"unpause", c.ActualName()}
		executeCommand("docker", args)
	}
}

// Exec command in container
func (c *container) Exec(cmds []string, configPath string) {
	if !c.Running() {
		c.Start([]string{}, configPath)
	}
	args := []string{"exec"}
	if c.ExecParams.Interactive {
		args = append(args, "--interactive")
	}
	if c.ExecParams.Tty {
		args = append(args, "--tty")
	}
	args = append(args, c.ActualName())
	args = append(args, cmds...)
	executeCommand("docker", args)
}

// Remove container
func (c *container) Rm(force bool) {
	if c.Unique() {
		fmt.Printf("Cannot remove uniquely named container(s) %s.\n", c.prefixedName())
		return
	}
	if c.Exists() {
		containerIsRunning := c.Running()
		if !force && containerIsRunning {
			printErrorf("Container %s is running and cannot be removed. Use --force to remove anyway.\n", c.ActualName())
		} else {
			args := []string{"rm"}
			if force && containerIsRunning {
				executeHook(c.Hooks().PreStop(), c.ActualName())
				args = append(args, "--force")
			}
			if c.RmParams.Volumes {
				fmt.Printf("Removing container %s and its volumes ... ", c.ActualName())
				args = append(args, "--volumes")
			} else {
				fmt.Printf("Removing container %s ... ", c.ActualName())
			}
			args = append(args, c.ActualName())
			executeCommand("docker", args)
			if force && containerIsRunning {
				executeHook(c.Hooks().PostStop(), c.ActualName())
			}
			c.id = ""
		}
	}
}

// Dump container logs
func (c *container) Logs(follow bool, since string, tail string) (stdout, stderr io.Reader) {
	if c.Unique() {
		fmt.Printf("Cannot show logs of uniquely named container(s) %s.\n", c.prefixedName())
		return
	}
	if c.Exists() {
		args := []string{"logs"}
		if follow {
			args = append(args, "-f")
		}
		if len(since) > 0 {
			args = append(args, "--since", since)
		}
		if len(tail) > 0 && tail != "all" {
			args = append(args, "--tail", tail)
		}
		// always include timestamps for ordering, we'll just strip
		// them if the user doesn't want to see them
		args = append(args, "-t")
		args = append(args, c.Id())
		_, stdout, stderr := executeCommandBackground("docker", args)
		return stdout, stderr
	} else {
		return nil, nil
	}
}

// Push container
func (c *container) Push() {
	if len(c.Image()) > 0 {
		fmt.Printf("Pushing image %s ... ", c.Image())
		args := []string{"push", c.Image()}
		executeCommand("docker", args)
	} else {
		printNoticef("Skipping %s as it does not have an image name.\n", c.ActualName())
	}
}

func (c *container) Hooks() Hooks {
	return &c.hooks
}

// Pull image for container
func (c *container) PullImage() {
	fmt.Printf("Pulling image %s ... ", c.Image())
	args := []string{"pull", c.Image()}
	executeCommand("docker", args)
}

// Build image for container
func (c *container) buildImage(nocache bool) {
	executeHook(c.Hooks().PreBuild(), c.ActualName())
	fmt.Printf("Building image %s ... ", c.Image())
	args := []string{"build"}
	if nocache {
		args = append(args, "--no-cache")
	}
	args = append(args, "--rm", "--tag="+c.Image(), c.BuildContext())
	executeCommand("docker", args)
	executeHook(c.Hooks().PostBuild(), c.ActualName())
}

func (c *container) prefixedName() string {
	return cfg.Prefix() + c.Name()
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
