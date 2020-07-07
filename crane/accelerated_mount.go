package crane

import (
	"crypto/md5"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"
)

type AcceleratedMount interface {
	Run()
	Reset()
	Logs(follow bool)
	VolumeArg() string
	Volume() string
}

type acceleratedMount struct {
	RawVolume  string
	RawIgnore  string   `json:"ignore" yaml:"ignore"`
	RawFlags   []string `json:"flags" yaml:"flags"`
	Uid        int      `json:"uid" yaml:"uid"`
	Gid        int      `json:"gid" yaml:"gid"`
	configPath string
	volume     string
	_digest    string
}

func (am *acceleratedMount) Volume() string {
	if am.volume == "" {
		am.volume = expandEnv(am.RawVolume)
	}
	return am.volume
}

func (am *acceleratedMount) Run() {
	if !am.running() {
		am.ensureDataVolume()
		am.initialSync()
		am.continuousSync()
	}
}

func (am *acceleratedMount) VolumeArg() string {
	return am.dataVolumeName() + ":" + am.bindMountContainerPart()
}

func (am *acceleratedMount) Reset() {
	args := []string{"rm", "-f", am.syncContainerName()}
	executeCommand("docker", args, os.Stdout, os.Stderr)
	args = []string{"volume", "rm", am.dataVolumeName()}
	executeCommand("docker", args, os.Stdout, os.Stderr)
}

func (am *acceleratedMount) Logs(follow bool) {
	args := []string{"logs"}
	if follow {
		args = append(args, "-f")
	}
	args = append(args, am.syncContainerName())
	executeCommand("docker", args, os.Stdout, os.Stderr)
}

func (am *acceleratedMount) running() bool {
	return am.syncContainerExists() && inspectBool(am.syncContainerName(), "{{.State.Running}}")
}

func (am *acceleratedMount) ensureDataVolume() {
	if !am.dataVolumeExists() {
		printInfof("Creating volume %s ...\n", am.dataVolumeName())
		args := []string{
			"volume",
			"create",
			"--name", am.dataVolumeName(),
			"--label", "com.crane-orchestration.accelerated-mount=" + am.Volume(),
		}
		executeHiddenCommand("docker", args)
	}
}

func (am *acceleratedMount) initialSync() {
	dockerArgs := []string{"run", "--rm"}
	dockerArgs = append(dockerArgs, "-e", "UNISON_CHOWN=1")
	dockerArgs = append(dockerArgs, am.syncContainerArgs()...)
	dockerArgs = append(dockerArgs, am.initialFlags()...)
	dockerArgs = append(dockerArgs, "/bind-mount", "/data-volume")
	printInfof("Doing initial sync for %s ... this might take a while\n", am.bindMountHostPart())
	executeHiddenCommand("docker", dockerArgs)
}

func (am *acceleratedMount) continuousSync() {
	if am.syncContainerExists() {
		dockerArgs := []string{"start", am.syncContainerName()}
		printInfof("Starting sync for %s via %s ...\n", am.bindMountHostPart(), am.syncContainerName())
		executeCommand("docker", dockerArgs, nil, nil)
	} else {
		dockerArgs := []string{"run", "--name", am.syncContainerName(), "-d"}
		dockerArgs = append(dockerArgs, am.syncContainerArgs()...)
		dockerArgs = append(dockerArgs, am.continuousFlags()...)
		dockerArgs = append(dockerArgs, "/bind-mount", "/data-volume")
		printInfof("Starting sync for %s via %s ...\n", am.bindMountHostPart(), am.syncContainerName())
		executeCommand("docker", dockerArgs, nil, nil)
	}
}

func (am *acceleratedMount) dataVolumeExists() bool {
	args := []string{"volume", "inspect", am.dataVolumeName()}
	_, err := commandOutput("docker", args)
	return err == nil
}

// If flags is given in the config, its value is used.
// Otherwise, we check if ingore is configured, and use
// its value for the -ignore flag. Otherwise the default
// flags are sent to Unison.
func (am *acceleratedMount) flags() []string {
	if len(am.RawFlags) > 0 {
		f := []string{}
		for _, rawFlag := range am.RawFlags {
			f = append(f, expandEnv(rawFlag))
		}
		return f
	}

	ignore := "Name {.git}"
	if len(am.RawIgnore) > 0 {
		ignore = expandEnv(am.RawIgnore)
	}
	ignoreFlag := fmt.Sprintf("-ignore='%s'", ignore)

	return []string{"-auto", "-batch", ignoreFlag, "-contactquietly", "-confirmbigdel=false", "-prefer=newer"}
}

func (am *acceleratedMount) initialFlags() []string {
	return append(am.flags(), "-ignorearchives")
}

func (am *acceleratedMount) continuousFlags() []string {
	return append(am.flags(), "-repeat=watch")
}

func (am *acceleratedMount) syncContainerExists() bool {
	return containerID(am.syncContainerName()) != ""
}

func (am *acceleratedMount) syncContainerArgs() []string {
	return []string{
		"-v", am.bindMountHostPart() + ":/bind-mount",
		"-v", am.dataVolumeName() + ":/data-volume",
		"-e", "UNISON_UID=" + strconv.Itoa(am.Uid),
		"-e", "UNISON_GID=" + strconv.Itoa(am.Gid),
		"--label", "com.crane-orchestration.accelerated-mount=" + am.Volume(),
		am.image(),
	}
}

func (am *acceleratedMount) dataVolumeName() string {
	return "crane_vol_" + am.digest()
}

func (am *acceleratedMount) syncContainerName() string {
	return "crane_sync_" + am.digest()
}

func (am *acceleratedMount) digest() string {
	if am._digest == "" {
		syncIdentifierParts := []string{
			am.configPath,
			am.Volume(),
			am.image(),
			strings.Join(am.flags(), " "),
			strconv.Itoa(am.Uid),
			strconv.Itoa(am.Gid),
		}
		syncIdentifier := []byte(strings.Join(syncIdentifierParts, ":"))
		am._digest = fmt.Sprintf("%x", md5.Sum(syncIdentifier))
	}
	return am._digest
}

func (am *acceleratedMount) image() string {
	return "michaelsauter/crane-sync:3.2.0"
}

func (am *acceleratedMount) bindMountHostPart() string {
	parts := strings.Split(actualVolumeArg(am.Volume()), ":")
	return parts[0]
}

func (am *acceleratedMount) bindMountContainerPart() string {
	parts := strings.Split(actualVolumeArg(am.Volume()), ":")
	return parts[1]
}

func accelerationEnabled() bool {
	return runtime.GOOS == "darwin" || runtime.GOOS == "windows"
}
