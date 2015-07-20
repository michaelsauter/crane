package crane

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDetermineTargetLinearChainDependencies(t *testing.T) {
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
				affected:     []string{"b", "a"},
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
