package crane

import (
	"bytes"
	"reflect"
	"testing"
)

func TestDOT(t *testing.T) {
	dependencyGraph := DependencyGraph{
		"b": &Dependencies{Link: []string{"c"}, VolumesFrom: []string{"a"}},
		"a": &Dependencies{Link: []string{"c"}},
		"c": &Dependencies{Net: "d"},
	}
	var buffer bytes.Buffer
	dependencyGraph.DOT(&buffer, Containers{&container{RawName: "a"}, &container{RawName: "b"}})
	expected := `digraph {
  "a" [style=bold,color=red]
  "a"->"c"
  "b" [style=bold,color=red]
  "b"->"c"
  "b"->"a" [style=dashed]
  "c" [style=bold]
  "c"->"d" [style=dotted]
}
`
	if expected != buffer.String() {
		t.Errorf("Invalid graph received. Expected `%v`, but got `%v`", expected, buffer.String())
	}
}

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
