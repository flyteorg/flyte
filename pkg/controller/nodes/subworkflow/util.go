package subworkflow

import (
	"strconv"

	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/lyft/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
	"github.com/lyft/flytepropeller/pkg/utils"
)

const maxLengthForSubWorkflow = 20

func GetChildWorkflowExecutionID(parentID *core.WorkflowExecutionIdentifier, id v1alpha1.NodeID, attempt uint32) (*core.WorkflowExecutionIdentifier, error) {
	name, err := utils.FixedLengthUniqueIDForParts(maxLengthForSubWorkflow, parentID.Name, id, strconv.Itoa(int(attempt)))
	if err != nil {
		return nil, err
	}
	// Restriction on name is 20 chars
	return &core.WorkflowExecutionIdentifier{
		Project: parentID.Project,
		Domain:  parentID.Domain,
		Name:    name,
	}, nil
}
