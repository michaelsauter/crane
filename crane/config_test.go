package crane

import (
	"reflect"
	"sort"
	"testing"
)

// Create a map of stubbed containers out of the provided set
func NewStubbedContainerMap(exists bool, containers ...Container) ContainerMap {
	containerMap := make(map[string]Container)
	for _, container := range containers {
		containerMap[container.Name()] = &StubbedContainer{container, exists}
	}
	return containerMap
}

type StubbedContainer struct {
	Container
	exists bool
}

func (stubbedContainer *StubbedContainer) Exists() bool {
	return stubbedContainer.exists
}

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

func TestUnmarshal(t *testing.T) {
	var actual *config
	json := []byte(
		`{
    "containers": {
        "apache": {
            "dockerfile": "apache",
            "image": "michaelsauter/apache",
            "run": {
                "volumes-from": ["crane_app"],
                "publish": ["80:80"],
                "link": ["crane_mysql:db", "crane_memcached:cache"],
                "detach": true
            }
        }
    },
    "groups": {
        "default": [
            "apache"
        ]
    }
}
`)
	actual = unmarshal(json, ".json")
	if _, ok := actual.RawContainerMap["apache"]; !ok {
		t.Errorf("Config should have one container, got %v", actual.RawContainerMap)
	}
	if len(actual.RawContainerMap["apache"].RunParams.Link()) != 2 {
		t.Errorf("Container should have been linked to 2 other containers, got %v", actual.RawContainerMap["apache"].RunParams.Link())
	}
	if group, ok := actual.RawGroups["default"]; !ok || len(group) != 1 {
		t.Errorf("Config should have one `default` group with one container, got %v", actual.RawGroups)
	}
	actual = nil

	yaml := []byte(
		`containers:
  apache:
    dockerfile: apache
    image: michaelsauter/apache
    run:
      volumes-from: ["crane_app"]
      publish: ["80:80"]
      link: ["crane_mysql:db", "crane_memcached:cache"]
      detach: true
groups:
  default:
    - apache
`)
	actual = unmarshal(yaml, ".yml")
	if _, ok := actual.RawContainerMap["apache"]; !ok {
		t.Errorf("Config should have one container, got %v", actual.RawContainerMap)
	}
	if len(actual.RawContainerMap["apache"].RunParams.Link()) != 2 {
		t.Errorf("Container should have been linked to 2 other containers, got %v", actual.RawContainerMap["apache"].RunParams.Link())
	}
	if group, ok := actual.RawGroups["default"]; !ok || len(group) != 1 {
		t.Errorf("Config should have one `default` group with one container, got %v", actual.RawGroups)
	}
}

func TestExpandEnv(t *testing.T) {
	rawContainerMap := map[string]*container{
		"a": &container{},
		"b": &container{},
	}
	c := &config{RawContainerMap: rawContainerMap}
	c.expandEnv()
	if c.containerMap["a"].Name() != "a" || c.containerMap["b"].Name() != "b" {
		t.Errorf("Names should be 'a' and 'b', got %s and %s", c.containerMap["a"].Name(), c.containerMap["b"].Name())
	}
}

func TestGraph(t *testing.T) {
	containerMap := NewStubbedContainerMap(true,
		&container{RawName: "a", RunParams: RunParameters{RawLink: []string{"b:b"}}},
		&container{RawName: "b", RunParams: RunParameters{RawLink: []string{"c:c"}}},
		&container{RawName: "c"},
	)
	c := &config{containerMap: containerMap}
	dependencyGraph := c.DependencyGraph()
	if len(dependencyGraph) != 3 {
		t.Errorf("Expecting the graph to contain all containers (defined in %v), got %v", containerMap, dependencyGraph)
	}
	// make sure a new graph is returned each time
	dependencyGraph.resolve("a") // mutate the previous graph
	dependencyGraph = c.DependencyGraph()
	if len(dependencyGraph) != 3 {
		t.Errorf("Expecting the graph to contain all containers (defined in %v), got %v", containerMap, dependencyGraph)
	}
}

