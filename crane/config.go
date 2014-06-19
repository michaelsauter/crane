package crane

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/michaelsauter/crane/print"
	"gopkg.in/v1/yaml"
	"io/ioutil"
	"os"
	"path/filepath"
)

// Types
type ContainerMap map[string]Container

type Config struct {
	ContainerMap *ContainerMap       `json:"containers" yaml:"containers"`
	Order        []string            `json:"order" yaml:"order"`
	Groups       map[string][]string `json:"groups" yaml:"groups"`
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
		return []string{"crane.json", "crane.yaml", "crane.yml", "Cranefile"}
	}
}

// readConfig will read the config file
// and return the created config.
func readConfig(filename string) *Config {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(StatusError{err, 74})
	}

	if filename == "Cranefile" {
		print.Notice("Using a Cranefile is deprecated. Please use crane.json/crane.yaml instead.\n")
	}

	ext := filepath.Ext(filename)
	if ext == ".json" {
		return unmarshalJSON(data)
	} else if ext == ".yml" || ext == ".yaml" {
		return unmarshalYAML(data)
	} else if ext == "" {
		return unmarshalJSON(data)
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
	config.setNames()
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
	for _, name := range c.Order {
		containerMap := *c.ContainerMap
		containers = append(containers, containerMap[name])
	}
	return containers
}

// setNames sets the RawName of each container
// to the container map key.
func (c *Config) setNames() {
	for name, container := range *c.ContainerMap {
		container.RawName = name
		containerMap := *c.ContainerMap
		containerMap[name] = container
	}
}

// determineOrder sets the Order field of the config.
// Containers will be ordered so that they can be
// brought up and down with Docker.
func (c *Config) determineOrder() error {
	if len(c.Order) > 0 {
		return nil // Order was set manually
	}
	// Setup dependencies
	dependencies := make(map[string][]string)
	for _, container := range *c.ContainerMap {
		dependencies[container.Name()] = container.Dependencies()
	}

	// Resolve dependencies
	success := true
	for success && len(dependencies) > 0 {
		success = false
		for name, unresolved := range dependencies {
			if len(unresolved) == 0 {
				// Resolve "name"
				success = true
				c.Order = append([]string{name}, c.Order...)
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
	return nil
}

// filter receives a target and deletes all containers
// from the map which are not targeted.
func (config Config) filter(target string) {
	targeted := config.targetedContainers(target)
	for name, container := range *config.ContainerMap {
		if !container.IsTargeted(targeted) {
			delete(*config.ContainerMap, name)
		}
	}
}

// targetedContainers receives a target and determines which
// containers of the map are targeted
func (config Config) targetedContainers(target string) []string {
	// target not given
	if len(target) == 0 {
		// If default group exists, return its containers
		for group, containers := range config.Groups {
			if os.ExpandEnv(group) == "default" {
				return containers
			}
		}
		// If no default group exists, return all containers
		var containers []string
		for name, _ := range *config.ContainerMap {
			containers = append(containers, os.ExpandEnv(name))
		}
		return containers
	}
	// target given
	target = os.ExpandEnv(target)
	// Select target from listed groups
	for group, containers := range config.Groups {
		if os.ExpandEnv(group) == target {
			return containers
		}
	}
	// The target might just be one container
	for name, _ := range *config.ContainerMap {
		if os.ExpandEnv(name) == target {
			return append([]string{}, target)
		}
	}
	// Otherwise, fail verbosely
	panic(StatusError{fmt.Errorf("No group or container matching `%s`", target), 64})
}
