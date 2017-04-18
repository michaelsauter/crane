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
	"runtime"
	"strconv"
	"strings"
	"sync"
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
	Run(cmds []string, detach bool)
	Start()
	Kill()
	Stop()
	Pause()
	Unpause()
	Exec(cmds []string, privileged bool, user string)
	Rm(force bool, volumes bool)
	Logs(follow bool, since string, tail string) (sources []LogSource)
	Push()
	SetCommandsOutput(stdout, stderr io.Writer)
	CommandsOut() io.Writer
	CommandsErr() io.Writer
	VolumeSources() []string
	Net() string
}

type ContainerInfo interface {
	Name() string
	PrefixedName() string
	ActualName(bool) string
	Image() string
	ID() string
	Dependencies() *Dependencies
	BuildParams() BuildParameters
	Hooks() Hooks
}

type container struct {
	id                   string
	RawName              string
	RawImage             string                `json:"image" yaml:"image"`
	RawRequires          []string              `json:"requires" yaml:"requires"`
	RawDependsOn         []string              `json:"depends_on" yaml:"depends_on"`
	RawBuild             BuildParameters       `json:"build" yaml:"build"`
	RawAddHost           []string              `json:"add-host" yaml:"add-host"`
	RawExtraHosts        []string              `json:"extra-hosts" yaml:"extra-hosts"`
	BlkioWeight          int                   `json:"blkio-weight" yaml:"blkio-weight"`
	RawBlkioWeightDevice []string              `json:"blkio-weight-device" yaml:"blkio-weight-device"`
	RawCapAdd            []string              `json:"cap-add" yaml:"cap-add"`
	RawCapDrop           []string              `json:"cap-drop" yaml:"cap-drop"`
	RawCap_Add           []string              `json:"cap_add" yaml:"cap_add"`
	RawCap_Drop          []string              `json:"cap_drop" yaml:"cap_drop"`
	RawCgroupParent      string                `json:"cgroup-parent" yaml:"cgroup-parent"`
	RawCgroup_Parent     string                `json:"cgroup_parent" yaml:"cgroup_parent"`
	CPUPeriod            int                   `json:"cpu-period" yaml:"cpu-period"`
	CPUQuota             int                   `json:"cpu-quota" yaml:"cpu-quota"`
	RawCidfile           string                `json:"cidfile" yaml:"cidfile"`
	CPUset               int                   `json:"cpuset" yaml:"cpuset"`
	CPUShares            int                   `json:"cpu-shares" yaml:"cpu-shares"`
	Detach               bool                  `json:"detach" yaml:"detach"`
	RawDetachKeys        string                `json:"detach-keys" yaml:"detach-keys"`
	RawDevice            []string              `json:"device" yaml:"device"`
	RawDevices           []string              `json:"devices" yaml:"devices"`
	RawDeviceReadBps     []string              `json:"device-read-bps" yaml:"device-read-bps"`
	RawDeviceReadIops    []string              `json:"device-read-iops" yaml:"device-read-iops"`
	RawDeviceWriteBps    []string              `json:"device-write-bps" yaml:"device-write-bps"`
	RawDeviceWriteIops   []string              `json:"device-rewritead-iops" yaml:"device-write-iops"`
	RawDNS               []string              `json:"dns" yaml:"dns"`
	RawDNSOpt            []string              `json:"dns-opt" yaml:"dns-opt"`
	RawDNSSearch         []string              `json:"dns-search" yaml:"dns-search"`
	RawDNS_Search        []string              `json:"dns_search" yaml:"dns_search"`
	RawEntrypoint        string                `json:"entrypoint" yaml:"entrypoint"`
	RawEnv               interface{}           `json:"env" yaml:"env"`
	RawEnvironment       interface{}           `json:"environment" yaml:"environment"`
	RawEnvFile           []string              `json:"env-file" yaml:"env-file"`
	RawEnv_File          []string              `json:"env_file" yaml:"env_file"`
	RawExpose            []string              `json:"expose" yaml:"expose"`
	RawGroupAdd          []string              `json:"group-add" yaml:"group-add"`
	RawGroup_Add         []string              `json:"group_add" yaml:"group_add"`
	RawHealthcheck       HealthcheckParameters `json:"healthcheck" yaml:"healthcheck"`
	RawHealthCmd         string                `json:"health-cmd" yaml:"health-cmd"`
	RawHealthInterval    string                `json:"health-interval" yaml:"health-interval"`
	HealthRetries        int                   `json:"health-retries" yaml:"health-retries"`
	RawHealthTimeout     string                `json:"health-timeout" yaml:"health-timeout"`
	RawHostname          string                `json:"hostname" yaml:"hostname"`
	Interactive          bool                  `json:"interactive" yaml:"interactive"`
	Stdin_Open           bool                  `json:"stdin_open" yaml:"stdin_open"`
	RawIp                string                `json:"ip" yaml:"ip"`
	RawIp6               string                `json:"ip6" yaml:"ip6"`
	RawIPC               string                `json:"ipc" yaml:"ipc"`
	RawIsolation         string                `json:"isolation" yaml:"isolation"`
	RawKernelMemory      string                `json:"kernel-memory" yaml:"kernel-memory"`
	RawLabel             interface{}           `json:"label" yaml:"label"`
	RawLabels            interface{}           `json:"labels" yaml:"labels"`
	RawLabelFile         []string              `json:"label-file" yaml:"label-file"`
	RawLink              []string              `json:"link" yaml:"link"`
	RawLinks             []string              `json:"links" yaml:"links"`
	RawLogDriver         string                `json:"log-driver" yaml:"log-driver"`
	RawLogOpt            []string              `json:"log-opt" yaml:"log-opt"`
	RawLogging           LoggingParameters     `json:"logging" yaml:"logging"`
	RawLxcConf           []string              `json:"lxc-conf" yaml:"lxc-conf"`
	RawMacAddress        string                `json:"mac-address" yaml:"mac-address"`
	RawMac_Address       string                `json:"mac_address" yaml:"mac_address"`
	RawMemory            string                `json:"memory" yaml:"memory"`
	RawMemoryReservation string                `json:"memory-reservation" yaml:"memory-reservation"`
	RawMemorySwap        string                `json:"memory-swap" yaml:"memory-swap"`
	MemorySwappiness     OptInt                `json:"memory-swappiness" yaml:"memory-swappiness"`
	RawNet               string                `json:"net" yaml:"net"`
	RawNetwork_Mode      string                `json:"network_mode" yaml:"network_mode"`
	RawNetAlias          []string              `json:"net-alias" yaml:"net-alias"`
	RawNetworks          []string              `json:"networks" yaml:"networks"`
	NoHealthcheck        bool                  `json:"no-healthcheck" yaml:"no-healthcheck"`
	OomKillDisable       bool                  `json:"oom-kill-disable" yaml:"oom-kill-disable"`
	RawOomScoreAdj       string                `json:"oom-score-adj" yaml:"oom-score-adj"`
	RawPid               string                `json:"pid" yaml:"pid"`
	Privileged           bool                  `json:"privileged" yaml:"privileged"`
	RawPublish           []string              `json:"publish" yaml:"publish"`
	RawPorts             []string              `json:"ports" yaml:"ports"`
	PublishAll           bool                  `json:"publish-all" yaml:"publish-all"`
	ReadOnly             bool                  `json:"read-only" yaml:"read-only"`
	Read_Only            bool                  `json:"read_only" yaml:"read_only"`
	RawRestart           string                `json:"restart" yaml:"restart"`
	RawRm                bool                  `json:"rm" yaml:"rm"`
	RawSecurityOpt       []string              `json:"security-opt" yaml:"security-opt"`
	RawSecurity_Opt      []string              `json:"security_opt" yaml:"security_opt"`
	RawShmSize           string                `json:"shm-size" yaml:"shm-size"`
	RawShm_Size          string                `json:"shm_size" yaml:"shm_size"`
	SigProxy             OptBool               `json:"sig-proxy" yaml:"sig-proxy"`
	RawStopSignal        string                `json:"stop-signal" yaml:"stop-signal"`
	RawStop_Signal       string                `json:"stop_signal" yaml:"stop_signal"`
	RawStopTimeout       string                `json:"stop-timeout" yaml:"stop-timeout"`
	RawStop_Grace_Period string                `json:"stop_grace_period" yaml:"stop_grace_period"`
	RawSysctl            interface{}           `json:"sysctl" yaml:"sysctl"`
	RawSysctls           interface{}           `json:"sysctls" yaml:"sysctls"`
	RawTmpfs             []string              `json:"tmpfs" yaml:"tmpfs"`
	Tty                  bool                  `json:"tty" yaml:"tty"`
	RawUlimit            []string              `json:"ulimit" yaml:"ulimit"`
	RawUser              string                `json:"user" yaml:"user"`
	RawUserns            string                `json:"userns" yaml:"userns"`
	RawUserns_Mode       string                `json:"userns_mode" yaml:"userns_mode"`
	RawUts               string                `json:"uts" yaml:"uts"`
	RawVolume            []string              `json:"volume" yaml:"volume"`
	RawVolumes           []string              `json:"volumes" yaml:"volumes"`
	RawVolumeDriver      string                `json:"volume-driver" yaml:"volume-driver"`
	RawVolume_Driver     string                `json:"volume_driver" yaml:"volume_driver"`
	RawVolumesFrom       []string              `json:"volumes-from" yaml:"volumes-from"`
	RawWorkdir           string                `json:"workdir" yaml:"workdir"`
	RawWorking_Dir       string                `json:"working_dir" yaml:"working_dir"`
	RawCmd               interface{}           `json:"command" yaml:"command"`
	hooks                hooks
	stdout               io.Writer
	stderr               io.Writer
}

