package service

import (
	"errors"

	"github.com/azurity/mini-sos/go-core/node"
)

var ErrExist = errors.New("service exist")
var ErrNotExist = errors.New("service not exist")

type Provider interface {
	Host() node.HostID
	Id() uint32
	WaitQuit()
	Kill()
	CallService(service string, data []byte, caller Provider) ([]byte, error)
}

type Manager interface {
	Register(entry string, proc Provider, local bool) error
	Unregister(entry string, proc Provider, local bool)
	List(local bool) []string
	Call(entry string, data []byte, caller Provider) ([]byte, error)
}
