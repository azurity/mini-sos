package service

import (
	"context"
	"errors"
	"io/fs"
	"sync"

	"github.com/azurity/mini-sos/go-core/node"
	"github.com/azurity/mini-sos/go-core/process"
	"github.com/azurity/mini-sos/go-core/service/capability"
	"github.com/azurity/mini-sos/go-core/service/structs"
	"github.com/azurity/mini-sos/go-core/utils/tree"
	"github.com/google/uuid"
	"github.com/vmihailenco/msgpack/v5"
)

var ErrExist = errors.New("service/cap exist")
var ErrNotExist = errors.New("service/cap not exist")

// var ErrIllegal = errors.New("service update illegal")

// type Provider interface {
// 	Host() node.HostID
// 	Id() uint32
// 	WaitQuit()
// 	Kill()
// 	CallProvider(id uint32, data []byte, caller Provider) ([]byte, error)
// }

// type Manager interface {
// 	Register(entry string, provider Provider, id uint32, sub bool, local bool) error
// 	Update(entry string, provider Provider, id uint32, local bool) error
// 	Unregister(entry string, provider Provider, local bool)
// 	List(local bool) []string
// 	Call(entry string, data []byte, caller Provider) ([]byte, error)
// }

var ErrWrongCaller = errors.New("wrong caller")
var ErrUnknownProcess = errors.New("unknown process")
var ErrNotService = errors.New("not a service")

type CallerInfo struct {
	Host    node.HostID
	Process uint32
}

type callRemote func(host node.HostID, process uint32, provider uint32, data []byte, caller CallerInfo) ([]byte, error)

type callBroadcast func(entry string, cap uuid.UUID, data []byte, caller CallerInfo)

type remoteCaller func(caller CallerInfo) (process.Process, error)

type lockFn func(path string) bool

type Manager struct {
	tree          *tree.MountTree
	lockTree      *lockTree[uuid.UUID]
	rootService   *serviceTree
	host          node.HostID
	procMan       *process.Manager
	callRemote    callRemote
	callBroadcast callBroadcast
	remoteCaller  remoteCaller
	lockFn        lockFn
	unlockFn      lockFn
	initChan      chan bool
}

func NewManager(host node.HostID) *Manager {
	ret := &Manager{
		tree: tree.NewMountTree(tree.NewBasicTree(tree.NewPath("/"))),
		lockTree: &lockTree[uuid.UUID]{
			locks: map[string]uuid.UUID{},
			mtx:   sync.Mutex{},
		},
		host:     host,
		initChan: make(chan bool),
	}
	return ret
}

func (man *Manager) Init(procMan *process.Manager, callRemote callRemote, callBroadcast callBroadcast, remoteCaller remoteCaller, lockFn lockFn, unlockFn lockFn) error {
	man.procMan = procMan
	man.callRemote = callRemote
	man.callBroadcast = callBroadcast
	man.remoteCaller = remoteCaller
	man.lockFn = lockFn
	man.unlockFn = unlockFn
	proc, err := procMan.NewNativeProcess(man.process)
	if err != nil {
		return err
	}
	go proc.Run()
	<-man.initChan
	return nil
}

func newCaller(proc process.Process) context.Context {
	return context.WithValue(context.Background(), "caller", CallerInfo{
		Host:    proc.Host(),
		Process: proc.Id(),
	})
}

func (man *Manager) getCaller(info CallerInfo) (process.Process, error) {
	if info.Host == uuid.Nil || info.Host == man.host {
		proc, ok := man.procMan.AliveProcess[info.Process]
		if ok {
			return proc, nil
		}
		return nil, ErrWrongCaller
	} else {
		return man.remoteCaller(info)
	}
}

func (man *Manager) callFn(host node.HostID, process uint32, provider uint32, data []byte, ctx context.Context) ([]byte, error) {
	info, ok := ctx.Value("caller").(CallerInfo)
	if !ok {
		return nil, ErrWrongCaller
	}
	caller, err := man.getCaller(info)
	if err != nil {
		return nil, err
	}
	if host == uuid.Nil || host == man.host {
		proc, ok := man.procMan.AliveProcess[process]
		if !ok {
			return []byte{}, ErrUnknownProcess
		}
		return proc.CallProvider(provider, data, caller)
	} else {
		return man.callRemote(host, process, provider, data, info)
	}
}

func (man *Manager) CallService(entry string, cap uuid.UUID, data []byte, caller process.Process) ([]byte, error) {
	path := tree.NewPath(entry).Normalize()
	relpath, err := tree.Rel(path, tree.NewPath("/"))
	if err != nil {
		return nil, err
	}
	node, err := man.tree.Find(relpath, newCaller(caller))
	if err != nil {
		return nil, err
	}
	cased, ok := node.(*serviceTree)
	if !ok {
		// TODO: checker auth here?
		return man.rootService.CallCap(cap, data, newCaller(caller))
	}
	// TODO: checker auth here?
	return cased.CallCap(cap, data, newCaller(caller))
}

