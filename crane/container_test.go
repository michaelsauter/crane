package crane

import (
	"encoding/json"
	"gopkg.in/v2/yaml"
	"os"
	"testing"
)

func TestDependencies(t *testing.T) {
	c := &container{RunParams: RunParameters{RawNet: "container:n", RawLink: []string{"a:b", "b:d"}, RawVolumesFrom: []string{"c"}}}
	if deps := c.Dependencies(); deps.All[0] != "a" || deps.All[1] != "b" || deps.All[2] != "c" || deps.All[3] != "n" || deps.Link[0] != "a" || deps.Link[1] != "b" || deps.All[2] != "c" || deps.Net != "n" {
		t.Errorf("Dependencies should have been a, b, c, n. Got %v", deps)
	}
	c = &container{RunParams: RunParameters{RawLink: []string{}, RawVolumesFrom: []string{}}}
	if deps := c.Dependencies(); len(deps.All) != 0 && len(deps.Link) != 0 && deps.Net == "" {
		t.Error("Dependencies should have been empty")
	}
}

func TestVolume(t *testing.T) {
	var c *container
	// Absolute path
	c = &container{RunParams: RunParameters{RawVolume: []string{"/a:b"}}}
	if c.RunParams.Volume()[0] != "/a:b" {
		t.Errorf("Volume mapping should have been a:b, was %v", c.RunParams.Volume()[0])
	}
	// Relative path
	c = &container{RunParams: RunParameters{RawVolume: []string{"a:b"}}}
	dir, _ := os.Getwd()
	if c.RunParams.Volume()[0] != (dir + "/a:b") {
		t.Errorf("Volume mapping should have been pwd/a:b, was %v", c.RunParams.Volume()[0])
	}
	// Environment variable
	c = &container{RunParams: RunParameters{RawVolume: []string{"$HOME/a:b"}}}
	os.Clearenv()
	os.Setenv("HOME", "/home")
	if c.RunParams.Volume()[0] != (os.Getenv("HOME") + "/a:b") {
		t.Errorf("Volume mapping should have been $HOME/a:b, was %v", c.RunParams.Volume()[0])
	}
	// Container-only path
	c = &container{RunParams: RunParameters{RawVolume: []string{"/b"}}}
	if c.RunParams.Volume()[0] != "/b" {
		t.Errorf("Volume mapping should have been /b, was %v", c.RunParams.Volume()[0])
	}
}

func TestNet(t *testing.T) {
	var c *container
	// Empty defaults to "bridge"
	c = &container{RunParams: RunParameters{}}
	if c.RunParams.Net() != "bridge" {
		t.Errorf("Net should have been bridge, got %v", c.RunParams.Net())
	}
	// Environment variable
	os.Clearenv()
	os.Setenv("NET", "container")
	c = &container{RunParams: RunParameters{RawNet: "$NET"}}
	if c.RunParams.Net() != "container" {
		t.Errorf("Net should have been container, got %v", c.RunParams.Net())
	}
}

func TestCmd(t *testing.T) {
	var c *container
	// String
	os.Clearenv()
	os.Setenv("CMD", "true")
	c = &container{RunParams: RunParameters{RawCmd: "$CMD"}}
	if len(c.RunParams.Cmd()) != 1 || c.RunParams.Cmd()[0] != "true" {
		t.Errorf("Command should have been true, got %v", c.RunParams.Cmd())
	}
	// String with multiple parts
	c = &container{RunParams: RunParameters{RawCmd: "bundle exec rails s -p 3000"}}
	if len(c.RunParams.Cmd()) != 6 || c.RunParams.Cmd()[0] != "bundle" || c.RunParams.Cmd()[1] != "exec" || c.RunParams.Cmd()[2] != "rails" || c.RunParams.Cmd()[3] != "s" || c.RunParams.Cmd()[4] != "-p" || c.RunParams.Cmd()[5] != "3000" {
		t.Errorf("Command should have been [bundle exec rails s -p 3000], got %v", c.RunParams.Cmd())
	}
	// Array
	os.Clearenv()
	os.Setenv("CMD", "1")
	c = &container{RunParams: RunParameters{RawCmd: []interface{}{"echo", "$CMD"}}}
	if len(c.RunParams.Cmd()) != 2 || c.RunParams.Cmd()[0] != "echo" || c.RunParams.Cmd()[1] != "1" {
		t.Errorf("Command should have been true, got %v", c.RunParams.Cmd())
	}
}

