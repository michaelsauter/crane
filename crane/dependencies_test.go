package crane

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIncludes(t *testing.T) {
	dependencies := Dependencies{
		All:         []string{"required", "link", "volumesFrom", "net", "ipc"},
		Requires:    []string{"required"},
		Link:        []string{"link"},
		VolumesFrom: []string{"volumesFrom"},
		Net:         "net",
		IPC:         "ipc",
	}
	assert.True(t, dependencies.includes("required"))
	assert.True(t, dependencies.includes("link"))
	assert.True(t, dependencies.includes("volumesFrom"))
	assert.True(t, dependencies.includes("net"))
	assert.True(t, dependencies.includes("ipc"))
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
