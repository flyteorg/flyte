package config

import (
	stdconfig "github.com/flyteorg/flyte/v2/flytestdlib/config"
)

const (
	tasksSectionKey       = "tasks"
	taskPluginsSectionKey = "task-plugins"
)

var (
	defaultTaskPluginConfig = &TaskPluginConfig{}

	tasksSection       = stdconfig.MustRegisterSection(tasksSectionKey, &TasksConfig{})
	taskPluginsSection = tasksSection.MustRegisterSection(taskPluginsSectionKey, defaultTaskPluginConfig)
)

// TasksConfig is the top-level tasks configuration container.
type TasksConfig struct{}

// TaskPluginConfig holds the task plugin routing configuration,
// read from the tasks.task-plugins config section.
type TaskPluginConfig struct {
	// DefaultForTaskTypes overrides which plugin handles a given task type.
	// Key: task type string, Value: plugin ID string.
	DefaultForTaskTypes map[string]string `json:"default-for-task-types"`
}

// GetTaskPluginConfig returns the parsed task plugin configuration from the
// tasks.task-plugins config section.
func GetTaskPluginConfig() *TaskPluginConfig {
	return taskPluginsSection.GetConfig().(*TaskPluginConfig)
}

// SetTaskPluginConfig replaces the active task plugin configuration. Intended for use in tests.
func SetTaskPluginConfig(cfg *TaskPluginConfig) error {
	return taskPluginsSection.SetConfig(cfg)
}
