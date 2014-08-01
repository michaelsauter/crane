package crane

import "testing"

func TestSatisfied(t *testing.T) {
	var dependencies Dependencies

	dependencies = Dependencies{
		list: []string{"a"},
	}
	if dependencies.satisfied() {
		t.Errorf("Dependencies was not empty, but appeared to be satisfied")
	}

	dependencies = Dependencies{
		list: []string{},
	}
	if !dependencies.satisfied() {
		t.Errorf("Dependencies was empty, but appeared not to be satisfied")
	}
}
