package crane

import (
	"reflect"
	"testing"
)

func TestOrder(t *testing.T) {

	examples := []struct {
		graph         DependencyGraph
		target        []string
		ignoreMissing string
		expected      []string
		err           bool
	}{
		{ // resolvable map -> works
			graph: DependencyGraph{
				"b": &Dependencies{All: []string{"c"}},
				"a": &Dependencies{All: []string{"b"}},
				"c": &Dependencies{All: []string{}},
			},
			target:        []string{"a", "b", "c"},
			ignoreMissing: "none",
			expected:      []string{"a", "b", "c"},
			err:           false,
		},
		{ // cyclic map -> fails
			graph: DependencyGraph{
				"b": &Dependencies{All: []string{"c"}},
				"a": &Dependencies{All: []string{"b"}},
				"c": &Dependencies{All: []string{"a"}},
			},
			target:        []string{"a", "b", "c"},
			ignoreMissing: "none",
			err:           true,
		},
		{ // cyclic map, ignoring all -> fails
			graph: DependencyGraph{
				"b": &Dependencies{All: []string{"c"}},
				"a": &Dependencies{All: []string{"b"}},
				"c": &Dependencies{All: []string{"a"}},
			},
			target:        []string{"a", "b", "c"},
			ignoreMissing: "all",
			err:           true,
		},
		{ // partial target -> fails
			graph: DependencyGraph{
				"b": &Dependencies{All: []string{"c"}, VolumesFrom: []string{"c"}},
				"a": &Dependencies{All: []string{"b"}, Link: []string{"b"}},
				"c": &Dependencies{All: []string{}},
			},
			target:        []string{"a", "b"},
			ignoreMissing: "none",
			err:           true,
		},
		{ // partial target, ignoring all -> works
			graph: DependencyGraph{
				"b": &Dependencies{All: []string{"c"}, VolumesFrom: []string{"c"}},
				"a": &Dependencies{All: []string{"b"}, Link: []string{"b"}},
				"c": &Dependencies{All: []string{}, Net: "d"},
			},
			target:        []string{"a", "b"},
			ignoreMissing: "all",
			expected:      []string{"a", "b"},
			err:           false,
		},
		{ // partial target, ignoring link -> works
			graph: DependencyGraph{
				"b": &Dependencies{All: []string{"c"}, VolumesFrom: []string{"c"}},
				"a": &Dependencies{All: []string{"b"}, Link: []string{"b"}},
				"c": &Dependencies{All: []string{}, Net: "d"},
			},
			target:        []string{"a"},
			ignoreMissing: "link",
			expected:      []string{"a"},
			err:           false,
		},
		{ // partial target, ignoring volumesFrom -> works
			graph: DependencyGraph{
				"b": &Dependencies{All: []string{"c"}, VolumesFrom: []string{"c"}},
				"a": &Dependencies{All: []string{"b"}, Link: []string{"b"}},
				"c": &Dependencies{All: []string{}, Net: "d"},
			},
			target:        []string{"b"},
			ignoreMissing: "volumesFrom",
			expected:      []string{"b"},
			err:           false,
		},
		{ // partial target, ignoring net -> works
			graph: DependencyGraph{
				"b": &Dependencies{All: []string{"c"}, VolumesFrom: []string{"c"}},
				"a": &Dependencies{All: []string{"b"}, Link: []string{"b"}},
				"c": &Dependencies{All: []string{}, Net: "d"},
			},
			target:        []string{"c"},
			ignoreMissing: "net",
			expected:      []string{"c"},
			err:           false,
		},
	}

	for _, example := range examples {
		order, err := example.graph.order(example.target, example.ignoreMissing)
		if example.err {
			if err == nil {
				t.Errorf("Should have not gotten an order, got %v", order)
			}
		} else {
			if err != nil || !reflect.DeepEqual(order, example.expected) {
				t.Errorf("Order should have been %v, got %v. Err: %v", example.expected, order, err)
			}
		}
	}
}
