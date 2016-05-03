package crane

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewTarget(t *testing.T) {
	containerMap := NewStubbedContainerMap(true,
		&Container{RawName: "a", RawRun: RunParameters{RawLink: []string{"b:b"}}},
		&Container{RawName: "b", RawRun: RunParameters{RawLink: []string{"c:c"}}},
		&Container{RawName: "c"},
	)
	cfg = &Config{containerMap: containerMap}
	dependencyMap := cfg.DependencyMap([]string{})

	examples := []struct {
		target   string
		expected Target
	}{
		{
			target: "a+dependencies",
			expected: Target{
				Initial:      []string{"a"},
				Dependencies: []string{"b", "c"},
				Affected:     []string{},
			},
		},
		{
			target: "b+dependencies",
			expected: Target{
				Initial:      []string{"b"},
				Dependencies: []string{"c"},
				Affected:     []string{},
			},
		},
		{
			target: "c+affected",
			expected: Target{
				Initial:      []string{"c"},
				Dependencies: []string{},
				Affected:     []string{"a", "b"},
			},
		},
		{
			target: "b+affected",
			expected: Target{
				Initial:      []string{"b"},
				Dependencies: []string{},
				Affected:     []string{"a"},
			},
		},
		{
			target: "b+dependencies+affected",
			expected: Target{
				Initial:      []string{"b"},
				Dependencies: []string{"c"},
				Affected:     []string{"a"},
			},
		},
		{
			target: "a+d",
			expected: Target{
				Initial:      []string{"a"},
				Dependencies: []string{"b", "c"},
				Affected:     []string{},
			},
		},
		{
			target: "c+a",
			expected: Target{
				Initial:      []string{"c"},
				Dependencies: []string{},
				Affected:     []string{"a", "b"},
			},
		},
	}

	for _, example := range examples {
		target, _ := NewTarget(dependencyMap, example.target, []string{})
		assert.Equal(t, example.expected, target)
	}
}

func TestNewTargetNonExisting(t *testing.T) {
	containerMap := NewStubbedContainerMap(false,
		&Container{RawName: "a", RawRun: RunParameters{RawLink: []string{"b:b"}}},
		&Container{RawName: "b"},
	)

	cfg = &Config{containerMap: containerMap}
	dependencyMap := cfg.DependencyMap([]string{})

	examples := []struct {
		target   string
		expected Target
	}{
		{
			target: "a+dependencies",
			expected: Target{
				Initial:      []string{"a"},
				Dependencies: []string{"b"},
				Affected:     []string{},
			},
		},
		{
			target: "b+affected",
			expected: Target{
				Initial:      []string{"b"},
				Dependencies: []string{},
				Affected:     []string{},
			},
		},
	}

	for _, example := range examples {
		target, _ := NewTarget(dependencyMap, example.target, []string{})
		assert.Equal(t, example.expected, target)
	}
}

func TestDeduplicationAll(t *testing.T) {
	containerMap := NewStubbedContainerMap(true,
		&Container{RawName: "a", RawRun: RunParameters{RawLink: []string{"b:b"}}},
		&Container{RawName: "b", RawRun: RunParameters{RawLink: []string{"c:c"}}},
		&Container{RawName: "c"},
	)
	groups := map[string][]string{
		"ab": []string{"a", "b", "a"},
	}
	cfg = &Config{containerMap: containerMap, groups: groups}
	dependencyMap := cfg.DependencyMap([]string{})

	target, _ := NewTarget(dependencyMap, "ab+dependencies+affected", []string{})
	assert.Equal(t, []string{"a", "b", "c"}, target.all())
}
