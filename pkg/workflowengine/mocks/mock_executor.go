package mocks

import (
	"context"

	"github.com/lyft/flyteadmin/pkg/workflowengine/interfaces"

	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/core"
)

type CreateExecutionFunc func(inputs interfaces.ExecuteWorkflowInputs) error
type TerminateWorkflowExecutionFunc func(ctx context.Context, executionID *core.WorkflowExecutionIdentifier) error

type MockExecutor struct {
	executeWorkflowCallback    CreateExecutionFunc
	terminateExecutionCallback TerminateWorkflowExecutionFunc
}

func (c *MockExecutor) SetExecuteWorkflowCallback(callback CreateExecutionFunc) {
	c.executeWorkflowCallback = callback
}

func (c *MockExecutor) ExecuteWorkflow(
	ctx context.Context, inputs interfaces.ExecuteWorkflowInputs) error {
	if c.executeWorkflowCallback != nil {
		return c.executeWorkflowCallback(inputs)
	}
	return nil
}

func (c *MockExecutor) SetTerminateExecutionCallback(callback TerminateWorkflowExecutionFunc) {
	c.terminateExecutionCallback = callback
}

func (c *MockExecutor) TerminateWorkflowExecution(ctx context.Context, executionID *core.WorkflowExecutionIdentifier) error {
	if c.terminateExecutionCallback != nil {
		return c.terminateExecutionCallback(ctx, executionID)
	}
	return nil
}

func NewMockExecutor() interfaces.Executor {
	return &MockExecutor{}
}
