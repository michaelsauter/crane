package crane

import (
	"bytes"
	"fmt"
	"github.com/bjaglin/multiplexio"
	ansi "github.com/fatih/color"
	"io"
	"os"
	"strconv"
	"strings"
	"sync"
	"text/tabwriter"
)

type Containers []Container

func (containers Containers) Reversed() Containers {
	var reversed []Container
	for i := len(containers) - 1; i >= 0; i-- {
		reversed = append(reversed, containers[i])
	}
	return reversed
}

// Provision containers.
func (containers Containers) Provision(nocache bool, parallel int) {
	var (
		throttle = make(chan struct{}, parallel)
		wg       sync.WaitGroup
	)
	if parallel != 1 {
		// As we won't dump the result of provisioning until the first
		// image succeeds, give some early feedback so that the operator
		// does not feel it's stuck
		fmt.Println("Provisioning containers ...")
	}
	for _, container := range containers.stripProvisioningDuplicates() {
		if parallel > 0 {
			throttle <- struct{}{}
		}
		wg.Add(1)
		go func(container Container) {
			var out, err bytes.Buffer
			defer func() {
				out.WriteTo(os.Stdout)
				err.WriteTo(os.Stderr)
				handleRecoveredError(recover())
				if parallel > 0 {
					<-throttle
				}
				wg.Done()
				container.SetCommandsOutput(nil, nil)
			}()
			if parallel != 1 {
				// Prevent parallel provisioning output interlacing
				// by redirecting the outputs to buffers
				container.SetCommandsOutput(&out, &err)
			}
			container.Provision(nocache)
		}(container)
	}
	wg.Wait()
}

// Dump container logs.
func (containers Containers) Logs(follow bool, timestamps bool, tail string, colorize bool, since string) {
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
	counter := 0
	for _, container := range containers {
		var (
			logs        = container.Logs(follow, since, tail)
			stdoutColor *ansi.Color
			stderrColor *ansi.Color
		)
		for _, log := range logs {
			if colorize {
				// red has a negative/error connotation, so skip it
				ansiAttribute := ansi.Attribute(int(ansi.FgGreen) + counter%int(ansi.FgWhite-ansi.FgGreen))
				stdoutColor = ansi.New(ansiAttribute)
				// To synchronize their output, we need to multiplex stdout & stderr
				// onto the same stream. Unfortunately, that means that the user won't
				// be able to pipe them separately, so we use bold as a distinguishing
				// characteristic.
				stderrColor = ansi.New(ansiAttribute).Add(ansi.Bold)
			}
			appendSources(log.Stdout, stdoutColor, log.Name, "|")
			appendSources(log.Stderr, stderrColor, log.Name, "*")
			counter = counter + 1
		}
	}
	if len(sources) > 0 {
		aggregatedReader := multiplexio.NewReader(multiplexio.Options{}, sources...)
		io.Copy(os.Stdout, aggregatedReader)
	}
}

// Status of containers.
func (containers Containers) Status(notrunc bool) {
	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 0, 8, 1, '\t', 0)
	fmt.Fprintln(w, "NAME\tIMAGE\tID\tUP TO DATE\tIP\tPORTS\tRUNNING")
	for _, container := range containers {
		rows := container.Status()
		for _, fields := range rows {
			if !notrunc {
				fields[2] = truncateID(fields[2])
			}
			fmt.Fprintf(w, "%s\n", strings.Join(fields, "\t"))
		}
	}
	w.Flush()
}

// Return the length of the longest container name.
func (containers Containers) maxNameLength() (maxPrefixLength int) {
	for _, container := range containers {
		prefixLength := len(container.ActualName(false))
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
		key := container.BuildParams().Context() + "#" + container.Image()
		if _, ok := seenProvisioningKeys[key]; !ok {
			deduplicated = append(deduplicated, container)
			seenProvisioningKeys[key] = true
		}
	}
	return
}

func truncateID(id string) string {
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
