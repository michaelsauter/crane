package crane

import (
	"fmt"
	"github.com/spf13/cobra"
)

type Options struct {
	verbose  bool
	recreate bool
	nocache  bool
	notrunc  bool
	kill     bool
	config   string
	target   string
}

var options = Options{
	verbose:  false,
	recreate: false,
	nocache:  false,
	notrunc:  false,
	kill:     false,
	config:   "",
	target:   "",
}

func isVerbose() bool {
	return options.verbose
}

// returns a function to be set as a cobra command run, wrapping a command meant to be run on a set of containers
func containersCommand(wrapped func(containers Containers), reversed bool, ignoreUnresolved bool) func(cmd *cobra.Command, args []string) {
	return func(cmd *cobra.Command, args []string) {
		if len(args) > 0 {
			cmd.Printf("Error: too many arguments given: %#q", args)
			cmd.Usage()
			panic(StatusError{status: 64})
		}
		wrapped(NewConfig(options, reversed, ignoreUnresolved).Containers())
	}
}

func handleCmd() {

	var cmdLift = &cobra.Command{
		Use:   "lift",
		Short: "Build or pull images, then run or start the containers",
		Long: `
lift will provision and run all targeted containers.`,
		Run: containersCommand(func(containers Containers) {
			containers.lift(options.recreate, options.nocache)
		}, false, false),
	}

	var cmdProvision = &cobra.Command{
		Use:   "provision",
		Short: "Build or pull images",
		Long: `
provision will use specified Dockerfiles to build all targeted images.
If no Dockerfile is given, it will pull the image(s) from the given registry.`,
		Run: containersCommand(func(containers Containers) {
			containers.provision(options.nocache)
		}, false, true),
	}

	var cmdRun = &cobra.Command{
		Use:   "run",
		Short: "Run the containers",
		Long:  `run will call docker run for all targeted containers.`,
		Run: containersCommand(func(containers Containers) {
			containers.run(options.recreate)
		}, false, false),
	}

	var cmdRm = &cobra.Command{
		Use:   "rm",
		Short: "Remove the containers",
		Long:  `rm will call docker rm for all targeted containers.`,
		Run: containersCommand(func(containers Containers) {
			containers.rm(options.kill)
		}, true, false),
	}

	var cmdKill = &cobra.Command{
		Use:   "kill",
		Short: "Kill the containers",
		Long:  `kill will call docker kill for all targeted containers.`,
		Run: containersCommand(func(containers Containers) {
			containers.kill()
		}, true, false),
	}

	var cmdStart = &cobra.Command{
		Use:   "start",
		Short: "Start the containers",
		Long:  `start will call docker start for all targeted containers.`,
		Run: containersCommand(func(containers Containers) {
			containers.start()
		}, false, false),
	}

	var cmdStop = &cobra.Command{
		Use:   "stop",
		Short: "Stop the containers",
		Long:  `stop will call docker stop for all targeted containers.`,
		Run: containersCommand(func(containers Containers) {
			containers.stop()
		}, true, false),
	}

	var cmdPause = &cobra.Command{
		Use:   "pause",
		Short: "Pause the containers",
		Long:  `pause will call docker pause for all targeted containers.`,
		Run: containersCommand(func(containers Containers) {
			containers.pause()
		}, true, false),
	}

	var cmdUnpause = &cobra.Command{
		Use:   "unpause",
		Short: "Unpause the containers",
		Long:  `unpause will call docker unpause for all targeted containers.`,
		Run: containersCommand(func(containers Containers) {
			containers.unpause()
		}, false, false),
	}

	var cmdPush = &cobra.Command{
		Use:   "push",
		Short: "Push the containers",
		Long:  `push will call docker push for all targeted containers.`,
		Run: containersCommand(func(containers Containers) {
			containers.push()
		}, false, false),
	}

	var cmdStatus = &cobra.Command{
		Use:   "status",
		Short: "Displays status of containers",
		Long:  `Displays the current status of all targeted containers.`,
		Run: containersCommand(func(containers Containers) {
			containers.status(options.notrunc)
		}, false, true),
	}

	var cmdVersion = &cobra.Command{
		Use:   "version",
		Short: "Display version",
		Long:  `Displays the version of Crane.`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("v0.8.0")
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

	craneCmd.PersistentFlags().BoolVarP(&options.verbose, "verbose", "v", false, "Verbose output")
	craneCmd.PersistentFlags().StringVarP(&options.config, "config", "c", "", "Config file to read from")
	craneCmd.PersistentFlags().StringVarP(&options.target, "target", "t", "", "Group or container to execute the command for")

	cmdLift.Flags().BoolVarP(&options.recreate, "recreate", "r", false, "Recreate containers (kill and remove containers, provision images, run containers)")
	cmdLift.Flags().BoolVarP(&options.nocache, "no-cache", "n", false, "Build the image without any cache")

	cmdProvision.Flags().BoolVarP(&options.nocache, "no-cache", "n", false, "Build the image without any cache")

	cmdRun.Flags().BoolVarP(&options.recreate, "recreate", "r", false, "Recreate containers (kill and remove containers first)")

	cmdRm.Flags().BoolVarP(&options.kill, "kill", "k", false, "Kill containers if they are running first")

	cmdStatus.Flags().BoolVarP(&options.notrunc, "no-trunc", "", false, "Don't truncate output")

	craneCmd.AddCommand(cmdLift, cmdProvision, cmdRun, cmdRm, cmdKill, cmdStart, cmdStop, cmdPause, cmdUnpause, cmdPush, cmdStatus, cmdVersion)
	err := craneCmd.Execute()
	if err != nil {
		panic(StatusError{status: 64})
	}
}
