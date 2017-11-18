package crane

import (
	"os"
)

type Volume interface {
	Name() string
	ActualName() string
	Create()
	Exists() bool
}

type volume struct {
	RawName string
}

func (v *volume) Name() string {
	return expandEnv(v.RawName)
}

func (v *volume) ActualName() string {
	return cfg.Prefix() + v.Name()
}

func (v *volume) Create() {
	printInfof("Creating volume %s ...\n", v.ActualName())

	args := []string{"volume", "create", "--name", v.ActualName()}
	executeCommand("docker", args, os.Stdout, os.Stderr)
}

func (v *volume) Exists() bool {
	args := []string{"volume", "inspect", v.ActualName()}
	_, err := commandOutput("docker", args)
	return err == nil
}
