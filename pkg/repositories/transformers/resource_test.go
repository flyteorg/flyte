package transformers

import (
	"testing"

	"github.com/golang/protobuf/proto"

	"github.com/lyft/flyteadmin/pkg/errors"
	"github.com/lyft/flyteadmin/pkg/repositories/models"
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/admin"
	"google.golang.org/grpc/codes"

	"github.com/stretchr/testify/assert"
)

var matchingClusterResourceAttributes = &admin.MatchingAttributes{
	Target: &admin.MatchingAttributes_ClusterResourceAttributes{
		ClusterResourceAttributes: &admin.ClusterResourceAttributes{
			Attributes: map[string]string{
				"foo": "bar",
			},
		},
	},
}

var workflowAttributes = admin.WorkflowAttributes{
	Project:            "project",
	Domain:             "domain",
	Workflow:           "workflow",
	MatchingAttributes: matchingClusterResourceAttributes,
}

var marshalledClusterResourceAttributes, _ = proto.Marshal(matchingClusterResourceAttributes)

var matchingExecutionQueueAttributes = &admin.MatchingAttributes{
	Target: &admin.MatchingAttributes_ExecutionQueueAttributes{
		ExecutionQueueAttributes: &admin.ExecutionQueueAttributes{
			Tags: []string{
				"foo",
			},
		},
	},
}

var projectDomainAttributes = admin.ProjectDomainAttributes{
	Project:            "project",
	Domain:             "domain",
	MatchingAttributes: matchingExecutionQueueAttributes,
}

var marshalledExecutionQueueAttributes, _ = proto.Marshal(matchingExecutionQueueAttributes)

func TestToProjectDomainAttributesModel(t *testing.T) {

	model, err := ProjectDomainAttributesToResourceModel(projectDomainAttributes, admin.MatchableResource_EXECUTION_QUEUE)
	assert.Nil(t, err)
	assert.EqualValues(t, models.Resource{
		Project:      "project",
		Domain:       "domain",
		ResourceType: admin.MatchableResource_EXECUTION_QUEUE.String(),
		Priority:     models.ResourcePriorityProjectDomainLevel,
		Attributes:   marshalledExecutionQueueAttributes,
	}, model)
}

func TestFromProjectDomainAttributesModel(t *testing.T) {
	model := models.Resource{
		Project:      "project",
		Domain:       "domain",
		ResourceType: admin.MatchableResource_EXECUTION_QUEUE.String(),
		Attributes:   marshalledExecutionQueueAttributes,
	}
	unmarshalledAttributes, err := FromResourceModelToProjectDomainAttributes(model)
	assert.Nil(t, err)
	assert.True(t, proto.Equal(&projectDomainAttributes, &unmarshalledAttributes))
}

func TestFromProjectDomainAttributesModel_InvalidResourceAttributes(t *testing.T) {
	model := models.Resource{
		Project:      "project",
		Domain:       "domain",
		ResourceType: admin.MatchableResource_EXECUTION_QUEUE.String(),
		Attributes:   []byte("i'm invalid!"),
	}
	_, err := FromResourceModelToProjectDomainAttributes(model)
	assert.NotNil(t, err)
	assert.Equal(t, codes.Internal, err.(errors.FlyteAdminError).Code())
}

func TestToWorkflowAttributesModel(t *testing.T) {
	model, err := WorkflowAttributesToResourceModel(workflowAttributes, admin.MatchableResource_EXECUTION_QUEUE)
	assert.Nil(t, err)
	assert.EqualValues(t, models.Resource{
		Project:      "project",
		Domain:       "domain",
		Workflow:     "workflow",
		ResourceType: admin.MatchableResource_EXECUTION_QUEUE.String(),
		Priority:     models.ResourcePriorityWorkflowLevel,
		Attributes:   marshalledClusterResourceAttributes,
	}, model)
}

func TestFromWorkflowAttributesModel(t *testing.T) {
	model := models.Resource{
		Project:      "project",
		Domain:       "domain",
		Workflow:     "workflow",
		ResourceType: admin.MatchableResource_EXECUTION_QUEUE.String(),
		Attributes:   marshalledClusterResourceAttributes,
	}
	unmarshalledAttributes, err := FromResourceModelToWorkflowAttributes(model)
	assert.Nil(t, err)
	assert.True(t, proto.Equal(&workflowAttributes, &unmarshalledAttributes))
}

func TestFromWorkflowAttributesModel_InvalidResourceAttributes(t *testing.T) {
	model := models.Resource{
		Project:      "project",
		Domain:       "domain",
		Workflow:     "workflow",
		ResourceType: admin.MatchableResource_EXECUTION_QUEUE.String(),
		Attributes:   []byte("i'm invalid!"),
	}
	_, err := FromResourceModelToWorkflowAttributes(model)
	assert.NotNil(t, err)
	assert.Equal(t, codes.Internal, err.(errors.FlyteAdminError).Code())
}
