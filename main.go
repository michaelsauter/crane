package main

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"os"
	"os/exec"
	"strings"
)

var verbose bool
var force bool
var kill bool
var printNotice func(format string, a ...interface{})
var printError func(format string, a ...interface{})

func init() {
	printNotice = color.New(color.FgYellow).PrintfFunc()
	printError = color.New(color.FgRed).PrintfFunc()
}

func main() {
	// On panic, recover the error and display it
	defer func() {
		if err := recover(); err != nil {
			printError("ERROR: %s", err)
		}
	}()

	var cmdLift = &cobra.Command{
		Use:   "lift",
		Short: "Build or pull images, then run or start the containers",
		Long: `
provision will use specified Dockerfiles to build the images.
If no Dockerfile is given, it will pull the image from the index.
       `,
		Run: func(cmd *cobra.Command, args []string) {
			containers := readCranefile("Cranefile")
			containers.lift(force, kill)
		},
	}

	var cmdProvision = &cobra.Command{
		Use:   "provision",
		Short: "Build or pull images",
		Long: `
provision will use specified Dockerfiles to build the images.
If no Dockerfile is given, it will pull the image from the index.
        `,
		Run: func(cmd *cobra.Command, args []string) {
			containers := readCranefile("Cranefile")
			containers.provision(force)
		},
	}

	var cmdRun = &cobra.Command{
		Use:   "run",
		Short: "Run the containers",
		Long:  `run will call docker run on all containers.`,
		Run: func(cmd *cobra.Command, args []string) {
			containers := readCranefile("Cranefile")
			containers.run(force, kill)
		},
	}

	var cmdRm = &cobra.Command{
		Use:   "rm",
		Short: "Remove the containers",
		Long:  `rm will call docker rm on all containers.`,
		Run: func(cmd *cobra.Command, args []string) {
			containers := readCranefile("Cranefile")
			containers.rm(force, kill)
		},
	}

	var cmdKill = &cobra.Command{
		Use:   "kill",
		Short: "Kill the containers",
		Long:  `kill will call docker kill on all containers.`,
		Run: func(cmd *cobra.Command, args []string) {
			containers := readCranefile("Cranefile")
			containers.kill()
		},
	}

	var cmdStart = &cobra.Command{
		Use:   "start",
		Short: "Start the containers",
		Long:  `start will call docker start on all containers.`,
		Run: func(cmd *cobra.Command, args []string) {
			containers := readCranefile("Cranefile")
			containers.start()
		},
	}

	var cmdStop = &cobra.Command{
		Use:   "stop",
		Short: "Stop the containers",
		Long:  `stop will call docker stop on all containers.`,
		Run: func(cmd *cobra.Command, args []string) {
			containers := readCranefile("Cranefile")
			containers.stop()
		},
	}

	var cmdStatus = &cobra.Command{
		Use:   "status",
		Short: "Displays status of containers",
		Long:  `Displays the current status of containers referenced in the Cranefile.`,
		Run: func(cmd *cobra.Command, args []string) {
			containers := readCranefile("Cranefile")
			containers.status()
		},
	}

	var cmdVersion = &cobra.Command{
		Use:   "version",
		Short: "Display version",
		Long:  `Displays the version of Crane.`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("v0.5.0")
		},
	}

	var craneCmd = &cobra.Command{
		Use:   "crane",
		Short: "crane - Lift containers with ease",
		Long: `
Crane is a little tool to orchestrate Docker containers.
It works by reading in a Cranefile (a JSON file) which describes how to obtain container images and how to run them.
See the corresponding docker commands for more information.
		`,
	}

	craneCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
	cmdLift.Flags().BoolVarP(&force, "force", "f", false, "force")
	cmdLift.Flags().BoolVarP(&kill, "kill", "k", false, "kill containers")
	cmdProvision.Flags().BoolVarP(&force, "force", "f", false, "force")
	cmdRun.Flags().BoolVarP(&force, "force", "f", false, "force")
	cmdRun.Flags().BoolVarP(&kill, "kill", "k", false, "kill containers")
	cmdRm.Flags().BoolVarP(&force, "force", "f", false, "force")
	cmdRm.Flags().BoolVarP(&kill, "kill", "k", false, "kill containers")
	craneCmd.AddCommand(cmdLift, cmdProvision, cmdRun, cmdRm, cmdKill, cmdStart, cmdStop, cmdStatus, cmdVersion)
	craneCmd.Execute()
}

func executeCommand(name string, args []string) {
	if verbose {
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
	out, err := exec.Command(name, args...).Output()
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
