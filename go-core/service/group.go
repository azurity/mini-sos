package service

import "github.com/azurity/mini-sos/go-core/node"

type GroupManager struct {
	Instance map[node.HostID]Manager
}

func (man *GroupManager) Register(entry string, proc Provider, local bool) error {
	for _, item := range man.Instance {
		err := item.Register(entry, proc, local)
		if err != nil {
			return err
		}
	}
	return nil
}

func (man *GroupManager) Unregister(entry string, proc Provider, local bool) {
	for _, item := range man.Instance {
		item.Unregister(entry, proc, local)
	}
}

func (man *GroupManager) List(local bool) []string {
	unique := map[string]interface{}{}
	for _, item := range man.Instance {
		for _, key := range item.List(local) {
			unique[key] = nil
		}
	}
	ret := []string{}
	for key := range unique {
		ret = append(ret, key)
	}
	return ret
}

func (man *GroupManager) Call(entry string, data []byte, caller Provider) ([]byte, error) {
	for _, item := range man.Instance {
		ret, err := item.Call(entry, data, caller)
		if err == ErrNotExist {
			continue
		}
		return ret, err
	}
	return nil, ErrNotExist
}
