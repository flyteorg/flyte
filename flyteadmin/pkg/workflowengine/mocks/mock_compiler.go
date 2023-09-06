package mocks

import (
	"github.com/flyteorg/flyteadmin/pkg/workflowengine/interfaces"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flytepropeller/pkg/compiler"
	"github.com/flyteorg/flytepropeller/pkg/compiler/common"
)

type MockCompiler struct {
	compileTaskCallback     func(task *core.TaskTemplate) (*core.CompiledTask, error)
	getRequirementsCallback func(fg *core.WorkflowTemplate, subWfs []*core.WorkflowTemplate) (
		reqs compiler.WorkflowExecutionRequirements, err error)
	compileWorkflowCallback func(
		primaryWf *core.WorkflowTemplate, subworkflows []*core.WorkflowTemplate, tasks []*core.CompiledTask,
		launchPlans []common.InterfaceProvider) (*core.CompiledWorkflowClosure, error)
}

func (c *MockCompiler) CompileTask(task *core.TaskTemplate) (*core.CompiledTask, error) {
	if c.compileTaskCallback != nil {
		return c.compileTaskCallback(task)
	}
	return nil, nil
}

func (c *MockCompiler) AddCompileTaskCallback(
	callback func(task *core.TaskTemplate) (*core.CompiledTask, error)) {
	c.compileTaskCallback = callback
}

func (c *MockCompiler) GetRequirements(fg *core.WorkflowTemplate, subWfs []*core.WorkflowTemplate) (
	reqs compiler.WorkflowExecutionRequirements, err error) {
	if c.getRequirementsCallback != nil {
		return c.getRequirementsCallback(fg, subWfs)
	}
	return compiler.WorkflowExecutionRequirements{}, nil
}

func (c *MockCompiler) AddGetRequirementCallback(
	callback func(fg *core.WorkflowTemplate, subWfs []*core.WorkflowTemplate) (
		reqs compiler.WorkflowExecutionRequirements, err error)) {
	c.getRequirementsCallback = callback
}

func (c *MockCompiler) CompileWorkflow(
	primaryWf *core.WorkflowTemplate, subworkflows []*core.WorkflowTemplate, tasks []*core.CompiledTask,
	launchPlans []common.InterfaceProvider) (*core.CompiledWorkflowClosure, error) {
	if c.compileWorkflowCallback != nil {
		return c.compileWorkflowCallback(primaryWf, subworkflows, tasks, launchPlans)
	}
	return nil, nil
}

func (c *MockCompiler) AddCompileWorkflowCallback(callback func(
	primaryWf *core.WorkflowTemplate, subworkflows []*core.WorkflowTemplate, tasks []*core.CompiledTask,
	launchPlans []common.InterfaceProvider) (*core.CompiledWorkflowClosure, error)) {
	c.compileWorkflowCallback = callback
}

func NewMockCompiler() interfaces.Compiler {
	return &MockCompiler{}
}
