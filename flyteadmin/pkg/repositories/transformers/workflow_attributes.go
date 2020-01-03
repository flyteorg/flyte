package transformers

import (
	"github.com/golang/protobuf/proto"

	"github.com/lyft/flyteadmin/pkg/errors"
	"github.com/lyft/flyteadmin/pkg/repositories/models"
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/admin"
	"google.golang.org/grpc/codes"
)

func ToWorkflowAttributesModel(attributes admin.WorkflowAttributes, resource admin.MatchableResource) (models.WorkflowAttributes, error) {
	attributeBytes, err := proto.Marshal(attributes.MatchingAttributes)
	if err != nil {
		return models.WorkflowAttributes{}, err
	}
	return models.WorkflowAttributes{
		Project:    attributes.Project,
		Domain:     attributes.Domain,
		Workflow:   attributes.Workflow,
		Resource:   resource.String(),
		Attributes: attributeBytes,
	}, nil
}

func FromWorkflowAttributesModel(model models.WorkflowAttributes) (admin.WorkflowAttributes, error) {
	var attributes admin.MatchingAttributes
	err := proto.Unmarshal(model.Attributes, &attributes)
	if err != nil {
		return admin.WorkflowAttributes{}, errors.NewFlyteAdminErrorf(
			codes.Internal, "Failed to decode project domain resource projectDomainAttributes with err: %v", err)
	}
	return admin.WorkflowAttributes{
		Project:            model.Project,
		Domain:             model.Domain,
		Workflow:           model.Workflow,
		MatchingAttributes: &attributes,
	}, nil
}
