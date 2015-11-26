package crane

import (
	"fmt"
	"os"
)

type Volume interface {
	Name() string
	Create()
	Exists() bool
}

type volume struct {
	RawName string
}

func (v *volume) Name() string {
	return os.ExpandEnv(v.RawName)
}

func (v *volume) Create() {
	fmt.Printf("Creating volume %s ...\n", v.Name())

	args := []string{"volume", "create", "--name", v.Name()}
	executeCommand("docker", args)
}

func (v *volume) Exists() bool {
	args := []string{"volume", "inspect", v.Name()}
	_, err := commandOutput("docker", args)
	return err == nil
}
