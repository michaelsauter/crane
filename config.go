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

type Config struct {
	Containers Containers          `json:"containers" yaml:"containers"`
	Groups     map[string][]string `json:"groups" yaml:"groups"`
}

func determineTargetedContainers(config Config, specifiedGroup string) []string {
	// If group is not given, all containers
	if len(specifiedGroup) == 0 {
		var containers []string
		for i := 0; i < len(config.Containers); i++ {
			containers = append(containers, config.Containers[i].Name())
		}
		return containers
	}
	// Select specified group from listed groups
	for name, containers := range config.Groups {
		if name == specifiedGroup {
			return containers
		}
	}
	// The group might just be a container reference itself
	for i := 0; i < len(config.Containers); i++ {
		if config.Containers[i].Name() == specifiedGroup {
			return append([]string{}, specifiedGroup)
		}
	}
	// Otherwise, fail verbosely
	panic(StatusError{fmt.Errorf("no group nor container matching `%s`", specifiedGroup), 64})
}

func getConfig(options Options) Config {
	if len(options.config) > 0 {
		return unmarshalJSON([]byte(options.config))
	} else {
		for _, f := range configFiles() {
			if _, err := os.Stat(f); err == nil {
				return readCraneData(f)
			}
		}
	}
	panic(StatusError{fmt.Errorf("No configuration found %v", configFiles()), 78})
}

func getContainers(options Options) Containers {
	config := getConfig(options)
	targetedContainers := determineTargetedContainers(config, options.group)
	return config.Containers.filter(targetedContainers)
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
