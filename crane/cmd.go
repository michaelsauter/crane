package crane

import (
	"fmt"
	"github.com/alecthomas/kingpin"
	"github.com/michaelsauter/crane/print"
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
	ignoreMissing       string
	config              string
	target              []string
}

var options Options
var cfg Config

var (
	app          = kingpin.New("crane", "Lift containers with ease")
	interspersed = app.Interspersed(false)
	verboseFlag  = app.Flag("verbose", "Enable verbose output.").Short('v').Bool()
	configFlag   = app.Flag(
		"config",
		"Location of config file.",
	).Short('c').PlaceHolder("~/crane.yaml").String()
	cascadeDependenciesFlag = app.Flag(
		"cascade-dependencies",
		"Also apply the command for the containers that (any of) the explicitly targeted one(s) depend on.",
	).Short('d').Default("none").String()
	cascadeAffectedFlag = app.Flag(
		"cascade-affected",
		"Also apply the command for the existing containers depending on (any of) the explicitly targeted one(s).",
	).Short('a').Default("none").String()

	liftCommand = app.Command(
		"lift",
		"Build or pull images if they don't exist, then run or start the containers.",
	)
	recreateFlag = liftCommand.Flag(
		"recreate",
		"Recreate containers (force-remove containers if they exist, force-provision images, run containers).",
	).Short('r').Bool()
	ignoreMissingFlag = liftCommand.Flag(
		"ignore-missing",
		"Rather than failing, ignore dependencies that are not fullfilled.",
	).Short('i').Default("none").String()
	noCacheFlag = liftCommand.Flag(
		"no-cache",
		"Build the image without any cache.",
	).Short('n').Bool()
	liftTargetArg = liftCommand.Arg("target", "Target of command").String()

	versionCommand = app.Command(
		"version",
		"Displays the version of Crane.",
	)

	graphCommand = app.Command(
		"graph",
		"Dumps the dependency graph as a DOT file.",
	)
	graphTargetArg = graphCommand.Arg("target", "Target of command").String()

	statsCommand = app.Command(
		"stats",
		"Displays statistics about containers.",
	)
	statsTargetArg = statsCommand.Arg("target", "Target of command").String()

	statusCommand = app.Command(
		"status",
		"Displays status of containers.",
	)
	noTruncFlag = liftCommand.Flag(
		"no-trunc",
		"Don't truncate output.",
	).Bool()
	statusTargetArg = statusCommand.Arg("target", "Target of command").String()

	pushCommand = app.Command(
		"push",
		"Push the containers to the registry.",
	)
	pushTargetArg = pushCommand.Arg("target", "Target of command").String()

	pauseCommand = app.Command(
		"pause",
		"Pause the containers.",
	)
	pauseTargetArg = pauseCommand.Arg("target", "Target of command").String()

	unpauseCommand = app.Command(
		"unpause",
		"Unpause the containers.",
	)
	unpauseTargetArg = unpauseCommand.Arg("target", "Target of command").String()

	startCommand = app.Command(
		"start",
		"Start the containers.",
	)
	startTargetArg = startCommand.Arg("target", "Target of command").String()

	stopCommand = app.Command(
		"stop",
		"Stop the containers.",
	)
	stopTargetArg = stopCommand.Arg("target", "Target of command").String()
)

func isVerbose() bool {
	return options.verbose
}

func action(target string, wrapped func(), forceOrder bool) {

	options.cascadeDependencies = *cascadeDependenciesFlag
	options.cascadeAffected = *cascadeAffectedFlag
	options.config = *configFlag
	options.recreate = *recreateFlag
	options.ignoreMissing = *ignoreMissingFlag
	options.nocache = *noCacheFlag
	options.verbose = *verboseFlag
	options.notrunc = *noTruncFlag

	// for _, value := range []string{options.cascadeDependencies, options.cascadeAffected, options.ignoreMissing} {
	// 	if value != "none" && value != "all" && value != "link" && value != "volumesFrom" && value != "net" {
	// 		print.Errorf("Error: invalid dependency type value: %v", value)
	// 		panic(StatusError{status: 64})
	// 	}
	// }

	// Set target from args
	options.target = []string{target}
	cfg = NewConfig(options, forceOrder)
	if containers := cfg.TargetedContainers(); len(containers) == 0 {
		print.Errorf("ERROR: Command cannot be applied to any container.")
	} else {
		if isVerbose() {
			print.Infof("Command will be applied to: %v\n\n", strings.Join(containers.names(), ", "))
		}
		wrapped()
	}
}