func (man *Manager) Sync(info *tree.Transfer) (*tree.Transfer, error) {
	if info == nil {
		return man.tree.Transfer()
	}
	node, err := tree.ExtractTransfer(*info)
	if err != nil {
		return nil, err
	}
	cased, ok := node.(*tree.MountTree)
	if !ok {
		return nil, fs.ErrInvalid
	}
	for _, item := range cased.Leaves() {
		item.(*serviceTree).callFn = man.callFn
	}
	err = man.tree.Merge(cased, context.WithValue(context.Background(), "caller", CallerInfo{
		Host:    man.host,
		Process: 0,
	}))
	if err != nil {
		return nil, err
	}
	return nil, nil
}

func (man *Manager) LockAction(path string, host uuid.UUID, action bool) bool {
	if action {
		return man.lockTree.Lock(path, host)
	} else {
		return man.lockTree.Unlock(path, host)
	}
}

func addTypedCap[Arg any, Ret any](man *Manager, op *process.NativeOperator, entrypoint *serviceTree, cap uuid.UUID, fn func(arg *Arg, caller process.Process) (*Ret, error)) error {
	id, err := op.ProviderManager().New(process.TypedProvider(fn))
	if err != nil {
		return err
	}
	if err := entrypoint.AddCap(cap, man.host, op.Pid(), id); err != nil {
		return err
	}
	return nil
}

func (man *Manager) process(ctx context.Context, op *process.NativeOperator) func() error {
	return func() error {
		service := tree.NewBasicTree(tree.NewPath("/service"))
		if _, err := man.tree.Create(tree.NewPath("service"), service, false, newCaller(op.Proc())); err != nil {
			return err
		}
		cap := tree.NewBasicTree(tree.NewPath("/service/cap"))
		if _, err := service.Create(tree.NewPath("cap"), cap, false, newCaller(op.Proc())); err != nil {
			return err
		}
		// add /service/register
		register := NewServiceTree(tree.NewPath("/service/register"), man.callFn, true)
		if _, err := man.tree.Mount(register, tree.NewPath("service/register"), newCaller(op.Proc())); err != nil {
			return err
		}
		if err := addTypedCap(man, op, register, capability.CapCall, man.register); err != nil {
			return err
		}
		// add /service/unregister
		unregister := NewServiceTree(tree.NewPath("/service/unregister"), man.callFn, true)
		if _, err := man.tree.Mount(unregister, tree.NewPath("service/unregister"), newCaller(op.Proc())); err != nil {
			return err
		}
		if err := addTypedCap(man, op, unregister, capability.CapCall, man.unregister); err != nil {
			return err
		}
		// add /service/cap/add
		addCap := NewServiceTree(tree.NewPath("/service/cap/add"), man.callFn, true)
		if _, err := man.tree.Mount(addCap, tree.NewPath("service/cap/add"), newCaller(op.Proc())); err != nil {
			return err
		}
		if err := addTypedCap(man, op, addCap, capability.CapCall, man.addCap); err != nil {
			return err
		}
		// add /service/cap/del
		delCap := NewServiceTree(tree.NewPath("/service/cap/del"), man.callFn, true)
		if _, err := man.tree.Mount(delCap, tree.NewPath("service/cap/del"), newCaller(op.Proc())); err != nil {
			return err
		}
		if err := addTypedCap(man, op, delCap, capability.CapCall, man.delCap); err != nil {
			return err
		}
		// add /service/cap/list
		listCap := NewServiceTree(tree.NewPath("/service/cap/list"), man.callFn, true)
		if _, err := man.tree.Mount(listCap, tree.NewPath("service/cap/list"), newCaller(op.Proc())); err != nil {
			return err
		}
		if err := addTypedCap(man, op, listCap, capability.CapCall, man.listCap); err != nil {
			return err
		}
		// add root fs caps
		rootService := NewServiceTree(tree.NewPath("/"), man.callFn, true)
		if err := addTypedCap(man, op, rootService, capability.CapListDir, man.listDir); err != nil {
			return err
		}
		if err := addTypedCap(man, op, rootService, capability.CapSetDir, man.setDir); err != nil {
			return err
		}
		man.rootService = rootService
		man.initChan <- true
		<-ctx.Done()
		return nil
	}
}

