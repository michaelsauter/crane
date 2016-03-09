package crane

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/flynn/go-shlex"
	"io"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
)

type Container interface {
	ContainerInfo
	Exists() bool
	InstancesOfStatus(status string) []string
	Status() [][]string
	Provision(nocache bool)
	PullImage()
	Create(cmds []string, excluded []string)
	Run(cmds []string, excluded []string)
	Start(excluded []string)
	Kill()
	Stop()
	Pause()
	Unpause()
	Exec(cmds []string)
	Rm(force bool)
	Logs(follow bool, since string, tail string) (sources []logSource)
	Push()
	SetCommandsOutput(stdout, stderr io.Writer)
	CommandsOut() io.Writer
	CommandsErr() io.Writer
}

type ContainerInfo interface {
	Name() string
	PrefixedName() string
	ActualName() string
	Image() string
	ID() string
	Dependencies() *Dependencies
	Unique() bool
	BuildParams() BuildParameters
	RunParams() RunParameters
	RmParams() RmParameters
	StartParams() StartParameters
	ExecParams() ExecParameters
	Hooks() Hooks
}

type container struct {
	id          string
	RawName     string
	RawUnique   bool            `json:"unique" yaml:"unique"`
	RawImage    string          `json:"image" yaml:"image"`
	RawRequires []string        `json:"requires" yaml:"requires"`
	RawBuild    BuildParameters `json:"build" yaml:"build"`
	RawRun      RunParameters   `json:"run" yaml:"run"`
	RawRm       RmParameters    `json:"rm" yaml:"rm"`
	RawStart    StartParameters `json:"start" yaml:"start"`
	RawExec     ExecParameters  `json:"exec" yaml:"exec"`
	hooks       hooks
	stdout      io.Writer
	stderr      io.Writer
}

type BuildParameters struct {
	RawContext   string      `json:"context" yaml:"context"`
	RawFile      string      `json:"file" yaml:"file"`
	RawBuildArgs interface{} `json:"build-arg" yaml:"build-arg"`
}

