package transformers

import (
	"github.com/golang/protobuf/proto"

	"github.com/lyft/flyteadmin/pkg/errors"
	"github.com/lyft/flyteadmin/pkg/repositories/models"
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/admin"
	"google.golang.org/grpc/codes"
)

func ToProjectDomainAttributesModel(attributes admin.ProjectDomainAttributes, resource admin.MatchableResource) (models.ProjectDomainAttributes, error) {
	attributeBytes, err := proto.Marshal(attributes.MatchingAttributes)
	if err != nil {
		return models.ProjectDomainAttributes{}, err
	}
	return models.ProjectDomainAttributes{
		Project:    attributes.Project,
		Domain:     attributes.Domain,
		Resource:   resource.String(),
		Attributes: attributeBytes,
	}, nil
}

func FromProjectDomainAttributesModel(model models.ProjectDomainAttributes) (admin.ProjectDomainAttributes, error) {
	var attributes admin.MatchingAttributes
	err := proto.Unmarshal(model.Attributes, &attributes)
	if err != nil {
		return admin.ProjectDomainAttributes{}, errors.NewFlyteAdminErrorf(
			codes.Internal, "Failed to decode project domain resource projectDomainAttributes with err: %v", err)
	}
	return admin.ProjectDomainAttributes{
		Project:            model.Project,
		Domain:             model.Domain,
		MatchingAttributes: &attributes,
	}, nil
}
