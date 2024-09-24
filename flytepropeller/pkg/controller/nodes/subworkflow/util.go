package subworkflow

import (
	"strconv"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/encoding"
	"github.com/flyteorg/flyte/flytestdlib/utils"
)

func GetChildWorkflowExecutionID(nodeExecID *core.NodeExecutionIdentifier, attempt uint32) (*core.WorkflowExecutionIdentifier, error) {
	name, err := encoding.FixedLengthUniqueIDForParts(utils.MaxUniqueIDLength, []string{nodeExecID.ExecutionId.Name, nodeExecID.NodeId, strconv.Itoa(int(attempt))})
	if err != nil {
		return nil, err
	}

	// Restriction on name is 20 chars
	return &core.WorkflowExecutionIdentifier{
		Project: nodeExecID.ExecutionId.Project,
		Domain:  nodeExecID.ExecutionId.Domain,
		Name:    name,
	}, nil
}

func GetChildWorkflowExecutionIDV2(nodeExecID *core.NodeExecutionIdentifier, attempt uint32) (*core.WorkflowExecutionIdentifier, error) {
	name, err := encoding.FixedLengthUniqueIDForParts(utils.MaxUniqueIDLength, []string{nodeExecID.ExecutionId.Name, nodeExecID.NodeId, strconv.Itoa(int(attempt))},
		encoding.NewAlgorithmOption(encoding.Algorithm64))
	if err != nil {
		return nil, err
	}

	// Restriction on name is 20 chars
	return &core.WorkflowExecutionIdentifier{
		Project: nodeExecID.ExecutionId.Project,
		Domain:  nodeExecID.ExecutionId.Domain,
		Name:    name,
	}, nil
}
