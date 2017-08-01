package crane

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/imdario/mergo"
	"gopkg.in/v2/yaml"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

type Config interface {
	DependencyMap() map[string]*Dependencies
	ContainersForReference(reference string) (result []string)
	Path() string
	UniqueID() string
	Prefix() string
	Tag() string
	NetworkNames() []string
	VolumeNames() []string
	Network(name string) Network
	Volume(name string) Volume
	AcceleratedMount(volume string) AcceleratedMount
	ContainerMap() ContainerMap
	Container(name string) Container
	ContainerInfo(name string) ContainerInfo
}

type config struct {
	RawPrefix            interface{}                  `json:"prefix" yaml:"prefix"`
	RawContainers        map[string]*container        `json:"services" yaml:"services"`
	RawGroups            map[string][]string          `json:"groups" yaml:"groups"`
	RawHooks             map[string]hooks             `json:"hooks" yaml:"hooks"`
	RawNetworks          map[string]*network          `json:"networks" yaml:"networks"`
	RawVolumes           map[string]*volume           `json:"volumes" yaml:"volumes"`
	RawAcceleratedMounts map[string]*acceleratedMount `json:"accelerated-mounts" yaml:"accelerated-mounts"`
	RawMacSyncs          map[string]*acceleratedMount `json:"mac-syncs" yaml:"mac-syncs"`
	containerMap         ContainerMap
	networkMap           NetworkMap
	volumeMap            VolumeMap
	acceleratedMountMap  AcceleratedMountMap
	groups               map[string][]string
	path                 string
	prefix               string
	tag                  string
	uniqueID             string
}

// ContainerMap maps the container name
// to its configuration
type ContainerMap map[string]Container

type NetworkMap map[string]Network

type VolumeMap map[string]Volume

type AcceleratedMountMap map[string]AcceleratedMount

// readFile will read the config file
// and return the created config.
func readFile(filename string) *config {
	verboseLog("Reading configuration " + filename)
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(StatusError{err, 74})
	}

	ext := filepath.Ext(filename)
	return unmarshal(data, ext)
}

// displaySyntaxError will display more information
// such as line and error type given an error and
// the data that was unmarshalled.
// Thanks to https://github.com/markpeek/packer/commit/5bf33a0e91b2318a40c42e9bf855dcc8dd4cdec5
func displaySyntaxError(data []byte, syntaxError error) (err error) {
	syntax, ok := syntaxError.(*json.SyntaxError)
	if !ok {
		err = syntaxError
		return
	}
	newline := []byte{'\x0a'}
	space := []byte{' '}

	start, end := bytes.LastIndex(data[:syntax.Offset], newline)+1, len(data)
	if idx := bytes.Index(data[start:], newline); idx >= 0 {
		end = start + idx
	}

	line, pos := bytes.Count(data[:start], newline)+1, int(syntax.Offset)-start-1

	err = fmt.Errorf("\nError in line %d: %s \n%s\n%s^", line, syntaxError, data[start:end], bytes.Repeat(space, pos))
	return
}

// unmarshal converts either JSON
// or YAML into a config object.
func unmarshal(data []byte, ext string) *config {
	var config *config
	var err error
	if ext == ".json" {
		err = json.Unmarshal(data, &config)
	} else if ext == ".yml" || ext == ".yaml" {
		err = yaml.Unmarshal(data, &config)
	} else {
		panic(StatusError{errors.New("Unrecognized file extension"), 65})
	}
	if err != nil {
		err = displaySyntaxError(data, err)
		panic(StatusError{err, 65})
	}
	return config
}

// NewConfig retus a new config based on given
// location.
// Containers will be ordered so that they can be
// brought up and down with Docker.
func NewConfig(files []string, prefix string, tag string) Config {
	var config *config
	// Files can be given colon-separated
	expandedFiles := []string{}
	for _, f := range files {
		expandedFiles = append(expandedFiles, strings.Split(f, ":")...)
	}
	configPath := findConfigPath(expandedFiles)
	config = readConfig(configPath, expandedFiles)
	config.path = configPath
	config.initialize(prefix)
	config.validate()
	config.tag = tag
	milliseconds := time.Now().UnixNano() / 1000000
	config.uniqueID = strconv.FormatInt(milliseconds, 10)
	return config
}

