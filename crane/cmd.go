package crane

import (
	"fmt"
	"github.com/michaelsauter/crane/print"
	"github.com/spf13/cobra"
	"os"
	"strings"
)

type Options struct {
	verbose             bool
	recreate            bool
	nocache             bool
	notrunc             bool
	forceRm             bool
	follow              bool
	timestamps          bool
	tail                string
	colorize            bool
	cascadeDependencies string
	cascadeAffected     string
	config              string
	target              []string
}

var options Options

func isVerbose() bool {
	return options.verbose
}

// returns a function to be set as a cobra command run, wrapping a command meant to be run according to the config
func configCommand(wrapped func(config Config), forceOrder bool) func(cmd *cobra.Command, args []string) {
	return func(cmd *cobra.Command, args []string) {
		for _, value := range []string{options.cascadeDependencies, options.cascadeAffected} {
			if value != "none" && value != "all" && value != "link" && value != "volumesFrom" && value != "net" {
				print.Errorf("Error: invalid cascading value: %v", value)
				cmd.Usage()
				panic(StatusError{status: 64})
			}
		}

		// Set target from args
		options.target = args

		config := NewConfig(options, forceOrder)
		if containers := config.TargetedContainers(); len(containers) == 0 {
			print.Errorf("ERROR: Command cannot be applied to any container.")
		} else {
			if isVerbose() {
				print.Infof("Command will be applied to: %v\n\n", strings.Join(containers.names(), ", "))
			}
			wrapped(config)
		}
	}
}

