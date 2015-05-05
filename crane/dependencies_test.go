package crane

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestIncludes(t *testing.T) {
	dependencies := Dependencies{
		All:         []string{"link", "volumesFrom", "net"},
		Link:        []string{"link"},
		VolumesFrom: []string{"volumesFrom"},
		Net:         "net",
	}
	assert.True(t, dependencies.includes("link"))
	assert.True(t, dependencies.includes("volumesFrom"))
	assert.True(t, dependencies.includes("net"))
	assert.False(t, dependencies.includes("non-existent"))
}

func TestIncludesAsKind(t *testing.T) {
	dependencies := Dependencies{
		All:         []string{"link", "volumesFrom", "net"},
		Link:        []string{"link"},
		VolumesFrom: []string{"volumesFrom"},
		Net:         "net",
	}

	examples := []struct {
		needle   string
		kind     string
		expected bool
	}{
		{ // kind all
			needle:   "link",
			kind:     "all",
			expected: true,
		},
		{
			needle:   "volumesFrom",
			kind:     "all",
			expected: true,
		},
		{
			needle:   "net",
			kind:     "all",
			expected: true,
		},
		{ // kind link
			needle:   "link",
			kind:     "link",
			expected: true,
		},
		{
			needle:   "volumesFrom",
			kind:     "link",
			expected: false,
		},
		{
			needle:   "net",
			kind:     "link",
			expected: false,
		},
		{ // kind volumesFrom
			needle:   "link",
			kind:     "volumesFrom",
			expected: false,
		},
		{
			needle:   "volumesFrom",
			kind:     "volumesFrom",
			expected: true,
		},
		{
			needle:   "net",
			kind:     "volumesFrom",
			expected: false,
		},
		{ // kind net
			needle:   "link",
			kind:     "net",
			expected: false,
		},
		{
			needle:   "volumesFrom",
			kind:     "net",
			expected: false,
		},
		{
			needle:   "net",
			kind:     "net",
			expected: true,
		},
	}

	for _, example := range examples {
		assert.Equal(t, example.expected, dependencies.includesAsKind(example.needle, example.kind))
	}

}

func TestForKind(t *testing.T) {
	dependencies := Dependencies{
		All:         []string{"link", "volumesFrom", "net"},
		Link:        []string{"link"},
		VolumesFrom: []string{"volumesFrom"},
		Net:         "net",
	}

	examples := []struct {
		kind     string
		expected []string
	}{
		{
			kind:     "all",
			expected: []string{"link", "volumesFrom", "net"},
		},
		{
			kind:     "link",
			expected: []string{"link"},
		},
		{
			kind:     "volumesFrom",
			expected: []string{"volumesFrom"},
		},
		{
			kind:     "net",
			expected: []string{"net"},
		},
		{
			kind:     "foobar",
			expected: []string{},
		},
	}

	for _, example := range examples {
		assert.Equal(t, example.expected, dependencies.forKind(example.kind))
	}

}

func TestSatisfied(t *testing.T) {
	var dependencies Dependencies

	dependencies = Dependencies{
		All: []string{"a"},
	}
	assert.False(t, dependencies.satisfied(), "Dependencies was not empty, but appeared to be satisfied")

	dependencies = Dependencies{
		All: []string{},
	}
	assert.True(t, dependencies.satisfied(), "Dependencies was empty, but appeared not to be satisfied")
}
