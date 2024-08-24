package structs

import "github.com/azurity/mini-sos/go-plugin/api"

//go:generate msgp

type Transfer struct {
	SubType string `msg:"sub_type"`
	Desc    []byte `msg:"desc"`
}

type ListDirReq struct {
	Abs  []string `msg:"base"`
	Path []string `msg:"path"`
}

type Cap struct {
	Name     api.UUID `msg:"name"`
	Host     api.UUID `msg:"host"`
	Process  uint32   `msg:"process"`
	Provider uint32   `msg:"provider"`
}

type Entry []Cap

type ListDirRes struct {
	Error   *Error           `msg:"error"`
	Entires map[string]Entry `msg:"entries"`
}

type SetDirReq struct {
	Abs    []string  `msg:"base"`
	Path   []string  `msg:"path"`
	Option bool      `msg:"option"`         // Create: replace, Remove: recurse
	Item   *Transfer `msg:"item,omitempty"` // valid: Create, empty: Remove
}

type SetDirRes struct {
	Error *Error    `msg:"error"`
	Item  *Transfer `msg:"item,omitempty"`
}
