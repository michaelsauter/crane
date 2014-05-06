package main

import (
	"fmt"
	"github.com/spf13/cobra"
)

type Options struct {
	verbose  bool
	force    bool
	kill     bool
	config   string
	manifest string
	group    string
}

var options = Options{
	false,
	false,
	false,
	"",
	"",
	"",
}
var defaultManifests = []string{"crane.json", "crane.yaml", "crane.yml", "Cranefile"}

func manifestFiles() []string {
	var result = []string(nil)
	if len(options.manifest) > 0 {
		result = []string{options.manifest}
	} else {
		result = defaultManifests
	}
	return result
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
			containers.lift(options.force, options.kill)
		}),
	}

	var cmdProvision = &cobra.Command{
		Use:   "provision",
		Short: "Build or pull images",
		Long: `
provision will use specified Dockerfiles to build all the containers, or the specified one(s).
If no Dockerfile is given, it will pull the image(s) from the given registry.`,
		Run: containersCommand(func(containers Containers) {
			containers.provision(options.force)
		}),
	}

	var cmdRun = &cobra.Command{
		Use:   "run",
		Short: "Run the containers",
		Long:  `run will call docker run on all containers, or the specified one(s).`,
		Run: containersCommand(func(containers Containers) {
			containers.run(options.force, options.kill)
		}),
	}

	var cmdRm = &cobra.Command{
		Use:   "rm",
		Short: "Remove the containers",
		Long:  `rm will call docker rm on all containers, or the specified one(s).`,
		Run: containersCommand(func(containers Containers) {
			containers.rm(options.force, options.kill)
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
			fmt.Println("v0.7.1")
		},
	}

	var craneCmd = &cobra.Command{
		Use:   "crane",
		Short: "crane - Lift containers with ease",
		Long: `
Crane is a little tool to orchestrate Docker containers.
It works by reading in JSON or YAML (either from crane.json, crane.yaml, the string specified in --config, or a json or yml file specified by --manifest) which describes how to obtain container images and how to run them.
See the corresponding docker commands for more information.`,
	}

	craneCmd.PersistentFlags().BoolVarP(&options.verbose, "verbose", "v", false, "verbose output")
	craneCmd.PersistentFlags().StringVarP(&options.config, "config", "c", "", "config to read from")
	craneCmd.PersistentFlags().StringVarP(&options.manifest, "manifest", "m", "", "config file to read from")
	craneCmd.PersistentFlags().StringVarP(&options.group, "group", "g", "", "group or container to restrict the command to")
	cmdLift.Flags().BoolVarP(&options.force, "force", "f", false, "rebuild all images")
	cmdLift.Flags().BoolVarP(&options.kill, "kill", "k", false, "kill containers")
	cmdProvision.Flags().BoolVarP(&options.force, "force", "f", false, "rebuild all images")
	cmdRun.Flags().BoolVarP(&options.force, "force", "f", false, "stop and remove running containers first")
	cmdRun.Flags().BoolVarP(&options.kill, "kill", "k", false, "when using --force, kill containers instead of stopping them")
	cmdRm.Flags().BoolVarP(&options.force, "force", "f", false, "stop running containers first")
	cmdRm.Flags().BoolVarP(&options.kill, "kill", "k", false, "when using --force, kill containers instead of stopping them")
	craneCmd.AddCommand(cmdLift, cmdProvision, cmdRun, cmdRm, cmdKill, cmdStart, cmdStop, cmdStatus, cmdVersion)
	err := craneCmd.Execute()
	if err != nil {
		panic(StatusError{status: 64})
	}
}
