package process

import (
	"sync"

	"github.com/azurity/mini-sos/go-core/node"
	"github.com/azurity/mini-sos/go-core/service"
)

type DummyProcess struct {
	host      node.HostID
	pid       uint32
	quit      sync.WaitGroup
	Providers map[string]func(data []byte, caller service.Provider) ([]byte, error)
}

func (man *Manager) NewDummyProcess() (*DummyProcess, error) {
	pid, err := man.allocate()
	if err != nil {
		return nil, err
	}

	proc := &DummyProcess{
		host:      man.host,
		pid:       pid,
		quit:      sync.WaitGroup{},
		Providers: map[string]func(data []byte, caller service.Provider) ([]byte, error){},
	}
	proc.quit.Add(1)
	return proc, nil
}

func (proc *DummyProcess) Host() node.HostID {
	return proc.host
}

func (proc *DummyProcess) Id() uint32 {
	return proc.pid
}

func (proc *DummyProcess) WaitQuit() {
	proc.quit.Wait()
}

func (proc *DummyProcess) Quit() {
	proc.quit.Done()
}

func (proc *DummyProcess) Kill() {}

func (proc *DummyProcess) ExitCode() uint32 {
	return 0
}

func (proc *DummyProcess) CallService(service string, data []byte, caller service.Provider) ([]byte, error) {
	fn, ok := proc.Providers[service]
	if !ok {
		return nil, ErrNotFound
	}
	return fn(data, caller)
}

func (proc *DummyProcess) Run() error {
	return nil
}
