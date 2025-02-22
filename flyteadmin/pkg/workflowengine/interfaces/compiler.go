package interfaces

import (
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flytepropeller/pkg/compiler"
	"github.com/flyteorg/flyte/flytepropeller/pkg/compiler/common"
)

//go:generate mockery-v2 --name=Compiler --output=../mocks --case=underscore

// Workflow compiler interface.
type Compiler interface {
	CompileTask(task *core.TaskTemplate) (*core.CompiledTask, error)
	GetRequirements(fg *core.WorkflowTemplate, subWfs []*core.WorkflowTemplate) (
		compiler.WorkflowExecutionRequirements, error)
	CompileWorkflow(primaryWf *core.WorkflowTemplate, subworkflows []*core.WorkflowTemplate, tasks []*core.CompiledTask,
		launchPlans []common.InterfaceProvider) (*core.CompiledWorkflowClosure, error)
}
