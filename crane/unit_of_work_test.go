package crane

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
	var uow *UnitOfWork
	var networkMap map[string]Network

	// no networks
	cfg = &config{networkMap: networkMap}
	uow = &UnitOfWork{}
	assert.Equal(t, []string{}, uow.RequiredNetworks())

	// some networks
	containerMap := NewStubbedContainerMap(true,
		&container{
			RawName: "a",
			RawNet:  "foo",
		},
		&container{
			RawName: "b",
			RawNet:  "bar",
		},
		&container{
			RawName: "c",
			RawNet:  "bar",
		},
	)
	networkMap = map[string]Network{
		"foo": &network{RawName: "foo"},
		"bar": &network{RawName: "bar"},
		"baz": &network{RawName: "baz"},
	}
	cfg = &config{containerMap: containerMap, networkMap: networkMap}
	uow = &UnitOfWork{order: []string{"a", "b", "c"}}
	assert.Equal(t, []string{"foo", "bar"}, uow.RequiredNetworks())
}

func TestRequiredVolumes(t *testing.T) {
	var uow *UnitOfWork
	var volumeMap map[string]Volume

	// no volumes
	cfg = &config{volumeMap: volumeMap}
	uow = &UnitOfWork{}
	assert.Equal(t, []string{}, uow.RequiredVolumes())

	// some volumes
	containerMap := NewStubbedContainerMap(true,
		&container{
			RawName:   "a",
			RawVolume: []string{"foo:/foo"},
		},
		&container{
			RawName:   "b",
			RawVolume: []string{"bar:/bar"},
		},
		&container{
			RawName:   "c",
			RawVolume: []string{"bar:/bar"},
		},
	)
	volumeMap = map[string]Volume{
		"foo": &volume{RawName: "foo"},
		"bar": &volume{RawName: "bar"},
		"baz": &volume{RawName: "baz"},
	}
	cfg = &config{containerMap: containerMap, volumeMap: volumeMap}
	uow = &UnitOfWork{order: []string{"a", "b", "c"}}
	assert.Equal(t, []string{"foo", "bar"}, uow.RequiredVolumes())
}
