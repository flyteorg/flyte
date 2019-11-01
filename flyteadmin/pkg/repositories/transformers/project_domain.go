package transformers

import (
	"github.com/golang/protobuf/proto"

	"github.com/lyft/flyteadmin/pkg/errors"
	"github.com/lyft/flyteadmin/pkg/repositories/models"
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/admin"
	"google.golang.org/grpc/codes"
)

func ToProjectDomainModel(attributes admin.ProjectDomainAttributes) (models.ProjectDomain, error) {
	attributeBytes, err := proto.Marshal(&attributes)
	if err != nil {
		return models.ProjectDomain{}, err
	}
	return models.ProjectDomain{
		Project:    attributes.Project,
		Domain:     attributes.Domain,
		Attributes: attributeBytes,
	}, nil
}

func FromProjectDomainModel(model models.ProjectDomain) (admin.ProjectDomainAttributes, error) {
	var attributes admin.ProjectDomainAttributes
	err := proto.Unmarshal(model.Attributes, &attributes)
	if err != nil {
		return admin.ProjectDomainAttributes{}, errors.NewFlyteAdminErrorf(
			codes.Internal, "Failed to decode project domain resource attributes with err: %v", err)
	}
	return attributes, nil
}