func handleCmd() {
	switch kingpin.MustParse(app.Parse(os.Args[1:])) {

	case liftCommand.FullCommand():
		action(*liftTargetArg, func() {
			cfg.TargetedContainers().lift(options.recreate, options.nocache, options.ignoreMissing, cfg.Path())
		}, false)

	case versionCommand.FullCommand():
		fmt.Println("v1.5.0")

	case graphCommand.FullCommand():
		action(*graphTargetArg, func() {
			cfg.DependencyGraph().DOT(os.Stdout, cfg.TargetedContainers())
		}, true)

	case statsCommand.FullCommand():
		action(*statsTargetArg, func() {
			cfg.TargetedContainers().stats()
		}, true)

	case statusCommand.FullCommand():
		action(*statusTargetArg, func() {
			cfg.TargetedContainers().status(options.notrunc)
		}, true)

	case pushCommand.FullCommand():
		action(*pushTargetArg, func() {
			cfg.TargetedContainers().push()
		}, true)

	case unpauseCommand.FullCommand():
		action(*unpauseTargetArg, func() {
			cfg.TargetedContainers().unpause()
		}, false)

	case pauseCommand.FullCommand():
		action(*pauseTargetArg, func() {
			cfg.TargetedContainers().reversed().pause()
		}, true)

	case startCommand.FullCommand():
		action(*startTargetArg, func() {
			cfg.TargetedContainers().start()
		}, false)

	case stopCommand.FullCommand():
		action(*stopTargetArg, func() {
			cfg.TargetedContainers().reversed().stop()
		}, true)
	}

	// 	var cmdProvision = &cobra.Command{
	// 		Use:   "provision",
	// 		Short: "Build or pull images",
	// 		Long: `
	// provision will use specified Dockerfiles to build all targeted images.
	// If no Dockerfile is given, it will pull the image(s) from the given registry.`,
	// 		Run: configCommand(func(config Config) {
	// 			config.TargetedContainers().provision(options.nocache)
	// 		}, true),
	// 	}

	// 	var cmdPull = &cobra.Command{
	// 		Use:   "pull",
	// 		Short: "Pull images",
	// 		Long: `
	// pull will pull the image(s) from the given registry.`,
	// 		Run: configCommand(func(config Config) {
	// 			config.TargetedContainers().pullImage()
	// 		}, true),
	// 	}

	// 	var cmdCreate = &cobra.Command{
	// 		Use:   "create",
	// 		Short: "Create the containers",
	// 		Long:  `run will call docker create for all targeted containers.`,
	// 		Run: configCommand(func(config Config) {
	// 			config.TargetedContainers().create(options.recreate, options.ignoreMissing, config.Path())
	// 		}, false),
	// 	}

	// 	var cmdRun = &cobra.Command{
	// 		Use:   "run",
	// 		Short: "Run the containers",
	// 		Long:  `run will call docker run for all targeted containers.`,
	// 		Run: configCommand(func(config Config) {
	// 			config.TargetedContainers().run(options.recreate, options.ignoreMissing, config.Path())
	// 		}, false),
	// 	}

	// 	var cmdRm = &cobra.Command{
	// 		Use:   "rm",
	// 		Short: "Remove the containers",
	// 		Long:  `rm will call docker rm for all targeted containers.`,
	// 		Run: configCommand(func(config Config) {
	// 			config.TargetedContainers().reversed().rm(options.forceRm)
	// 		}, true),
	// 	}

	// 	var cmdKill = &cobra.Command{
	// 		Use:   "kill",
	// 		Short: "Kill the containers",
	// 		Long:  `kill will call docker kill for all targeted containers.`,
	// 		Run: configCommand(func(config Config) {
	// 			config.TargetedContainers().reversed().kill()
	// 		}, true),
	// 	}

	// 	var cmdLogs = &cobra.Command{
	// 		Use:   "logs",
	// 		Short: "Display container logs",
	// 		Long: `Display an aggregated, name-prefixed view of the logs for the targeted
	// containers. To distinguish the sources better, lines can be colorized by
	// enabling the 'colorize' flag. Containers' STDERR and STDOUT are multiplexed
	// together into the process STDOUT in order to interlace lines properly. Logs
	// originally dumped to STDERR have a line header ending with a '*', and are
	// formatted in bold provided the 'colorize' flag is on.`,
	// 		Run: configCommand(func(config Config) {
	// 			config.TargetedContainers().logs(options.follow, options.timestamps, options.tail, options.colorize)
	// 		}, true),
	// 	}



	// 	cmdProvision.Flags().BoolVarP(&options.nocache, "no-cache", "n", false, "Build the image without any cache")

	// 	cmdCreate.Flags().BoolVarP(&options.recreate, "recreate", "r", false, "Recreate containers (force-remove containers first)")
	// 	cmdCreate.Flags().StringVarP(&options.ignoreMissing, "ignore-missing", "i", "none", "Rather than failing, ignore dependencies that are not fullfilled for:"+dependencyTypeValuesSuffix)

	// 	cmdRun.Flags().BoolVarP(&options.recreate, "recreate", "r", false, "Recreate containers (force-remove containers first)")
	// 	cmdRun.Flags().StringVarP(&options.ignoreMissing, "ignore-missing", "i", "none", "Rather than failing, ignore dependencies that are not fullfilled for:"+dependencyTypeValuesSuffix)

	// 	cmdRm.Flags().BoolVarP(&options.forceRm, "force", "f", false, "Kill containers if they are running first")

	// 	cmdLogs.Flags().BoolVarP(&options.follow, "follow", "f", false, "Follow log output")
	// 	cmdLogs.Flags().BoolVarP(&options.timestamps, "timestamps", "t", false, "Show timestamps")
	// 	cmdLogs.Flags().StringVarP(&options.tail, "tail", "", "all", "Output the specified number of lines at the end of logs")
	// 	cmdLogs.Flags().BoolVarP(&options.colorize, "colorize", "z", false, "Output the lines with one color per container")

	// 	craneCmd.AddCommand(cmdLift, cmdProvision, cmdPull, cmdCreate, cmdRun, cmdRm, cmdKill, cmdStart, cmdStop, cmdPause, cmdUnpause, cmdPush, cmdLogs, cmdStatus, cmdStats, cmdGraph, cmdVersion)
	// 	err := craneCmd.Execute()
	// 	if err != nil {
	// 		panic(StatusError{status: 64})
	// 	}
}
