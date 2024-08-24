package tree

import (
	"context"
	"errors"

	"github.com/vmihailenco/msgpack/v5"
)

var ErrUnsupported = errors.New("unsupported")

type Tree interface {
	Abs() Path
	Find(path Path, ctx context.Context) (Tree, error)
	Create(path Path, subTree Tree, replace bool, ctx context.Context) (old Tree, err error)
	Remove(path Path, recurse bool, ctx context.Context) error
	List(ctx context.Context) (map[string]Tree, error)
	Transfer() (*Transfer, error)
	Merge(other Tree, ctx context.Context) error
	Leaves() []Tree
}

type Transfer struct {
	SubType string             `msgpack:"sub_type"`
	Desc    msgpack.RawMessage `msgpack:"desc"`
}

var ErrUnknownSubType = errors.New("unknown sub-type")
var ErrWrongPath = errors.New("wrong path")
var Extractors = map[string]func(msgpack.RawMessage) (Tree, error){}

func ExtractTransfer(transfer Transfer) (Tree, error) {
	fn, ok := Extractors[transfer.SubType]
	if !ok {
		return nil, ErrUnknownSubType
	}
	return fn(transfer.Desc)
}

func init() {
	Extractors["basic"] = basicExtractor
	Extractors["mount"] = mountExtractor
}
