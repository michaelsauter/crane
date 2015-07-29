package crane

import (
	"fmt"
	"github.com/bjaglin/multiplexio"
	ansi "github.com/fatih/color"
	"github.com/michaelsauter/crane/print"
	"io"
	"os"
	"strconv"
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
func (containers Containers) lift(recreate bool, nocache bool, ignoreMissing string, configPath string) {
	containers.provisionOrSkip(recreate, nocache)
	containers.runOrStart(recreate, ignoreMissing, configPath)
}

// Provision containers.
func (containers Containers) provision(nocache bool) {
	for _, container := range containers.stripProvisioningDuplicates() {
		container.Provision(nocache)
	}
}

// Pull images.
func (containers Containers) pullImage() {
	for _, container := range containers.stripProvisioningDuplicates() {
		if len(container.Dockerfile()) == 0 {
			container.PullImage()
		}
	}
}

// Create containers.
// When recreate is true, removes existing containers first.
func (containers Containers) create(recreate bool, ignoreMissing string, configPath string) {
	if recreate {
		containers.rm(true)
	}
	for _, container := range containers {
		container.Create(ignoreMissing, configPath)
	}
}

// Run containers.
// When recreate is true, removes existing containers first.
func (containers Containers) run(recreate bool, ignoreMissing string, configPath string) {
	if recreate {
		containers.rm(true)
	}
	for _, container := range containers {
		container.Run(ignoreMissing, configPath)
	}
}

// Run or start containers.
// When recreate is true, removes existing containers first.
func (containers Containers) runOrStart(recreate bool, ignoreMissing string, configPath string) {
	if recreate {
		containers.rm(true)
	}
	for _, container := range containers {
		container.RunOrStart(ignoreMissing, configPath)
	}
}

// Provision or skip images.
// When update is true, provisions all images.
func (containers Containers) provisionOrSkip(update bool, nocache bool) {
	for _, container := range containers.stripProvisioningDuplicates() {
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

// Dump container logs.
func (containers Containers) logs(follow bool, timestamps bool, tail string, colorize bool) {
	var (
		sources         = make([]multiplexio.Source, 0, 2*len(containers))
		maxPrefixLength = strconv.Itoa(containers.maxNameLength())
	)
	appendSources := func(reader io.Reader, color *ansi.Color, name string, separator string) {
		if reader != nil {
			prefix := fmt.Sprintf("%"+maxPrefixLength+"s "+separator+" ", name)
			sources = append(sources, multiplexio.Source{
				Reader: reader,
				Write:  write(prefix, color, timestamps),
			})
		}
	}
	for i, container := range containers {
		var (
			stdout, stderr = container.Logs(follow, tail)
			stdoutColor    *ansi.Color
			stderrColor    *ansi.Color
		)
		if colorize {
			// red has a negative/error connotation, so skip it
			ansiAttribute := ansi.Attribute(int(ansi.FgGreen) + i%int(ansi.FgWhite-ansi.FgGreen))
			stdoutColor = ansi.New(ansiAttribute)
			// To synchronize their output, we need to multiplex stdout & stderr
			// onto the same stream. Unfortunately, that means that the user won't
			// be able to pipe them separately, so we use bold as a distinguishing
			// characteristic.
			stderrColor = ansi.New(ansiAttribute).Add(ansi.Bold)
		}
		appendSources(stdout, stdoutColor, container.Name(), "|")
		appendSources(stderr, stderrColor, container.Name(), "*")
	}
	if len(sources) > 0 {
		aggregatedReader := multiplexio.NewReader(multiplexio.Options{}, sources...)
		io.Copy(os.Stdout, aggregatedReader)
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

// Stats about containers.
func (containers Containers) stats() {
	args := []string{"stats"}
	for _, container := range containers {
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

// Return the length of the longest container name.
func (containers Containers) maxNameLength() (maxPrefixLength int) {
	for _, container := range containers {
		prefixLength := len(container.Name())
		if prefixLength > maxPrefixLength {
			maxPrefixLength = prefixLength
		}
	}
	return
}

// returns another list of containers, stripping out containers which
// would trigger some commands more than once for provisioning.
func (containers Containers) stripProvisioningDuplicates() (deduplicated Containers) {
	seenProvisioningKeys := make(map[string]bool)
	for _, container := range containers {
		// for 2 containers that would the same provisioning
		// commands, the key should be equal
		key := container.Dockerfile() + "#" + container.Image()
		if _, ok := seenProvisioningKeys[key]; !ok {
			deduplicated = append(deduplicated, container)
			seenProvisioningKeys[key] = true
		}
	}
	return
}

func truncateId(id string) string {
	shortLen := 12
	if len(id) < shortLen {
		shortLen = len(id)
	}
	return id[:shortLen]
}

// wraps an io.Writer, counting the number of bytes written
type countingWriter struct {
	io.Writer
	written int
}

func (c *countingWriter) Write(p []byte) (n int, err error) {
	n, err = c.Writer.Write(p)
	c.written += n
	return
}

// returns a function that will format and writes the line extracted from the logs of a given container
func write(prefix string, color *ansi.Color, timestamps bool) func(dest io.Writer, token []byte) (n int, err error) {
	return func(dest io.Writer, token []byte) (n int, err error) {
		countingWriter := countingWriter{Writer: dest}
		if color != nil {
			ansi.Output = &countingWriter
			color.Set()
		}
		_, err = countingWriter.Write([]byte(prefix))
		if err == nil {
			if !timestamps {
				// timestamps are always present in the incoming stream for
				// sorting purposes, so we strip them if the user didn't ask
				// for them
				const timestampPrefixLength = 31
				strip := timestampPrefixLength
				if string(token[0]) == "[" {
					// it seems that timestamps are wrapped in [] for events
					// streamed  in real time during a `docker logs -f`
					strip = strip + 2
				}
				token = token[strip:]
			}
			_, err = countingWriter.Write(token)
		}
		if err == nil {
			if color != nil {
				ansi.Unset()
			}
			_, err = dest.Write([]byte("\n"))
		}
		return countingWriter.written, err

	}
}
