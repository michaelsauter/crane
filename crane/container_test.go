package crane

import (
	"os"
	"testing"
)

func TestDependencies(t *testing.T) {
	c := &container{RunParams: RunParameters{RawNet: "container:n", RawLink: []string{"a:b", "b:d"}, RawVolumesFrom: []string{"c"}}}
	if deps := c.Dependencies(); deps.all[0] != "a" || deps.all[1] != "b" || deps.all[2] != "c" || deps.all[3] != "n" || deps.link[0] != "a" || deps.link[1] != "b" || deps.all[2] != "c" || deps.net != "n" {
		t.Errorf("Dependencies should have been a, b, c, n. Got %v", deps)
	}
	c = &container{RunParams: RunParameters{RawLink: []string{}, RawVolumesFrom: []string{}}}
	if deps := c.Dependencies(); len(deps.all) != 0 && len(deps.link) != 0 && deps.net == "" {
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
}
