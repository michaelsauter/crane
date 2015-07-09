package crane

import (
	"fmt"
)

type UnitOfWork struct {
	targeted   []string
	containers []string
	order      []string
	mustRun    []string
}

func includes(haystack []string, needle string) bool {
	for _, name := range haystack {
		if name == needle {
			return true
		}
	}
	return false
}

func NewUnitOfWork(graph DependencyGraph, targeted []string) (unitOfWork UnitOfWork, err error) {

	unitOfWork = UnitOfWork{
		targeted:   targeted,
		containers: targeted,
		order:      []string{},
		mustRun:    []string{},
	}

	// select all containers which we care about
	for {
		initialLenContainers := len(unitOfWork.containers)
		for _, name := range unitOfWork.containers {
			dependencies := graph[name]
			for _, dep := range dependencies.All {
				unitOfWork.ensureInContainers(dep)
				if dependencies.mustRun(dep) {
					unitOfWork.ensureInMustRun(dep)
				}
			}
		}
		if len(unitOfWork.containers) == initialLenContainers {
			break
		}
	}

	// bring containers into order
	for {
		initialLenOrdered := len(unitOfWork.order)
		for _, name := range unitOfWork.containers {
			if dependencies, ok := graph[name]; ok {
				if dependencies.satisfied() {
					unitOfWork.order = append([]string{name}, unitOfWork.order...)
					graph.resolve(name)
				}
			}
		}
		if len(unitOfWork.order) == initialLenOrdered {
			err = fmt.Errorf("Dependencies for container(s) %s could not be resolved.", unitOfWork.targeted)
			return
		}
	}

	if len(unitOfWork.containers) == 0 {
		err = fmt.Errorf("ERROR: Command cannot be applied to any container.")
	}

	return
}

func (uow UnitOfWork) ensureInContainers(name string) {
	if !includes(uow.containers, name) {
		uow.containers = append(uow.containers, name)
	}
}

func (uow UnitOfWork) ensureInMustRun(name string) {
	if !includes(uow.mustRun, name) {
		uow.mustRun = append(uow.mustRun, name)
	}
}
