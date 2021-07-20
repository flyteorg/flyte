package launchplan

import (
	"context"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
)

//go:generate mockery -all -case=underscore

// A simple context that is used to start an execution of a LaunchPlan. It encapsulates enough parent information
// to tie the executions
type LaunchContext struct {
	// Nesting level of the current workflow (parent)
	NestingLevel uint32
	// Principal of the current workflow, so that billing can be tied correctly
	Principal string
	// If a node launched the execution, this specifies which node execution
	ParentNodeExecution *core.NodeExecutionIdentifier
	// If a node in recovery mode launched this execution, propagate recovery mode to the child execution.
	RecoveryExecution *core.WorkflowExecutionIdentifier
}

// Interface to be implemented by the remote system that can allow workflow launching capabilities
type Executor interface {
	// Start an execution of a launchplan
	Launch(ctx context.Context, launchCtx LaunchContext, executionID *core.WorkflowExecutionIdentifier, launchPlanRef *core.Identifier, inputs *core.LiteralMap) error

	// Retrieve status of a LaunchPlan execution
	GetStatus(ctx context.Context, executionID *core.WorkflowExecutionIdentifier) (*admin.ExecutionClosure, error)

	// Kill a remote execution
	Kill(ctx context.Context, executionID *core.WorkflowExecutionIdentifier, reason string) error

	// Initializes Executor.
	Initialize(ctx context.Context) error
}

type Reader interface {
	// Get the definition of a launch plan. This is primarily used to ensure all the TypedInterfaces match up before actually executing.
	GetLaunchPlan(ctx context.Context, launchPlanRef *core.Identifier) (*admin.LaunchPlan, error)
}

type FlyteAdmin interface {
	Executor
	Reader
}
