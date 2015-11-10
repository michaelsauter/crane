package crane

// DependencyGraph maps container names
// to their dependencies
type DependencyGraph map[string]*Dependencies

// resolve deletes the given name from the
// dependency graph and removes it from all
// dependencies.
func (graph DependencyGraph) resolve(resolved string) {
	delete(graph, resolved)
	for _, dependencies := range graph {
		dependencies.remove(resolved)
	}
}
