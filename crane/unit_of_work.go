package crane

import (
	"fmt"
	"os"
	"strings"
	"text/template"
)

type UnitOfWork struct {
	targeted       []string
	containers     []string
	order          []string
	requireStarted []string
}

func NewUnitOfWork(dependencyMap map[string]*Dependencies, targeted []string) (uow *UnitOfWork, err error) {

	uow = &UnitOfWork{
		targeted:       targeted,
		containers:     targeted,
		order:          []string{},
		requireStarted: []string{},
	}

	// select all containers which we care about
	for {
		c := uow.containers
		initialLenContainers := len(c)
		for _, name := range c {
			dependencies := dependencyMap[name]
			if dependencies == nil {
				err = fmt.Errorf("Container %s referenced, but not defined.", name)
				return
			}
			for _, dep := range dependencies.All {
				uow.ensureInContainers(dep)
				if dependencies.requireStarted(dep) {
					uow.ensureInRequireStarted(dep)
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
			if dependencies, ok := dependencyMap[name]; ok {
				if dependencies.satisfied() {
					uow.order = append(uow.order, name)
					delete(dependencyMap, name)
					for _, dependencies := range dependencyMap {
						dependencies.remove(name)
					}
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
		err = fmt.Errorf("Command cannot be applied to any container.")
	}

	return
}

func (uow *UnitOfWork) Run(cmds []string, excluded []string) {
	uow.prepareRequirements()
	for _, container := range uow.Containers() {
		if includes(uow.targeted, container.Name()) {
			container.Run(cmds, excluded)
		} else if includes(uow.requireStarted, container.Name()) || !container.Exists() {
			container.Start(excluded)
		}
	}
}

func (uow *UnitOfWork) Lift(cmds []string, excluded []string, noCache bool, parallel int) {
	uow.Targeted().Provision(noCache, parallel)
	uow.prepareRequirements()
	for _, container := range uow.Containers() {
		if includes(uow.targeted, container.Name()) {
			container.Run(cmds, excluded)
		} else if includes(uow.requireStarted, container.Name()) || !container.Exists() {
			container.Start(excluded)
		}
	}
}

func (uow *UnitOfWork) Stats() {
	args := []string{"stats"}
	for _, container := range uow.Targeted() {
		for _, name := range container.InstancesOfStatus("running") {
			args = append(args, name)
		}
	}
	if len(args) > 1 {
		executeCommand("docker", args, os.Stdout, os.Stderr)
	} else {
		printNoticef("None of the targeted container is running.\n")
	}
}

func (uow *UnitOfWork) Status(noTrunc bool) {
	uow.Targeted().Status(noTrunc)
}

// Push containers.
func (uow *UnitOfWork) Push() {
	for _, container := range uow.Targeted() {
		container.Push()
	}
}

// Unpause containers.
func (uow *UnitOfWork) Unpause() {
	for _, container := range uow.Targeted() {
		container.Unpause()
	}
}

// Pause containers.
func (uow *UnitOfWork) Pause() {
	for _, container := range uow.Targeted().Reversed() {
		container.Pause()
	}
}

// Start containers.
func (uow *UnitOfWork) Start() {
	uow.prepareRequirements()
	for _, container := range uow.Containers() {
		if includes(uow.targeted, container.Name()) {
			container.Start(excluded)
		} else if includes(uow.requireStarted, container.Name()) || !container.Exists() {
			container.Start(excluded)
		}
	}
}

// Stop containers.
func (uow *UnitOfWork) Stop() {
	for _, container := range uow.Targeted().Reversed() {
		container.Stop()
	}
}

// Kill containers.
func (uow *UnitOfWork) Kill() {
	for _, container := range uow.Targeted().Reversed() {
		container.Kill()
	}
}

func (uow *UnitOfWork) Exec(cmds []string) {
	for _, container := range uow.Containers() {
		if includes(uow.targeted, container.Name()) {
			container.Exec(cmds)
		} else if includes(uow.requireStarted, container.Name()) || !container.Exists() {
			container.Start(excluded)
		}
	}
}

// Rm containers.
func (uow *UnitOfWork) Rm(force bool) {
	for _, container := range uow.Targeted().Reversed() {
		container.Rm(force)
	}
}

// Create containers.
func (uow *UnitOfWork) Create(cmds []string, excluded []string) {
	uow.prepareRequirements()
	for _, container := range uow.Containers() {
		if includes(uow.targeted, container.Name()) {
			container.Create(cmds, excluded)
		} else if includes(uow.requireStarted, container.Name()) || !container.Exists() {
			container.Start(excluded)
		}
	}
}

// Provision containers.
func (uow *UnitOfWork) Provision(noCache bool, parallel int) {
	uow.Targeted().Provision(noCache, parallel)
}

// Pull containers.
func (uow *UnitOfWork) PullImage() {
	for _, container := range uow.Targeted() {
		if len(container.BuildParams().Context()) == 0 {
			container.PullImage()
		}
	}
}

// Log containers.
func (uow *UnitOfWork) Logs(follow bool, timestamps bool, tail string, colorize bool, since string) {
	uow.Targeted().Logs(follow, timestamps, tail, colorize, since)
}

// Generate files.
func (uow *UnitOfWork) Generate(templateFile string, output string) {
	templateFileParts := strings.Split(templateFile, "/")
	templateName := templateFileParts[len(templateFileParts)-1]

	tmpl, err := template.New(templateName).ParseFiles(templateFile)
	if err != nil {
		printErrorf("ERROR: %s\n", err)
		return
	}

	executeTemplate := func(outputFile string, templateInfo interface{}) {
		writer := os.Stdout
		if len(outputFile) > 0 {
			writer, _ = os.Create(outputFile)
		}
		err = tmpl.Execute(writer, templateInfo)
		if err != nil {
			printErrorf("ERROR: %s\n", err)
		}
	}

	if strings.Contains(output, "%s") {
		for _, container := range uow.TargetedInfo() {
			executeTemplate(fmt.Sprintf(output, container.PrefixedName()), container)
		}
	} else {
		tmplInfo := struct {
			Containers []ContainerInfo
		}{
			Containers: uow.TargetedInfo(),
		}
		executeTemplate(output, tmplInfo)
	}
}

func (uow *UnitOfWork) Containers() Containers {
	c := []Container{}
	for _, name := range uow.order {
		c = append(c, cfg.Container(name))
	}
	return c
}

func (uow *UnitOfWork) Targeted() Containers {
	c := []Container{}
	for _, name := range uow.order {
		if includes(uow.targeted, name) {
			c = append(c, cfg.Container(name))
		}
	}
	return c
}

func (uow *UnitOfWork) TargetedInfo() []ContainerInfo {
	c := []ContainerInfo{}
	for _, name := range uow.order {
		if includes(uow.targeted, name) {
			c = append([]ContainerInfo{cfg.ContainerInfo(name)}, c...)
		}
	}
	return c
}

func (uow *UnitOfWork) Associated() []string {
	c := []string{}
	for _, name := range uow.order {
		if !includes(uow.targeted, name) {
			c = append(c, name)
		}
	}
	return c
}

func (uow *UnitOfWork) RequiredNetworks() []string {
	required := []string{}
	networks := cfg.NetworkNames()
	if len(networks) == 0 {
		return required
	}
	for _, container := range uow.Containers() {
		net := container.RunParams().Net()
		if includes(networks, net) && !includes(required, net) {
			required = append(required, net)
		}
	}
	return required
}

func (uow *UnitOfWork) RequiredVolumes() []string {
	required := []string{}
	volumes := cfg.VolumeNames()
	if len(volumes) == 0 {
		return required
	}
	for _, container := range uow.Containers() {
		for _, volumeSource := range container.RunParams().VolumeSources() {
			if includes(volumes, volumeSource) && !includes(required, volumeSource) {
				required = append(required, volumeSource)
			}
		}
	}
	return required
}

func (uow *UnitOfWork) ensureInContainers(name string) {
	if !includes(uow.containers, name) {
		uow.containers = append(uow.containers, name)
	}
}

func (uow *UnitOfWork) ensureInRequireStarted(name string) {
	if !includes(uow.requireStarted, name) {
		uow.requireStarted = append(uow.requireStarted, name)
	}
}

func (uow *UnitOfWork) prepareRequirements() {
	uow.prepareNetworks()
	uow.prepareVolumes()
}

func (uow *UnitOfWork) prepareNetworks() {
	for _, n := range uow.RequiredNetworks() {
		net := cfg.Network(n)
		if !net.Exists() {
			net.Create()
		}
	}
}

func (uow *UnitOfWork) prepareVolumes() {
	for _, v := range uow.RequiredVolumes() {
		vol := cfg.Volume(v)
		if !vol.Exists() {
			vol.Create()
		}
	}
}