func (man *Manager) register(arg *structs.RegisterReq, caller process.Process) (*structs.RegisterRes, error) {
	path := tree.NewPath(arg.Path).Normalize()
	if path.IsRel() || len(path) == 0 {
		return structs.NewErr(fs.ErrInvalid), nil
	}
	lockStr := path.Str()
	if !man.lockTree.Lock(lockStr, man.host) {
		return structs.NewErr(ErrOccupied), nil
	}
	defer man.lockTree.Unlock(lockStr, man.host)
	if !man.lockFn(lockStr) {
		return structs.NewErr(ErrOccupied), nil
	}
	defer man.unlockFn(lockStr)
	node := NewServiceTree(path, man.callFn, arg.Local)
	relPath, _ := tree.Rel(path, tree.Path{})
	// TODO: checker auth here, except pid:0
	_, err := man.tree.Create(relPath, node, false, newCaller(caller))
	if err != nil {
		return structs.NewErr(err), nil
	}
	if !arg.Local && caller.Id() != 0 {
		data, _ := msgpack.Marshal(arg)
		man.callBroadcast("/service/register", capability.CapCall, data, CallerInfo{
			Host:    man.host,
			Process: 0,
		})
	}
	return structs.NewSuccess(), nil
}

func (man *Manager) unregister(arg *structs.UnregisterReq, caller process.Process) (*structs.UnregisterRes, error) {
	path := tree.NewPath(*arg).Normalize()
	if path.IsRel() || len(path) == 0 {
		return structs.NewErr(fs.ErrInvalid), nil
	}
	lockStr := path.Str()
	if !man.lockTree.Lock(lockStr, man.host) {
		return structs.NewErr(ErrOccupied), nil
	}
	defer man.lockTree.Unlock(lockStr, man.host)
	if !man.lockFn(lockStr) {
		return structs.NewErr(ErrOccupied), nil
	}
	defer man.unlockFn(lockStr)
	relPath, _ := tree.Rel(path, tree.Path{})
	// TODO: checker auth here, except pid:0
	err := man.tree.Remove(relPath, false, newCaller(caller))
	if err != nil {
		return structs.NewErr(err), nil
	}
	if caller.Id() != 0 {
		data, _ := msgpack.Marshal(arg)
		man.callBroadcast("/service/unregister", capability.CapCall, data, CallerInfo{
			Host:    man.host,
			Process: 0,
		})
	}
	return structs.NewSuccess(), nil
}

func (man *Manager) addCap(arg *structs.AddCapReq, caller process.Process) (*structs.AddCapRes, error) {
	path := tree.NewPath(arg.Path).Normalize()
	if path.IsRel() || len(path) == 0 {
		return structs.NewErr(fs.ErrInvalid), nil
	}
	lockStr := path.Str()
	if !man.lockTree.Lock(lockStr, man.host) {
		return structs.NewErr(ErrOccupied), nil
	}
	defer man.lockTree.Unlock(lockStr, man.host)
	if !man.lockFn(lockStr) {
		return structs.NewErr(ErrOccupied), nil
	}
	defer man.unlockFn(lockStr)
	relPath, _ := tree.Rel(path, tree.Path{})
	node, err := man.tree.Find(relPath, newCaller(caller))
	if err != nil {
		return structs.NewErr(err), nil
	}
	cased, ok := node.(*serviceTree)
	if !ok {
		return structs.NewErr(ErrNotService), nil
	}
	validHost := arg.Desc.Host
	if validHost == uuid.Nil {
		validHost = man.host
	}
	validPid := arg.Desc.Process
	if caller.Id() != 0 {
		validPid = caller.Id()
	}
	err = cased.AddCap(arg.Desc.Name, validHost, validPid, arg.Desc.Provider)
	if err != nil {
		return structs.NewErr(err), nil
	}
	if caller.Id() != 0 {
		data, _ := msgpack.Marshal(arg)
		man.callBroadcast("/service/cap/add", capability.CapCall, data, CallerInfo{
			Host:    man.host,
			Process: 0,
		})
	}
	return structs.NewSuccess(), nil
}

func (man *Manager) delCap(arg *structs.DelCapReq, caller process.Process) (*structs.DelCapRes, error) {
	path := tree.NewPath(arg.Path).Normalize()
	if path.IsRel() || len(path) == 0 {
		return structs.NewErr(fs.ErrInvalid), nil
	}
	lockStr := path.Str()
	if !man.lockTree.Lock(lockStr, man.host) {
		return structs.NewErr(ErrOccupied), nil
	}
	defer man.lockTree.Unlock(lockStr, man.host)
	if !man.lockFn(lockStr) {
		return structs.NewErr(ErrOccupied), nil
	}
	defer man.unlockFn(lockStr)
	relPath, _ := tree.Rel(path, tree.Path{})
	node, err := man.tree.Find(relPath, newCaller(caller))
	if err != nil {
		return structs.NewErr(err), nil
	}
	cased, ok := node.(*serviceTree)
	if !ok {
		return structs.NewErr(ErrNotService), nil
	}
	cased.DelCap(arg.Cap)
	if caller.Id() != 0 {
		data, _ := msgpack.Marshal(arg)
		man.callBroadcast("/service/cap/del", capability.CapCall, data, CallerInfo{
			Host:    man.host,
			Process: 0,
		})
	}
	return structs.NewSuccess(), nil
}

