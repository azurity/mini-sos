package service

import (
	"errors"
	"sync"

	"github.com/azurity/mini-sos/go-core/utils/tree"
)

var ErrOccupied = errors.New("occupied")

type lockTree[Info comparable] struct {
	locks map[string]Info
	mtx   sync.Mutex
}

func (ltree *lockTree[Info]) Lock(path string, info Info) bool {
	normlaized := tree.NewPath(path).Normalize()
	if normlaized.IsRel() {
		return false
	}
	ltree.mtx.Lock()
	defer ltree.mtx.Unlock()
	sub := append(tree.Path{}, normlaized...)
	for {
		if _, ok := ltree.locks[sub.Str()]; ok {
			return false
		}
		if len(sub) == 0 {
			break
		}
		sub = sub[:len(sub)-1]
	}
	ltree.locks[normlaized.Str()] = info
	return true
}

func (ltree *lockTree[Info]) Unlock(path string, info Info) bool {
	normlaized := tree.NewPath(path).Normalize()
	if normlaized.IsRel() {
		return false
	}
	ltree.mtx.Lock()
	defer ltree.mtx.Unlock()
	if value, ok := ltree.locks[normlaized.Str()]; ok {
		if value != info {
			return false
		}
		delete(ltree.locks, normlaized.Str())
	}
	return true
}
