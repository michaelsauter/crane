package crane

import "testing"

func TestSubset(t *testing.T) {
	containerMap := ContainerMap{
		"a": Container{},
		"b": Container{},
		"c": Container{},
	}

	subset := containerMap.subset([]string{"a", "c"})
	if _, present := subset["a"]; !present {
		t.Errorf("a should have been kept")
	}
	if _, present := subset["b"]; present {
		t.Errorf("b should have been removed")
	}
	if _, present := subset["c"]; !present {
		t.Errorf("c should have been kept")
	}

	subset = containerMap.subset([]string{"a", "a"})
	if _, present := subset["a"]; !present {
		t.Errorf("a should have been kept")
	}
	if _, present := subset["b"]; present {
		t.Errorf("b should have been removed")
	}
	if _, present := subset["c"]; present {
		t.Errorf("c should have been removed")
	}

	subset = containerMap.subset([]string{"d"})
	if len(subset) != 0 {
		t.Errorf("everything should have been removed but got %v", subset)
	}
}

func TestOrder(t *testing.T) {
	var err error
	var order []string
	var containerMap ContainerMap

	// Resolvable map
	containerMap = ContainerMap{
		"b": Container{RawName: "b", Run: RunParameters{RawLink: []string{"c:c"}}},
		"a": Container{RawName: "a", Run: RunParameters{RawLink: []string{"b:b"}}},
		"c": Container{RawName: "c"},
	}
	// Default order
	order, err = containerMap.order(false, false)
	if err != nil || order[0] != "a" || order[1] != "b" || order[2] != "c" {
		t.Errorf("Order should have been [a b c], got %v. Err: %v", order, err)
	}
	// Reversed order
	order, err = containerMap.order(true, false)
	if err != nil || order[0] != "c" || order[1] != "b" || order[2] != "a" {
		t.Errorf("Order should have been [c b a], got %v. Err: %v", order, err)
	}

	// Unresolvable map
	containerMap = ContainerMap{
		"b": Container{RawName: "b", Run: RunParameters{RawLink: []string{"c:c"}}},
		"a": Container{RawName: "a", Run: RunParameters{RawLink: []string{"b:b"}}},
		"c": Container{RawName: "c", Run: RunParameters{RawLink: []string{"a:a"}}},
	}
	// Default order
	order, err = containerMap.order(false, false)
	if err == nil {
		t.Errorf("Cyclic dependency a -> b -> c -> a should not have been resolvable, got %v. Err: %v", order, err)
	}
	// Default order, ignoring unresolved
	order, err = containerMap.order(false, true)
	if err != nil || order[0] != "a" || order[1] != "b" || order[2] != "c" {
		t.Errorf("Order should have been [a b c], got %v. Err: %v", order, err)
	}
	// Reversed order
	order, err = containerMap.order(true, true)
	if err != nil || order[0] != "c" || order[1] != "b" || order[2] != "a" {
		t.Errorf("Order should have been [c b a], got %v", order)
	}
	// Reversed order, ignoring unresolved
	order, err = containerMap.order(true, true)
	if err != nil || order[0] != "c" || order[1] != "b" || order[2] != "a" {
		t.Errorf("Order should have been [c b a], got %v. Err: %v", order, err)
	}
}

func TestAlphabetical(t *testing.T) {
	var containerMap ContainerMap

	containerMap = ContainerMap{
		"b": Container{},
		"c": Container{},
		"a": Container{},
		"e": Container{},
		"d": Container{},
	}

	alphabetical := containerMap.alphabetical(false)
	if alphabetical[0] != "a" || alphabetical[1] != "b" || alphabetical[2] != "c" || alphabetical[3] != "d" || alphabetical[4] != "e" {
		t.Errorf("Order should have been [a b c d e], got %v", alphabetical)
	}
	alphabetical = containerMap.alphabetical(true)
	if alphabetical[0] != "e" || alphabetical[1] != "d" || alphabetical[2] != "c" || alphabetical[3] != "b" || alphabetical[4] != "a" {
		t.Errorf("Order should have been [e d c b a], got %v", alphabetical)
	}
}
