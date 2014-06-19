package crane

import "testing"

func TestConfigFiles(t *testing.T) {
	// With given filename
	filename := "some/file.yml"
	options := Options{config: filename}
	files := configFiles(options)
	if len(files) > 1 {
		t.Errorf("Config files should be just [%s], got %v", filename, files)
	}
	// Without given filename
	files = configFiles(Options{})
	if len(files) != 4 {
		t.Errorf("Config files should be [crane.json, crane.yaml, crane.yml, Cranefile], got %v", files)
	}
}

func TestSetNames(t *testing.T) {
	containerMap := &ContainerMap{
		"a": Container{},
		"b": Container{},
	}
	c := &Config{ContainerMap: containerMap}
	c.setNames()
	cMap := *c.ContainerMap
	if cMap["a"].RawName != "a" || cMap["b"].RawName != "b" {
		t.Errorf("Names should be 'a' and 'b', got %s and %s", cMap["a"].RawName, cMap["b"].RawName)
	}
}

func TestDetermineOrder(t *testing.T) {
	// Order set manually
	order := []string{"a", "b", "c"}
	c := &Config{Order: order}
	c.determineOrder()
	if order[0] != "a" || order[1] != "b" || order[2] != "c" {
		t.Errorf("Order should have been %v, got %v", order, c.Order)
	}
	// Resolvable map
	containerMap := &ContainerMap{
		"b": Container{Run: RunParameters{RawLink: []string{"c:c"}}},
		"a": Container{Run: RunParameters{RawLink: []string{"b:b"}}},
		"c": Container{},
	}
	c = &Config{ContainerMap: containerMap}
	err := c.determineOrder()
	if err != nil || order[0] != "a" || order[1] != "b" || order[2] != "c" {
		t.Errorf("Order should have been [a, b, c], got %v", c.Order)
	}
	// Unresolvable map returns error
	containerMap = &ContainerMap{
		"b": Container{Run: RunParameters{RawLink: []string{"c:c"}}},
		"a": Container{Run: RunParameters{RawLink: []string{"b:b"}}},
		"c": Container{Run: RunParameters{RawLink: []string{"a:a"}}},
	}
	c = &Config{ContainerMap: containerMap}
	err = c.determineOrder()
	if err == nil {
		t.Errorf("Cyclic dependency a -> b -> c -> a should not have been resolvable, got %v", c.Order)
	}
}

func TestTargetedContainers(t *testing.T) {
	var result []string
	var containers []string
	var groups = make(map[string][]string)
	containerMap := &ContainerMap{
		"a": Container{},
		"b": Container{},
		"c": Container{},
	}

	// No target given
	// If default groups exist, it returns its containers
	result = []string{"a", "b"}
	groups = map[string][]string{
		"default": result,
	}
	c := &Config{Groups: groups}
	containers = c.targetedContainers("")
	if len(containers) != 2 || containers[0] != "a" || containers[1] != "b" {
		t.Errorf("Expected %v, got %v", result, containers)
	}
	// If no default group, returns all containers
	result = []string{"a", "b", "c"}
	c = &Config{ContainerMap: containerMap}
	containers = c.targetedContainers("")
	if len(containers) != 3 || containers[0] != "a" || containers[1] != "b" || containers[2] != "c" {
		t.Errorf("Expected %v, got %v", result, containers)
	}
	// Target given
	// Target is a group
	result = []string{"b", "c"}
	groups = map[string][]string{
		"second": result,
	}
	c = &Config{ContainerMap: containerMap, Groups: groups}
	containers = c.targetedContainers("second")
	if len(containers) != 2 || containers[0] != "b" || containers[1] != "c" {
		t.Errorf("Expected %v, got %v", result, containers)
	}
	// Target is a container
	result = []string{"a"}
	containers = c.targetedContainers("a")
	if len(containers) != 1 || containers[0] != "a" {
		t.Errorf("Expected %v, got %v", result, containers)
	}
}
