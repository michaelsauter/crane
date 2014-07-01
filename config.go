package main

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

var defaultConfigFiles = []string{"crane.json", "crane.yaml", "crane.yml", "Cranefile"}

func configFiles() []string {
	if len(options.config) > 0 {
		return []string{options.config}
	} else {
		return defaultConfigFiles
	}
}

type Config struct {
	Containers Containers          `json:"containers" yaml:"containers"`
	Groups     map[string][]string `json:"groups" yaml:"groups"`
}

func targetedContainers(config Config, target string) []string {
	// target not given
	if len(target) == 0 {
		// If default group exists, return its containers
		for group, containers := range config.Groups {
			if group == "default" {
				return containers
			}
		}
		// If no default group exists, return all containers
		var containers []string
		for i := 0; i < len(config.Containers); i++ {
			containers = append(containers, config.Containers[i].Name())
		}
		return containers
	}
	// target given
	// Select target from listed groups
	for group, containers := range config.Groups {
		if group == target {
			return containers
		}
	}
	// The target might just be one container
	for i := 0; i < len(config.Containers); i++ {
		if config.Containers[i].Name() == target {
			return append([]string{}, target)
		}
	}
	// Otherwise, fail verbosely
	panic(StatusError{fmt.Errorf("No group or container matching `%s`", target), 64})
}

func getConfig(options Options) Config {
	for _, f := range configFiles() {
		if _, err := os.Stat(f); err == nil {
			return readCraneData(f)
		}
	}
	panic(StatusError{fmt.Errorf("No configuration found %v", configFiles()), 78})
}

func getContainers(options Options) Containers {
	config := getConfig(options)
	return config.Containers.filter(targetedContainers(config, options.target))
}

func readCraneData(filename string) Config {
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

func unmarshalJSON(data []byte) Config {
	var config Config
	err := json.Unmarshal(data, &config)
	if err != nil {
		err = displaySyntaxError(data, err)
		panic(StatusError{err, 65})
	}
	return config
}

func unmarshalYAML(data []byte) Config {
	var config Config
	err := yaml.Unmarshal(data, &config)
	if err != nil {
		err = displaySyntaxError(data, err)
		panic(StatusError{err, 65})
	}
	return config
}
