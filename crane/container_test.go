package crane

import "testing"

func TestDependencies(t *testing.T) {
	container := &Container{Run: RunParameters{RawLink: []string{"a:b", "b:d"}}}
	if deps := container.Dependencies(); deps[0] != "a" || deps[1] != "b" {
		t.Error("Dependencies should have been a and b")
	}
	container = &Container{Run: RunParameters{RawLink: []string{}}}
	if deps := container.Dependencies(); len(deps) != 0 {
		t.Error("Dependencies should have been empty")
	}
}

func TestIsTargeted(t *testing.T) {
	container := &Container{RawName: "a"}
	if container.IsTargeted([]string{"b"}) {
		t.Error("Container name was a, got targeted with b")
	}
	if !container.IsTargeted([]string{"x", "a"}) {
		t.Error("Container name was a, should have been targeted with a")
	}
}
