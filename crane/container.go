package crane

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/flynn/go-shlex"
)

type Container interface {
	ContainerInfo
	Exists() bool
	Running() bool
	Paused() bool
	Status() [][]string
	Provision(nocache bool)
	PullImage()
	Create(cmds []string)
	Run(cmds []string)
	Start()
	Kill()
	Stop()
	Pause()
	Unpause()
	Exec(cmds []string)
	Rm(force bool)
	Logs(follow bool, since string, tail string) (sources []LogSource)
	Push()
	Tag(tagName string)
	SetCommandsOutput(stdout, stderr io.Writer)
	CommandsOut() io.Writer
	CommandsErr() io.Writer
}

type ContainerInfo interface {
	Name() string
	PrefixedName() string
	ActualName(bool) string
	Image() string
	ID() string
	Dependencies() *Dependencies
	BuildParams() BuildParameters
	PushParameters() PushParameters
	PullParameters() PullParameters
	RunParams() RunParameters
	RmParams() RmParameters
	StartParams() StartParameters
	ExecParams() ExecParameters
	Hooks() Hooks
}

type container struct {
	id          string
	RawName     string
	RawImage    string          `json:"image" yaml:"image"`
	RawRequires []string        `json:"requires" yaml:"requires"`
	RawBuild    BuildParameters `json:"build" yaml:"build"`
	RawPush     PushParameters  `json:"push" yaml:"push"`
	RawPull     PullParameters  `json:"pull" yaml:"pull"`
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

type RegistryAwareParameters struct {
	RawRegistry     string `json:"registry" yaml:"registry"`
	RawOverrideUser string `json:"override_user" yaml:"override_user"`
}

type PushParameters struct {
	RegistryAwareParameters `yaml:",inline"`
	RawSkip                 bool `json:"skip" yaml:"skip"`
}

type PullParameters struct {
	RegistryAwareParameters `yaml:",inline"`
}

type RunParameters struct {
	RawAddHost           []string    `json:"add-host" yaml:"add-host"`
	BlkioWeight          int         `json:"blkio-weight" yaml:"blkio-weight"`
	RawBlkioWeightDevice []string    `json:"blkio-weight-device" yaml:"blkio-weight-device"`
	RawCapAdd            []string    `json:"cap-add" yaml:"cap-add"`
	RawCapDrop           []string    `json:"cap-drop" yaml:"cap-drop"`
	RawCgroupParent      string      `json:"cgroup-parent" yaml:"cgroup-parent"`
	CPUPeriod            int         `json:"cpu-period" yaml:"cpu-period"`
	CPUQuota             int         `json:"cpu-quota" yaml:"cpu-quota"`
	RawCidfile           string      `json:"cidfile" yaml:"cidfile"`
	CPUset               int         `json:"cpuset" yaml:"cpuset"`
	CPUShares            int         `json:"cpu-shares" yaml:"cpu-shares"`
	Detach               bool        `json:"detach" yaml:"detach"`
	RawDetachKeys        string      `json:"detach-keys" yaml:"detach-keys"`
	RawDevice            []string    `json:"device" yaml:"device"`
	RawDeviceReadBps     []string    `json:"device-read-bps" yaml:"device-read-bps"`
	RawDeviceReadIops    []string    `json:"device-read-iops" yaml:"device-read-iops"`
	RawDeviceWriteBps    []string    `json:"device-write-bps" yaml:"device-write-bps"`
	RawDeviceWriteIops   []string    `json:"device-rewritead-iops" yaml:"device-write-iops"`
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
	RawIp                string      `json:"ip" yaml:"ip"`
	RawIp6               string      `json:"ip6" yaml:"ip6"`
	RawIPC               string      `json:"ipc" yaml:"ipc"`
	RawIsolation         string      `json:"isolation" yaml:"isolation"`
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
	RawNetAlias          []string    `json:"net-alias" yaml:"net-alias"`
	OomKillDisable       bool        `json:"oom-kill-disable" yaml:"oom-kill-disable"`
	RawOomScoreAdj       string      `json:"oom-score-adj" yaml:"oom-score-adj"`
	RawPid               string      `json:"pid" yaml:"pid"`
	Privileged           bool        `json:"privileged" yaml:"privileged"`
	RawPublish           []string    `json:"publish" yaml:"publish"`
	PublishAll           bool        `json:"publish-all" yaml:"publish-all"`
	ReadOnly             bool        `json:"read-only" yaml:"read-only"`
	RawRestart           string      `json:"restart" yaml:"restart"`
	Rm                   bool        `json:"rm" yaml:"rm"`
	RawSecurityOpt       []string    `json:"security-opt" yaml:"security-opt"`
	RawShmSize           string      `json:"shm-size" yaml:"shm-size"`
	SigProxy             OptBool     `json:"sig-proxy" yaml:"sig-proxy"`
	RawStopSignal        string      `json:"stop-signal" yaml:"stop-signal"`
	RawTmpfs             []string    `json:"tmpfs" yaml:"tmpfs"`
	Tty                  bool        `json:"tty" yaml:"tty"`
	RawUlimit            []string    `json:"ulimit" yaml:"ulimit"`
	RawUser              string      `json:"user" yaml:"user"`
	RawUts               string      `json:"uts" yaml:"uts"`
	RawVolume            []string    `json:"volume" yaml:"volume"`
	RawVolumeDriver      string      `json:"volume-driver" yaml:"volume-driver"`
	RawVolumesFrom       []string    `json:"volumes-from" yaml:"volumes-from"`
	RawWorkdir           string      `json:"workdir" yaml:"workdir"`
	RawCmd               interface{} `json:"cmd" yaml:"cmd"`
}

type RmParameters struct {
	Volumes bool `json:"volumes" yaml:"volumes"`
}

type StartParameters struct {
	Attach        bool   `json:"attach" yaml:"attach"`
	RawDetachKeys string `json:"detach-keys" yaml:"detach-keys"`
	Interactive   bool   `json:"interactive" yaml:"interactive"`
}

type ExecParameters struct {
	Detach        bool   `json:"detach" yaml:"detach"`
	RawDetachKeys string `json:"detach-keys" yaml:"detach-keys"`
	Interactive   bool   `json:"interactive" yaml:"interactive"`
	Privileged    bool   `json:"privileged" yaml:"privileged"`
	Tty           bool   `json:"tty" yaml:"tty"`
	RawUser       string `json:"user" yaml:"user"`
}

type OptInt struct {
	Defined bool
	Value   int
}

type OptBool struct {
	Defined bool
	Value   bool
}

type LogSource struct {
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

func (c *container) PushParameters() PushParameters {
	return c.RawPush
}

func (c *container) PullParameters() PullParameters {
	return c.RawPull
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
		if includes(allowed, required) && !dependencies.includes(required) {
			dependencies.All = append(dependencies.All, required)
			dependencies.Requires = append(dependencies.Requires, required)
		}
	}
	if c.RunParams().Net() == "bridge" {
		// links are strict dependencies only on bridge networks
		for _, link := range c.RunParams().Link() {
			linkName := strings.Split(link, ":")[0]
			if includes(allowed, linkName) && !dependencies.includes(linkName) {
				dependencies.All = append(dependencies.All, linkName)
				dependencies.Link = append(dependencies.Link, linkName)
			}
		}
	}
	for _, volumesFrom := range c.RunParams().VolumesFrom() {
		volumesFromName := strings.Split(volumesFrom, ":")[0]
		if includes(allowed, volumesFromName) && !dependencies.includes(volumesFromName) {
			dependencies.All = append(dependencies.All, volumesFromName)
			dependencies.VolumesFrom = append(dependencies.VolumesFrom, volumesFromName)
		}
	}
	if net := containerReference(c.RunParams().Net()); net != "" {
		if includes(allowed, net) && !dependencies.includes(net) {
			dependencies.Net = net
			dependencies.All = append(dependencies.All, net)
		}
	}
	if ipc := containerReference(c.RunParams().IPC()); ipc != "" {
		if includes(allowed, ipc) && !dependencies.includes(ipc) {
			dependencies.IPC = ipc
			dependencies.All = append(dependencies.All, ipc)
		}
	}
	return dependencies
}

func (c *container) Name() string {
	return expandEnv(c.RawName)
}

func (c *container) ActualName(adHoc bool) string {
	if adHoc {
		return c.PrefixedName() + "-" + cfg.UniqueID()
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

func (r RegistryAwareParameters) Registry() string {
	return expandEnv(r.RawRegistry)
}

func (r RegistryAwareParameters) ImageWithRegistry(name string) string {
	return r.Registry() + "/" + name
}

func (r RegistryAwareParameters) findUserIndex(nameParts []string) int {
	maybeRepoIndex := len(nameParts) - 2
	maybeRepo := nameParts[maybeRepoIndex]
	if strings.Contains(maybeRepo, ".") {
		return maybeRepoIndex + 1
	}
	return maybeRepoIndex
}

func (r RegistryAwareParameters) OverrideUser() string {
	return expandEnv(r.RawOverrideUser)
}

func (r RegistryAwareParameters) OverrideImageName(image string) string {
	if strings.Index(image, "/") == -1 {
		return r.OverrideUser() + "/" + image
	}
	nameParts := strings.Split(image, "/")
	index := r.findUserIndex(nameParts)
	if index == len(nameParts)-1 {
		nameParts = append(nameParts, nameParts[index])
	}
	nameParts[index] = r.OverrideUser()
	fmt.Fprintf(os.Stdout, "Using override user %s\n", r.OverrideUser())
	return strings.Join(nameParts, "/")
}

func (p PushParameters) Skip() bool {
	return p.RawSkip
}

func (p PullParameters) CanBePulled() bool {
	return len(p.Registry()) > 0 || len(p.OverrideUser()) > 0
}

func (b BuildParameters) BuildArgs() []string {
	return sliceOrMap2ExpandedSlice(b.RawBuildArgs)
}

func (r RegistryAwareParameters) Registry() string {
	return expandEnv(r.RawRegistry)
}

func (r RegistryAwareParameters) ImageWithRegistry(name string) string {
	return r.Registry() + "/" + name
}

func (r RegistryAwareParameters) findUserIndex(nameParts []string) int {
	maybeRepoIndex := len(nameParts) - 2
	maybeRepo := nameParts[maybeRepoIndex]
	if strings.Contains(maybeRepo, ".") {
		return maybeRepoIndex + 1
	}
	return maybeRepoIndex
}

func (r RegistryAwareParameters) OverrideUser() string {
	return expandEnv(r.RawOverrideUser)
}

func (r RegistryAwareParameters) OverrideImageName(image string) string {
	if strings.Index(image, "/") == -1 {
		return r.OverrideUser() + "/" + image
	}
	nameParts := strings.Split(image, "/")
	index := r.findUserIndex(nameParts)
	if index == len(nameParts)-1 {
		nameParts = append(nameParts, nameParts[index])
	}
	nameParts[index] = r.OverrideUser()
	fmt.Fprintf(os.Stdout, "Using override user %s\n", r.OverrideUser())
	return strings.Join(nameParts, "/")
}

func (p PushParameters) Skip() bool {
	return p.RawSkip
}

func (p PullParameters) CanBePulled() bool {
	return len(p.Registry()) > 0 || len(p.OverrideUser()) > 0
}

func (r RunParameters) AddHost() []string {
	var addHost []string
	for _, rawAddHost := range r.RawAddHost {
		addHost = append(addHost, expandEnv(rawAddHost))
	}
	return addHost
}

func (r RunParameters) BlkioWeightDevice() []string {
	var blkioWeightDevice []string
	for _, rawBlkioWeightDevice := range r.RawBlkioWeightDevice {
		blkioWeightDevice = append(blkioWeightDevice, expandEnv(rawBlkioWeightDevice))
	}
	return blkioWeightDevice
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

func (r RunParameters) DetachKeys() string {
	return expandEnv(r.RawDetachKeys)
}

func (r RunParameters) Device() []string {
	var device []string
	for _, rawDevice := range r.RawDevice {
		device = append(device, expandEnv(rawDevice))
	}
	return device
}

func (r RunParameters) DeviceReadBps() []string {
	var deviceReadBps []string
	for _, rawDeviceReadBps := range r.RawDeviceReadBps {
		deviceReadBps = append(deviceReadBps, expandEnv(rawDeviceReadBps))
	}
	return deviceReadBps
}

func (r RunParameters) DeviceReadIops() []string {
	var deviceReadIops []string
	for _, rawDeviceReadIops := range r.RawDeviceReadIops {
		deviceReadIops = append(deviceReadIops, expandEnv(rawDeviceReadIops))
	}
	return deviceReadIops
}

func (r RunParameters) DeviceWriteBps() []string {
	var deviceWriteBps []string
	for _, rawDeviceWriteBps := range r.RawDeviceWriteBps {
		deviceWriteBps = append(deviceWriteBps, expandEnv(rawDeviceWriteBps))
	}
	return deviceWriteBps
}

func (r RunParameters) DeviceWriteIops() []string {
	var deviceWriteIops []string
	for _, rawDeviceWriteIops := range r.RawDeviceWriteIops {
		deviceWriteIops = append(deviceWriteIops, expandEnv(rawDeviceWriteIops))
	}
	return deviceWriteIops
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

func (r RunParameters) Ip() string {
	return expandEnv(r.RawIp)
}

func (r RunParameters) Ip6() string {
	return expandEnv(r.RawIp6)
}

func (r RunParameters) IPC() string {
	return expandEnv(r.RawIPC)
}

func (r RunParameters) Isolation() string {
	return expandEnv(r.RawIsolation)
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
		if includes(allowed, netContainer) {
			return "container:" + cfg.Container(netContainer).ActualName(false)
		}
	} else {
		if includes(cfg.NetworkNames(), netParam) {
			return cfg.Network(netParam).ActualName()
		}
	}
	return netParam
}

func (r RunParameters) NetAlias() []string {
	var netAlias []string
	for _, rawNetAlias := range r.RawNetAlias {
		netAlias = append(netAlias, expandEnv(rawNetAlias))
	}
	return netAlias
}

func (r RunParameters) OomScoreAdj() string {
	return expandEnv(r.RawOomScoreAdj)
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

func (r RunParameters) ShmSize() string {
	return expandEnv(r.RawShmSize)
}

func (r RunParameters) StopSignal() string {
	return expandEnv(r.RawStopSignal)
}

func (r RunParameters) Tmpfs() []string {
	var tmpfs []string
	for _, rawTmpfs := range r.RawTmpfs {
		tmpfs = append(tmpfs, expandEnv(rawTmpfs))
	}
	return tmpfs
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

func (r RunParameters) VolumeDriver() string {
	return expandEnv(r.RawVolumeDriver)
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

func (s StartParameters) DetachKeys() string {
	return expandEnv(s.RawDetachKeys)
}

func (e ExecParameters) DetachKeys() string {
	return expandEnv(e.RawDetachKeys)
}

func (e ExecParameters) User() string {
	return expandEnv(e.RawUser)
}

func (c *container) ID() string {
	if len(c.id) == 0 {
		// `docker inspect` works for both image and containers, make sure this is a
		// container payload we get back, otherwise we might end up getting the ID
		// of the image of the same name.
		c.id = inspectString(c.ActualName(false), "{{if .State}}{{.Id}}{{else}}{{end}}")
	}
	return c.id
}

func (c *container) Exists() bool {
	return c.ID() != ""
}

func (c *container) Status() [][]string {
	rows := [][]string{}
	if !c.Exists() {
		fields := []string{c.ActualName(false), c.Image(), "-", "-", "-", "-", "-"}
		rows = append(rows, fields)
	} else {
		name := c.ActualName(false)
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
func (c *container) Create(cmds []string) {
	adHoc := (len(cmds) > 0)
	if !adHoc {
		c.Rm(true)
	}
	fmt.Fprintf(c.CommandsOut(), "Creating container %s ...\n", c.ActualName(adHoc))

	args := append([]string{"create"}, c.createArgs(cmds)...)
	executeCommand("docker", args, c.CommandsOut(), c.CommandsErr())
}

// Run container, or start it if already existing
func (c *container) Run(cmds []string) {
	adHoc := (len(cmds) > 0)
	if !adHoc {
		c.Rm(true)
	}
	executeHook(c.Hooks().PreStart(), c.ActualName(adHoc))
	fmt.Fprintf(c.CommandsOut(), "Running container %s ...\n", c.ActualName(adHoc))

	args := []string{"run"}
	// Detach
	if !adHoc && c.RunParams().Detach {
		args = append(args, "--detach")
	}
	args = append(args, c.createArgs(cmds)...)
	wg := c.executePostStartHook(adHoc)
	executeCommand("docker", args, c.CommandsOut(), c.CommandsErr())
	wg.Wait()
}

func (c *container) executePostStartHook(adHoc bool) *sync.WaitGroup {

	var wg sync.WaitGroup

	if len(c.Hooks().PostStart()) > 0 {
		wg.Add(1)
		cmd, cmdOut, _ := executeCommandBackground("docker", []string{"events", "--filter", "event=start", "--filter", "container=" + c.ActualName(adHoc)})
		go func() {
			defer func() {
				handleRecoveredError(recover())
				wg.Done()
			}()
			r := bufio.NewReader(cmdOut)
			_, _, err := r.ReadLine()
			cmd.Process.Kill()
			if err != nil {
				printNoticef("Could not execute post-start hook for %s.", c.ActualName(adHoc))
			} else {
				executeHook(c.Hooks().PostStart(), c.ActualName(adHoc))
			}
		}()
	}

	return &wg
}

// Returns all the flags to be passed to `docker create`
func (c *container) createArgs(cmds []string) []string {
	adHoc := (len(cmds) > 0)
	args := []string{}
	// AddHost
	for _, addHost := range c.RunParams().AddHost() {
		args = append(args, "--add-host", addHost)
	}
	// BlkioWeight
	if c.RunParams().BlkioWeight > 0 {
		args = append(args, "--blkio-weight", strconv.Itoa(c.RunParams().BlkioWeight))
	}
	// BlkioWeightDevice
	for _, blkioWeightDevice := range c.RunParams().BlkioWeightDevice() {
		args = append(args, "--blkio-weight-device", blkioWeightDevice)
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
	// DetachKeys
	if len(c.RunParams().DetachKeys()) > 0 {
		args = append(args, "--detach-keys", c.RunParams().DetachKeys())
	}
	// Device
	for _, device := range c.RunParams().Device() {
		args = append(args, "--device", device)
	}
	// DeviceReadBps
	for _, deviceReadBps := range c.RunParams().DeviceReadBps() {
		args = append(args, "--device-read-bps", deviceReadBps)
	}
	// DeviceReadIops
	for _, deviceReadIops := range c.RunParams().DeviceReadIops() {
		args = append(args, "--device-read-iops", deviceReadIops)
	}
	// DeviceWriteBps
	for _, deviceWriteBps := range c.RunParams().DeviceWriteBps() {
		args = append(args, "--device-write-bps", deviceWriteBps)
	}
	// DeviceWriteIops
	for _, deviceWriteIops := range c.RunParams().DeviceWriteIops() {
		args = append(args, "--device-write-iops", deviceWriteIops)
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
	// Ip
	if !adHoc {
		if len(c.RunParams().Ip()) > 0 {
			args = append(args, "--ip", c.RunParams().Ip())
		}
	}
	// Ip6
	if !adHoc {
		if len(c.RunParams().Ip6()) > 0 {
			args = append(args, "--ip6", c.RunParams().Ip6())
		}
	}
	// IPC
	if len(c.RunParams().IPC()) > 0 {
		ipcContainer := containerReference(c.RunParams().IPC())
		if len(ipcContainer) > 0 {
			if includes(allowed, ipcContainer) {
				args = append(args, "--ipc", "container:"+cfg.Container(ipcContainer).ActualName(false))
			}
		} else {
			args = append(args, "--ipc", c.RunParams().IPC())
		}
	}
	// Isolation
	if len(c.RunParams().Isolation()) > 0 {
		args = append(args, "--isolation", c.RunParams().Isolation())
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
		if includes(allowed, linkName) {
			linkParts[0] = cfg.Container(linkName).ActualName(false)
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
	// NetAlias
	for _, netAlias := range c.RunParams().NetAlias() {
		args = append(args, "--net-alias", netAlias)
	}
	// OomKillDisable
	if c.RunParams().OomKillDisable {
		args = append(args, "--oom-kill-disable")
	}
	// OomScoreAdj
	if len(c.RunParams().OomScoreAdj()) > 0 {
		args = append(args, "--oom-score-adj", c.RunParams().OomScoreAdj())
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
	if !adHoc {
		for _, port := range c.RunParams().Publish() {
			args = append(args, "--publish", port)
		}
	}
	// PublishAll
	if !adHoc {
		if c.RunParams().PublishAll {
			args = append(args, "--publish-all")
		}
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
	if adHoc || c.RunParams().Rm {
		args = append(args, "--rm")
	}
	// SecurityOpt
	for _, securityOpt := range c.RunParams().SecurityOpt() {
		args = append(args, "--security-opt", securityOpt)
	}
	// ShmSize
	if len(c.RunParams().ShmSize()) > 0 {
		args = append(args, "--shm-size", c.RunParams().ShmSize())
	}
	// SigProxy
	if c.RunParams().SigProxy.Falsy() {
		args = append(args, "--sig-proxy=false")
	}
	// StopSignal
	if len(c.RunParams().StopSignal()) > 0 {
		args = append(args, "--stop-signal", c.RunParams().StopSignal())
	}
	// Tmpfs
	for _, tmpfs := range c.RunParams().Tmpfs() {
		args = append(args, "--tmpfs", tmpfs)
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
	// VolumeDriver
	if len(c.RunParams().VolumeDriver()) > 0 {
		args = append(args, "--volume-driver", c.RunParams().VolumeDriver())
	}
	// VolumesFrom
	for _, volumesFrom := range c.RunParams().VolumesFrom() {
		volumesFromParts := strings.Split(volumesFrom, ":")
		volumesFromName := volumesFromParts[0]
		if includes(allowed, volumesFromName) {
			volumesFromParts[0] = cfg.Container(volumesFromName).ActualName(false)
			args = append(args, "--volumes-from", strings.Join(volumesFromParts, ":"))
		}
	}
	// Workdir
	if len(c.RunParams().Workdir()) > 0 {
		args = append(args, "--workdir", c.RunParams().Workdir())
	}
	// Name
	args = append(args, "--name", c.ActualName(adHoc))
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
func (c *container) Start() {
	if c.Exists() {
		if !c.Running() {
			executeHook(c.Hooks().PreStart(), c.ActualName(false))
			fmt.Fprintf(c.CommandsOut(), "Starting container %s ...\n", c.ActualName(false))
			args := []string{"start"}
			if c.StartParams().Attach {
				args = append(args, "--attach")
			}
			if len(c.StartParams().DetachKeys()) > 0 {
				args = append(args, "--detach-keys", c.StartParams().DetachKeys())
			}
			if c.StartParams().Interactive {
				args = append(args, "--interactive")
			}
			args = append(args, c.ActualName(false))
			wg := c.executePostStartHook(false)
			executeCommand("docker", args, c.CommandsOut(), c.CommandsErr())
			wg.Wait()
		}
	} else {
		c.Run([]string{})
	}
}

// Kill container
func (c *container) Kill() {
	if c.Running() {
		name := c.ActualName(false)
		executeHook(c.Hooks().PreStop(), name)
		fmt.Fprintf(c.CommandsOut(), "Killing container %s ...\n", name)
		args := []string{"kill", name}
		executeCommand("docker", args, c.CommandsOut(), c.CommandsErr())
		executeHook(c.Hooks().PostStop(), name)
	}
}

// Stop container
func (c *container) Stop() {
	if c.Running() {
		name := c.ActualName(false)
		executeHook(c.Hooks().PreStop(), name)
		fmt.Fprintf(c.CommandsOut(), "Stopping container %s ...\n", name)
		args := []string{"stop", name}
		executeCommand("docker", args, c.CommandsOut(), c.CommandsErr())
		executeHook(c.Hooks().PostStop(), name)
	}
}

// Pause container
func (c *container) Pause() {
	if c.Running() {
		name := c.ActualName(false)
		fmt.Fprintf(c.CommandsOut(), "Pausing container %s ...\n", name)
		args := []string{"pause", name}
		executeCommand("docker", args, c.CommandsOut(), c.CommandsErr())
	}
}

// Unpause container
func (c *container) Unpause() {
	if c.Paused() {
		name := c.ActualName(false)
		fmt.Fprintf(c.CommandsOut(), "Unpausing container %s ...\n", name)
		args := []string{"unpause", name}
		executeCommand("docker", args, c.CommandsOut(), c.CommandsErr())
	}
}

// Exec command in container
func (c *container) Exec(cmds []string) {
	name := c.ActualName(false)
	if !c.Running() {
		c.Start()
	}
	args := []string{"exec"}
	if c.ExecParams().Detach {
		args = append(args, "--detach")
	}
	if len(c.ExecParams().DetachKeys()) > 0 {
		args = append(args, "--detach-keys", c.ExecParams().DetachKeys())
	}
	if c.ExecParams().Privileged {
		args = append(args, "--privileged")
	}
	if c.ExecParams().Interactive {
		args = append(args, "--interactive")
	}
	if c.ExecParams().Tty {
		args = append(args, "--tty")
	}
	if len(c.ExecParams().User()) > 0 {
		args = append(args, "--user", c.ExecParams().User())
	}
	args = append(args, name)
	args = append(args, cmds...)
	executeCommand("docker", args, c.CommandsOut(), c.CommandsErr())
}

// Remove container
func (c *container) Rm(force bool) {
	if c.Exists() {
		name := c.ActualName(false)
		containerIsRunning := c.Running()
		if !force && containerIsRunning {
			fmt.Fprintf(c.CommandsOut(), "Cannot remove running container %s, use --force to remove anyway.\n", name)
			return
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
func (c *container) Logs(follow bool, since string, tail string) (sources []LogSource) {
	if c.Exists() {
		name := c.ActualName(false)
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
		sources = append(sources, LogSource{
			Stdout: stdout,
			Stderr: stderr,
			Name:   name,
		})
	}
	return
}

func (c *container) imageTag(image string, tag string) {
	fmt.Fprintf(c.CommandsOut(), "Tagging image %s as %s...\n", image, tag)
	var args []string
	if validateDockerClientAbove([]int{1, 10}) {
		args = []string{"tag", image, tag}
	} else {
		args = []string{"tag", "--force", image, tag}
	}
	executeCommand("docker", args, c.CommandsOut(), c.CommandsErr())
}

// Tag container image
func (c *container) Tag(tag string) {
	c.imageTag(c.Image(), tag)
}

// Push container
func (c *container) Push() {
	fmt.Fprintf(c.CommandsOut(), "Pushing image %s ...", c.Image())
	if c.PushParameters().Skip() {
		fmt.Fprintf(c.CommandsOut(), " Skipping\n")
		return
	}
	fmt.Fprintf(c.CommandsOut(), "\n")
	image := c.Image()
	if len(c.PushParameters().OverrideUser()) > 0 {
		image = c.PushParameters().OverrideImageName(image)
		c.Tag(image)
	}
	if len(c.PushParameters().Registry()) > 0 {
		image = c.PushParameters().ImageWithRegistry(image)
		c.Tag(image)
	}
	args := []string{"push", image}
	executeCommand("docker", args, c.CommandsOut(), c.CommandsErr())
}

func (c *container) Hooks() Hooks {
	return &c.hooks
}

// Pull image for container
func (c *container) PullImage() {
	image := c.Image()
	if len(c.PullParameters().OverrideUser()) > 0 {
		image = c.PullParameters().OverrideImageName(image)
	}
	if len(c.PullParameters().Registry()) > 0 {
		image = c.PullParameters().ImageWithRegistry(image)
	}
	fmt.Fprintf(c.CommandsOut(), "Pulling image %s ...\n", image)
	args := []string{"pull", image}
	executeCommand("docker", args, c.CommandsOut(), c.CommandsErr())
	if image != c.Image() {
		c.imageTag(image, c.Image())
	}
}

func (c *container) PrefixedName() string {
	return cfg.Prefix() + c.Name()
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

func (c *container) Running() bool {
	if !c.Exists() {
		return false
	}
	return inspectBool(c.ID(), "{{.State.Running}}")
}

func (c *container) Paused() bool {
	if !c.Exists() {
		return false
	}
	return inspectBool(c.ID(), "{{.State.Paused}}")
}

// Build image for container
func (c *container) buildImage(nocache bool) {
	executeHook(c.Hooks().PreBuild(), c.ActualName(false))
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
	executeHook(c.Hooks().PostBuild(), c.ActualName(false))
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
	expandedStringOrPanic := func(v interface{}) string {
		switch concreteValue := v.(type) {
		case []interface{}: // YAML or JSON
			panic(StatusError{fmt.Errorf("unknown type: %v", v), 65})
		case map[interface{}]interface{}: // YAML
			panic(StatusError{fmt.Errorf("unknown type: %v", v), 65})
		case map[string]interface{}: // JSON
			panic(StatusError{fmt.Errorf("unknown type: %v", v), 65})
		default:
			return expandEnv(fmt.Sprintf("%v", concreteValue))
		}
	}
	if value != nil {
		switch concreteValue := value.(type) {
		case []interface{}: // YAML or JSON
			for _, v := range concreteValue {
				result = append(result, expandedStringOrPanic(v))
			}
		case map[interface{}]interface{}: // YAML
			for k, v := range concreteValue {
				result = append(result, expandedStringOrPanic(k)+"="+expandedStringOrPanic(v))
			}
		case map[string]interface{}: // JSON
			for k, v := range concreteValue {
				result = append(result, expandEnv(k)+"="+expandedStringOrPanic(v))
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