func (man *Manager) listCap(arg *structs.ListCapReq, caller process.Process) (*structs.ListCapRes, error) {
	path := tree.NewPath(arg.Path).Normalize()
	if path.IsRel() || len(path) == 0 {
		return &structs.ListCapRes{Error: structs.NewErr(fs.ErrInvalid)}, nil
	}
	relPath, _ := tree.Rel(path, tree.Path{})
	node, err := man.tree.Find(relPath, newCaller(caller))
	if err != nil {
		return &structs.ListCapRes{Error: structs.NewErr(err)}, nil
	}
	cased, ok := node.(*serviceTree)
	if !ok {
		return &structs.ListCapRes{Error: structs.NewErr(ErrNotService)}, nil
	}
	return &structs.ListCapRes{
		Error: structs.NewSuccess(),
		Caps:  cased.toEntry(),
	}, nil
}

func (man *Manager) listDir(arg *structs.ListDirReq, caller process.Process) (*structs.ListDirRes, error) {
	base := tree.Path(arg.Abs).Normalize()
	path := tree.Path(arg.Path).Normalize()
	if base.IsRel() || !path.IsRel() {
		return &structs.ListDirRes{Error: structs.NewErr(fs.ErrInvalid)}, nil
	}
	relPath, _ := tree.Rel(base, tree.Path{})
	relPath = tree.Join(relPath, path)
	node, err := man.tree.Find(relPath, newCaller(caller))
	if err != nil {
		return &structs.ListDirRes{Error: structs.NewErr(err)}, nil
	}
	entries, err := node.List(newCaller(caller))
	if err != nil {
		return &structs.ListDirRes{Error: structs.NewErr(err)}, nil
	}
	ret := &structs.ListDirRes{
		Error:   structs.NewSuccess(),
		Entires: map[string]structs.Entry{},
	}
	for name, entry := range entries {
		if cased, ok := entry.(*serviceTree); ok {
			ret.Entires[name] = cased.toEntry()
		} else {
			ret.Entires[name] = man.rootService.toEntry()
		}
	}
	return ret, nil
}

func (man *Manager) setDir(arg *structs.SetDirReq, caller process.Process) (*structs.SetDirRes, error) {
	broadcast := func() {
		if caller.Id() != 0 {
			data, _ := msgpack.Marshal(arg)
			man.callBroadcast("/", capability.CapSetDir, data, CallerInfo{
				Host:    man.host,
				Process: 0,
			})
		}
	}
	base := tree.Path(arg.Abs).Normalize()
	path := tree.Path(arg.Path).Normalize()
	if base.IsRel() || !path.IsRel() {
		return &structs.SetDirRes{Error: structs.NewErr(fs.ErrInvalid)}, nil
	}
	lockStr := tree.Join(base, path).Str()
	if !man.lockTree.Lock(lockStr, man.host) {
		return &structs.SetDirRes{Error: structs.NewErr(ErrOccupied)}, nil
	}
	defer man.lockTree.Unlock(lockStr, man.host)
	if !man.lockFn(lockStr) {
		return &structs.SetDirRes{Error: structs.NewErr(ErrOccupied)}, nil
	}
	defer man.unlockFn(lockStr)
	relPath, _ := tree.Rel(base, tree.Path{})
	node, err := man.tree.Find(relPath, newCaller(caller))
	if err != nil {
		return &structs.SetDirRes{Error: structs.NewErr(err)}, nil
	}
	if arg.Item != nil {
		subTree, err := tree.ExtractTransfer(*arg.Item)
		if err != nil {
			return &structs.SetDirRes{Error: structs.NewErr(err)}, nil
		}
		retTree, err := node.Create(path, subTree, arg.Option, newCaller(caller))
		if err != nil {
			return &structs.SetDirRes{Error: structs.NewErr(err)}, nil
		}
		if retTree == nil {
			broadcast()
			return &structs.SetDirRes{Error: structs.NewSuccess()}, nil
		}
		transfer, err := retTree.Transfer()
		if err != nil {
			return &structs.SetDirRes{Error: structs.NewErr(err)}, nil
		}
		broadcast()
		return &structs.SetDirRes{
			Error: structs.NewSuccess(),
			Item:  transfer,
		}, nil
	} else {
		err := node.Remove(path, arg.Option, newCaller(caller))
		if err != nil {
			return &structs.SetDirRes{Error: structs.NewErr(err)}, nil
		}
		broadcast()
		return &structs.SetDirRes{Error: structs.NewSuccess()}, nil
	}
}
