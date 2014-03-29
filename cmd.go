package main

import (
	"github.com/spf13/cobra"
	"fmt"
)

type Options struct {
	verbose bool
	force bool
	kill bool
	config string
	manifest string
}
var options = Options{
	false,
	false,
	false,
	"",
	"",
}
var defaultManifests = []string{"crane.json","crane.yaml","crane.yml","Cranefile"}

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

func handleCmd() {
	var cmdLift = &cobra.Command{
		Use:   "lift",
		Short: "Build or pull images, then run or start the containers",
		Long: `
provision will use specified Dockerfiles to build the images.
If no Dockerfile is given, it will pull the image from the index.`,
		Run: func(cmd *cobra.Command, args []string) {
			containers := getContainers(options)
			containers.lift(options.force, options.kill)
		},
	}

	var cmdProvision = &cobra.Command{
		Use:   "provision",
		Short: "Build or pull images",
		Long: `
provision will use specified Dockerfiles to build the images.
If no Dockerfile is given, it will pull the image from the index.`,
		Run: func(cmd *cobra.Command, args []string) {
			containers := getContainers(options)
			containers.provision(options.force)
		},
	}

	var cmdRun = &cobra.Command{
		Use:   "run",
		Short: "Run the containers",
		Long:  `run will call docker run on all containers.`,
		Run: func(cmd *cobra.Command, args []string) {
			containers := getContainers(options)
			containers.run(options.force, options.kill)
		},
	}

	var cmdRm = &cobra.Command{
		Use:   "rm",
		Short: "Remove the containers",
		Long:  `rm will call docker rm on all containers.`,
		Run: func(cmd *cobra.Command, args []string) {
			containers := getContainers(options)
			containers.rm(options.force, options.kill)
		},
	}

	var cmdKill = &cobra.Command{
		Use:   "kill",
		Short: "Kill the containers",
		Long:  `kill will call docker kill on all containers.`,
		Run: func(cmd *cobra.Command, args []string) {
			containers := getContainers(options)
			containers.kill()
		},
	}

	var cmdStart = &cobra.Command{
		Use:   "start",
		Short: "Start the containers",
		Long:  `start will call docker start on all containers.`,
		Run: func(cmd *cobra.Command, args []string) {
			containers := getContainers(options)
			containers.start()
		},
	}

	var cmdStop = &cobra.Command{
		Use:   "stop",
		Short: "Stop the containers",
		Long:  `stop will call docker stop on all containers.`,
		Run: func(cmd *cobra.Command, args []string) {
			containers := getContainers(options)
			containers.stop()
		},
	}

	var cmdStatus = &cobra.Command{
		Use:   "status",
		Short: "Displays status of containers",
		Long:  `Displays the current status of the containers.`,
		Run: func(cmd *cobra.Command, args []string) {
			containers := getContainers(options)
			containers.status()
		},
	}

	var cmdCheck = &cobra.Command{
		Use:   "check",
		Short: "Check configuration",
		Long: `check will read the configuration and print it to stdout for debugging.`,
		Run: func(cmd *cobra.Command, args []string) {
			containers := getContainers(options)
			fmt.Printf("%v\n", containers)
		},
	}


	var cmdVersion = &cobra.Command{
		Use:   "version",
		Short: "Display version",
		Long:  `Displays the version of Crane.`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("v0.6.1")
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
	cmdLift.Flags().BoolVarP(&options.force, "force", "f", false, "force")
	cmdLift.Flags().BoolVarP(&options.kill, "kill", "k", false, "kill containers")
	cmdProvision.Flags().BoolVarP(&options.force, "force", "f", false, "force")
	cmdRun.Flags().BoolVarP(&options.force, "force", "f", false, "force")
	cmdRun.Flags().BoolVarP(&options.kill, "kill", "k", false, "kill containers")
	cmdRm.Flags().BoolVarP(&options.force, "force", "f", false, "force")
	cmdRm.Flags().BoolVarP(&options.kill, "kill", "k", false, "kill containers")
	craneCmd.AddCommand(cmdLift, cmdProvision, cmdRun, cmdRm, cmdKill, cmdStart, cmdStop, cmdStatus, cmdVersion, cmdCheck)
	craneCmd.Execute()
}
