package impl

import (
	"github.com/flyteorg/flyteadmin/pkg/errors"
	"github.com/flyteorg/flyteadmin/pkg/workflowengine/interfaces"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flytepropeller/pkg/compiler"
	"github.com/flyteorg/flytepropeller/pkg/compiler/common"
	"google.golang.org/grpc/codes"
)

type workflowCompiler struct {
}

func (c *workflowCompiler) CompileTask(task *core.TaskTemplate) (*core.CompiledTask, error) {
	compiledTask, err := compiler.CompileTask(task)
	if err != nil {
		return nil, errors.NewFlyteAdminErrorf(codes.InvalidArgument, "task failed compilation with %v", err)
	}
	return compiledTask, nil
}

func (c *workflowCompiler) GetRequirements(fg *core.WorkflowTemplate, subWfs []*core.WorkflowTemplate) (
	compiler.WorkflowExecutionRequirements, error) {
	requirements, err := compiler.GetRequirements(fg, subWfs)
	if err != nil {
		return compiler.WorkflowExecutionRequirements{},
			errors.NewFlyteAdminErrorf(codes.InvalidArgument, "failed to get workflow requirements with err %v", err)
	}
	return requirements, nil
}

func (c *workflowCompiler) CompileWorkflow(
	primaryWf *core.WorkflowTemplate, subworkflows []*core.WorkflowTemplate, tasks []*core.CompiledTask,
	launchPlans []common.InterfaceProvider) (*core.CompiledWorkflowClosure, error) {

	compiledWorkflowClosure, err := compiler.CompileWorkflow(primaryWf, subworkflows, tasks, launchPlans)
	if err != nil {
		return nil, errors.NewFlyteAdminErrorf(codes.InvalidArgument, "failed to compile workflow with err %v", err)
	}
	return compiledWorkflowClosure, nil
}

func NewCompiler() interfaces.Compiler {
	return &workflowCompiler{}
}
