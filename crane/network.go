package crane

import (
	"fmt"
	"os"
)

type NetworkCommander interface {
	Name() string
	ActualName() string
	Create()
	Exists() bool
}

type Network struct {
	RawName string
}

func (n *Network) Name() string {
	return expandEnv(n.RawName)
}

func (n *Network) ActualName() string {
	return cfg.Prefix() + n.Name()
}

func (n *Network) Create() {
	fmt.Printf("Creating network %s ...\n", n.ActualName())

	args := []string{"network", "create", n.ActualName()}
	executeCommand("docker", args, os.Stdout, os.Stderr)
}

func (n *Network) Exists() bool {
	args := []string{"network", "inspect", n.ActualName()}
	_, err := commandOutput("docker", args)
	return err == nil
}
