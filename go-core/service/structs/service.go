package structs

import (
	"github.com/google/uuid"
)

type RegisterReq struct {
	Path  string `msgpack:"path"`
	Local bool   `msgpack:"local"`
}

type RegisterRes = Error

type UnregisterReq = string

type UnregisterRes = Error

type ListReq struct{}

type ListRes []string

type AddCapReq struct {
	Path string `msgpack:"path"`
	Desc Cap    `msgpack:"desc"`
}

type AddCapRes = Error

type DelCapReq struct {
	Path string    `msgpack:"path"`
	Cap  uuid.UUID `msgpack:"cap"`
}

type DelCapRes = Error

type ListCapReq struct {
	Path string `msgpack:"path"`
}

type ListCapRes struct {
	Error *Error `msgpack:"error"`
	Caps  []Cap  `msgpack:"caps"`
}
