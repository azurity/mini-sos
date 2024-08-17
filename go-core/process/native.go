package process

import (
	"context"
	"sync"

	"github.com/azurity/mini-sos/go-core/node"
	"github.com/azurity/mini-sos/go-core/service"
)

type NativeProcess struct {
	host      node.HostID
	pid       uint32
	quit      sync.WaitGroup
	providers *ProviderManager[func(data []byte, caller service.Provider) ([]byte, error)]
	cancel    context.CancelFunc
	run       func() error
	quitFn    func()
}

type NativeOperator struct {
	self *NativeProcess
	man  *Manager
}

func (op *NativeOperator) Call(service string, data []byte) ([]byte, error) {
	return op.man.callFn(service, data, op.self)
}

func (op *NativeOperator) ProviderManager() *ProviderManager[func(data []byte, caller service.Provider) ([]byte, error)] {
	return op.self.providers
}

func (man *Manager) NewNativeProcess(init func(ctx context.Context, op *NativeOperator) func() error) (*NativeProcess, error) {
	pid, err := man.allocate()
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())

	proc := &NativeProcess{
		host:      man.host,
		pid:       pid,
		quit:      sync.WaitGroup{},
		providers: NewProviderManager[func(data []byte, caller service.Provider) ([]byte, error)](),
		cancel:    cancel,
	}
	proc.run = init(ctx, &NativeOperator{self: proc, man: man})

	man.AliveProcess[pid] = proc
	proc.quitFn = func() {
		man.release(pid)
	}
	proc.quit.Add(1)
	if man.CreateCallback != nil {
		man.CreateCallback(proc.pid)
	}
	return proc, nil
}

func (proc *NativeProcess) Host() node.HostID {
	return proc.host
}

func (proc *NativeProcess) Id() uint32 {
	return proc.pid
}

func (proc *NativeProcess) WaitQuit() {
	proc.quit.Wait()
}

func (proc *NativeProcess) Kill() {
	proc.cancel()
}

func (proc *NativeProcess) CallProvider(id uint32, data []byte, caller service.Provider) ([]byte, error) {
	fn := proc.providers.Get(id)
	if fn == nil {
		return nil, ErrNotFound
	}
	return (*fn)(data, caller)
}

func (proc *NativeProcess) Run() error {
	defer proc.quitFn()
	return proc.run()
}

func Dummy(ctx context.Context, op *NativeOperator) func() error {
	return func() error {
		return nil
	}
}
