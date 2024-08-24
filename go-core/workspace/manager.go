package workspace

import (
	"errors"

	"github.com/azurity/mini-sos/go-core/node"
	"github.com/azurity/mini-sos/go-core/utils/tree"
	"github.com/google/uuid"
	"github.com/vmihailenco/msgpack/v5"
)

var ErrFullOfWorkspace = errors.New("full of workspace")
var ErrUnknownWorkspace = errors.New("unknown workspace")
var ErrUnknownRemote = errors.New("unknown remote")
var ErrUnknownProcess = errors.New("unknown process")

type WSID = uuid.UUID

type Manager struct {
	Workspaces map[WSID]*Workspace
	network    *node.Manager
}

type processArg struct {
	Workspace WSID   `msgpack:"workspace"`
	Process   uint32 `msgpack:"process"`
}

type extendWorkspaceArg struct {
	Workspace WSID          `msgpack:"workspace"`
	Parts     []node.HostID `msgpack:"parts"`
}

type serviceArg struct {
	Workspace WSID           `msgpack:"workspace"`
	Desc      *tree.Transfer `msgpack:"desc,omitempty"`
}

type callArg struct {
	Workspace WSID   `msgpack:"workspace"`
	Process   uint32 `msgpack:"process"`
	Provider  uint32 `msgpack:"provider"`
	Caller    uint32 `msgpack:"caller"`
	Data      []byte `msgpack:"data"`
}

type broadcastArg struct {
	Workspace WSID      `msgpack:"workspace"`
	Entry     string    `msgpack:"entry"`
	Cap       uuid.UUID `msgpack:"cap"`
	Caller    uint32    `msgpack:"caller"`
	Data      []byte    `msgpack:"data"`
}

type lockArg struct {
	Workspace WSID   `msgpack:"workspace"`
	Lock      string `msgpack:"lock"`
	Action    bool   `msgpack:"action"`
}

func NewManager(network *node.Manager) *Manager {
	man := &Manager{
		Workspaces: map[WSID]*Workspace{},
		network:    network,
	}

	// workspace operator
	network.Callback["extendWorkspace"] = man.extendWorkspace
	network.Callback["reduceWorkspace"] = man.reduceWorkspace
	network.Callback["closeWorkspace"] = man.closeWorkspace
	// process operator
	network.Callback["createProcess"] = man.createProcess
	network.Callback["releaseProcess"] = man.releaseProcess
	network.Callback["syncProcess"] = man.syncProcess
	// service operator
	network.Callback["syncService"] = man.syncService
	network.Callback["callService"] = man.callService
	network.Callback["callBroadcast"] = man.callBroadcast
	network.Callback["lockAction"] = man.lockAction
	// network.Callback["registerService"] = man.registerService
	// network.Callback["unregisterService"] = man.unregisterService
	// network.Callback["listService"] = man.listService
	// network.Callback["callService"] = man.callService
	return man
}

func (man *Manager) createProcess(arg []byte, caller node.HostID) ([]byte, error) {
	parsedArg := processArg{}
	err := msgpack.Unmarshal(arg, &parsedArg)
	if err != nil {
		return nil, err
	}
	if ws, ok := man.Workspaces[parsedArg.Workspace]; ok {
		if part, ok := ws.parts[caller]; ok {
			part.createProcess(parsedArg.Process)
		}
	}
	return []byte{}, nil
}

func (man *Manager) releaseProcess(arg []byte, caller node.HostID) ([]byte, error) {
	parsedArg := processArg{}
	err := msgpack.Unmarshal(arg, &parsedArg)
	if err != nil {
		return nil, err
	}
	if ws, ok := man.Workspaces[parsedArg.Workspace]; ok {
		if part, ok := ws.parts[caller]; ok {
			part.releaseProcess(parsedArg.Process)
		}
	}
	return []byte{}, nil
}

