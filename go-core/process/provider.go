package process

import (
	"errors"
	"sync"
)

var ErrFullOfProvider = errors.New("full of provider")

type ProviderManager[T any] struct {
	currentId    uint32
	allocateLock sync.RWMutex
	content      map[uint32]*T
}

func (man *ProviderManager[T]) allocate() (uint32, error) {
	man.allocateLock.Lock()
	defer man.allocateLock.Unlock()
	if len(man.content) == 0xffffffff {
		man.allocateLock.RUnlock()
		return 0, ErrFullOfProvider
	}
	for {
		if _, ok := man.content[man.currentId]; !ok {
			man.content[man.currentId] = nil
			return man.currentId, nil
		}
		man.currentId += 1
		if man.currentId == 0 {
			man.currentId += 1
		}
	}
}

func (man *ProviderManager[T]) release(id uint32) {
	man.allocateLock.Lock()
	defer man.allocateLock.Unlock()
	delete(man.content, id)
}

func NewProviderManager[T any]() *ProviderManager[T] {
	return &ProviderManager[T]{
		currentId:    1,
		allocateLock: sync.RWMutex{},
		content:      map[uint32]*T{},
	}
}

func (man *ProviderManager[T]) New(item T) (uint32, error) {
	id, err := man.allocate()
	if err != nil {
		return 0, err
	}
	man.Set(id, &item)
	return id, nil
}

func (man *ProviderManager[T]) Set(id uint32, item *T) {
	man.allocateLock.Lock()
	defer man.allocateLock.Unlock()
	man.content[id] = item
}

func (man *ProviderManager[T]) Get(id uint32) *T {
	man.allocateLock.RLock()
	defer man.allocateLock.RUnlock()
	if item, ok := man.content[id]; ok {
		return item
	}
	return nil
}

func (man *ProviderManager[T]) Del(id uint32) {
	man.release(id)
}
