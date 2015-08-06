package crane

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewTarget(t *testing.T) {
	containerMap := NewStubbedContainerMap(true,
		&container{RawName: "a", RunParams: RunParameters{RawLink: []string{"b:b"}}},
		&container{RawName: "b", RunParams: RunParameters{RawLink: []string{"c:c"}}},
		&container{RawName: "c"},
	)
	cfg = &config{containerMap: containerMap}
	dependencyGraph := cfg.DependencyGraph([]string{})

	examples := []struct {
		target   string
		expected Target
	}{
		{
			target: "a+dependencies",
			expected: Target{
				initial:      []string{"a"},
				dependencies: []string{"b", "c"},
				affected:     []string{},
			},
		},
		{
			target: "b+dependencies",
			expected: Target{
				initial:      []string{"b"},
				dependencies: []string{"c"},
				affected:     []string{},
			},
		},
		{
			target: "c+affected",
			expected: Target{
				initial:      []string{"c"},
				dependencies: []string{},
				affected:     []string{"a", "b"},
			},
		},
		{
			target: "b+affected",
			expected: Target{
				initial:      []string{"b"},
				dependencies: []string{},
				affected:     []string{"a"},
			},
		},
		{
			target: "b+dependencies+affected",
			expected: Target{
				initial:      []string{"b"},
				dependencies: []string{"c"},
				affected:     []string{"a"},
			},
		},
	}

	for _, example := range examples {
		target := NewTarget(dependencyGraph, example.target)
		assert.Equal(t, example.expected, target)
	}
}

func TestDeduplicationAll(t *testing.T) {
	containerMap := NewStubbedContainerMap(true,
		&container{RawName: "a", RunParams: RunParameters{RawLink: []string{"b:b"}}},
		&container{RawName: "b", RunParams: RunParameters{RawLink: []string{"c:c"}}},
		&container{RawName: "c"},
	)
	groups := map[string][]string{
		"ab": []string{"a", "b", "a"},
	}
	cfg = &config{containerMap: containerMap, groups: groups}
	dependencyGraph := cfg.DependencyGraph([]string{})

	target := NewTarget(dependencyGraph, "ab+dependencies+affected")
	assert.Equal(t, []string{"a", "b", "c"}, target.all())
}
