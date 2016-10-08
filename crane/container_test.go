package crane

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"gopkg.in/v2/yaml"
	"os"
	"testing"
)

func TestDependencies(t *testing.T) {
	defer func() {
		allowed = []string{}
	}()

	c := &container{}
	expected := &Dependencies{}

	// no dependencies
	assert.Equal(t, expected, c.Dependencies())

	// network v2 links
	allowed = []string{"foo", "bar", "c"}
	c = &container{
		RawRequires: []string{"foo", "bar"},
		RawRun: RunParameters{
			RawNet:         "network",
			RawLink:        []string{"a:b", "b:d"},
			RawVolumesFrom: []string{"c"},
		},
	}
	expected = &Dependencies{
		All:         []string{"foo", "bar", "c"},
		Requires:    []string{"foo", "bar"},
		VolumesFrom: []string{"c"},
	}
	assert.Equal(t, expected, c.Dependencies())

	// legacy links
	allowed = []string{"a", "b", "c"}
	c = &container{
		RawRun: RunParameters{
			RawLink:        []string{"a:b", "b:d"},
			RawVolumesFrom: []string{"c"},
		},
	}
	expected = &Dependencies{
		All:         []string{"a", "b", "c"},
		Link:        []string{"a", "b"},
		VolumesFrom: []string{"c"},
	}
	assert.Equal(t, expected, c.Dependencies())

	// container network
	allowed = []string{"c", "n"}
	c = &container{
		RawRun: RunParameters{
			RawNet:         "container:n",
			RawVolumesFrom: []string{"c"},
		},
	}
	expected = &Dependencies{
		All:         []string{"c", "n"},
		VolumesFrom: []string{"c"},
		Net:         "n",
	}
	assert.Equal(t, expected, c.Dependencies())

	// with restricted allowed containers
	allowed = []string{"foo", "c"}
	c = &container{
		RawRequires: []string{"foo", "bar"},
		RawRun: RunParameters{
			RawLink:        []string{"a:b", "b:d"},
			RawVolumesFrom: []string{"c", "d"},
		},
	}
	expected = &Dependencies{
		All:         []string{"foo", "c"},
		Requires:    []string{"foo"},
		VolumesFrom: []string{"c"},
	}
	assert.Equal(t, expected, c.Dependencies())
}

func TestVolumesFromSuffixes(t *testing.T) {
	defer func() {
		allowed = []string{}
	}()
	allowed = []string{"a", "b"}
	c := &container{RawRun: RunParameters{RawVolumesFrom: []string{"a:rw", "b:ro"}}}
	expected := &Dependencies{
		All:         []string{"a", "b"},
		VolumesFrom: []string{"a", "b"},
	}
	assert.Equal(t, expected, c.Dependencies())
}

func TestMultipleLinkAliases(t *testing.T) {
	defer func() {
		allowed = []string{}
	}()
	allowed = []string{"a"}
	c := &container{RawRun: RunParameters{RawLink: []string{"a:b", "a:c"}}}
	expected := &Dependencies{
		All:  []string{"a"},
		Link: []string{"a"},
	}
	assert.Equal(t, expected, c.Dependencies())
}

func TestImplicitLinkAliases(t *testing.T) {
	defer func() {
		allowed = []string{}
	}()
	allowed = []string{"a"}
	c := &container{RawRun: RunParameters{RawLink: []string{"a"}}}
	expected := &Dependencies{
		All:  []string{"a"},
		Link: []string{"a"},
	}
	assert.Equal(t, expected, c.Dependencies())
}

func TestImage(t *testing.T) {
	containers := []*container{
		&container{RawName: "full-spec", RawImage: "test/image-a:1.0"},
		&container{RawName: "without-repo", RawImage: "image-b:latest"},
		&container{RawName: "without-tag", RawImage: "test/image-c"},
		&container{RawName: "image-only", RawImage: "image-d"},
		&container{RawName: "private-registry", RawImage: "localhost:5000/foo/image-e:2.0"},
		&container{RawName: "private-registry-without-tag", RawImage: "localhost:5000/foo/image-e"},
		&container{RawName: "digest", RawImage: "localhost:5000/foo/image-f@sha256:xxx"},
	}
	containerMap := make(map[string]*container)
	for _, container := range containers {
		containerMap[container.Name()] = container
	}
	cfg = &config{
		tag: "rc-1",
	}

	assert.Equal(t, "test/image-a:rc-1", containerMap["full-spec"].Image())

	assert.Equal(t, "image-b:rc-1", containerMap["without-repo"].Image())

	assert.Equal(t, "test/image-c:rc-1", containerMap["without-tag"].Image())

	assert.Equal(t, "image-d:rc-1", containerMap["image-only"].Image())

	assert.Equal(t, "localhost:5000/foo/image-e:rc-1", containerMap["private-registry"].Image())

	assert.Equal(t, "localhost:5000/foo/image-e:rc-1", containerMap["private-registry-without-tag"].Image())

	assert.NotEqual(t, "localhost:5000/foo/image-f@sha256:rc-1", containerMap["digest"].Image())
}

func TestVolume(t *testing.T) {
	var c *container
	// Absolute path
	c = &container{RawRun: RunParameters{RawVolume: []string{"/a:/b"}}}
	cfg = &config{path: "foo"}
	assert.Equal(t, "/a:/b", c.RunParams().Volume()[0])
	// Environment variable
	c = &container{RawRun: RunParameters{RawVolume: []string{"$HOME/a:/b"}}}
	os.Clearenv()
	os.Setenv("HOME", "/home")
	cfg = &config{path: "foo"}
	assert.Equal(t, os.Getenv("HOME")+"/a:/b", c.RunParams().Volume()[0])
}

