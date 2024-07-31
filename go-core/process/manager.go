package process

import (
	"errors"
	"log"
	"sync"

	"github.com/azurity/mini-sos/go-core/node"
	"github.com/azurity/mini-sos/go-core/service"
	"github.com/vmihailenco/msgpack/v5"
)

var ErrFullOfProcess = errors.New("full of process")
var ErrNotFound = errors.New("service not found")

const (
	RunStateReady uint32 = iota
	RunStateRunning
	RunStateExit
)

const (
	SignalExit uint32 = iota
	SignalKill
)

type Manager struct {
	host            node.HostID
	pidAllocateLock sync.RWMutex
	currentId       uint32
	AliveProcess    map[uint32]service.Provider
	callFn          func(entry string, data []byte, proc service.Provider) ([]byte, error)
	CreateCallback  func(pid uint32)
}

func NewManager(host node.HostID, callFn func(entry string, data []byte, proc service.Provider) ([]byte, error)) *Manager {
	return &Manager{
		host:            host,
		pidAllocateLock: sync.RWMutex{},
		currentId:       0,
		AliveProcess:    map[uint32]service.Provider{},
		callFn:          callFn,
		CreateCallback:  nil,
	}
}

func (man *Manager) allocate() (uint32, error) {
	man.pidAllocateLock.RLock()
	if len(man.AliveProcess) == 0x100000000 {
		man.pidAllocateLock.RUnlock()
		return 0, ErrFullOfProcess
	}
	man.pidAllocateLock.RUnlock()
	man.pidAllocateLock.Lock()
	defer man.pidAllocateLock.Unlock()
	for {
		if _, ok := man.AliveProcess[man.currentId]; !ok {
			return man.currentId, nil
		}
		man.currentId += 1
	}
}

func (man *Manager) release(pid uint32) {
	man.pidAllocateLock.Lock()
	defer man.pidAllocateLock.Unlock()
	delete(man.AliveProcess, pid)
}

type RegisterArgs struct {
	Service  string `msgpack:"service"`
	Function string `msgpack:"function"`
}

type UnregisterArgs struct {
	Service string `msgpack:"service"`
}

func (man *Manager) InitServiceProcess(serviceMan service.Manager) (*DummyProcess, error) {
	serviceProc, err := man.NewDummyProcess()
	if err != nil {
		return nil, err
	}

	serviceProc.Providers["/service/register"] = func(data []byte, caller service.Provider) ([]byte, error) {
		args := RegisterArgs{}
		err := msgpack.Unmarshal(data, &args)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		err = serviceMan.Register(args.Service, caller, false)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		ret, err := msgpack.Marshal(true)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		if proc, ok := caller.(*LocalProcess); ok {
			proc.channelMap[args.Service] = args.Function
		}
		return ret, nil
	}
	serviceMan.Register("/service/register", serviceProc, true)

	serviceProc.Providers["/service/unregister"] = func(data []byte, caller service.Provider) ([]byte, error) {
		args := UnregisterArgs{}
		err := msgpack.Unmarshal(data, &args)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		serviceMan.Unregister(args.Service, caller, false)
		ret, err := msgpack.Marshal(true)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		return ret, nil
	}
	serviceMan.Register("/service/unregister", serviceProc, true)

	serviceProc.Providers["/service/list"] = func(data []byte, caller service.Provider) ([]byte, error) {
		list := serviceMan.List(false)
		ret, err := msgpack.Marshal(list)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		return ret, nil
	}
	serviceMan.Register("/service/list", serviceProc, true)

	return serviceProc, nil
}
