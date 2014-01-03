package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"os/exec"
)

var verbose bool

func main() {
	// On panic, recover the error and display it
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("ERROR: ", err)
		}
	}()

	// Read the Cranefile in order to have
	// a root container to execute the commands.
	container := readCranefile("Cranefile")

	var cmdProvision = &cobra.Command{
		Use:   "provision",
		Short: "Build or pull images",
		Long: `
provision will use specified Dockerfiles to build the images.
If no Dockerfile is given, it will pull the image from the index.
        `,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Provisioning...")
			container.provision()
		},
	}

	var cmdRun = &cobra.Command{
		Use:   "run",
		Short: "Run the containers",
		Long:  `run will call docker run on all containers.`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Running...")
			// "Entry" container is attachable
			container.run(true)
		},
	}

	var cmdRm = &cobra.Command{
		Use:   "rm",
		Short: "Remove the containers",
		Long:  `rm will call docker rm on all containers.`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Removing...")
			container.rm()
		},
	}

	var cmdKill = &cobra.Command{
		Use:   "kill",
		Short: "Kill the containers",
		Long:  `kill will call docker kill on all containers.`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Killing...")
			container.kill()
		},
	}

	var cmdStart = &cobra.Command{
		Use:   "start",
		Short: "Start the containers",
		Long:  `start will call docker start on all containers.`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Starting...")
			container.start()
		},
	}

	var cmdStop = &cobra.Command{
		Use:   "stop",
		Short: "Stop the containers",
		Long:  `stop will call docker stop on all containers.`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Stopping...")
			container.stop()
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
	craneCmd.AddCommand(cmdProvision, cmdRun, cmdRm, cmdKill, cmdStart, cmdStop)
	craneCmd.Execute()
}

func executeCommand(name string, args []string) {
	if verbose {
		fmt.Printf("Calling docker with %v\n", args)
	}
	cmd := exec.Command("docker", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Run()
	if !cmd.ProcessState.Success() {
		panic(cmd.ProcessState.String()) // pass the error?
	}
}
