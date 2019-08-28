package cri

import (
	"syscall"
	"time"

	"github.com/pkg/errors"

	"github.com/iximiuz/conman/pkg/container"
	"github.com/iximiuz/conman/pkg/oci"
	"github.com/iximiuz/conman/pkg/rollback"
	"github.com/iximiuz/conman/pkg/storage"
)

// RuntimeService is a CRI-alike service
// to manage container & sandbox runtimes.
type RuntimeService interface {
	CreateContainer(ContainerOptions) (*container.Container, error)
	StartContainer(container.ID) error
	StopContainer(id container.ID, timeout time.Duration) error
	// +RemoveContainer(container.ID) error
	// +ListContainers
	// +ContainerStatus
	// UpdateContainerResources
	// ReopenContainerLog
	// ExecSync
	// Exec
	// Attach

	// RunPodSandbox
	// StopPodSandbox
	// RemovePodSandbox
	// PodSandboxStatus
	// ListPodSandbox
}

type runtimeService struct {
	runtime oci.Runtime
	cmap    *container.Map
	cstore  storage.ContainerStore
}

func NewRuntimeService(
	runtime oci.Runtime,
	cstore storage.ContainerStore,
) RuntimeService {
	return &runtimeService{
		runtime: runtime,
		cmap:    container.NewMap(),
		cstore:  cstore,
	}
}

func (s *runtimeService) CreateContainer(
	opts ContainerOptions,
) (cont *container.Container, err error) {
	rb := rollback.New()
	defer func() { _ = err == nil || rb.Execute() }()

	cont, err = container.New(
		container.RandID(),
		opts.Name,
	)
	if err != nil {
		return
	}

	if err = s.cmap.Add(cont, rb); err != nil {
		return
	}

	h, err := s.cstore.CreateContainer(cont, rb)
	if err != nil {
		return
	}

	spec, err := oci.NewSpec(oci.SpecOptions{
		Command:      opts.Command,
		Args:         opts.Args,
		RootPath:     h.RootfsDir(),
		RootReadonly: opts.RootfsReadonly,
	})
	if err != nil {
		return
	}

	err = s.cstore.CreateContainerBundle(cont.ID(), spec, opts.RootfsPath)
	if err != nil {
		return
	}

	err = s.runtime.CreateContainer(cont.ID(), h.BundleDir())
	return
}

func (r *runtimeService) StartContainer(
	id container.ID,
) error {
	cont := r.cmap.Get(id)
	if cont == nil {
		return errors.New("container not found")
	}
	return r.runtime.StartContainer(cont.ID())
}

func (r *runtimeService) StopContainer(
	id container.ID,
	timeout time.Duration,
) error {
	cont := r.cmap.Get(id)
	if cont == nil {
		return errors.New("container not found")
	}

	return r.runtime.KillContainer(cont.ID(), syscall.SIGTERM)
	// TODO: wait for `timeout` ms. If the container proc is still there
	// r.runtime.KillContainer(cont.ID(), syscall.SIGKILL)
	// wait for some default timeout. If the container proc is still there
	// kill(PID)
}

type ContainerOptions struct {
	Name           string
	Command        string
	Args           []string
	RootfsPath     string
	RootfsReadonly bool
}
