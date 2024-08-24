package tree

import (
	"context"
	"io/fs"
	"sync/atomic"

	"github.com/azurity/mini-sos/go-core/utils/syncx"
	"github.com/vmihailenco/msgpack/v5"
)

type BasicTree struct {
	abs   Path
	items syncx.Map[string, Tree]
	size  atomic.Int32
}

func NewBasicTree(abs Path) *BasicTree {
	normalized := abs.Normalize()
	if normalized.IsRel() {
		return nil
	}
	return &BasicTree{
		abs:   normalized.Normalize(),
		items: syncx.Map[string, Tree]{},
		size:  atomic.Int32{},
	}
}

func (tree *BasicTree) Abs() Path {
	return NewPath(tree.abs.Str())
}

func (tree *BasicTree) Find(path Path, ctx context.Context) (Tree, error) {
	relpath := path.Normalize()
	if !relpath.IsRel() {
		return nil, fs.ErrInvalid
	}
	var node *BasicTree = tree
	for len(relpath) > 1 {
		relpath = relpath[1:]
		if sub, ok := node.items.Load(relpath[0]); ok {
			if len(relpath) == 1 {
				return sub, nil
			}
			if cased, ok := sub.(*BasicTree); ok {
				node = cased
				continue
			} else {
				return sub.Find(relpath[1:], ctx)
			}
		}
		return nil, fs.ErrNotExist
	}
	return tree, nil
}

func (tree *BasicTree) Create(path Path, subTree Tree, replace bool, ctx context.Context) (Tree, error) {
	relpath := path.Normalize()
	if !relpath.IsRel() {
		return nil, fs.ErrInvalid
	}
	var node *BasicTree = tree
	for len(relpath) > 1 {
		relpath = relpath[1:]
		if len(relpath) == 1 {
			if replace {
				if value, loaded := node.items.Swap(relpath[0], subTree); loaded {
					return value, nil
				}
			} else if _, ok := node.items.LoadOrStore(relpath[0], subTree); ok {
				return nil, fs.ErrExist
			}
			node.size.Add(1)
			return nil, nil
		}
		if sub, ok := node.items.Load(relpath[0]); ok {
			if cased, ok := sub.(*BasicTree); ok {
				node = cased
				continue
			} else {
				return sub.Create(relpath[1:], subTree, replace, ctx)
			}
		}
		return nil, fs.ErrNotExist
	}
	return nil, fs.ErrInvalid
}

func (tree *BasicTree) Remove(path Path, recurse bool, ctx context.Context) error {
	relpath := path.Normalize()
	if !relpath.IsRel() {
		return fs.ErrInvalid
	}
	if relpath[0] == "." {
		if recurse {
			var err error = nil
			tree.items.Range(func(key string, value Tree) bool {
				err = value.Remove(Path{"."}, recurse, ctx)
				return err == nil
			})
			return err
		} else if tree.size.Load() != 0 {
			return fs.ErrInvalid
		}
		return nil
	}
	var node *BasicTree = tree
	for len(relpath) > 1 {
		relpath = relpath[1:]
		if len(relpath) == 1 {
			break
		}
		if sub, ok := node.items.Load(relpath[0]); ok {
			if cased, ok := sub.(*BasicTree); ok {
				node = cased
				continue
			} else {
				return sub.Remove(append(Path{"."}, relpath[1:]...), recurse, ctx)
			}
		}
		break
	}
	if _, ok := node.items.LoadAndDelete(relpath[0]); ok {
		node.size.Add(-1)
	}
	return nil
}

func (tree *BasicTree) List(ctx context.Context) (map[string]Tree, error) {
	ret := map[string]Tree{}
	tree.items.Range(func(key string, value Tree) bool {
		ret[key] = value
		return true
	})
	return ret, nil
}

type basicTreeTransfer struct {
	path string              `msgpack:"path"`
	subs map[string]Transfer `msgpack:"subs"`
}

func (tree *BasicTree) Transfer() (*Transfer, error) {
	path := tree.abs.Str()
	info := basicTreeTransfer{
		path: path,
		subs: map[string]Transfer{},
	}
	var totalError error
	tree.items.Range(func(key string, value Tree) bool {
		transfer, err := value.Transfer()
		if err != nil {
			totalError = err
			return false
		}
		if transfer != nil {
			info.subs[key] = *transfer
		}
		return true
	})
	if totalError != nil {
		return nil, totalError
	}
	data, err := msgpack.Marshal(info)
	if err != nil {
		return nil, err
	}
	return &Transfer{
		SubType: "basic",
		Desc:    data,
	}, nil
}

func basicExtractor(rm msgpack.RawMessage) (Tree, error) {
	info := basicTreeTransfer{}
	err := msgpack.Unmarshal(rm, &info)
	if err != nil {
		return nil, err
	}
	ret := NewBasicTree(NewPath(info.path))
	if ret == nil {
		return nil, fs.ErrInvalid
	}
	for name, value := range info.subs {
		tree, err := ExtractTransfer(value)
		if err != nil {
			return nil, err
		}
		ret.items.Store(name, tree)
		ret.size.Add(1)
	}
	return ret, nil
}

func (tree *BasicTree) Merge(other Tree, ctx context.Context) error {
	cased, ok := other.(*BasicTree)
	if !ok {
		return ErrUnsupported
	}
	if tree.abs.Str() != cased.abs.Str() {
		return fs.ErrInvalid
	}
	var totalError error
	tree.items.Range(func(key string, value Tree) bool {
		if loaded, ok := tree.items.LoadOrStore(key, value); ok {
			err := loaded.Merge(value, ctx)
			if err != nil {
				totalError = err
				return false
			}
		}
		return true
	})
	return totalError
}

func (tree *BasicTree) Leaves() []Tree {
	ret := []Tree{}
	queue := []Tree{tree}
	for len(queue) > 0 {
		cur := queue[0].(*BasicTree)
		queue = queue[1:]
		cur.items.Range(func(key string, value Tree) bool {
			if _, ok := value.(*BasicTree); ok {
				queue = append(queue, value)
			} else {
				ret = append(ret, value.Leaves()...)
			}
			return true
		})
	}
	return ret
}