func TestEnv(t *testing.T) {
	var c *container
	// Array
	os.Clearenv()
	os.Setenv("FOO", "fooKey=fooValue")
	c = &container{RunParams: RunParameters{RawEnv: []interface{}{"$FOO", "barKey=barValue"}}}
	if len(c.RunParams.Env()) != 2 || c.RunParams.Env()[0] != "fooKey=fooValue" || c.RunParams.Env()[1] != "barKey=barValue" {
		t.Errorf("Env should have been [fooKey=fooValue barKey=barValue], got %v", c.RunParams.Env())
	}
	// Mapping
	os.Clearenv()
	os.Setenv("FOO_KEY", "fooKey")
	os.Setenv("FOO_VALUE", "fooValue")
	c = &container{RunParams: RunParameters{RawEnv: map[interface{}]interface{}{"$FOO_KEY": "$FOO_VALUE", "barKey": "barValue"}}}
	if len(c.RunParams.Env()) != 2 || c.RunParams.Env()[0] != "fooKey=fooValue" || c.RunParams.Env()[1] != "barKey=barValue" {
		t.Errorf("Env should have been [fooKey=fooValue barKey=barValue], got %v", c.RunParams.Env())
	}
}

type OptBoolWrapper struct {
	OptBool OptBool `json:"OptBool" yaml:"OptBool"`
}

func TestOptBoolJSON(t *testing.T) {
	wrapper := OptBoolWrapper{}
	json.Unmarshal([]byte("{\"OptBool\": true}"), &wrapper)
	if !wrapper.OptBool.Defined || !wrapper.OptBool.Value {
		t.Errorf("OptBool should have been defined and true, got %v", wrapper.OptBool)
	}

	wrapper = OptBoolWrapper{}
	json.Unmarshal([]byte("{\"OptBool\": false}"), &wrapper)
	if !wrapper.OptBool.Defined || wrapper.OptBool.Value {
		t.Errorf("OptBool should have been defined and false, got %v", wrapper.OptBool)
	}

	wrapper = OptBoolWrapper{}
	json.Unmarshal([]byte("{}"), &wrapper)
	if wrapper.OptBool.Defined {
		t.Errorf("OptBool should have been undefined, got %v", wrapper.OptBool)
	}

	wrapper = OptBoolWrapper{}
	err := json.Unmarshal([]byte("{\"OptBool\": \"notaboolean\"}"), &wrapper)
	if err == nil {
		t.Errorf("Error expected but not found")
	}
}

func TestOptBoolYAML(t *testing.T) {
	wrapper := OptBoolWrapper{}
	yaml.Unmarshal([]byte("OptBool: true"), &wrapper)
	if !wrapper.OptBool.Defined || !wrapper.OptBool.Value {
		t.Errorf("OptBool should have been defined and true, got %v", wrapper.OptBool)
	}

	wrapper = OptBoolWrapper{}
	yaml.Unmarshal([]byte("OptBool: false"), &wrapper)
	if !wrapper.OptBool.Defined || wrapper.OptBool.Value {
		t.Errorf("OptBool should have been defined and false, got %v", wrapper.OptBool)
	}

	wrapper = OptBoolWrapper{}
	yaml.Unmarshal([]byte(""), &wrapper)
	if wrapper.OptBool.Defined {
		t.Errorf("OptBool should have been undefined, got %v", wrapper.OptBool)
	}

	wrapper = OptBoolWrapper{}
	err := yaml.Unmarshal([]byte("OptBool: notaboolean"), &wrapper)
	if err == nil {
		t.Errorf("Error expected but not found")
	}
}
