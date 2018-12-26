package crane

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewTarget(t *testing.T) {
	defer func() {
		allowed = []string{}
	}()
	allowed = []string{"a", "b", "c"}
	containerMap := NewStubbedContainerMap(true,
		&container{RawName: "a", RawNet: "bridge", RawLink: []string{"b:b"}},
		&container{RawName: "b", RawNet: "bridge", RawLink: []string{"c:c"}},
		&container{RawName: "c", RawNet: "bridge"},
	)
	cfg = &config{containerMap: containerMap}
	dependencyMap := cfg.DependencyMap()

	examples := []struct {
		target   string
		extend   bool
		expected Target
	}{
		{
			target: "a",
			extend: true,
			expected: Target{
				initial:      []string{"a"},
				dependencies: []string{"b", "c"},
			},
		},
		{
			target: "b",
			extend: true,
			expected: Target{
				initial:      []string{"b"},
				dependencies: []string{"c"},
			},
		},
		{
			target: "c",
			extend: false,
			expected: Target{
				initial:      []string{"c"},
				dependencies: []string{},
			},
		},
		{
			target: "b",
			extend: false,
			expected: Target{
				initial:      []string{"b"},
				dependencies: []string{},
			},
		},
		{
			target: "b",
			extend: true,
			expected: Target{
				initial:      []string{"b"},
				dependencies: []string{"c"},
			},
		},
	}

	for _, example := range examples {
		target, _ := NewTarget(dependencyMap, example.target, example.extend)
		assert.Equal(t, example.expected, target)
	}
}

func TestNewTargetNonExisting(t *testing.T) {
	defer func() {
		allowed = []string{}
	}()
	allowed = []string{"a", "b"}
	containerMap := NewStubbedContainerMap(false,
		&container{RawName: "a", RawNet: "bridge", RawLink: []string{"b:b"}},
		&container{RawName: "b", RawNet: "bridge"},
	)

	cfg = &config{containerMap: containerMap}
	dependencyMap := cfg.DependencyMap()

	examples := []struct {
		target   string
		extend   bool
		expected Target
	}{
		{
			target: "a",
			extend: true,
			expected: Target{
				initial:      []string{"a"},
				dependencies: []string{"b"},
			},
		},
		{
			target: "b",
			extend: false,
			expected: Target{
				initial:      []string{"b"},
				dependencies: []string{},
			},
		},
	}

	for _, example := range examples {
		target, _ := NewTarget(dependencyMap, example.target, example.extend)
		assert.Equal(t, example.expected, target)
	}
}

func TestDeduplicationAll(t *testing.T) {
	defer func() {
		allowed = []string{}
	}()
	allowed = []string{"a", "b", "c"}
	containerMap := NewStubbedContainerMap(true,
		&container{RawName: "a", RawNet: "bridge", RawLink: []string{"b:b"}},
		&container{RawName: "b", RawNet: "bridge", RawLink: []string{"c:c"}},
		&container{RawName: "c", RawNet: "bridge"},
	)
	groups := map[string][]string{
		"ab": []string{"a", "b", "a"},
	}
	cfg = &config{containerMap: containerMap, groups: groups}
	dependencyMap := cfg.DependencyMap()

	target, _ := NewTarget(dependencyMap, "ab", true)
	assert.Equal(t, []string{"a", "b", "c"}, target.all())
}
