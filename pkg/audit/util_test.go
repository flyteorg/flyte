package audit

import (
	"testing"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/stretchr/testify/assert"
)

func TestParametersFromIdentifier(t *testing.T) {
	assert.EqualValues(t, map[string]string{
		"project": "proj",
		"domain":  "development",
		"name":    "foo",
		"version": "123",
	}, ParametersFromIdentifier(&core.Identifier{
		Project: "proj",
		Domain:  "development",
		Name:    "foo",
		Version: "123",
	}))
}

func TestParametersFromNamedEntityIdentifier(t *testing.T) {
	assert.EqualValues(t, map[string]string{
		"project": "proj",
		"domain":  "development",
		"name":    "foo",
	}, ParametersFromNamedEntityIdentifier(&admin.NamedEntityIdentifier{
		Project: "proj",
		Domain:  "development",
		Name:    "foo",
	}))
}

func TestParametersFromNamedEntityIdentifierAndResource(t *testing.T) {
	assert.EqualValues(t, map[string]string{
		"project":  "proj",
		"domain":   "development",
		"name":     "foo",
		"resource": "LAUNCH_PLAN",
	}, ParametersFromNamedEntityIdentifierAndResource(&admin.NamedEntityIdentifier{
		Project: "proj",
		Domain:  "development",
		Name:    "foo",
	}, core.ResourceType_LAUNCH_PLAN))
}

func TestParametersFromExecutionIdentifier(t *testing.T) {
	assert.EqualValues(t, map[string]string{
		"project": "proj",
		"domain":  "development",
		"name":    "foo",
	}, ParametersFromExecutionIdentifier(&core.WorkflowExecutionIdentifier{
		Project: "proj",
		Domain:  "development",
		Name:    "foo",
	}))
}

func TestParametersFromNodeExecutionIdentifier(t *testing.T) {
	assert.EqualValues(t, map[string]string{
		"project": "proj",
		"domain":  "development",
		"name":    "foo",
		"node_id": "nodey",
	}, ParametersFromNodeExecutionIdentifier(&core.NodeExecutionIdentifier{
		NodeId: "nodey",
		ExecutionId: &core.WorkflowExecutionIdentifier{
			Project: "proj",
			Domain:  "development",
			Name:    "foo",
		},
	}))
}

func TestParametersFromTaskExecutionIdentifier(t *testing.T) {
	assert.EqualValues(t, map[string]string{
		"project":       "proj",
		"domain":        "development",
		"name":          "foo",
		"node_id":       "nodey",
		"retry_attempt": "1",
		"task_project":  "proj2",
		"task_domain":   "production",
		"task_name":     "bar",
		"task_version":  "version",
	}, ParametersFromTaskExecutionIdentifier(&core.TaskExecutionIdentifier{
		TaskId: &core.Identifier{
			Project: "proj2",
			Domain:  "production",
			Name:    "bar",
			Version: "version",
		},
		NodeExecutionId: &core.NodeExecutionIdentifier{
			NodeId: "nodey",
			ExecutionId: &core.WorkflowExecutionIdentifier{
				Project: "proj",
				Domain:  "development",
				Name:    "foo",
			},
		},
		RetryAttempt: 1,
	}))
}
