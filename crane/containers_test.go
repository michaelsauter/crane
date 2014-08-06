package crane

import "testing"

func TestReversed(t *testing.T) {
	var containers Containers
	containers = []Container{&container{RawName: "a"}, &container{RawName: "b"}}
	reversed := containers.reversed()
	if reversed[0].Name() != "b" || reversed[1].Name() != "a" {
		t.Errorf("Containers should have been ordered [b a], got %v", reversed)
	}

}
