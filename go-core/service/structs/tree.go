package structs

import (
	"github.com/azurity/mini-sos/go-core/utils/tree"
	"github.com/google/uuid"
)

type ListDirReq struct {
	Abs  []string `msgpack:"base"`
	Path []string `msgpack:"path"`
}

type Cap struct {
	Name     uuid.UUID `msgpack:"name"`
	Host     uuid.UUID `msgpack:"host"`
	Process  uint32    `msgpack:"process"`
	Provider uint32    `msgpack:"provider"`
}

type Entry = []Cap

type ListDirRes struct {
	Error   *Error           `msgpack:"error"`
	Entires map[string]Entry `msgpack:"entries"`
}

type SetDirReq struct {
	Abs    []string       `msgpack:"base"`
	Path   []string       `msgpack:"path"`
	Option bool           `msgpack:"option"`         // Create: replace, Remove: recurse
	Item   *tree.Transfer `msgpack:"item,omitempty"` // valid: Create, empty: Remove
}

type SetDirRes struct {
	Error *Error         `msgpack:"error"`
	Item  *tree.Transfer `msgpack:"item,omitempty"`
}
