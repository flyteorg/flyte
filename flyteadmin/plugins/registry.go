package plugins

import (
	"fmt"
	"sync"
	"sync/atomic"
)

type PluginID = string

const (
	PluginIDWorkflowExecutor       PluginID = "WorkflowExecutor"
	PluginIDDataProxy              PluginID = "DataProxy"
	PluginIDUnaryServiceMiddleware PluginID = "UnaryServiceMiddleware"
	PluginIDPreRedirectHook        PluginID = "PreRedirectHook"
)

type AtomicRegistry struct {
	atomic.Value
}

// Store stores the Registry to be retrieved later.
func (a *AtomicRegistry) Store(r *Registry) {
	a.Value.Store(r)
}

// Load loads the stored Registry or nil if non exists
func (a *AtomicRegistry) Load() *Registry {
	return a.Value.Load().(*Registry)
}

// NewAtomicRegistry creates a new envelope to hold a Registry object.
func NewAtomicRegistry(initialValue *Registry) AtomicRegistry {
	val := atomic.Value{}
	val.Store(initialValue)
	return AtomicRegistry{
		Value: val,
	}
}

// Registry is a generic plugin registrar for dependency injection.
type Registry struct {
	m        sync.Map
	mDefault sync.Map
}

// Register registers a new implementation for the pluginID. Only one plugin is allowed to be registered
// for a given ID.
func (r *Registry) Register(id PluginID, impl interface{}) error {
	_, loaded := r.m.LoadOrStore(id, impl)
	if loaded {
		return fmt.Errorf("an existing implementation has already been registered for [%v]", id)
	}

	return nil
}

// RegisterDefault registers a new implementation for the pluginID. This is the implementation that will be used
// if no other plugin is registered for the ID.
func (r *Registry) RegisterDefault(id PluginID, impl interface{}) {
	r.mDefault.Store(id, impl)
}

// Get retrieves a registered implementation for the ID. If one doesn't exist, it returns the default implementation.
// If the id isn't found, it returns nil.
func Get[T any](r *Registry, id PluginID) T {
	obj := r.Get(id)
	res, casted := obj.(T)
	if !casted {
		var zeroVal T
		return zeroVal
	}

	return res
}

// Get retrieves a registered implementation for the ID. If one doesn't exist, it returns the default implementation.
// If the id isn't found, it returns nil.
func (r *Registry) Get(id PluginID) interface{} {
	val, exists := r.m.Load(id)
	if exists {
		return val
	}

	val, exists = r.mDefault.Load(id)
	if exists {
		return val
	}

	return nil
}

// NewRegistry creates a new Registry
func NewRegistry() *Registry {
	return &Registry{
		m:        sync.Map{},
		mDefault: sync.Map{},
	}
}
