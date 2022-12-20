package transformers

import (
	"context"

	repoInterfaces "github.com/flyteorg/flyteadmin/pkg/repositories/interfaces"
	"github.com/flyteorg/flytestdlib/logger"
	"github.com/golang/protobuf/proto"

	"github.com/flyteorg/flyteadmin/pkg/errors"
	"github.com/flyteorg/flyteadmin/pkg/repositories/models"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
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

func mergeUpdatePluginOverrides(existingAttributes admin.MatchingAttributes,
	newMatchingAttributes *admin.MatchingAttributes) *admin.MatchingAttributes {
	taskPluginOverrides := make(map[string]*admin.PluginOverride)
	if existingAttributes.GetPluginOverrides() != nil && len(existingAttributes.GetPluginOverrides().Overrides) > 0 {
		for _, pluginOverride := range existingAttributes.GetPluginOverrides().Overrides {
			taskPluginOverrides[pluginOverride.TaskType] = pluginOverride
		}
	}
	if newMatchingAttributes.GetPluginOverrides() != nil &&
		len(newMatchingAttributes.GetPluginOverrides().Overrides) > 0 {
		for _, pluginOverride := range newMatchingAttributes.GetPluginOverrides().Overrides {
			taskPluginOverrides[pluginOverride.TaskType] = pluginOverride
		}
	}

	updatedPluginOverrides := make([]*admin.PluginOverride, 0, len(taskPluginOverrides))
	for _, pluginOverride := range taskPluginOverrides {
		updatedPluginOverrides = append(updatedPluginOverrides, pluginOverride)
	}
	return &admin.MatchingAttributes{
		Target: &admin.MatchingAttributes_PluginOverrides{
			PluginOverrides: &admin.PluginOverrides{
				Overrides: updatedPluginOverrides,
			},
		},
	}
}

func MergeUpdateWorkflowAttributes(ctx context.Context, model models.Resource, resource admin.MatchableResource,
	resourceID *repoInterfaces.ResourceID, workflowAttributes *admin.WorkflowAttributes) (models.Resource, error) {
	switch resource {
	case admin.MatchableResource_PLUGIN_OVERRIDE:
		var existingAttributes admin.MatchingAttributes
		err := proto.Unmarshal(model.Attributes, &existingAttributes)
		if err != nil {
			return models.Resource{}, errors.NewFlyteAdminErrorf(codes.Internal,
				"Unable to unmarshal existing resource attributes for [%+v] with err: %v", resourceID, err)
		}
		updatedAttributes := mergeUpdatePluginOverrides(existingAttributes, workflowAttributes.GetMatchingAttributes())
		marshaledAttributes, err := proto.Marshal(updatedAttributes)
		if err != nil {
			return models.Resource{}, errors.NewFlyteAdminErrorf(codes.Internal,
				"Failed to marshal merge-updated attributes for [%+v] with err: %v", resourceID, err)
		}
		model.Attributes = marshaledAttributes
		return model, nil
	default:
		logger.Warningf(ctx, "Tried to merge-update an unsupported resource type [%s] for [%+v]",
			resource.String(), resourceID)
		return models.Resource{}, errors.NewFlyteAdminErrorf(codes.Internal,
			"Tried to merge-update an unsupported resource type [%s] for [%+v]",
			resource.String(), resourceID)
	}
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

func ProjectAttributesToResourceModel(attributes admin.ProjectAttributes, resource admin.MatchableResource) (models.Resource, error) {
	attributeBytes, err := proto.Marshal(attributes.MatchingAttributes)
	if err != nil {
		return models.Resource{}, err
	}
	return models.Resource{
		Project:      attributes.Project,
		ResourceType: resource.String(),
		Priority:     models.ResourcePriorityProjectLevel,
		Attributes:   attributeBytes,
	}, nil
}

// MergeUpdatePluginAttributes only handles plugin overrides. Other attributes are just overridden when an
// update happens.
func MergeUpdatePluginAttributes(ctx context.Context, model models.Resource, resource admin.MatchableResource,
	resourceID *repoInterfaces.ResourceID, matchingAttributes *admin.MatchingAttributes) (models.Resource, error) {
	switch resource {
	case admin.MatchableResource_PLUGIN_OVERRIDE:
		var existingAttributes admin.MatchingAttributes
		err := proto.Unmarshal(model.Attributes, &existingAttributes)
		if err != nil {
			return models.Resource{}, errors.NewFlyteAdminErrorf(codes.Internal,
				"Unable to unmarshal existing resource attributes for [%+v] with err: %v", resourceID, err)
		}
		updatedAttributes := mergeUpdatePluginOverrides(existingAttributes, matchingAttributes)
		marshaledAttributes, err := proto.Marshal(updatedAttributes)
		if err != nil {
			return models.Resource{}, errors.NewFlyteAdminErrorf(codes.Internal,
				"Failed to marshal merge-updated attributes for [%+v] with err: %v", resourceID, err)
		}
		model.Attributes = marshaledAttributes
		return model, nil
	default:
		logger.Warningf(ctx, "Tried to merge-update an unsupported resource type [%s] for [%+v]",
			resource.String(), resourceID)
		return models.Resource{}, errors.NewFlyteAdminErrorf(codes.Internal,
			"Tried to merge-update an unsupported resource type [%s] for [%+v]",
			resource.String(), resourceID)
	}
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

func FromResourceModelToMatchableAttributes(model models.Resource) (admin.MatchableAttributesConfiguration, error) {
	var attributes admin.MatchingAttributes
	err := proto.Unmarshal(model.Attributes, &attributes)
	if err != nil {
		return admin.MatchableAttributesConfiguration{}, errors.NewFlyteAdminErrorf(
			codes.Internal, "Failed to decode MatchableAttributesConfiguration with err: %v", err)
	}
	return admin.MatchableAttributesConfiguration{
		Attributes: &attributes,
		Project:    model.Project,
		Domain:     model.Domain,
		Workflow:   model.Workflow,
		LaunchPlan: model.LaunchPlan,
	}, nil
}

func FromResourceModelsToMatchableAttributes(models []models.Resource) (
	[]*admin.MatchableAttributesConfiguration, error) {
	configs := make([]*admin.MatchableAttributesConfiguration, len(models))
	for idx, model := range models {
		attributesConfig, err := FromResourceModelToMatchableAttributes(model)
		if err != nil {
			return nil, err
		}
		configs[idx] = &attributesConfig
	}
	return configs, nil
}
