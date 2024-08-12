package common

import (
	"context"
	"fmt"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flytestdlib/logger"
	"github.com/wolfeidau/humanhash"
	"k8s.io/apimachinery/pkg/util/rand"
)

const ExecutionIDLength = 20
const ExecutionIDLengthLimit = 63
const ExecutionStringFormat = "a%s"

/* #nosec */
func GetExecutionName(seed int64, enableHumanHash bool) (string, error) {
	rand.Seed(seed)
	if enableHumanHash {
		hashKey := []byte(rand.String(ExecutionIDLength))
		result, err := humanhash.Humanize(hashKey, 3)
		if err != nil {
			logger.Errorf(context.Background(), "failed to generate execution name using key %v: %v", hashKey, err)
			return "", err
		}
		return result, nil
	}
	return fmt.Sprintf(ExecutionStringFormat, rand.String(ExecutionIDLength-1)), nil
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
