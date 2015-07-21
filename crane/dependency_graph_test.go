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
