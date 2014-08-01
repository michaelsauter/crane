package crane

import "testing"

func TestOrder(t *testing.T) {
	var err error
	var order []string
	var dependencyGraph DependencyGraph

	// Resolvable map
	dependencyGraph = DependencyGraph{
		"b": &Dependencies{list: []string{"c"}},
		"a": &Dependencies{list: []string{"b"}},
		"c": &Dependencies{list: []string{}},
	}
	order, err = dependencyGraph.order([]string{"a", "b", "c"}, false)
	if err != nil || len(order) != 3 || order[0] != "a" || order[1] != "b" || order[2] != "c" {
		t.Errorf("Order should have been [a b c], got %v. Err: %v", order, err)
	}

	// Cyclic map without forced order fails
	dependencyGraph = DependencyGraph{
		"b": &Dependencies{list: []string{"c"}},
		"a": &Dependencies{list: []string{"b"}},
		"c": &Dependencies{list: []string{"a"}},
	}
	order, err = dependencyGraph.order([]string{"a", "b", "c"}, false)
	if err == nil {
		t.Errorf("Cyclic dependency a -> b -> c -> a should not have been resolvable, got %v. Err: %v", order, err)
	}

	// Cyclic map with forced order fails
	dependencyGraph = DependencyGraph{
		"b": &Dependencies{list: []string{"c"}},
		"a": &Dependencies{list: []string{"b"}},
		"c": &Dependencies{list: []string{"a"}},
	}
	order, err = dependencyGraph.order([]string{"a", "b", "c"}, true)
	if err == nil {
		t.Errorf("Cyclic dependency a -> b -> c -> a should not have been resolvable, got %v. Err: %v", order, err)
	}

	// Resolvable map, partial target fails non-forced
	dependencyGraph = DependencyGraph{
		"b": &Dependencies{list: []string{"c"}},
		"a": &Dependencies{list: []string{"b"}},
		"c": &Dependencies{list: []string{}},
	}
	order, err = dependencyGraph.order([]string{"a", "b"}, false)
	if err == nil {
		t.Errorf("Dependency c was missing, so graph should not have been resolvable, got %v. Err: %v", order, err)
	}

	// Resolvable map, partial target, succeeds forced
	dependencyGraph = DependencyGraph{
		"b": &Dependencies{list: []string{"c"}},
		"a": &Dependencies{list: []string{"b"}},
		"c": &Dependencies{list: []string{}},
	}
	order, err = dependencyGraph.order([]string{"a", "b"}, true)
	if err != nil || len(order) != 2 || order[0] != "a" || order[1] != "b" {
		t.Errorf("Dependency c was missing, but order was forced, so graph should have been resolvable, got %v. Err: %v", order, err)
	}
}
