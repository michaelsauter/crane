package crane

import (
	"github.com/stretchr/testify/assert"
	"sort"
	"testing"
)

func TestAllowedContainers(t *testing.T) {
	rawContainerMap := map[string]Container{
		"a": &container{},
		"b": &container{},
		"c": &container{},
	}
	groups := map[string][]string{
		"group": []string{"a", "b"},
	}
	cfg = &config{containerMap: rawContainerMap, groups: groups}

	examples := []struct {
		excluded []string
		only     string
		expected []string
	}{
		{
			excluded: []string{},
			only:     "",
			expected: []string{"a", "b", "c"},
		},
		{
			excluded: []string{"a"},
			only:     "",
			expected: []string{"b", "c"},
		},
		{
			excluded: []string{},
			only:     "a",
			expected: []string{"a"},
		},
		{
			excluded: []string{},
			only:     "group",
			expected: []string{"a", "b"},
		},
		{
			excluded: []string{"b"},
			only:     "group",
			expected: []string{"a"},
		},
	}

	for _, example := range examples {
		containers := allowedContainers(example.excluded, example.only)
		sort.Strings(containers)
		assert.Equal(t, example.expected, containers)
	}

	// with a default group
	rawContainerMap = map[string]Container{
		"a": &container{},
		"b": &container{},
		"c": &container{},
	}
	groups = map[string][]string{
		"default": []string{"a", "b"},
	}
	cfg = &config{containerMap: rawContainerMap, groups: groups}
	containers := allowedContainers([]string{}, "")
	sort.Strings(containers)
	assert.Equal(t, []string{"a", "b", "c"}, containers)
}
