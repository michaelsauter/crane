package crane

import (
	"crypto/md5"
	"errors"
	"fmt"
	"os/exec"
	"path"
	"strconv"
	"strings"
)

type UnisonSync interface {
	ContainerName() string
	Volume() string
	Start(bool)
	Stop()
}

type unisonSync struct {
	RawVolume  string
	RawFlags   string `json:"flags" yaml:"flags"`
	RawImage   string `json:"image" yaml:"image"`
	Uid        int    `json:"uid" yaml:"uid"`
	Gid        int    `json:"gid" yaml:"gid"`
	configPath string
	cName      string
	volume     string
}

func (s *unisonSync) ContainerName() string {
	if s.cName == "" {
		syncIdentifier := []byte(s.configPath + ":" + s.Volume())
		digest := fmt.Sprintf("%x", md5.Sum(syncIdentifier))
		s.cName = "crane_unison_" + digest
	}
	return s.cName
}

func (s *unisonSync) Volume() string {
	if s.volume == "" {
		v := expandEnv(s.RawVolume)
		parts := strings.Split(v, ":")
		if !path.IsAbs(parts[0]) {
			parts[0] = s.configPath + "/" + parts[0]
		}
		s.volume = strings.Join(parts, ":")
	}
	return s.volume
}

func (s *unisonSync) Start(sync bool) {
	unisonArgs := []string{}

	// Start sync container if needed
	if containerID(s.ContainerName()) != "" {
		if !inspectBool(s.ContainerName(), "{{.State.Running}}") {
			verboseLog("Starting unison sync for " + s.hostDir())
			dockerArgs := []string{"start", s.ContainerName()}
			executeHiddenCommand("docker", dockerArgs)
			unisonArgs = s.unisonArgs()
		}
	} else {
		verboseLog("Starting unison sync for " + s.hostDir())
		dockerArgs := []string{"run", "--name", s.ContainerName(), "-d", "-P", "-e", "UNISON_DIR=" + s.containerDir(), "-e", "UNISON_UID=" + strconv.Itoa(s.Uid), "-e", "UNISON_GID=" + strconv.Itoa(s.Gid), "-v", s.containerDir(), s.image()}
		executeHiddenCommand("docker", dockerArgs)
		fmt.Printf("Doing initial snyc for %s ...\n", s.hostDir())
		unisonArgs = s.unisonArgs()
		initialSyncArgs := []string{}
		for _, a := range unisonArgs {
			if !strings.HasPrefix(a, "-repeat") {
				initialSyncArgs = append(initialSyncArgs, a)
			}
		}
		executeCommand("unison", initialSyncArgs, nil, nil)
	}

	// Start unison in background if needed
	if sync && len(unisonArgs) > 0 {
		verboseLog("unison " + strings.Join(unisonArgs, " "))
		if !isDryRun() {
			cmd := exec.Command("unison", unisonArgs...)
			cmd.Dir = cfg.Path()
			cmd.Stdout = nil
			cmd.Stderr = nil
			cmd.Stdin = nil
			cmd.Start()
		}
	}
}

func (s *unisonSync) Stop() {
	verboseLog("Stopping unison sync for " + s.hostDir())

	// stop container (also stops Unison sync)
	dockerArgs := []string{"kill", s.ContainerName()}
	executeHiddenCommand("docker", dockerArgs)
}

func (s *unisonSync) unisonArgs() []string {
	unisonArgs := []string{s.hostDir(), "socket://localhost:" + s.publishedPort() + "/"}
	return append(unisonArgs, s.flags()...)
}

func (s *unisonSync) image() string {
	if len(s.RawImage) > 0 {
		return expandEnv(s.RawImage)
	}
	allowedVersions := []string{"2.48.4"}
	versionOut, err := commandOutput("unison", []string{"-version"})
	if err != nil {
		return "michaelsauter/unison:2.48.4"
	}
	versionParts := strings.Split(versionOut, " ")
	installedVersion := versionParts[len(versionParts)-1]
	if !includes(allowedVersions, installedVersion) {
		panic(StatusError{errors.New("Unison version " + installedVersion + " is not supported. You need to install: " + strings.Join(allowedVersions, ", ")), 69})
	}
	return "michaelsauter/unison:" + installedVersion
}

func (s *unisonSync) flags() []string {
	f := "-auto -batch -confirmbigdel=false -repeat=watch"
	if len(s.RawFlags) > 0 {
		f = expandEnv(s.RawFlags)
	}
	return strings.Split(f, " ")
}

func (s *unisonSync) hostDir() string {
	parts := strings.Split(s.Volume(), ":")
	return parts[0]
}

func (s *unisonSync) containerDir() string {
	parts := strings.Split(s.Volume(), ":")
	return parts[1]
}

func (s *unisonSync) publishedPort() string {
	args := []string{"port", s.ContainerName(), "5000/tcp"}
	published, err := commandOutput("docker", args)
	if err != nil {
		printErrorf("Could not detect port of container %s. Sync will not work properly.", s.ContainerName())
		return ""
	}
	parts := strings.Split(published, ":")
	return parts[1]
}

func checkUnisonRequirements() {
	_, err := commandOutput("which", []string{"unison"})
	if err != nil {
		panic(StatusError{errors.New("`unison` is not installed or not in your $PATH.\nSee https://github.com/michaelsauter/crane/wiki/Unison-installation."), 69})
	}

	_, err = commandOutput("which", []string{"unison-fsmonitor"})
	if err != nil {
		panic(StatusError{errors.New("`unison-fsmonitor` is not installed or not in your $PATH.\nSee https://github.com/michaelsauter/crane/wiki/Unison-installation."), 69})
	}
}
