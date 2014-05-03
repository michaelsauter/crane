package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/michaelsauter/crane/print"
	"gopkg.in/v1/yaml"
	"io/ioutil"
	"os"
	"path/filepath"
)

type Manifest struct {
	Containers Containers          `json:"containers" yaml:"containers"`
	Groups     map[string][]string `json:"groups" yaml:"groups"`
}

func determineTargetedContainers(manifest Manifest, specifiedGroup string) []string {
	// If group is not given, all containers
	if len(specifiedGroup) == 0 {
		var containers []string
		for i := 0; i < len(manifest.Containers); i++ {
			containers = append(containers, manifest.Containers[i].Name)
		}
		return containers
	}
	// Select specified group from listed groups
	for name, containers := range manifest.Groups {
		if name == specifiedGroup {
			return containers
		}
	}
	// Otherwise, the group is just the specified container
	return append([]string{}, specifiedGroup)
}

func getManifest(options Options) Manifest {
	if len(options.config) > 0 {
		return unmarshalJSON([]byte(options.config))
	} else {
		for _, f := range manifestFiles() {
			if _, err := os.Stat(f); err == nil {
				return readCraneData(f)
			}
		}
	}
	panic(StatusError{fmt.Errorf("No configuration found %v", manifestFiles()), 78})
}

func getContainers(options Options) Containers {
	manifest := getManifest(options)
	targetedContainers := determineTargetedContainers(manifest, options.group)
	return manifest.Containers.filter(targetedContainers)
}

func readCraneData(filename string) Manifest {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(err)
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
		panic("Unrecognized file extension")
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

func unmarshalJSON(data []byte) Manifest {
	var manifest Manifest
	err := json.Unmarshal(data, &manifest)
	if err != nil {
		err = displaySyntaxError(data, err)
		panic(err)
	}
	return manifest
}

func unmarshalYAML(data []byte) Manifest {
	var manifest Manifest
	err := yaml.Unmarshal(data, &manifest)
	if err != nil {
		err = displaySyntaxError(data, err)
		panic(err)
	}
	return manifest
}
