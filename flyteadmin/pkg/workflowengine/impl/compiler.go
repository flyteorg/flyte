package impl

import (
	"google.golang.org/grpc/codes"

	"github.com/flyteorg/flyte/flyteadmin/pkg/errors"
	"github.com/flyteorg/flyte/flyteadmin/pkg/workflowengine/interfaces"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flytepropeller/pkg/compiler"
	"github.com/flyteorg/flyte/flytepropeller/pkg/compiler/common"
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
		return nil, errors.NewFlyteAdminError(codes.InvalidArgument, err.Error())
	}
	return compiledWorkflowClosure, nil
}

func NewCompiler() interfaces.Compiler {
	return &workflowCompiler{}
}
