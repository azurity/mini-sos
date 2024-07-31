package service

import "sync"

type LocalManager struct {
	Entries   sync.Map
	processes sync.Map
}

func NewLocalManager() *LocalManager {
	return &LocalManager{
		Entries:   sync.Map{},
		processes: sync.Map{},
	}
}

func (man *LocalManager) Register(entry string, proc Provider, local bool) error {
	if _, ok := man.Entries.LoadOrStore(entry, proc.Id()); ok {
		return ErrExist
	}
	if _, ok := man.processes.LoadOrStore(proc.Id(), proc); !ok {
		go (func() {
			proc.WaitQuit()
			man.Entries.Range(func(key, value any) bool {
				man.Entries.CompareAndDelete(key, proc.Id())
				return true
			})
			man.processes.Delete(proc.Id())
		})()
	}
	return nil
}

func (man *LocalManager) Unregister(entry string, proc Provider, local bool) {
	man.Entries.CompareAndDelete(entry, proc.Id())
}

func (man *LocalManager) List(local bool) []string {
	ret := []string{}
	man.Entries.Range(func(key, value any) bool {
		ret = append(ret, key.(string))
		return true
	})
	return ret
}

func (man *LocalManager) Call(entry string, data []byte, caller Provider) ([]byte, error) {
	pid, ok := man.Entries.Load(entry)
	if !ok {
		return nil, ErrNotExist
	}
	proc, ok := man.processes.Load(pid)
	if !ok {
		return nil, ErrNotExist
	}
	return proc.(Provider).CallService(entry, data, caller)
}
