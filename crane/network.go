package crane

import (
	"fmt"
	"os"
)

type Network interface {
	Name() string
	ActualName() string
	Create()
	Exists() bool
}

type network struct {
	RawName string
}

func (n *network) Name() string {
	return os.ExpandEnv(n.RawName)
}

func (n *network) ActualName() string {
	return cfg.Prefix() + n.Name()
}

func (n *network) Create() {
	fmt.Printf("Creating network %s ...\n", n.ActualName())

	args := []string{"network", "create", n.ActualName()}
	executeCommand("docker", args)
}

func (n *network) Exists() bool {
	args := []string{"network", "inspect", n.ActualName()}
	_, err := commandOutput("docker", args)
	return err == nil
}
