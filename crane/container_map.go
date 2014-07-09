package crane

import (
	"fmt"
	"strings"
)

type ContainerMap map[string]Container

type Dependencies map[string][]string

func (m ContainerMap) order(reversed bool) (order []string, err error) {
	dependencies := m.extractDependencies(reversed)

	success := true
	for success && len(dependencies) > 0 {
		success = false
		for name, unresolved := range dependencies {
			if len(unresolved) == 0 {
				// Resolve "name" and continue with next iteration
				success = true
				order = append([]string{name}, order...)
				dependencies.resolve(name)
				break
			}
		}

		if !success && !reversed {
			// Could not resolve a dependency so far in this iteration,
			// but maybe one of the container already exists?
			// This check does only make sense for the default order.
			for _, unresolved := range dependencies {
				// Loop over unresolved containers
				for _, name := range unresolved {
					// Container must not be part of the map that
					// is currently targeted.
					if _, ok := m[name]; !ok {
						// Need to "fake" a container here because
						// it can't be retrieved from the map (as it was removed).
						container := &Container{RawName: name}
						if container.exists() {
							success = true
							dependencies.resolve(name)
							break
						}
					}
				}
			}
		}
	}

	// If we still have dependencies, the container map
	// cannot be resolved (cyclic or missing dependency found).
	if len(dependencies) > 0 {
		unresolved := []string{}
		for name, _ := range dependencies {
			unresolved = append(unresolved, name)
		}
		// For reversed order, that is okay.
		// Otherwise, this is an error that needs
		// to be resolved.
		if reversed {
			order = append(order, unresolved...)
		} else {
			err = fmt.Errorf("Container(s) %s could not be resolved. Check for cyclic or missing dependencies.", strings.Join(unresolved, ", "))
		}
	}

	return
}

func (m ContainerMap) extractDependencies(reversed bool) Dependencies {
	dependencies := make(map[string][]string)

	if reversed {
		for _, container := range m {
			for _, dep := range container.Dependencies() {
				if _, ok := dependencies[dep]; ok {
					dependencies[dep] = append(dependencies[dep], container.Name())
				} else {
					dependencies[dep] = []string{container.Name()}
				}
			}
			if _, ok := dependencies[container.Name()]; !ok {
				dependencies[container.Name()] = []string{}
			}
		}
	} else {
		for _, container := range m {
			dependencies[container.Name()] = container.Dependencies()
		}
	}

	return dependencies
}

func (d Dependencies) resolve(resolved string) {
	if _, ok := d[resolved]; ok {
		delete(d, resolved)
	}
	// Remove "resolved" from "unresolved" slices
	for name, unresolved := range d {
		for i, dep := range unresolved {
			if dep == resolved {
				d[name] = append(unresolved[:i], unresolved[i+1:]...)
			}
		}
	}
}
