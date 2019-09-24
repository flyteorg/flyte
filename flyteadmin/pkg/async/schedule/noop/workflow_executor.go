package noop

import "github.com/lyft/flyteadmin/pkg/async/schedule/interfaces"

type workflowExecutor struct{}

func (w *workflowExecutor) Run() {}

func (w *workflowExecutor) Stop() error {
	return nil
}

func NewWorkflowExecutor() interfaces.WorkflowExecutor {
	return &workflowExecutor{}
}
