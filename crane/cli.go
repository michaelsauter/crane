package crane

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/alecthomas/kingpin"
)

var cfg Config
var allowed []string

var (
	app         = kingpin.New("crane", "Lift containers with ease").Interspersed(false).DefaultEnvars()
	verboseFlag = app.Flag("verbose", "Enable verbose output.").Short('v').Bool()
	dryRunFlag = app.Flag("dry-run", "Dry run (implicit verbose, no side effects).").Bool()
	configFlag  = app.Flag(
		"config",
		"Location of config file.",
	).Short('c').PlaceHolder("~/crane.yaml").String()
	prefixFlag = app.Flag(
		"prefix",
		"Container prefix.",
	).Short('p').String()
	excludeFlag = app.Flag(
		"exclude",
		"Exclude group or container. Can be repeated.",
	).Short('e').PlaceHolder("container|group").Strings()
	onlyFlag = app.Flag(
		"only",
		"Include only group or container.",
	).Short('o').PlaceHolder("container|group").String()
	tagFlag = app.Flag(
		"tag",
		"Override image tags.",
	).OverrideDefaultFromEnvar("CRANE_TAG").String()

	liftCommand = app.Command(
		"lift",
		"Build or pull images if they don't exist, then run or start the containers.",
	)
	liftNoCacheFlag = liftCommand.Flag(
		"no-cache",
		"Build the image without any cache.",
	).Short('n').Bool()
	liftParallelFlag = liftCommand.Flag(
		"parallel",
		"Define how many containers are provisioned in parallel.",
	).Short('l').Default("1").Int()
	liftTargetArg = liftCommand.Arg("target", "Target of command").String()
	liftCmdArg    = liftCommand.Arg("cmd", "Command for container").Strings()

	versionCommand = app.Command(
		"version",
		"Displays the version of Crane.",
	)

	statsCommand = app.Command(
		"stats",
		"Displays statistics about containers.",
	)
	statsNoStreamFlag = statsCommand.Flag(
		"no-stream",
		"Disable stats streaming (Docker >= 1.7).",
	).Short('n').Bool()
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

	execCommand = app.Command(
		"exec",
		"Execute command in the container(s).",
	)
	execTargetArg = execCommand.Arg("target", "Target of command").String()
	execCmdArg    = execCommand.Arg("cmd", "Command for container").Strings()

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
	runTargetArg = runCommand.Arg("target", "Target of command").String()
	runCmdArg    = runCommand.Arg("cmd", "Command for container").Strings()

	createCommand = app.Command(
		"create",
		"Create the containers.",
	)
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
	provisionParallelFlag = provisionCommand.Flag(
		"parallel",
		"Define how many containers are provisioned in parallel.",
	).Short('l').Default("1").Int()
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
	sinceFlag = logsCommand.Flag(
		"since",
		"Show logs since timestamp (Docker >= 1.7).",
	).String()
	logsTargetArg = logsCommand.Arg("target", "Target of command").String()

	generateCommand = app.Command(
		"generate",
		"Generate files by passing the config to a given template.",
	)
	templateFlag = generateCommand.Flag(
		"template",
		"Template to use.",
	).Short('t').String()
	outputFlag = generateCommand.Flag(
		"output",
		"The file(s) to write the output to.",
	).Short('O').String()
	generateTargetArg = generateCommand.Arg("target", "Target of command").String()

	syncCommand = app.Command(
		"mac-sync",
		"Docker for Mac sync",
	)
	syncStartCommand = syncCommand.Command(
		"start",
		"Start Docker for Mac sync",
	)
	syncStartVolumeArg = syncStartCommand.Arg("volume", "Folders to sync").String()
	syncStopCommand = syncCommand.Command(
		"stop",
		"Stop Docker for Mac sync",
	)
	syncStopVolumeArg = syncStopCommand.Arg("volume", "Folders to sync").String()
	syncStatusCommand = syncCommand.Command(
		"status",
		"Status of Docker for Mac syncs",
	)
)

func isVerbose() bool {
	return *verboseFlag || *dryRunFlag
}

func isDryRun() bool {
	return *dryRunFlag
}

func commandAction(targetFlag string, wrapped func(unitOfWork *UnitOfWork), mightStartRelated bool) {

	cfg = NewConfig(*configFlag, *prefixFlag, *tagFlag)
	allowed = allowedContainers(*excludeFlag, *onlyFlag)
	dependencyMap := cfg.DependencyMap()
	target, err := NewTarget(dependencyMap, targetFlag)
	if err != nil {
		panic(StatusError{err, 78})
	}
	unitOfWork, err := NewUnitOfWork(dependencyMap, target.all())
	if err != nil {
		panic(StatusError{err, 78})
	}

	if isVerbose() {
		printInfof("Command will be applied to: %s", strings.Join(unitOfWork.targeted, ", "))
		if mightStartRelated {
			associated := unitOfWork.Associated()
			if len(associated) > 0 {
				printInfof("\nIf needed, also starts containers: %s", strings.Join(associated, ", "))
			}
			requiredNetworks := unitOfWork.RequiredNetworks()
			if len(requiredNetworks) > 0 {
				printInfof("\nIf needed, also creates networks: %s", strings.Join(requiredNetworks, ", "))
			}
			requiredVolumes := unitOfWork.RequiredVolumes()
			if len(requiredVolumes) > 0 {
				printInfof("\nIf needed, also creates volumes: %s", strings.Join(requiredVolumes, ", "))
			}
		}
		fmt.Printf("\n\n")
	}
	wrapped(unitOfWork)
}

