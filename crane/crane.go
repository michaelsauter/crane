package crane

import (
	"errors"
	"fmt"
	"github.com/michaelsauter/crane/print"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"syscall"
)

type StatusError struct {
	error  error
	status int
}

var (
	minimalDockerVersion     = []int{1, 0}
	recommendedDockerVersion = []int{1, 2}
)

func RealMain() {
	// On panic, recover the error, display it and return the given status code if any
	defer func() {
		var statusError StatusError

		switch err := recover().(type) {
		case StatusError:
			statusError = err
		case error:
			statusError = StatusError{err, 1}
		case string:
			statusError = StatusError{errors.New(err), 1}
		default:
			statusError = StatusError{}
		}

		if statusError.error != nil {
			print.Errorf("ERROR: %s\n", statusError.error)
		}
		os.Exit(statusError.status)
	}()
	checkDockerClient()
	handleCmd()
}

// Ensure there is a docker binary in the path, printing a warning if its version
// is below the minimal requirement, and printing a warning if its version
// is below the recommended requirement.
func checkDockerClient() {
	output, err := commandOutput("docker", []string{"--version"})
	if err != nil {
		panic(StatusError{errors.New("Error when probing Docker's client version. Is docker installed and within the $PATH?"), 69})
	}
	re := regexp.MustCompile("([0-9]+)\\.([0-9]+)\\.?([0-9]+)?")
	rawVersions := re.FindStringSubmatch(string(output))
	var versions []int
	for _, rawVersion := range rawVersions[1:] {
		version, err := strconv.Atoi(rawVersion)
		if err != nil {
			print.Errorf("Error when parsing Docker's version %v: %v", rawVersion, err)
			break
		}
		versions = append(versions, version)
	}

	for i, expectedVersion := range minimalDockerVersion {
		if versions[i] > expectedVersion {
			break
		}
		if versions[i] < expectedVersion {
			print.Errorf("Unsupported client version. Please upgrade to Docker %v or later.", intJoin(recommendedDockerVersion, "."))
		}
	}

	for i, expectedVersion := range recommendedDockerVersion {
		if versions[i] > expectedVersion {
			break
		}
		if versions[i] < expectedVersion {
			print.Noticef("WARNING: outdated Docker client, behavior might not be optimal. Please upgrade to Docker %v or later.\n", intJoin(recommendedDockerVersion, "."))
			break
		}
	}
}

// Similar to strings.Join() for int slices.
func intJoin(intSlice []int, sep string) string {
	var stringSlice []string
	for _, v := range intSlice {
		stringSlice = append(stringSlice, fmt.Sprint(v))
	}
	return strings.Join(stringSlice, ".")
}

func executeCommand(name string, args []string) {
	if isVerbose() {
		fmt.Printf("\n--> %s %s\n", name, strings.Join(args, " "))
	}
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Run()
	if !cmd.ProcessState.Success() {
		status := cmd.ProcessState.Sys().(syscall.WaitStatus).ExitStatus()
		panic(StatusError{errors.New(cmd.ProcessState.String()), status})
	}
}

func commandOutput(name string, args []string) (string, error) {
	out, err := exec.Command(name, args...).CombinedOutput()
	return strings.TrimSpace(string(out)), err
}

// From https://gist.github.com/dagoof/1477401
func pipedCommandOutput(pipedCommandArgs ...[]string) ([]byte, error) {
	var commands []exec.Cmd
	for _, commandArgs := range pipedCommandArgs {
		cmd := exec.Command(commandArgs[0], commandArgs[1:]...)
		commands = append(commands, *cmd)
	}
	for i, command := range commands[:len(commands)-1] {
		out, err := command.StdoutPipe()
		if err != nil {
			return nil, err
		}
		command.Start()
		commands[i+1].Stdin = out
	}
	final, err := commands[len(commands)-1].Output()
	if err != nil {
		return nil, err
	}
	return final, nil
}
