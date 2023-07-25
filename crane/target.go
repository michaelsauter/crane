package crane

import (
	"sort"
)

type Target struct {
	initial      []string
	dependencies []string
}

// NewTarget receives the specified target
// and determines which containers should be targeted.
// The target might be extended to dependencies if --extend is given.
// Additionally, the target is sorted alphabetically.
func NewTarget(dependencyMap map[string]*Dependencies, targetArg string, extendFlag bool) (target Target, err error) {

	target = Target{
		initial:      []string{},
		dependencies: []string{},
	}

	initialTarget := cfg.ContainersForReference(targetArg)
	for _, c := range initialTarget {
		if includes(allowed, c) {
			target.initial = append(target.initial, c)
		}
	}

	if extendFlag {
		var (
			dependenciesSet = make(map[string]struct{})
			cascadingSeeds  = []string{}
		)
		// start from the explicitly targeted target
		for _, name := range target.initial {
			dependenciesSet[name] = struct{}{}
			cascadingSeeds = append(cascadingSeeds, name)
		}

		// Cascade until the dependency map has been fully traversed
		// according to the cascading flags.
		for len(cascadingSeeds) > 0 {
			nextCascadingSeeds := []string{}
			for _, seed := range cascadingSeeds {
				if dependencies, ok := dependencyMap[seed]; ok {
					// Queue direct dependencies if we haven't already considered them
					for _, name := range dependencies.All {
						if _, alreadyIncluded := dependenciesSet[name]; !alreadyIncluded {
							dependenciesSet[name] = struct{}{}
							nextCascadingSeeds = append(nextCascadingSeeds, name)
						}
					}
				}
			}
			cascadingSeeds = nextCascadingSeeds
		}

		for name := range dependenciesSet {
			if !includes(target.initial, name) {
				target.dependencies = append(target.dependencies, name)
			}
		}

		sort.Strings(target.dependencies)
	}

	return
}

// Return all targeted containers, sorted alphabetically
func (t Target) all() []string {
	all := t.initial
	all = append(all, t.dependencies...)
	sort.Strings(all)
	return all
}
