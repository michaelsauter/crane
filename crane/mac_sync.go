package crane

import (
	"crypto/md5"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"time"
)

type MacSync interface {
	ContainerName() string
	Volume() string
	Exists() bool
	Running() bool
	Start()
	Stop()
}

type macSync struct {
	RawVolume  string
	RawFlags   []string `json:"flags" yaml:"flags"`
	RawImage   string   `json:"image" yaml:"image"`
	Uid        int      `json:"uid" yaml:"uid"`
	Gid        int      `json:"gid" yaml:"gid"`
	configPath string
	cName      string
	volume     string
}

func (s *macSync) ContainerName() string {
	if s.cName == "" {
		syncIdentifier := []byte(s.configPath + ":" + s.Volume())
		digest := fmt.Sprintf("%x", md5.Sum(syncIdentifier))
		s.cName = "crane_unison_" + digest
	}
	return s.cName
}

func (s *macSync) Volume() string {
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

func (s *macSync) Exists() bool {
	return containerID(s.ContainerName()) != ""
}

func (s *macSync) Running() bool {
	return s.Exists() && inspectBool(s.ContainerName(), "{{.State.Running}}")
}

func (s *macSync) Start() {
	unisonArgs := []string{}

	// Start sync container if needed
	if s.Exists() {
		if !s.Running() {
			checkUnisonRequirements()
			verboseLog("Starting unison sync for " + s.hostDir())
			dockerArgs := []string{"start", s.ContainerName()}
			executeHiddenCommand("docker", dockerArgs)
			unisonArgs = s.unisonArgs()
		} else {
			verboseLog("Unison sync for " + s.hostDir() + " already running")
		}
	} else {
		checkUnisonRequirements()
		verboseLog("Starting unison sync for " + s.hostDir())
		dockerArgs := []string{
			"run",
			"--name", s.ContainerName(),
			"-d",
			"-P",
			"-e", "UNISON_DIR=" + s.containerDir(),
			"-e", "UNISON_UID=" + strconv.Itoa(s.Uid),
			"-e", "UNISON_GID=" + strconv.Itoa(s.Gid),
			"-v", s.containerDir(),
			s.image(),
		}
		executeHiddenCommand("docker", dockerArgs)
		fmt.Printf("Doing initial snyc for %s ...\n", s.hostDir())
		unisonArgs = s.unisonArgs()
		initialSyncArgs := []string{}
		for _, a := range unisonArgs {
			if !strings.HasPrefix(a, "-repeat") {
				initialSyncArgs = append(initialSyncArgs, a)
			}
		}
		// Wait a bit for the Unison server to start
		time.Sleep(time.Second)
		executeCommand("unison", initialSyncArgs, nil, os.Stderr)
	}

	// Start unison in background if not already running
	if len(unisonArgs) > 0 {
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

func (s *macSync) Stop() {
	verboseLog("Stopping unison sync for " + s.hostDir())

	// stop container (also stops Unison sync)
	dockerArgs := []string{"kill", s.ContainerName()}
	executeHiddenCommand("docker", dockerArgs)
}

func (s *macSync) unisonArgs() []string {
	unisonArgs := []string{s.hostDir(), "socket://localhost:" + s.publishedPort() + "/"}
	return append(unisonArgs, s.flags()...)
}

func (s *macSync) image() string {
	if len(s.RawImage) > 0 {
		return expandEnv(s.RawImage)
	}
	allowedVersions := []string{"2.48.4"}
	versionOut, err := commandOutput("unison", []string{"-version"})
	if err != nil {
		return "michaelsauter/unison:2.48.4"
	}
	// `unison -version` returns sth like "unison version 2.48.4"
	versionParts := strings.Split(versionOut, " ")
	installedVersion := versionParts[len(versionParts)-1]
	if !includes(allowedVersions, installedVersion) {
		panic(StatusError{errors.New("Unison version " + installedVersion + " is not supported. You need to install: " + strings.Join(allowedVersions, ", ")), 69})
	}
	return "michaelsauter/unison:" + installedVersion
}

func (s *macSync) flags() []string {
	if len(s.RawFlags) > 0 {
		f := []string{}
		for _, rawFlag := range s.RawFlags {
			f = append(f, expandEnv(rawFlag))
		}
		return f
	}
	return []string{"-auto", "-batch", "-ignore=Name {.git}", "-confirmbigdel=false", "-prefer=newer", "-repeat=watch"}
}

func (s *macSync) hostDir() string {
	parts := strings.Split(s.Volume(), ":")
	return parts[0]
}

func (s *macSync) containerDir() string {
	parts := strings.Split(s.Volume(), ":")
	return parts[1]
}

func (s *macSync) publishedPort() string {
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
