package pluginmachinery

import (
	"context"
	"sync"

	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/core"
	internalRemote "github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/internal/webapi"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/k8s"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/webapi"
	"github.com/flyteorg/flyte/flytestdlib/logger"
)

const defaultPluginBufferSize = 100

type PluginInfo struct {
	VersionedTaskType string
	DeploymentID string
}

type taskPluginRegistry struct {
	m          sync.Mutex
	k8sPlugin  []k8s.PluginEntry
	corePlugin []core.PluginEntry
	connectorCorePlugin map[string]map[string]core.PluginEntry
	pluginChan chan PluginInfo
}

// A singleton variable that maintains a registry of all plugins. The framework uses this to access all plugins
var pluginRegistry = &taskPluginRegistry{
	corePlugin: []core.PluginEntry{},
	connectorCorePlugin: make(map[string]map[string]core.PluginEntry),
	pluginChan:  make(chan PluginInfo, defaultPluginBufferSize),
}

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

// RegisterConnectorCorePlugin registers a core plugin for a specific connector deployment
func (p *taskPluginRegistry) RegisterConnectorCorePlugin(info webapi.PluginEntry, deploymentID string) {
	ctx := context.Background()
	if info.ID == "" {
		logger.Panicf(ctx, "ID is required attribute for connector core plugin")
	}

	if len(info.SupportedTaskTypes) == 0 {
		logger.Panicf(ctx, "Plugin should be registered to handle at least one task type")
	}

	p.m.Lock()
	defer p.m.Unlock()

	if p.connectorCorePlugin == nil {
		p.connectorCorePlugin = make(map[string]map[string]core.PluginEntry)
	}

	if p.connectorCorePlugin[info.ID] == nil {
		p.connectorCorePlugin[info.ID] = make(map[string]core.PluginEntry)
	}

	p.connectorCorePlugin[info.ID][deploymentID] = internalRemote.CreateRemotePlugin(info)
}

// GetConnectorCorePlugin returns a specific connector core plugin for a task type and deployment ID
func (p *taskPluginRegistry) GetConnectorCorePlugin(taskType string, deploymentID string) (core.PluginEntry, bool) {
	p.m.Lock()
	defer p.m.Unlock()

	if p.connectorCorePlugin == nil {
		return core.PluginEntry{}, false
	}

	if plugins, exists := p.connectorCorePlugin[taskType]; exists {
		if plugin, exists := plugins[deploymentID]; exists {
			return plugin, true
		}
	}

	return core.PluginEntry{}, false
}

func (p *taskPluginRegistry) IsConnectorCorePluginRegistered(taskType string, deploymentID string) bool {
	p.m.Lock()
	defer p.m.Unlock()

	if p.connectorCorePlugin == nil {
		return false
	}

	if plugins, exists := p.connectorCorePlugin[taskType]; exists {
		if _, exists := plugins[deploymentID]; exists {
			return true
		}
	}

	return false
}

func (p *taskPluginRegistry) GetPluginChan() chan PluginInfo {
	return p.pluginChan
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

	if info.Plugin == nil {
		logger.Panicf(context.TODO(), "K8s AsyncPlugin cannot be nil")
	}

	if info.ResourceToWatch == nil {
		logger.Panicf(context.TODO(), "The framework requires a K8s resource to watch, for valid plugin registration")
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
	RegisterConnectorCorePlugin(info webapi.PluginEntry, deploymentID string)
	GetConnectorCorePlugin(taskType string, deploymentID string) (core.PluginEntry, bool)
	IsConnectorCorePluginRegistered(taskType string, deploymentID string) bool
	GetCorePlugins() []core.PluginEntry
	GetK8sPlugins() []k8s.PluginEntry
	GetPluginChan() chan PluginInfo
}
