package crane

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestIncludes(t *testing.T) {
	dependencies := Dependencies{
		All:         []string{"link", "volumesFrom", "net"},
		Link:        []string{"link"},
		VolumesFrom: []string{"volumesFrom"},
		Net:         "net",
	}
	assert.True(t, dependencies.includes("link"))
	assert.True(t, dependencies.includes("volumesFrom"))
	assert.True(t, dependencies.includes("net"))
	assert.False(t, dependencies.includes("non-existent"))
}

func TestSatisfied(t *testing.T) {
	var dependencies Dependencies

	dependencies = Dependencies{
		All: []string{"a"},
	}
	assert.False(t, dependencies.satisfied(), "Dependencies was not empty, but appeared to be satisfied")

	dependencies = Dependencies{
		All: []string{},
	}
	assert.True(t, dependencies.satisfied(), "Dependencies was empty, but appeared not to be satisfied")
}
