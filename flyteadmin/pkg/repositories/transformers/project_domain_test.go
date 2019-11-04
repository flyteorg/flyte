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

var attributes = admin.ProjectDomainAttributes{
	Project: "project",
	Domain:  "domain",
	Attributes: map[string]string{
		"cpu": "100",
	},
}

var marshalledAttributes, _ = proto.Marshal(&attributes)

func TestToProjectDomainModel(t *testing.T) {

	model, err := ToProjectDomainModel(attributes)
	assert.Nil(t, err)
	assert.EqualValues(t, models.ProjectDomain{
		Project:    "project",
		Domain:     "domain",
		Attributes: marshalledAttributes,
	}, model)
}

func TestFromProjectDomainModel(t *testing.T) {
	model := models.ProjectDomain{
		Project:    "project",
		Domain:     "domain",
		Attributes: marshalledAttributes,
	}
	unmarshalledAttributes, err := FromProjectDomainModel(model)
	assert.Nil(t, err)
	assert.True(t, proto.Equal(&attributes, &unmarshalledAttributes))
}

func TestFromProjectDomainModel_InvalidResourceAttributes(t *testing.T) {
	model := models.ProjectDomain{
		Project:    "project",
		Domain:     "domain",
		Attributes: []byte("i'm invalid!"),
	}
	_, err := FromProjectDomainModel(model)
	assert.NotNil(t, err)
	assert.Equal(t, codes.Internal, err.(errors.FlyteAdminError).Code())
}