func readConfig(configPath string, files []string) *config {
	var config *config

	for _, f := range files {
		filename := filepath.Base(f)
		absFile := filepath.Join(configPath, filename)
		if _, err := os.Stat(absFile); err == nil {
			fileConfig := readFile(absFile)
			if config == nil {
				config = fileConfig
			} else {
				mergo.Merge(config, fileConfig)
			}
		} else if !includes(defaultFiles, filename) {
			panic(StatusError{fmt.Errorf("Configuration file %v was not found!", filename), 78})
		}
	}

	return config
}

func findConfigPath(files []string) string {
	// If the first of the locations array is specified as an absolute
	// path, we use its directory as the config path.
	if filepath.IsAbs(files[0]) {
		return filepath.Dir(files[0])
	}

	// Otherwise, we traverse directories upwards, until we find a
	// directory which has one of the locations, then use that
	// directory as the config path.
	configPath, _ := os.Getwd()
	for {
		for _, f := range files {
			filename := filepath.Join(configPath, f)
			if _, err := os.Stat(filename); err == nil {
				return configPath
			}
		}
		// Loop only if we haven't yet reached the root
		if parentPath := filepath.Dir(configPath); len(parentPath) != len(configPath) {
			configPath = parentPath
		} else {
			break
		}
	}

	panic(StatusError{fmt.Errorf("No config files found for: %v", files), 78})
}

// Return path of config file
func (c *config) Path() string {
	return c.path
}

func (c *config) UniqueID() string {
	return c.uniqueID
}

func (c *config) Prefix() string {
	return c.prefix
}

func (c *config) Tag() string {
	return c.tag
}

func (c *config) ContainerMap() ContainerMap {
	return c.containerMap
}

func (c *config) Container(name string) Container {
	return c.containerMap[name]
}

func (c *config) ContainerInfo(name string) ContainerInfo {
	return c.Container(name)
}

func (c *config) NetworkNames() []string {
	networks := []string{}
	for name, _ := range c.networkMap {
		networks = append(networks, name)
	}
	sort.Strings(networks)
	return networks
}

func (c *config) VolumeNames() []string {
	volumes := []string{}
	for name, _ := range c.volumeMap {
		volumes = append(volumes, name)
	}
	sort.Strings(volumes)
	return volumes
}

func (c *config) Network(name string) Network {
	return c.networkMap[name]
}

func (c *config) Volume(name string) Volume {
	return c.volumeMap[name]
}

func (c *config) AcceleratedMount(name string) AcceleratedMount {
	return c.acceleratedMountMap[name]
}

// Load configuration into the internal structs from the raw, parsed ones
func (c *config) initialize(prefixFlag string) {
	// Local container map to query by expanded name
	containerMap := make(map[string]*container)
	for rawName, container := range c.RawContainers {
		container.RawName = rawName
		containerMap[container.Name()] = container
	}
	// Local hooks map to query by expanded name
	hooksMap := make(map[string]hooks)
	for hooksRawName, hooks := range c.RawHooks {
		hooksMap[expandEnv(hooksRawName)] = hooks
	}
	// Groups
	c.groups = make(map[string][]string)
	for groupRawName, rawNames := range c.RawGroups {
		groupName := expandEnv(groupRawName)
		for _, rawName := range rawNames {
			c.groups[groupName] = append(c.groups[groupName], expandEnv(rawName))
		}
		if hooks, ok := hooksMap[groupName]; ok {
			// attach group-defined hooks to the group containers
			for _, name := range c.groups[groupName] {
				if overriden := containerMap[name].hooks.CopyFrom(hooks); overriden {
					panic(StatusError{fmt.Errorf("Multiple conflicting hooks inherited from groups for container `%s`", name), 64})
				}
			}
		}
	}
	// Container map
	c.containerMap = make(map[string]Container)
	for name, container := range containerMap {
		if hooks, ok := hooksMap[name]; ok {
			// attach container-defined hooks, overriding potential group-inherited hooks
			container.hooks.CopyFrom(hooks)
		}
		c.containerMap[name] = container
	}

	c.determinePrefix(prefixFlag)
	c.setNetworkMap()
	c.setVolumeMap()
	c.setAcceleratedMountMap()
}

