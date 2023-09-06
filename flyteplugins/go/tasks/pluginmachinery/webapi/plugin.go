// Defines the  interfaces to implement to add a Web API Plugin (AsyncPlugin and SyncPlugin) to the Flyte system. A
// WebAPI plugin is a plugin that runs the compute for a task on a separate system through a web call (REST/Grpc...
// etc.). By implementing either the AsyncPlugin or SyncPlugin interfaces, the users of the Flyte system can then
// declare tasks of the handled task type in their workflows and the engine (Propeller) will route these tasks to your
// plugin to interact with the remote system.
//
// A sample skeleton plugin is defined in the example directory.
package webapi

import (
	"context"

	"github.com/flyteorg/flytestdlib/storage"

	"github.com/flyteorg/flytestdlib/promutils"

	pluginsCore "github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/io"
)

//go:generate mockery -all -case=underscore

// A Lazy loading function, that will load the plugin. Plugins should be initialized in this method. It is guaranteed
// that the plugin loader will be called before any Handle/Abort/Finalize functions are invoked
type PluginLoader func(ctx context.Context, iCtx PluginSetupContext) (AsyncPlugin, error)

// PluginEntry is a structure that is used to indicate to the system a K8s plugin
type PluginEntry struct {
	// ID/Name of the plugin. This will be used to identify this plugin and has to be unique in the entire system
	// All functions like enabling and disabling a plugin use this ID
	ID pluginsCore.TaskType

	// A list of all the task types for which this plugin is applicable.
	SupportedTaskTypes []pluginsCore.TaskType

	// An instance of the plugin
	PluginLoader PluginLoader

	// Boolean that indicates if this plugin can be used as the default for unknown task types. There can only be
	// one default in the system
	IsDefault bool

	// A list of all task types for which this plugin should be default handler when multiple registered plugins
	// support the same task type. This must be a subset of RegisteredTaskTypes and at most one default per task type
	// is supported.
	DefaultForTaskTypes []pluginsCore.TaskType
}

// PluginSetupContext is the interface made available to the plugin loader when initializing the plugin.
type PluginSetupContext interface {
	// a metrics scope to publish stats under
	MetricsScope() promutils.Scope
}

type TaskExecutionContextReader interface {
	// Returns a secret manager that can retrieve configured secrets for this plugin
	SecretManager() pluginsCore.SecretManager

	// Returns a TaskReader, to retrieve the task details
	TaskReader() pluginsCore.TaskReader

	// Returns an input reader to retrieve input data
	InputReader() io.InputReader

	// Returns a handle to the Task's execution metadata.
	TaskExecutionMetadata() pluginsCore.TaskExecutionMetadata

	// Provides an output sync of type io.OutputWriter
	OutputWriter() io.OutputWriter
}

type TaskExecutionContext interface {
	TaskExecutionContextReader

	// Provides the raw datastore to enable persisting outputs.
	DataStore() *storage.DataStore

	// Returns the max allowed dataset size that the outputwriter will accept
	MaxDatasetSizeBytes() int64
}

type GetContext interface {
	ResourceMeta() ResourceMeta
}

type DeleteContext interface {
	ResourceMeta() ResourceMeta
	Reason() string
}

type StatusContext interface {
	TaskExecutionContext

	ResourceMeta() ResourceMeta
	Resource() Resource
}

// Metadata about the resource to be synced from the remote service.
type ResourceMeta = interface{}
type Resource = interface{}

// AsyncPlugin defines the interface for plugins that call Async Web APIs.
type AsyncPlugin interface {
	// GetConfig gets the loaded plugin config. This will be used to control the interactions with the remote service.
	GetConfig() PluginConfig

	// ResourceRequirements analyzes the task to execute and determines the ResourceNamespace to be used when allocating tokens.
	ResourceRequirements(ctx context.Context, tCtx TaskExecutionContextReader) (
		namespace pluginsCore.ResourceNamespace, constraints pluginsCore.ResourceConstraintsSpec, err error)

	// Create a new resource using the TaskExecutionContext provided. Ideally, the remote service uses the name in the
	// TaskExecutionMetadata to launch the resource in an idempotent fashion. This function will be on the critical path
	// of the execution of a workflow and therefore it should not do expensive operations before making the webAPI call.
	// Flyte will call this api at least once. It's important that the callee service is idempotent to ensure no
	// resource leakage or duplicate requests. Flyte has an in-memory cache that does a best effort idempotency
	// guarantee.
	// It's required to return a resourceMeta object (that will be cached in memory). In case the remote API returns the
	// actually created resource, it's advisable to also return that in optionalResource output parameter. Doing so will
	// instruct the system to call Status() immediately after Create() and potentially terminate early if the resource
	// has already been executed/failed.
	// If the remote API failed due to a system error (network failure, timeout... etc.), the plugin should return a
	// non-nil error. The system will automatically retry the operation based on the plugin config.
	Create(ctx context.Context, tCtx TaskExecutionContextReader) (resourceMeta ResourceMeta, optionalResource Resource,
		err error)

	// Get the resource that matches the keys. If the plugin hits any failure, it should stop and return the failure.
	// This API will be called asynchronously and periodically to update the set of tasks currently in progress. It's
	// acceptable if this API is blocking since it'll be called from a background go-routine.
	// Best practices:
	//  1) Instead of returning the entire response object retrieved from the WebAPI, construct a smaller object that
	//     has enough information to construct the status/phase, error and/or output.
	//  2) This object will NOT be serialized/marshaled. It's, therefore, not a requirement to make it so.
	//  3) There is already client-side throttling in place. If the WebAPI returns a throttling error, you should return
	//     it as is so that the appropriate metrics are updated and the system administrator can update throttling
	//     params accordingly.
	Get(ctx context.Context, tCtx GetContext) (latest Resource, err error)

	// Delete the object in the remote service using the resource key. Flyte will call this API at least once. If the
	// resource has already been deleted, the API should not fail.
	Delete(ctx context.Context, tCtx DeleteContext) error

	// Status checks the status of a given resource and translates it to a Flyte-understandable PhaseInfo. This API
	// should avoid making any network calls and should run very efficiently.
	Status(ctx context.Context, tCtx StatusContext) (phase pluginsCore.PhaseInfo, err error)
}

// SyncPlugin defines the interface for plugins that call Web APIs synchronously.
type SyncPlugin interface {
	// GetConfig gets the loaded plugin config. This will be used to control the interactions with the remote service.
	GetConfig() PluginConfig

	// Do performs the action associated with this plugin.
	Do(ctx context.Context, tCtx TaskExecutionContext) (phase pluginsCore.PhaseInfo, err error)
}
