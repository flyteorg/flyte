package mocks

import (
	"context"

	"github.com/lyft/flyteadmin/pkg/workflowengine/interfaces"
)

type ExecuteWorkflowFunc func(input interfaces.ExecuteWorkflowInput) (*interfaces.ExecutionInfo, error)
type ExecuteTaskFunc func(ctx context.Context, input interfaces.ExecuteTaskInput) (*interfaces.ExecutionInfo, error)
type TerminateWorkflowExecutionFunc func(ctx context.Context, input interfaces.TerminateWorkflowInput) error

type MockExecutor struct {
	executeWorkflowCallback    ExecuteWorkflowFunc
	executeTaskCallback        ExecuteTaskFunc
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

func (c *MockExecutor) SetExecuteTaskCallback(callback ExecuteTaskFunc) {
	c.executeTaskCallback = callback
}

func (c *MockExecutor) ExecuteTask(ctx context.Context, input interfaces.ExecuteTaskInput) (*interfaces.ExecutionInfo, error) {
	if c.executeTaskCallback != nil {
		return c.executeTaskCallback(ctx, input)
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