func allowedContainers(excludedReference []string, onlyReference string) (containers []string) {
	allContainers := []string{}
	if len(onlyReference) == 0 {
		for name := range cfg.ContainerMap() {
			allContainers = append(allContainers, name)
		}
	} else {
		allContainers = cfg.ContainersForReference(onlyReference)
	}
	excludedContainers := []string{}
	for _, reference := range excludedReference {
		excludedContainers = append(excludedContainers, cfg.ContainersForReference(reference)...)
	}
	for _, name := range allContainers {
		if !includes(excludedContainers, name) {
			containers = append(containers, name)
		}
	}
	return
}

func runCli() {
	switch kingpin.MustParse(app.Parse(os.Args[1:])) {

	case liftCommand.FullCommand():
		commandAction(*liftTargetArg, func(uow *UnitOfWork) {
			uow.Lift(*liftCmdArg, *liftNoCacheFlag, *liftParallelFlag)
		}, true)

	case versionCommand.FullCommand():
		fmt.Println("v2.9.1")

	case statsCommand.FullCommand():
		commandAction(*statsTargetArg, func(uow *UnitOfWork) {
			uow.Stats(*statsNoStreamFlag)
		}, false)

	case statusCommand.FullCommand():
		commandAction(*statusTargetArg, func(uow *UnitOfWork) {
			uow.Status(*noTruncFlag)
		}, false)

	case pushCommand.FullCommand():
		commandAction(*pushTargetArg, func(uow *UnitOfWork) {
			uow.Push()
		}, false)

	case unpauseCommand.FullCommand():
		commandAction(*unpauseTargetArg, func(uow *UnitOfWork) {
			uow.Unpause()
		}, false)

	case pauseCommand.FullCommand():
		commandAction(*pauseTargetArg, func(uow *UnitOfWork) {
			uow.Pause()
		}, false)

	case startCommand.FullCommand():
		commandAction(*startTargetArg, func(uow *UnitOfWork) {
			uow.Start()
		}, true)

	case stopCommand.FullCommand():
		commandAction(*stopTargetArg, func(uow *UnitOfWork) {
			uow.Stop()
		}, false)

	case killCommand.FullCommand():
		commandAction(*killTargetArg, func(uow *UnitOfWork) {
			uow.Kill()
		}, false)

	case execCommand.FullCommand():
		commandAction(*execTargetArg, func(uow *UnitOfWork) {
			uow.Exec(*execCmdArg)
		}, false)

	case rmCommand.FullCommand():
		commandAction(*rmTargetArg, func(uow *UnitOfWork) {
			uow.Rm(*forceRmFlag)
		}, false)

	case runCommand.FullCommand():
		commandAction(*runTargetArg, func(uow *UnitOfWork) {
			uow.Run(*runCmdArg)
		}, true)

	case createCommand.FullCommand():
		commandAction(*createTargetArg, func(uow *UnitOfWork) {
			uow.Create(*createCmdArg)
		}, true)

	case provisionCommand.FullCommand():
		commandAction(*provisionTargetArg, func(uow *UnitOfWork) {
			uow.Provision(*provisionNoCacheFlag, *provisionParallelFlag)
		}, false)

	case pullCommand.FullCommand():
		commandAction(*pullTargetArg, func(uow *UnitOfWork) {
			uow.PullImage()
		}, false)

	case logsCommand.FullCommand():
		commandAction(*logsTargetArg, func(uow *UnitOfWork) {
			uow.Logs(*followFlag, *timestampsFlag, *tailFlag, *colorizeFlag, *sinceFlag)
		}, false)

	case generateCommand.FullCommand():
		if len(*templateFlag) == 0 {
			printErrorf("ERROR: No template specified. The flag `--template` is required.\n")
			return
		}
		commandAction(*generateTargetArg, func(uow *UnitOfWork) {
			uow.Generate(*templateFlag, *outputFlag)
		}, false)

	case syncStartCommand.FullCommand():
		cfg = NewConfig(*configFlag, *prefixFlag, *tagFlag)
		sync := cfg.MacSync(*syncStartVolumeArg)
		sync.Start()

	case syncStopCommand.FullCommand():
		cfg = NewConfig(*configFlag, *prefixFlag, *tagFlag)
		sync := cfg.MacSync(*syncStopVolumeArg)
		sync.Stop()

	case syncStatusCommand.FullCommand():
		cfg = NewConfig(*configFlag, *prefixFlag, *tagFlag)
		w := new(tabwriter.Writer)
		w.Init(os.Stdout, 0, 8, 1, '\t', 0)
		fmt.Fprintln(w, "VOLUME\tCONTAINER\tSTATUS")
		for _, name := range cfg.MacSyncNames() {
			s := cfg.MacSync(name)
			status := "-"
			if s.Exists() {
				status = "stopped"
				if s.Running() {
					status = "running"
				}
			}
			fmt.Fprintf(w, "%s\n", s.Volume() + "\t" + s.ContainerName() + "\t" + status)
		}
		w.Flush()
	}
}
