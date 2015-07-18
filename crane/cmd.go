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
var excluded []string

var (
	app          = kingpin.New("crane", "Lift containers with ease")
	interspersed = app.Interspersed(false)
	verboseFlag  = app.Flag("verbose", "Enable verbose output.").Short('v').Bool()
	configFlag   = app.Flag(
		"config",
		"Location of config file.",
	).Short('c').PlaceHolder("~/crane.yaml").String()

	liftCommand = app.Command(
		"lift",
		"Build or pull images if they don't exist, then run or start the containers.",
	)
	liftNoCacheFlag = liftCommand.Flag(
		"no-cache",
		"Build the image without any cache.",
	).Short('n').Bool()
	liftExcludeFlag = liftCommand.Flag(
		"exclude",
		"Exclude group or container",
	).Short('e').String()
	liftTargetArg = liftCommand.Arg("target", "Target of command").String()
	liftCmdArg    = liftCommand.Arg("cmd", "Command for container").Strings()

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
	runExcludeFlag = runCommand.Flag(
		"exclude",
		"Exclude group or container",
	).Short('e').String()
	runTargetArg = runCommand.Arg("target", "Target of command").String()
	runCmdArg    = runCommand.Arg("cmd", "Command for container").Strings()

	createCommand = app.Command(
		"create",
		"Create the containers.",
	)
	createExcludeFlag = createCommand.Flag(
		"exclude",
		"Exclude group or container",
	).Short('e').String()
	createTargetArg = createCommand.Arg("target", "Target of command").String()
	createCmdArg    = createCommand.Arg("cmd", "Command for container").Strings()

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

func commandAction(targetFlag string, wrapped func(unitOfWork *UnitOfWork)) {

	cfg = NewConfig(*configFlag)
	excluded = excludedContainers([]string{*liftExcludeFlag, *createExcludeFlag, *runExcludeFlag})
	dependencyGraph = cfg.DependencyGraph(excluded)
	target := NewTarget(dependencyGraph, targetFlag, excluded)
	unitOfWork, err := NewUnitOfWork(dependencyGraph, target.all())
	if err != nil {
		panic(StatusError{err, 78})
	}

	if isVerbose() {
		print.Infof("Command will be applied to: %v\n\n", strings.Join(unitOfWork.order, ", "))
	}
	wrapped(unitOfWork)
}

func excludedContainers(flags []string) []string {
	for _, flag := range flags {
		if len(flag) > 0 {
			return cfg.ContainersForReference(flag)
		}
	}
	return []string{}
}

func handleCmd() {
	switch kingpin.MustParse(app.Parse(os.Args[1:])) {

	case liftCommand.FullCommand():
		commandAction(*liftTargetArg, func(uow *UnitOfWork) {
			uow.Lift(*liftCmdArg, excluded, *liftNoCacheFlag)
		})

	case versionCommand.FullCommand():
		fmt.Println("v1.5.0")

	case graphCommand.FullCommand():
		commandAction(*graphTargetArg, func(uow *UnitOfWork) {
			dependencyGraph.DOT(os.Stdout, uow.Targeted())
		})

	case statsCommand.FullCommand():
		commandAction(*statsTargetArg, func(uow *UnitOfWork) {
			uow.Stats()
		})

	case statusCommand.FullCommand():
		commandAction(*statusTargetArg, func(uow *UnitOfWork) {
			uow.Status(*noTruncFlag)
		})

	case pushCommand.FullCommand():
		commandAction(*pushTargetArg, func(uow *UnitOfWork) {
			uow.Push()
		})

	case unpauseCommand.FullCommand():
		commandAction(*unpauseTargetArg, func(uow *UnitOfWork) {
			uow.Unpause()
		})

	case pauseCommand.FullCommand():
		commandAction(*pauseTargetArg, func(uow *UnitOfWork) {
			uow.Pause()
		})

	case startCommand.FullCommand():
		commandAction(*startTargetArg, func(uow *UnitOfWork) {
			uow.Start()
		})

	case stopCommand.FullCommand():
		commandAction(*stopTargetArg, func(uow *UnitOfWork) {
			uow.Stop()
		})

	case killCommand.FullCommand():
		commandAction(*killTargetArg, func(uow *UnitOfWork) {
			uow.Kill()
		})

	case rmCommand.FullCommand():
		commandAction(*rmTargetArg, func(uow *UnitOfWork) {
			uow.Rm(*forceRmFlag)
		})

	case runCommand.FullCommand():
		commandAction(*runTargetArg, func(uow *UnitOfWork) {
			uow.Run(*runCmdArg, excluded)
		})

	case createCommand.FullCommand():
		commandAction(*createTargetArg, func(uow *UnitOfWork) {
			uow.Create(*createCmdArg, excluded)
		})

	case provisionCommand.FullCommand():
		commandAction(*provisionTargetArg, func(uow *UnitOfWork) {
			uow.Provision(*provisionNoCacheFlag)
		})

	case pullCommand.FullCommand():
		commandAction(*pullTargetArg, func(uow *UnitOfWork) {
			uow.PullImage()
		})

	case logsCommand.FullCommand():
		commandAction(*logsTargetArg, func(uow *UnitOfWork) {
			uow.Logs(*followFlag, *timestampsFlag, *tailFlag, *colorizeFlag)
		})
	}
}
