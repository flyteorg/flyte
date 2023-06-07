package task

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/flyteorg/flytepropeller/pkg/controller/nodes/task/backoff"

	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/flyteorg/flytestdlib/logger"

	"github.com/flyteorg/flyteplugins/go/tasks/plugins/webapi/agent"
	"github.com/flyteorg/flytepropeller/pkg/controller/nodes/task/config"
	"github.com/flyteorg/flytepropeller/pkg/controller/nodes/task/k8s"
)

var once sync.Once

func WranglePluginsAndGenerateFinalList(ctx context.Context, cfg *config.TaskPluginConfig, pr PluginRegistryIface) (enabledPlugins []core.PluginEntry, defaultForTaskTypes map[pluginID][]taskType, err error) {
	if cfg == nil {
		return nil, nil, fmt.Errorf("unable to initialize plugin list, cfg is a required argument")
	}

	// Register the GRPC plugin after the config is loaded
	once.Do(func() { agent.RegisterAgentPlugin() })
	pluginsConfigMeta, err := cfg.GetEnabledPlugins()
	if err != nil {
		return nil, nil, err
	}

	allPluginsEnabled := false
	if pluginsConfigMeta.EnabledPlugins.Len() == 0 {
		allPluginsEnabled = true
	}

	logger.Infof(ctx, "Enabled plugins: %v", pluginsConfigMeta.EnabledPlugins.List())
	logger.Infof(ctx, "Loading core Plugins, plugin configuration [all plugins enabled: %v]", allPluginsEnabled)
	for _, cpe := range pr.GetCorePlugins() {
		id := strings.ToLower(cpe.ID)
		if !allPluginsEnabled && !pluginsConfigMeta.EnabledPlugins.Has(id) {
			logger.Infof(ctx, "Plugin [%s] is DISABLED (not found in enabled plugins list).", id)
		} else {
			logger.Infof(ctx, "Plugin [%s] ENABLED", id)
			enabledPlugins = append(enabledPlugins, cpe)
		}
	}

	// Create a single backOffManager for all the plugins
	backOffController := backoff.NewController(ctx)

	// Create a single resource monitor object for all plugins to use
	monitorIndex := k8s.NewResourceMonitorIndex()

	k8sPlugins := pr.GetK8sPlugins()
	for i := range k8sPlugins {
		kpe := k8sPlugins[i]
		id := strings.ToLower(kpe.ID)
		if !allPluginsEnabled && !pluginsConfigMeta.EnabledPlugins.Has(id) {
			logger.Infof(ctx, "K8s Plugin [%s] is DISABLED (not found in enabled plugins list).", id)
		} else {
			logger.Infof(ctx, "K8s Plugin [%s] is ENABLED.", id)
			plugin := core.PluginEntry{
				ID:                  id,
				RegisteredTaskTypes: kpe.RegisteredTaskTypes,
				LoadPlugin: func(ctx context.Context, iCtx core.SetupContext) (plugin core.Plugin, e error) {
					return k8s.NewPluginManagerWithBackOff(ctx, iCtx, kpe, backOffController, monitorIndex)
				},
				IsDefault: kpe.IsDefault,
			}
			enabledPlugins = append(enabledPlugins, plugin)
		}
	}
	return enabledPlugins, pluginsConfigMeta.AllDefaultForTaskTypes, nil
}
