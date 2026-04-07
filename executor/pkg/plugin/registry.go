package plugin

import (
	"context"
	"fmt"
	"sync"

	pluginsCore "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/core"
	k8sPlugin "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/k8s"
	"github.com/flyteorg/flyte/v2/flytestdlib/logger"

	executorK8s "github.com/flyteorg/flyte/v2/executor/pkg/plugin/k8s"
)

// PluginRegistryIface provides access to registered plugin entries.
type PluginRegistryIface interface {
	GetCorePlugins() []pluginsCore.PluginEntry
	GetK8sPlugins() []k8sPlugin.PluginEntry
}

// Registry resolves a pluginsCore.Plugin for a given task type by wrapping the global
// plugin registry. K8s plugins are wrapped in a PluginManager; core plugins are loaded
// via their PluginLoader.
type Registry struct {
	mu sync.RWMutex

	setupCtx       pluginsCore.SetupContext
	pluginRegistry PluginRegistryIface

	// taskType -> pluginsCore.Plugin
	plugins       map[string]pluginsCore.Plugin
	defaultPlugin pluginsCore.Plugin
	initialized   bool
}

// NewRegistry creates a Registry backed by the given plugin source and setup context.
func NewRegistry(setupCtx pluginsCore.SetupContext, pluginRegistry PluginRegistryIface) *Registry {
	return &Registry{
		setupCtx:       setupCtx,
		pluginRegistry: pluginRegistry,
		plugins:        make(map[string]pluginsCore.Plugin),
	}
}

// Initialize loads all registered plugins. Must be called once during startup.
func (r *Registry) Initialize(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.initialized {
		return nil
	}

	// Load k8s plugins
	for _, entry := range r.pluginRegistry.GetK8sPlugins() {
		pm := executorK8s.NewPluginManager(
			entry.ID,
			entry.Plugin,
			r.setupCtx.KubeClient(),
		)
		if err := pm.InitializeObjectEventWatcher(ctx); err != nil {
			return fmt.Errorf("failed to initialize k8s object event watcher for plugin %s: %w", entry.ID, err)
		}

		for _, taskType := range entry.RegisteredTaskTypes {
			if existing, ok := r.plugins[taskType]; ok {
				logger.Warnf(ctx, "Task type %q already registered by plugin %q, overwriting with %q",
					taskType, existing.GetID(), entry.ID)
			}
			r.plugins[taskType] = pm
		}

		if entry.IsDefault {
			if r.defaultPlugin != nil {
				logger.Warnf(ctx, "Multiple default plugins found, overwriting %q with %q",
					r.defaultPlugin.GetID(), entry.ID)
			}
			r.defaultPlugin = pm
		}

		logger.Infof(ctx, "Registered k8s plugin [%s] for task types %v", entry.ID, entry.RegisteredTaskTypes)
	}

	// Load core plugins
	for _, entry := range r.pluginRegistry.GetCorePlugins() {
		plugin, err := pluginsCore.LoadPlugin(ctx, r.setupCtx, entry)
		if err != nil {
			return fmt.Errorf("failed to load core plugin %s: %w", entry.ID, err)
		}

		for _, taskType := range entry.RegisteredTaskTypes {
			if existing, ok := r.plugins[taskType]; ok {
				logger.Warnf(ctx, "Task type %q already registered by plugin %q, overwriting with %q",
					taskType, existing.GetID(), entry.ID)
			}
			r.plugins[taskType] = plugin
		}

		if entry.IsDefault && r.defaultPlugin == nil {
			r.defaultPlugin = plugin
		}

		logger.Infof(ctx, "Registered core plugin [%s] for task types %v", entry.ID, entry.RegisteredTaskTypes)
	}

	r.initialized = true
	return nil
}

// ResolvePlugin returns the plugin registered for the given task type.
// Falls back to the default plugin if no specific match is found.
func (r *Registry) ResolvePlugin(taskType string) (pluginsCore.Plugin, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if p, ok := r.plugins[taskType]; ok {
		return p, nil
	}

	if r.defaultPlugin != nil {
		return r.defaultPlugin, nil
	}

	return nil, fmt.Errorf("no plugin registered for task type %q and no default plugin available", taskType)
}
