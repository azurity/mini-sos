package workspace

import (
	"errors"

	"github.com/azurity/mini-sos/go-core/node"
	"github.com/azurity/mini-sos/go-core/service"
	"github.com/google/uuid"
	"github.com/vmihailenco/msgpack/v5"
)

var ErrSelf = errors.New("cannot remove self node")

type Workspace struct {
	id      WSID
	service *service.GroupManager
	local   *LocalPart
	parts   map[node.HostID]*RemotePart
}

func (man *Manager) NewWorkspace() (*Workspace, error) {
	return man.newWorkspace(uuid.Nil)
}

func (man *Manager) newWorkspace(preAllocatedId WSID) (*Workspace, error) {
	host := man.network.HostId
	id, err := uuid.NewUUID()
	if err != nil {
		return nil, err
	}
	if preAllocatedId != uuid.Nil {
		id = preAllocatedId
	}
	ws := &Workspace{
		id: id,
		service: &service.GroupManager{
			Instance: map[uuid.UUID]service.Manager{},
		},
		local: nil,
		parts: map[node.HostID]*RemotePart{},
	}
	local, err := newLocalPart(host, ws.service)
	if err != nil {
		return nil, err
	}
	ws.local = local
	man.Workspaces[id] = ws

	local.processMan.CreateCallback = func(pid uint32) {
		proc, ok := local.processMan.AliveProcess[pid]
		if !ok {
			return
		}
		arg, err := msgpack.Marshal(processArg{
			Workspace: ws.id,
			Process:   pid,
		})
		if err != nil {
			return
		}
		for _, part := range ws.parts {
			part.node.Call("createProcess", arg)
		}
		go func() {
			proc.WaitQuit()
			for _, part := range ws.parts {
				part.node.Call("releaseProcess", arg)
			}
		}()
	}
	return ws, nil
}

func (workspace *Workspace) Id() WSID {
	return workspace.id
}

func (workspace *Workspace) Local() *LocalPart {
	return workspace.local
}

func (workspace *Workspace) AddNode(n node.Node) {
	workspace.newRemotePart(n)
	arg := extendWorkspaceArg{
		Workspace: workspace.id,
		Parts:     []node.HostID{workspace.local.host()},
	}
	for it := range workspace.parts {
		arg.Parts = append(arg.Parts, it)
	}
	data, _ := msgpack.Marshal(arg)
	n.Call("extendWorkspace", data)
}

func (workspace *Workspace) DelNode(n node.Node) error {
	if n.Host() == workspace.local.host() {
		return ErrSelf
	}
	part, ok := workspace.parts[n.Host()]
	if !ok {
		return ErrUnknownRemote
	}
	for _, part := range workspace.parts {
		arg := extendWorkspaceArg{
			Workspace: workspace.id,
			Parts:     []node.HostID{n.Host()},
		}
		data, _ := msgpack.Marshal(arg)
		part.node.Call("reduceWorkspace", data)
	}
	part.close()
	return nil
}

func (workspace *Workspace) Close() {
	arg, _ := msgpack.Marshal(workspace.id)
	for _, part := range workspace.parts {
		part.node.Call("closeWorkspace", arg)
	}
	workspace.closeImpl()
}

func (workspace *Workspace) closeImpl() {
	workspace.local.close()
	for _, part := range workspace.parts {
		part.close()
	}
}
