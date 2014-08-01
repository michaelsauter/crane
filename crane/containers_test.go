package crane

import "testing"

func TestReversed(t *testing.T) {
	var containers Containers
	containers = []Container{Container{RawName: "a"}, Container{RawName: "b"}}
	reversed := containers.reversed()
	if reversed[0].RawName != "b" || reversed[1].RawName != "a" {
		t.Errorf("Containers should have been ordered [b a], got %v", reversed)
	}

}
