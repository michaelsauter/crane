package crane

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewUnitOfWork(t *testing.T) {

	examples := []struct {
		dependencyMap map[string]*Dependencies
		targeted      []string
		expected      *UnitOfWork
		err           bool
	}{
		{ // resolvable map -> works
			dependencyMap: map[string]*Dependencies{
				"b": &Dependencies{All: []string{"c"}},
				"a": &Dependencies{All: []string{"b"}},
				"c": &Dependencies{All: []string{}},
			},
			targeted: []string{"a", "b", "c"},
			expected: &UnitOfWork{
				targeted:       []string{"a", "b", "c"},
				containers:     []string{"a", "b", "c"},
				order:          []string{"c", "b", "a"},
				requireStarted: []string{},
			},
			err: false,
		},
		{ // cyclic map -> fails
			dependencyMap: map[string]*Dependencies{
				"b": &Dependencies{All: []string{"c"}},
				"a": &Dependencies{All: []string{"b"}},
				"c": &Dependencies{All: []string{"a"}},
			},
			targeted: []string{"a", "b", "c"},
			err:      true,
		},
		{ // incomplete map -> fails
			dependencyMap: map[string]*Dependencies{
				"a": &Dependencies{All: []string{"b"}},
			},
			targeted: []string{"a"},
			err:      true,
		},
		{ // partial target -> works
			dependencyMap: map[string]*Dependencies{
				"b": &Dependencies{All: []string{"c"}, Link: []string{"c"}},
				"a": &Dependencies{All: []string{"b"}},
				"c": &Dependencies{All: []string{}},
			},
			targeted: []string{"a", "b"},
			expected: &UnitOfWork{
				targeted:       []string{"a", "b"},
				containers:     []string{"a", "b", "c"},
				order:          []string{"c", "b", "a"},
				requireStarted: []string{"c"},
			},
			err: false,
		},
	}

	for _, example := range examples {
		uow, err := NewUnitOfWork(example.dependencyMap, example.targeted)
		if example.err {
			assert.Error(t, err)
		} else {
			if assert.NoError(t, err) {
				assert.Equal(t, example.expected, uow)
			}
		}
	}
}

func TestRequiredNetworks(t *testing.T) {
	containerMap := NewStubbedContainerMap(true,
		&container{
			RawName: "a",
			RawRun: RunParameters{
				RawNet: "foo",
			},
		},
		&container{
			RawName: "b",
			RawRun: RunParameters{
				RawNet: "bar",
			},
		},
	)

	networkMap := map[string]Network{
		"foo": &network{RawName: "foo"},
		"bar": &network{RawName: "bar"},
		"baz": &network{RawName: "baz"},
	}

	cfg = &config{containerMap: containerMap, networkMap: networkMap}
	uow := &UnitOfWork{order: []string{"a", "b"}}
	requiredNetworks := []string{}
	for _, network := range uow.requiredNetworks() {
		requiredNetworks = append(requiredNetworks, network.Name())
	}
	assert.Equal(t, []string{"foo", "bar"}, requiredNetworks)
}

func TestRequiredVolumes(t *testing.T) {
	containerMap := NewStubbedContainerMap(true,
		&container{
			RawName: "a",
			RawRun: RunParameters{
				RawVolume: []string{"foo:/foo"},
			},
		},
		&container{
			RawName: "b",
			RawRun: RunParameters{
				RawVolume: []string{"bar:/bar"},
			},
		},
	)

	volumeMap := map[string]Volume{
		"foo": &volume{RawName: "foo"},
		"bar": &volume{RawName: "bar"},
		"baz": &volume{RawName: "baz"},
	}

	cfg = &config{containerMap: containerMap, volumeMap: volumeMap}
	uow := &UnitOfWork{order: []string{"a", "b"}}
	requiredVolumes := []string{}
	for _, network := range uow.requiredVolumes() {
		requiredVolumes = append(requiredVolumes, network.Name())
	}
	assert.Equal(t, []string{"foo", "bar"}, requiredVolumes)
}