func TestDetermineTargetLinearChainDependencies(t *testing.T) {
	containerMap := NewStubbedContainerMap(true,
		&container{RawName: "a", RunParams: RunParameters{RawLink: []string{"b:b"}}},
		&container{RawName: "b", RunParams: RunParameters{RawLink: []string{"c:c"}}},
		&container{RawName: "c"},
	)
	c := &config{containerMap: containerMap}
	c.dependencyGraph = c.DependencyGraph()

	examples := []struct {
		target              []string
		cascadeDependencies string
		cascadeAffected     string
		expected            Target
	}{
		{
			target:              []string{"a"},
			cascadeDependencies: "all",
			cascadeAffected:     "none",
			expected:            []string{"a", "b", "c"},
		},
		{
			target:              []string{"b"},
			cascadeDependencies: "all",
			cascadeAffected:     "none",
			expected:            []string{"b", "c"},
		},
		{
			target:              []string{"c"},
			cascadeDependencies: "none",
			cascadeAffected:     "all",
			expected:            []string{"a", "b", "c"},
		},
		{
			target:              []string{"b"},
			cascadeDependencies: "none",
			cascadeAffected:     "all",
			expected:            []string{"a", "b"},
		},
		{
			target:              []string{"b"},
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
	containerMap := NewStubbedContainerMap(true,
		&container{RawName: "a", RunParams: RunParameters{RawLink: []string{"b:b", "c:c"}}},
		&container{RawName: "b", RunParams: RunParameters{RawLink: []string{"d:d"}}},
		&container{RawName: "c", RunParams: RunParameters{RawLink: []string{"e:e"}}},
		&container{RawName: "d"},
		&container{RawName: "e"},
	)
	c := &config{containerMap: containerMap}
	c.dependencyGraph = c.DependencyGraph()
	c.determineTarget([]string{"a"}, "all", "none")
	if len(c.target) != 5 {
		t.Errorf("all containers should have been targeted but got %v", c.target)
	}
	c.determineTarget([]string{"b"}, "all", "none")
	if c.target[0] != "b" || c.target[1] != "d" || len(c.target) != 2 {
		t.Errorf("all b and d should have been targeted but got %v", c.target)
	}
	c.determineTarget([]string{"b", "c"}, "all", "none")
	if c.target[0] != "b" || c.target[1] != "c" || c.target[2] != "d" || c.target[3] != "e" || len(c.target) != 4 {
		t.Errorf("a should have been left out but got %v", c.target)
	}
	c.determineTarget([]string{"b", "c"}, "none", "all")
	if c.target[0] != "a" || c.target[1] != "b" || c.target[2] != "c" || len(c.target) != 3 {
		t.Errorf("d and e should have been left out but got %v", c.target)
	}
	c.determineTarget([]string{"b", "c"}, "all", "all")
	if len(c.target) != 5 {
		t.Errorf("all containers should have been targeted but got %v", c.target)
	}
}

func TestDetermineTargetMissingDependencies(t *testing.T) {
	containerMap := NewStubbedContainerMap(true,
		&container{RawName: "a", RunParams: RunParameters{RawLink: []string{"b:b", "d:d"}}},
		&container{RawName: "b", RunParams: RunParameters{RawLink: []string{"c:c"}}},
		&container{RawName: "c", RunParams: RunParameters{RawLink: []string{"d:d"}}},
	)
	c := &config{containerMap: containerMap}
	c.dependencyGraph = c.DependencyGraph()
	c.determineTarget([]string{"a"}, "all", "none")
	if len(c.target) != 3 {
		t.Errorf("only declared containers should have been targeted but got %v", c.target)
	}
	c.determineTarget([]string{"c"}, "none", "all")
	if len(c.target) != 3 {
		t.Errorf("only declared containers should have been targeted but got %v", c.target)
	}
	c.determineTarget([]string{"a"}, "all", "all")
	if len(c.target) != 3 {
		t.Errorf("only declared containers should have been targeted but got %v", c.target)
	}
}

func TestDetermineTargetCustomCascading(t *testing.T) {
	containerMap := NewStubbedContainerMap(true,
		&container{RawName: "linkSource", RunParams: RunParameters{RawLink: []string{"x:x"}}},
		&container{RawName: "netSource", RunParams: RunParameters{RawNet: "container:x"}},
		&container{RawName: "volumesFromSource", RunParams: RunParameters{RawVolumesFrom: []string{"x"}}},
		&container{RawName: "x", RunParams: RunParameters{RawLink: []string{"linkTarget:linkTarget"}, RawNet: "container:netTarget", RawVolumesFrom: []string{"volumesFromTarget"}}},
		&container{RawName: "linkTarget"},
		&container{RawName: "netTarget"},
		&container{RawName: "volumesFromTarget"},
	)
	c := &config{containerMap: containerMap}
	c.dependencyGraph = c.DependencyGraph()
	c.determineTarget([]string{"x"}, "all", "none")
	if c.target[0] != "linkTarget" || c.target[1] != "netTarget" || c.target[2] != "volumesFromTarget" || c.target[3] != "x" || len(c.target) != 4 {
		t.Errorf("all *Target containers should have been targeted but got %v", c.target)
	}
	c.determineTarget([]string{"x"}, "link", "none")
	if c.target[0] != "linkTarget" || c.target[1] != "x" || len(c.target) != 2 {
		t.Errorf("linkTarget should have been targeted but got %v", c.target)
	}
	c.determineTarget([]string{"x"}, "net", "none")
	if c.target[0] != "netTarget" || c.target[1] != "x" || len(c.target) != 2 {
		t.Errorf("netTarget should have been targeted but got %v", c.target)
	}
	c.determineTarget([]string{"x"}, "volumesFrom", "none")
	if c.target[0] != "volumesFromTarget" || c.target[1] != "x" || len(c.target) != 2 {
		t.Errorf("volumesFromTarget should have been targeted but got %v", c.target)
	}
	c.determineTarget([]string{"x"}, "none", "all")
	if c.target[0] != "linkSource" || c.target[1] != "netSource" || c.target[2] != "volumesFromSource" || c.target[3] != "x" || len(c.target) != 4 {
		t.Errorf("all *Source containers should have been targeted but got %v", c.target)
	}
	c.determineTarget([]string{"x"}, "none", "net")
	if c.target[0] != "netSource" || c.target[1] != "x" || len(c.target) != 2 {
		t.Errorf("netSource should have been targeted but got %v", c.target)
	}
	c.determineTarget([]string{"x"}, "none", "volumesFrom")
	if c.target[0] != "volumesFromSource" || c.target[1] != "x" || len(c.target) != 2 {
		t.Errorf("volumesFromSource should have been targeted but got %v", c.target)
	}
	c.determineTarget([]string{"x"}, "volumesFrom", "volumesFrom")
	if c.target[0] != "volumesFromSource" || c.target[1] != "volumesFromTarget" || c.target[2] != "x" || len(c.target) != 3 {
		t.Errorf("all volumesFrom* containers should have been targeted but got %v", c.target)
	}
}

func TestDetermineTargetCascadingToExisting(t *testing.T) {
	containerMap := NewStubbedContainerMap(true,
		&container{RawName: "existingSource", RunParams: RunParameters{RawLink: []string{"x:x"}}},
		&container{RawName: "nonExistingSource", RunParams: RunParameters{RawLink: []string{"x:x"}}},
		&container{RawName: "x", RunParams: RunParameters{RawLink: []string{"existingTarget:existingTarget", "nonExistingTarget:nonExistingTarget"}}},
		&container{RawName: "existingTarget"},
		&container{RawName: "nonExistingTarget"},
	)
	containerMap["nonExistingSource"].(*StubbedContainer).exists = false
	containerMap["nonExistingTarget"].(*StubbedContainer).exists = false
	c := &config{containerMap: containerMap}
	c.dependencyGraph = c.DependencyGraph()
	c.determineTarget([]string{"x"}, "all", "none")
	if c.target[0] != "existingTarget" || c.target[1] != "nonExistingTarget" || c.target[2] != "x" || len(c.target) != 3 {
		t.Errorf("all *Target containers should have been targeted but got %v", c.target)
	}
	c.determineTarget([]string{"x"}, "none", "all")
	if c.target[0] != "existingSource" || c.target[1] != "x" || len(c.target) != 2 {
		t.Errorf("from the *Source containers, only existingSource should have been targeted but got %v", c.target)
	}
}

func TestExplicitlyTargeted(t *testing.T) {
	var expected []string
	var containers []string
	containerMap := NewStubbedContainerMap(true,
		&container{RawName: "a"},
		&container{RawName: "b"},
		&container{RawName: "c"},
	)

	// No target given
	// If default group exist, it returns its containers
	expected = []string{"a", "b"}
	groups := map[string][]string{
		"default": expected,
	}
	c := &config{containerMap: containerMap, groups: groups}
	containers = c.explicitlyTargeted([]string{})
	if len(containers) != 2 || containers[0] != "a" || containers[1] != "b" {
		t.Errorf("Expected %v, got %v", expected, containers)
	}
	// If no default group, returns all containers
	expected = []string{"a", "b", "c"}
	c = &config{containerMap: containerMap}
	containers = c.explicitlyTargeted([]string{})
	sort.Strings(containers)
	if len(containers) != 3 || containers[0] != "a" || containers[1] != "b" || containers[2] != "c" {
		t.Errorf("Expected %v, got %v", expected, containers)
	}
	// Target given
	// Target is a group
	expected = []string{"b", "c"}
	groups = map[string][]string{
		"second": expected,
	}
	c = &config{containerMap: containerMap, groups: groups}
	containers = c.explicitlyTargeted([]string{"second"})
	if len(containers) != 2 || containers[0] != "b" || containers[1] != "c" {
		t.Errorf("Expected %v, got %v", expected, containers)
	}
	// Target is a container
	expected = []string{"a"}
	containers = c.explicitlyTargeted([]string{"a"})
	if len(containers) != 1 || containers[0] != "a" {
		t.Errorf("Expected %v, got %v", expected, containers)
	}
	// Target is 2 containers
	expected = []string{"a", "b"}
	containers = c.explicitlyTargeted([]string{"a", "b"})
	if len(containers) != 2 || containers[0] != "a" || containers[1] != "b" {
		t.Errorf("Expected %v, got %v", expected, containers)
	}
	// Target is a container and a group
	expected = []string{"a", "b", "c"}
	containers = c.explicitlyTargeted([]string{"a", "second"})
	if len(containers) != 3 || containers[0] != "a" || containers[1] != "b" || containers[2] != "c" {
		t.Errorf("Expected %v, got %v", expected, containers)
	}
}

func TestTargetedContainers(t *testing.T) {
	c := &config{
		containerMap: NewStubbedContainerMap(true, &container{RawName: "a"}, &container{RawName: "b"}),
		order:        []string{"a", "b"},
	}
	containers := c.TargetedContainers()
	if containers[0].Name() != "b" || containers[1].Name() != "a" {
		t.Errorf("Expected [b a], got %v", containers)
	}
}
