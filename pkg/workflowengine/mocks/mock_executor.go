package mocks

import (
	"context"

	"github.com/lyft/flyteadmin/pkg/workflowengine/interfaces"
)

type ExecuteWorkflowFunc func(input interfaces.ExecuteWorkflowInput) (*interfaces.ExecutionInfo, error)
type TerminateWorkflowExecutionFunc func(ctx context.Context, input interfaces.TerminateWorkflowInput) error

type MockExecutor struct {
	executeWorkflowCallback    ExecuteWorkflowFunc
	terminateExecutionCallback TerminateWorkflowExecutionFunc
}

func (c *MockExecutor) SetExecuteWorkflowCallback(callback ExecuteWorkflowFunc) {
	c.executeWorkflowCallback = callback
}

func (c *MockExecutor) ExecuteWorkflow(
	ctx context.Context, inputs interfaces.ExecuteWorkflowInput) (*interfaces.ExecutionInfo, error) {
	if c.executeWorkflowCallback != nil {
		return c.executeWorkflowCallback(inputs)
	}
	return &interfaces.ExecutionInfo{}, nil
}

func (c *MockExecutor) SetTerminateExecutionCallback(callback TerminateWorkflowExecutionFunc) {
	c.terminateExecutionCallback = callback
}

func (c *MockExecutor) TerminateWorkflowExecution(ctx context.Context, input interfaces.TerminateWorkflowInput) error {
	if c.terminateExecutionCallback != nil {
		return c.terminateExecutionCallback(ctx, input)
	}
	return nil
}

func NewMockExecutor() interfaces.Executor {
	return &MockExecutor{}
}
