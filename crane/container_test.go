package crane

import "testing"

func TestDependencies(t *testing.T) {
	container := &Container{Run: RunParameters{RawLink: []string{"a:b", "b:d"}, RawVolumesFrom: []string{"c"}}}
	if deps := container.Dependencies(); deps[0] != "a" || deps[1] != "b" || deps[2] != "c" {
		t.Errorf("Dependencies should have been a, b and c. Got %v", deps)
	}
	container = &Container{Run: RunParameters{RawLink: []string{"${DOES_NOT_EXIST}a:b", "${DOES_NOT_EXIST_EITHER}b:d"}}}
	if deps := container.Dependencies(); deps[0] != "a" || deps[1] != "b" {
		t.Errorf("Dependencies should have been [a b]. Got %v", deps)
	}
	container = &Container{Run: RunParameters{RawLink: []string{}, RawVolumesFrom: []string{}}}
	if deps := container.Dependencies(); len(deps) != 0 {
		t.Error("Dependencies should have been empty. Got %v", deps)
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
