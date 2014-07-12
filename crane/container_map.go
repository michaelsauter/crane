package crane

import (
	"fmt"
	"strings"
)

type ContainerMap map[string]Container

type Unordered map[string]*Dependencies

type Dependencies struct {
	list   []string
	linked []string
}

func (d *Dependencies) includes(needle string) bool {
	for _, name := range d.list {
		if name == needle {
			return true
		}
	}
	return false
}

func (d *Dependencies) mustRun(needle string) bool {
	for _, name := range d.linked {
		if name == needle {
			return true
		}
	}
	return false
}

func (d *Dependencies) satisfied() bool {
	return len(d.list) == 0
}

func (d *Dependencies) remove(resolved string) {
	for i, name := range d.list {
		if name == resolved {
			d.list = append(d.list[:i], d.list[i+1:]...)
		}
	}
}

func (m ContainerMap) order(reversed bool) (order []string, err error) {
	unordered := m.unordered(reversed)

	success := true
	for success && len(unordered) > 0 {
		success = false
		for name, dependencies := range unordered {
			if dependencies.satisfied() {
				// Resolve "name" and continue with next iteration
				success = true
				order = append([]string{name}, order...)
				unordered.resolve(name)
				break
			}
		}

		if !success && !reversed {
			// Could not resolve a dependency so far in this iteration,
			// but maybe one of the container already runs/exists?
			// This check does only make sense for the default order.
			for _, dependencies := range unordered {
				// Loop over dependencies that need to be running
				for _, name := range dependencies.list {
					// Container must not be part of the map that
					// is currently targeted.
					if _, ok := m[name]; !ok {
						// Need to "fake" a container here because
						// it can't be retrieved from the map (as it was removed).
						container := &Container{RawName: name}
						satisfied := false
						if dependencies.mustRun(name) {
							satisfied = container.running()
						} else {
							satisfied = container.exists()
						}
						if satisfied {
							success = true
							unordered.resolve(name)
							break
						}
					}
				}
			}
		}
	}

	// If we still have dependencies, the container map
	// cannot be resolved (cyclic or missing dependency found).
	if len(unordered) > 0 {
		unresolved := []string{}
		for name, _ := range unordered {
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

func (m ContainerMap) unordered(reversed bool) Unordered {
	unordered := make(map[string]*Dependencies)

	if reversed {
		// Need to set every "dependency" as the key of the
		// unordered map and list the containers that depend
		// on it as the dependencies ...
		// Iterate over map
		for _, container := range m {
			// Get dependency list of each container
			for _, dep := range container.Dependencies().list {
				// If unordered already has the key, append to the list,
				// otherwise create that dependecy
				if _, ok := unordered[dep]; ok {
					unordered[dep].list = append(unordered[dep].list, container.Name())
				} else {
					unordered[dep] = &Dependencies{list: []string{container.Name()}}
				}
			}
			// If we haven't created the key yet, do it now
			if _, ok := unordered[container.Name()]; !ok {
				unordered[container.Name()] = &Dependencies{list: []string{}}
			}
		}
	} else {
		// For default order, just map the container names to their dependencies
		for _, container := range m {
			unordered[container.Name()] = container.Dependencies()
		}
	}

	return unordered
}

func (u Unordered) resolve(resolved string) {
	if _, ok := u[resolved]; ok {
		delete(u, resolved)
	}
	for _, dependencies := range u {
		dependencies.remove(resolved)
	}
}
