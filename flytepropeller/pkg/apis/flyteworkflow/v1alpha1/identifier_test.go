package v1alpha1

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
)

func TestIdentifierJSONMarshalling(t *testing.T) {
	identifier := Identifier{
		&core.Identifier{
			ResourceType: core.ResourceType_TASK,
			Project:      "TestProject",
			Domain:       "TestDomain",
			Name:         "TestName",
			Version:      "TestVersion",
		},
	}

	expected, mockErr := mockMarshalPbToBytes(identifier.Identifier)
	assert.Nil(t, mockErr)

	// MarshalJSON
	identifierBytes, mErr := identifier.MarshalJSON()
	assert.Nil(t, mErr)
	assert.Equal(t, expected, identifierBytes)

	// UnmarshalJSON
	identifierObj := &Identifier{}
	uErr := identifierObj.UnmarshalJSON(identifierBytes)
	assert.Nil(t, uErr)
	assert.Equal(t, identifier.Project, identifierObj.Project)
	assert.Equal(t, identifier.Domain, identifierObj.Domain)
	assert.Equal(t, identifier.Name, identifierObj.Name)
	assert.Equal(t, identifier.Version, identifierObj.Version)
}

func TestIdentifier_DeepCopyInto(t *testing.T) {
	identifier := Identifier{
		&core.Identifier{
			ResourceType: core.ResourceType_TASK,
			Project:      "TestProject",
			Domain:       "TestDomain",
			Name:         "TestName",
			Version:      "TestVersion",
		},
	}

	identifierCopy := Identifier{}
	identifier.DeepCopyInto(&identifierCopy)
	assert.Equal(t, identifier.Project, identifierCopy.Project)
	assert.Equal(t, identifier.Domain, identifierCopy.Domain)
	assert.Equal(t, identifier.Name, identifierCopy.Name)
	assert.Equal(t, identifier.Version, identifierCopy.Version)
}

func TestWorkflowExecutionIdentifier_DeepCopyInto(t *testing.T) {
	weIdentifier := WorkflowExecutionIdentifier{
		&core.WorkflowExecutionIdentifier{
			Project: "TestProject",
			Domain:  "TestDomain",
			Name:    "TestName",
			Org:     "TestOrg",
		},
	}

	weIdentifierCopy := WorkflowExecutionIdentifier{}
	weIdentifier.DeepCopyInto(&weIdentifierCopy)
	assert.Equal(t, weIdentifier.Project, weIdentifierCopy.Project)
	assert.Equal(t, weIdentifier.Domain, weIdentifierCopy.Domain)
	assert.Equal(t, weIdentifier.Name, weIdentifierCopy.Name)
	assert.Equal(t, weIdentifier.Org, weIdentifierCopy.Org)
}

func TestTaskExecutionIdentifier_DeepCopyInto(t *testing.T) {
	teIdentifier := TaskExecutionIdentifier{
		&core.TaskExecutionIdentifier{
			TaskId: &core.Identifier{
				ResourceType: core.ResourceType_TASK,
				Project:      "TestProject",
				Domain:       "TestDomain",
				Name:         "TestName",
				Version:      "TestVersion",
				Org:          "TestOrg",
			},
			NodeExecutionId: &core.NodeExecutionIdentifier{
				ExecutionId: &core.WorkflowExecutionIdentifier{
					Project: "TestProject",
					Domain:  "TestDomain",
					Name:    "TestName",
					Org:     "TestOrg",
				},
				NodeId: "TestNodeId",
			},
			RetryAttempt: 1,
		},
	}

	teIdentifierCopy := TaskExecutionIdentifier{}
	teIdentifier.DeepCopyInto(&teIdentifierCopy)
	assert.Equal(t, teIdentifier.TaskId.ResourceType, teIdentifierCopy.TaskId.ResourceType)
	assert.Equal(t, teIdentifier.TaskId.Project, teIdentifierCopy.TaskId.Project)
	assert.Equal(t, teIdentifier.TaskId.Domain, teIdentifierCopy.TaskId.Domain)
	assert.Equal(t, teIdentifier.TaskId.Name, teIdentifierCopy.TaskId.Name)
	assert.Equal(t, teIdentifier.TaskId.Version, teIdentifierCopy.TaskId.Version)
	assert.Equal(t, teIdentifier.TaskId.Org, teIdentifierCopy.TaskId.Org)
	assert.Equal(t, teIdentifier.NodeExecutionId.ExecutionId.Project, teIdentifierCopy.NodeExecutionId.ExecutionId.Project)
	assert.Equal(t, teIdentifier.NodeExecutionId.ExecutionId.Domain, teIdentifierCopy.NodeExecutionId.ExecutionId.Domain)
	assert.Equal(t, teIdentifier.NodeExecutionId.ExecutionId.Name, teIdentifierCopy.NodeExecutionId.ExecutionId.Name)
	assert.Equal(t, teIdentifier.NodeExecutionId.ExecutionId.Org, teIdentifierCopy.NodeExecutionId.ExecutionId.Org)
	assert.Equal(t, teIdentifier.NodeExecutionId.NodeId, teIdentifierCopy.NodeExecutionId.NodeId)
	assert.Equal(t, teIdentifier.RetryAttempt, teIdentifierCopy.RetryAttempt)
}
