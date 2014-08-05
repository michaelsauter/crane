package crane

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"gopkg.in/v1/yaml"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
)

type Config struct {
	RawContainerMap ContainerMap        `json:"containers" yaml:"containers"`
	RawOrder        []string            `json:"order" yaml:"order"`
	RawGroups       map[string][]string `json:"groups" yaml:"groups"`
	containerMap    ContainerMap
	dependencyGraph DependencyGraph
	target          Target
	order           []string
	groups          map[string][]string
}

// ContainerMap maps the container name
// to its configuration
type ContainerMap map[string]Container

type Target []string

// configFiles returns a slice of
// files to read the config from.
// If the --config option was given,
// it will just use the given file.
func configFiles(options Options) []string {
	if len(options.config) > 0 {
		return []string{options.config}
	} else {
		return []string{"crane.json", "crane.yaml", "crane.yml"}
	}
}

// readConfig will read the config file
// and return the created config.
func readConfig(filename string) *Config {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(StatusError{err, 74})
	}

	ext := filepath.Ext(filename)
	if ext == ".json" {
		return unmarshalJSON(data)
	} else if ext == ".yml" || ext == ".yaml" {
		return unmarshalYAML(data)
	} else {
		panic(StatusError{errors.New("Unrecognized file extension"), 65})
	}
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

// unmarshalJSON converts given JSON data
// into a config object.
func unmarshalJSON(data []byte) *Config {
	var config *Config
	err := json.Unmarshal(data, &config)
	if err != nil {
		err = displaySyntaxError(data, err)
		panic(StatusError{err, 65})
	}
	return config
}

// unmarshalYAML converts given YAML data
// into a config object.
func unmarshalYAML(data []byte) *Config {
	var config *Config
	err := yaml.Unmarshal(data, &config)
	if err != nil {
		err = displaySyntaxError(data, err)
		panic(StatusError{err, 65})
	}
	return config
}

func NewConfig(options Options, forceOrder bool) *Config {
	var config *Config
	for _, f := range configFiles(options) {
		if _, err := os.Stat(f); err == nil {
			config = readConfig(f)
			break
		}
	}
	if config == nil {
		panic(StatusError{fmt.Errorf("No configuration found %v", configFiles(options)), 78})
	}
	config.expandEnv()
	config.determineGraph()
	config.determineTarget(options.target, options.cascadeDependencies, options.cascadeAffected)
	err := config.determineOrder(forceOrder)
	if err != nil {
		panic(StatusError{err, 78})
	}
	return config
}

// Containers returns the containers of the config in order
func (c *Config) Containers() Containers {
	var containers Containers
	for _, name := range c.order {
		containers = append([]Container{c.containerMap[name]}, containers...)
	}
	return containers
}

// expandEnv creates a new container map
// with expanded names and sets the RawName of each
// container to the map key.
// It also expand variables in the order and the groups.
func (c *Config) expandEnv() {
	// Container map
	c.containerMap = make(map[string]Container)
	for rawName, container := range c.RawContainerMap {
		container.RawName = rawName
		name := os.ExpandEnv(rawName)
		c.containerMap[name] = container
	}
	// Order
	for _, rawName := range c.RawOrder {
		c.order = append(c.order, os.ExpandEnv(rawName))
	}
	// Groups
	c.groups = make(map[string][]string)
	for groupRawName, rawNames := range c.RawGroups {
		for _, rawName := range rawNames {
			c.groups[groupRawName] = append(c.groups[groupRawName], os.ExpandEnv(rawName))
		}
	}
}

// generateGraph generated the dependency graph, which is
// a map describing the dependencies between the containers.
func (c *Config) determineGraph() {
	c.dependencyGraph = make(DependencyGraph)
	for _, container := range c.containerMap {
		c.dependencyGraph[container.Name()] = container.Dependencies()
	}
}

// determineOrder sets the Order field of the config.
// Containers will be ordered so that they can be
// brought up and down with Docker.
func (c *Config) determineOrder(force bool) error {
	if len(c.order) > 0 {
		return nil // Order was set manually
	}

	order, err := c.dependencyGraph.order(c.target, force)
	if err != nil {
		return err
	} else {
		c.order = order
	}
	return nil
}

// determineTarget receives the specified target
// and determines which containers should be targeted.
// Additionally, ot sorts these alphabetically.
func (c *Config) determineTarget(target string, cascadeDependencies string, cascadeAffected string) {
	// start from the explicitly targeted target
	includedSet := make(map[string]bool)
	cascadingSeeds := []string{}
	for _, name := range c.explicitlyTargeted(target) {
		includedSet[name] = true
		cascadingSeeds = append(cascadingSeeds, name)
	}

	// cascade until the graph has been fully traversed according to the cascading flags
	for len(cascadingSeeds) > 0 {
		nextCascadingSeeds := []string{}
		for _, seed := range cascadingSeeds {
			if cascadeDependencies != "none" {
				if dependencies, ok := c.dependencyGraph[seed]; ok {
					// queue direct dependencies if we haven't already considered them
					for _, name := range dependencies.forKind(cascadeDependencies) {
						if _, alreadyIncluded := includedSet[name]; !alreadyIncluded {
							includedSet[name] = true
							nextCascadingSeeds = append(nextCascadingSeeds, name)
						}
					}
				}
			}
			if cascadeAffected != "none" {
				// queue all containers we haven't considered yet which directly depend on the seed
				for name, dependencies := range c.dependencyGraph {
					if _, alreadyIncluded := includedSet[name]; !alreadyIncluded {
						if dependencies.includesAsKind(seed, cascadeAffected) {
							includedSet[name] = true
							nextCascadingSeeds = append(nextCascadingSeeds, name)
						}
					}
				}
			}
		}
		cascadingSeeds = nextCascadingSeeds
	}

	// keep the ones that we know of
	included := []string{}
	for name := range includedSet {
		if _, exists := c.containerMap[name]; exists {
			included = append(included, name)
		}
	}

	// sort
	c.target = Target(included)
	sort.Strings(c.target)
}

// explicitlyTargeted receives a target and determines which
// containers of the map are targeted
func (c *Config) explicitlyTargeted(target string) []string {
	// target not given
	if len(target) == 0 {
		// If default group exists, return its containers
		for group, containers := range c.groups {
			if group == "default" {
				return containers
			}
		}
		// If no default group exists, return all containers
		var containers []string
		for name, _ := range c.containerMap {
			containers = append(containers, name)
		}
		return containers
	}
	// target given
	target = os.ExpandEnv(target)
	// Select target from listed groups
	for group, containers := range c.groups {
		if group == target {
			return containers
		}
	}
	// The target might just be one container
	for name, _ := range c.containerMap {
		if name == target {
			return append([]string{}, target)
		}
	}
	// Otherwise, fail verbosely
	panic(StatusError{fmt.Errorf("No group or container matching `%s`", target), 64})
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