func (man *Manager) extendWorkspace(arg []byte, caller node.HostID) ([]byte, error) {
	parsedArg := extendWorkspaceArg{}
	err := msgpack.Unmarshal(arg, &parsedArg)
	if err != nil {
		return nil, err
	}
	remotes := []node.Node{}
	for _, node := range parsedArg.Parts {
		remote, ok := man.network.Nodes[node]
		if !ok {
			return nil, ErrUnknownRemote
		}
		remotes = append(remotes, remote)
	}
	// get ws
	ws, ok := man.Workspaces[parsedArg.Workspace]
	if !ok {
		ws, err = man.newWorkspace(parsedArg.Workspace)
		if err != nil {
			return nil, err
		}
	}
	// sync parts
	for _, remote := range remotes {
		part, ok := ws.parts[remote.Host()]
		if !ok {
			part = newRemoteProcMan(ws, remote)
			ws.parts[remote.Host()] = part
		}
		// sync process
		arg, _ := msgpack.Marshal(parsedArg.Workspace)
		data, err := part.node.Call("syncProcess", arg)
		if err == nil {
			res := []uint32{}
			err := msgpack.Unmarshal(data, &res)
			if err == nil {
				pidMap := map[uint32]bool{}
				for _, pid := range res {
					pidMap[pid] = true
					if _, ok := part.AliveProcess[pid]; !ok {
						part.createProcess(pid)
					}
				}
				removes := []uint32{}
				for pid := range part.AliveProcess {
					if _, ok := pidMap[pid]; !ok {
						removes = append(removes, pid)
					}
				}
				for _, pid := range removes {
					part.releaseProcess(pid)
				}
			}
		}
		// sync service
		// TODO:
		// data, err = part.node.Call("listService", []byte{})
		// if err == nil {
		// 	res := []string{}
		// 	err := msgpack.Unmarshal(data, &res)
		// 	if err == nil {
		// 		entryMap := map[string]bool{}
		// 		for _, entry := range res {
		// 			entryMap[entry] = true
		// 			if _, ok := part.service.services[entry]; !ok {
		// 				part.service.services[entry] = 0
		// 			}
		// 		}
		// 		removes := []string{}
		// 		for entry := range part.service.services {
		// 			if _, ok := entryMap[entry]; !ok {
		// 				removes = append(removes, entry)
		// 			}
		// 		}
		// 		for _, entry := range removes {
		// 			delete(part.service.services, entry)
		// 		}
		// 	}
		// }
	}
	return []byte{}, nil
}

func (man *Manager) reduceWorkspace(arg []byte, caller node.HostID) ([]byte, error) {
	parsedArg := extendWorkspaceArg{}
	err := msgpack.Unmarshal(arg, &parsedArg)
	if err != nil || len(parsedArg.Parts) != 1 {
		return nil, err
	}
	ws, ok := man.Workspaces[parsedArg.Workspace]
	if !ok {
		return []byte{}, nil
	}
	if parsedArg.Parts[0] == man.network.HostId {
		ws.closeImpl()
		delete(man.Workspaces, parsedArg.Workspace)
		return []byte{}, nil
	}
	part, ok := ws.parts[parsedArg.Parts[0]]
	if !ok {
		return []byte{}, nil
	}
	part.close()
	return []byte{}, nil
}

func (man *Manager) closeWorkspace(arg []byte, caller node.HostID) ([]byte, error) {
	parsedArg := WSID{}
	err := msgpack.Unmarshal(arg, &parsedArg)
	if err != nil {
		return nil, err
	}
	ws, ok := man.Workspaces[parsedArg]
	if !ok {
		return []byte{}, nil
	}
	ws.closeImpl()
	delete(man.Workspaces, parsedArg)
	return []byte{}, nil
}

func (man *Manager) syncProcess(arg []byte, caller node.HostID) ([]byte, error) {
	parsedArg := WSID{}
	err := msgpack.Unmarshal(arg, &parsedArg)
	if err != nil {
		return nil, err
	}
	ws, ok := man.Workspaces[parsedArg]
	if !ok {
		return nil, ErrUnknownWorkspace
	}
	list := []uint32{}
	for it := range ws.processes.AliveProcess {
		list = append(list, it)
	}
	return msgpack.Marshal(list)
}