func handleCmd() {

	var cmdLift = &cobra.Command{
		Use:   "lift",
		Short: "Build or pull images if they don't exist, then run or start the containers",
		Long: `
lift will provision missing images and run all targeted containers.`,
		Run: configCommand(func(config Config) {
			config.TargetedContainers().lift(options.recreate, options.nocache)
		}, false),
	}

	var cmdProvision = &cobra.Command{
		Use:   "provision",
		Short: "Build or pull images",
		Long: `
provision will use specified Dockerfiles to build all targeted images.
If no Dockerfile is given, it will pull the image(s) from the given registry.`,
		Run: configCommand(func(config Config) {
			config.TargetedContainers().provision(options.nocache)
		}, true),
	}

	var cmdCreate = &cobra.Command{
		Use:   "create",
		Short: "Create the containers",
		Long:  `run will call docker create for all targeted containers.`,
		Run: configCommand(func(config Config) {
			config.TargetedContainers().create(options.recreate)
		}, false),
	}

	var cmdRun = &cobra.Command{
		Use:   "run",
		Short: "Run the containers",
		Long:  `run will call docker run for all targeted containers.`,
		Run: configCommand(func(config Config) {
			config.TargetedContainers().run(options.recreate)
		}, false),
	}

	var cmdRm = &cobra.Command{
		Use:   "rm",
		Short: "Remove the containers",
		Long:  `rm will call docker rm for all targeted containers.`,
		Run: configCommand(func(config Config) {
			config.TargetedContainers().reversed().rm(options.forceRm)
		}, true),
	}

	var cmdKill = &cobra.Command{
		Use:   "kill",
		Short: "Kill the containers",
		Long:  `kill will call docker kill for all targeted containers.`,
		Run: configCommand(func(config Config) {
			config.TargetedContainers().reversed().kill()
		}, true),
	}

	var cmdStart = &cobra.Command{
		Use:   "start",
		Short: "Start the containers",
		Long:  `start will call docker start for all targeted containers.`,
		Run: configCommand(func(config Config) {
			config.TargetedContainers().start()
		}, false),
	}

	var cmdStop = &cobra.Command{
		Use:   "stop",
		Short: "Stop the containers",
		Long:  `stop will call docker stop for all targeted containers.`,
		Run: configCommand(func(config Config) {
			config.TargetedContainers().reversed().stop()
		}, true),
	}

	var cmdPause = &cobra.Command{
		Use:   "pause",
		Short: "Pause the containers",
		Long:  `pause will call docker pause for all targeted containers.`,
		Run: configCommand(func(config Config) {
			config.TargetedContainers().reversed().pause()
		}, true),
	}

	var cmdUnpause = &cobra.Command{
		Use:   "unpause",
		Short: "Unpause the containers",
		Long:  `unpause will call docker unpause for all targeted containers.`,
		Run: configCommand(func(config Config) {
			config.TargetedContainers().unpause()
		}, false),
	}

	var cmdPush = &cobra.Command{
		Use:   "push",
		Short: "Push the containers",
		Long:  `push will call docker push for all targeted containers.`,
		Run: configCommand(func(config Config) {
			config.TargetedContainers().push()
		}, true),
	}

	var cmdLogs = &cobra.Command{
		Use:   "logs",
		Short: "Display container logs",
		Long: `Display an aggregated, name-prefixed view of the logs for the targeted
containers. To distinguish the sources better, lines can be colorized by
enabling the 'colorize' flag. Containers' STDERR and STDOUT are multiplexed
together into the process STDOUT in order to interlace lines properly. Logs
originally dumped to STDERR have a line header ending with a '*', and are
formatted in bold provided the 'colorize' flag is on.`,
		Run: configCommand(func(config Config) {
			config.TargetedContainers().logs(options.follow, options.timestamps, options.tail, options.colorize)
		}, true),
	}

	var cmdStatus = &cobra.Command{
		Use:   "status",
		Short: "Displays status of containers",
		Long:  `Displays the current status of all targeted containers.`,
		Run: configCommand(func(config Config) {
			config.TargetedContainers().status(options.notrunc)
		}, true),
	}

	var cmdGraph = &cobra.Command{
		Use:   "graph",
		Short: "Dumps the dependency graph as a DOT file",
		Long: `Generate a DOT file representing the dependency graph. Bold nodes represent the
containers declared in the config (as opposed to non-bold ones that are referenced
in the config, but not defined). Targeted containers are highlighted with color
borders. Solid edges represent links, dashed edges volumesFrom, and dotted edges
net=container relations.`,
		Run: configCommand(func(config Config) {
			config.DependencyGraph().DOT(os.Stdout, config.TargetedContainers())
		}, true),
	}

	var cmdVersion = &cobra.Command{
		Use:   "version",
		Short: "Display version",
		Long:  `Displays the version of Crane.`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("v1.0.0")
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
	cascadingValuesSuffix := `
					"all": follow any kind of dependency
					"link": follow --link dependencies only
					"volumesFrom": follow --volumesFrom dependencies only
					"net": follow --net dependencies only
	`
	craneCmd.PersistentFlags().StringVarP(&options.cascadeDependencies, "cascade-dependencies", "d", "none", "Also apply the command for the containers that (any of) the explicitly targeted one(s) depend on"+cascadingValuesSuffix)
	craneCmd.PersistentFlags().StringVarP(&options.cascadeAffected, "cascade-affected", "a", "none", "Also apply the command for the existing containers depending on (any of) the explicitly targeted one(s)"+cascadingValuesSuffix)

	cmdLift.Flags().BoolVarP(&options.recreate, "recreate", "r", false, "Recreate containers (force-remove containers if they exist, force-provision images, run containers)")
	cmdLift.Flags().BoolVarP(&options.nocache, "no-cache", "n", false, "Build the image without any cache")

	cmdProvision.Flags().BoolVarP(&options.nocache, "no-cache", "n", false, "Build the image without any cache")

	cmdCreate.Flags().BoolVarP(&options.recreate, "recreate", "r", false, "Recreate containers (force-remove containers first)")

	cmdRun.Flags().BoolVarP(&options.recreate, "recreate", "r", false, "Recreate containers (force-remove containers first)")

	cmdRm.Flags().BoolVarP(&options.forceRm, "force", "f", false, "Kill containers if they are running first")

	cmdStatus.Flags().BoolVarP(&options.notrunc, "no-trunc", "", false, "Don't truncate output")

	cmdLogs.Flags().BoolVarP(&options.follow, "follow", "f", false, "Follow log output")
	cmdLogs.Flags().BoolVarP(&options.timestamps, "timestamps", "t", false, "Show timestamps")
	cmdLogs.Flags().StringVarP(&options.tail, "tail", "", "all", "Output the specified number of lines at the end of logs")
	cmdLogs.Flags().BoolVarP(&options.colorize, "colorize", "z", false, "Output the lines with one color per container")

	// default usage template with target arguments & description
	craneCmd.SetUsageTemplate(`{{ $cmd := . }}
Usage: {{if .Runnable}}
  {{.UseLine}}{{if .HasFlags}} [flags]{{end}}{{end}}{{if .HasSubCommands}}
  {{ .CommandPath}} [command]{{end}} [target1 [target2 [...]]]
{{ if .HasSubCommands}}
Available Commands: {{range .Commands}}{{if .Runnable}}
  {{rpad .Use .UsagePadding }} {{.Short}}{{end}}{{end}}
{{end}}
Explicit targeting:
  By default, the command is applied to all containers declared in the
  config,  or to the containers defined in the group ` + "`" + `default` + "`" + ` if it is
  defined. If one or several container or group reference(s) is/are
  passed as  argument(s), the command will only be applied to containers
  matching these references. Note however that providing cascading flags
  might extend the set of targeted containers.

{{ if .HasFlags}}Available Flags:
{{.Flags.FlagUsages}}{{end}}{{if .HasParent}}{{if and (gt .Commands 0) (gt .Parent.Commands 1) }}
Additional help topics: {{if gt .Commands 0 }}{{range .Commands}}{{if not .Runnable}} {{rpad .CommandPath .CommandPathPadding}} {{.Short}}{{end}}{{end}}{{end}}{{if gt .Parent.Commands 1 }}{{range .Parent.Commands}}{{if .Runnable}}{{if not (eq .Name $cmd.Name) }}{{end}}
  {{rpad .CommandPath .CommandPathPadding}} {{.Short}}{{end}}{{end}}{{end}}{{end}}
{{end}}
Use "{{.Root.Name}} help [command]" for more information about that command.
`)

	craneCmd.AddCommand(cmdLift, cmdProvision, cmdCreate, cmdRun, cmdRm, cmdKill, cmdStart, cmdStop, cmdPause, cmdUnpause, cmdPush, cmdLogs, cmdStatus, cmdGraph, cmdVersion)
	err := craneCmd.Execute()
	if err != nil {
		panic(StatusError{status: 64})
	}
}
