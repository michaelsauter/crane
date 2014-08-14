package crane

import (
	"reflect"
	"testing"
)

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
		"a": &container{},
		"b": &container{},
	}
	c := &Config{RawContainerMap: rawContainerMap}
	c.expandEnv()
	if c.containerMap["a"].Name() != "a" || c.containerMap["b"].Name() != "b" {
		t.Errorf("Names should be 'a' and 'b', got %s and %s", c.containerMap["a"].Name(), c.containerMap["b"].Name())
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
		"a": &container{RunParams: RunParameters{RawLink: []string{"b:b"}}},
		"b": &container{RunParams: RunParameters{RawLink: []string{"c:c"}}},
		"c": &container{},
	}
	c := &Config{RawContainerMap: rawContainerMap}
	c.expandEnv()
	c.determineGraph()

	examples := []struct {
		target              string
		cascadeDependencies string
		cascadeAffected     string
		expected            Target
	}{
		{
			target:              "a",
			cascadeDependencies: "all",
			cascadeAffected:     "none",
			expected:            []string{"a", "b", "c"},
		},
		{
			target:              "b",
			cascadeDependencies: "all",
			cascadeAffected:     "none",
			expected:            []string{"b", "c"},
		},
		{
			target:              "c",
			cascadeDependencies: "none",
			cascadeAffected:     "all",
			expected:            []string{"a", "b", "c"},
		},
		{
			target:              "b",
			cascadeDependencies: "none",
			cascadeAffected:     "all",
			expected:            []string{"a", "b"},
		},
		{
			target:              "b",
			cascadeDependencies: "all",
			cascadeAffected:     "all",
			expected:            []string{"a", "b", "c"},
		},
	}

	for _, example := range examples {
		c.determineTarget(example.target, example.cascadeDependencies, example.cascadeAffected)
		if !reflect.DeepEqual(c.target, example.expected) {
			t.Errorf("Target should have been %v, got %v", example.expected, c.target)
		}
	}
}

func TestDetermineTargetGraphDependencies(t *testing.T) {
	rawContainerMap := ContainerMap{
		"a": &container{RunParams: RunParameters{RawLink: []string{"b:b", "c:c"}}},
		"b": &container{RunParams: RunParameters{RawLink: []string{"d:d"}}},
		"c": &container{RunParams: RunParameters{RawLink: []string{"e:e"}}},
		"d": &container{},
		"e": &container{},
	}
	rawGroups := map[string][]string{
		"bc": []string{"b", "c"},
	}
	c := &Config{RawContainerMap: rawContainerMap, RawGroups: rawGroups}
	c.expandEnv()
	c.determineGraph()
	c.determineTarget("a", "all", "none")
	if len(c.target) != 5 {
		t.Errorf("all containers should have been targeted but got %v", c.target)
	}
	c.determineTarget("b", "all", "none")
	if c.target[0] != "b" || c.target[1] != "d" || len(c.target) != 2 {
		t.Errorf("all b and d should have been targeted but got %v", c.target)
	}
	c.determineTarget("bc", "all", "none")
	if c.target[0] != "b" || c.target[1] != "c" || c.target[2] != "d" || c.target[3] != "e" || len(c.target) != 4 {
		t.Errorf("a should have been left out but got %v", c.target)
	}
	c.determineTarget("bc", "none", "all")
	if c.target[0] != "a" || c.target[1] != "b" || c.target[2] != "c" || len(c.target) != 3 {
		t.Errorf("d and e should have been left out but got %v", c.target)
	}
	c.determineTarget("bc", "all", "all")
	if len(c.target) != 5 {
		t.Errorf("all containers should have been targeted but got %v", c.target)
	}
}

