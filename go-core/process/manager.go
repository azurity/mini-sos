package process

import (
	"errors"
	"sync"

	"github.com/azurity/mini-sos/go-core/node"
	"github.com/google/uuid"
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

type Process interface {
	Host() node.HostID
	Id() uint32
	WaitQuit()
	Kill()
	Run() error
	CallProvider(id uint32, data []byte, caller Process) ([]byte, error)
}

type Manager struct {
	host            node.HostID
	pidAllocateLock sync.RWMutex
	currentId       uint32
	AliveProcess    map[uint32]Process
	callFn          func(entry string, cap uuid.UUID, data []byte, caller Process) ([]byte, error)
	CreateCallback  func(pid uint32)
}

func NewManager(host node.HostID, callFn func(entry string, cap uuid.UUID, data []byte, caller Process) ([]byte, error)) *Manager {
	return &Manager{
		host:            host,
		pidAllocateLock: sync.RWMutex{},
		currentId:       0,
		AliveProcess:    map[uint32]Process{},
		callFn:          callFn,
		CreateCallback:  nil,
	}
}

func (man *Manager) allocate() (uint32, error) {
	man.pidAllocateLock.Lock()
	defer man.pidAllocateLock.Unlock()
	if len(man.AliveProcess) == 0x100000000 {
		man.pidAllocateLock.RUnlock()
		return 0, ErrFullOfProcess
	}
	for {
		if _, ok := man.AliveProcess[man.currentId]; !ok {
			man.AliveProcess[man.currentId] = nil
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
