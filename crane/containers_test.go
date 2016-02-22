package crane

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReversed(t *testing.T) {
	var containers Containers
	containers = []Container{&container{RawName: "a"}, &container{RawName: "b"}}
	reversed := containers.Reversed()
	assert.Len(t, reversed, 2)
	assert.Equal(t, "b", reversed[0].Name())
	assert.Equal(t, "a", reversed[1].Name())
}

func TestProvisioningDuplicates(t *testing.T) {
	var containers Containers
	containers = []Container{
		&container{RawName: "A", RawBuild: BuildParameters{RawContext: "dockerfile1"}, RawImage: "image1"},
		&container{RawName: "B", RawBuild: BuildParameters{RawContext: "dockerfile1"}, RawImage: "image1"}, //dup of A
		&container{RawName: "C", RawBuild: BuildParameters{RawContext: "dockerfile1"}, RawImage: "image2"},
		&container{RawName: "D", RawBuild: BuildParameters{RawContext: "dockerfile1"}, RawImage: "image2"}, //dup of C
		&container{RawName: "E", RawBuild: BuildParameters{RawContext: "dockerfile2"}, RawImage: "image1"},
		&container{RawName: "F", RawBuild: BuildParameters{RawContext: "dockerfile2"}, RawImage: "image1"}, //dup of E
		&container{RawName: "G", RawBuild: BuildParameters{RawContext: "dockerfile1"}},
		&container{RawName: "H", RawBuild: BuildParameters{RawContext: "dockerfile1"}}, //dup of G
		&container{RawName: "I", RawImage: "image1"},
		&container{RawName: "J", RawImage: "image1"}, //dup of I
	}
	deduplicated := containers.stripProvisioningDuplicates()
	assert.Len(t, deduplicated, 5)
	assert.Len(t, containers, 10) // input was not mutated - further operations won't be affected
}
