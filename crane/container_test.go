package crane

import (
	"os"
	"testing"
)

func TestDependencies(t *testing.T) {
	container := &Container{Run: RunParameters{RawNet: "container:n", RawLink: []string{"a:b", "b:d"}, RawVolumesFrom: []string{"c"}}}
	if deps := container.Dependencies(); deps.list[0] != "a" || deps.list[1] != "b" || deps.list[2] != "c" || deps.list[3] != "n" || deps.linked[0] != "a" || deps.linked[1] != "b" || deps.list[2] != "c" || deps.net != "n" {
		t.Errorf("Dependencies should have been a, b, c, n. Got %v", deps)
	}
	container = &Container{Run: RunParameters{RawLink: []string{}, RawVolumesFrom: []string{}}}
	if deps := container.Dependencies(); len(deps.list) != 0 && len(deps.linked) != 0 && deps.net == "" {
		t.Error("Dependencies should have been empty")
	}
}

func TestVolume(t *testing.T) {
	var container *Container
	// Absolute path
	container = &Container{Run: RunParameters{RawVolume: []string{"/a:b"}}}
	if container.Run.Volume()[0] != "/a:b" {
		t.Errorf("Volume mapping should have been a:b, was %v", container.Run.Volume()[0])
	}
	// Relative path
	container = &Container{Run: RunParameters{RawVolume: []string{"a:b"}}}
	dir, _ := os.Getwd()
	if container.Run.Volume()[0] != (dir + "/a:b") {
		t.Errorf("Volume mapping should have been pwd/a:b, was %v", container.Run.Volume()[0])
	}
	// Environment variable
	container = &Container{Run: RunParameters{RawVolume: []string{"$HOME/a:b"}}}
	os.Clearenv()
	os.Setenv("HOME", "/home")
	if container.Run.Volume()[0] != (os.Getenv("HOME") + "/a:b") {
		t.Errorf("Volume mapping should have been $HOME/a:b, was %v", container.Run.Volume()[0])
	}
}
