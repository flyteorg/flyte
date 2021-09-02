package interfaces

import (
	"context"
	"time"

	runtime "github.com/flyteorg/flyteadmin/pkg/runtime/interfaces"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
)

type TaskResources struct {
	Defaults runtime.TaskResourceSet
	Limits   runtime.TaskResourceSet
}

type ExecuteWorkflowInput struct {
	ExecutionID         *core.WorkflowExecutionIdentifier
	WfClosure           core.CompiledWorkflowClosure
	Inputs              *core.LiteralMap
	Reference           admin.LaunchPlan
	AcceptedAt          time.Time
	Labels              map[string]string
	Annotations         map[string]string
	QueueingBudget      time.Duration
	TaskPluginOverrides []*admin.PluginOverride
	ExecutionConfig     *admin.WorkflowExecutionConfig
	Auth                *admin.AuthRole
	RecoveryExecution   *core.WorkflowExecutionIdentifier
	TaskResources       *TaskResources
}

type ExecuteTaskInput struct {
	ExecutionID         *core.WorkflowExecutionIdentifier
	WfClosure           core.CompiledWorkflowClosure
	Inputs              *core.LiteralMap
	ReferenceName       string
	Auth                *admin.AuthRole
	AcceptedAt          time.Time
	Labels              map[string]string
	Annotations         map[string]string
	QueueingBudget      time.Duration
	TaskPluginOverrides []*admin.PluginOverride
	ExecutionConfig     *admin.WorkflowExecutionConfig
	TaskResources       *TaskResources
}

type TerminateWorkflowInput struct {
	ExecutionID *core.WorkflowExecutionIdentifier
	Cluster     string
}

type ExecutionInfo struct {
	Cluster string
}

type FlyteWorkflowInterface interface {
	BuildFlyteWorkflow(
		wfClosure *core.CompiledWorkflowClosure, inputs *core.LiteralMap, executionID *core.WorkflowExecutionIdentifier,
		namespace string) (*v1alpha1.FlyteWorkflow, error)
}

type Executor interface {
	ExecuteWorkflow(
		ctx context.Context, input ExecuteWorkflowInput) (*ExecutionInfo, error)
	ExecuteTask(ctx context.Context, input ExecuteTaskInput) (*ExecutionInfo, error)
	TerminateWorkflowExecution(ctx context.Context, input TerminateWorkflowInput) error
}
