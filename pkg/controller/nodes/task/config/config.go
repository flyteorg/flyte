package config

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/flyteorg/flytestdlib/logger"
	"k8s.io/apimachinery/pkg/util/sets"

	"github.com/flyteorg/flytestdlib/config"
)

//go:generate pflags Config

const SectionKey = "tasks"

var (
	defaultConfig = &Config{
		TaskPlugins:            TaskPluginConfig{EnabledPlugins: []string{}, DefaultForTaskTypes: map[string]string{}},
		MaxPluginPhaseVersions: 100000,
		BackOffConfig: BackOffConfig{
			BaseSecond:  2,
			MaxDuration: config.Duration{Duration: time.Second * 20},
		},
	}

	section = config.MustRegisterSection(SectionKey, defaultConfig)
)

type Config struct {
	TaskPlugins            TaskPluginConfig `json:"task-plugins" pflag:",Task plugin configuration"`
	MaxPluginPhaseVersions int32            `json:"max-plugin-phase-versions" pflag:",Maximum number of plugin phase versions allowed for one phase."`
	BackOffConfig          BackOffConfig    `json:"backoff" pflag:",Config for Exponential BackOff implementation"`
	MaxErrorMessageLength  int              `json:"maxLogMessageLength" pflag:",Deprecated!!! Max length of error message."`
}

type TaskPluginConfig struct {
	EnabledPlugins []string `json:"enabled-plugins" pflag:",Plugins enabled currently"`
	// Maps task types to their plugin handler (by ID).
	DefaultForTaskTypes map[string]string `json:"default-for-task-types" pflag:"-,"`
}

type BackOffConfig struct {
	BaseSecond  int             `json:"base-second" pflag:",The number of seconds representing the base duration of the exponential backoff"`
	MaxDuration config.Duration `json:"max-duration" pflag:",The cap of the backoff duration"`
}

type PluginID = string
type TaskType = string

// Contains the set of enabled plugins for this flytepropeller deployment along with default plugin handlers
// for specific task types.
type PluginsConfigMeta struct {
	EnabledPlugins         sets.String
	AllDefaultForTaskTypes map[PluginID][]TaskType
}

func cleanString(source string) string {
	cleaned := strings.Trim(source, " ")
	cleaned = strings.ToLower(cleaned)
	return cleaned
}

func (p TaskPluginConfig) GetEnabledPlugins() (PluginsConfigMeta, error) {
	enabledPluginsNames := sets.NewString()
	for _, pluginName := range p.EnabledPlugins {
		cleanedPluginName := cleanString(pluginName)
		enabledPluginsNames.Insert(cleanedPluginName)
	}

	pluginDefaultForTaskType := make(map[PluginID][]TaskType)
	// Reverse the DefaultForTaskTypes map. Having the config use task type as a key guarantees only one default plugin can be specified per
	// task type but now we need to sort for which tasks a plugin needs to be the default.
	for taskName, pluginName := range p.DefaultForTaskTypes {
		existing, found := pluginDefaultForTaskType[pluginName]
		if !found {
			existing = make([]string, 0, 1)
		}
		pluginDefaultForTaskType[cleanString(pluginName)] = append(existing, cleanString(taskName))
	}

	// All plugins are enabled, nothing further to validate here.
	if enabledPluginsNames.Len() == 0 {
		return PluginsConfigMeta{
			EnabledPlugins:         enabledPluginsNames,
			AllDefaultForTaskTypes: pluginDefaultForTaskType,
		}, nil
	}

	// Finally, validate that default plugins for task types only reference enabled plugins
	for pluginName, taskTypes := range pluginDefaultForTaskType {
		if !enabledPluginsNames.Has(pluginName) {
			logger.Errorf(context.TODO(), "Cannot set default plugin [%s] for task types [%+v] when it is not "+
				"configured to be an enabled plugin. Please double check the flytepropeller config.", pluginName, taskTypes)
			return PluginsConfigMeta{}, fmt.Errorf("cannot set default plugin [%s] for task types [%+v] when it is not "+
				"configured to be an enabled plugin", pluginName, taskTypes)
		}
	}
	return PluginsConfigMeta{
		EnabledPlugins:         enabledPluginsNames,
		AllDefaultForTaskTypes: pluginDefaultForTaskType,
	}, nil
}

func GetConfig() *Config {
	return section.GetConfig().(*Config)
}
