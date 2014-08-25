package crane

import (
	"fmt"
	"github.com/michaelsauter/crane/print"
	"github.com/spf13/cobra"
	"strings"
)

type Options struct {
	verbose             bool
	recreate            bool
	nocache             bool
	notrunc             bool
	kill                bool
	cascadeDependencies string
	cascadeAffected     string
	config              string
	target              []string
}

var options = Options{
	verbose:             false,
	recreate:            false,
	nocache:             false,
	notrunc:             false,
	kill:                false,
	cascadeDependencies: "",
	cascadeAffected:     "",
	config:              "",
	target:              make([]string, 1), //FIXME: remove pre-allocation when -t/--target is removed
}

func isVerbose() bool {
	return options.verbose
}

// returns a function to be set as a cobra command run, wrapping a command meant to be run on a set of containers
func containersCommand(wrapped func(containers Containers), forceOrder bool) func(cmd *cobra.Command, args []string) {
	return func(cmd *cobra.Command, args []string) {
		for _, value := range []string{options.cascadeDependencies, options.cascadeAffected} {
			if value != "none" && value != "all" && value != "link" && value != "volumesFrom" && value != "net" {
				cmd.Printf("Error: invalid cascading value: %v", value)
				cmd.Usage()
				panic(StatusError{status: 64})
			}
		}
		if options.target[0] != "" { //FIXME: remove when -t/--target is removed
			print.Noticef("DEPRECATION: -t/--target is now implicit and will be removed in an upcoming release\n")
			if len(args) > 0 {
				options.target = append(args, options.target[0])
			}
		} else {
			options.target = args
		}

		containers := NewConfig(options, forceOrder).Containers()
		if len(containers) == 0 {
			print.Errorf("ERROR: Command cannot be applied to any container.")
		} else {
			if isVerbose() {
				print.Infof("Command will be applied to: %v\n\n", strings.Join(containers.names(), ", "))
			}
			wrapped(containers)
		}
	}
}

func handleCmd() {

	var cmdLift = &cobra.Command{
		Use:   "lift",
		Short: "Build or pull images if they don't exist, then run or start the containers",
		Long: `
lift will provision missing images and run all targeted containers.`,
		Run: containersCommand(func(containers Containers) {
			containers.lift(options.recreate, options.nocache)
		}, false),
	}

	var cmdProvision = &cobra.Command{
		Use:   "provision",
		Short: "Build or pull images",
		Long: `
provision will use specified Dockerfiles to build all targeted images.
If no Dockerfile is given, it will pull the image(s) from the given registry.`,
		Run: containersCommand(func(containers Containers) {
			containers.provision(options.nocache)
		}, true),
	}

	var cmdRun = &cobra.Command{
		Use:   "run",
		Short: "Run the containers",
		Long:  `run will call docker run for all targeted containers.`,
		Run: containersCommand(func(containers Containers) {
			containers.run(options.recreate)
		}, false),
	}

	var cmdRm = &cobra.Command{
		Use:   "rm",
		Short: "Remove the containers",
		Long:  `rm will call docker rm for all targeted containers.`,
		Run: containersCommand(func(containers Containers) {
			containers.reversed().rm(options.kill)
		}, true),
	}

	var cmdKill = &cobra.Command{
		Use:   "kill",
		Short: "Kill the containers",
		Long:  `kill will call docker kill for all targeted containers.`,
		Run: containersCommand(func(containers Containers) {
			containers.reversed().kill()
		}, true),
	}

	var cmdStart = &cobra.Command{
		Use:   "start",
		Short: "Start the containers",
		Long:  `start will call docker start for all targeted containers.`,
		Run: containersCommand(func(containers Containers) {
			containers.start()
		}, false),
	}

	var cmdStop = &cobra.Command{
		Use:   "stop",
		Short: "Stop the containers",
		Long:  `stop will call docker stop for all targeted containers.`,
		Run: containersCommand(func(containers Containers) {
			containers.reversed().stop()
		}, true),
	}

	var cmdPause = &cobra.Command{
		Use:   "pause",
		Short: "Pause the containers",
		Long:  `pause will call docker pause for all targeted containers.`,
		Run: containersCommand(func(containers Containers) {
			containers.reversed().pause()
		}, true),
	}

	var cmdUnpause = &cobra.Command{
		Use:   "unpause",
		Short: "Unpause the containers",
		Long:  `unpause will call docker unpause for all targeted containers.`,
		Run: containersCommand(func(containers Containers) {
			containers.unpause()
		}, false),
	}

	var cmdPush = &cobra.Command{
		Use:   "push",
		Short: "Push the containers",
		Long:  `push will call docker push for all targeted containers.`,
		Run: containersCommand(func(containers Containers) {
			containers.push()
		}, true),
	}

	var cmdStatus = &cobra.Command{
		Use:   "status",
		Short: "Displays status of containers",
		Long:  `Displays the current status of all targeted containers.`,
		Run: containersCommand(func(containers Containers) {
			containers.status(options.notrunc)
		}, true),
	}

	var cmdVersion = &cobra.Command{
		Use:   "version",
		Short: "Display version",
		Long:  `Displays the version of Crane.`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("v0.9.0")
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
	craneCmd.PersistentFlags().StringVarP(&options.target[0], "target", "t", "", "Group or container to execute the command for [DEPRECATED, NOW IMPLICIT]")
	cascadingValuesSuffix := `
					"all": follow any kind of dependency
					"link": follow --link dependencies only
					"volumesFrom": follow --volumesFrom dependencies only
					"net": follow --net dependencies only
	`
	craneCmd.PersistentFlags().StringVarP(&options.cascadeDependencies, "cascade-dependencies", "d", "none", "Also apply the command for the containers that (any of) the explicitly targeted one(s) depend on"+cascadingValuesSuffix)
	craneCmd.PersistentFlags().StringVarP(&options.cascadeAffected, "cascade-affected", "a", "none", "Also apply the command for the existing containers depending on (any of) the explicitly targeted one(s)"+cascadingValuesSuffix)

	cmdLift.Flags().BoolVarP(&options.recreate, "recreate", "r", false, "Recreate containers (kill and remove containers if they exist, force-provision images, run containers)")
	cmdLift.Flags().BoolVarP(&options.nocache, "no-cache", "n", false, "Build the image without any cache")

	cmdProvision.Flags().BoolVarP(&options.nocache, "no-cache", "n", false, "Build the image without any cache")

	cmdRun.Flags().BoolVarP(&options.recreate, "recreate", "r", false, "Recreate containers (kill and remove containers first)")

	cmdRm.Flags().BoolVarP(&options.kill, "kill", "k", false, "Kill containers if they are running first")

	cmdStatus.Flags().BoolVarP(&options.notrunc, "no-trunc", "", false, "Don't truncate output")

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

	craneCmd.AddCommand(cmdLift, cmdProvision, cmdRun, cmdRm, cmdKill, cmdStart, cmdStop, cmdPause, cmdUnpause, cmdPush, cmdStatus, cmdVersion)
	err := craneCmd.Execute()
	if err != nil {
		panic(StatusError{status: 64})
	}
}
