package crane

import (
	"fmt"
	"os"
)

type Network interface {
	Name() string
	Subnet() string
	ActualName() string
	Create()
	Exists() bool
}

type network struct {
	RawName string
	RawSubnet string `json:"subnet" yaml:"subnet"`
}

func (n *network) Name() string {
	return expandEnv(n.RawName)
}

func (n *network) Subnet() string {
	return expandEnv(n.RawSubnet)
}

func (n *network) ActualName() string {
	return cfg.Prefix() + n.Name()
}

func (n *network) Create() {
	fmt.Printf("Creating network %s ...\n", n.ActualName())

	args := []string{"network", "create"}

	if len(n.Subnet()) > 0 {
		args = append(args, "--subnet", n.Subnet())
	}

	args = append(args, n.ActualName())
	executeCommand("docker", args, os.Stdout, os.Stderr)
}

func (n *network) Exists() bool {
	args := []string{"network", "inspect", n.ActualName()}
	_, err := commandOutput("docker", args)
	return err == nil
}