func TestDetermineTargetMissingDependencies(t *testing.T) {
	rawContainerMap := ContainerMap{
		"a": &container{RunParams: RunParameters{RawLink: []string{"b:b", "d:d"}}},
		"b": &container{RunParams: RunParameters{RawLink: []string{"c:c"}}},
		"c": &container{RunParams: RunParameters{RawLink: []string{"d:d"}}},
	}
	c := &Config{RawContainerMap: rawContainerMap}
	c.expandEnv()
	c.determineGraph()
	c.determineTarget("a", "all", "none")
	if len(c.target) != 3 {
		t.Errorf("only declared containers should have been targeted but got %v", c.target)
	}
	c.determineTarget("c", "none", "all")
	if len(c.target) != 3 {
		t.Errorf("only declared containers should have been targeted but got %v", c.target)
	}
	c.determineTarget("a", "all", "all")
	if len(c.target) != 3 {
		t.Errorf("only declared containers should have been targeted but got %v", c.target)
	}
}

func TestDetermineTargetCustomCascading(t *testing.T) {
	rawContainerMap := ContainerMap{
		"linkSource":        &container{RunParams: RunParameters{RawLink: []string{"x:x"}}},
		"netSource":         &container{RunParams: RunParameters{RawNet: "container:x"}},
		"volumesFromSource": &container{RunParams: RunParameters{RawVolumesFrom: []string{"x"}}},
		"x":                 &container{RunParams: RunParameters{RawLink: []string{"linkTarget:linkTarget"}, RawNet: "container:netTarget", RawVolumesFrom: []string{"volumesFromTarget"}}},
		"linkTarget":        &container{},
		"netTarget":         &container{},
		"volumesFromTarget": &container{},
	}
	c := &Config{RawContainerMap: rawContainerMap}
	c.expandEnv()
	c.determineGraph()
	c.determineTarget("x", "all", "none")
	if c.target[0] != "linkTarget" || c.target[1] != "netTarget" || c.target[2] != "volumesFromTarget" || c.target[3] != "x" || len(c.target) != 4 {
		t.Errorf("all *Target containers should have been targeted but got %v", c.target)
	}
	c.determineTarget("x", "link", "none")
	if c.target[0] != "linkTarget" || c.target[1] != "x" || len(c.target) != 2 {
		t.Errorf("linkTarget should have been targeted but got %v", c.target)
	}
	c.determineTarget("x", "net", "none")
	if c.target[0] != "netTarget" || c.target[1] != "x" || len(c.target) != 2 {
		t.Errorf("netTarget should have been targeted but got %v", c.target)
	}
	c.determineTarget("x", "volumesFrom", "none")
	if c.target[0] != "volumesFromTarget" || c.target[1] != "x" || len(c.target) != 2 {
		t.Errorf("volumesFromTarget should have been targeted but got %v", c.target)
	}
	c.determineTarget("x", "none", "all")
	if c.target[0] != "linkSource" || c.target[1] != "netSource" || c.target[2] != "volumesFromSource" || c.target[3] != "x" || len(c.target) != 4 {
		t.Errorf("all *Source containers should have been targeted but got %v", c.target)
	}
	c.determineTarget("x", "none", "net")
	if c.target[0] != "netSource" || c.target[1] != "x" || len(c.target) != 2 {
		t.Errorf("netSource should have been targeted but got %v", c.target)
	}
	c.determineTarget("x", "none", "volumesFrom")
	if c.target[0] != "volumesFromSource" || c.target[1] != "x" || len(c.target) != 2 {
		t.Errorf("volumesFromSource should have been targeted but got %v", c.target)
	}
	c.determineTarget("x", "volumesFrom", "volumesFrom")
	if c.target[0] != "volumesFromSource" || c.target[1] != "volumesFromTarget" || c.target[2] != "x" || len(c.target) != 3 {
		t.Errorf("all volumesFrom* containers should have been targeted but got %v", c.target)
	}
}

func TestExplicitlyTargeted(t *testing.T) {
	var result []string
	var containers []string
	var rawGroups = make(map[string][]string)
	rawContainerMap := ContainerMap{
		"a": &container{},
		"b": &container{},
		"c": &container{},
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
		containerMap: ContainerMap{"a": &container{RawName: "a"}, "b": &container{RawName: "b"}},
		order:        []string{"a", "b"},
	}
	containers := c.Containers()
	if containers[0].Name() != "b" || containers[1].Name() != "a" {
		t.Errorf("Expected [b a], got %v", containers)
	}
}
