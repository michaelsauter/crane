package crane

import (
	"sort"
	"strings"
)

type Target []string

// NewTarget receives the specified target
// and determines which containers should be targeted.
// The target might be extended depending on the value
// given for cascadeDependencies and cascadeAffected.
// Additionally, the target is sorted alphabetically.
func NewTarget(targetFlag string) Target {

	containerMap := cfg.ContainerMap()

	targetParts := strings.Split(targetFlag, "+")
	target := targetParts[0]
	cascadeDependencies := "none"
	cascadeAffected := "none"
	for _, v := range targetParts[1:] {
		if v == "dependencies" {
			cascadeDependencies = "all"
		} else if v == "affected" {
			cascadeAffected = "all"
		} else {

		}
	}

	// start from the explicitly targeted target
	includedSet := make(map[string]bool)
	cascadingSeeds := []string{}
	for _, name := range cfg.ContainersForReference(target) {
		includedSet[name] = true
		cascadingSeeds = append(cascadingSeeds, name)
	}

	// Cascade until the graph has been fully traversed
	// according to the cascading flags.
	for len(cascadingSeeds) > 0 {
		nextCascadingSeeds := []string{}
		for _, seed := range cascadingSeeds {
			if cascadeDependencies != "none" {
				if dependencies, ok := dependencyGraph[seed]; ok {
					// Queue direct dependencies if we haven't already considered them
					for _, name := range dependencies.forKind(cascadeDependencies) {
						if _, alreadyIncluded := includedSet[name]; !alreadyIncluded {
							includedSet[name] = true
							nextCascadingSeeds = append(nextCascadingSeeds, name)
						}
					}
				}
			}
			if cascadeAffected != "none" {
				// Queue all containers we haven't considered yet which exist
				// and directly depend on the seed.
				for name, container := range containerMap {
					if _, alreadyIncluded := includedSet[name]; !alreadyIncluded {
						if container.Dependencies().includesAsKind(seed, cascadeAffected) && container.Exists() {
							includedSet[name] = true
							nextCascadingSeeds = append(nextCascadingSeeds, name)
						}
					}
				}
			}
		}
		cascadingSeeds = nextCascadingSeeds
	}

	// Keep the ones that are part of the container map
	included := []string{}
	for name := range includedSet {
		if _, exists := containerMap[name]; exists {
			included = append(included, name)
		}
	}

	// Sort alphabetically
	sortedTarget := Target(included)
	sort.Strings(sortedTarget)
	return sortedTarget
}

// includes checks whether the given needle is
// included in the target
func (t Target) includes(needle string) bool {
	for _, name := range t {
		if name == needle {
			return true
		}
	}
	return false
}
