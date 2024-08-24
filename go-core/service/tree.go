package service

import (
	"context"
	"io/fs"

	"github.com/azurity/mini-sos/go-core/node"
	"github.com/azurity/mini-sos/go-core/service/capability"
	"github.com/azurity/mini-sos/go-core/service/structs"
	"github.com/azurity/mini-sos/go-core/utils/syncx"
	"github.com/azurity/mini-sos/go-core/utils/tree"
	"github.com/google/uuid"
	"github.com/vmihailenco/msgpack/v5"
)

type capDesc struct {
	host     node.HostID
	process  uint32
	provider uint32
	// INFO: add priv info here.
}

type callFnType = func(host node.HostID, process uint32, provider uint32, data []byte, ctx context.Context) ([]byte, error)

type serviceTree struct {
	abs    tree.Path
	caps   syncx.Map[uuid.UUID, *capDesc]
	callFn callFnType
	local  bool
}

func callCap[Arg any, Ret any](callFn callFnType, cap *capDesc, arg *Arg, ctx context.Context) (*Ret, error) {
	argRaw, err := msgpack.Marshal(arg)
	if err != nil {
		return nil, err
	}
	retRaw, err := callFn(cap.host, cap.process, cap.provider, argRaw, ctx)
	if err != nil {
		return nil, err
	}
	var ret = new(Ret)
	err = msgpack.Unmarshal(retRaw, ret)
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func NewServiceTree(abs tree.Path, callFn func(host node.HostID, process uint32, provider uint32, data []byte, ctx context.Context) ([]byte, error), local bool) *serviceTree {
	return &serviceTree{
		abs:    abs.Normalize(),
		caps:   syncx.Map[uuid.UUID, *capDesc]{},
		callFn: callFn,
		local:  local,
	}
}

func fromEntry(name string, entry structs.Entry, base *serviceTree) *serviceTree {
	ret := &serviceTree{
		abs:    tree.Join(base.Abs(), tree.Path{name}),
		caps:   syncx.Map[uuid.UUID, *capDesc]{},
		callFn: base.callFn,
	}
	for _, cap := range entry {
		ret.AddCap(cap.Name, cap.Host, cap.Process, cap.Provider)
	}
	return ret
}

func (stree *serviceTree) toEntry() structs.Entry {
	ret := structs.Entry{}
	stree.caps.Range(func(key uuid.UUID, value *capDesc) bool {
		ret = append(ret, structs.Cap{
			Name:     key,
			Host:     value.host,
			Process:  value.process,
			Provider: value.provider,
		})
		return true
	})
	return ret
}

func (stree *serviceTree) Abs() tree.Path {
	return tree.NewPath(stree.abs.Str())
}

func (stree *serviceTree) Find(path tree.Path, ctx context.Context) (tree.Tree, error) {
	if !path.IsRel() {
		return nil, fs.ErrInvalid
	}
	if len(path) == 1 {
		return stree, nil
	}
	if cap, ok := stree.caps.Load(capability.CapListDir); ok {
		res, err := callCap[structs.ListDirReq, structs.ListDirRes](stree.callFn, cap, &structs.ListDirReq{
			Abs:  stree.abs,
			Path: path[:len(path)-1],
		}, ctx)
		if err != nil {
			return nil, err
		}
		if res.Error.Err() != nil {
			return nil, err
		}
		item := path[len(path)-1]
		if entry, ok := res.Entires[item]; ok {
			return fromEntry(item, entry, stree), nil
		}
	}
	return nil, fs.ErrInvalid
}

func (stree *serviceTree) Create(path tree.Path, subTree tree.Tree, replace bool, ctx context.Context) (old tree.Tree, err error) {
	if len(path) == 1 || !path.IsRel() {
		return nil, fs.ErrInvalid
	}
	if len(path) == 2 {
		if cap, ok := stree.caps.Load(capability.CapListDir); ok {
			transfer, err := subTree.Transfer()
			if err != nil {
				return nil, err
			}
			if transfer == nil {
				return nil, fs.ErrInvalid
			}
			res, err := callCap[structs.SetDirReq, structs.SetDirRes](stree.callFn, cap, &structs.SetDirReq{
				Abs:    stree.abs,
				Path:   path,
				Option: replace,
				Item:   transfer,
			}, ctx)
			if err != nil {
				return nil, err
			}
			if res.Error.Err() != nil {
				return nil, res.Error.Err()
			}
			if res.Item == nil {
				return nil, nil
			}
			ret, err := tree.ExtractTransfer(*res.Item)
			if err != nil {
				return nil, err
			}
			if cased, ok := ret.(*serviceTree); ok {
				cased.callFn = stree.callFn
			}
			return ret, nil
		}
		return nil, fs.ErrInvalid
	}
	aimTree, err := stree.Find(path[:len(path)-1], ctx)
	if err != nil {
		return nil, err
	}
	return aimTree.Create(tree.Path{".", path[len(path)-1]}, subTree, replace, ctx)
}

func (stree *serviceTree) Remove(path tree.Path, recurse bool, ctx context.Context) error {
	if !path.IsRel() {
		return fs.ErrInvalid
	}
	if len(path) <= 2 {
		if cap, ok := stree.caps.Load(capability.CapListDir); ok {
			res, err := callCap[structs.SetDirReq, structs.SetDirRes](stree.callFn, cap, &structs.SetDirReq{
				Abs:    stree.abs,
				Path:   path,
				Option: recurse,
				Item:   nil,
			}, ctx)
			if err != nil {
				return err
			}
			if res.Error.Err() != nil {
				return res.Error.Err()
			}
			return nil
		}
		return fs.ErrInvalid
	}
	aimTree, err := stree.Find(path[:len(path)-1], ctx)
	if err != nil {
		return err
	}
	return aimTree.Remove(tree.Path{".", path[len(path)-1]}, recurse, ctx)
}

func (stree *serviceTree) List(ctx context.Context) (map[string]tree.Tree, error) {
	if cap, ok := stree.caps.Load(capability.CapListDir); ok {
		res, err := callCap[structs.ListDirReq, structs.ListDirRes](stree.callFn, cap, &structs.ListDirReq{
			Abs:  stree.abs,
			Path: tree.Path{"."},
		}, ctx)
		if err != nil {
			return nil, err
		}
		if res.Error.Err() != nil {
			return nil, res.Error.Err()
		}
		ret := map[string]tree.Tree{}
		for name, entry := range res.Entires {
			ret[name] = fromEntry(name, entry, stree)
		}
		return ret, nil
	}
	return nil, fs.ErrInvalid
}

type serivceTreeTransfer struct {
	path string        `msgpack:"path"`
	caps []structs.Cap `msgpack:"caps"`
}

func (stree *serviceTree) Transfer() (*tree.Transfer, error) {
	if stree.local {
		return nil, nil
	}
	desc := serivceTreeTransfer{
		path: stree.abs.Str(),
		caps: []structs.Cap{},
	}
	stree.caps.Range(func(key uuid.UUID, value *capDesc) bool {
		desc.caps = append(desc.caps, structs.Cap{
			Name:     key,
			Host:     value.host,
			Process:  value.process,
			Provider: value.provider,
		})
		return true
	})
	raw, err := msgpack.Marshal(desc)
	if err != nil {
		return nil, err
	}
	return &tree.Transfer{
		SubType: "service",
		Desc:    raw,
	}, nil
}

func init() {
	tree.Extractors["service"] = func(rm msgpack.RawMessage) (tree.Tree, error) {
		parsed := serivceTreeTransfer{}
		if err := msgpack.Unmarshal(rm, parsed); err != nil {
			return nil, err
		}
		path := tree.NewPath(parsed.path).Normalize()
		if path.IsRel() {
			return nil, tree.ErrWrongPath
		}
		ret := &serviceTree{
			abs:    path,
			caps:   syncx.Map[uuid.UUID, *capDesc]{},
			callFn: nil,
		}
		for _, cap := range parsed.caps {
			if err := ret.AddCap(cap.Name, cap.Host, cap.Process, cap.Provider); err != nil {
				return nil, err
			}
		}
		return ret, nil
	}
}

func (stree *serviceTree) Merge(other tree.Tree, ctx context.Context) error {
	return tree.ErrUnsupported
}

func (stree *serviceTree) Leaves() []tree.Tree {
	return []tree.Tree{stree}
}

func (tree *serviceTree) AddCap(cap uuid.UUID, host uuid.UUID, process uint32, provider uint32) error {
	if _, ok := tree.caps.LoadOrStore(cap, &capDesc{
		host:     host,
		process:  process,
		provider: provider,
	}); ok {
		return ErrExist
	}
	return nil
}

func (tree *serviceTree) DelCap(cap uuid.UUID) {
	tree.caps.Delete(cap)
}

func (tree *serviceTree) CallCap(cap uuid.UUID, data []byte, ctx context.Context) ([]byte, error) {
	desc, ok := tree.caps.Load(cap)
	if !ok {
		return nil, ErrNotExist
	}
	return tree.callFn(desc.host, desc.process, desc.provider, data, ctx)
}
