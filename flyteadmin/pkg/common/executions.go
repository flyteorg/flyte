package common

import (
	"fmt"

	"github.com/samber/lo"
	"k8s.io/apimachinery/pkg/util/rand"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
)

const ExecutionIDLength = 20
const ExecutionStringFormat = "a%s"

/* #nosec */
func GetExecutionName(seed int64) string {
	rand.Seed(seed)
	return fmt.Sprintf(ExecutionStringFormat, rand.String(ExecutionIDLength-1))
}

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

// ActiveExecutionPhases is computed by mapping over all workflow execution phases (expressed as int32s and converted to enums)
// and filters out terminal phases.
var ActiveExecutionPhases = lo.Filter[core.WorkflowExecution_Phase](
	lo.Map[int32](
		lo.Keys(core.WorkflowExecution_Phase_name),
		func(item int32, index int) core.WorkflowExecution_Phase {
			return core.WorkflowExecution_Phase(item)
		}),
	func(item core.WorkflowExecution_Phase, index int) bool {
		return !IsExecutionTerminal(item)
	})
