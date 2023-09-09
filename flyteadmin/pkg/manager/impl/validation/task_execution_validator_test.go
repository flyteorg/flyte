package validation

import (
	"testing"
	"time"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/event"
	"github.com/golang/protobuf/ptypes"
	"github.com/stretchr/testify/assert"
)

var taskEventOccurredAt = time.Now()
var taskEventOccurredAtProto, _ = ptypes.TimestampProto(taskEventOccurredAt)
var maxOutputSizeInBytes = int64(1000000)

func TestValidateTaskExecutionRequest(t *testing.T) {
	assert.Nil(t, ValidateTaskExecutionRequest(admin.TaskExecutionEventRequest{
		Event: &event.TaskExecutionEvent{
			OccurredAt: taskEventOccurredAtProto,
			TaskId: &core.Identifier{
				ResourceType: core.ResourceType_TASK,
				Project:      "project",
				Domain:       "domain",
				Name:         "name",
				Version:      "version",
			},
			ParentNodeExecutionId: &core.NodeExecutionIdentifier{
				NodeId: "nodey",
				ExecutionId: &core.WorkflowExecutionIdentifier{
					Project: "project",
					Domain:  "domain",
					Name:    "name",
				},
			},
			RetryAttempt: 0,
		},
	}, maxOutputSizeInBytes))
}

func TestValidateTaskExecutionRequest_MissingFields(t *testing.T) {
	err := ValidateTaskExecutionRequest(admin.TaskExecutionEventRequest{
		Event: &event.TaskExecutionEvent{
			TaskId: &core.Identifier{
				ResourceType: core.ResourceType_TASK,
				Project:      "project",
				Domain:       "domain",
				Name:         "name",
				Version:      "version",
			},
			ParentNodeExecutionId: &core.NodeExecutionIdentifier{
				NodeId: "nodey",
				ExecutionId: &core.WorkflowExecutionIdentifier{
					Project: "project",
					Domain:  "domain",
					Name:    "name",
				},
			},
			RetryAttempt: 0,
		},
	}, maxOutputSizeInBytes)
	assert.EqualError(t, err, "missing occurred_at")

	err = ValidateTaskExecutionRequest(admin.TaskExecutionEventRequest{
		Event: &event.TaskExecutionEvent{
			OccurredAt: taskEventOccurredAtProto,
			TaskId: &core.Identifier{
				ResourceType: core.ResourceType_TASK,
				Project:      "project",
				Domain:       "domain",
				Name:         "name",
			},
			ParentNodeExecutionId: &core.NodeExecutionIdentifier{
				NodeId: "nodey",
				ExecutionId: &core.WorkflowExecutionIdentifier{
					Project: "project",
					Domain:  "domain",
					Name:    "name",
				},
			},
			RetryAttempt: 0,
		},
	}, maxOutputSizeInBytes)
	assert.EqualError(t, err, "missing version")

	err = ValidateTaskExecutionRequest(admin.TaskExecutionEventRequest{
		Event: &event.TaskExecutionEvent{
			OccurredAt: taskEventOccurredAtProto,
			TaskId: &core.Identifier{
				ResourceType: core.ResourceType_TASK,
				Project:      "project",
				Domain:       "domain",
				Name:         "name",
				Version:      "version",
			},
			ParentNodeExecutionId: &core.NodeExecutionIdentifier{
				ExecutionId: &core.WorkflowExecutionIdentifier{
					Project: "project",
					Domain:  "domain",
					Name:    "name",
				},
			},
			RetryAttempt: 0,
		},
	}, maxOutputSizeInBytes)
	assert.EqualError(t, err, "missing node_id")

	err = ValidateTaskExecutionRequest(admin.TaskExecutionEventRequest{}, maxOutputSizeInBytes)
	assert.EqualError(t, err, "missing event")
}

func TestValidateTaskExecutionIdentifier(t *testing.T) {
	assert.Nil(t, ValidateTaskExecutionIdentifier(&core.TaskExecutionIdentifier{
		TaskId: &core.Identifier{
			ResourceType: core.ResourceType_TASK,
			Project:      "project",
			Domain:       "domain",
			Name:         "name",
			Version:      "version",
		},
		NodeExecutionId: &core.NodeExecutionIdentifier{
			NodeId: "nodey",
			ExecutionId: &core.WorkflowExecutionIdentifier{
				Project: "project",
				Domain:  "domain",
				Name:    "name",
			},
		},
		RetryAttempt: 0,
	}))
}

func TestValidateTaskExecutionListRequest(t *testing.T) {
	assert.Nil(t, ValidateTaskExecutionListRequest(admin.TaskExecutionListRequest{
		NodeExecutionId: &core.NodeExecutionIdentifier{
			NodeId: "nodey",
			ExecutionId: &core.WorkflowExecutionIdentifier{
				Project: "project",
				Domain:  "domain",
				Name:    "name",
			},
		},
		Limit: 200,
	}))
}

func TestValidateTaskExecutionListRequest_MissingFields(t *testing.T) {
	err := ValidateTaskExecutionListRequest(admin.TaskExecutionListRequest{
		NodeExecutionId: &core.NodeExecutionIdentifier{
			NodeId: "nodey",
			ExecutionId: &core.WorkflowExecutionIdentifier{
				Project: "project",
				Name:    "name",
			},
		},
		Limit: 200,
	})
	assert.EqualError(t, err, "missing domain")

	err = ValidateTaskExecutionListRequest(admin.TaskExecutionListRequest{
		NodeExecutionId: &core.NodeExecutionIdentifier{
			NodeId: "nodey",
			ExecutionId: &core.WorkflowExecutionIdentifier{
				Project: "project",
				Domain:  "domain",
				Name:    "name",
			},
		},
	})
	assert.EqualError(t, err, "invalid value for limit")
}
