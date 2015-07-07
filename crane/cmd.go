package crane

import (
	"fmt"
	"github.com/alecthomas/kingpin"
	"github.com/michaelsauter/crane/print"
	"os"
	"strings"
)

type Options struct {
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
	liftRecreateFlag = liftCommand.Flag(
		"recreate",
		"Recreate containers (force-remove containers if they exist, force-provision images, run containers).",
	).Short('r').Bool()
	liftIgnoreMissingFlag = liftCommand.Flag(
		"ignore-missing",
		"Rather than failing, ignore dependencies that are not fullfilled.",
	).Short('i').Default("none").String()
	liftNoCacheFlag = liftCommand.Flag(
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

	killCommand = app.Command(
		"kill",
		"Kill the containers.",
	)
	killTargetArg = killCommand.Arg("target", "Target of command").String()

	rmCommand = app.Command(
		"rm",
		"Remove the containers.",
	)
	forceRmFlag = rmCommand.Flag(
		"force",
		"Kill containers if they are running first.",
	).Short('f').Bool()
	rmTargetArg = rmCommand.Arg("target", "Target of command").String()

	runCommand = app.Command(
		"run",
		"Run the containers.",
	)
	runRecreateFlag = runCommand.Flag(
		"recreate",
		"Recreate containers (force-remove containers if they exist, force-provision images, run containers).",
	).Short('r').Bool()
	runIgnoreMissingFlag = runCommand.Flag(
		"ignore-missing",
		"Rather than failing, ignore dependencies that are not fullfilled.",
	).Short('i').Default("none").String()
	runTargetArg = runCommand.Arg("target", "Target of command").String()

	createCommand = app.Command(
		"create",
		"Create the containers.",
	)
	createRecreateFlag = createCommand.Flag(
		"recreate",
		"Recreate containers (force-remove containers if they exist, force-provision images, run containers).",
	).Short('r').Bool()
	createIgnoreMissingFlag = createCommand.Flag(
		"ignore-missing",
		"Rather than failing, ignore dependencies that are not fullfilled.",
	).Short('i').Default("none").String()
	createTargetArg = createCommand.Arg("target", "Target of command").String()

	provisionCommand = app.Command(
		"provision",
		"Build or pull images.",
	)
	provisionNoCacheFlag = provisionCommand.Flag(
		"no-cache",
		"Build the image without any cache.",
	).Short('n').Bool()
	provisionTargetArg = provisionCommand.Arg("target", "Target of command").String()

	pullCommand = app.Command(
		"pull",
		"Pull images.",
	)
	pullTargetArg = pullCommand.Arg("target", "Target of command").String()

	logsCommand = app.Command(
		"logs",
		"Display container logs.",
	)
	followFlag = logsCommand.Flag(
		"follow",
		"Follow log output.",
	).Short('f').Bool()
	tailFlag = logsCommand.Flag(
		"tail",
		"Output the specified number of lines at the end of logs.",
	).String()
	timestampsFlag = logsCommand.Flag(
		"timestamps",
		"Show timestamps.",
	).Short('t').Bool()
	colorizeFlag = logsCommand.Flag(
		"colorize",
		"Output the lines with one color per container.",
	).Short('z').Bool()
	logsTargetArg = logsCommand.Arg("target", "Target of command").String()
)

func isVerbose() bool {
	return *verboseFlag
}

func action(target string, wrapped func(), forceOrder bool, ignoreMissing string) {

	options.cascadeDependencies = *cascadeDependenciesFlag
	options.cascadeAffected = *cascadeAffectedFlag
	options.config = *configFlag
	options.ignoreMissing = ignoreMissing

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
			cfg.TargetedContainers().lift(*liftRecreateFlag, *liftNoCacheFlag, *liftIgnoreMissingFlag, cfg.Path())
		}, false, *liftIgnoreMissingFlag)

	case versionCommand.FullCommand():
		fmt.Println("v1.5.0")

	case graphCommand.FullCommand():
		action(*graphTargetArg, func() {
			cfg.DependencyGraph().DOT(os.Stdout, cfg.TargetedContainers())
		}, true, "none")

	case statsCommand.FullCommand():
		action(*statsTargetArg, func() {
			cfg.TargetedContainers().stats()
		}, true, "none")

	case statusCommand.FullCommand():
		action(*statusTargetArg, func() {
			cfg.TargetedContainers().status(*noTruncFlag)
		}, true, "none")

	case pushCommand.FullCommand():
		action(*pushTargetArg, func() {
			cfg.TargetedContainers().push()
		}, true, "none")

	case unpauseCommand.FullCommand():
		action(*unpauseTargetArg, func() {
			cfg.TargetedContainers().unpause()
		}, false, "none")

	case pauseCommand.FullCommand():
		action(*pauseTargetArg, func() {
			cfg.TargetedContainers().reversed().pause()
		}, true, "none")

	case startCommand.FullCommand():
		action(*startTargetArg, func() {
			cfg.TargetedContainers().start()
		}, false, "none")

	case stopCommand.FullCommand():
		action(*stopTargetArg, func() {
			cfg.TargetedContainers().reversed().stop()
		}, true, "none")

	case killCommand.FullCommand():
		action(*killTargetArg, func() {
			cfg.TargetedContainers().reversed().kill()
		}, true, "none")

	case rmCommand.FullCommand():
		action(*rmTargetArg, func() {
			cfg.TargetedContainers().reversed().rm(*forceRmFlag)
		}, true, "none")

	case runCommand.FullCommand():
		action(*runTargetArg, func() {
			cfg.TargetedContainers().run(*runRecreateFlag, *runIgnoreMissingFlag, cfg.Path())
		}, false, *runIgnoreMissingFlag)

	case createCommand.FullCommand():
		action(*createTargetArg, func() {
			cfg.TargetedContainers().create(*createRecreateFlag, *createIgnoreMissingFlag, cfg.Path())
		}, false, *createIgnoreMissingFlag)

	case provisionCommand.FullCommand():
		action(*provisionTargetArg, func() {
			cfg.TargetedContainers().provision(*provisionNoCacheFlag)
		}, true, "none")

	case pullCommand.FullCommand():
		action(*pullTargetArg, func() {
			cfg.TargetedContainers().pullImage()
		}, true, "none")

	case logsCommand.FullCommand():
		action(*logsTargetArg, func() {
			cfg.TargetedContainers().logs(*followFlag, *timestampsFlag, *tailFlag, *colorizeFlag)
		}, true, "none")
	}
}
