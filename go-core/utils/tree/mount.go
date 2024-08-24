package tree

import (
	"context"
	"io/fs"
	"sort"

	"github.com/azurity/mini-sos/go-core/utils/syncx"
	"github.com/vmihailenco/msgpack/v5"
)

type MountTree struct {
	mounts syncx.Map[string, Tree]
}

func NewMountTree(root Tree) *MountTree {
	tree := &MountTree{
		mounts: syncx.Map[string, Tree]{},
	}
	tree.mounts.Store(".", root)
	return tree
}

func (tree *MountTree) Abs() Path {
	return NewPath("/")
}

func (tree *MountTree) Mount(subTree Tree, path Path, ctx context.Context) (Tree, error) {
	_, err := tree.Create(path, subTree, false, ctx)
	if err != nil {
		return nil, err
	}
	tree.mounts.Store(path.Str(), subTree)
	return nil, nil
}

func (tree *MountTree) Find(path Path, ctx context.Context) (Tree, error) {
	relpath := path.Normalize()
	if !relpath.IsRel() {
		return nil, fs.ErrInvalid
	}
	subPath := Path{"."}
	for len(relpath) >= 1 {
		if sub, ok := tree.mounts.Load(relpath.Str()); ok {
			return sub.Find(subPath, ctx)
		}
		subPath = append(Path{".", relpath[len(relpath)-1]}, subPath[1:]...)
		relpath = relpath[:len(relpath)-1]
	}
	return nil, fs.ErrNotExist
}

func (tree *MountTree) Create(path Path, subTree Tree, replace bool, ctx context.Context) (old Tree, err error) {
	relpath := path.Normalize()
	if !relpath.IsRel() {
		return nil, fs.ErrInvalid
	}
	subPath := Path{"."}
	for len(relpath) >= 1 {
		if sub, ok := tree.mounts.Load(relpath.Str()); ok {
			return sub.Create(subPath, subTree, replace, ctx)
		}
		subPath = append(Path{".", relpath[len(relpath)-1]}, subPath[1:]...)
		relpath = relpath[:len(relpath)-1]
	}
	return nil, fs.ErrNotExist
}

func (tree *MountTree) Remove(path Path, recurse bool, ctx context.Context) error {
	relpath := path.Normalize()
	if !relpath.IsRel() {
		return fs.ErrInvalid
	}
	subPath := Path{"."}
	for len(relpath) > 1 {
		if sub, ok := tree.mounts.Load(relpath.Str()); ok {
			err := sub.Remove(subPath, recurse, ctx)
			if err != nil {
				return err
			}
			if len(subPath) == 1 {
				tree.mounts.CompareAndDelete(relpath.Str(), sub)
			}
			return nil
		}
		subPath = append(Path{".", relpath[len(relpath)-1]}, subPath[1:]...)
		relpath = relpath[:len(relpath)-1]
	}
	return fs.ErrInvalid
}

func (tree *MountTree) List(ctx context.Context) (map[string]Tree, error) {
	var errFinal error
	ret := map[string]Tree{}
	tree.mounts.Range(func(key string, value Tree) bool {
		path := NewPath(key)
		if len(path) == 1 {
			listed, err := value.List(ctx)
			if err != nil {
				errFinal = err
				return false
			}
			for k, v := range listed {
				ret[k] = v
			}
		} else if len(path) == 2 {
			ret[path[1]] = value
		}
		return true
	})
	if errFinal != nil {
		return nil, errFinal
	}
	return ret, nil
}

func (tree *MountTree) Transfer() (*Transfer, error) {
	info := map[string]Transfer{}
	var totalError error
	tree.mounts.Range(func(key string, value Tree) bool {
		transfer, err := value.Transfer()
		if err != nil {
			totalError = err
			return false
		}
		if transfer != nil {
			info[key] = *transfer
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
		SubType: "mount",
		Desc:    data,
	}, nil
}

func mountExtractor(rm msgpack.RawMessage) (Tree, error) {
	info := map[string]Transfer{}
	err := msgpack.Unmarshal(rm, &info)
	if err != nil {
		return nil, err
	}
	ret := &MountTree{
		mounts: syncx.Map[string, Tree]{},
	}
	for name, value := range info {
		tree, err := ExtractTransfer(value)
		if err != nil {
			return nil, err
		}
		ret.mounts.Store(name, tree)
	}
	if _, ok := ret.mounts.Load("."); !ok {
		return nil, fs.ErrInvalid
	}
	return ret, nil
}

type subDesc struct {
	key   string
	value Tree
}

type subDescSorter []*subDesc

func (s subDescSorter) Len() int      { return len(s) }
func (s subDescSorter) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s subDescSorter) Less(i, j int) bool {
	return s[i].key < s[j].key
}

func (tree *MountTree) Merge(other Tree, ctx context.Context) error {
	cased, ok := other.(*MountTree)
	if !ok {
		return ErrUnsupported
	}
	items := subDescSorter{}
	cased.mounts.Range(func(key string, value Tree) bool {
		items = append(items, &subDesc{
			key:   key,
			value: value,
		})
		return true
	})
	sort.Sort(items)
	for _, item := range items {
		if loaded, ok := tree.mounts.LoadOrStore(item.key, item.value); ok {
			loaded.Merge(loaded, ctx)
		} else {
			tree.Create(NewPath(item.key), item.value, false, ctx)
		}
	}
	return nil
}

func (tree *MountTree) Leaves() []Tree {
	ret := []Tree{}
	tree.mounts.Range(func(key string, value Tree) bool {
		ret = append(ret, value.Leaves()...)
		return true
	})
	return ret
}
