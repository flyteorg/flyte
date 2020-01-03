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

func TestToWorkflowAttributesModel(t *testing.T) {
	model, err := ToWorkflowAttributesModel(workflowAttributes, admin.MatchableResource_EXECUTION_QUEUE)
	assert.Nil(t, err)
	assert.EqualValues(t, models.WorkflowAttributes{
		Project:    "project",
		Domain:     "domain",
		Workflow:   "workflow",
		Resource:   admin.MatchableResource_EXECUTION_QUEUE.String(),
		Attributes: marshalledClusterResourceAttributes,
	}, model)
}

func TestFromWorkflowAttributesModel(t *testing.T) {
	model := models.WorkflowAttributes{
		Project:    "project",
		Domain:     "domain",
		Workflow:   "workflow",
		Resource:   admin.MatchableResource_EXECUTION_QUEUE.String(),
		Attributes: marshalledClusterResourceAttributes,
	}
	unmarshalledAttributes, err := FromWorkflowAttributesModel(model)
	assert.Nil(t, err)
	assert.True(t, proto.Equal(&workflowAttributes, &unmarshalledAttributes))
}

func TestFromWorkflowAttributesModel_InvalidResourceAttributes(t *testing.T) {
	model := models.WorkflowAttributes{
		Project:    "project",
		Domain:     "domain",
		Workflow:   "workflow",
		Resource:   admin.MatchableResource_EXECUTION_QUEUE.String(),
		Attributes: []byte("i'm invalid!"),
	}
	_, err := FromWorkflowAttributesModel(model)
	assert.NotNil(t, err)
	assert.Equal(t, codes.Internal, err.(errors.FlyteAdminError).Code())
}
