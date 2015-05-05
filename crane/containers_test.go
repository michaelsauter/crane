package crane

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNames(t *testing.T) {
	var containers Containers
	containers = []Container{&container{RawName: "a"}, &container{RawName: "b"}}
	assert.Equal(t, []string{"a", "b"}, containers.names())

}

func TestReversed(t *testing.T) {
	var containers Containers
	containers = []Container{&container{RawName: "a"}, &container{RawName: "b"}}
	reversed := containers.reversed()
	assert.Len(t, reversed, 2)
	assert.Equal(t, "b", reversed[0].Name())
	assert.Equal(t, "a", reversed[1].Name())
}
