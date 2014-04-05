package main

import (
	"fmt"
	"github.com/michaelsauter/crane/print"
	"os"
	"os/exec"
	"strings"
)

func main() {
	// On panic, recover the error and display it
	defer func() {
		if err := recover(); err != nil {
			print.Error("ERROR: %s\n", err)
		}
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
		panic(cmd.ProcessState.String()) // pass the error?
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
