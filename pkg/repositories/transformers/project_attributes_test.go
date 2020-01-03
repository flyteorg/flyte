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

var matchingTaskResourceAttributes = &admin.MatchingAttributes{
	Target: &admin.MatchingAttributes_TaskResourceAttributes{
		TaskResourceAttributes: &admin.TaskResourceAttributes{
			Defaults: &admin.TaskResourceSpec{
				Cpu: "1",
			},
		},
	},
}

var projectAttributes = admin.ProjectAttributes{
	Project:            "project",
	MatchingAttributes: matchingTaskResourceAttributes,
}

var marshalledAttributes, _ = proto.Marshal(matchingTaskResourceAttributes)

func TestToProjectAttributesModel(t *testing.T) {
	model, err := ToProjectAttributesModel(projectAttributes, admin.MatchableResource_TASK_RESOURCE)
	assert.Nil(t, err)
	assert.EqualValues(t, models.ProjectAttributes{
		Project:    "project",
		Resource:   admin.MatchableResource_TASK_RESOURCE.String(),
		Attributes: marshalledAttributes,
	}, model)
}

func TestFromProjectAttributesModel(t *testing.T) {
	model := models.ProjectAttributes{
		Project:    "project",
		Resource:   admin.MatchableResource_TASK_RESOURCE.String(),
		Attributes: marshalledAttributes,
	}
	unmarshalledAttributes, err := FromProjectAttributesModel(model)
	assert.Nil(t, err)
	assert.True(t, proto.Equal(&projectAttributes, &unmarshalledAttributes))
}

func TestFromProjectAttributesModel_InvalidResourceAttributes(t *testing.T) {
	model := models.ProjectAttributes{
		Project:    "project",
		Resource:   admin.MatchableResource_TASK_RESOURCE.String(),
		Attributes: []byte("i'm invalid!"),
	}
	_, err := FromProjectAttributesModel(model)
	assert.NotNil(t, err)
	assert.Equal(t, codes.Internal, err.(errors.FlyteAdminError).Code())
}
