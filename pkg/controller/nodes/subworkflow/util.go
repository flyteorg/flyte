package subworkflow

import (
	"strconv"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"

	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/encoding"
)

const maxLengthForSubWorkflow = 20

func GetChildWorkflowExecutionID(nodeExecID *core.NodeExecutionIdentifier, attempt uint32) (*core.WorkflowExecutionIdentifier, error) {
	name, err := encoding.FixedLengthUniqueIDForParts(maxLengthForSubWorkflow, []string{nodeExecID.ExecutionId.Name, nodeExecID.NodeId, strconv.Itoa(int(attempt))})
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
	name, err := encoding.FixedLengthUniqueIDForParts(maxLengthForSubWorkflow, []string{nodeExecID.ExecutionId.Name, nodeExecID.NodeId, strconv.Itoa(int(attempt))},
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
