package crane

import (
	"fmt"
	"strings"
)

// DependencyGraph maps container names
// to their dependencies
type DependencyGraph map[string]*Dependencies

// order works on the dependency graph and returns the order
// of the given the target (a subset of the graph).
func (graph DependencyGraph) order(target Target, ignoreMissing string) (order []string, err error) {
	success := true
	for success && len(order) < len(target) {
		success = false

		// Try to resolve container with satisfied dependencies
		for _, name := range target {
			if dependencies, ok := graph[name]; ok {
				if dependencies.satisfied() {
					// Resolve "name" and continue with next iteration
					success = true
					order = append([]string{name}, order...)
					graph.resolve(name)
					break
				}
			}
		}

		// If we could not resolve a container so far in this iteration,
		// we check if one of the non-targeted dependencies already runs or exists.
		if !success {
			for _, name := range target {
				if dependencies, ok := graph[name]; ok {
					for _, name := range dependencies.All {
						if !target.includes(name) {
							container := graph.tmpContainer(name)
							satisfied := false
							if dependencies.mustRun(name) {
								satisfied = container.Running()
							} else {
								satisfied = container.Exists()
							}
							if satisfied {
								success = true
								graph.resolve(name)
								break
							}
						}
					}
				}
			}
		}

		// If we still have not resolved a dependency, we can't
		// order the given target properly. However, if some
		// type of dependencies can be ignored is missing, we
		// just resolve the first non-targeted dependency.
		if !success && ignoreMissing != "none" {
			for _, name := range target {
				if dependencies, ok := graph[name]; ok {
					// Note about the corner case of dependencies of two kinds (both
					// link & volumesFrom for example): a missing dependency is resolved
					// if at least one of its kinds is requested to be ignored, given the
					// current naive check below. That can trigger false negatives here, but
					// further on, the "other" kind won't be removed from the CLI command, so
					// it will just fail a bit less nicely
					for _, name := range dependencies.forKind(ignoreMissing) {
						if !target.includes(name) {
							success = true
							graph.resolve(name)
							break
						}
					}
				}
			}
		}
	}

	// If we the order is not complete yet, the target
	// cannot be resolved (cyclic or missing dependency found).
	if len(order) < len(target) {
		unresolved := []string{}
		for _, name := range target {
			if _, ok := graph[name]; ok {
				unresolved = append(unresolved, name)
			}
		}
		err = fmt.Errorf("Dependencies for container(s) %s could not be resolved. Check for cyclic or missing dependencies, ignore them via -i/--ignore-missing or use -d/--cascade-dependencies to automatically attempt to recursively include dependencies in the set of targeted containers.", strings.Join(unresolved, ", "))
	}

	return
}

func (d DependencyGraph) tmpContainer(name string) Container {
	return &container{RawName: name}
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
