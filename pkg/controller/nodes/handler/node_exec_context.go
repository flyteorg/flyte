package handler

import (
	"context"

	"github.com/lyft/flyteidl/clients/go/events"
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/lyft/flyteplugins/go/tasks/pluginmachinery/io"
	"github.com/lyft/flytestdlib/promutils"
	"github.com/lyft/flytestdlib/storage"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/lyft/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
)

type TaskReader interface {
	Read(ctx context.Context) (*core.TaskTemplate, error)
	GetTaskType() v1alpha1.TaskType
	GetTaskID() *core.Identifier
}

type SetupContext interface {
	EnqueueOwner() func(string)
	OwnerKind() string
	MetricsScope() promutils.Scope
}

type NodeExecutionMetadata interface {
	GetOwnerID() types.NamespacedName
	// TODO we should covert this to a generic execution identifier instead of a workflow identifier
	GetExecutionID() v1alpha1.WorkflowExecutionIdentifier
	GetNamespace() string
	GetOwnerReference() v1.OwnerReference
	GetLabels() map[string]string
	GetAnnotations() map[string]string
	GetK8sServiceAccount() string
	IsInterruptible() bool
}

type NodeExecutionContext interface {
	DataStore() *storage.DataStore
	InputReader() io.InputReader
	EventsRecorder() events.TaskEventRecorder
	NodeID() v1alpha1.NodeID
	Node() v1alpha1.ExecutableNode
	CurrentAttempt() uint32
	TaskReader() TaskReader

	NodeStateReader() NodeStateReader
	NodeStateWriter() NodeStateWriter

	NodeExecutionMetadata() NodeExecutionMetadata
	MaxDatasetSizeBytes() int64

	EnqueueOwnerFunc() func() error

	// Deprecated
	Workflow() v1alpha1.ExecutableWorkflow
	// TODO We should not need to pass NodeStatus, we probably only need it for DataDir, which should actually be sent using an OutputWriter interface
	// Deprecated
	NodeStatus() v1alpha1.ExecutableNodeStatus
}
