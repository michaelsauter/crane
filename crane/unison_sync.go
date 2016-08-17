package crane

import (
	"fmt"
	"crypto/md5"
	"strconv"
	"path"
	"strings"
	"os"
)

type UnisonSync interface {
	Start()
	Stop()
}

type unisonSync struct {
	configPath string
	hostDir string
	containerDir string
	cName string
}

func NewUnisonSync(configFlag string, syncArg string) UnisonSync {
	sync := &unisonSync{}

	sync.configPath = path.Dir(findConfig(configFlag))

	parts := strings.Split(syncArg, ":")
	if !path.IsAbs(parts[0]) {
		parts[0] = sync.configPath + "/" + parts[0]
	}
	sync.hostDir = parts[0]
	sync.containerDir = parts[1]
	bindMount := strings.Join(parts, ":")
	sync.cName = unisonSyncContainerName(sync.configPath, bindMount)

	return sync
}

func (s *unisonSync) Start() {
	fmt.Printf("Starting unison sync for %s ...\n", s.hostDir)

	// bring container up
	if containerID(s.cName) != "" {
		if !inspectBool(s.cName, "{{.State.Running}}") {
			dockerArgs := []string{"start", s.cName}
			executeCommand("docker", dockerArgs, os.Stdout, os.Stderr)
		}
	} else {
		dockerArgs := []string{"run", "--name", s.cName, "-d", "-P", "-e", "UNISON_DIR=" + s.containerDir, "-v", s.containerDir, "onnimonni/unison:2.48.4"}
		executeCommand("docker", dockerArgs, os.Stdout, os.Stderr)
	}

	// start unison
	unisonArgs := []string{s.hostDir, "socket://localhost:" + s.publishedPort() + "/", "-auto", "-batch", "-repeat", "watch"}
	executeCommand("unison", unisonArgs, os.Stdout, os.Stderr)
}

func (s *unisonSync) Stop() {
	fmt.Printf("Stopping unison sync for %s ...\n", s.hostDir)

	// stop sync (does not work yet!)
	pgrepArgs := []string{"-f", "\"unison " + s.hostDir + " socket://localhost:" + s.publishedPort() + "/\""}
	spid, _ := commandOutput("pgrep", pgrepArgs)
	ipid, _ := strconv.Atoi(spid)
	p, _ := os.FindProcess(ipid)
	p.Kill()
	// stop container
	dockerArgs := []string{"stop", s.cName}
	executeCommand("docker", dockerArgs, os.Stdout, os.Stderr)
}

func (s *unisonSync) publishedPort() string {
	args := []string{"port", s.cName, "5000/tcp"}
	published, _ := commandOutput("docker", args)
	parts := strings.Split(published, ":")
	return parts[1]
}

func unisonRequirementsMet() bool {
	met := true

	_, err := commandOutput("which", []string{"unison"})
	if err != nil {
		printErrorf("ERROR: Unison is not installed. You need version 2.48.4. Install with:\n  brew install unison\n")
		met = false
	}

	_, err = commandOutput("which", []string{"unison-fsmonitor"})
	if err != nil {
		printErrorf("ERROR: unison-fsmonitor is not installed. Install with:\n  pip install MacFSEvents\n  curl -o /usr/local/bin/unison-fsmonitor -L https://raw.githubusercontent.com/hnsl/unox/master/unox.py\n  chmod +x /usr/local/bin/unison-fsmonitor\n")
		met = false
	}

	return met
}

func unisonSyncContainerName(configPath string, bindMount string) string {
	syncIdentifier := []byte(configPath + ":" + bindMount)
	digest := fmt.Sprintf("%x", md5.Sum(syncIdentifier))
	return "crane_unison_" + digest
}
