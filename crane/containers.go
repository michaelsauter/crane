package crane

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
)

type Containers []Container

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
		container.provision(nocache)
	}
}

// Run containers.
// When recreate is true, removes existing containers first.
func (containers Containers) run(recreate bool) {
	if recreate {
		containers.rm(true)
	}
	for _, container := range containers {
		container.run()
	}
}

// Run or start containers.
// When recreate is true, removes existing containers first.
func (containers Containers) runOrStart(recreate bool) {
	if recreate {
		containers.rm(true)
	}
	for _, container := range containers {
		container.runOrStart()
	}
}

// Provision or skip images.
// When update is true, provisions all images.
func (containers Containers) provisionOrSkip(update bool, nocache bool) {
	for _, container := range containers {
		container.provisionOrSkip(update, nocache)
	}
}

// Start containers.
func (containers Containers) start() {
	for _, container := range containers {
		container.start()
	}
}

// Kill containers.
func (containers Containers) kill() {
	for _, container := range containers {
		container.kill()
	}
}

// Stop containers.
func (containers Containers) stop() {
	for _, container := range containers {
		container.stop()
	}
}

// Pause containers.
func (containers Containers) pause() {
	for _, container := range containers {
		container.pause()
	}
}

// Unpause containers.
func (containers Containers) unpause() {
	for _, container := range containers {
		container.unpause()
	}
}

// Remove containers.
// When kill is true, kills existing containers first.
func (containers Containers) rm(kill bool) {
	if kill {
		containers.kill()
	}
	for _, container := range containers {
		container.rm()
	}
}

// Push containers.
func (containers Containers) push() {
	for _, container := range containers {
		container.push()
	}
}

// Status of containers.
func (containers Containers) status(notrunc bool) {
	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 0, 8, 1, '\t', 0)
	fmt.Fprintln(w, "NAME\tIMAGE\tID\tIP\tPORTS\tRUNNING")
	for _, container := range containers {
		fields := container.status(notrunc)
		fmt.Fprintf(w, "%s\n", strings.Join(fields, "\t"))
	}
	w.Flush()
}