// func (man *Manager) registerService(arg []byte, caller node.HostID) ([]byte, error) {
// 	parsedArg := serviceArg{}
// 	err := msgpack.Unmarshal(arg, &parsedArg)
// 	if err != nil {
// 		return nil, err
// 	}
// 	ws, ok := man.Workspaces[parsedArg.Workspace]
// 	if !ok {
// 		return nil, ErrUnknownWorkspace
// 	}
// 	part, ok := ws.parts[caller]
// 	if !ok {
// 		return nil, ErrUnknownRemote
// 	}
// 	part.service.services[parsedArg.Entry] = parsedArg.Process
// 	return nil, err
// }

// func (man *Manager) unregisterService(arg []byte, caller node.HostID) ([]byte, error) {
// 	parsedArg := serviceArg{}
// 	err := msgpack.Unmarshal(arg, &parsedArg)
// 	if err != nil {
// 		return nil, err
// 	}
// 	ws, ok := man.Workspaces[parsedArg.Workspace]
// 	if !ok {
// 		return nil, ErrUnknownWorkspace
// 	}
// 	part, ok := ws.parts[caller]
// 	if !ok {
// 		return nil, ErrUnknownRemote
// 	}
// 	delete(part.service.services, parsedArg.Entry)
// 	return nil, err
// }

func (man *Manager) syncService(arg []byte, caller node.HostID) ([]byte, error) {
	parsedArg := serviceArg{}
	err := msgpack.Unmarshal(arg, &parsedArg)
	if err != nil {
		return nil, err
	}
	ws, ok := man.Workspaces[parsedArg.Workspace]
	if !ok {
		return nil, ErrUnknownWorkspace
	}
	ret, err := ws.service.Sync(parsedArg.Desc)
	if err != nil {
		return nil, err
	}
	return msgpack.Marshal(ret)
}

func (man *Manager) callService(arg []byte, caller node.HostID) ([]byte, error) {
	parsedArg := callArg{}
	err := msgpack.Unmarshal(arg, &parsedArg)
	if err != nil {
		return nil, err
	}
	ws, ok := man.Workspaces[parsedArg.Workspace]
	if !ok {
		return nil, ErrUnknownWorkspace
	}
	part, ok := ws.parts[caller]
	if !ok {
		return nil, ErrUnknownRemote
	}
	callerProc, ok := part.AliveProcess[parsedArg.Caller]
	if !ok {
		return nil, ErrUnknownProcess
	}
	providerProc, ok := ws.processes.AliveProcess[parsedArg.Process]
	if !ok {
		return nil, ErrUnknownProcess
	}
	return providerProc.CallProvider(parsedArg.Provider, parsedArg.Data, callerProc)
}

func (man *Manager) callBroadcast(arg []byte, caller node.HostID) ([]byte, error) {
	parsedArg := broadcastArg{}
	err := msgpack.Unmarshal(arg, &parsedArg)
	if err != nil {
		return nil, err
	}
	ws, ok := man.Workspaces[parsedArg.Workspace]
	if !ok {
		return nil, ErrUnknownWorkspace
	}
	part, ok := ws.parts[caller]
	if !ok {
		return nil, ErrUnknownRemote
	}
	callerProc, ok := part.AliveProcess[parsedArg.Caller]
	if !ok {
		return nil, ErrUnknownProcess
	}
	return ws.service.CallService(parsedArg.Entry, parsedArg.Cap, parsedArg.Data, callerProc)
}

func (man *Manager) lockAction(arg []byte, caller node.HostID) ([]byte, error) {
	parsedArg := lockArg{}
	err := msgpack.Unmarshal(arg, &parsedArg)
	if err != nil {
		return nil, err
	}
	ws, ok := man.Workspaces[parsedArg.Workspace]
	if !ok {
		return nil, ErrUnknownWorkspace
	}
	ret := ws.service.LockAction(parsedArg.Lock, caller, parsedArg.Action)
	return msgpack.Marshal(ret)
}
