package backoff

import "sync"

type HandlerMap struct {
	sync.Map
}

func (m *HandlerMap) Set(key string, value *ComputeResourceAwareBackOffHandler) {
	m.Store(key, value)
}

func (m *HandlerMap) Get(key string) (*ComputeResourceAwareBackOffHandler, bool) {
	value, found := m.Load(key)
	if !found {
		return nil, false
	}
	h, ok := value.(*ComputeResourceAwareBackOffHandler)
	return h, found && ok
}
