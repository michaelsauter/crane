// +build !pro

package crane

import (
	"errors"
)

type MacSync interface {
	ContainerName() string
	Volume() string
	Autostart() bool
	Exists() bool
	Running() bool
	Start(debug bool)
	Stop()
	Status() string
}

type macSync struct {
	RawVolume    string
	RawFlags     []string `json:"flags" yaml:"flags"`
	RawImage     string   `json:"image" yaml:"image"`
	Uid          int      `json:"uid" yaml:"uid"`
	Gid          int      `json:"gid" yaml:"gid"`
	RawAutostart bool     `json:"autostart" yaml:"autostart"`
	configPath   string
	cName        string
	volume       string
}

func (s *macSync) ContainerName() string {
	return ""
}

func (s *macSync) Volume() string {
	return ""
}

func (s *macSync) Autostart() bool {
	return false
}

func (s *macSync) Exists() bool {
	return false
}

func (s *macSync) Running() bool {
	return false
}

func (s *macSync) Status() string {
	return "-"
}

var proOnly = "Mac Sync is not available in the free version, please purchase the pro version: https://www.craneup.tech"

func (s *macSync) Start(debug bool) {
	panic(StatusError{errors.New(proOnly), 69})
}

func (s *macSync) Stop() {
	panic(StatusError{errors.New(proOnly), 69})
}

func startSync(volumeArg string, debugFlag bool) {
	panic(StatusError{errors.New(proOnly), 69})
}

func stopSync(volumeArg string) {
	panic(StatusError{errors.New(proOnly), 69})
}

func printSyncStatus() {
	panic(StatusError{errors.New(proOnly), 69})
}

func isSyncPossible() bool {
	printInfof("%s\n", proOnly)
	return false
}
