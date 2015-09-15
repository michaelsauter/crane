package crane

import (
	"io"
	"text/template"
)

// DependencyGraph maps container names
// to their dependencies
type DependencyGraph map[string]*Dependencies

type dotInput struct {
	Graph              DependencyGraph
	TargetedContainers Containers
}

// dumps the dependency graph as a DOT to the writer
func (graph DependencyGraph) DOT(writer io.Writer, targetedContainers Containers) {
	const dotTemplate = `{{ $targetedContainers := .TargetedContainers }}digraph {
{{ range $name, $dependencies := .Graph }}{{ with $dependencies }}  "{{ $name }}" [style=bold{{ range $targetedContainers }}{{ if eq $name .Name }},color=red{{ end }}{{ end }}]
{{ range .Link }}  "{{ $name }}"->"{{ . }}"
{{ end }}{{ range .VolumesFrom }}  "{{ $name }}"->"{{ . }}" [style=dashed]
{{ end }}{{ if ne .Net "" }}  "{{ $name }}"->"{{ .Net }}" [style=dotted]
{{ end }}{{ end }}{{ end }}}
`
	template, err := template.New("dot").Parse(dotTemplate)
	if err != nil {
		printErrorf("ERROR: %s\n", err)
		return
	}
	err = template.Execute(writer, dotInput{graph, targetedContainers})
	if err != nil {
		printErrorf("ERROR: %s\n", err)
	}
}

// resolve deletes the given name from the
// dependency graph and removes it from all
// dependencies.
func (d DependencyGraph) resolve(resolved string) {
	delete(d, resolved)
	for _, dependencies := range d {
		dependencies.remove(resolved)
	}
}
