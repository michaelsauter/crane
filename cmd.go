package main

import (
	"fmt"
	"github.com/spf13/cobra"
)

type Options struct {
	verbose  bool
	recreate bool
	nocache  bool
	kill     bool
	config   string
	target   string
}

var options = Options{
	verbose:  false,
	recreate: false,
	nocache:  false,
	kill:     false,
	config:   "",
	target:   "",
}
var defaultConfigFiles = []string{"crane.json", "crane.yaml", "crane.yml", "Cranefile"}

func configFiles() []string {
	if len(options.config) > 0 {
		return []string{options.config}
	} else {
		return defaultConfigFiles
	}
}

func isVerbose() bool {
	return options.verbose
}

// returns a function to be set as a cobra command run, wrapping a command meant to be run on a set of containers
func containersCommand(wrapped func(containers Containers)) func(cmd *cobra.Command, args []string) {
	return func(cmd *cobra.Command, args []string) {
		if len(args) > 0 {
			cmd.Printf("Error: too many arguments given: %#q", args)
			cmd.Usage()
			panic(StatusError{status: 64})
		}
		containers := getContainers(options)
		wrapped(containers)
	}
}

func handleCmd() {

	var cmdLift = &cobra.Command{
		Use:   "lift",
		Short: "Build or pull images, then run or start the containers",
		Long: `
lift will use specified Dockerfiles to build all the containers, or the specified one(s).
If no Dockerfile is given, it will pull the image(s) from the given registry.`,
		Run: containersCommand(func(containers Containers) {
			containers.lift(options.recreate, options.nocache)
		}),
	}

	var cmdProvision = &cobra.Command{
		Use:   "provision",
		Short: "Build or pull images",
		Long: `
provision will use specified Dockerfiles to build all the containers, or the specified one(s).
If no Dockerfile is given, it will pull the image(s) from the given registry.`,
		Run: containersCommand(func(containers Containers) {
			containers.provision(options.nocache)
		}),
	}

	var cmdRun = &cobra.Command{
		Use:   "run",
		Short: "Run the containers",
		Long:  `run will call docker run on all containers, or the specified one(s).`,
		Run: containersCommand(func(containers Containers) {
			containers.run(options.recreate)
		}),
	}

	var cmdRm = &cobra.Command{
		Use:   "rm",
		Short: "Remove the containers",
		Long:  `rm will call docker rm on all containers, or the specified one(s).`,
		Run: containersCommand(func(containers Containers) {
			containers.rm(options.kill)
		}),
	}

	var cmdKill = &cobra.Command{
		Use:   "kill",
		Short: "Kill the containers",
		Long:  `kill will call docker kill on all containers, or the specified one(s).`,
		Run: containersCommand(func(containers Containers) {
			containers.kill()
		}),
	}

	var cmdStart = &cobra.Command{
		Use:   "start",
		Short: "Start the containers",
		Long:  `start will call docker start on all containers, or the specified one(s).`,
		Run: containersCommand(func(containers Containers) {
			containers.start()
		}),
	}

	var cmdStop = &cobra.Command{
		Use:   "stop",
		Short: "Stop the containers",
		Long:  `stop will call docker stop on all containers, or the specified one(s).`,
		Run: containersCommand(func(containers Containers) {
			containers.stop()
		}),
	}

	var cmdPush = &cobra.Command{
		Use:   "push",
		Short: "Push the containers",
		Long:  `push will call docker push on all containers, or the specified one(s).`,
		Run: containersCommand(func(containers Containers) {
			containers.push()
		}),
	}

	var cmdStatus = &cobra.Command{
		Use:   "status",
		Short: "Displays status of containers",
		Long:  `Displays the current status of all the containers, or the specified one(s).`,
		Run: containersCommand(func(containers Containers) {
			containers.status()
		}),
	}

	var cmdVersion = &cobra.Command{
		Use:   "version",
		Short: "Display version",
		Long:  `Displays the version of Crane.`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("v0.7.3")
		},
	}

	var craneCmd = &cobra.Command{
		Use:   "crane",
		Short: "crane - Lift containers with ease",
		Long: `
Crane is a little tool to orchestrate Docker containers.
It works by reading in JSON or YAML which describes how to obtain container images and how to run them.
See the corresponding docker commands for more information.`,
	}

	craneCmd.PersistentFlags().BoolVarP(&options.verbose, "verbose", "v", false, "verbose output")
	craneCmd.PersistentFlags().StringVarP(&options.config, "config", "c", "", "config file to read from")
	craneCmd.PersistentFlags().StringVarP(&options.target, "target", "t", "", "group or container to execute the command for")

	cmdLift.Flags().BoolVarP(&options.recreate, "recreate", "r", false, "recreate containers (kill and remove containers, provision images, run containers)")
	cmdLift.Flags().BoolVarP(&options.nocache, "nocache", "n", false, "Build the image without any cache")

	cmdProvision.Flags().BoolVarP(&options.nocache, "nocache", "n", false, "Build the image without any cache")

	cmdRun.Flags().BoolVarP(&options.recreate, "recreate", "r", false, "recreate containers (kill and remove containers first)")

	cmdRm.Flags().BoolVarP(&options.kill, "kill", "k", false, "kill containers if they are running first")

	craneCmd.AddCommand(cmdLift, cmdProvision, cmdRun, cmdRm, cmdKill, cmdStart, cmdStop, cmdPush, cmdStatus, cmdVersion)
	err := craneCmd.Execute()
	if err != nil {
		panic(StatusError{status: 64})
	}
}
