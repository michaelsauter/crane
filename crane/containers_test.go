package crane

import (
	"github.com/stretchr/testify/assert"
	"testing"
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
		&container{RawName: "A", RawDockerfile: "dockerfile1", RawImage: "image1"},
		&container{RawName: "B", RawDockerfile: "dockerfile1", RawImage: "image1"}, //dup of A
		&container{RawName: "C", RawDockerfile: "dockerfile1", RawImage: "image2"},
		&container{RawName: "D", RawDockerfile: "dockerfile1", RawImage: "image2"}, //dup of C
		&container{RawName: "E", RawDockerfile: "dockerfile2", RawImage: "image1"},
		&container{RawName: "F", RawDockerfile: "dockerfile2", RawImage: "image1"}, //dup of E
		&container{RawName: "G", RawDockerfile: "dockerfile1"},
		&container{RawName: "H", RawDockerfile: "dockerfile1"}, //dup of G
		&container{RawName: "I", RawImage: "image1"},
		&container{RawName: "J", RawImage: "image1"}, //dup of I
	}
	deduplicated := containers.stripProvisioningDuplicates()
	assert.Len(t, deduplicated, 5)
	assert.Len(t, containers, 10) // input was not mutated - further operations won't be affected
}
