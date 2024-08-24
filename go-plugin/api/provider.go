package api

import (
	"errors"
	"sync"

	"github.com/extism/go-pdk"
	"github.com/google/uuid"
)

var ErrFullOfProvider = errors.New("full of provider")
var ErrNotExist = errors.New("provider not exist")

type Manager[T any] struct {
	currentId    uint32
	allocateLock sync.RWMutex
	content      map[uint32]*T
}

func (man *Manager[T]) allocate() (uint32, error) {
	man.allocateLock.Lock()
	defer man.allocateLock.Unlock()
	if len(man.content) == 0x7fffffff {
		man.allocateLock.RUnlock()
		return 0, ErrFullOfProvider
	}
	for {
		if _, ok := man.content[man.currentId]; !ok {
			man.content[man.currentId] = nil
			return man.currentId, nil
		}
		man.currentId += 1
		man.currentId &= 0x7fffffff
		if man.currentId == 0 {
			man.currentId += 1
		}
	}
}

func (man *Manager[T]) release(id uint32) {
	man.allocateLock.Lock()
	defer man.allocateLock.Unlock()
	delete(man.content, id)
}

func newManager[T any]() *Manager[T] {
	return &Manager[T]{
		currentId:    1,
		allocateLock: sync.RWMutex{},
		content:      map[uint32]*T{},
	}
}

func (man *Manager[T]) New(item T) (uint32, error) {
	id, err := man.allocate()
	if err != nil {
		return 0, err
	}
	man.Set(id, &item)
	return id, nil
}

func (man *Manager[T]) Set(id uint32, item *T) {
	man.allocateLock.Lock()
	defer man.allocateLock.Unlock()
	man.content[id] = item
}

func (man *Manager[T]) Get(id uint32) *T {
	man.allocateLock.RLock()
	defer man.allocateLock.RUnlock()
	if item, ok := man.content[id]; ok {
		return item
	}
	return nil
}

func (man *Manager[T]) Del(id uint32) {
	man.release(id)
}

type Caller struct {
	Host uuid.UUID
	Id   uint32
}

var ProviderManager = newManager[func(data []byte, caller Caller) ([]byte, error)]()

//go:export _sos_call
func _sos_call() int32 {
	raw := pdk.Input()
	arg := ProviderArg{}
	_, err := arg.UnmarshalMsg(raw)
	if err != nil {
		pdk.SetError(err)
		return 1
	}
	fn := ProviderManager.Get(arg.Id)
	if fn == nil {
		pdk.SetError(ErrNotExist)
		return 1
	}
	data, err := (*fn)(arg.Data, Caller{Host: uuid.Must(uuid.FromBytes(arg.CallerHost[:])), Id: arg.Id})
	if err != nil {
		pdk.SetError(err)
		return 1
	}
	pdk.Output(data)
	return 0
}
