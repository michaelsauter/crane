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
	"sort"
)

type Config interface {
	DependencyGraph() DependencyGraph
	DetermineTarget(target string, cascadeDependencies string, cascadeAffected string) Target
	ContainersForReference(reference string) (result []string)
	Path() string
	ContainerMap() ContainerMap
}

type config struct {
	RawContainerMap map[string]*container `json:"containers" yaml:"containers"`
	RawGroups       map[string][]string   `json:"groups" yaml:"groups"`
	RawHooksMap     map[string]hooks      `json:"hooks" yaml:"hooks"`
	containerMap    ContainerMap
	target          Target
	order           []string
	groups          map[string][]string
	path            string
}

// ContainerMap maps the container name
// to its configuration
type ContainerMap map[string]Container

type Target []string

// configFilenames returns a slice of
// files to read the config from.
// If the --config option was given,
// it will only use the given file.
func configFilenames(location string) []string {
	if len(location) > 0 {
		return []string{location}
	} else {
		return []string{"crane.json", "crane.yaml", "crane.yml"}
	}
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
		for len(configPath) > 1 {
			for _, f := range configFiles {
				filename := configPath + "/" + f
				if _, err := os.Stat(filename); err == nil {
					return filename
				}
			}
			configPath = path.Dir(configPath)
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
func NewConfig(location string) Config {
	var config *config
	configFile := findConfig(location)
	config = readConfig(configFile)
	config.initialize()
	config.path = path.Dir(configFile)
	return config
}

// Return path of config file
func (c *config) Path() string {
	return c.path
}

func (c *config) ContainerMap() ContainerMap {
	return c.containerMap
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

// DependencyGraph returns the dependency graph, which is
// a map describing the dependencies between the containers.
func (c *config) DependencyGraph() DependencyGraph {
	dependencyGraph := make(DependencyGraph)
	for _, container := range c.containerMap {
		dependencyGraph[container.Name()] = container.Dependencies()
	}
	return dependencyGraph
}

// determineTarget receives the specified target
// and determines which containers should be targeted.
// The target might be extended depending on the value
// given for cascadeDependencies and cascadeAffected.
// Additionally, the target is sorted alphabetically.
func (c *config) DetermineTarget(target string, cascadeDependencies string, cascadeAffected string) Target {
	// start from the explicitly targeted target
	includedSet := make(map[string]bool)
	cascadingSeeds := []string{}
	for _, name := range c.ContainersForReference(target) {
		includedSet[name] = true
		cascadingSeeds = append(cascadingSeeds, name)
	}

	// Cascade until the graph has been fully traversed
	// according to the cascading flags.
	for len(cascadingSeeds) > 0 {
		nextCascadingSeeds := []string{}
		for _, seed := range cascadingSeeds {
			if cascadeDependencies != "none" {
				if dependencies, ok := dependencyGraph[seed]; ok {
					// Queue direct dependencies if we haven't already considered them
					for _, name := range dependencies.forKind(cascadeDependencies) {
						if _, alreadyIncluded := includedSet[name]; !alreadyIncluded {
							includedSet[name] = true
							nextCascadingSeeds = append(nextCascadingSeeds, name)
						}
					}
				}
			}
			if cascadeAffected != "none" {
				// Queue all containers we haven't considered yet which exist
				// and directly depend on the seed.
				for name, container := range c.containerMap {
					if _, alreadyIncluded := includedSet[name]; !alreadyIncluded {
						if container.Dependencies().includesAsKind(seed, cascadeAffected) && container.Exists() {
							includedSet[name] = true
							nextCascadingSeeds = append(nextCascadingSeeds, name)
						}
					}
				}
			}
		}
		cascadingSeeds = nextCascadingSeeds
	}

	// Keep the ones that are part of the container map
	included := []string{}
	for name := range includedSet {
		if _, exists := c.containerMap[name]; exists {
			included = append(included, name)
		}
	}

	// Sort alphabetically
	sortedTarget := Target(included)
	sort.Strings(sortedTarget)
	return sortedTarget
}

// ContainersForReference receives a reference and determines which
// containers of the map that resolves to.
func (c *config) ContainersForReference(reference string) (result []string) {
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
			result = defaultGroup
		} else {
			// Otherwise, return all containers
			for name, _ := range c.containerMap {
				result = append(result, name)
			}
		}
	} else {
		// reference given
		reference = os.ExpandEnv(reference)
		// Select reference from listed groups
		for group, containers := range c.groups {
			if group == reference {
				result = append(result, containers...)
				break
			}
		}
		if len(result) == 0 {
			// The reference might just be one container
			for name, _ := range c.containerMap {
				if name == reference {
					result = append(result, reference)
					break
				}
			}
		}
		if len(result) == 0 {
			// reference was not found anywhere
			panic(StatusError{fmt.Errorf("No group or container matching `%s`", reference), 64})
		}
	}
	// ensure all container references exist
	for _, container := range result {
		containerDeclared := false
		for name, _ := range c.containerMap {
			if container == name {
				containerDeclared = true
				break
			}
		}
		if !containerDeclared {
			panic(StatusError{fmt.Errorf("Invalid container reference `%s`", container), 64})
		}
	}
	return
}

// includes checks whether the given needle is
// included in the target
func (t Target) includes(needle string) bool {
	for _, name := range t {
		if name == needle {
			return true
		}
	}
	return false
}
