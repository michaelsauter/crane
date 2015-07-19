package crane

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"gopkg.in/v2/yaml"
	"os"
	"testing"
)

func TestImageWithTag(t *testing.T) {
	var c *container
	// No tag
	c = &container{
		RawImage:  "ubuntu",
		RunParams: RunParameters{},
	}
	assert.Equal(t, "ubuntu:latest", c.ImageWithTag())
	// Given tag
	c = &container{
		RawImage:  "ubuntu:14.04",
		RunParams: RunParameters{},
	}
	assert.Equal(t, "ubuntu:14.04", c.ImageWithTag())
	// With port in registry part
	c = &container{
		RawImage:  "private.registry.com:5000/ubuntu",
		RunParams: RunParameters{},
	}
	assert.Equal(t, "ubuntu:latest", c.ImageWithTag())
}

func TestDependencies(t *testing.T) {
	c := &container{
		RunParams: RunParameters{
			RawNet:         "container:n",
			RawLink:        []string{"a:b", "b:d"},
			RawVolumesFrom: []string{"c"},
		},
	}
	expected := &Dependencies{
		All:         []string{"a", "b", "c", "n"},
		Link:        []string{"a", "b"},
		VolumesFrom: []string{"c"},
		Net:         "n",
	}
	assert.Equal(t, expected, c.Dependencies([]string{}))

	c = &container{}
	expected = &Dependencies{}
	assert.Equal(t, expected, c.Dependencies([]string{}))
}

func TestVolumesFromSuffixes(t *testing.T) {
	c := &container{RunParams: RunParameters{RawVolumesFrom: []string{"a:rw", "b:ro"}}}
	expected := &Dependencies{
		All:         []string{"a", "b"},
		VolumesFrom: []string{"a", "b"},
	}
	assert.Equal(t, expected, c.Dependencies([]string{}))
}

func TestMultipleLinkAliases(t *testing.T) {
	c := &container{RunParams: RunParameters{RawLink: []string{"a:b", "a:c"}}}
	expected := &Dependencies{
		All:  []string{"a"},
		Link: []string{"a"},
	}
	assert.Equal(t, expected, c.Dependencies([]string{}))
}

func TestImplicitLinkAliases(t *testing.T) {
	c := &container{RunParams: RunParameters{RawLink: []string{"a"}}}
	expected := &Dependencies{
		All:  []string{"a"},
		Link: []string{"a"},
	}
	assert.Equal(t, expected, c.Dependencies([]string{}))
}

func TestVolume(t *testing.T) {
	var c *container
	// Absolute path
	c = &container{RunParams: RunParameters{RawVolume: []string{"/a:b"}}}
	assert.Equal(t, "/a:b", c.RunParams.Volume("foo")[0])
	// Relative path
	c = &container{RunParams: RunParameters{RawVolume: []string{"a:b"}}}
	dir, _ := os.Getwd()
	assert.Equal(t, dir+"/a:b", c.RunParams.Volume(dir)[0])
	// Environment variable
	c = &container{RunParams: RunParameters{RawVolume: []string{"$HOME/a:b"}}}
	os.Clearenv()
	os.Setenv("HOME", "/home")
	assert.Equal(t, os.Getenv("HOME")+"/a:b", c.RunParams.Volume("foo")[0])
	// Container-only path
	c = &container{RunParams: RunParameters{RawVolume: []string{"/b"}}}
	assert.Equal(t, "/b", c.RunParams.Volume("foo")[0])
}

func TestNet(t *testing.T) {
	var c *container
	// Empty defaults to "bridge"
	c = &container{RunParams: RunParameters{}}
	assert.Equal(t, "bridge", c.RunParams.Net())
	// Environment variable
	os.Clearenv()
	os.Setenv("NET", "container")
	c = &container{RunParams: RunParameters{RawNet: "$NET"}}
	assert.Equal(t, "container", c.RunParams.Net())
}

func TestCmd(t *testing.T) {
	var c *container
	// String
	os.Clearenv()
	os.Setenv("CMD", "true")
	c = &container{RunParams: RunParameters{RawCmd: "$CMD"}}
	assert.Equal(t, []string{"true"}, c.RunParams.Cmd())
	// String with multiple parts
	c = &container{RunParams: RunParameters{RawCmd: "bundle exec rails s -p 3000"}}
	assert.Equal(t, []string{"bundle", "exec", "rails", "s", "-p", "3000"}, c.RunParams.Cmd())
	// Array
	os.Clearenv()
	os.Setenv("CMD", "1")
	c = &container{RunParams: RunParameters{RawCmd: []interface{}{"echo", "$CMD"}}}
	if len(c.RunParams.Cmd()) != 2 || c.RunParams.Cmd()[0] != "echo" || c.RunParams.Cmd()[1] != "1" {
		t.Errorf("Command should have been true, got %v", c.RunParams.Cmd())
	}
}

type OptBoolWrapper struct {
	OptBool OptBool `json:"OptBool" yaml:"OptBool"`
}

func TestOptBoolJSON(t *testing.T) {
	wrapper := OptBoolWrapper{}
	json.Unmarshal([]byte("{\"OptBool\": true}"), &wrapper)
	assert.Equal(t, OptBool{Defined: true, Value: true}, wrapper.OptBool)

	wrapper = OptBoolWrapper{}
	json.Unmarshal([]byte("{\"OptBool\": false}"), &wrapper)
	assert.Equal(t, OptBool{Defined: true, Value: false}, wrapper.OptBool)

	wrapper = OptBoolWrapper{}
	json.Unmarshal([]byte("{}"), &wrapper)
	assert.False(t, wrapper.OptBool.Defined)

	wrapper = OptBoolWrapper{}
	err := json.Unmarshal([]byte("{\"OptBool\": \"notaboolean\"}"), &wrapper)
	assert.Error(t, err)
}

func TestOptBoolYAML(t *testing.T) {
	wrapper := OptBoolWrapper{}
	yaml.Unmarshal([]byte("OptBool: true"), &wrapper)
	assert.Equal(t, OptBool{Defined: true, Value: true}, wrapper.OptBool)

	wrapper = OptBoolWrapper{}
	yaml.Unmarshal([]byte("OptBool: false"), &wrapper)
	assert.Equal(t, OptBool{Defined: true, Value: false}, wrapper.OptBool)

	wrapper = OptBoolWrapper{}
	yaml.Unmarshal([]byte(""), &wrapper)
	assert.False(t, wrapper.OptBool.Defined)

	wrapper = OptBoolWrapper{}
	err := yaml.Unmarshal([]byte("OptBool: notaboolean"), &wrapper)
	assert.Error(t, err)
}
