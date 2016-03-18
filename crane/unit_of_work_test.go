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
	var uow *UnitOfWork
	var networkMap map[string]NetworkCommander

	// no networks
	cfg = &Config{networkMap: networkMap}
	uow = &UnitOfWork{}
	assert.Equal(t, []string{}, uow.RequiredNetworks())

	// some networks
	containerMap := NewStubbedContainerMap(true,
		&Container{
			RawName: "a",
			RawRun: RunParameters{
				RawNet: "foo",
			},
		},
		&Container{
			RawName: "b",
			RawRun: RunParameters{
				RawNet: "bar",
			},
		},
		&Container{
			RawName: "c",
			RawRun: RunParameters{
				RawNet: "bar",
			},
		},
	)
	networkMap = map[string]NetworkCommander{
		"foo": &Network{RawName: "foo"},
		"bar": &Network{RawName: "bar"},
		"baz": &Network{RawName: "baz"},
	}
	cfg = &Config{containerMap: containerMap, networkMap: networkMap}
	uow = &UnitOfWork{order: []string{"a", "b", "c"}}
	assert.Equal(t, []string{"foo", "bar"}, uow.RequiredNetworks())
}

func TestRequiredVolumes(t *testing.T) {
	var uow *UnitOfWork
	var volumeMap map[string]VolumeCommander

	// no volumes
	cfg = &Config{volumeMap: volumeMap}
	uow = &UnitOfWork{}
	assert.Equal(t, []string{}, uow.RequiredVolumes())

	// some volumes
	containerMap := NewStubbedContainerMap(true,
		&Container{
			RawName: "a",
			RawRun: RunParameters{
				RawVolume: []string{"foo:/foo"},
			},
		},
		&Container{
			RawName: "b",
			RawRun: RunParameters{
				RawVolume: []string{"bar:/bar"},
			},
		},
		&Container{
			RawName: "c",
			RawRun: RunParameters{
				RawVolume: []string{"bar:/bar"},
			},
		},
	)
	volumeMap = map[string]VolumeCommander{
		"foo": &Volume{RawName: "foo"},
		"bar": &Volume{RawName: "bar"},
		"baz": &Volume{RawName: "baz"},
	}
	cfg = &Config{containerMap: containerMap, volumeMap: volumeMap}
	uow = &UnitOfWork{order: []string{"a", "b", "c"}}
	assert.Equal(t, []string{"foo", "bar"}, uow.RequiredVolumes())
}
