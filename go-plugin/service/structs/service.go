package structs

import "github.com/azurity/mini-sos/go-plugin/api"

//go:generate msgp

type RegisterReq struct {
	Path  string `msg:"path"`
	Local bool   `msg:"local"`
}

type RegisterRes = Error

type UnregisterReq string

type UnregisterRes = Error

type ListReq struct{}

type ListRes []string

type AddCapReq struct {
	Path string `msg:"path"`
	Desc Cap    `msg:"desc"`
}

type AddCapRes = Error

type DelCapReq struct {
	Path string   `msg:"path"`
	Cap  api.UUID `msg:"cap"`
}

type DelCapRes = Error

type ListCapReq struct {
	Path string `msg:"path"`
}

type ListCapRes struct {
	Error *Error `msg:"error"`
	Caps  []Cap  `msg:"caps"`
}
