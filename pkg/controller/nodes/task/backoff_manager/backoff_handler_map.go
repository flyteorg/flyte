package backoff_manager

import "sync"

type BackOffHandlerMap struct {
	sync.Map
}

func (m *BackOffHandlerMap) Set(key string, value *ComputeResourceAwareBackOffHandler) {
	m.Store(key, value)
}

func (m *BackOffHandlerMap) Get(key string) (*ComputeResourceAwareBackOffHandler, bool) {
	value, found := m.Load(key)
	if found == false {
		return nil, false
	} else {
		h, ok := value.(*ComputeResourceAwareBackOffHandler)
		return h, found && ok
	}
}
