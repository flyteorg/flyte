package k8s

import (
	"context"

	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/flyteorg/flytestdlib/storage"

	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/io"

	pluginsCore "github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/core"
)

//go:generate mockery -all -case=underscore

// PluginEntry is a structure that is used to indicate to the system a K8s plugin
type PluginEntry struct {
	// ID/Name of the plugin. This will be used to identify this plugin and has to be unique in the entire system
	// All functions like enabling and disabling a plugin use this ID
	ID pluginsCore.TaskType
	// A list of all the task types for which this plugin is applicable.
	RegisteredTaskTypes []pluginsCore.TaskType
	// An instance of the kubernetes resource this plugin is responsible for, for example v1.Pod{}
	ResourceToWatch client.Object
	// An instance of the plugin
	Plugin Plugin
	// Boolean that indicates if this plugin can be used as the default for unknown task types. There can only be
	// one default in the system
	IsDefault bool
	// A list of all task types for which this plugin should be default handler when multiple registered plugins
	// support the same task type. This must be a subset of RegisteredTaskTypes and at most one default per task type
	// is supported.
	DefaultForTaskTypes []pluginsCore.TaskType
	// Returns a new KubeClient to be used instead of the internal controller-runtime client.
	CustomKubeClient func(ctx context.Context) (pluginsCore.KubeClient, error)
}

// System level properties that this Plugin supports
type PluginProperties struct {
	// Disables the inclusion of OwnerReferences in kubernetes resources that this plugin is responsible for.
	// Disabling is only useful if resources will be created in a remote cluster.
	DisableInjectOwnerReferences bool
	// Boolean that indicates if finalizer injection should be disabled for resources that this plugin is
	// responsible for.
	DisableInjectFinalizer bool
	// Specifies the length of TaskExecutionID generated name. default: 50
	GeneratedNameMaxLength *int
	// DisableDeleteResourceOnFinalize disables deleting the created resource on finalize. That behavior is controllable
	// on the base K8sPluginConfig level but can be disabled for individual plugins. Plugins should generally not
	// override that behavior unless the resource that gets created for this plugin does not consume resources (cluster's
	// cpu/memory... etc. or external resources) once the plugin's Plugin.GetTaskPhase() returns a terminal phase.
	DisableDeleteResourceOnFinalize bool
}

// Special context passed in to plugins when checking task phase
type PluginContext interface {
	// Returns a TaskReader, to retrieve task details
	TaskReader() pluginsCore.TaskReader

	// Returns an input reader to retrieve input data
	InputReader() io.InputReader

	// Provides an output sync of type io.OutputWriter
	OutputWriter() io.OutputWriter

	// Returns a handle to the currently configured storage backend that can be used to communicate with the tasks or write metadata
	DataStore() *storage.DataStore

	// Returns the max allowed dataset size that the outputwriter will accept
	MaxDatasetSizeBytes() int64

	// Returns a handle to the Task's execution metadata.
	TaskExecutionMetadata() pluginsCore.TaskExecutionMetadata
}

// Defines a simplified interface to author plugins for k8s resources.
type Plugin interface {
	// Defines a func to create a query object (typically just object and type meta portions) that's used to query k8s
	// resources.
	BuildIdentityResource(ctx context.Context, taskCtx pluginsCore.TaskExecutionMetadata) (client.Object, error)

	// Defines a func to create the full resource object that will be posted to k8s.
	BuildResource(ctx context.Context, taskCtx pluginsCore.TaskExecutionContext) (client.Object, error)

	// Analyses the k8s resource and reports the status as TaskPhase. This call is expected to be relatively fast,
	// any operations that might take a long time (limits are configured system-wide) should be offloaded to the
	// background.
	GetTaskPhase(ctx context.Context, pluginContext PluginContext, resource client.Object) (pluginsCore.PhaseInfo, error)

	// Properties desired by the plugin
	GetProperties() PluginProperties
}
