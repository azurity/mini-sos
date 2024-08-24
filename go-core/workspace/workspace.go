package workspace

import (
	"errors"

	"github.com/azurity/mini-sos/go-core/node"
	"github.com/azurity/mini-sos/go-core/process"
	"github.com/azurity/mini-sos/go-core/service"
	"github.com/google/uuid"
	"github.com/vmihailenco/msgpack/v5"
)

var ErrSelf = errors.New("cannot remove self node")

type Workspace struct {
	id        WSID
	man       *Manager
	processes *process.Manager
	service   *service.Manager
	local     node.HostID
	parts     map[node.HostID]*remoteProcMan
}

func (man *Manager) NewWorkspace() (*Workspace, error) {
	return man.newWorkspace(uuid.Nil)
}

func (man *Manager) newWorkspace(preAllocatedId WSID) (*Workspace, error) {
	host := man.network.HostId
	id, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}
	if preAllocatedId != uuid.Nil {
		id = preAllocatedId
	}
	ws := &Workspace{
		id:        id,
		man:       man,
		processes: nil,
		service:   nil,
		local:     preAllocatedId,
		parts:     map[node.HostID]*remoteProcMan{},
	}
	ws.service = service.NewManager(host)
	ws.processes = process.NewManager(host, ws.service.CallService)
	if err := ws.service.Init(ws.processes, ws.callRemoteProvider, ws.callBroadcast, ws.newRemoteProcess, func(path string) bool {
		return ws.lockAction(path, true)
	}, func(path string) bool {
		return ws.lockAction(path, false)
	}); err != nil {
		return nil, err
	}
	man.Workspaces[id] = ws
	return ws, nil
}

func (workspace *Workspace) callBroadcast(entry string, cap uuid.UUID, data []byte, caller service.CallerInfo) {
	arg := broadcastArg{
		Workspace: workspace.id,
		Entry:     entry,
		Cap:       cap,
		Caller:    caller.Process,
		Data:      data,
	}
	raw, _ := msgpack.Marshal(arg)
	for id := range workspace.parts {
		if node, ok := workspace.man.network.Nodes[id]; ok {
			node.Call("callBroadcast", raw)
		}
	}
}

func (workspace *Workspace) lockAction(lock string, action bool) bool {
	arg := lockArg{
		Workspace: workspace.id,
		Lock:      lock,
		Action:    action,
	}
	data, _ := msgpack.Marshal(arg)
	ret := true
	for id := range workspace.parts {
		if node, ok := workspace.man.network.Nodes[id]; ok {
			res, err := node.Call("lockAction", data)
			if err != nil {
				return false
			}
			var result bool
			msgpack.Unmarshal(res, &result)
			ret = result
		} else {
			ret = false
		}
	}
	return ret
}

func (workspace *Workspace) callRemoteProvider(host node.HostID, process uint32, provider uint32, data []byte, caller service.CallerInfo) ([]byte, error) {
	if _, ok := workspace.parts[host]; !ok {
		return []byte{}, ErrUnknownRemote
	}
	if node, ok := workspace.man.network.Nodes[host]; ok {
		arg := callArg{
			Workspace: workspace.id,
			Process:   process,
			Provider:  provider,
			Caller:    caller.Process,
			Data:      data,
		}
		data, _ := msgpack.Marshal(arg)
		return node.Call("callProvider", data)
	}
	return []byte{}, ErrUnknownRemote
}

func (workspace *Workspace) newRemoteProcess(caller service.CallerInfo) (process.Process, error) {
	procMan, ok := workspace.parts[caller.Host]
	if !ok {
		return nil, ErrUnknownRemote
	}
	if proc, ok := procMan.AliveProcess[caller.Process]; ok {
		return proc, nil
	}
	return nil, ErrUnknownProcess
}

func (workspace *Workspace) Id() WSID {
	return workspace.id
}

func (workspace *Workspace) ProcMan() *process.Manager {
	return workspace.processes
}

func (workspace *Workspace) AddNode(n node.Node) {
	if _, ok := workspace.parts[n.Host()]; ok || n.Host() == workspace.local {
		return
	}
	workspace.parts[n.Host()] = &remoteProcMan{}
	arg := extendWorkspaceArg{
		Workspace: workspace.id,
		Parts:     []node.HostID{workspace.local},
	}
	for it := range workspace.parts {
		arg.Parts = append(arg.Parts, it)
	}
	data, _ := msgpack.Marshal(arg)
	n.Call("extendWorkspace", data)
}

func (workspace *Workspace) DelNode(n node.Node) error {
	if n.Host() == workspace.local {
		return ErrSelf
	}
	if _, ok := workspace.parts[n.Host()]; !ok {
		return ErrUnknownRemote
	}
	arg := extendWorkspaceArg{
		Workspace: workspace.id,
		Parts:     []node.HostID{n.Host()},
	}
	data, _ := msgpack.Marshal(arg)
	for part := range workspace.parts {
		if node, ok := workspace.man.network.Nodes[part]; ok {
			node.Call("reduceWorkspace", data)
		}
	}
	delete(workspace.parts, n.Host())
	return nil
}

func (workspace *Workspace) Close() {
	arg, _ := msgpack.Marshal(workspace.id)
	for part := range workspace.parts {
		if node, ok := workspace.man.network.Nodes[part]; ok {
			node.Call("closeWorkspace", arg)
		}
	}
	workspace.closeImpl()
	delete(workspace.man.Workspaces, workspace.id)
}

func (workspace *Workspace) closeImpl() {
	for _, proc := range workspace.processes.AliveProcess {
		proc.Kill()
	}
	for _, man := range workspace.parts {
		man.close()
	}
}
