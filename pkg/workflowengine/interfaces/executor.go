package interfaces

import (
	"context"
	"time"

	runtime "github.com/flyteorg/flyteadmin/pkg/runtime/interfaces"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"

	"github.com/flyteorg/flytestdlib/storage"
)

//go:generate mockery -name=WorkflowExecutor -output=../mocks/ -case=underscore

type TaskResources struct {
	Defaults runtime.TaskResourceSet
	Limits   runtime.TaskResourceSet
}

type ExecutionParameters struct {
	Inputs              *core.LiteralMap
	AcceptedAt          time.Time
	Labels              map[string]string
	Annotations         map[string]string
	TaskPluginOverrides []*admin.PluginOverride
	ExecutionConfig     *admin.WorkflowExecutionConfig
	RecoveryExecution   *core.WorkflowExecutionIdentifier
	TaskResources       *TaskResources
	EventVersion        int
	RoleNameKey         string
	RawOutputDataConfig *admin.RawOutputDataConfig
	ClusterAssignment   *admin.ClusterAssignment
}

// ExecutionData includes all parameters required to create an execution CRD object.
type ExecutionData struct {
	// Execution namespace.
	Namespace string
	// Execution identifier.
	ExecutionID *core.WorkflowExecutionIdentifier
	// Underlying workflow name for the execution.
	ReferenceWorkflowName string
	// Launch plan name used to trigger the execution.
	ReferenceLaunchPlanName string
	// Compiled workflow closure used to build the flyte workflow
	WorkflowClosure *core.CompiledWorkflowClosure
	// Storage Data Reference for the WorkflowClosure
	WorkflowClosureReference storage.DataReference
	// Additional parameters used to build a workflow execution
	ExecutionParameters ExecutionParameters
}

// ExecutionResponse is returned when a Flyte workflow execution is successfully created.
type ExecutionResponse struct {
	// Cluster identifier where the execution was created
	Cluster string
}

// AbortData includes all parameters required to abort an execution CRD object.
type AbortData struct {
	// Execution namespace.
	Namespace string
	// Execution identifier.
	ExecutionID *core.WorkflowExecutionIdentifier
	// Cluster identifier where the execution was created
	Cluster string
}

// WorkflowExecutor is a client interface used to create and delete Flyte workflow CRD objects.
type WorkflowExecutor interface {
	// ID returns the unique name of this executor implementation.
	ID() string
	// Execute creates a Flyte workflow execution CRD object.
	Execute(ctx context.Context, data ExecutionData) (ExecutionResponse, error)
	// Abort aborts a running Flyte workflow execution CRD object.
	Abort(ctx context.Context, data AbortData) error
}
