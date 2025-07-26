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

// PluginRegistrationInfo contains information about plugin registration
type PluginRegistrationInfo struct {
	Plugin   webapi.PluginEntry
	DeploymentID string
}

type PluginUpdateInfo struct {
	TaskType string
	DeploymentID string
}

type taskPluginRegistry struct {
	m          sync.Mutex
	k8sPlugin  []k8s.PluginEntry
	corePlugin []core.PluginEntry
	pluginRegistrationChan chan PluginRegistrationInfo
	pluginUpdateChan chan PluginUpdateInfo
	registeredPlugins map[string]map[string]struct{}
}

// A singleton variable that maintains a registry of all plugins. The framework uses this to access all plugins
var pluginRegistry = &taskPluginRegistry{
	pluginRegistrationChan: make(chan PluginRegistrationInfo, defaultPluginBufferSize),
	pluginUpdateChan:  make(chan PluginUpdateInfo, defaultPluginBufferSize),
	registeredPlugins: make(map[string]map[string]struct{}),
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

func (p *taskPluginRegistry) GetPluginRegistrationChan() chan PluginRegistrationInfo {
	return p.pluginRegistrationChan
}

func (p *taskPluginRegistry) GetPluginUpdateChan() chan PluginUpdateInfo {
	return p.pluginUpdateChan
}

// IsPluginForTaskTypeRegistered checks if a task type is registered
func (p *taskPluginRegistry) IsPluginForTaskTypeRegistered(taskType string, deploymentID string) bool {
	p.m.Lock()
	defer p.m.Unlock()
	
	if p.registeredPlugins == nil {
		return false
	}
	
	deploymentMap, exists := p.registeredPlugins[taskType]
	if !exists {
		return false
	}
	
	_, exists = deploymentMap[deploymentID]
	return exists
}

// RegisterTaskType registers a single task type
func (p *taskPluginRegistry) AddRegisteredPluginForTaskType(taskType string, deploymentID string) {
	p.m.Lock()
	defer p.m.Unlock()
	
	if p.registeredPlugins[taskType] == nil {
		p.registeredPlugins[taskType] = make(map[string]struct{})
	}
	
	p.registeredPlugins[taskType][deploymentID] = struct{}{}
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
	GetCorePlugins() []core.PluginEntry
	GetK8sPlugins() []k8s.PluginEntry
	GetPluginRegistrationChan() chan PluginRegistrationInfo
	GetPluginUpdateChan() chan PluginUpdateInfo
	IsPluginForTaskTypeRegistered(taskType string, deploymentID string) bool
	AddRegisteredPluginForTaskType(taskType string, deploymentID string)
}
