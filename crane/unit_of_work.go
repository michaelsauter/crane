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
		targeted:   targeted,
		containers: targeted,
		order:      []string{},
		mustRun:    []string{},
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

	return
}

func (uow *UnitOfWork) Run() {
	for _, container := range uow.containers() {
		if includes(uow.targeted, container.Name()) {
			container.Run("none", cfg.Path())
		} else if includes(uow.mustRun, container.Name()) {
			container.Start("none", cfg.Path())
		}
	}
}

func (uow *UnitOfWork) Lift(noCache bool) {
	for _, container := range uow.containers() {
		if includes(uow.targeted, container.Name()) {
			container.Lift(noCache, "none", cfg.Path())
		} else if includes(uow.mustRun, container.Name()) {
			container.Start("none", cfg.Path())
		}
	}
}

func (uow *UnitOfWork) Stats() {
	args := []string{"stats"}
	for _, container := range uow.targeted() {
		if container.Running() {
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
	uow.targeted().Status(noTrunc)
}

// Push containers.
func (uow *UnitOfWork) Push() {
	for _, container := range uow.targeted() {
		container.Push()
	}
}

// Unpause containers.
func (uow *UnitOfWork) Unpause() {
	for _, container := range uow.targeted() {
		container.Unpause()
	}
}

// Pause containers.
func (uow *UnitOfWork) Pause() {
	for _, container := range uow.targeted().Reversed() {
		container.Pause()
	}
}

// Start containers.
func (uow *UnitOfWork) Start() {
	for _, container := range uow.containers() {
		if includes(uow.targeted, container.Name()) {
			container.Start()
		} else if includes(uow.mustRun, container.Name()) {
			container.Start("none", cfg.Path())
		}
	}
}

// Stop containers.
func (uow *UnitOfWork) Stop() {
	for _, container := range uow.targeted().Reversed() {
		container.Stop()
	}
}

// Kill containers.
func (uow *UnitOfWork) Kill() {
	for _, container := range uow.targeted().Reversed() {
		container.Kill()
	}
}

// Rm containers.
func (uow *UnitOfWork) Rm() {
	for _, container := range uow.targeted().Reversed() {
		container.Rm()
	}
}

// Create containers.
func (uow *UnitOfWork) Create() {
	for _, container := range uow.containers() {
		if includes(uow.targeted, container.Name()) {
			container.Create("none", cfg.Path())
		} else if includes(uow.mustRun, container.Name()) {
			container.Start("none", cfg.Path())
		}
	}
}

// Provision containers.
func (uow *UnitOfWork) Provision(noCache bool) {
	uow.targeted().Provision(noCache)
}

// Pull containers.
func (uow *UnitOfWork) PullImage() {
	for _, container := range uow.targeted() {
		if len(container.Dockerfile()) == 0 {
			container.PullImage()
		}
	}
}

// Log containers.
func (uow *UnitOfWork) Logs(follow bool, timestamps bool, tail bool, colorize bool) {
	uow.targeted().Logs(follow, timestamps, tail, colorize)
}

func (uow *UnitOfWork) containers() Containers {
	c = []Container{}
	for _, name := range uow.order {
		c = append(c, cfg.Container(name))
	}
	c
}

func (uow *UnitOfWork) targeted() Containers {
	c = []Container{}
	for _, name := range uow.order {
		if includes(uow.targeted, name) {
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
