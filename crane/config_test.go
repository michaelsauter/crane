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

func TestSetNames(t *testing.T) {
	rawContainerMap := ContainerMap{
		"a": Container{},
		"b": Container{},
	}
	c := &Config{RawContainerMap: rawContainerMap}
	c.process()
	if c.containerMap["a"].RawName != "a" || c.containerMap["b"].RawName != "b" {
		t.Errorf("Names should be 'a' and 'b', got %s and %s", c.containerMap["a"].RawName, c.containerMap["b"].RawName)
	}
}

func TestDetermineOrder(t *testing.T) {
	// Order set manually
	rawOrder := []string{"a", "b", "c"}
	c := &Config{RawOrder: rawOrder}
	c.process()
	c.determineOrder(false)
	if c.order[0] != "a" || c.order[1] != "b" || c.order[2] != "c" {
		t.Errorf("Order should have been %v, got %v", rawOrder, c.order)
	}
}

func TestTargetedContainers(t *testing.T) {
	var result []string
	var containers []string
	var rawGroups = make(map[string][]string)
	rawContainerMap := ContainerMap{
		"a": Container{},
		"b": Container{},
		"c": Container{},
	}

	// No target given
	// If default group exist, it returns its containers
	result = []string{"a", "b"}
	rawGroups = map[string][]string{
		"default": result,
	}
	c := &Config{RawGroups: rawGroups}
	c.process()
	containers = c.targetedContainers("")
	if len(containers) != 2 || containers[0] != "a" || containers[1] != "b" {
		t.Errorf("Expected %v, got %v", result, containers)
	}
	// If no default group, returns all containers
	result = []string{"a", "b", "c"}
	c = &Config{RawContainerMap: rawContainerMap}
	c.process()
	containers = c.targetedContainers("")
	if len(containers) != 3 || containers[0] != "a" || containers[1] != "b" || containers[2] != "c" {
		t.Errorf("Expected %v, got %v", result, containers)
	}
	// Target given
	// Target is a group
	result = []string{"b", "c"}
	rawGroups = map[string][]string{
		"second": result,
	}
	c = &Config{RawContainerMap: rawContainerMap, RawGroups: rawGroups}
	c.process()
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
