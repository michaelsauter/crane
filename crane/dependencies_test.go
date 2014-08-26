package crane

import (
	"reflect"
	"testing"
)

func TestIncludes(t *testing.T) {
	dependencies := Dependencies{
		All:         []string{"link", "volumesFrom", "net"},
		Link:        []string{"link"},
		VolumesFrom: []string{"volumesFrom"},
		Net:         "net",
	}

	if !dependencies.includes("link") || !dependencies.includes("volumesFrom") || !dependencies.includes("net") {
		t.Errorf("Dependencies should have included link, volumesFrom and net")
	}
	if dependencies.includes("non-existant") {
		t.Errorf("Dependencies should not have included 'non-existant'")
	}
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
		actual := dependencies.includesAsKind(example.needle, example.kind)
		if actual != example.expected {
			t.Errorf("includesAsKind should have returned %v, got %v for %v, %v", example.expected, actual, example.needle, example.kind)
		}
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
		kindDeps := dependencies.forKind(example.kind)
		if !reflect.DeepEqual(kindDeps, example.expected) {
			t.Errorf("%v dependencies expected for kind %v, got %v", example.expected, example.kind, kindDeps)
		}
	}

}

func TestSatisfied(t *testing.T) {
	var dependencies Dependencies

	dependencies = Dependencies{
		All: []string{"a"},
	}
	if dependencies.satisfied() {
		t.Errorf("Dependencies was not empty, but appeared to be satisfied")
	}

	dependencies = Dependencies{
		All: []string{},
	}
	if !dependencies.satisfied() {
		t.Errorf("Dependencies was empty, but appeared not to be satisfied")
	}
}
