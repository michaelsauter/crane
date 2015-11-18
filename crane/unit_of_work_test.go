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
