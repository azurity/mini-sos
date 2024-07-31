package workspace

import (
	"context"
	"errors"
	"sync"

	"github.com/azurity/mini-sos/go-core/node"
	"github.com/azurity/mini-sos/go-core/process"
	"github.com/azurity/mini-sos/go-core/service"
	"github.com/vmihailenco/msgpack/v5"
)

var ErrDirectCallRemoteProcess = errors.New("cannot direct call remote process")

type Part interface {
	host() node.HostID
	close()
	NewProcess(data []byte) (process.Process, error)
}

type LocalPart struct {
	hostId     node.HostID
	service    *service.LocalManager
	processMan *process.Manager
}

func newLocalPart(host node.HostID, totalService service.Manager) (*LocalPart, error) {
	service := service.NewLocalManager()
	man := process.NewManager(host, totalService.Call)
	_, err := man.InitServiceProcess(totalService)
	if err != nil {
		return nil, err
	}
	return &LocalPart{
		hostId:     host,
		service:    service,
		processMan: man,
	}, nil
}

func (part *LocalPart) host() node.HostID {
	return part.hostId
}

func (part *LocalPart) close() {
	for _, proc := range part.processMan.AliveProcess {
		proc.Kill()
	}
}

func (part *LocalPart) NewProcess(data []byte) (process.Process, error) {
	return part.processMan.NewLocalProcess(data)
}

type remoteService struct {
	node      node.Node
	workspace WSID
	services  map[string]uint32
}

func (service *remoteService) Register(entry string, proc service.Provider, local bool) error {
	if local {
		return nil
	}
	data, err := msgpack.Marshal(serviceArg{
		Workspace: service.workspace,
		Process:   proc.Id(),
		Entry:     entry,
	})
	if err != nil {
		return err
	}
	_, err = service.node.Call("registerService", data)
	return err
}

func (service *remoteService) Unregister(entry string, proc service.Provider, local bool) {
	if local {
		return
	}
	data, err := msgpack.Marshal(serviceArg{
		Workspace: service.workspace,
		Process:   proc.Id(),
		Entry:     entry,
	})
	if err != nil {
		return
	}
	service.node.Call("unregisterService", data)
}

func (service *remoteService) List(local bool) []string {
	if local {
		return []string{}
	}
	ret := []string{}
	for it := range service.services {
		ret = append(ret, it)
	}
	return ret
}

func (srv *remoteService) Call(entry string, data []byte, caller service.Provider) ([]byte, error) {
	if _, ok := srv.services[entry]; !ok {
		return nil, service.ErrNotExist
	}
	data, err := msgpack.Marshal(callArg{
		Workspace: srv.workspace,
		Entry:     entry,
		Caller:    caller.Id(),
		Data:      data,
	})
	if err != nil {
		return nil, err
	}
	return srv.node.Call("callService", data)
}

type remoteProcess struct {
	node node.Node
	pid  uint32
	quit sync.WaitGroup
}

func (proc *remoteProcess) Host() node.HostID {
	return proc.node.Host()
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

func (proc *remoteProcess) ExitCode() uint32 {
	return 0
}

func (proc *remoteProcess) CallService(service string, data []byte, caller service.Provider) ([]byte, error) {
	return nil, ErrDirectCallRemoteProcess
}

func (proc *remoteProcess) Run() error {
	return nil
}

type RemotePart struct {
	workspace *Workspace
	node      node.Node
	service   *remoteService
	process   map[uint32]*remoteProcess
	cancel    func()
}

func (part *RemotePart) createProcess(pid uint32) {
	ret := &remoteProcess{
		node: part.node,
		pid:  pid,
		quit: sync.WaitGroup{},
	}
	ret.quit.Add(1)
	part.process[pid] = ret
}

func (part *RemotePart) releaseProcess(pid uint32) {
	if proc, ok := part.process[pid]; ok {
		delete(part.process, pid)
		proc.quit.Done()
	}
}

func (part *RemotePart) close() {
	if part.cancel != nil {
		part.cancel()
	}
}

func (part *RemotePart) clear() {
	id := part.host()
	for _, proc := range part.process {
		proc.quit.Done()
	}
	delete(part.workspace.parts, id)
	delete(part.workspace.service.Instance, id)
}

func (part *RemotePart) start() {
	go func() {
		ctx, cancel := context.WithCancel(context.Background())
		part.cancel = cancel
		part.node.WaitClose(ctx)
		part.clear()
	}()
}

func (workspace *Workspace) newRemotePart(node node.Node) *RemotePart {
	ret := &RemotePart{
		workspace: workspace,
		node:      node,
		service: &remoteService{
			node:      node,
			workspace: workspace.id,
			services:  map[string]uint32{},
		},
		process: map[uint32]*remoteProcess{},
		cancel:  nil,
	}
	ret.start()
	workspace.parts[ret.host()] = ret
	workspace.service.Instance[ret.host()] = ret.service
	return ret
}

func (part *RemotePart) host() node.HostID {
	return part.node.Host()
}
