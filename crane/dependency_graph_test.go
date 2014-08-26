package crane

import (
	"reflect"
	"testing"
)

func TestOrder(t *testing.T) {

	examples := []struct {
		graph      DependencyGraph
		target     []string
		forceOrder bool
		expected   []string
		err        bool
	}{
		{ // resolvable map -> works
			graph: DependencyGraph{
				"b": &Dependencies{All: []string{"c"}},
				"a": &Dependencies{All: []string{"b"}},
				"c": &Dependencies{All: []string{}},
			},
			target:     []string{"a", "b", "c"},
			forceOrder: false,
			expected:   []string{"a", "b", "c"},
			err:        false,
		},
		{ // cyclic map, unforced -> fails
			graph: DependencyGraph{
				"b": &Dependencies{All: []string{"c"}},
				"a": &Dependencies{All: []string{"b"}},
				"c": &Dependencies{All: []string{"a"}},
			},
			target:     []string{"a", "b", "c"},
			forceOrder: false,
			expected:   []string{},
			err:        true,
		},
		{ // cyclic map, forced -> fails
			graph: DependencyGraph{
				"b": &Dependencies{All: []string{"c"}},
				"a": &Dependencies{All: []string{"b"}},
				"c": &Dependencies{All: []string{"a"}},
			},
			target:     []string{"a", "b", "c"},
			forceOrder: true,
			expected:   []string{},
			err:        true,
		},
		{ // partial target, unforced -> fails
			graph: DependencyGraph{
				"b": &Dependencies{All: []string{"c"}},
				"a": &Dependencies{All: []string{"b"}},
				"c": &Dependencies{All: []string{}},
			},
			target:     []string{"a", "b"},
			forceOrder: false,
			expected:   []string{},
			err:        true,
		},
		{ // partial target, forced -> works
			graph: DependencyGraph{
				"b": &Dependencies{All: []string{"c"}},
				"a": &Dependencies{All: []string{"b"}},
				"c": &Dependencies{All: []string{}},
			},
			target:     []string{"a", "b"},
			forceOrder: true,
			expected:   []string{"a", "b"},
			err:        false,
		},
	}

	for _, example := range examples {
		order, err := example.graph.order(example.target, example.forceOrder)
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
