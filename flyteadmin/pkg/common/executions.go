package common

import (
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
)

var terminalExecutionPhases = map[core.WorkflowExecution_Phase]bool{
	core.WorkflowExecution_SUCCEEDED: true,
	core.WorkflowExecution_FAILED:    true,
	core.WorkflowExecution_TIMED_OUT: true,
	core.WorkflowExecution_ABORTED:   true,
}

var terminalNodeExecutionPhases = map[core.NodeExecution_Phase]bool{
	core.NodeExecution_SUCCEEDED: true,
	core.NodeExecution_FAILED:    true,
	core.NodeExecution_TIMED_OUT: true,
	core.NodeExecution_ABORTED:   true,
	core.NodeExecution_SKIPPED:   true,
	core.NodeExecution_RECOVERED: true,
}

var terminalTaskExecutionPhases = map[core.TaskExecution_Phase]bool{
	core.TaskExecution_SUCCEEDED: true,
	core.TaskExecution_FAILED:    true,
	core.TaskExecution_ABORTED:   true,
}

func IsExecutionTerminal(phase core.WorkflowExecution_Phase) bool {
	return terminalExecutionPhases[phase]
}

func IsNodeExecutionTerminal(phase core.NodeExecution_Phase) bool {
	return terminalNodeExecutionPhases[phase]
}

func IsTaskExecutionTerminal(phase core.TaskExecution_Phase) bool {
	return terminalTaskExecutionPhases[phase]
}
