package crane

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
)

type Containers []Container

func (containers Containers) names() []string {
	var names []string
	for _, container := range containers {
		names = append(names, container.Name())
	}
	return names
}

func (containers Containers) reversed() Containers {
	var reversed []Container
	for i := len(containers) - 1; i >= 0; i-- {
		reversed = append(reversed, containers[i])
	}
	return reversed
}

// Lift containers (provision + run).
// When recreate is set, this will re-provision all images
// and recreate all containers.
func (containers Containers) lift(recreate bool, nocache bool) {
	containers.provisionOrSkip(recreate, nocache)
	containers.runOrStart(recreate)
}

// Provision containers.
func (containers Containers) provision(nocache bool) {
	for _, container := range containers {
		container.Provision(nocache)
	}
}

// Run containers.
// When recreate is true, removes existing containers first.
func (containers Containers) run(recreate bool) {
	if recreate {
		containers.rm(true)
	}
	for _, container := range containers {
		container.Run()
	}
}

// Run or start containers.
// When recreate is true, removes existing containers first.
func (containers Containers) runOrStart(recreate bool) {
	if recreate {
		containers.rm(true)
	}
	for _, container := range containers {
		container.RunOrStart()
	}
}

// Provision or skip images.
// When update is true, provisions all images.
func (containers Containers) provisionOrSkip(update bool, nocache bool) {
	for _, container := range containers {
		container.ProvisionOrSkip(update, nocache)
	}
}

// Start containers.
func (containers Containers) start() {
	for _, container := range containers {
		container.Start()
	}
}

// Kill containers.
func (containers Containers) kill() {
	for _, container := range containers {
		container.Kill()
	}
}

// Stop containers.
func (containers Containers) stop() {
	for _, container := range containers {
		container.Stop()
	}
}

// Pause containers.
func (containers Containers) pause() {
	for _, container := range containers {
		container.Pause()
	}
}

// Unpause containers.
func (containers Containers) unpause() {
	for _, container := range containers {
		container.Unpause()
	}
}

// Remove containers.
// When force is true, kills existing containers first.
func (containers Containers) rm(force bool) {
	for _, container := range containers {
		container.Rm(force)
	}
}

// Push containers.
func (containers Containers) push() {
	for _, container := range containers {
		container.Push()
	}
}

// Status of containers.
func (containers Containers) status(notrunc bool) {
	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 0, 8, 1, '\t', 0)
	fmt.Fprintln(w, "NAME\tIMAGE\tID\tUP TO DATE\tIP\tPORTS\tRUNNING")
	for _, container := range containers {
		fields := container.Status()
		if !notrunc {
			fields[2] = truncateId(fields[2])
		}
		fmt.Fprintf(w, "%s\n", strings.Join(fields, "\t"))
	}
	w.Flush()
}

func truncateId(id string) string {
	shortLen := 12
	if len(id) < shortLen {
		shortLen = len(id)
	}
	return id[:shortLen]
}
