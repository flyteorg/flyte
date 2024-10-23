package v1alpha1

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
)

func TestWorkflowNodeSpec_GetLaunchPlanRefID(t *testing.T) {
	wfNodeSpec := &WorkflowNodeSpec{
		LaunchPlanRefID: &LaunchPlanRefID{
			&core.Identifier{
				Project: "TestProject",
			},
		},
	}

	nilWfNodeSpec := &WorkflowNodeSpec{}

	assert.Equal(t, wfNodeSpec.GetLaunchPlanRefID(), wfNodeSpec.LaunchPlanRefID)
	assert.Empty(t, nilWfNodeSpec.GetLaunchPlanRefID())
}

func TestWorkflowNodeSpec_GetSubWorkflowRef(t *testing.T) {
	workflowID := "TestWorkflowID"
	wfNodeSpec := &WorkflowNodeSpec{
		SubWorkflowReference: &workflowID,
	}

	nilWfNodeSpec := &WorkflowNodeSpec{}

	assert.Equal(t, wfNodeSpec.GetSubWorkflowRef(), wfNodeSpec.SubWorkflowReference)
	assert.Empty(t, nilWfNodeSpec.GetSubWorkflowRef())
}
