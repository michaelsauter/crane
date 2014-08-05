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

func TestExpandEnv(t *testing.T) {
	rawContainerMap := ContainerMap{
		"a": Container{},
		"b": Container{},
	}
	c := &Config{RawContainerMap: rawContainerMap}
	c.expandEnv()
	if c.containerMap["a"].RawName != "a" || c.containerMap["b"].RawName != "b" {
		t.Errorf("Names should be 'a' and 'b', got %s and %s", c.containerMap["a"].RawName, c.containerMap["b"].RawName)
	}
}

func TestDetermineOrder(t *testing.T) {
	// Order set manually
	rawOrder := []string{"a", "b", "c"}
	c := &Config{RawOrder: rawOrder}
	c.expandEnv()
	c.determineOrder(false)
	if c.order[0] != "a" || c.order[1] != "b" || c.order[2] != "c" {
		t.Errorf("Order should have been %v, got %v", rawOrder, c.order)
	}
}

func TestDetermineTargetLinearChainDependencies(t *testing.T) {
	rawContainerMap := ContainerMap{
		"a": Container{Run: RunParameters{RawLink: []string{"b:b"}}},
		"b": Container{Run: RunParameters{RawLink: []string{"c:c"}}},
		"c": Container{},
	}
	c := &Config{RawContainerMap: rawContainerMap}
	c.expandEnv()
	c.determineGraph()
	c.determineTarget("a", true, false)
	if len(c.target) != 3 {
		t.Errorf("all containers should have been targeted but got %v", c.target)
	}
	c.determineTarget("b", true, false)
	if c.target[0] != "b" || c.target[1] != "c" || len(c.target) != 2 {
		t.Errorf("a should have been left out but got %v", c.target)
	}
	c.determineTarget("c", false, true)
	if len(c.target) != 3 {
		t.Errorf("all containers should have been targeted but got %v", c.target)
	}
	c.determineTarget("b", false, true)
	if c.target[0] != "a" || c.target[1] != "b" || len(c.target) != 2 {
		t.Errorf("c should have been left out but got %v", c.target)
	}
	c.determineTarget("b", true, true)
	if len(c.target) != 3 {
		t.Errorf("all containers should have been targeted but got %v", c.target)
	}
}

func TestDetermineTargetGraphDependencies(t *testing.T) {
	rawContainerMap := ContainerMap{
		"a": Container{Run: RunParameters{RawLink: []string{"b:b", "c:c"}}},
		"b": Container{Run: RunParameters{RawLink: []string{"d:d"}}},
		"c": Container{Run: RunParameters{RawLink: []string{"e:e"}}},
		"d": Container{},
		"e": Container{},
	}
	rawGroups := map[string][]string{
		"bc": []string{"b", "c"},
	}
	c := &Config{RawContainerMap: rawContainerMap, RawGroups: rawGroups}
	c.expandEnv()
	c.determineGraph()
	c.determineTarget("a", true, false)
	if len(c.target) != 5 {
		t.Errorf("all containers should have been targeted but got %v", c.target)
	}
	c.determineTarget("b", true, false)
	if c.target[0] != "b" || c.target[1] != "d" || len(c.target) != 2 {
		t.Errorf("all b and d should have been targeted but got %v", c.target)
	}
	c.determineTarget("bc", true, false)
	if c.target[0] != "b" || c.target[1] != "c" || c.target[2] != "d" || c.target[3] != "e" || len(c.target) != 4 {
		t.Errorf("a should have been left out but got %v", c.target)
	}
	c.determineTarget("bc", false, true)
	if c.target[0] != "a" || c.target[1] != "b" || c.target[2] != "c" || len(c.target) != 3 {
		t.Errorf("d and e should have been left out but got %v", c.target)
	}
	c.determineTarget("bc", true, true)
	if len(c.target) != 5 {
		t.Errorf("all containers should have been targeted but got %v", c.target)
	}
}

func TestDetermineTargetMissingDependencies(t *testing.T) {
	rawContainerMap := ContainerMap{
		"a": Container{Run: RunParameters{RawLink: []string{"b:b", "d:d"}}},
		"b": Container{Run: RunParameters{RawLink: []string{"c:c"}}},
		"c": Container{Run: RunParameters{RawLink: []string{"d:d"}}},
	}
	c := &Config{RawContainerMap: rawContainerMap}
	c.expandEnv()
	c.determineGraph()
	c.determineTarget("a", true, false)
	if len(c.target) != 3 {
		t.Errorf("only declared containers should have been targeted but got %v", c.target)
	}
	c.determineTarget("c", false, true)
	if len(c.target) != 3 {
		t.Errorf("only declared containers should have been targeted but got %v", c.target)
	}
	c.determineTarget("a", true, true)
	if len(c.target) != 3 {
		t.Errorf("only declared containers should have been targeted but got %v", c.target)
	}
}

func TestExplicitlyTargeted(t *testing.T) {
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
	c.expandEnv()
	containers = c.explicitlyTargeted("")
	if len(containers) != 2 || containers[0] != "a" || containers[1] != "b" {
		t.Errorf("Expected %v, got %v", result, containers)
	}
	// If no default group, returns all containers
	result = []string{"a", "b", "c"}
	c = &Config{RawContainerMap: rawContainerMap}
	c.expandEnv()
	containers = c.explicitlyTargeted("")
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
	c.expandEnv()
	containers = c.explicitlyTargeted("second")
	if len(containers) != 2 || containers[0] != "b" || containers[1] != "c" {
		t.Errorf("Expected %v, got %v", result, containers)
	}
	// Target is a container
	result = []string{"a"}
	containers = c.explicitlyTargeted("a")
	if len(containers) != 1 || containers[0] != "a" {
		t.Errorf("Expected %v, got %v", result, containers)
	}
}

func TestContainers(t *testing.T) {
	c := &Config{
		containerMap: ContainerMap{"a": Container{RawName: "a"}, "b": Container{RawName: "b"}},
		order:        []string{"a", "b"},
	}
	containers := c.Containers()
	if containers[0].Name() != "b" || containers[1].Name() != "a" {
		t.Errorf("Expected [b a], got %v", containers)
	}
}
