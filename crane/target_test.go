package crane

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewTarget(t *testing.T) {
	defer func() {
		allowed = []string{}
	}()
	allowed = []string{"a", "b", "c"}
	containerMap := NewStubbedContainerMap(true,
		&container{RawName: "a", RawRun: RunParameters{RawLink: []string{"b:b"}}},
		&container{RawName: "b", RawRun: RunParameters{RawLink: []string{"c:c"}}},
		&container{RawName: "c"},
	)
	cfg = &config{containerMap: containerMap}
	dependencyMap := cfg.DependencyMap()

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
		{
			target: "a+d",
			expected: Target{
				initial:      []string{"a"},
				dependencies: []string{"b", "c"},
				affected:     []string{},
			},
		},
		{
			target: "c+a",
			expected: Target{
				initial:      []string{"c"},
				dependencies: []string{},
				affected:     []string{"a", "b"},
			},
		},
	}

	for _, example := range examples {
		target, _ := NewTarget(dependencyMap, example.target)
		assert.Equal(t, example.expected, target)
	}
}

func TestNewTargetNonExisting(t *testing.T) {
	defer func() {
		allowed = []string{}
	}()
	allowed = []string{"a", "b"}
	containerMap := NewStubbedContainerMap(false,
		&container{RawName: "a", RawRun: RunParameters{RawLink: []string{"b:b"}}},
		&container{RawName: "b"},
	)

	cfg = &config{containerMap: containerMap}
	dependencyMap := cfg.DependencyMap()

	examples := []struct {
		target   string
		expected Target
	}{
		{
			target: "a+dependencies",
			expected: Target{
				initial:      []string{"a"},
				dependencies: []string{"b"},
				affected:     []string{},
			},
		},
		{
			target: "b+affected",
			expected: Target{
				initial:      []string{"b"},
				dependencies: []string{},
				affected:     []string{},
			},
		},
	}

	for _, example := range examples {
		target, _ := NewTarget(dependencyMap, example.target)
		assert.Equal(t, example.expected, target)
	}
}

func TestDeduplicationAll(t *testing.T) {
	defer func() {
		allowed = []string{}
	}()
	allowed = []string{"a", "b", "c"}
	containerMap := NewStubbedContainerMap(true,
		&container{RawName: "a", RawRun: RunParameters{RawLink: []string{"b:b"}}},
		&container{RawName: "b", RawRun: RunParameters{RawLink: []string{"c:c"}}},
		&container{RawName: "c"},
	)
	groups := map[string][]string{
		"ab": []string{"a", "b", "a"},
		"ac": []string{"a", "c"},
	}
	cfg = &config{containerMap: containerMap, groups: groups}
	dependencyMap := cfg.DependencyMap()

	target1, _ := NewTarget(dependencyMap, "ab+dependencies+affected")
	assert.Equal(t, []string{"a", "b", "c"}, target1.all())

	target2, _ := NewTarget(dependencyMap, "ac+dependencies+affected")
	assert.Equal(t, []string{"a", "b", "c"}, target2.all())
}
