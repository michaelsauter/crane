package crane

import "testing"

func TestIncludes(t *testing.T) {
	dependencies := Dependencies{
		all:         []string{"link", "volumesFrom", "net"},
		link:        []string{"link"},
		volumesFrom: []string{"volumesFrom"},
		net:         "net",
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
		all:         []string{"link", "volumesFrom", "net"},
		link:        []string{"link"},
		volumesFrom: []string{"volumesFrom"},
		net:         "net",
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
		all:         []string{"link", "volumesFrom", "net"},
		link:        []string{"link"},
		volumesFrom: []string{"volumesFrom"},
		net:         "net",
	}

	examples := []struct {
		kind           string
		expectedLength int
		expectedFirst  string
	}{
		{
			kind:           "all",
			expectedLength: 3,
			expectedFirst:  "",
		},
		{
			kind:           "link",
			expectedLength: 1,
			expectedFirst:  "link",
		},
		{
			kind:           "volumesFrom",
			expectedLength: 1,
			expectedFirst:  "volumesFrom",
		},
		{
			kind:           "net",
			expectedLength: 1,
			expectedFirst:  "net",
		},
		{
			kind:           "foobar",
			expectedLength: 0,
			expectedFirst:  "",
		},
	}

	for _, example := range examples {
		kindDeps := dependencies.forKind(example.kind)
		if len(kindDeps) != example.expectedLength {
			t.Errorf("%v dependencies expected for kind %v, got %v", example.expectedLength, example.kind, len(kindDeps))
		}
		if len(example.expectedFirst) > 0 && kindDeps[0] != example.expectedFirst {
			t.Errorf("Expected %v at index 0 for kind %v, got %v", example.expectedFirst, example.kind, kindDeps[0])
		}
	}

}

func TestSatisfied(t *testing.T) {
	var dependencies Dependencies

	dependencies = Dependencies{
		all: []string{"a"},
	}
	if dependencies.satisfied() {
		t.Errorf("Dependencies was not empty, but appeared to be satisfied")
	}

	dependencies = Dependencies{
		all: []string{},
	}
	if !dependencies.satisfied() {
		t.Errorf("Dependencies was empty, but appeared not to be satisfied")
	}
}
