package crane

import (
	"fmt"
	"os"
)

type VolumeCommander interface {
	Name() string
	ActualName() string
	Create()
	Exists() bool
}

type Volume struct {
	RawName string
}

func (v *Volume) Name() string {
	return expandEnv(v.RawName)
}

func (v *Volume) ActualName() string {
	return cfg.Prefix() + v.Name()
}

func (v *Volume) Create() {
	fmt.Printf("Creating volume %s ...\n", v.ActualName())

	args := []string{"volume", "create", "--name", v.ActualName()}
	executeCommand("docker", args, os.Stdout, os.Stderr)
}

func (v *Volume) Exists() bool {
	args := []string{"volume", "inspect", v.ActualName()}
	_, err := commandOutput("docker", args)
	return err == nil
}
