package main

import (
	"github.com/michaelsauter/crane/print"
	"bytes"
	"encoding/json"
	"path/filepath"
	"fmt"
	"gopkg.in/v1/yaml"
	"io/ioutil"
	"os"
)

func getContainers(config Config) Containers {

	if len(config.config) > 0 {
		return unmarshalJSON([]byte(config.config))
	}

	for _, f := range configFiles() {
		if _, err := os.Stat(f); err == nil {
			return readCraneData(f)
		}
	}
	panic(fmt.Sprintf("No configuration found %v", configFiles()))
}

func readCraneData(filename string) Containers {
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

func unmarshalJSON(data []byte) Containers {
	var containers Containers
	err := json.Unmarshal(data, &containers)
	if err != nil {
		err = displaySyntaxError(data, err)
		panic(err)
	}
	return containers
}

func unmarshalYAML(data []byte) Containers {
	var containers Containers
	err := yaml.Unmarshal(data, &containers)
	if err != nil {
		err = displaySyntaxError(data, err)
		panic(err)
	}
	return containers
}
