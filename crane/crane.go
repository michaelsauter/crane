package crane

import (
	"errors"
	"fmt"
	"github.com/michaelsauter/crane/print"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

type StatusError struct {
	error  error
	status int
}

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

	handleCmd()
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

// from https://gist.github.com/dagoof/1477401
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
