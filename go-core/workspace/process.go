package workspace

import (
	"context"
	"errors"
	"sync"

	"github.com/azurity/mini-sos/go-core/node"
	"github.com/azurity/mini-sos/go-core/process"
)

var ErrDirectCallRemoteProcess = errors.New("cannot direct call remote process")

type remoteProcess struct {
	man  *remoteProcMan
	pid  uint32
	quit sync.WaitGroup
}

func (proc *remoteProcess) Host() node.HostID {
	return proc.man.node.Host()
}

func (proc *remoteProcess) Id() uint32 {
	return proc.pid
}

func (proc *remoteProcess) WaitQuit() {
	proc.quit.Wait()
}

func (proc *remoteProcess) Kill() {
	// TODO: kill remote proc
}

func (proc *remoteProcess) CallProvider(id uint32, data []byte, caller process.Process) ([]byte, error) {
	return nil, ErrDirectCallRemoteProcess
}

func (proc *remoteProcess) Run() error {
	return nil
}

type remoteProcMan struct {
	workspace    *Workspace
	node         node.Node
	AliveProcess map[uint32]*remoteProcess
	cancel       func()
}

func newRemoteProcMan(ws *Workspace, node node.Node) *remoteProcMan {
	ret := &remoteProcMan{
		workspace:    ws,
		node:         node,
		AliveProcess: map[uint32]*remoteProcess{},
		cancel:       nil,
	}
	go func() {
		ctx, cancel := context.WithCancel(context.Background())
		ret.cancel = cancel
		ret.node.WaitClose(ctx)
		ret.clear()
	}()
	return ret
}

func (man *remoteProcMan) createProcess(pid uint32) {
	ret := &remoteProcess{
		man:  man,
		pid:  pid,
		quit: sync.WaitGroup{},
	}
	ret.quit.Add(1)
	man.AliveProcess[pid] = ret
}

func (man *remoteProcMan) releaseProcess(pid uint32) {
	if proc, ok := man.AliveProcess[pid]; ok {
		delete(man.AliveProcess, pid)
		proc.quit.Done()
	}
}

func (man *remoteProcMan) clear() {
	id := man.node.Host()
	for _, proc := range man.AliveProcess {
		proc.quit.Done()
	}
	delete(man.workspace.parts, id)
}

func (man *remoteProcMan) close() {
	if man.cancel != nil {
		man.cancel()
	}
}
