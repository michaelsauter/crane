package crane

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/alecthomas/kingpin"
)

var cfg Config
var allowed []string

var defaultFiles = []string{"docker-compose.yml", "docker-compose.override.yml", "crane.yml", "crane.override.yml"}

var (
	app         = kingpin.New("crane", "Lift containers with ease - https://www.crane-orchestration.com").Interspersed(false).DefaultEnvars()
	verboseFlag = app.Flag("verbose", "Enable verbose output.").Short('v').Bool()
	dryRunFlag  = app.Flag("dry-run", "Dry run (implicitly verbose; no side effects).").Bool()
	configFlag  = app.Flag(
		"config",
		"Location of config file (repeatable).",
	).Short('c').Default(defaultFiles...).PlaceHolder("~/crane.yml").Strings()
	prefixFlag = app.Flag(
		"prefix",
		"Container/Network/Volume prefix.",
	).Short('p').String()
	excludeFlag = app.Flag(
		"exclude",
		"Exclude group or container (repeatable).",
	).Short('x').PlaceHolder("container|group").Strings()
	onlyFlag = app.Flag(
		"only",
		"Limit scope to group or container.",
	).Short('o').PlaceHolder("container|group").String()
	extendFlag = app.Flag(
		"extend",
		"Extend command from target to dependencies.",
	).Short('e').Bool()
	tagFlag = app.Flag(
		"tag",
		"Override image tags.",
	).String()

	upCommand = app.Command(
		"up",
		"Build or pull images if they don't exist, then run or start the containers. Alias of `lift`.",
	)
	upNoCacheFlag = upCommand.Flag(
		"no-cache",
		"Build the image(s) without any cache.",
	).Short('n').Bool()
	upParallelFlag = upCommand.Flag(
		"parallel",
		"Define how many containers are provisioned in parallel.",
	).Short('l').Default("1").Int()
	upDetachFlag = upCommand.Flag(
		"detach",
		"Detach from targeted container.",
	).Short('d').Bool()
	upTargetArg = upCommand.Arg("target", "Target of command").String()
	upCmdArg    = upCommand.Arg("cmd", "Command for container").Strings()

	liftCommand = app.Command(
		"lift",
		"Build or pull images if they don't exist, then run or start the containers. Alias of `up`.",
	)
	liftNoCacheFlag = liftCommand.Flag(
		"no-cache",
		"Build the image(s) without any cache.",
	).Short('n').Bool()
	liftParallelFlag = liftCommand.Flag(
		"parallel",
		"Define how many containers are provisioned in parallel.",
	).Short('l').Default("1").Int()
	liftDetachFlag = liftCommand.Flag(
		"detach",
		"Detach from targeted container.",
	).Short('d').Bool()
	liftTargetArg = liftCommand.Arg("target", "Target of command").String()
	liftCmdArg    = liftCommand.Arg("cmd", "Command for container").Strings()

	runCommand = app.Command(
		"run",
		"Run containers. Already existing containers will be removed first.",
	)
	runDetachFlag = runCommand.Flag(
		"detach",
		"Detach from container.",
	).Short('d').Bool()
	runTargetArg = runCommand.Arg("target", "Target of command").String()
	runCmdArg    = runCommand.Arg("cmd", "Command for container").Strings()

	createCommand = app.Command(
		"create",
		"Create containers. Already existing containers will be removed first.",
	)
	createTargetArg = createCommand.Arg("target", "Target of command").String()
	createCmdArg    = createCommand.Arg("cmd", "Command for container").Strings()

	startCommand = app.Command(
		"start",
		"Start stopped containers. Non-existant containers will be created.",
	)
	startTargetArg = startCommand.Arg("target", "Target of command").String()

	stopCommand = app.Command(
		"stop",
		"Stop running containers.",
	)
	stopTargetArg = stopCommand.Arg("target", "Target of command").String()

	killCommand = app.Command(
		"kill",
		"Kill running containers.",
	)
	killTargetArg = killCommand.Arg("target", "Target of command").String()

	execCommand = app.Command(
		"exec",
		"Execute command in the targeted container(s). Stopped containers will be started, non-existant containers will be created first.",
	)
	execPrivilegedFlag = execCommand.Flag(
		"privileged",
		"Give extended privileges to the process.",
	).Bool()
	execUserFlag = execCommand.Flag(
		"user",
		"Run the command as this user.",
	).String()
	execTargetArg = execCommand.Arg("target", "Target of command").String()
	execCmdArg    = execCommand.Arg("cmd", "Command for container").Strings()

	rmCommand = app.Command(
		"rm",
		"Remove stopped containers.",
	)
	rmForceFlag = rmCommand.Flag(
		"force",
		"Remove running containers, too.",
	).Short('f').Bool()
	rmVolumesFlag = rmCommand.Flag(
		"volumes",
		"Remove volumes as well.",
	).Bool()
	rmTargetArg = rmCommand.Arg("target", "Target of command").String()

	pauseCommand = app.Command(
		"pause",
		"Pause running containers.",
	)
	pauseTargetArg = pauseCommand.Arg("target", "Target of command").String()

	unpauseCommand = app.Command(
		"unpause",
		"Unpause paused containers.",
	)
	unpauseTargetArg = unpauseCommand.Arg("target", "Target of command").String()

	provisionCommand = app.Command(
		"provision",
		"Build or pull images.",
	)
	provisionNoCacheFlag = provisionCommand.Flag(
		"no-cache",
		"Build the image(s) without any cache.",
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

	pushCommand = app.Command(
		"push",
		"Push containers to the registry.",
	)
	pushTargetArg = pushCommand.Arg("target", "Target of command").String()

	logsCommand = app.Command(
		"logs",
		"Show container logs.",
	)
	followFlag = logsCommand.Flag(
		"follow",
		"Follow log output.",
	).Short('f').Bool()
	tailFlag = logsCommand.Flag(
		"tail",
		"Define number of lines to display at the end of the logs.",
	).String()
	timestampsFlag = logsCommand.Flag(
		"timestamps",
		"Show timestamps.",
	).Short('t').Bool()
	colorizeFlag = logsCommand.Flag(
		"colorize",
		"Use different color for each container.",
	).Short('z').Bool()
	sinceFlag = logsCommand.Flag(
		"since",
		"Show logs since timestamp.",
	).String()
	logsTargetArg = logsCommand.Arg("target", "Target of command").String()

	statsCommand = app.Command(
		"stats",
		"Display statistics about containers.",
	)
	statsNoStreamFlag = statsCommand.Flag(
		"no-stream",
		"Disable stats streaming.",
	).Short('n').Bool()
	statsTargetArg = statsCommand.Arg("target", "Target of command").String()

	statusCommand = app.Command(
		"status",
		"Display status of containers (similar to `docker ps`).",
	)
	noTruncFlag = statusCommand.Flag(
		"no-trunc",
		"Don't truncate output.",
	).Short('n').Bool()
	statusTargetArg = statusCommand.Arg("target", "Target of command").String()

	cmdCommand = app.Command(
		"cmd",
		"Execute predefined shortcut command.",
	)
	cmdCmdArg       = cmdCommand.Arg("command", "Command to execute").String()
	cmdArgumentsArg = cmdCommand.Arg("arguments", "Arguments for the command").Strings()

	generateCommand = app.Command(
		"generate",
		"Generate files by passing the config to given template.",
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

	amCommand = app.Command(
		"am",
		"Sub-commands for accelerated mounts.",
	)
	amResetCommand = amCommand.Command(
		"reset",
		"Reset accelerated mount.",
	)
	amResetTargetArg = amResetCommand.Arg("target", "Target of command").String()
	amLogsCommand    = amCommand.Command(
		"logs",
		"Show logs of accelerated mount.",
	)
	amFollowFlag = amLogsCommand.Flag(
		"follow",
		"Follow log output.",
	).Short('f').Bool()
	amLogsTargetArg = amLogsCommand.Arg("target", "Target of command").String()

	versionCommand = app.Command(
		"version",
		"Display the current version.",
	)
)

func isVerbose() bool {
	return *verboseFlag || *dryRunFlag
}

func isDryRun() bool {
	return *dryRunFlag
}

func commandAction(targetArg string, wrapped func(unitOfWork *UnitOfWork), mightStartRelated bool) {

	cfg = NewConfig(*configFlag, *prefixFlag, *tagFlag)
	allowed = allowedContainers(*excludeFlag, *onlyFlag)
	dependencyMap := cfg.DependencyMap()
	target, err := NewTarget(dependencyMap, targetArg, *extendFlag)
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
	command := kingpin.MustParse(app.Parse(os.Args[1:]))

	switch command {
	case cmdCommand.FullCommand():
		cfg = NewConfig(*configFlag, *prefixFlag, *tagFlag)
		printCmds := func() {
			cmds := cfg.Cmds()
			cmdNames := []string{}
			for name, _ := range cmds {
				cmdNames = append(cmdNames, name)
			}
			sort.Strings(cmdNames)
			for _, name := range cmdNames {
				fmt.Fprintf(os.Stdout, "%s: ", name)
				printInfof("%s\n", strings.Join(cmds[name], " "))
			}
		}
		if len(*cmdCmdArg) == 0 {
			printCmds()
			return
		}
		definedCmd := cfg.Cmd(*cmdCmdArg)
		if definedCmd == nil {
			printErrorf("No such command: %s\n\n", *cmdCmdArg)
			fmt.Fprintf(os.Stdout, "Available commands:\n")
			printCmds()
			return
		}
		args := []string{}
		if *verboseFlag {
			args = append(args, "--verbose")
		}
		if !equalSlices(*configFlag, defaultFiles) {
			for _, conf := range *configFlag {
				args = append(args, "--config", conf)
			}
		}
		if len(*prefixFlag) > 0 {
			args = append(args, "--prefix", *prefixFlag)
		}
		if len(*tagFlag) > 0 {
			args = append(args, "--tag", *tagFlag)
		}
		args = append(args, definedCmd...)
		args = append(args, *cmdArgumentsArg...)
		executeCommand("crane", args, os.Stdout, os.Stderr)

	case upCommand.FullCommand():
		commandAction(*upTargetArg, func(uow *UnitOfWork) {
			uow.Up(*upCmdArg, *upDetachFlag, *upNoCacheFlag, *upParallelFlag)
		}, true)

	case liftCommand.FullCommand():
		commandAction(*liftTargetArg, func(uow *UnitOfWork) {
			uow.Up(*liftCmdArg, *liftDetachFlag, *liftNoCacheFlag, *liftParallelFlag)
		}, true)

	case versionCommand.FullCommand():
		printVersion()

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
			uow.Exec(*execCmdArg, *execPrivilegedFlag, *execUserFlag)
		}, false)

	case rmCommand.FullCommand():
		commandAction(*rmTargetArg, func(uow *UnitOfWork) {
			uow.Rm(*rmForceFlag, *rmVolumesFlag)
		}, false)

	case runCommand.FullCommand():
		commandAction(*runTargetArg, func(uow *UnitOfWork) {
			uow.Run(*runCmdArg, *runDetachFlag)
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

	case amResetCommand.FullCommand():
		cfg = NewConfig(*configFlag, *prefixFlag, *tagFlag)
		resetTargets := []string{}
		container := cfg.Container(*amResetTargetArg)
		if container != nil {
			for _, bm := range container.BindMounts(cfg.VolumeNames()) {
				am := &acceleratedMount{RawVolume: bm, configPath: cfg.Path()}
				resetTargets = append(resetTargets, am.Volume())
			}
		} else {
			resetTargets = append(resetTargets, *amResetTargetArg)
		}
		for _, resetTarget := range resetTargets {
			am := cfg.AcceleratedMount(resetTarget)
			if accelerationEnabled() && am != nil {
				printInfof("Resetting accelerated mount %s ...\n", resetTarget)
				am.Reset()
			} else {
				amNames := cfg.AcceleratedMountNames()
				if len(amNames) > 0 {
					printErrorf("ERROR: No such accelerated mount. Configured mounts: %s\n", strings.Join(amNames, ", "))
				} else {
					printErrorf("ERROR: No accelerated mounts configured.\n")
				}
			}
		}

	case amLogsCommand.FullCommand():
		cfg = NewConfig(*configFlag, *prefixFlag, *tagFlag)
		var logsTarget string
		configuredAcceleratedMounts := cfg.AcceleratedMountNames()

		container := cfg.Container(*amLogsTargetArg)
		if container != nil {
			allBindMounts := container.BindMounts(cfg.VolumeNames())
			acceleratedBindMounts := []string{}
			for _, name := range allBindMounts {
				if includes(configuredAcceleratedMounts, name) {
					acceleratedBindMounts = append(acceleratedBindMounts, name)
				}
			}
			am := &acceleratedMount{RawVolume: acceleratedBindMounts[0], configPath: cfg.Path()}
			logsTarget = am.Volume()
			if len(acceleratedBindMounts) > 1 {
				printNoticef("WARNING: %s has more than one bind-mount configured. The first one will be selected. To select a different bind-mount, pass it directly.\n", *amLogsTargetArg)
			}
		} else {
			logsTarget = *amLogsTargetArg
		}
		am := cfg.AcceleratedMount(logsTarget)
		if accelerationEnabled() && am != nil {
			kind := "Showing"
			if *amFollowFlag {
				kind = "Following"
			}
			printInfof("%s logs of accelerated mount %s ...\n", kind, *amLogsTargetArg)
			am.Logs(*amFollowFlag)
		} else {
			if len(configuredAcceleratedMounts) > 0 {
				printErrorf("ERROR: No such accelerated mount. Configured mounts: %s\n", strings.Join(configuredAcceleratedMounts, ", "))
			} else {
				printErrorf("ERROR: No accelerated mounts configured.\n")
			}
		}
	}
}
