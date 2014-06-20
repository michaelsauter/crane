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
)

// Types
type ContainerMap map[string]Container

type Config struct {
	ContainersByRawName *ContainerMap       `json:"containers" yaml:"containers"`
	RawOrder            []string            `json:"order" yaml:"order"`
	RawGroups           map[string][]string `json:"groups" yaml:"groups"`
	containersByName    *ContainerMap
	order               []string
}

// Package-level functions
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

// Constructor
func NewConfig(options Options) *Config {
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
	config.setContainersByName()
	config.filter(options.target)
	err := config.determineOrder()
	if err != nil {
		panic(StatusError{err, 78})
	}
	return config
}

// Config methods
// Containers returns the containers of the config in order
func (c *Config) Containers() Containers {
	var containers Containers
	for _, name := range c.order {
		containerMap := *c.containersByName
		containers = append(containers, containerMap[name])
	}
	return containers
}

// setContainersByName set the containersByName field with
// containers (with their raw name filled) grouped by expanded
// name
func (c *Config) setContainersByName() {
	containersByName := make(ContainerMap)
	for rawName, container := range *c.ContainersByRawName {
		container.RawName = rawName
		containersByName[os.ExpandEnv(rawName)] = container
	}
	c.containersByName = &containersByName
}

// determineOrder sets the order field of the config.
// Containers will be ordered so that they can be
// brought up and down with Docker.
func (c *Config) determineOrder() error {
	if len(c.RawOrder) > 0 {
		// Order was set manually, just expand
		c.order = make([]string, len(c.RawOrder))
		for i, rawName := range c.RawOrder {
			c.order[i] = os.ExpandEnv(rawName)
		}
		return nil
	}
	// Setup dependencies
	dependencies := make(map[string][]string)
	for name, container := range *c.containersByName {
		dependencies[name] = container.Dependencies()
	}

	// Resolve dependencies
	success := true
	var order []string
	for success && len(dependencies) > 0 {
		success = false
		for name, unresolved := range dependencies {
			if len(unresolved) == 0 {
				// Resolve "name"
				success = true
				order = append([]string{name}, order...)
				delete(dependencies, name)
				// Remove "name" from "unresolved" slices
				for name2, deps2 := range dependencies {
					for i, unresolved2 := range deps2 {
						if unresolved2 == name {
							dependencies[name2] = append(deps2[:i], deps2[i+1:]...)
						}
					}
				}
				// Continue with next iteration
				break
			}
		}
	}

	// If we still have dependencies, the container map
	// cannot be resolved (cyclic or missing dependency found).
	if len(dependencies) > 0 {
		return errors.New("Container map cannot be resolved. Check for cyclic or missing dependencies.")
	}
	// No error
	c.order = order
	return nil
}

// filter receives a target and deletes all containers
// from the map which are not targeted.
func (config Config) filter(target string) {
	targeted := config.targetedContainers(target)
	for i, t := range targeted {
		targeted[i] = os.ExpandEnv(t)
	}
	for name, container := range *config.containersByName {
		if !container.IsTargeted(targeted) {
			delete(*config.containersByName, name)
		}
	}
}

// targetedContainers receives a target and determines which
// containers of the map are targeted
func (config Config) targetedContainers(target string) []string {
	// target not given
	if len(target) == 0 {
		// If default group exists, return its containers
		for group, containers := range config.RawGroups {
			if os.ExpandEnv(group) == "default" {
				return containers
			}
		}
		// If no default group exists, return all containers
		var containers []string
		for rawName, _ := range *config.ContainersByRawName {
			containers = append(containers, rawName)
		}
		return containers
	}
	// target given
	target = os.ExpandEnv(target)
	// Select target from listed groups
	for group, containers := range config.RawGroups {
		if os.ExpandEnv(group) == target {
			return containers
		}
	}
	// The target might just be one container
	for rawName, _ := range *config.ContainersByRawName {
		if os.ExpandEnv(rawName) == target {
			return append([]string{}, rawName)
		}
	}
	// Otherwise, fail verbosely
	panic(StatusError{fmt.Errorf("No group or container matching `%s`", target), 64})
}