func (c *config) setNetworkMap() {
	c.networkMap = make(map[string]Network)
	for rawName, net := range c.RawNetworks {
		if net == nil {
			net = &network{}
		}
		net.RawName = rawName
		c.networkMap[net.Name()] = net
	}
}

func (c *config) setVolumeMap() {
	c.volumeMap = make(map[string]Volume)
	for rawName, vol := range c.RawVolumes {
		if vol == nil {
			vol = &volume{}
		}
		vol.RawName = rawName
		c.volumeMap[vol.Name()] = vol
	}
}

func (c *config) setAcceleratedMountMap() {
	c.acceleratedMountMap = make(map[string]AcceleratedMount)
	for rawVolume, am := range c.RawAcceleratedMounts {
		if am == nil {
			am = &acceleratedMount{}
		}
		am.RawVolume = rawVolume
		am.configPath = c.path
		c.acceleratedMountMap[am.Volume()] = am
	}
	for rawVolume, am := range c.RawMacSyncs {
		if am == nil {
			am = &acceleratedMount{}
		}
		am.RawVolume = rawVolume
		am.configPath = c.path
		c.acceleratedMountMap[am.Volume()] = am
	}
}

// CLI > Config > Default
func (c *config) determinePrefix(prefixFlag string) {
	// CLI takes precedence over config
	if len(prefixFlag) > 0 {
		c.prefix = prefixFlag
		return
	}
	// If prefix is not configured, don't use any
	if c.RawPrefix == nil {
		c.prefix = ""
		return
	}
	// Use configured prefix:
	// true -> folder name
	// false -> no prefix
	// string -> use as-is
	switch concretePrefix := c.RawPrefix.(type) {
	case bool:
		if concretePrefix {
			c.prefix = filepath.Base(c.path) + "_"
		} else {
			c.prefix = ""
		}
	case string:
		c.prefix = expandEnv(concretePrefix)
	default:
		panic(StatusError{fmt.Errorf("prefix must be either string or boolean", c.RawPrefix), 65})
	}
}

func (c *config) validate() {
	for name, container := range c.RawContainers {
		if len(container.RawImage) == 0 && container.RawBuild == (BuildParameters{}) {
			panic(StatusError{fmt.Errorf("Neither image or build specified for `%s`", name), 64})
		}
	}
}

// DependencyMap returns a map of containers to their dependencies.
func (c *config) DependencyMap() map[string]*Dependencies {
	dependencyMap := make(map[string]*Dependencies)
	for _, container := range c.containerMap {
		if includes(allowed, container.Name()) {
			dependencyMap[container.Name()] = container.Dependencies()
		}
	}
	return dependencyMap
}

// ContainersForReference receives a reference and determines which
// containers of the map that resolves to.
func (c *config) ContainersForReference(reference string) (result []string) {
	containers := []string{}
	if len(reference) == 0 {
		// reference not given
		var defaultGroup []string
		for group, containers := range c.groups {
			if group == "default" {
				defaultGroup = containers
				break
			}
		}
		if defaultGroup != nil {
			// If default group exists, return its containers
			containers = defaultGroup
		} else {
			// Otherwise, return all containers
			for name := range c.containerMap {
				containers = append(containers, name)
			}
		}
	} else {
		// reference given
		reference = expandEnv(reference)
		// Select reference from listed groups
		for group, groupContainers := range c.groups {
			if group == reference {
				containers = append(containers, groupContainers...)
				break
			}
		}
		if len(containers) == 0 {
			// The reference might just be one container
			for name := range c.containerMap {
				if name == reference {
					containers = append(containers, reference)
					break
				}
			}
		}
		if len(containers) == 0 {
			// reference was not found anywhere
			panic(StatusError{fmt.Errorf("No group or container matching `%s`", reference), 64})
		}
	}
	// ensure all container references exist
	for _, container := range containers {
		containerDeclared := false
		for name := range c.containerMap {
			if container == name {
				containerDeclared = true
				break
			}
		}
		if !containerDeclared {
			panic(StatusError{fmt.Errorf("Invalid container reference `%s`", container), 64})
		}
		if !includes(result, container) {
			result = append(result, container)
		}
	}
	return
}
