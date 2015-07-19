package crane

import (
	"bytes"
	"github.com/stretchr/testify/assert"
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
	assert.Equal(t, expected, buffer.String())
}

// func TestOrder(t *testing.T) {

// 	examples := []struct {
// 		graph         DependencyGraph
// 		target        []string
// 		ignoreMissing string
// 		expected      []string
// 		err           bool
// 	}{
// 		{ // resolvable map -> works
// 			graph: DependencyGraph{
// 				"b": &Dependencies{All: []string{"c"}},
// 				"a": &Dependencies{All: []string{"b"}},
// 				"c": &Dependencies{All: []string{}},
// 			},
// 			target:        []string{"a", "b", "c"},
// 			ignoreMissing: "none",
// 			expected:      []string{"a", "b", "c"},
// 			err:           false,
// 		},
// 		{ // cyclic map -> fails
// 			graph: DependencyGraph{
// 				"b": &Dependencies{All: []string{"c"}},
// 				"a": &Dependencies{All: []string{"b"}},
// 				"c": &Dependencies{All: []string{"a"}},
// 			},
// 			target:        []string{"a", "b", "c"},
// 			ignoreMissing: "none",
// 			err:           true,
// 		},
// 		{ // cyclic map, ignoring all -> fails
// 			graph: DependencyGraph{
// 				"b": &Dependencies{All: []string{"c"}},
// 				"a": &Dependencies{All: []string{"b"}},
// 				"c": &Dependencies{All: []string{"a"}},
// 			},
// 			target:        []string{"a", "b", "c"},
// 			ignoreMissing: "all",
// 			err:           true,
// 		},
// 		{ // partial target -> fails
// 			graph: DependencyGraph{
// 				"b": &Dependencies{All: []string{"c"}, VolumesFrom: []string{"c"}},
// 				"a": &Dependencies{All: []string{"b"}, Link: []string{"b"}},
// 				"c": &Dependencies{All: []string{}},
// 			},
// 			target:        []string{"a", "b"},
// 			ignoreMissing: "none",
// 			err:           true,
// 		},
// 		{ // partial target, ignoring all -> works
// 			graph: DependencyGraph{
// 				"b": &Dependencies{All: []string{"c"}, VolumesFrom: []string{"c"}},
// 				"a": &Dependencies{All: []string{"b"}, Link: []string{"b"}},
// 				"c": &Dependencies{All: []string{}, Net: "d"},
// 			},
// 			target:        []string{"a", "b"},
// 			ignoreMissing: "all",
// 			expected:      []string{"a", "b"},
// 			err:           false,
// 		},
// 		{ // partial target, ignoring link -> works
// 			graph: DependencyGraph{
// 				"b": &Dependencies{All: []string{"c"}, VolumesFrom: []string{"c"}},
// 				"a": &Dependencies{All: []string{"b"}, Link: []string{"b"}},
// 				"c": &Dependencies{All: []string{}, Net: "d"},
// 			},
// 			target:        []string{"a"},
// 			ignoreMissing: "link",
// 			expected:      []string{"a"},
// 			err:           false,
// 		},
// 		{ // partial target, ignoring volumesFrom -> works
// 			graph: DependencyGraph{
// 				"b": &Dependencies{All: []string{"c"}, VolumesFrom: []string{"c"}},
// 				"a": &Dependencies{All: []string{"b"}, Link: []string{"b"}},
// 				"c": &Dependencies{All: []string{}, Net: "d"},
// 			},
// 			target:        []string{"b"},
// 			ignoreMissing: "volumesFrom",
// 			expected:      []string{"b"},
// 			err:           false,
// 		},
// 		{ // partial target, ignoring net -> works
// 			graph: DependencyGraph{
// 				"b": &Dependencies{All: []string{"c"}, VolumesFrom: []string{"c"}},
// 				"a": &Dependencies{All: []string{"b"}, Link: []string{"b"}},
// 				"c": &Dependencies{All: []string{}, Net: "d"},
// 			},
// 			target:        []string{"c"},
// 			ignoreMissing: "net",
// 			expected:      []string{"c"},
// 			err:           false,
// 		},
// 	}

// 	for _, example := range examples {
// 		order, err := example.graph.order(example.target, example.ignoreMissing)
// 		if example.err {
// 			assert.Error(t, err)
// 		} else {
// 			if assert.NoError(t, err) {
// 				assert.Equal(t, example.expected, order)
// 			}
// 		}
// 	}
// }
