package transformers

import (
	"github.com/golang/protobuf/proto"

	"github.com/lyft/flyteadmin/pkg/errors"
	"github.com/lyft/flyteadmin/pkg/repositories/models"
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/admin"
	"google.golang.org/grpc/codes"
)

func WorkflowAttributesToResourceModel(attributes admin.WorkflowAttributes, resource admin.MatchableResource) (models.Resource, error) {
	attributeBytes, err := proto.Marshal(attributes.MatchingAttributes)
	if err != nil {
		return models.Resource{}, err
	}
	return models.Resource{
		Project:      attributes.Project,
		Domain:       attributes.Domain,
		Workflow:     attributes.Workflow,
		ResourceType: resource.String(),
		Priority:     models.ResourcePriorityWorkflowLevel,
		Attributes:   attributeBytes,
	}, nil
}

func FromResourceModelToWorkflowAttributes(model models.Resource) (admin.WorkflowAttributes, error) {
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

func ProjectDomainAttributesToResourceModel(attributes admin.ProjectDomainAttributes, resource admin.MatchableResource) (models.Resource, error) {
	attributeBytes, err := proto.Marshal(attributes.MatchingAttributes)
	if err != nil {
		return models.Resource{}, err
	}
	return models.Resource{
		Project:      attributes.Project,
		Domain:       attributes.Domain,
		ResourceType: resource.String(),
		Priority:     models.ResourcePriorityProjectDomainLevel,
		Attributes:   attributeBytes,
	}, nil
}

func FromResourceModelToProjectDomainAttributes(model models.Resource) (admin.ProjectDomainAttributes, error) {
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
