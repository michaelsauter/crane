package crane

import (
	"fmt"
	"strings"
)

// ContainerMap maps the container name
// to its configuration
type ContainerMap map[string]Container

// Unordered maps a container name
// to its dependencies
type Unordered map[string]*Dependencies

// Dependencies contain two fields:
// list: contains all dependencies
// linked: contains dependencies that
// are being linked to (which means they
// need to be running in order to be satisfied).
type Dependencies struct {
	list   []string
	linked []string
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
// to be running (for now that just means it's in
// the linked list)
func (d *Dependencies) mustRun(needle string) bool {
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

// order works on the containerMap and retuens the order
// The order can be reversed as well. This is needed as any command
// bringing up containers needs to bring up the dependencies first,
// but any command shutting down containers needs to shutdown the
// dependent containers first.
// If the order between 2 containers does not matter, they are sorted
// alphabetically.
// If the map cannot be resolved, and error is returned detailing
// which containers still have unresolved dependencies.
func (m ContainerMap) order(reversed bool) (order []string, err error) {
	unordered := m.unordered(reversed)
	alphabetical := m.alphabetical(reversed)

	success := true
	for success && len(unordered) > 0 {
		success = false
		for _, name := range alphabetical {
			if _, ok := unordered[name]; ok {
				dependencies := unordered[name]
				if dependencies.satisfied() {
					// Resolve "name" and continue with next iteration
					success = true
					order = append([]string{name}, order...)
					unordered.resolve(name)
					break
				}
			}
		}

		if !success && !reversed {
			// Could not resolve a dependency so far in this iteration,
			// but maybe one of the container already runs/exists?
			// This check does only make sense for the default order.
			for _, name := range alphabetical {
				if _, ok := unordered[name]; ok {
					dependencies := unordered[name]
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
	}

	// If we still have dependencies, the container map
	// cannot be resolved (cyclic or missing dependency found).
	if len(unordered) > 0 {
		unresolvedSet := make(map[string]bool)
		for _, name := range alphabetical {
			if _, ok := unordered[name]; ok {
				unresolvedSet[name] = true
			}
		}
		unresolved := []string{}
		for name, _ := range unresolvedSet {
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

// unordered returns a map describing the dependencies
// between the containers. This is used as a basis to
// determine the order.
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

// alphabetical returns the containers of the map in
// alphabetical order. If reversed is set to true, it
// returns reverse alphabetical order.
func (m ContainerMap) alphabetical(reversed bool) []string {
	alphabetical := []string{}
	inserted := false
	for toInsert, _ := range m {
		inserted = false
		for i, name := range alphabetical {
			if name > toInsert {
				before := make([]string, len(alphabetical[:i]))
				copy(before, alphabetical[:i])
				before = append(before, toInsert)
				alphabetical = append(before, alphabetical[i:]...)
				inserted = true
			}
		}
		if !inserted {
			alphabetical = append(alphabetical, toInsert)
		}
	}
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
// unordered map and remoes it from all
// dependencies.
func (u Unordered) resolve(resolved string) {
	if _, ok := u[resolved]; ok {
		delete(u, resolved)
	}
	for _, dependencies := range u {
		dependencies.remove(resolved)
	}
}
