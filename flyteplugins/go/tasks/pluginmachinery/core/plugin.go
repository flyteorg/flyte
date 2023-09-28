package core

import (
	"context"
	"fmt"
)

//go:generate mockery -all -case=underscore

// https://github.com/flyteorg/flytepropeller/blob/979fabe1d1b22b01645259a03b8096f227681d08/pkg/utils/encoder.go#L25-L26
const minGeneratedNameLength = 8

type TaskType = string

// A Lazy loading function, that will load the plugin. Plugins should be initialized in this method. It is guaranteed
// that the plugin loader will be called before any Handle/Abort/Finalize functions are invoked
type PluginLoader func(ctx context.Context, iCtx SetupContext) (Plugin, error)

// An entry that identifies the CorePlugin
type PluginEntry struct {
	// System wide unique identifier for the plugin
	ID TaskType
	// A list of all the task types for which this plugin is applicable.
	RegisteredTaskTypes []TaskType
	// A Lazy loading function, that will load the plugin. Plugins should be initialized in this method. It is guaranteed
	// that the plugin loader will be called before any Handle/Abort/Finalize functions are invoked
	LoadPlugin PluginLoader
	// Boolean that indicates if this plugin can be used as the default for unknown task types. There can only be
	// one default in the system
	IsDefault bool
}

// System level properties that this Plugin supports
type PluginProperties struct {
	// Instructs the execution engine to not attempt to cache lookup or write for the node.
	DisableNodeLevelCaching bool
	// Specifies the length of TaskExecutionID generated name. default: 50
	GeneratedNameMaxLength *int
}

// Interface for the core Flyte plugin
type Plugin interface {
	// Unique ID for the plugin, should be ideally the same the ID in PluginEntry
	GetID() string
	// Properties desired by the plugin from the available set
	GetProperties() PluginProperties
	// The actual method that is invoked for every single task execution. The method should be a non blocking method.
	// It maybe invoked multiple times and hence all actions should be idempotent. If idempotency is not possible look at
	// Transitions to get some system level guarantees
	Handle(ctx context.Context, tCtx TaskExecutionContext) (Transition, error)
	// Called when the task is to be killed/aborted, because the top level entity was aborted or some other failure happened.
	// Abort should always be idempotent
	Abort(ctx context.Context, tCtx TaskExecutionContext) error
	// Finalize is always called, after Handle or Abort. Finalize should be an idempotent operation
	Finalize(ctx context.Context, tCtx TaskExecutionContext) error
}

// Loads and validates a plugin.
func LoadPlugin(ctx context.Context, iCtx SetupContext, entry PluginEntry) (Plugin, error) {
	plugin, err := entry.LoadPlugin(ctx, iCtx)
	if err != nil {
		return nil, err
	}

	length := plugin.GetProperties().GeneratedNameMaxLength
	if length != nil && *length < minGeneratedNameLength {
		return nil, fmt.Errorf("GeneratedNameMaxLength needs to be greater then %d", minGeneratedNameLength)
	}

	return plugin, err
}
