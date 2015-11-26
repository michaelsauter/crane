package crane

import (
	"fmt"
	"os"
)

type Network interface {
	Name() string
	Create()
	Exists() bool
}

type network struct {
	RawName string
}

func (n *network) Name() string {
	return os.ExpandEnv(n.RawName)
}

func (n *network) Create() {
	fmt.Printf("Creating network %s ...\n", n.Name())

	args := []string{"network", "create", n.Name()}
	executeCommand("docker", args)
}

func (n *network) Exists() bool {
	args := []string{"network", "inspect", n.Name()}
	_, err := commandOutput("docker", args)
	return err == nil
}
