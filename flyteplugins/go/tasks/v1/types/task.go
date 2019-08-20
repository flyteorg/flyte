package types

import (
	"context"

	"github.com/lyft/flytestdlib/promutils"
	"k8s.io/apimachinery/pkg/types"

	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/event"

	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/lyft/flytestdlib/storage"
)

//go:generate mockery -name Executor

type TaskType = string
type WorkflowID = string
type VarName = string
type TaskExecutorName = string

type EnqueueOwner func(name types.NamespacedName) error

// Defines optional properties for the executor.
type ExecutorProperties struct {
	// If the executor needs to clean-up external resources that won't automatically be garbage-collected by the fact that
	// the containing-k8s object is being deleted, it should set this value to true. This ensures that the containing-k8s
	// object is not deleted until all executors of non-terminal phase tasks report success for KillTask calls.
	RequiresFinalizer bool

	// If set, the execution engine will not perform node-level task caching and retrieval. This can be useful for more
	// fine-grained executors that implement their own logic for caching.
	DisableNodeLevelCaching bool
}

// Defines the exposed interface for plugins to record task events.
// TODO: Add a link to explain how events are structred and linked together.
type EventRecorder interface {
	RecordTaskEvent(ctx context.Context, event *event.TaskExecutionEvent) error
}

// Defines the Catalog client interface exposed for plugins
type CatalogClient interface {
	Get(ctx context.Context, task *core.TaskTemplate, inputPath storage.DataReference) (*core.LiteralMap, error)
	Put(ctx context.Context, task *core.TaskTemplate, execID *core.TaskExecutionIdentifier, inputPath storage.DataReference, outputPath storage.DataReference) error
}

// Defines the all-optional initialization parameters passed to plugins.
type ExecutorInitializationParameters struct {
	CatalogClient CatalogClient
	EventRecorder EventRecorder
	DataStore     *storage.DataStore
	EnqueueOwner  EnqueueOwner
	OwnerKind     string
	MetricsScope  promutils.Scope
}

// Defines a task executor interface.
type Executor interface {
	// Gets a unique identifier for the executor. No two executors can have the same ID.
	GetID() TaskExecutorName

	// Gets optional properties about this executor. These properties are not task-specific.
	GetProperties() ExecutorProperties

	// Initializes the executor. The executor should not have any heavy initialization logic in its constructor and should
	// delay all initialization logic till this method is called.
	Initialize(ctx context.Context, param ExecutorInitializationParameters) error

	// Start the task with an initial state that could be empty and return the new state of the task once it started
	StartTask(ctx context.Context, taskCtx TaskContext, task *core.TaskTemplate, inputs *core.LiteralMap) (
		status TaskStatus, err error)

	// ChecksTaskStatus is called every time client needs to know the latest status of a given task this specific
	// executor launched. It passes the same task context that was used when StartTask was called as well as the last
	// known state of the task. Note that there is no strict guarantee that the previous state is literally the last
	// status returned due to the nature of eventual consistency in the system. The system guarantees idempotency as long
	// as it's within kubernetes boundaries or if external services support idempotency.
	CheckTaskStatus(ctx context.Context, taskCtx TaskContext, task *core.TaskTemplate) (status TaskStatus, err error)

	// The engine will ensure kill task is called in abort scenarios. KillTask will not be called in case CheckTaskStatus
	// ever returned a terminal phase.
	KillTask(ctx context.Context, taskCtx TaskContext, reason string) error

	// ResolveOutputs is responsible for retrieving outputs variables from a task. For simple tasks, adding OutputsResolver
	// in the executor is enough to get a default implementation.
	ResolveOutputs(ctx context.Context, taskCtx TaskContext, outputVariables ...VarName) (
		values map[VarName]*core.Literal, err error)
}

// Represents a free-form state that allows plugins to store custom information between invocations.
type CustomState = map[string]interface{}
