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

	model, err := ToProjectDomainAttributesModel(projectDomainAttributes, admin.MatchableResource_EXECUTION_QUEUE)
	assert.Nil(t, err)
	assert.EqualValues(t, models.ProjectDomainAttributes{
		Project:    "project",
		Domain:     "domain",
		Resource:   admin.MatchableResource_EXECUTION_QUEUE.String(),
		Attributes: marshalledExecutionQueueAttributes,
	}, model)
}

func TestFromProjectDomainAttributesModel(t *testing.T) {
	model := models.ProjectDomainAttributes{
		Project:    "project",
		Domain:     "domain",
		Resource:   admin.MatchableResource_EXECUTION_QUEUE.String(),
		Attributes: marshalledExecutionQueueAttributes,
	}
	unmarshalledAttributes, err := FromProjectDomainAttributesModel(model)
	assert.Nil(t, err)
	assert.True(t, proto.Equal(&projectDomainAttributes, &unmarshalledAttributes))
}

func TestFromProjectDomainAttributesModel_InvalidResourceAttributes(t *testing.T) {
	model := models.ProjectDomainAttributes{
		Project:    "project",
		Domain:     "domain",
		Resource:   admin.MatchableResource_EXECUTION_QUEUE.String(),
		Attributes: []byte("i'm invalid!"),
	}
	_, err := FromProjectDomainAttributesModel(model)
	assert.NotNil(t, err)
	assert.Equal(t, codes.Internal, err.(errors.FlyteAdminError).Code())
}
