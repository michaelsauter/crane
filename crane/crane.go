package crane

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"syscall"

	"github.com/fatih/color"
	shlex "github.com/flynn/go-shlex"
)

var printSuccessf func(format string, a ...any)
var printInfof func(format string, a ...any)
var printNoticef func(format string, a ...any)
var printErrorf func(format string, a ...any)

func init() {
	color.Output = os.Stderr
	printSuccessf = color.New(color.FgGreen).PrintfFunc()
	printInfof = color.New(color.FgBlue).PrintfFunc()
	printNoticef = color.New(color.FgYellow).PrintfFunc()
	printErrorf = color.New(color.FgRed).PrintfFunc()
}

type StatusError struct {
	error  error
	status int
}

func handleRecoveredError(recovered any) {
	if recovered == nil {
		return
	}

	var statusError StatusError

	switch err := recovered.(type) {
	case StatusError:
		statusError = err
	case error:
		statusError = StatusError{err, 1}
	case string:
		statusError = StatusError{errors.New(err), 1}
	default:
		statusError = StatusError{}
	}

	if statusError.error != nil {
		printErrorf("ERROR: %s\n", statusError.error)
	}
	os.Exit(statusError.status)
}

var requiredDockerVersion = []int{1, 13}

func RealMain() {
	// On panic, recover the error, display it and return the given status code if any
	defer func() {
		handleRecoveredError(recover())
	}()
	checkDockerClient()
	runCli()
}

// Ensure there is a docker binary in the path,
// and printing an error if its version is below the minimal requirement.
func checkDockerClient() {
	output, err := commandOutput("docker", []string{"--version"})
	if err != nil {
		panic(StatusError{errors.New("Error when probing Docker's client version. Is docker installed and within the $PATH?"), 69})
	}
	re := regexp.MustCompile("([0-9]+)\\.([0-9]+)\\.?([0-9]+)?")
	rawVersions := re.FindStringSubmatch(output)
	var versions []int
	for _, rawVersion := range rawVersions[1:] {
		version, err := strconv.Atoi(rawVersion)
		if err != nil {
			printErrorf("Error when parsing Docker's version %v: %v", rawVersion, err)
			break
		}
		versions = append(versions, version)
	}

	for i, expectedVersion := range requiredDockerVersion {
		if versions[i] > expectedVersion {
			break
		}
		if versions[i] < expectedVersion {
			printErrorf("Unsupported client version! Please upgrade to Docker %v or later.\n", intJoin(requiredDockerVersion, "."))
		}
	}
}

// Assemble slice of strings from slice or string with spaces
func stringSlice(sliceLike any) []string {
	var strSlice []string
	switch sl := sliceLike.(type) {
	case string:
		if len(sl) > 0 {
			parts, err := shlex.Split(expandEnv(sl))
			if err != nil {
				printErrorf("Error when parsing cmd `%v`: %v. Proceeding with %q.", sl, err, parts)
			}
			strSlice = append(strSlice, parts...)
		}
	case []any:
		parts := make([]string, len(sl))
		for i, v := range sl {
			parts[i] = expandEnv(fmt.Sprintf("%v", v))
		}
		strSlice = append(strSlice, parts...)
	default:
		panic(StatusError{fmt.Errorf("unknown type: %v", sl), 65})
	}
	return strSlice
}

func equalSlices(a, b []string) bool {

	if a == nil && b == nil {
		return true
	}

	if a == nil || b == nil {
		return false
	}

	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}

// Similar to strings.Join() for int slices.
func intJoin(intSlice []int, sep string) string {
	var stringSlice []string
	for _, v := range intSlice {
		stringSlice = append(stringSlice, fmt.Sprint(v))
	}
	return strings.Join(stringSlice, sep)
}

func executeHook(hook string, containerName string) {
	os.Setenv("CRANE_HOOKED_CONTAINER", containerName)
	cmds, err := shlex.Split(hook)
	if err != nil {
		panic(StatusError{fmt.Errorf("Error when parsing hook `%v`: %v", hook, err), 64})
	}
	switch len(cmds) {
	case 0:
		return
	case 1:
		executeCommand(cmds[0], []string{}, os.Stdout, os.Stderr)
	default:
		executeCommand(cmds[0], cmds[1:], os.Stdout, os.Stderr)
	}
}

// Print message when verbose mode is enabled
func verboseMsg(message string) {
	if isVerbose() {
		printInfof("%s\n", message)
	}
}

// Log command when verbose mode is enabled
func verboseLog(message string) {
	if isVerbose() {
		printInfof("--> %s\n", message)
	}
}

func executeCommand(name string, args []string, stdout, stderr io.Writer) {
	verboseLog(name + " " + strings.Join(args, " "))
	if !isDryRun() {
		cmd := exec.Command(name, args...)
		if cfg != nil {
			cmd.Dir = cfg.Path()
		}
		cmd.Stdout = stdout
		cmd.Stderr = stderr
		cmd.Stdin = os.Stdin
		cmd.Run()
		if !cmd.ProcessState.Success() {
			status := cmd.ProcessState.Sys().(syscall.WaitStatus).ExitStatus()
			panic(StatusError{errors.New(cmd.ProcessState.String()), status})
		}
	}
}

func executeHiddenCommand(name string, args []string) {
	if isVerbose() {
		executeCommand(name, args, os.Stdout, os.Stderr)
	} else {
		executeCommand(name, args, nil, nil)
	}
}

func executeCommandBackground(name string, args []string) (cmd *exec.Cmd, stdout io.ReadCloser, stderr io.ReadCloser) {
	verboseLog(name + " " + strings.Join(args, " "))
	if !isDryRun() {
		cmd = exec.Command(name, args...)
		if cfg != nil {
			cmd.Dir = cfg.Path()
		}
		stdout, _ = cmd.StdoutPipe()
		stderr, _ = cmd.StderrPipe()
		cmd.Start()
		return cmd, stdout, stderr
	}
	return nil, nil, nil
}

func commandOutput(name string, args []string) (string, error) {
	cmd := exec.Command(name, args...)
	if cfg != nil {
		cmd.Dir = cfg.Path()
	}
	out, err := cmd.CombinedOutput()
	return strings.TrimSpace(string(out)), err
}

// Follow os.ExpandEnv's contract except for `$$` which is transformed to `$`
func expandEnv(s string) string {
	os.Setenv("CRANE_DOLLAR", "$")
	return os.ExpandEnv(strings.Replace(s, "$$", "${CRANE_DOLLAR}", -1))
}

func includes(haystack []string, needle string) bool {
	for _, name := range haystack {
		if name == needle {
			return true
		}
	}
	return false
}
