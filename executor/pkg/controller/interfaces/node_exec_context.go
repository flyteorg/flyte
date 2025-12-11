package interfaces

import (
	"context"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	v1Api "github.com/flyteorg/flyte/v2/executor/api/v1"
	"github.com/flyteorg/flyte/v2/executor/events"
	"github.com/flyteorg/flyte/v2/executor/pkg/controller/executors"
	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/io"
	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/ioutils"
	"github.com/flyteorg/flyte/v2/flytestdlib/storage"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/core"
)

//go:generate mockery --all --case=underscore --with-expecter

type TaskReader interface {
	Read(ctx context.Context) (*core.TaskTemplate, error)
	GetTaskType() v1Api.TaskType
	GetTaskID() *core.Identifier
}

type EventRecorder interface {
	events.TaskEventRecorder
	events.NodeEventRecorder
}

type NodeExecutionMetadata interface {
	GetOwnerID() types.NamespacedName
	GetNodeExecutionID() *core.NodeExecutionIdentifier
	GetNamespace() string
	GetOwnerReference() v1.OwnerReference
	GetLabels() map[string]string
	GetAnnotations() map[string]string
	GetK8sServiceAccount() string
	GetSecurityContext() core.SecurityContext
	IsInterruptible() bool
	GetInterruptibleFailureThreshold() int32
	GetConsoleURL() string
}

type NodeExecutionContext interface {
	// This path is never read by propeller, but allows using some container or prefix in a specific container for all output from tasks
	// Sandboxes provide exactly once execution semantics and only the successful sandbox wins. Ideally a sandbox should be a path that is
	// available to the task at High Bandwidth (for example the base path of a sharded s3 bucket.
	// This with a prefix based sharded strategy, could improve the throughput from S3 manifold)
	RawOutputPrefix() storage.DataReference

	// Sharding strategy for the output data for this node execution.
	OutputShardSelector() ioutils.ShardSelector

	DataStore() *storage.DataStore
	InputReader() io.InputReader
	EventsRecorder() EventRecorder
	NodeID() v1Api.NodeID
	Node() v1Api.ExecutableNode
	CurrentAttempt() uint32
	TaskReader() TaskReader

	NodeStateReader() NodeStateReader
	NodeStateWriter() NodeStateWriter

	NodeExecutionMetadata() NodeExecutionMetadata

	EnqueueOwnerFunc() func() error

	ContextualNodeLookup() executors.NodeLookup
	ExecutionContext() executors.ExecutionContext
	// TODO We should not need to pass NodeStatus, we probably only need it for DataDir, which should actually be sent using an OutputWriter interface
	// Deprecated
	NodeStatus() v1Api.ExecutableNodeStatus
}
