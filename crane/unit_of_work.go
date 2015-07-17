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

func NewUnitOfWork(graph DependencyGraph, targeted []string) (uow *UnitOfWork, err error) {

	uow = &UnitOfWork{
		targeted:     targeted,
		containers:   targeted,
		order:        []string{},
		reverseOrder: []string{},
		mustRun:      []string{},
	}

	// select all containers which we care about
	for {
		c := uow.containers
		initialLenContainers := len(c)
		for _, name := range c {
			dependencies := graph[name]
			for _, dep := range dependencies.All {
				uow.ensureInContainers(dep)
				if dependencies.mustRun(dep) {
					uow.ensureInMustRun(dep)
				}
			}
		}
		if len(uow.containers) == initialLenContainers {
			break
		}
	}

	// bring containers into order
	for {
		initialLenOrdered := len(uow.order)
		for _, name := range uow.containers {
			if dependencies, ok := graph[name]; ok {
				if dependencies.satisfied() {
					uow.order = append(uow.order, name)
					graph.resolve(name)
				}
			}
		}
		if len(uow.order) == initialLenOrdered {
			break
		}
	}

	if len(uow.order) < len(uow.containers) {
		err = fmt.Errorf("Dependencies for container(s) %s could not be resolved.", uow.targeted)
	} else if len(uow.containers) == 0 {
		err = fmt.Errorf("ERROR: Command cannot be applied to any container.")
	}

	for _, name := range uow.order {
		uow.reverseOrder = append([]string{name}, uow.reverseOrder...)
	}

	return
}

func (uow *UnitOfWork) orderedTargetedContainers() Containers {
	c = []Container{}
	for _, name := range uow.order {
		if !includes(uow.targeted, name) {
			c = append(c, cfg.Container(name))
		}
	}
	c
}

func (uow *UnitOfWork) ensureInContainers(name string) {
	if !includes(uow.containers, name) {
		uow.containers = append(uow.containers, name)
	}
}

func (uow *UnitOfWork) ensureInMustRun(name string) {
	if !includes(uow.mustRun, name) {
		uow.mustRun = append(uow.mustRun, name)
	}
}

func (uow *UnitOfWork) Run() {
	for _, name := range uow.order {
		if includes(uow.targeted, name) {
			cfg.Container(name).Run("none", cfg.Path())
		} else if includes(uow.mustRun, name) {
			cfg.Container(name).Start("none", cfg.Path())
		}
	}
}

func (uow *UnitOfWork) Lift(noCache bool) {
	for _, name := range uow.order {
		if includes(uow.targeted, name) {
			cfg.Container(name).Lift(noCache, "none", cfg.Path())
		} else if includes(uow.mustRun, name) {
			cfg.Container(name).Start("none", cfg.Path())
		}
	}
}

func (uow *UnitOfWork) Stats() {
	args := []string{"stats"}
	for _, name := range uow.order {
		container := cfg.Container(name)
		if includes(uow.targeted, name) && container.Running() {
			args = append(args, container.Name())
		}
	}
	if len(args) > 1 {
		executeCommand("docker", args)
	} else {
		print.Errorf("None of the targeted container is running.\n")
	}
}

func (uow *UnitOfWork) Status(noTrunc bool) {
	uow.orderedTargetedContainers().status(noTrunc)
}

// Push containers.
func (uow *UnitOfWork) Push() {
	for _, name := range uow.order {
		if includes(uow.targeted, name) {
			cfg.Container(name).Push()
		}
	}
}

// Unpause containers.
func (uow *UnitOfWork) Unpause() {
	for _, name := range uow.order {
		if includes(uow.targeted, name) {
			cfg.Container(name).Unpause()
		}
	}
}

// Pause containers.
func (uow *UnitOfWork) Pause() {
	for _, name := range uow.reverseOrder {
		if includes(uow.targeted, name) {
			cfg.Container(name).Pause()
		}
	}
}

// Start containers.
func (uow *UnitOfWork) Start() {
	for _, name := range uow.order {
		if includes(uow.targeted, name) {
			cfg.Container(name).Start()
		} else if includes(uow.mustRun, name) {
			cfg.Container(name).Start("none", cfg.Path())
		}
	}
}

// Stop containers.
func (uow *UnitOfWork) Stop() {
	for _, name := range uow.reverseOrder {
		if includes(uow.targeted, name) {
			cfg.Container(name).Stop()
		}
	}
}

// Kill containers.
func (uow *UnitOfWork) Kill() {
	for _, name := range uow.reverseOrder {
		if includes(uow.targeted, name) {
			cfg.Container(name).Kill()
		}
	}
}

// Rm containers.
func (uow *UnitOfWork) Rm() {
	for _, name := range uow.reverseOrder {
		if includes(uow.targeted, name) {
			cfg.Container(name).Rm()
		}
	}
}

// Create containers.
func (uow *UnitOfWork) Create() {
	for _, name := range uow.order {
		if includes(uow.targeted, name) {
			cfg.Container(name).Create("none", cfg.Path())
		} else if includes(uow.mustRun, name) {
			cfg.Container(name).Start("none", cfg.Path())
		}
	}
}

// Provision containers.
func (uow *UnitOfWork) Provision(noCache bool) {
	uow.orderedTargetedContainers().provision(noCache)
}

// Pull containers.
func (uow *UnitOfWork) PullImage() {
	for _, name := range uow.order {
		container := cfg.Container(name)
		if includes(uow.targeted, name) && len(container.Dockerfile()) == 0 {
			container.PullImage()
		}
	}
}

// Log containers.
func (uow *UnitOfWork) Logs(follow bool, timestamps bool, tail bool, colorize bool) {
	uow.orderedTargetedContainers().logs(follow, timestamps, tail, colorize)
}
