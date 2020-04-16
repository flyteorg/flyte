package subworkflow

import (
	"strconv"

	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/core"

	"github.com/lyft/flytepropeller/pkg/utils"
)

const maxLengthForSubWorkflow = 20

func GetChildWorkflowExecutionID(nodeExecID *core.NodeExecutionIdentifier, attempt uint32) (*core.WorkflowExecutionIdentifier, error) {
	name, err := utils.FixedLengthUniqueIDForParts(maxLengthForSubWorkflow, nodeExecID.ExecutionId.Name, nodeExecID.NodeId, strconv.Itoa(int(attempt)))
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
