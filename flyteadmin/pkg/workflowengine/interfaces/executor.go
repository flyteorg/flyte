package interfaces

import (
	"context"
	"time"

	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/lyft/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
)

type ExecuteWorkflowInputs struct {
	ExecutionID *core.WorkflowExecutionIdentifier
	WfClosure   core.CompiledWorkflowClosure
	Inputs      *core.LiteralMap
	Reference   admin.LaunchPlan
	AcceptedAt  time.Time
	Labels      map[string]string
	Annotations map[string]string
}

type FlyteWorkflowInterface interface {
	BuildFlyteWorkflow(
		wfClosure *core.CompiledWorkflowClosure, inputs *core.LiteralMap, executionID *core.WorkflowExecutionIdentifier,
		namespace string) (*v1alpha1.FlyteWorkflow, error)
}

type Executor interface {
	ExecuteWorkflow(
		ctx context.Context, inputs ExecuteWorkflowInputs) error
	TerminateWorkflowExecution(ctx context.Context, executionID *core.WorkflowExecutionIdentifier) error
}
