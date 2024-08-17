package service

import "sync"

type entryInfo struct {
	provider uint32
	id       uint32
}

type ProviderInfo struct {
	provider Provider
	entries  map[string]bool
}

type LocalManager struct {
	entries   sync.Map // TODO: change to tree
	providers sync.Map
}

func NewLocalManager() *LocalManager {
	return &LocalManager{
		entries:   sync.Map{},
		providers: sync.Map{},
	}
}

func (man *LocalManager) Register(entry string, provider Provider, id uint32, local bool) error {
	if _, ok := man.entries.LoadOrStore(entry, &entryInfo{provider: provider.Id(), id: id}); ok {
		return ErrExist
	}
	if _, ok := man.providers.LoadOrStore(provider.Id(), &ProviderInfo{provider: provider, entries: map[string]bool{}}); !ok {
		go (func() {
			provider.WaitQuit()
			value, _ := man.providers.LoadAndDelete(provider.Id())
			info := value.(*ProviderInfo)
			man.entries.Range(func(key, value any) bool {
				if _, ok := info.entries[key.(string)]; ok {
					man.entries.Delete(key)
				}
				return true
			})
		})()
	}
	info, _ := man.providers.Load(provider.Id())
	info.(*ProviderInfo).entries[entry] = true
	return nil
}

func (man *LocalManager) Update(entry string, proc Provider, id uint32, local bool) error {
	info, ok := man.entries.Load(entry)
	if !ok {
		return ErrNotExist
	}
	if info.(*entryInfo).provider != proc.Id() {
		return ErrIllegal
	}
	info.(*entryInfo).id = id
	return nil
}

func (man *LocalManager) Unregister(entry string, provider Provider, local bool) {
	if info, ok := man.providers.Load(provider.Id()); ok {
		if _, ok := info.(*ProviderInfo).entries[entry]; ok {
			man.entries.Delete(entry)
		}
	}
}

func (man *LocalManager) List(local bool) []string {
	ret := []string{}
	man.entries.Range(func(key, value any) bool {
		ret = append(ret, key.(string))
		return true
	})
	return ret
}

func (man *LocalManager) Call(entry string, data []byte, caller Provider) ([]byte, error) {
	info, ok := man.entries.Load(entry)
	if !ok {
		return nil, ErrNotExist
	}
	provider, ok := man.providers.Load(info.(*entryInfo).provider)
	if !ok {
		return nil, ErrNotExist
	}
	data, err := provider.(*ProviderInfo).provider.CallProvider(info.(*entryInfo).id, data, caller)
	if err != nil && err.Error() == "provider not exist" {
		err = ErrNotExist
	}
	return data, err
}