type BuildParameters struct {
	RawContext    string      `json:"context" yaml:"context"`
	RawFile       string      `json:"file" yaml:"file"`
	RawDockerfile string      `json:"dockerfile" yaml:"dockerfile"`
	RawBuildArgs  interface{} `json:"build-arg" yaml:"build-arg"`
	RawArgs       interface{} `json:"args" yaml:"args"`
}

type LoggingParameters struct {
	RawDriver  string      `json:"driver" yaml:"driver"`
	RawOptions interface{} `json:"options" yaml:"options"`
}

type HealthcheckParameters struct {
	RawTest     string `json:"test" yaml:"test"`
	RawInterval string `json:"interval" yaml:"interval"`
	RawTimeout  string `json:"timeout" yaml:"timeout"`
	Retries     int    `json:"retries" yaml:"retries"`
	Disable     bool   `json:"disable" yaml:"disable"`
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

func (c *container) LoggingParams() LoggingParameters {
	return c.RawLogging
}

func (c *container) HealthcheckParams() HealthcheckParameters {
	return c.RawHealthcheck
}

func (c *container) Dependencies() *Dependencies {
	dependencies := &Dependencies{}
	for _, required := range c.Requires() {
		if includes(allowed, required) && !dependencies.includes(required) {
			dependencies.All = append(dependencies.All, required)
			dependencies.Requires = append(dependencies.Requires, required)
		}
	}
	if c.Net() == "bridge" {
		// links are strict dependencies only on bridge networks
		for _, link := range c.Link() {
			linkName := strings.Split(link, ":")[0]
			if includes(allowed, linkName) && !dependencies.includes(linkName) {
				dependencies.All = append(dependencies.All, linkName)
				dependencies.Link = append(dependencies.Link, linkName)
			}
		}
	}
	for _, volumesFrom := range c.VolumesFrom() {
		volumesFromName := strings.Split(volumesFrom, ":")[0]
		if includes(allowed, volumesFromName) && !dependencies.includes(volumesFromName) {
			dependencies.All = append(dependencies.All, volumesFromName)
			dependencies.VolumesFrom = append(dependencies.VolumesFrom, volumesFromName)
		}
	}
	if net := containerReference(c.Net()); net != "" {
		if includes(allowed, net) && !dependencies.includes(net) {
			dependencies.Net = net
			dependencies.All = append(dependencies.All, net)
		}
	}
	if ipc := containerReference(c.IPC()); ipc != "" {
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
	if len(c.RawImage) == 0 {
		return c.ActualName(false)
	}

	image := expandEnv(c.RawImage)

	// Return if no global tag given or image is a digest
	if len(cfg.Tag()) == 0 || strings.Contains(image, "@") {
		return image
	}

	// Replace image tag with global tag
	startOfTag := strings.LastIndex(image, ":")
	if startOfTag != -1 {
		// Check that the colon does not refer to a port.
		// Tags must not contain slashes.
		tagPart := image[startOfTag+1:]
		slashPosition := strings.LastIndex(tagPart, "/")
		if slashPosition == -1 {
			image = image[:startOfTag]
		}
	}
	return image + ":" + cfg.Tag()
}

func (c *container) Requires() []string {
	var requires []string
	rawRequires := c.RawDependsOn
	if len(c.RawRequires) > 0 {
		rawRequires = c.RawRequires
	}
	for _, rawRequired := range rawRequires {
		requires = append(requires, expandEnv(rawRequired))
	}
	return requires
}

func (b BuildParameters) Context() string {
	return expandEnv(b.RawContext)
}

func (b BuildParameters) File() string {
	if len(b.RawFile) > 0 {
		return expandEnv(b.RawFile)
	}
	return expandEnv(b.RawDockerfile)
}

func (b BuildParameters) BuildArgs() []string {
	buildArgs := sliceOrMap2ExpandedSlice(b.RawBuildArgs)
	if len(buildArgs) == 0 {
		return sliceOrMap2ExpandedSlice(b.RawArgs)
	}
	return buildArgs
}

func (c *container) AddHost() []string {
	var addHost []string
	rawAddHost := c.RawExtraHosts
	if len(c.RawAddHost) > 0 {
		rawAddHost = c.RawAddHost
	}
	for _, raw := range rawAddHost {
		addHost = append(addHost, expandEnv(raw))
	}
	return addHost
}

func (c *container) BlkioWeightDevice() []string {
	var blkioWeightDevice []string
	for _, rawBlkioWeightDevice := range c.RawBlkioWeightDevice {
		blkioWeightDevice = append(blkioWeightDevice, expandEnv(rawBlkioWeightDevice))
	}
	return blkioWeightDevice
}

func (c *container) CapAdd() []string {
	var capAdd []string
	rawCapAdd := c.RawCap_Add
	if len(c.RawCapAdd) > 0 {
		rawCapAdd = c.RawCapAdd
	}
	for _, raw := range rawCapAdd {
		capAdd = append(capAdd, expandEnv(raw))
	}
	return capAdd
}

func (c *container) CapDrop() []string {
	var capDrop []string
	rawCapDrop := c.RawCap_Drop
	if len(c.RawCapDrop) > 0 {
		rawCapDrop = c.RawCap_Drop
	}
	for _, raw := range rawCapDrop {
		capDrop = append(capDrop, expandEnv(raw))
	}
	return capDrop
}

func (c *container) CgroupParent() string {
	if len(c.RawCgroupParent) > 0 {
		return expandEnv(c.RawCgroupParent)
	}
	return expandEnv(c.RawCgroup_Parent)
}

func (c *container) Cidfile() string {
	return expandEnv(c.RawCidfile)
}

func (c *container) DetachKeys() string {
	return expandEnv(c.RawDetachKeys)
}

func (c *container) Device() []string {
	var device []string
	rawDevice := c.RawDevices
	if len(c.RawDevice) > 0 {
		rawDevice = c.RawDevice
	}
	for _, rawDevice := range rawDevice {
		device = append(device, expandEnv(rawDevice))
	}
	return device
}

func (c *container) DeviceReadBps() []string {
	var deviceReadBps []string
	for _, rawDeviceReadBps := range c.RawDeviceReadBps {
		deviceReadBps = append(deviceReadBps, expandEnv(rawDeviceReadBps))
	}
	return deviceReadBps
}

func (c *container) DeviceReadIops() []string {
	var deviceReadIops []string
	for _, rawDeviceReadIops := range c.RawDeviceReadIops {
		deviceReadIops = append(deviceReadIops, expandEnv(rawDeviceReadIops))
	}
	return deviceReadIops
}

func (c *container) DeviceWriteBps() []string {
	var deviceWriteBps []string
	for _, rawDeviceWriteBps := range c.RawDeviceWriteBps {
		deviceWriteBps = append(deviceWriteBps, expandEnv(rawDeviceWriteBps))
	}
	return deviceWriteBps
}

func (c *container) DeviceWriteIops() []string {
	var deviceWriteIops []string
	for _, rawDeviceWriteIops := range c.RawDeviceWriteIops {
		deviceWriteIops = append(deviceWriteIops, expandEnv(rawDeviceWriteIops))
	}
	return deviceWriteIops
}

func (c *container) DNS() []string {
	var dns []string
	for _, rawDNS := range c.RawDNS {
		dns = append(dns, expandEnv(rawDNS))
	}
	return dns
}

func (c *container) DNSOpt() []string {
	var dnsOpt []string
	for _, rawDNSOpt := range c.RawDNSOpt {
		dnsOpt = append(dnsOpt, expandEnv(rawDNSOpt))
	}
	return dnsOpt
}

func (c *container) DNSSearch() []string {
	var dnsSearch []string
	rawDNSSearch := c.RawDNS_Search
	if len(c.RawDNSSearch) > 0 {
		rawDNSSearch = c.RawDNSSearch
	}
	for _, rawDNSSearch := range rawDNSSearch {
		dnsSearch = append(dnsSearch, expandEnv(rawDNSSearch))
	}
	return dnsSearch
}

func (c *container) Entrypoint() string {
	return expandEnv(c.RawEntrypoint)
}

func (c *container) Env() []string {
	env := sliceOrMap2ExpandedSlice(c.RawEnv)
	if len(env) == 0 {
		return sliceOrMap2ExpandedSlice(c.RawEnvironment)
	}
	return env
}

func (c *container) EnvFile() []string {
	var envFile []string
	rawEnvFile := c.RawEnv_File
	if len(c.RawEnvFile) > 0 {
		rawEnvFile = c.RawEnvFile
	}
	for _, rawEnvFile := range rawEnvFile {
		envFile = append(envFile, expandEnv(rawEnvFile))
	}
	return envFile
}

func (c *container) Expose() []string {
	var expose []string
	for _, rawExpose := range c.RawExpose {
		expose = append(expose, expandEnv(rawExpose))
	}
	return expose
}

func (c *container) GroupAdd() []string {
	var groupAdd []string
	rawGroupAdd := c.RawGroup_Add
	if len(c.RawGroupAdd) > 0 {
		rawGroupAdd = c.RawGroupAdd
	}
	for _, raw := range rawGroupAdd {
		groupAdd = append(groupAdd, expandEnv(raw))
	}
	return groupAdd
}

func (h HealthcheckParameters) Test() string {
	return expandEnv(h.RawTest)
}

func (h HealthcheckParameters) Interval() string {
	return expandEnv(h.RawInterval)
}

func (h HealthcheckParameters) Timeout() string {
	return expandEnv(h.RawTimeout)
}

func (c *container) HealthCmd() string {
	cmd := c.HealthcheckParams().Test()
	if len(c.RawLogDriver) > 0 {
		cmd = expandEnv(c.RawHealthCmd)
	}
	return cmd
}

func (c *container) HealthInterval() string {
	interval := c.HealthcheckParams().Interval()
	if len(c.RawHealthInterval) > 0 {
		interval = expandEnv(c.RawHealthInterval)
	}
	return interval
}

func (c *container) HealthTimeout() string {
	timeout := c.HealthcheckParams().Timeout()
	if len(c.RawHealthTimeout) > 0 {
		timeout = expandEnv(c.RawHealthTimeout)
	}
	return timeout
}

func (c *container) Hostname() string {
	return expandEnv(c.RawHostname)
}

func (c *container) Ip() string {
	return expandEnv(c.RawIp)
}

func (c *container) Ip6() string {
	return expandEnv(c.RawIp6)
}

func (c *container) IPC() string {
	return expandEnv(c.RawIPC)
}

func (c *container) Isolation() string {
	return expandEnv(c.RawIsolation)
}

func (c *container) KernelMemory() string {
	return expandEnv(c.RawKernelMemory)
}

func (c *container) Label() []string {
	label := sliceOrMap2ExpandedSlice(c.RawLabel)
	if len(label) == 0 {
		return sliceOrMap2ExpandedSlice(c.RawLabels)
	}
	return label
}

func (c *container) LabelFile() []string {
	var labelFile []string
	for _, rawLabelFile := range c.RawLabelFile {
		labelFile = append(labelFile, expandEnv(rawLabelFile))
	}
	return labelFile
}

func (c *container) Link() []string {
	var link []string
	rawLink := c.RawLinks
	if len(c.RawLink) > 0 {
		rawLink = c.RawLink
	}
	for _, raw := range rawLink {
		link = append(link, expandEnv(raw))
	}
	return link
}

func (l LoggingParameters) Options() []string {
	return sliceOrMap2ExpandedSlice(l.RawOptions)
}

func (l LoggingParameters) Driver() string {
	return expandEnv(l.RawDriver)
}

func (c *container) LogDriver() string {
	driver := c.LoggingParams().Driver()
	if len(c.RawLogDriver) > 0 {
		driver = expandEnv(c.RawLogDriver)
	}
	return driver
}

func (c *container) LogOpt() []string {
	var logOpt []string
	rawLogOpt := c.LoggingParams().Options()
	if len(c.RawLogOpt) > 0 {
		rawLogOpt = c.RawLogOpt
	}
	for _, raw := range rawLogOpt {
		logOpt = append(logOpt, expandEnv(raw))
	}
	return logOpt
}

func (c *container) LxcConf() []string {
	var lxcConf []string
	for _, rawLxcConf := range c.RawLxcConf {
		lxcConf = append(lxcConf, expandEnv(rawLxcConf))
	}
	return lxcConf
}

func (c *container) MacAddress() string {
	if len(c.RawMacAddress) > 0 {
		return expandEnv(c.RawMacAddress)
	}
	return expandEnv(c.RawMac_Address)
}

func (c *container) Memory() string {
	return expandEnv(c.RawMemory)
}

func (c *container) MemoryReservation() string {
	return expandEnv(c.RawMemoryReservation)
}

func (c *container) MemorySwap() string {
	return expandEnv(c.RawMemorySwap)
}

func (c *container) Net() string {
	rawNet := c.RawNetwork_Mode
	// docker-compose's "networks" actually works via "docker network connect".
	// This is not supported yet, but we can at least take the first network
	// until we fully support it.
	if len(c.RawNetworks) > 0 {
		rawNet = c.RawNetworks[0]
	}
	if len(c.RawNet) > 0 {
		rawNet = c.RawNet
	}
	// Fall back to the "default" network
	if len(rawNet) == 0 {
		return "default"
	}
	return expandEnv(rawNet)
}

func (c *container) ActualNet() string {
	netParam := c.Net()
	if netParam == "bridge" {
		return "bridge"
	} else if netParam == "none" {
		return "none"
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

func (c *container) NetAlias() []string {
	var netAlias []string
	for _, rawNetAlias := range c.RawNetAlias {
		netAlias = append(netAlias, expandEnv(rawNetAlias))
	}
	return netAlias
}

func (c *container) OomScoreAdj() string {
	return expandEnv(c.RawOomScoreAdj)
}

func (c *container) Pid() string {
	return expandEnv(c.RawPid)
}

func (c *container) Publish() []string {
	var publish []string
	rawPublish := c.RawPorts
	if len(c.RawPublish) > 0 {
		rawPublish = c.RawPublish
	}
	for _, raw := range rawPublish {
		publish = append(publish, expandEnv(raw))
	}
	return publish
}

func (c *container) Restart() string {
	return expandEnv(c.RawRestart)
}

func (c *container) SecurityOpt() []string {
	var securityOpt []string
	rawSecurityOpt := c.RawSecurity_Opt
	if len(c.RawSecurityOpt) > 0 {
		rawSecurityOpt = c.RawSecurityOpt
	}
	for _, raw := range rawSecurityOpt {
		securityOpt = append(securityOpt, expandEnv(raw))
	}
	return securityOpt
}

func (c *container) ShmSize() string {
	if len(c.RawShmSize) > 0 {
		return expandEnv(c.RawShmSize)
	}
	return expandEnv(c.RawShm_Size)
}

func (c *container) StopSignal() string {
	if len(c.RawStopSignal) > 0 {
		return expandEnv(c.RawStopSignal)
	}
	return expandEnv(c.RawStop_Signal)
}

func (c *container) StopTimeout() string {
	if len(c.RawStopTimeout) > 0 {
		return expandEnv(c.RawStopTimeout)
	}
	return expandEnv(c.RawStop_Grace_Period)
}

func (c *container) Sysctl() []string {
	sysctl := sliceOrMap2ExpandedSlice(c.RawSysctl)
	if len(sysctl) > 0 {
		return sysctl
	}
	return sliceOrMap2ExpandedSlice(c.RawSysctls)
}

func (c *container) Tmpfs() []string {
	var tmpfs []string
	for _, rawTmpfs := range c.RawTmpfs {
		tmpfs = append(tmpfs, expandEnv(rawTmpfs))
	}
	return tmpfs
}

func (c *container) Ulimit() []string {
	var ulimit []string
	rawUlimit := c.RawUlimit
	for _, raw := range rawUlimit {
		ulimit = append(ulimit, expandEnv(raw))
	}
	return ulimit
}

func (c *container) User() string {
	return expandEnv(c.RawUser)
}

func (c *container) Userns() string {
	if len(c.RawUserns) > 0 {
		return expandEnv(c.RawUserns)
	}
	return expandEnv(c.RawUserns_Mode)
}

func (c *container) Uts() string {
	return expandEnv(c.RawUts)
}

func (c *container) Volume() []string {
	var volumes []string
	rawVolume := c.RawVolumes
	if len(c.RawVolume) > 0 {
		rawVolume = c.RawVolume
	}
	for _, raw := range rawVolume {
		volume := expandEnv(raw)
		volumes = append(volumes, volume)
	}
	return volumes
}

func (c *container) VolumeSources() []string {
	volumes := c.Volume()
	var volumeSources []string
	for _, volume := range volumes {
		parts := strings.Split(volume, ":")
		volumeSources = append(volumeSources, parts[0])
	}
	return volumeSources
}

func (c *container) ActualVolume() []string {
	vols := []string{}
	for _, volume := range c.Volume() {
		parts := strings.Split(volume, ":")
		if includes(cfg.VolumeNames(), parts[0]) {
			parts[0] = cfg.Volume(parts[0]).ActualName()
		} else if !path.IsAbs(parts[0]) {
			parts[0] = cfg.Path() + "/" + parts[0]
		}
		vols = append(vols, strings.Join(parts, ":"))
	}
	return vols
}

func (c *container) VolumeDriver() string {
	if len(c.RawVolumeDriver) > 0 {
		return expandEnv(c.RawVolumeDriver)
	}
	return expandEnv(c.RawVolume_Driver)
}

func (c *container) VolumesFrom() []string {
	var volumesFrom []string
	for _, rawVolumesFrom := range c.RawVolumesFrom {
		volumesFrom = append(volumesFrom, expandEnv(rawVolumesFrom))
	}
	return volumesFrom
}

func (c *container) Workdir() string {
	if len(c.RawWorkdir) > 0 {
		return expandEnv(c.RawWorkdir)
	}
	return expandEnv(c.RawWorking_Dir)
}

func (c *container) Cmd() []string {
	var cmd []string
	if c.RawCmd != nil {
		switch rawCmd := c.RawCmd.(type) {
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

func containerID(name string) string {
	// `docker inspect` works for both image and containers, make sure this is a
	// container payload we get back, otherwise we might end up getting the ID
	// of the image of the same name.
	return inspectString(name, "{{if .State}}{{.Id}}{{else}}{{end}}")
}

func (c *container) ID() string {
	if len(c.id) == 0 {
		c.id = containerID(c.ActualName(false))
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
		c.Rm(true, false)
	}
	fmt.Fprintf(c.CommandsOut(), "Creating container %s ...\n", c.ActualName(adHoc))

	args := append([]string{"create"}, c.createArgs(cmds)...)
	executeCommand("docker", args, c.CommandsOut(), c.CommandsErr())
}

// Run container, or start it if already existing
func (c *container) Run(cmds []string, detach bool) {
	adHoc := (len(cmds) > 0)
	if !adHoc {
		c.Rm(true, false)
	}
	executeHook(c.Hooks().PreStart(), c.ActualName(adHoc))
	fmt.Fprintf(c.CommandsOut(), "Running container %s ...\n", c.ActualName(adHoc))

	args := []string{"run"}
	// Detach
	if !adHoc && (detach || c.Detach) {
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
		cmd, cmdOut, _ := executeCommandBackground("docker", []string{"events", "--filter", "event=start", "--filter", "container=" + c.ActualName(adHoc)})
		if cmd != nil {
			wg.Add(1)
			go func() {
				defer func() {
					handleRecoveredError(recover())
					wg.Done()
				}()
				r := bufio.NewReader(cmdOut)
				_, _, err := r.ReadLine()
				cmd.Process.Kill()
				if err != nil {
					printNoticef("Could not execute post-start hook for %s: %s.", c.ActualName(adHoc), err)
				} else {
					executeHook(c.Hooks().PostStart(), c.ActualName(adHoc))
				}
			}()
		}
	}

	return &wg
}

// Returns all the flags to be passed to `docker create`
func (c *container) createArgs(cmds []string) []string {
	adHoc := (len(cmds) > 0)
	args := []string{}
	// AddHost
	for _, addHost := range c.AddHost() {
		args = append(args, "--add-host", addHost)
	}
	// BlkioWeight
	if c.BlkioWeight > 0 {
		args = append(args, "--blkio-weight", strconv.Itoa(c.BlkioWeight))
	}
	// BlkioWeightDevice
	for _, blkioWeightDevice := range c.BlkioWeightDevice() {
		args = append(args, "--blkio-weight-device", blkioWeightDevice)
	}
	// CapAdd
	for _, capAdd := range c.CapAdd() {
		args = append(args, "--cap-add", capAdd)
	}
	// CapDrop
	for _, capDrop := range c.CapDrop() {
		args = append(args, "--cap-drop", capDrop)
	}
	// CgroupParent
	if len(c.CgroupParent()) > 0 {
		args = append(args, "--cgroup-parent", c.CgroupParent())
	}
	// Cidfile
	if len(c.Cidfile()) > 0 {
		args = append(args, "--cidfile", c.Cidfile())
	}
	// CPUPeriod
	if c.CPUPeriod > 0 {
		args = append(args, "--cpu-period", strconv.Itoa(c.CPUPeriod))
	}
	// CPUQuota
	if c.CPUQuota > 0 {
		args = append(args, "--cpu-quota", strconv.Itoa(c.CPUQuota))
	}
	// CPU set
	if c.CPUset > 0 {
		args = append(args, "--cpuset", strconv.Itoa(c.CPUset))
	}
	// CPU shares
	if c.CPUShares > 0 {
		args = append(args, "--cpu-shares", strconv.Itoa(c.CPUShares))
	}
	// DetachKeys
	if len(c.DetachKeys()) > 0 {
		args = append(args, "--detach-keys", c.DetachKeys())
	}
	// Device
	for _, device := range c.Device() {
		args = append(args, "--device", device)
	}
	// DeviceReadBps
	for _, deviceReadBps := range c.DeviceReadBps() {
		args = append(args, "--device-read-bps", deviceReadBps)
	}
	// DeviceReadIops
	for _, deviceReadIops := range c.DeviceReadIops() {
		args = append(args, "--device-read-iops", deviceReadIops)
	}
	// DeviceWriteBps
	for _, deviceWriteBps := range c.DeviceWriteBps() {
		args = append(args, "--device-write-bps", deviceWriteBps)
	}
	// DeviceWriteIops
	for _, deviceWriteIops := range c.DeviceWriteIops() {
		args = append(args, "--device-write-iops", deviceWriteIops)
	}
	// DNS
	for _, dns := range c.DNS() {
		args = append(args, "--dns", dns)
	}

	// DNSOpt
	for _, dnsOpt := range c.DNSOpt() {
		args = append(args, "--dns-opt", dnsOpt)
	}
	// DNS Search
	for _, dnsSearch := range c.DNSSearch() {
		args = append(args, "--dns-search", dnsSearch)
	}
	// Entrypoint
	if len(c.Entrypoint()) > 0 {
		args = append(args, "--entrypoint", c.Entrypoint())
	}
	// Env
	for _, env := range c.Env() {
		args = append(args, "--env", env)
	}
	// Env file
	for _, envFile := range c.EnvFile() {
		args = append(args, "--env-file", envFile)
	}
	// Expose
	for _, expose := range c.Expose() {
		args = append(args, "--expose", expose)
	}
	// GroupAdd
	for _, groupAdd := range c.GroupAdd() {
		args = append(args, "--group-add", groupAdd)
	}
	// Health Cmd
	if len(c.HealthCmd()) > 0 {
		args = append(args, "--health-cmd", c.HealthCmd())
	}
	// Health Interval
	if len(c.HealthInterval()) > 0 {
		args = append(args, "--health-interval", c.HealthInterval())
	}
	// Health Retries
	if c.HealthRetries > 0 {
		args = append(args, "--health-retries", strconv.Itoa(c.HealthRetries))
	} else if c.HealthcheckParams().Retries > 0 {
		args = append(args, "--health-retries", strconv.Itoa(c.HealthcheckParams().Retries))
	}
	// Health Timeout
	if len(c.HealthTimeout()) > 0 {
		args = append(args, "--health-timeout", c.HealthTimeout())
	}
	// Host
	if len(c.Hostname()) > 0 {
		args = append(args, "--hostname", c.Hostname())
	}
	// Interactive
	if c.Stdin_Open || c.Interactive {
		args = append(args, "--interactive")
	}
	// Ip
	if !adHoc {
		if len(c.Ip()) > 0 {
			args = append(args, "--ip", c.Ip())
		}
	}
	// Ip6
	if !adHoc {
		if len(c.Ip6()) > 0 {
			args = append(args, "--ip6", c.Ip6())
		}
	}
	// IPC
	if len(c.IPC()) > 0 {
		ipcContainer := containerReference(c.IPC())
		if len(ipcContainer) > 0 {
			if includes(allowed, ipcContainer) {
				args = append(args, "--ipc", "container:"+cfg.Container(ipcContainer).ActualName(false))
			}
		} else {
			args = append(args, "--ipc", c.IPC())
		}
	}
	// Isolation
	if len(c.Isolation()) > 0 {
		args = append(args, "--isolation", c.Isolation())
	}
	// KernelMemory
	if len(c.KernelMemory()) > 0 {
		args = append(args, "--kernel-memory", c.KernelMemory())
	}
	// Label
	for _, label := range c.Label() {
		args = append(args, "--label", label)
	}
	// LabelFile
	for _, labelFile := range c.LabelFile() {
		args = append(args, "--label-file", labelFile)
	}
	// Link
	for _, link := range c.Link() {
		linkParts := strings.Split(link, ":")
		linkName := linkParts[0]
		if includes(allowed, linkName) {
			linkParts[0] = cfg.Container(linkName).ActualName(false)
			args = append(args, "--link", strings.Join(linkParts, ":"))
		}
	}
	// LogDriver
	if len(c.LogDriver()) > 0 {
		args = append(args, "--log-driver", c.LogDriver())
	}
	// LogOpt
	for _, opt := range c.LogOpt() {
		args = append(args, "--log-opt", opt)
	}
	// LxcConf
	for _, lxcConf := range c.LxcConf() {
		args = append(args, "--lxc-conf", lxcConf)
	}
	// Mac address
	if len(c.MacAddress()) > 0 {
		args = append(args, "--mac-address", c.MacAddress())
	}
	// Memory
	if len(c.Memory()) > 0 {
		args = append(args, "--memory", c.Memory())
	}
	// MemoryReservation
	if len(c.MemoryReservation()) > 0 {
		args = append(args, "--memory-reservation", c.MemoryReservation())
	}
	// MemorySwap
	if len(c.MemorySwap()) > 0 {
		args = append(args, "--memory-swap", c.MemorySwap())
	}
	// MemorySwappiness
	if c.MemorySwappiness.Defined {
		args = append(args, "--memory-swappiness", strconv.Itoa(c.MemorySwappiness.Value))
	}
	// Net
	netParam := c.ActualNet()
	if len(c.RawNetworks) > 1 {
		printErrorf("Crane does not support joining more than one network yet. Connecting service %s to network %s.\n", c.ActualName(adHoc), netParam)
	}
	if netParam != "bridge" {
		args = append(args, "--net", netParam)
	}
	// NetAlias
	for _, netAlias := range c.NetAlias() {
		args = append(args, "--net-alias", netAlias)
	}
	// NoHealthcheck
	if c.NoHealthcheck || c.HealthcheckParams().Disable {
		args = append(args, "--no-healthcheck")
	}
	// OomKillDisable
	if c.OomKillDisable {
		args = append(args, "--oom-kill-disable")
	}
	// OomScoreAdj
	if len(c.OomScoreAdj()) > 0 {
		args = append(args, "--oom-score-adj", c.OomScoreAdj())
	}
	// PID
	if len(c.Pid()) > 0 {
		args = append(args, "--pid", c.Pid())
	}
	// Privileged
	if c.Privileged {
		args = append(args, "--privileged")
	}
	// Publish
	if !adHoc {
		for _, port := range c.Publish() {
			args = append(args, "--publish", port)
		}
	}
	// PublishAll
	if !adHoc {
		if c.PublishAll {
			args = append(args, "--publish-all")
		}
	}
	// ReadOnly
	if c.ReadOnly || c.Read_Only {
		args = append(args, "--read-only")
	}
	// Restart
	if len(c.Restart()) > 0 {
		args = append(args, "--restart", c.Restart())
	}
	// Rm
	if adHoc || c.RawRm {
		args = append(args, "--rm")
	}
	// SecurityOpt
	for _, securityOpt := range c.SecurityOpt() {
		args = append(args, "--security-opt", securityOpt)
	}
	// ShmSize
	if len(c.ShmSize()) > 0 {
		args = append(args, "--shm-size", c.ShmSize())
	}
	// SigProxy
	if c.SigProxy.Falsy() {
		args = append(args, "--sig-proxy=false")
	}
	// StopSignal
	if len(c.StopSignal()) > 0 {
		args = append(args, "--stop-signal", c.StopSignal())
	}
	// StopTimeout
	if len(c.StopTimeout()) > 0 {
		args = append(args, "--stop-timeout", c.StopTimeout())
	}
	// Tmpfs
	for _, tmpfs := range c.Tmpfs() {
		args = append(args, "--tmpfs", tmpfs)
	}
	// Tty
	if c.Tty {
		args = append(args, "--tty")
	}
	// Ulimit
	for _, ulimit := range c.Ulimit() {
		args = append(args, "--ulimit", ulimit)
	}
	// User
	if len(c.User()) > 0 {
		args = append(args, "--user", c.User())
	}
	// Userns
	if len(c.Userns()) > 0 {
		args = append(args, "--userns", c.Userns())
	}
	// Uts
	if len(c.Uts()) > 0 {
		args = append(args, "--uts", c.Uts())
	}
	// Volumes
	for _, volume := range c.ActualVolume() {
		volumeArgs := []string{"--volume", volume}
		if runtime.GOOS == "darwin" && isSyncPossible() {
			if s := cfg.MacSync(volume); s != nil {
				syncRunning := s.Running()
				if !syncRunning && s.Autostart() {
					s.Start(false)
					syncRunning = true
				}
				if syncRunning {
					volumeArgs = []string{
						"--volumes-from", s.ContainerName(),
						"--label", "tech.craneup.mac-sync=" + s.ContainerName(),
					}
				}
			}
		}
		args = append(args, volumeArgs...)
	}
	// VolumeDriver
	if len(c.VolumeDriver()) > 0 {
		args = append(args, "--volume-driver", c.VolumeDriver())
	}
	// VolumesFrom
	for _, volumesFrom := range c.VolumesFrom() {
		volumesFromParts := strings.Split(volumesFrom, ":")
		volumesFromName := volumesFromParts[0]
		if includes(allowed, volumesFromName) {
			volumesFromParts[0] = cfg.Container(volumesFromName).ActualName(false)
			args = append(args, "--volumes-from", strings.Join(volumesFromParts, ":"))
		}
	}
	// Workdir
	if len(c.Workdir()) > 0 {
		args = append(args, "--workdir", c.Workdir())
	}
	// Name
	args = append(args, "--name", c.ActualName(adHoc))
	// Image
	args = append(args, c.Image())
	// Command
	if len(cmds) > 0 {
		args = append(args, cmds...)
	} else {
		args = append(args, c.Cmd()...)
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
			args = append(args, c.ActualName(false))
			wg := c.executePostStartHook(false)
			executeCommand("docker", args, c.CommandsOut(), c.CommandsErr())
			wg.Wait()
		}
	} else {
		c.Run([]string{}, true)
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
func (c *container) Exec(cmds []string, privileged bool, user string) {
	name := c.ActualName(false)
	if !c.Running() {
		c.Start()
	}
	args := []string{"exec"}
	if privileged {
		args = append(args, "--privileged")
	}
	args = append(args, "--interactive")
	args = append(args, "--tty")
	if len(user) > 0 {
		args = append(args, "--user", user)
	}
	args = append(args, name)
	args = append(args, cmds...)
	executeCommand("docker", args, c.CommandsOut(), c.CommandsErr())
}

// Remove container
func (c *container) Rm(force bool, volumes bool) {
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
		if volumes {
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
		cmd, stdout, stderr := executeCommandBackground("docker", args)
		if cmd != nil {
			sources = append(sources, LogSource{
				Stdout: stdout,
				Stderr: stderr,
				Name:   name,
			})
		}
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
	if parts := strings.Split(reference, ":"); len(parts) == 2 && (parts[0] == "container" || parts[0] == "service") {
		// We'll just assume here that the reference is a name, and not an id, even
		// though docker supports it, since we have no bullet-proof way to tell:
		// heuristics to detect whether it's an id could bring false positives, and
		// a lookup in the list of container names could bring false negatives
		name = parts[1]
	}
	return
}

// Transform an unmarshalled payload (YAML or JSON) of type slice or map to a slice of env-expanded "K=V" strings
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
		case []interface{}: // YAML or JSON: array
			for _, v := range concreteValue {
				result = append(result, expandedStringOrPanic(v))
			}
		case map[interface{}]interface{}: // YAML: hash
			for k, v := range concreteValue {
				result = append(result, expandedStringOrPanic(k)+"="+expandedStringOrPanic(v))
			}
		case map[string]interface{}: // JSON: hash
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
// for the `docker inspect` as a boolean, falling back to
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
