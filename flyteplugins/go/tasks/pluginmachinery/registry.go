package pluginmachinery

import (
	"context"
	"sync"

	"k8s.io/apimachinery/pkg/runtime"

	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/k8s"
	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/webapi"
	"github.com/flyteorg/flyte/v2/flytestdlib/logger"

	internalRemote "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/internal/webapi"
)

// SchemeRegistration associates a plugin ID with its AddToScheme function.
type SchemeRegistration struct {
	ID          string
	AddToScheme func(*runtime.Scheme) error
}

type taskPluginRegistry struct {
	m               sync.Mutex
	k8sPlugin       []k8s.PluginEntry
	corePlugin      []core.PluginEntry
	schemeRegisters []SchemeRegistration
}

// A singleton variable that maintains a registry of all plugins. The framework uses this to access all plugins
var pluginRegistry = &taskPluginRegistry{}

func PluginRegistry() TaskPluginRegistry {
	return pluginRegistry
}

func (p *taskPluginRegistry) RegisterRemotePlugin(info webapi.PluginEntry) {
	ctx := context.Background()
	if info.ID == "" {
		logger.Panicf(ctx, "ID is required attribute for k8s plugin")
	}

	if len(info.SupportedTaskTypes) == 0 {
		logger.Panicf(ctx, "AsyncPlugin should be registered to handle at least one task type")
	}

	if info.PluginLoader == nil {
		logger.Panicf(ctx, "PluginLoader cannot be nil")
	}

	p.m.Lock()
	defer p.m.Unlock()
	p.corePlugin = append(p.corePlugin, internalRemote.CreateRemotePlugin(info))
}

func CreateRemotePlugin(pluginEntry webapi.PluginEntry) core.PluginEntry {
	return internalRemote.CreateRemotePlugin(pluginEntry)
}

// Use this method to register Kubernetes Plugins
func (p *taskPluginRegistry) RegisterK8sPlugin(info k8s.PluginEntry) {
	if info.ID == "" {
		logger.Panicf(context.TODO(), "ID is required attribute for k8s plugin")
	}

	if len(info.RegisteredTaskTypes) == 0 {
		logger.Panicf(context.TODO(), "K8s AsyncPlugin should be registered to handle at least one task type")
	}

	if info.Plugin == nil && info.ClusterPlugin == nil {
		logger.Panicf(context.TODO(), "K8s plugin requires either Plugin or ClusterPlugin to be set")
	}

	if info.Plugin != nil && info.ClusterPlugin != nil {
		logger.Panicf(context.TODO(), "K8s plugin cannot set both Plugin and ClusterPlugin")
	}

	if info.ResourceToWatch == nil {
		logger.Panicf(context.TODO(), "The framework requires a K8s resource to watch, for valid plugin registration")
	}

	if info.ClusterPlugin != nil && info.ClusterResourceToWatch == nil {
		logger.Panicf(context.TODO(), "ClusterPlugin requires ClusterResourceToWatch to be set")
	}

	p.m.Lock()
	defer p.m.Unlock()
	p.k8sPlugin = append(p.k8sPlugin, info)
}

// Use this method to register core plugins
func (p *taskPluginRegistry) RegisterCorePlugin(info core.PluginEntry) {
	if info.ID == "" {
		logger.Panicf(context.TODO(), "ID is required attribute for k8s plugin")
	}
	if len(info.RegisteredTaskTypes) == 0 {
		logger.Panicf(context.TODO(), "AsyncPlugin should be registered to handle at least one task type")
	}
	if info.LoadPlugin == nil {
		logger.Panicf(context.TODO(), "PluginLoader cannot be nil")
	}

	p.m.Lock()
	defer p.m.Unlock()
	p.corePlugin = append(p.corePlugin, info)
}

// RegisterScheme registers an AddToScheme function for a plugin's CRD types.
func (p *taskPluginRegistry) RegisterScheme(id string, addToScheme func(*runtime.Scheme) error) {
	p.m.Lock()
	defer p.m.Unlock()
	p.schemeRegisters = append(p.schemeRegisters, SchemeRegistration{ID: id, AddToScheme: addToScheme})
}

// GetSchemeRegisters returns a snapshot of all registered SchemeRegistrations.
func (p *taskPluginRegistry) GetSchemeRegisters() []SchemeRegistration {
	p.m.Lock()
	defer p.m.Unlock()
	return append(p.schemeRegisters[:0:0], p.schemeRegisters...)
}

// Returns a snapshot of all the registered core plugins.
func (p *taskPluginRegistry) GetCorePlugins() []core.PluginEntry {
	p.m.Lock()
	defer p.m.Unlock()
	return append(p.corePlugin[:0:0], p.corePlugin...)
}

// Returns a snapshot of all registered K8s plugins
func (p *taskPluginRegistry) GetK8sPlugins() []k8s.PluginEntry {
	p.m.Lock()
	defer p.m.Unlock()
	return append(p.k8sPlugin[:0:0], p.k8sPlugin...)
}

type TaskPluginRegistry interface {
	RegisterK8sPlugin(info k8s.PluginEntry)
	RegisterCorePlugin(info core.PluginEntry)
	RegisterRemotePlugin(info webapi.PluginEntry)
	RegisterScheme(id string, addToScheme func(*runtime.Scheme) error)
	GetCorePlugins() []core.PluginEntry
	GetK8sPlugins() []k8s.PluginEntry
	GetSchemeRegisters() []SchemeRegistration
}
