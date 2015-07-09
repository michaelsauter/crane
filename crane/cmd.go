package crane

import (
	"fmt"
	"github.com/alecthomas/kingpin"
	"github.com/michaelsauter/crane/print"
	"os"
	"strings"
)

var cfg Config
var dependencyGraph DependencyGraph

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
	startIgnoreMissingFlag = startCommand.Flag(
		"ignore-missing",
		"Rather than failing, ignore dependencies that are not fullfilled.",
	).Short('i').Default("none").String()
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
	runIgnoreMissingFlag = runCommand.Flag(
		"ignore-missing",
		"Rather than failing, ignore dependencies that are not fullfilled.",
	).Short('i').Default("none").String()
	runTargetArg = runCommand.Arg("target", "Target of command").String()

	createCommand = app.Command(
		"create",
		"Create the containers.",
	)
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

func action(targetFlag string, wrapped func(containers Containers), forceOrder bool, ignoreMissing string) {

	cfg = NewConfig(*configFlag)
	dependencyGraph = cfg.DependencyGraph()
	target := NewTarget(targetFlag, *cascadeDependenciesFlag, *cascadeAffectedFlag)
	order, err := dependencyGraph.order(target, ignoreMissing)
	if err != nil {
		panic(StatusError{err, 78})
	}

	var containers Containers
	containerMap := cfg.ContainerMap()
	for _, name := range order {
		containers = append([]Container{containerMap[name]}, containers...)
	}

	if len(containers) == 0 {
		print.Errorf("ERROR: Command cannot be applied to any container.")
	} else {
		if isVerbose() {
			print.Infof("Command will be applied to: %v\n\n", strings.Join(containers.names(), ", "))
		}
		wrapped(containers)
	}
}

func handleCmd() {
	switch kingpin.MustParse(app.Parse(os.Args[1:])) {

	case liftCommand.FullCommand():
		action(*liftTargetArg, func(containers Containers) {
			containers.lift(*liftNoCacheFlag, *liftIgnoreMissingFlag, cfg.Path())
		}, false, *liftIgnoreMissingFlag)

	case versionCommand.FullCommand():
		fmt.Println("v1.5.0")

	case graphCommand.FullCommand():
		action(*graphTargetArg, func(containers Containers) {
			cfg.DependencyGraph().DOT(os.Stdout, containers)
		}, true, "none")

	case statsCommand.FullCommand():
		action(*statsTargetArg, func(containers Containers) {
			containers.stats()
		}, true, "none")

	case statusCommand.FullCommand():
		action(*statusTargetArg, func(containers Containers) {
			containers.status(*noTruncFlag)
		}, true, "none")

	case pushCommand.FullCommand():
		action(*pushTargetArg, func(containers Containers) {
			containers.push()
		}, true, "none")

	case unpauseCommand.FullCommand():
		action(*unpauseTargetArg, func(containers Containers) {
			containers.unpause()
		}, false, "none")

	case pauseCommand.FullCommand():
		action(*pauseTargetArg, func(containers Containers) {
			containers.reversed().pause()
		}, true, "none")

	case startCommand.FullCommand():
		action(*startTargetArg, func(containers Containers) {
			containers.start(*startIgnoreMissingFlag, cfg.Path())
		}, false, "none")

	case stopCommand.FullCommand():
		action(*stopTargetArg, func(containers Containers) {
			containers.reversed().stop()
		}, true, "none")

	case killCommand.FullCommand():
		action(*killTargetArg, func(containers Containers) {
			containers.reversed().kill()
		}, true, "none")

	case rmCommand.FullCommand():
		action(*rmTargetArg, func(containers Containers) {
			containers.reversed().rm(*forceRmFlag)
		}, true, "none")

	case runCommand.FullCommand():
		action(*runTargetArg, func(containers Containers) {
			containers.run(*runIgnoreMissingFlag, cfg.Path())
		}, false, *runIgnoreMissingFlag)

	case createCommand.FullCommand():
		action(*createTargetArg, func(containers Containers) {
			containers.create(*createIgnoreMissingFlag, cfg.Path())
		}, false, *createIgnoreMissingFlag)

	case provisionCommand.FullCommand():
		action(*provisionTargetArg, func(containers Containers) {
			containers.provision(*provisionNoCacheFlag)
		}, true, "none")

	case pullCommand.FullCommand():
		action(*pullTargetArg, func(containers Containers) {
			containers.pullImage()
		}, true, "none")

	case logsCommand.FullCommand():
		action(*logsTargetArg, func(containers Containers) {
			containers.logs(*followFlag, *timestampsFlag, *tailFlag, *colorizeFlag)
		}, true, "none")
	}
}
