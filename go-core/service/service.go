package service

import (
	"errors"

	"github.com/azurity/mini-sos/go-core/node"
)

var ErrExist = errors.New("service exist")
var ErrNotExist = errors.New("service not exist")
var ErrIllegal = errors.New("service update illegal")

type Provider interface {
	Host() node.HostID
	Id() uint32
	WaitQuit()
	Kill()
	CallProvider(id uint32, data []byte, caller Provider) ([]byte, error)
}

type Manager interface {
	Register(entry string, provider Provider, id uint32, local bool) error
	Update(entry string, provider Provider, id uint32, local bool) error
	Unregister(entry string, provider Provider, local bool)
	List(local bool) []string
	Call(entry string, data []byte, caller Provider) ([]byte, error)
}
