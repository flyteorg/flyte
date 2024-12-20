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
	assert.Equal(t, teIdentifier.TaskId.GetResourceType(), teIdentifierCopy.TaskId.GetResourceType())
	assert.Equal(t, teIdentifier.TaskId.GetProject(), teIdentifierCopy.TaskId.GetProject())
	assert.Equal(t, teIdentifier.TaskId.GetDomain(), teIdentifierCopy.TaskId.GetDomain())
	assert.Equal(t, teIdentifier.TaskId.GetName(), teIdentifierCopy.TaskId.GetName())
	assert.Equal(t, teIdentifier.TaskId.GetVersion(), teIdentifierCopy.TaskId.GetVersion())
	assert.Equal(t, teIdentifier.TaskId.GetOrg(), teIdentifierCopy.TaskId.GetOrg())
	assert.Equal(t, teIdentifier.NodeExecutionId.GetExecutionId().GetProject(), teIdentifierCopy.NodeExecutionId.GetExecutionId().GetProject())
	assert.Equal(t, teIdentifier.NodeExecutionId.GetExecutionId().GetDomain(), teIdentifierCopy.NodeExecutionId.GetExecutionId().GetDomain())
	assert.Equal(t, teIdentifier.NodeExecutionId.GetExecutionId().GetName(), teIdentifierCopy.NodeExecutionId.GetExecutionId().GetName())
	assert.Equal(t, teIdentifier.NodeExecutionId.GetExecutionId().GetOrg(), teIdentifierCopy.NodeExecutionId.GetExecutionId().GetOrg())
	assert.Equal(t, teIdentifier.NodeExecutionId.GetNodeId(), teIdentifierCopy.NodeExecutionId.GetNodeId())
	assert.Equal(t, teIdentifier.RetryAttempt, teIdentifierCopy.RetryAttempt)
}