type RunParameters struct {
	RawAddHost           []string    `json:"add-host" yaml:"add-host"`
	BlkioWeight          int         `json:"blkio-weight" yaml:"blkio-weight"`
	RawCapAdd            []string    `json:"cap-add" yaml:"cap-add"`
	RawCapDrop           []string    `json:"cap-drop" yaml:"cap-drop"`
	RawCgroupParent      string      `json:"cgroup-parent" yaml:"cgroup-parent"`
	CPUPeriod            int         `json:"cpu-period" yaml:"cpu-period"`
	CPUQuota             int         `json:"cpu-quota" yaml:"cpu-quota"`
	RawCidfile           string      `json:"cidfile" yaml:"cidfile"`
	CPUset               int         `json:"cpuset" yaml:"cpuset"`
	CPUShares            int         `json:"cpu-shares" yaml:"cpu-shares"`
	Detach               bool        `json:"detach" yaml:"detach"`
	RawDevice            []string    `json:"device" yaml:"device"`
	RawDNS               []string    `json:"dns" yaml:"dns"`
	RawDNSOpt            []string    `json:"dns-opt" yaml:"dns-opt"`
	RawDNSSearch         []string    `json:"dns-search" yaml:"dns-search"`
	RawEntrypoint        string      `json:"entrypoint" yaml:"entrypoint"`
	RawEnv               interface{} `json:"env" yaml:"env"`
	RawEnvFile           []string    `json:"env-file" yaml:"env-file"`
	RawExpose            []string    `json:"expose" yaml:"expose"`
	RawGroupAdd          []string    `json:"group-add" yaml:"group-add"`
	RawHostname          string      `json:"hostname" yaml:"hostname"`
	Interactive          bool        `json:"interactive" yaml:"interactive"`
	RawIPC               string      `json:"ipc" yaml:"ipc"`
	RawKernelMemory      string      `json:"kernel-memory" yaml:"kernel-memory"`
	RawLabel             interface{} `json:"label" yaml:"label"`
	RawLabelFile         []string    `json:"label-file" yaml:"label-file"`
	RawLink              []string    `json:"link" yaml:"link"`
	RawLogDriver         string      `json:"log-driver" yaml:"log-driver"`
	RawLogOpt            []string    `json:"log-opt" yaml:"log-opt"`
	RawLxcConf           []string    `json:"lxc-conf" yaml:"lxc-conf"`
	RawMacAddress        string      `json:"mac-address" yaml:"mac-address"`
	RawMemory            string      `json:"memory" yaml:"memory"`
	RawMemoryReservation string      `json:"memory-reservation" yaml:"memory-reservation"`
	RawMemorySwap        string      `json:"memory-swap" yaml:"memory-swap"`
	MemorySwappiness     OptInt      `json:"memory-swappiness" yaml:"memory-swappiness"`
	RawNet               string      `json:"net" yaml:"net"`
	OomKillDisable       bool        `json:"oom-kill-disable" yaml:"oom-kill-disable"`
	RawPid               string      `json:"pid" yaml:"pid"`
	Privileged           bool        `json:"privileged" yaml:"privileged"`
	RawPublish           []string    `json:"publish" yaml:"publish"`
	PublishAll           bool        `json:"publish-all" yaml:"publish-all"`
	ReadOnly             bool        `json:"read-only" yaml:"read-only"`
	RawRestart           string      `json:"restart" yaml:"restart"`
	Rm                   bool        `json:"rm" yaml:"rm"`
	RawSecurityOpt       []string    `json:"security-opt" yaml:"security-opt"`
	SigProxy             OptBool     `json:"sig-proxy" yaml:"sig-proxy"`
	RawStopSignal        string      `json:"stop-signal" yaml:"stop-signal"`
	Tty                  bool        `json:"tty" yaml:"tty"`
	RawUlimit            []string    `json:"ulimit" yaml:"ulimit"`
	RawUser              string      `json:"user" yaml:"user"`
	RawUts               string      `json:"uts" yaml:"uts"`
	RawVolume            []string    `json:"volume" yaml:"volume"`
	RawVolumesFrom       []string    `json:"volumes-from" yaml:"volumes-from"`
	RawWorkdir           string      `json:"workdir" yaml:"workdir"`
	RawCmd               interface{} `json:"cmd" yaml:"cmd"`
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

type OptInt struct {
	Defined bool
	Value   int
}

type OptBool struct {
	Defined bool
	Value   bool
}

type logSource struct {
	Stdout io.Reader
	Stderr io.Reader
	Name   string
}

func (o *OptInt) UnmarshalYAML(unmarshal func(interface{}) error) error {
	if err := unmarshal(&o.Value); err != nil {
		return err
	}
	o.Defined = true
	return nil
}

func (o *OptInt) UnmarshalJSON(b []byte) (err error) {
	if err := json.Unmarshal(b, &o.Value); err != nil {
		return err
	}
	o.Defined = true
	return
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

func (o OptBool) Truthy() bool {
	return !o.Defined || o.Value
}

func (o OptBool) Falsy() bool {
	return o.Defined && !o.Value
}

func (c *container) BuildParams() BuildParameters {
	return c.RawBuild
}

func (c *container) RunParams() RunParameters {
	return c.RawRun
}

func (c *container) RmParams() RmParameters {
	return c.RawRm
}

func (c *container) StartParams() StartParameters {
	return c.RawStart
}

func (c *container) ExecParams() ExecParameters {
	return c.RawExec
}

func (c *container) Dependencies() *Dependencies {
	dependencies := &Dependencies{}
	for _, required := range c.Requires() {
		if !includes(excluded, required) && !dependencies.includes(required) {
			dependencies.All = append(dependencies.All, required)
			dependencies.Requires = append(dependencies.Requires, required)
		}
	}
	if c.RunParams().Net() == "bridge" {
		// links are strict dependencies only on bridge networks
		for _, link := range c.RunParams().Link() {
			linkName := strings.Split(link, ":")[0]
			if !includes(excluded, linkName) && !dependencies.includes(linkName) {
				dependencies.All = append(dependencies.All, linkName)
				dependencies.Link = append(dependencies.Link, linkName)
			}
		}
	}
	for _, volumesFrom := range c.RunParams().VolumesFrom() {
		volumesFromName := strings.Split(volumesFrom, ":")[0]
		if !includes(excluded, volumesFromName) && !dependencies.includes(volumesFromName) {
			dependencies.All = append(dependencies.All, volumesFromName)
			dependencies.VolumesFrom = append(dependencies.VolumesFrom, volumesFromName)
		}
	}
	if net := containerReference(c.RunParams().Net()); net != "" {
		if !includes(excluded, net) && !dependencies.includes(net) {
			dependencies.Net = net
			dependencies.All = append(dependencies.All, net)
		}
	}
	if ipc := containerReference(c.RunParams().IPC()); ipc != "" {
		if !includes(excluded, ipc) && !dependencies.includes(ipc) {
			dependencies.IPC = ipc
			dependencies.All = append(dependencies.All, ipc)
		}
	}
	return dependencies
}

func (c *container) Name() string {
	return expandEnv(c.RawName)
}

func (c *container) ActualName() string {
	if c.Unique() {
		return c.PrefixedName() + "-unique-" + cfg.UniqueID()
	}
	return c.PrefixedName()
}

func (c *container) Image() string {
	image := expandEnv(c.RawImage)

	// Return if no global tag given or image is a digest
	if len(cfg.Tag()) == 0 || strings.Contains(image, "@") {
		return image
	}

	// Replace image tag with global tag
	startOfTag := strings.LastIndex(image, ":")
	if startOfTag != -1 {
		image = image[:startOfTag]
	}
	return image + ":" + cfg.Tag()
}

func (c *container) Unique() bool {
	return c.RawUnique
}

func (c *container) Requires() []string {
	var requires []string
	for _, rawRequired := range c.RawRequires {
		requires = append(requires, expandEnv(rawRequired))
	}
	return requires
}

func (b BuildParameters) Context() string {
	return expandEnv(b.RawContext)
}

func (b BuildParameters) File() string {
	return expandEnv(b.RawFile)
}

func (b BuildParameters) BuildArgs() []string {
	return sliceOrMap2ExpandedSlice(b.RawBuildArgs)
}

func (r RunParameters) AddHost() []string {
	var addHost []string
	for _, rawAddHost := range r.RawAddHost {
		addHost = append(addHost, expandEnv(rawAddHost))
	}
	return addHost
}

func (r RunParameters) CapAdd() []string {
	var capAdd []string
	for _, rawCapAdd := range r.RawCapAdd {
		capAdd = append(capAdd, expandEnv(rawCapAdd))
	}
	return capAdd
}

func (r RunParameters) CapDrop() []string {
	var capDrop []string
	for _, rawCapDrop := range r.RawCapDrop {
		capDrop = append(capDrop, expandEnv(rawCapDrop))
	}
	return capDrop
}

func (r RunParameters) CgroupParent() string {
	return expandEnv(r.RawCgroupParent)
}

func (r RunParameters) Cidfile() string {
	return expandEnv(r.RawCidfile)
}

func (r RunParameters) Device() []string {
	var device []string
	for _, rawDevice := range r.RawDevice {
		device = append(device, expandEnv(rawDevice))
	}
	return device
}

func (r RunParameters) DNS() []string {
	var dns []string
	for _, rawDNS := range r.RawDNS {
		dns = append(dns, expandEnv(rawDNS))
	}
	return dns
}

func (r RunParameters) DNSOpt() []string {
	var dnsOpt []string
	for _, rawDNSOpt := range r.RawDNSOpt {
		dnsOpt = append(dnsOpt, expandEnv(rawDNSOpt))
	}
	return dnsOpt
}

func (r RunParameters) DNSSearch() []string {
	var dnsSearch []string
	for _, rawDNSSearch := range r.RawDNSSearch {
		dnsSearch = append(dnsSearch, expandEnv(rawDNSSearch))
	}
	return dnsSearch
}

func (r RunParameters) Entrypoint() string {
	return expandEnv(r.RawEntrypoint)
}

func (r RunParameters) Env() []string {
	return sliceOrMap2ExpandedSlice(r.RawEnv)
}

func (r RunParameters) EnvFile() []string {
	var envFile []string
	for _, rawEnvFile := range r.RawEnvFile {
		envFile = append(envFile, expandEnv(rawEnvFile))
	}
	return envFile
}

func (r RunParameters) Expose() []string {
	var expose []string
	for _, rawExpose := range r.RawExpose {
		expose = append(expose, expandEnv(rawExpose))
	}
	return expose
}

func (r RunParameters) GroupAdd() []string {
	var groupAdd []string
	for _, rawGroupAdd := range r.RawGroupAdd {
		groupAdd = append(groupAdd, expandEnv(rawGroupAdd))
	}
	return groupAdd
}

func (r RunParameters) Hostname() string {
	return expandEnv(r.RawHostname)
}

func (r RunParameters) IPC() string {
	return expandEnv(r.RawIPC)
}

func (r RunParameters) KernelMemory() string {
	return expandEnv(r.RawKernelMemory)
}

func (r RunParameters) Label() []string {
	return sliceOrMap2ExpandedSlice(r.RawLabel)
}

func (r RunParameters) LabelFile() []string {
	var labelFile []string
	for _, rawLabelFile := range r.RawLabelFile {
		labelFile = append(labelFile, expandEnv(rawLabelFile))
	}
	return labelFile
}

func (r RunParameters) Link() []string {
	var link []string
	for _, rawLink := range r.RawLink {
		link = append(link, expandEnv(rawLink))
	}
	return link
}

func (r RunParameters) LogDriver() string {
	return expandEnv(r.RawLogDriver)
}

func (r RunParameters) LogOpt() []string {
	var opt []string
	for _, rawOpt := range r.RawLogOpt {
		opt = append(opt, expandEnv(rawOpt))
	}
	return opt
}

func (r RunParameters) LxcConf() []string {
	var lxcConf []string
	for _, rawLxcConf := range r.RawLxcConf {
		lxcConf = append(lxcConf, expandEnv(rawLxcConf))
	}
	return lxcConf
}

func (r RunParameters) MacAddress() string {
	return expandEnv(r.RawMacAddress)
}

func (r RunParameters) Memory() string {
	return expandEnv(r.RawMemory)
}

func (r RunParameters) MemoryReservation() string {
	return expandEnv(r.RawMemoryReservation)
}

func (r RunParameters) MemorySwap() string {
	return expandEnv(r.RawMemorySwap)
}

func (r RunParameters) Net() string {
	// Default to bridge
	if len(r.RawNet) == 0 {
		return "bridge"
	}
	return expandEnv(r.RawNet)
}

func (r RunParameters) ActualNet() string {
	netParam := r.Net()
	if netParam == "bridge" {
		return "bridge"
	}
	netContainer := containerReference(netParam)
	if len(netContainer) > 0 {
		if !includes(excluded, netContainer) {
			return "container:" + cfg.Container(netContainer).ActualName()
		}
	} else {
		if includes(cfg.NetworkNames(), netParam) {
			return cfg.Network(netParam).ActualName()
		}
	}
	return netParam
}

func (r RunParameters) Pid() string {
	return expandEnv(r.RawPid)
}

func (r RunParameters) Publish() []string {
	var publish []string
	for _, rawPublish := range r.RawPublish {
		publish = append(publish, expandEnv(rawPublish))
	}
	return publish
}

func (r RunParameters) Restart() string {
	return expandEnv(r.RawRestart)
}

func (r RunParameters) SecurityOpt() []string {
	var securityOpt []string
	for _, rawSecurityOpt := range r.RawSecurityOpt {
		securityOpt = append(securityOpt, expandEnv(rawSecurityOpt))
	}
	return securityOpt
}

func (r RunParameters) StopSignal() string {
	return expandEnv(r.RawStopSignal)
}

func (r RunParameters) Ulimit() []string {
	var ulimit []string
	for _, rawUlimit := range r.RawUlimit {
		ulimit = append(ulimit, expandEnv(rawUlimit))
	}
	return ulimit
}

func (r RunParameters) User() string {
	return expandEnv(r.RawUser)
}

func (r RunParameters) Uts() string {
	return expandEnv(r.RawUts)
}

func (r RunParameters) Volume() []string {
	var volumes []string
	for _, rawVolume := range r.RawVolume {
		volume := expandEnv(rawVolume)
		parts := strings.Split(volume, ":")
		if !includes(cfg.VolumeNames(), parts[0]) && !path.IsAbs(parts[0]) {
			parts[0] = cfg.Path() + "/" + parts[0]
		}
		volumes = append(volumes, strings.Join(parts, ":"))
	}
	return volumes
}

func (r RunParameters) VolumeSources() []string {
	volumes := r.Volume()
	var volumeSources []string
	for _, volume := range volumes {
		parts := strings.Split(volume, ":")
		volumeSources = append(volumeSources, parts[0])
	}
	return volumeSources
}

func (r RunParameters) ActualVolume() []string {
	vols := []string{}
	for _, volume := range r.Volume() {
		parts := strings.Split(volume, ":")
		if includes(cfg.VolumeNames(), parts[0]) {
			parts[0] = cfg.Volume(parts[0]).ActualName()
		}
		vols = append(vols, strings.Join(parts, ":"))
	}
	return vols
}

func (r RunParameters) VolumesFrom() []string {
	var volumesFrom []string
	for _, rawVolumesFrom := range r.RawVolumesFrom {
		volumesFrom = append(volumesFrom, expandEnv(rawVolumesFrom))
	}
	return volumesFrom
}

func (r RunParameters) Workdir() string {
	return expandEnv(r.RawWorkdir)
}

func (r RunParameters) Cmd() []string {
	var cmd []string
	if r.RawCmd != nil {
		switch rawCmd := r.RawCmd.(type) {
		case string:
			if len(rawCmd) > 0 {
				cmds, err := shlex.Split(expandEnv(rawCmd))
				if err != nil {
					printErrorf("Error when parsing cmd `%v`: %v. Proceeding with %q.", rawCmd, err, cmds)
				}
				cmd = append(cmd, cmds...)
			}
		case []interface{}:
			cmds := make([]string, len(rawCmd))
			for i, v := range rawCmd {
				cmds[i] = expandEnv(fmt.Sprintf("%v", v))
			}
			cmd = append(cmd, cmds...)
		default:
			panic(StatusError{fmt.Errorf("unknown type: %v", rawCmd), 65})
		}
	}
	return cmd
}

func (c *container) ID() string {
	if len(c.id) == 0 && !c.Unique() {
		// `docker inspect` works for both image and containers, make sure this is a
		// container payload we get back, otherwise we might end up getting the ID
		// of the image of the same name.
		c.id = inspectString(c.ActualName(), "{{if .State}}{{.Id}}{{else}}{{end}}")
	}
	return c.id
}

func (c *container) Exists() bool {
	return c.ID() != ""
}

func (c *container) Status() [][]string {
	rows := [][]string{}
	existingInstances := c.InstancesOfStatus("existing")
	if len(existingInstances) == 0 {
		fields := []string{c.ActualName(), c.Image(), "-", "-", "-", "-", "-"}
		rows = append(rows, fields)
	} else {
		for _, name := range existingInstances {
			fields := []string{name, c.Image(), "-", "-", "-", "-", "-"}
			// When using a `--tag` global flag, c.Image() may not represent an actual image tag.
			// Instead we should get an image tag by inspecting "Config.Image".
			output := inspectString(name, "{{.Config.Image}}+++{{.Id}}+++{{.Image}}+++{{if .NetworkSettings.IPAddress}}{{.NetworkSettings.IPAddress}}{{else}}-{{end}}+++{{range $k,$v := $.NetworkSettings.Ports}}{{$k}},{{else}}-{{end}}+++{{.State.Running}}")
			if output != "" {
				copy(fields[1:], strings.Split(output, "+++"))
				// We asked for the image id the container was created from
				fields[3] = strconv.FormatBool(imageIDFromTag(fields[1]) == fields[3])
			}
			rows = append(rows, fields)
		}
	}
	return rows
}

func (c *container) Provision(nocache bool) {
	if len(c.BuildParams().Context()) > 0 {
		c.buildImage(nocache)
	} else {
		c.PullImage()
	}
}

// Create container
func (c *container) Create(cmds []string, excluded []string) {
	if !c.Unique() {
		c.Rm(true)
	}
	fmt.Fprintf(c.CommandsOut(), "Creating container %s ...\n", c.ActualName())

	args := append([]string{"create"}, c.createArgs(cmds, excluded)...)
	executeCommand("docker", args, c.CommandsOut(), c.CommandsErr())
}

// Run container, or start it if already existing
func (c *container) Run(cmds []string, excluded []string) {
	if !c.Unique() {
		c.Rm(true)
	}
	executeHook(c.Hooks().PreStart(), c.ActualName())
	fmt.Fprintf(c.CommandsOut(), "Running container %s ...\n", c.ActualName())

	args := []string{"run"}
	// Detach
	if c.RunParams().Detach {
		args = append(args, "--detach")
	}
	args = append(args, c.createArgs(cmds, excluded)...)
	wg := c.executePostStartHook()
	executeCommand("docker", args, c.CommandsOut(), c.CommandsErr())
	wg.Wait()
}

func (c *container) executePostStartHook() *sync.WaitGroup {

	var wg sync.WaitGroup

	if len(c.Hooks().PostStart()) > 0 {
		wg.Add(1)
		cmd, cmdOut, _ := executeCommandBackground("docker", []string{"events", "--filter", "event=start", "--filter", "container=" + c.ActualName()})
		go func() {
			defer func() {
				handleRecoveredError(recover())
				wg.Done()
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

	return &wg
}

// Returns all the flags to be passed to `docker create`
func (c *container) createArgs(cmds []string, excluded []string) []string {
	args := []string{}
	// AddHost
	for _, addHost := range c.RunParams().AddHost() {
		args = append(args, "--add-host", addHost)
	}
	// BlkioWeight
	if c.RunParams().BlkioWeight > 0 {
		args = append(args, "--blkio-weight", strconv.Itoa(c.RunParams().BlkioWeight))
	}
	// CapAdd
	for _, capAdd := range c.RunParams().CapAdd() {
		args = append(args, "--cap-add", capAdd)
	}
	// CapDrop
	for _, capDrop := range c.RunParams().CapDrop() {
		args = append(args, "--cap-drop", capDrop)
	}
	// CgroupParent
	if len(c.RunParams().CgroupParent()) > 0 {
		args = append(args, "--cgroup-parent", c.RunParams().CgroupParent())
	}
	// Cidfile
	if len(c.RunParams().Cidfile()) > 0 {
		args = append(args, "--cidfile", c.RunParams().Cidfile())
	}
	// CPUPeriod
	if c.RunParams().CPUPeriod > 0 {
		args = append(args, "--cpu-period", strconv.Itoa(c.RunParams().CPUPeriod))
	}
	// CPUQuota
	if c.RunParams().CPUQuota > 0 {
		args = append(args, "--cpu-quota", strconv.Itoa(c.RunParams().CPUQuota))
	}
	// CPU set
	if c.RunParams().CPUset > 0 {
		args = append(args, "--cpuset", strconv.Itoa(c.RunParams().CPUset))
	}
	// CPU shares
	if c.RunParams().CPUShares > 0 {
		args = append(args, "--cpu-shares", strconv.Itoa(c.RunParams().CPUShares))
	}
	// Device
	for _, device := range c.RunParams().Device() {
		args = append(args, "--device", device)
	}
	// DNS
	for _, dns := range c.RunParams().DNS() {
		args = append(args, "--dns", dns)
	}

	// DNSOpt
	for _, dnsOpt := range c.RunParams().DNSOpt() {
		args = append(args, "--dns-opt", dnsOpt)
	}
	// DNS Search
	for _, dnsSearch := range c.RunParams().DNSSearch() {
		args = append(args, "--dns-search", dnsSearch)
	}
	// Entrypoint
	if len(c.RunParams().Entrypoint()) > 0 {
		args = append(args, "--entrypoint", c.RunParams().Entrypoint())
	}
	// Env
	for _, env := range c.RunParams().Env() {
		args = append(args, "--env", env)
	}
	// Env file
	for _, envFile := range c.RunParams().EnvFile() {
		args = append(args, "--env-file", envFile)
	}
	// Expose
	for _, expose := range c.RunParams().Expose() {
		args = append(args, "--expose", expose)
	}
	// GroupAdd
	for _, groupAdd := range c.RunParams().GroupAdd() {
		args = append(args, "--group-add", groupAdd)
	}
	// Host
	if len(c.RunParams().Hostname()) > 0 {
		args = append(args, "--hostname", c.RunParams().Hostname())
	}
	// Interactive
	if c.RunParams().Interactive {
		args = append(args, "--interactive")
	}
	// IPC
	if len(c.RunParams().IPC()) > 0 {
		ipcContainer := containerReference(c.RunParams().IPC())
		if len(ipcContainer) > 0 {
			if !includes(excluded, ipcContainer) {
				args = append(args, "--ipc", "container:"+cfg.Container(ipcContainer).ActualName())
			}
		} else {
			args = append(args, "--ipc", c.RunParams().IPC())
		}
	}
	// KernelMemory
	if len(c.RunParams().KernelMemory()) > 0 {
		args = append(args, "--kernel-memory", c.RunParams().KernelMemory())
	}
	// Label
	for _, label := range c.RunParams().Label() {
		args = append(args, "--label", label)
	}
	// LabelFile
	for _, labelFile := range c.RunParams().LabelFile() {
		args = append(args, "--label-file", labelFile)
	}
	// Link
	for _, link := range c.RunParams().Link() {
		linkParts := strings.Split(link, ":")
		linkName := linkParts[0]
		if !includes(excluded, linkName) {
			linkParts[0] = cfg.Container(linkName).ActualName()
			args = append(args, "--link", strings.Join(linkParts, ":"))
		}
	}
	// LogDriver
	if len(c.RunParams().LogDriver()) > 0 {
		args = append(args, "--log-driver", c.RunParams().LogDriver())
	}
	// LogOpt
	for _, opt := range c.RunParams().LogOpt() {
		args = append(args, "--log-opt", opt)
	}
	// LxcConf
	for _, lxcConf := range c.RunParams().LxcConf() {
		args = append(args, "--lxc-conf", lxcConf)
	}
	// Mac address
	if len(c.RunParams().MacAddress()) > 0 {
		args = append(args, "--mac-address", c.RunParams().MacAddress())
	}
	// Memory
	if len(c.RunParams().Memory()) > 0 {
		args = append(args, "--memory", c.RunParams().Memory())
	}
	// MemoryReservation
	if len(c.RunParams().MemoryReservation()) > 0 {
		args = append(args, "--memory-reservation", c.RunParams().MemoryReservation())
	}
	// MemorySwap
	if len(c.RunParams().MemorySwap()) > 0 {
		args = append(args, "--memory-swap", c.RunParams().MemorySwap())
	}
	// MemorySwappiness
	if c.RunParams().MemorySwappiness.Defined {
		args = append(args, "--memory-swappiness", strconv.Itoa(c.RunParams().MemorySwappiness.Value))
	}
	// Net
	netParam := c.RunParams().ActualNet()
	if netParam != "bridge" {
		args = append(args, "--net", netParam)
	}
	// OomKillDisable
	if c.RunParams().OomKillDisable {
		args = append(args, "--oom-kill-disable")
	}
	// PID
	if len(c.RunParams().Pid()) > 0 {
		args = append(args, "--pid", c.RunParams().Pid())
	}
	// Privileged
	if c.RunParams().Privileged {
		args = append(args, "--privileged")
	}
	// Publish
	for _, port := range c.RunParams().Publish() {
		args = append(args, "--publish", port)
	}
	// PublishAll
	if c.RunParams().PublishAll {
		args = append(args, "--publish-all")
	}
	// ReadOnly
	if c.RunParams().ReadOnly {
		args = append(args, "--read-only")
	}
	// Restart
	if len(c.RunParams().Restart()) > 0 {
		args = append(args, "--restart", c.RunParams().Restart())
	}
	// Rm
	if c.RunParams().Rm {
		args = append(args, "--rm")
	}
	// SecurityOpt
	for _, securityOpt := range c.RunParams().SecurityOpt() {
		args = append(args, "--security-opt", securityOpt)
	}
	// SigProxy
	if c.RunParams().SigProxy.Falsy() {
		args = append(args, "--sig-proxy=false")
	}
	// StopSignal
	if len(c.RunParams().StopSignal()) > 0 {
		args = append(args, "--stop-signal", c.RunParams().StopSignal())
	}
	// Tty
	if c.RunParams().Tty {
		args = append(args, "--tty")
	}
	// Ulimit
	for _, ulimit := range c.RunParams().Ulimit() {
		args = append(args, "--ulimit", ulimit)
	}
	// User
	if len(c.RunParams().User()) > 0 {
		args = append(args, "--user", c.RunParams().User())
	}
	// Uts
	if len(c.RunParams().Uts()) > 0 {
		args = append(args, "--uts", c.RunParams().Uts())
	}
	// Volumes
	for _, volume := range c.RunParams().ActualVolume() {
		args = append(args, "--volume", volume)
	}
	// VolumesFrom
	for _, volumesFrom := range c.RunParams().VolumesFrom() {
		volumesFromParts := strings.Split(volumesFrom, ":")
		volumesFromName := volumesFromParts[0]
		if !includes(excluded, volumesFromName) {
			volumesFromParts[0] = cfg.Container(volumesFromName).ActualName()
			args = append(args, "--volumes-from", strings.Join(volumesFromParts, ":"))
		}
	}
	// Workdir
	if len(c.RunParams().Workdir()) > 0 {
		args = append(args, "--workdir", c.RunParams().Workdir())
	}
	// Name
	args = append(args, "--name", c.ActualName())
	// Image
	args = append(args, c.Image())
	// Command
	if len(cmds) > 0 {
		args = append(args, cmds...)
	} else {
		args = append(args, c.RunParams().Cmd()...)
	}
	return args
}

// Start container
func (c *container) Start(excluded []string) {
	if !c.Unique() && c.Exists() {
		if !c.running() {
			executeHook(c.Hooks().PreStart(), c.ActualName())
			fmt.Fprintf(c.CommandsOut(), "Starting container %s ...\n", c.ActualName())
			args := []string{"start"}
			if c.StartParams().Attach {
				args = append(args, "--attach")
			}
			if c.StartParams().Interactive {
				args = append(args, "--interactive")
			}
			args = append(args, c.ActualName())
			wg := c.executePostStartHook()
			executeCommand("docker", args, c.CommandsOut(), c.CommandsErr())
			wg.Wait()
		}
	} else {
		c.Run([]string{}, excluded)
	}
}

// Kill container
func (c *container) Kill() {
	for _, name := range c.InstancesOfStatus("running") {
		executeHook(c.Hooks().PreStop(), name)
		fmt.Fprintf(c.CommandsOut(), "Killing container %s ...\n", name)
		args := []string{"kill", name}
		executeCommand("docker", args, c.CommandsOut(), c.CommandsErr())
		executeHook(c.Hooks().PostStop(), name)
	}
}

// Stop container
func (c *container) Stop() {
	for _, name := range c.InstancesOfStatus("running") {
		executeHook(c.Hooks().PreStop(), name)
		fmt.Fprintf(c.CommandsOut(), "Stopping container %s ...\n", name)
		args := []string{"stop", name}
		executeCommand("docker", args, c.CommandsOut(), c.CommandsErr())
		executeHook(c.Hooks().PostStop(), name)
	}
}

// Pause container
func (c *container) Pause() {
	for _, name := range c.InstancesOfStatus("running") {
		fmt.Fprintf(c.CommandsOut(), "Pausing container %s ...\n", name)
		args := []string{"pause", name}
		executeCommand("docker", args, c.CommandsOut(), c.CommandsErr())
	}
}

// Unpause container
func (c *container) Unpause() {
	for _, name := range c.InstancesOfStatus("paused") {
		fmt.Fprintf(c.CommandsOut(), "Unpausing container %s ...\n", name)
		args := []string{"unpause", name}
		executeCommand("docker", args, c.CommandsOut(), c.CommandsErr())
	}
}

// Exec command in container
func (c *container) Exec(cmds []string) {
	runningInstances := c.InstancesOfStatus("running")
	if len(runningInstances) == 0 {
		c.Start([]string{})
		runningInstances = []string{c.ActualName()}
	}
	for _, name := range runningInstances {
		args := []string{"exec"}
		if c.ExecParams().Interactive {
			args = append(args, "--interactive")
		}
		if c.ExecParams().Tty {
			args = append(args, "--tty")
		}
		args = append(args, name)
		args = append(args, cmds...)
		executeCommand("docker", args, c.CommandsOut(), c.CommandsErr())
	}
}

// Remove container
func (c *container) Rm(force bool) {
	runningInstances := c.InstancesOfStatus("running")
	for _, name := range c.InstancesOfStatus("existing") {
		containerIsRunning := includes(runningInstances, name)
		if !force && containerIsRunning {
			fmt.Fprintf(c.CommandsOut(), "Cannot remove running container %s, use --force to remove anyway.\n", name)
			break
		}
		args := []string{"rm"}
		if force && containerIsRunning {
			executeHook(c.Hooks().PreStop(), name)
			args = append(args, "--force")
		}
		if c.RmParams().Volumes {
			fmt.Fprintf(c.CommandsOut(), "Removing container %s and its volumes ...\n", name)
			args = append(args, "--volumes")
		} else {
			fmt.Fprintf(c.CommandsOut(), "Removing container %s ...\n", name)
		}
		args = append(args, name)
		executeCommand("docker", args, c.CommandsOut(), c.CommandsErr())
		if force && containerIsRunning {
			executeHook(c.Hooks().PostStop(), name)
		}
		c.id = ""
	}
}

// Dump container logs
func (c *container) Logs(follow bool, since string, tail string) (sources []logSource) {
	for _, name := range c.InstancesOfStatus("existing") {
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
		args = append(args, name)
		_, stdout, stderr := executeCommandBackground("docker", args)
		sources = append(sources, logSource{
			Stdout: stdout,
			Stderr: stderr,
			Name:   name,
		})
	}
	return
}

// Push container
func (c *container) Push() {
	fmt.Fprintf(c.CommandsOut(), "Pushing image %s ...\n", c.Image())
	args := []string{"push", c.Image()}
	executeCommand("docker", args, c.CommandsOut(), c.CommandsErr())
}

func (c *container) Hooks() Hooks {
	return &c.hooks
}

// Pull image for container
func (c *container) PullImage() {
	fmt.Fprintf(c.CommandsOut(), "Pulling image %s ...\n", c.Image())
	args := []string{"pull", c.Image()}
	executeCommand("docker", args, c.CommandsOut(), c.CommandsErr())
}

func (c *container) PrefixedName() string {
	return cfg.Prefix() + c.Name()
}

func (c *container) InstancesOfStatus(status string) []string {
	if c.Unique() {
		args := []string{
			"ps",
			"--quiet",
			"--format='{{.Names}}'",
			"--filter='name=" + c.PrefixedName() + "-unique-'",
		}
		if status == "running" {
			args = append(args, "--filter='status=running'")
		} else if status == "paused" {
			args = append(args, "--filter='status=paused'")
		} else if status == "existing" {
			args = append(args, "--all")
		}
		output, err := commandOutput("docker", args)
		if err != nil || len(output) == 0 {
			return []string{}
		} else {
			return strings.Split(output, "\n")
		}
	} else {
		if status == "running" && c.running() {
			return []string{c.ActualName()}
		} else if status == "paused" && c.paused() {
			return []string{c.ActualName()}
		} else if status == "existing" && c.Exists() {
			return []string{c.ActualName()}
		}
	}
	return []string{}
}

func (c *container) SetCommandsOutput(stdout, stderr io.Writer) {
	c.stdout = stdout
	c.stderr = stderr
}

func (c *container) CommandsOut() io.Writer {
	if c.stdout == nil {
		return os.Stdout
	}
	return c.stdout
}

func (c *container) CommandsErr() io.Writer {
	if c.stderr == nil {
		return os.Stderr
	}
	return c.stderr
}

func (c *container) running() bool {
	if !c.Exists() {
		return false
	}
	return inspectBool(c.ID(), "{{.State.Running}}")
}

func (c *container) paused() bool {
	if !c.Exists() {
		return false
	}
	return inspectBool(c.ID(), "{{.State.Paused}}")
}

// Build image for container
func (c *container) buildImage(nocache bool) {
	executeHook(c.Hooks().PreBuild(), c.ActualName())
	fmt.Fprintf(c.CommandsOut(), "Building image %s ...\n", c.Image())
	args := []string{"build"}
	if nocache {
		args = append(args, "--no-cache")
	}
	args = append(args, "--rm", "--tag="+c.Image())
	if len(c.BuildParams().File()) > 0 {
		args = append(args, "--file="+filepath.FromSlash(c.BuildParams().Context()+"/"+c.BuildParams().File()))
	}
	for _, arg := range c.BuildParams().BuildArgs() {
		args = append(args, "--build-arg", arg)
	}

	args = append(args, c.BuildParams().Context())
	executeCommand("docker", args, c.CommandsOut(), c.CommandsErr())
	executeHook(c.Hooks().PostBuild(), c.ActualName())
}

// Return the image id of a tag, or an empty string if it doesn't exist
func imageIDFromTag(tag string) string {
	args := []string{"inspect", "--format={{.Id}}", tag}
	output, err := commandOutput("docker", args)
	if err != nil {
		return ""
	}
	return string(output)
}

// If the reference follows the `container:foo` pattern, return "foo"; otherwise, return an empty string
func containerReference(reference string) (name string) {
	if parts := strings.Split(reference, ":"); len(parts) == 2 && parts[0] == "container" {
		// We'll just assume here that the reference is a name, and not an id, even
		// though docker supports it, since we have no bullet-proof way to tell:
		// heuristics to detect whether it's an id could bring false positives, and
		// a lookup in the list of container names could bring false negatives
		name = parts[1]
	}
	return
}

// Transform an unmarshalled payload (YAML or JSON) of type slice or map to an slice of env-expanded "K=V" strings
func sliceOrMap2ExpandedSlice(value interface{}) []string {
	var result []string
	if value != nil {
		switch concreteValue := value.(type) {
		case []interface{}: // YAML or JSON
			for _, v := range concreteValue {
				result = append(result, expandEnv(fmt.Sprintf("%v", v)))
			}
		case map[interface{}]interface{}: // YAML
			for k, v := range concreteValue {
				result = append(result, expandEnv(fmt.Sprintf("%v", k))+"="+expandEnv(fmt.Sprintf("%v", v)))
			}
		case map[string]interface{}: // JSON
			for k, v := range concreteValue {
				result = append(result, expandEnv(k)+"="+expandEnv(fmt.Sprintf("%v", v)))
			}
		default:
			panic(StatusError{fmt.Errorf("unknown type: %v", value), 65})
		}
	}
	return result
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
