package oci

import (
	"os"
	"time"

	"github.com/iximiuz/conman/pkg/container"
)

// Runtime represents an OCI container runtime interface.
type Runtime interface {
	CreateContainer(
		id container.ID,
		bundleDir string,
		logfile string,
		exitfile string,
		attachfile string,
		stdin bool,
		stdinOnce bool,
		timeout time.Duration,
	) (pid int, err error)
	StartContainer(id container.ID) error
	KillContainer(id container.ID, sig os.Signal) error
	DeleteContainer(id container.ID) error
	ContainerState(container.ID) (StateResp, error)
}

type StateResp struct {
	Id      string `json:"id"`
	Pid     int    `json:"pid"`
	Status  string `json:"status"`
	Created string `json:"created"`
}
