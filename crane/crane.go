package crane

import (
	"errors"
	"fmt"
	"github.com/flynn/go-shlex"
	"github.com/michaelsauter/crane/print"
	"io"
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

var requiredDockerVersion = []int{1, 3}

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

// Ensure there is a docker binary in the path,
// and printing an error if its version is below the minimal requirement.
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

	for i, expectedVersion := range requiredDockerVersion {
		if versions[i] > expectedVersion {
			break
		}
		if versions[i] < expectedVersion {
			print.Errorf("Unsupported client version! Please upgrade to Docker %v or later.\n", intJoin(requiredDockerVersion, "."))
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

func executeHook(hook string) {
	cmds, err := shlex.Split(hook)
	if err != nil {
		panic(StatusError{fmt.Errorf("Error when parsing hook `%v`: %v", hook, err), 64})
	}
	switch len(cmds) {
	case 0:
		return
	case 1:
		executeCommand(cmds[0], []string{})
	default:
		executeCommand(cmds[0], cmds[1:])
	}
}

func executeCommand(name string, args []string) {
	if isVerbose() {
		print.Infof("\n--> %s %s\n", name, strings.Join(args, " "))
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

func executeCommandBackground(name string, args []string) (stdout, stderr io.ReadCloser) {
	if isVerbose() {
		print.Infof("--> %s %s\n\n", name, strings.Join(args, " "))
	}
	cmd := exec.Command(name, args...)
	stdout, _ = cmd.StdoutPipe()
	stderr, _ = cmd.StderrPipe()
	cmd.Start()
	return stdout, stderr
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
