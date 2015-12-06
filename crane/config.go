package crane

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"gopkg.in/v2/yaml"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type Config interface {
	DependencyMap(excluded []string) map[string]*Dependencies
	ContainersForReference(reference string) (result []string)
	Path() string
	UniqueID() string
	Prefix() string
	Tag() string
	ContainerMap() ContainerMap
	Container(name string) Container
	ContainerInfo(name string) ContainerInfo
}

type config struct {
	RawContainerMap map[string]*container `json:"containers" yaml:"containers"`
	RawGroups       map[string][]string   `json:"groups" yaml:"groups"`
	RawHooksMap     map[string]hooks      `json:"hooks" yaml:"hooks"`
	containerMap    ContainerMap
	groups          map[string][]string
	path            string
	prefix          string
	tag             string
	uniqueID        string
}

// ContainerMap maps the container name
// to its configuration
type ContainerMap map[string]Container

// configFilenames returns a slice of
// files to read the config from.
// If the --config option was given,
// it will only use the given file.
func configFilenames(location string) []string {
	if len(location) > 0 {
		return []string{location}
	}
	return []string{"crane.json", "crane.yaml", "crane.yml"}
}

// findConfig returns the filename of the
// config. It searches parent directories
// if it can't find any of the config
// filenames in the current directory.
func findConfig(location string) string {
	configFiles := configFilenames(location)
	// Absolute path to config given
	if len(location) > 0 && path.IsAbs(location) {
		if _, err := os.Stat(location); err == nil {
			return location
		}
	} else { // Relative config
		configPath, _ := os.Getwd()
		for {
			for _, f := range configFiles {
				// the root path is a `/` but others don't have a trailing `/`
				filename := strings.TrimSuffix(configPath, "/") + "/" + f
				if _, err := os.Stat(filename); err == nil {
					return filename
				}
			}
			// loop only if we haven't yet reached the root
			if parentPath := path.Dir(configPath); len(parentPath) != len(configPath) {
				configPath = parentPath
			} else {
				break
			}
		}
	}
	panic(StatusError{fmt.Errorf("No configuration found %v", configFiles), 78})
}

// readConfig will read the config file
// and return the created config.
func readConfig(filename string) *config {
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
func NewConfig(location string, prefix string, tag string) Config {
	var config *config
	configFile := findConfig(location)
	if isVerbose() {
		printInfof("Using configuration file `%s`\n", configFile)
	}
	config = readConfig(configFile)
	config.initialize()
	config.validate()
	config.path = path.Dir(configFile)
	config.prefix = prefix
	config.tag = tag
	milliseconds := time.Now().UnixNano() / 1000000
	config.uniqueID = strconv.FormatInt(milliseconds, 10)
	return config
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

// Load configuration into the internal structs from the raw, parsed ones
func (c *config) initialize() {
	// Local container map to query by expanded name
	containerMap := make(map[string]*container)
	for rawName, container := range c.RawContainerMap {
		container.RawName = rawName
		containerMap[container.Name()] = container
	}
	// Local hooks map to query by expanded name
	hooksMap := make(map[string]hooks)
	for hooksRawName, hooks := range c.RawHooksMap {
		hooksMap[os.ExpandEnv(hooksRawName)] = hooks
	}
	// Groups
	c.groups = make(map[string][]string)
	for groupRawName, rawNames := range c.RawGroups {
		groupName := os.ExpandEnv(groupRawName)
		for _, rawName := range rawNames {
			c.groups[groupName] = append(c.groups[groupName], os.ExpandEnv(rawName))
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
}

func (c *config) validate() {
	for name, container := range c.RawContainerMap {
		if len(container.RawImage) == 0 {
			panic(StatusError{fmt.Errorf("No image specified for `%s`", name), 64})
		}
	}
}

// DependencyMap returns a map of containers to their dependencies.
func (c *config) DependencyMap(excluded []string) map[string]*Dependencies {
	dependencyMap := make(map[string]*Dependencies)
	for _, container := range c.containerMap {
		if !includes(excluded, container.Name()) {
			dependencyMap[container.Name()] = container.Dependencies(excluded)
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
		reference = os.ExpandEnv(reference)
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
