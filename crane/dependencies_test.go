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

	if !dependencies.includesAsKind("link", "all") {
		t.Errorf("false negative")
	}
	if !dependencies.includesAsKind("volumesFrom", "all") {
		t.Errorf("false negative")
	}
	if !dependencies.includesAsKind("net", "all") {
		t.Errorf("false negative")
	}

	if !dependencies.includesAsKind("link", "link") {
		t.Errorf("false negative")
	}
	if dependencies.includesAsKind("volumesFrom", "link") {
		t.Errorf("false positive")
	}
	if dependencies.includesAsKind("net", "link") {
		t.Errorf("false positive")
	}

	if dependencies.includesAsKind("link", "volumesFrom") {
		t.Errorf("false positive")
	}
	if !dependencies.includesAsKind("volumesFrom", "volumesFrom") {
		t.Errorf("false negative")
	}
	if dependencies.includesAsKind("net", "volumesFrom") {
		t.Errorf("false positive")
	}

	if dependencies.includesAsKind("link", "net") {
		t.Errorf("false positive")
	}
	if dependencies.includesAsKind("volumesFrom", "net") {
		t.Errorf("false positive")
	}
	if !dependencies.includesAsKind("net", "net") {
		t.Errorf("false negative")
	}
}

func TestForKind(t *testing.T) {
	dependencies := Dependencies{
		all:         []string{"link", "volumesFrom", "net"},
		link:        []string{"link"},
		volumesFrom: []string{"volumesFrom"},
		net:         "net",
	}

	if kindDeps := dependencies.forKind("all"); len(kindDeps) != 3 {
		t.Errorf("all dependencies expected but got %v", kindDeps)
	}
	if kindDeps := dependencies.forKind("link"); len(kindDeps) != 1 || kindDeps[0] != "link" {
		t.Errorf("link expected but got %v", kindDeps)
	}
	if kindDeps := dependencies.forKind("volumesFrom"); len(kindDeps) != 1 || kindDeps[0] != "volumesFrom" {
		t.Errorf("volumesFrom expected but got %v", kindDeps)
	}
	if kindDeps := dependencies.forKind("net"); len(kindDeps) != 1 || kindDeps[0] != "net" {
		t.Errorf("net expected but got %v", kindDeps)
	}
	if kindDeps := dependencies.forKind("foobar"); len(kindDeps) != 0 {
		t.Errorf("no dependencies expected but got %v", kindDeps)
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
