package crane

import (
	"fmt"
	"sort"
	"strings"
)

// ContainerMap maps the container name
// to its configuration
type ContainerMap map[string]Container

// DependenciesMap maps the container name
// to its dependencies
type DependenciesMap map[string]*Dependencies

// Dependencies contain 3 fields:
// list: contains all dependencies
// linked: contains dependencies that
// are being linked to.
// net: container the net stack is shared with
type Dependencies struct {
	list   []string
	linked []string
	net    string
}

// includes checks whether the given needle is
// included in the dependency list
func (d *Dependencies) includes(needle string) bool {
	for _, name := range d.list {
		if name == needle {
			return true
		}
	}
	return false
}

// mustRun checks whether the given needle needs
// to be running
func (d *Dependencies) mustRun(needle string) bool {
	if needle == d.net {
		return true
	}
	for _, name := range d.linked {
		if name == needle {
			return true
		}
	}
	return false
}

// satisfied is true when the list is empty
func (d *Dependencies) satisfied() bool {
	return len(d.list) == 0
}

// remove removes the given name from the list
func (d *Dependencies) remove(resolved string) {
	for i, name := range d.list {
		if name == resolved {
			d.list = append(d.list[:i], d.list[i+1:]...)
		}
	}
}

// returns a copy of the map where all containers are within the given list
func (m ContainerMap) subset(included []string) ContainerMap {
	newMap := make(ContainerMap)
	for _, name := range included {
		if container, present := m[name]; present {
			newMap[name] = container
		}
	}
	return newMap
}

// order works on the containerMap and retuens the order
// The order can be reversed as well. This is needed as any command
// bringing up containers needs to bring up the dependencies first,
// but any command shutting down containers needs to shutdown the
// dependent containers first.
// If the order between 2 containers does not matter, they are sorted
// alphabetically.
// If the map cannot be resolved, and error is returned detailing
// which containers still have unresolved dependencies.
func (m ContainerMap) order(reversed bool, ignoreUnresolved bool) (order []string, err error) {
	dependenciesMap := m.dependencies(reversed)
	alphabetical := m.alphabetical(reversed)

	success := true
	for success && len(dependenciesMap) > 0 {
		success = false
		for _, name := range alphabetical {
			if dependencies, ok := dependenciesMap[name]; ok {
				if dependencies.satisfied() {
					// Resolve "name" and continue with next iteration
					success = true
					order = append([]string{name}, order...)
					dependenciesMap.resolve(name)
					break
				}
			}
		}

		if !success && !reversed {
			// Could not resolve a dependency so far in this iteration,
			// but maybe one of the container already runs/exists?
			// This check does only make sense for the default order.
			for _, name := range alphabetical {
				if dependencies, ok := dependenciesMap[name]; ok {
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
								dependenciesMap.resolve(name)
								break
							}
						}
					}
				}
			}
		}
	}

	// If we still have dependencies, the container map
	// cannot be resolved (cyclic or missing dependency found).
	if len(dependenciesMap) > 0 {
		unresolved := []string{}
		for _, name := range alphabetical {
			if _, ok := dependenciesMap[name]; ok {
				unresolved = append(unresolved, name)
			}
		}
		if ignoreUnresolved {
			order = append(order, unresolved...)
		} else {
			err = fmt.Errorf("Dependencies for container(s) %s could not be resolved. Check for cyclic or missing dependencies.", strings.Join(unresolved, ", "))
		}
	}

	return
}

// dependencies returns a map describing the dependencies
// between the containers.
func (m ContainerMap) dependencies(reversed bool) DependenciesMap {
	dependenciesMap := make(DependenciesMap)

	if reversed {
		// Need to set every "dependency" as the key of the
		// dependenciesMap map and list the containers that depend
		// on it as the dependencies ...
		// Iterate over map
		for _, container := range m {
			// Get dependency list of each container
			for _, dep := range container.Dependencies().list {
				// If dependenciesMap already has the key, append to the list,
				// otherwise create that dependecy
				if dependencies, ok := dependenciesMap[dep]; ok {
					dependencies.list = append(dependencies.list, container.Name())
				} else {
					dependenciesMap[dep] = &Dependencies{list: []string{container.Name()}}
				}
			}
			// If we haven't created the key yet, do it now
			if _, ok := dependenciesMap[container.Name()]; !ok {
				dependenciesMap[container.Name()] = &Dependencies{list: []string{}}
			}
		}
	} else {
		// For default order, just map the container names to their dependencies
		for _, container := range m {
			dependenciesMap[container.Name()] = container.Dependencies()
		}
	}

	return dependenciesMap
}

// alphabetical returns the containers of the map in
// alphabetical order. If reversed is set to true, it
// returns reverse alphabetical order.
func (m ContainerMap) alphabetical(reversed bool) []string {
	alphabetical := []string{}

	for name, _ := range m {
		alphabetical = append(alphabetical, name)
	}

	sort.Strings(alphabetical)

	if reversed {
		reverse := []string{}
		for _, name := range alphabetical {
			reverse = append([]string{name}, reverse...)
		}
		alphabetical = reverse
	}

	return alphabetical
}

// resolve deletes the given name from the
// dependenciesMap and removes it from all
// dependencies.
func (d DependenciesMap) resolve(resolved string) {
	delete(d, resolved)
	for _, dependencies := range d {
		dependencies.remove(resolved)
	}
}
