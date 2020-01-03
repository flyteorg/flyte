package transformers

import (
	"github.com/golang/protobuf/proto"

	"github.com/lyft/flyteadmin/pkg/errors"
	"github.com/lyft/flyteadmin/pkg/repositories/models"
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/admin"
	"google.golang.org/grpc/codes"
)

func ToProjectAttributesModel(attributes admin.ProjectAttributes, resource admin.MatchableResource) (models.ProjectAttributes, error) {
	attributeBytes, err := proto.Marshal(attributes.MatchingAttributes)
	if err != nil {
		return models.ProjectAttributes{}, err
	}
	return models.ProjectAttributes{
		Project:    attributes.Project,
		Resource:   resource.String(),
		Attributes: attributeBytes,
	}, nil
}

func FromProjectAttributesModel(model models.ProjectAttributes) (admin.ProjectAttributes, error) {
	var attributes admin.MatchingAttributes
	err := proto.Unmarshal(model.Attributes, &attributes)
	if err != nil {
		return admin.ProjectAttributes{}, errors.NewFlyteAdminErrorf(
			codes.Internal, "Failed to decode project domain resource projectDomainAttributes with err: %v", err)
	}
	return admin.ProjectAttributes{
		Project:            model.Project,
		MatchingAttributes: &attributes,
	}, nil
}
