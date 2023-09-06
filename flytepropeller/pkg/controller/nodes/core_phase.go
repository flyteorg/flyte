package nodes

import "github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"

// IsTerminalNodePhase returns true if node phase is one of the terminal phases, else false
func IsTerminalNodePhase(p core.NodeExecution_Phase) bool {
	return p == core.NodeExecution_ABORTED || p == core.NodeExecution_FAILED || p == core.NodeExecution_SKIPPED || p == core.NodeExecution_SUCCEEDED || p == core.NodeExecution_TIMED_OUT
}

// IsTerminalTaskPhase returns true if task phase is terminal, else false
func IsTerminalTaskPhase(p core.TaskExecution_Phase) bool {
	return p == core.TaskExecution_ABORTED || p == core.TaskExecution_FAILED || p == core.TaskExecution_SUCCEEDED
}
