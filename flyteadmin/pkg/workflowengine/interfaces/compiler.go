package interfaces

import (
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/lyft/flytepropeller/pkg/compiler"
	"github.com/lyft/flytepropeller/pkg/compiler/common"
)

// Workflow compiler interface.
type Compiler interface {
	CompileTask(task *core.TaskTemplate) (*core.CompiledTask, error)
	GetRequirements(fg *core.WorkflowTemplate, subWfs []*core.WorkflowTemplate) (
		compiler.WorkflowExecutionRequirements, error)
	CompileWorkflow(primaryWf *core.WorkflowTemplate, subworkflows []*core.WorkflowTemplate, tasks []*core.CompiledTask,
		launchPlans []common.InterfaceProvider) (*core.CompiledWorkflowClosure, error)
}
