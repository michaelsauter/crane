package crane

import (
	"fmt"
	"sort"
	"strings"
)

type Target struct {
	initial      []string
	dependencies []string
	affected     []string
}

// NewTarget receives the specified target
// and determines which containers should be targeted.
// The target might be extended depending whether the
// dynamic targets "dependencies" and/or "affected"
// are included in the targetFlag.
// Additionally, the target is sorted alphabetically.
func NewTarget(dependencyMap map[string]*Dependencies, targetFlag string, excluded []string) (target Target, err error) {

	targetParts := strings.Split(targetFlag, "+")
	targetName := targetParts[0]
	extendDependencies := false
	extendAffected := false
	for _, v := range targetParts[1:] {
		if v == "dependencies" || v == "d" {
			extendDependencies = true
		} else if v == "affected" || v == "a" {
			extendAffected = true
		} else {
			err = fmt.Errorf("Unknown target extension %s. Available options are 'dependencies'/'d' and 'affected'/'a'", v)
			return
		}
	}

	target = Target{
		initial:      []string{},
		dependencies: []string{},
		affected:     []string{},
	}

	initialTarget := cfg.ContainersForReference(targetName)
	for _, c := range initialTarget {
		if !includes(excluded, c) {
			target.initial = append(target.initial, c)
		}
	}

	includedSet := make(map[string]bool)
	cascadingSeeds := []string{}

	if extendDependencies {
		// start from the explicitly targeted target
		includedSet = make(map[string]bool)
		cascadingSeeds = []string{}
		for _, name := range target.initial {
			includedSet[name] = true
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
						if _, alreadyIncluded := includedSet[name]; !alreadyIncluded {
							includedSet[name] = true
							nextCascadingSeeds = append(nextCascadingSeeds, name)
						}
					}
				}
			}
			cascadingSeeds = nextCascadingSeeds
		}

		for name := range includedSet {
			if !includes(target.initial, name) {
				target.dependencies = append(target.dependencies, name)
			}
		}

		sort.Strings(target.dependencies)
	}

	if extendAffected {
		// start from the explicitly targeted target
		includedSet = make(map[string]bool)
		cascadingSeeds = []string{}
		for _, name := range target.initial {
			includedSet[name] = true
			cascadingSeeds = append(cascadingSeeds, name)
		}

		for len(cascadingSeeds) > 0 {
			nextCascadingSeeds := []string{}
			for _, seed := range cascadingSeeds {
				for name, dependencies := range dependencyMap {
					if _, alreadyIncluded := includedSet[name]; !alreadyIncluded && cfg.Container(name).Exists() {
						if dependencies.includes(seed) {
							includedSet[name] = true
							nextCascadingSeeds = append(nextCascadingSeeds, name)
						}
					}
				}
			}
			cascadingSeeds = nextCascadingSeeds
		}

		for name := range includedSet {
			if !includes(target.initial, name) {
				target.affected = append(target.affected, name)
			}
		}

		sort.Strings(target.affected)
	}

	return
}

// includes checks whether the given needle is
// included in the target
func (t Target) includes(needle string) bool {
	for _, name := range t.all() {
		if name == needle {
			return true
		}
	}
	return false
}

// Return all targeted containers, sorted alphabetically
func (t Target) all() []string {
	all := t.initial
	for _, name := range t.dependencies {
		all = append(all, name)
	}
	for _, name := range t.affected {
		all = append(all, name)
	}
	sort.Strings(all)
	return all
}