func TestActualVolume(t *testing.T) {
	var c *container
	// Simple case
	c = &container{RawRun: RunParameters{RawVolume: []string{"/a:/b"}}}
	cfg = &config{path: "foo"}
	assert.Equal(t, "/a:/b", c.RunParams().ActualVolume()[0])
	// Relative path
	c = &container{RawRun: RunParameters{RawVolume: []string{"a:/b"}}}
	dir, _ := os.Getwd()
	cfg = &config{path: dir}
	assert.Equal(t, dir+"/a:/b", c.RunParams().ActualVolume()[0])
	// Container-only path
	c = &container{RawRun: RunParameters{RawVolume: []string{"/b"}}}
	assert.Equal(t, "/b", c.RunParams().Volume()[0])
	// Using Docker volume
	c = &container{RawRun: RunParameters{RawVolume: []string{"a:/b"}}}
	cfg = &config{volumeMap: map[string]Volume{"a": &volume{RawName: "a"}}}
	assert.Equal(t, "a:/b", c.RunParams().Volume()[0])
	// With prefix Docker volume
	c = &container{RawRun: RunParameters{RawVolume: []string{"a:/b"}}}
	cfg = &config{prefix: "foo_", volumeMap: map[string]Volume{"a": &volume{RawName: "a"}}}
	assert.Equal(t, "foo_a:/b", c.RunParams().ActualVolume()[0])
}

func TestNet(t *testing.T) {
	var c *container
	// Empty defaults to "bridge"
	c = &container{RawRun: RunParameters{}}
	assert.Equal(t, "bridge", c.RunParams().Net())
	// Environment variable
	os.Clearenv()
	os.Setenv("NET", "container")
	c = &container{RawRun: RunParameters{RawNet: "$NET"}}
	assert.Equal(t, "container", c.RunParams().Net())
}

func TestActualNet(t *testing.T) {
	var c *container
	// Empty defaults to "bridge"
	c = &container{RawRun: RunParameters{}}
	assert.Equal(t, "bridge", c.RunParams().ActualNet())
	// Container
	c = &container{RawName: "foo", RawRun: RunParameters{RawNet: "container:foo"}}
	cfg = &config{containerMap: map[string]Container{"foo": c}}
	assert.Equal(t, "container:foo", c.RunParams().ActualNet())
	// Network
	c = &container{RawName: "foo", RawRun: RunParameters{RawNet: "bar"}}
	cfg = &config{
		containerMap: map[string]Container{"foo": c},
		networkMap:   map[string]Network{"bar": &network{RawName: "bar"}},
	}
	assert.Equal(t, "bar", c.RunParams().ActualNet())
	// Network with prefix
	cfg = &config{
		prefix:       "qux_",
		containerMap: map[string]Container{"foo": c},
		networkMap:   map[string]Network{"bar": &network{RawName: "bar"}},
	}
	assert.Equal(t, "qux_bar", c.RunParams().ActualNet())
}

func TestCmd(t *testing.T) {
	var c *container
	// String
	os.Clearenv()
	os.Setenv("CMD", "true")
	c = &container{RawRun: RunParameters{RawCmd: "$$CMD is $CMD"}}
	assert.Equal(t, []string{"$CMD", "is", "true"}, c.RunParams().Cmd())
	// String with multiple parts
	c = &container{RawRun: RunParameters{RawCmd: "bundle exec rails s -p 3000"}}
	assert.Equal(t, []string{"bundle", "exec", "rails", "s", "-p", "3000"}, c.RunParams().Cmd())
	// Array
	os.Clearenv()
	os.Setenv("CMD", "1")
	c = &container{RawRun: RunParameters{RawCmd: []interface{}{"echo", "$CMD", "$$CMD"}}}
	assert.Equal(t, []string{"echo", "1", "$CMD"}, c.RunParams().Cmd())
}

type OptIntWrapper struct {
	OptInt OptInt `json:"OptInt" yaml:"OptInt"`
}

func TestOptIntJSON(t *testing.T) {
	wrapper := OptIntWrapper{}
	json.Unmarshal([]byte("{\"OptInt\": 1}"), &wrapper)
	assert.Equal(t, OptInt{Defined: true, Value: 1}, wrapper.OptInt)

	wrapper = OptIntWrapper{}
	json.Unmarshal([]byte("{}"), &wrapper)
	assert.False(t, wrapper.OptInt.Defined)

	wrapper = OptIntWrapper{}
	err := json.Unmarshal([]byte("{\"OptInt\": \"notanumber\"}"), &wrapper)
	assert.Error(t, err)
}

func TestOptIntYAML(t *testing.T) {
	wrapper := OptIntWrapper{}
	yaml.Unmarshal([]byte("OptInt: 1"), &wrapper)
	assert.Equal(t, OptInt{Defined: true, Value: 1}, wrapper.OptInt)

	wrapper = OptIntWrapper{}
	yaml.Unmarshal([]byte(""), &wrapper)
	assert.False(t, wrapper.OptInt.Defined)

	wrapper = OptIntWrapper{}
	err := yaml.Unmarshal([]byte("OptInt: notanumber"), &wrapper)
	assert.Error(t, err)
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

func TestBuildArgs(t *testing.T) {
	var c *container
	c = &container{RawBuild: BuildParameters{RawBuildArgs: []interface{}{"key1=value1"}}}
	cfg = &config{path: "foo"}
	assert.Equal(t, "key1=value1", c.BuildParams().BuildArgs()[0])
}
