package common

import (
	"math/rand"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
)

const ExecutionIDLength = 10

// In kubernetes, resource names must comply with this regex: '[a-z]([-a-z0-9]*[a-z0-9])?'
const AllowedExecutionIDStartCharStr = "abcdefghijklmnopqrstuvwxyz"
const AllowedExecutionIDStr = "abcdefghijklmnopqrstuvwxyz1234567890"

var AllowedExecutionIDStartChars = []rune(AllowedExecutionIDStartCharStr)
var AllowedExecutionIDChars = []rune(AllowedExecutionIDStr)

/* #nosec */
func GetExecutionName(seed int64) string {
	executionName := make([]rune, ExecutionIDLength)
	rand.Seed(seed)
	executionName[0] = AllowedExecutionIDStartChars[rand.Intn(len(AllowedExecutionIDStartChars))]
	for i := 1; i < len(executionName); i++ {
		executionName[i] = AllowedExecutionIDChars[rand.Intn(len(AllowedExecutionIDChars))]
	}
	return string(executionName)
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
