package validation

import (
	"testing"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/stretchr/testify/assert"
)

var testExecutionID = core.WorkflowExecutionIdentifier{
	Project: "project",
	Domain:  "domain",
	Name:    "name",
}

func TestValidateNodeExecutionIdentifier(t *testing.T) {
	err := ValidateNodeExecutionIdentifier(&core.NodeExecutionIdentifier{
		ExecutionId: &testExecutionID,
		NodeId:      "node id",
	})
	assert.Nil(t, err)
}

func TestValidateNodeExecutionIdentifier_MissingFields(t *testing.T) {
	err := ValidateNodeExecutionIdentifier(&core.NodeExecutionIdentifier{
		NodeId: "node id",
	})
	assert.EqualError(t, err, "missing execution_id")

	err = ValidateNodeExecutionIdentifier(&core.NodeExecutionIdentifier{
		ExecutionId: &testExecutionID,
	})
	assert.EqualError(t, err, "missing node_id")
}

func TestValidateNodeExecutionListRequest(t *testing.T) {
	err := ValidateNodeExecutionListRequest(admin.NodeExecutionListRequest{
		WorkflowExecutionId: &testExecutionID,
		Filters:             "foo",
		Limit:               2,
	})
	assert.Nil(t, err)
}

func TestValidateNodeExecutionListRequest_MissingFields(t *testing.T) {
	err := ValidateNodeExecutionListRequest(admin.NodeExecutionListRequest{
		Limit: 2,
	})
	assert.EqualError(t, err, "missing execution_id")

	err = ValidateNodeExecutionListRequest(admin.NodeExecutionListRequest{
		WorkflowExecutionId: &testExecutionID,
	})
	assert.EqualError(t, err, "invalid value for limit")
}

func TestValidateNodeExecutionForTaskListRequest(t *testing.T) {
	err := ValidateNodeExecutionForTaskListRequest(admin.NodeExecutionForTaskListRequest{
		TaskExecutionId: &core.TaskExecutionIdentifier{
			NodeExecutionId: &core.NodeExecutionIdentifier{
				ExecutionId: &testExecutionID,
				NodeId:      "nodey",
			},
			TaskId: &core.Identifier{
				ResourceType: core.ResourceType_TASK,
				Project:      "project",
				Domain:       "domain",
				Name:         "name",
				Version:      "version",
			},
		},
		Filters: "foo",
		Limit:   2,
	})
	assert.Nil(t, err)
}

func TestValidateNodeExecutionForTaskListRequest_MissingFields(t *testing.T) {
	err := ValidateNodeExecutionForTaskListRequest(admin.NodeExecutionForTaskListRequest{
		TaskExecutionId: &core.TaskExecutionIdentifier{
			NodeExecutionId: &core.NodeExecutionIdentifier{
				ExecutionId: &testExecutionID,
				NodeId:      "nodey",
			},
			TaskId: &core.Identifier{
				ResourceType: core.ResourceType_TASK,
				Project:      "project",
				Domain:       "domain",
				Name:         "name",
				Version:      "version",
			},
		},
	})
	assert.EqualError(t, err, "invalid value for limit")

	err = ValidateNodeExecutionForTaskListRequest(admin.NodeExecutionForTaskListRequest{
		TaskExecutionId: &core.TaskExecutionIdentifier{
			NodeExecutionId: &core.NodeExecutionIdentifier{
				ExecutionId: &testExecutionID,
			},
			TaskId: &core.Identifier{
				ResourceType: core.ResourceType_TASK,
				Project:      "project",
				Domain:       "domain",
				Name:         "name",
				Version:      "version",
			},
		},
		Limit: 2,
	})
	assert.EqualError(t, err, "missing node_id")

	err = ValidateNodeExecutionForTaskListRequest(admin.NodeExecutionForTaskListRequest{
		TaskExecutionId: &core.TaskExecutionIdentifier{
			NodeExecutionId: &core.NodeExecutionIdentifier{
				ExecutionId: &testExecutionID,
				NodeId:      "nodey",
			},
			TaskId: &core.Identifier{
				ResourceType: core.ResourceType_TASK,
				Domain:       "domain",
				Name:         "name",
				Version:      "version",
			},
		},
		Limit: 2,
	})
	assert.EqualError(t, err, "missing project")
}
