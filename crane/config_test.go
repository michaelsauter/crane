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
	if len(files) != 3 {
		t.Errorf("Config files should be [crane.json crane.yaml crane.yml], got %v", files)
	}
}

func TestSetContainersByNames(t *testing.T) {
	containerMap := &ContainerMap{
		"a": Container{},
		"b": Container{},
	}
	c := &Config{ContainersByRawName: containerMap}
	c.setContainersByName()
	cMap := *c.containersByName
	if cMap["a"].RawName != "a" || cMap["b"].RawName != "b" {
		t.Errorf("Names should be 'a' and 'b', got %s and %s", cMap["a"].RawName, cMap["b"].RawName)
	}
}

func TestDetermineOrder(t *testing.T) {
	// Order set manually
	rawOrder := []string{"a", "${DOES_NOT_EXIST}b", "c"}
	c := &Config{RawOrder: rawOrder}
	c.determineOrder()
	if c.order[0] != "a" || c.order[1] != "b" || c.order[2] != "c" {
		t.Errorf("Order should have been %v, got %v", c.order, c.RawOrder)
	}
	// Resolvable map
	containerMap := &ContainerMap{
		"b":                  Container{Run: RunParameters{RawLink: []string{"${DOES_NOT_EXIST}c:c"}}},
		"${DOES_NOT_EXIST}a": Container{Run: RunParameters{RawLink: []string{"b:b"}}},
		"c":                  Container{},
	}
	c = &Config{ContainersByRawName: containerMap}
	c.setContainersByName()
	err := c.determineOrder()
	if err != nil || c.order[0] != "a" || c.order[1] != "b" || c.order[2] != "c" {
		t.Errorf("Order should have been [a, b, c], got %v", c.RawOrder)
	}
	// Unresolvable map returns error
	containerMap = &ContainerMap{
		"b": Container{Run: RunParameters{RawLink: []string{"c:c"}}},
		"a": Container{Run: RunParameters{RawLink: []string{"b:b"}}},
		"c": Container{Run: RunParameters{RawLink: []string{"a:a"}}},
	}
	c = &Config{ContainersByRawName: containerMap}
	c.setContainersByName()
	err = c.determineOrder()
	if err == nil {
		t.Errorf("Cyclic dependency a -> b -> c -> a should not have been resolvable, got %v", c.RawOrder)
	}
}

func TestFilter(t *testing.T) {
	containerMap := &ContainerMap{
		"a":                  Container{},
		"${DOES_NOT_EXIST}b": Container{},
		"c":                  Container{},
	}
	c := &Config{ContainersByRawName: containerMap}
	c.setContainersByName()
	c.filter("b")
	containersByName := *c.containersByName
	if len(containersByName) != 1 {
		t.Errorf("Only one container should remain after filtering, got %v", c.containersByName)
	}
}

func TestTargetedContainers(t *testing.T) {
	var result []string
	var containers []string
	var groups = make(map[string][]string)
	containerMap := &ContainerMap{
		"a":                  Container{},
		"${DOES_NOT_EXIST}b": Container{},
		"c":                  Container{},
	}

	// No target given
	// If default groups exist, it returns its containers
	result = []string{"a", "${DOES_NOT_EXIST}b"}
	groups = map[string][]string{
		"default": result,
	}
	c := &Config{RawGroups: groups}
	containers = c.targetedContainers("")
	if len(containers) != len(result) || containers[0] != result[0] || containers[1] != result[1] {
		t.Errorf("Expected %v, got %v", result, containers)
	}
	// If no default group, returns all containers
	result = []string{"a", "${DOES_NOT_EXIST}b", "c"}
	c = &Config{ContainersByRawName: containerMap}
	containers = c.targetedContainers("")
	if len(containers) != len(result) || containers[0] != result[0] || containers[1] != result[1] || containers[2] != result[2] {
		t.Errorf("Expected %v, got %v", result, containers)
	}
	// Target given
	// Target is a group
	result = []string{"${DOES_NOT_EXIST}b", "c"}
	groups = map[string][]string{
		"second": result,
	}
	c = &Config{ContainersByRawName: containerMap, RawGroups: groups}
	containers = c.targetedContainers("second")
	if len(containers) != len(result) || containers[0] != result[0] || containers[1] != result[1] {
		t.Errorf("Expected %v, got %v", result, containers)
	}
	// Target is a container
	result = []string{"a"}
	containers = c.targetedContainers("a")
	if len(containers) != len(result) || containers[0] != result[0] {
		t.Errorf("Expected %v, got %v", result, containers)
	}
}
