package transformers

import (
	"google.golang.org/grpc/codes"

	flyteErrs "github.com/flyteorg/flyte/flyteadmin/pkg/errors"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
)

var attributeNotFoundError = flyteErrs.NewFlyteAdminErrorf(codes.NotFound, "attribute not found")

func updateMatchingAttributesInConfiguration(matchingAttributes *admin.MatchingAttributes, configuration *admin.Configuration) *admin.Configuration {
	switch x := matchingAttributes.GetTarget().(type) {
	case *admin.MatchingAttributes_TaskResourceAttributes:
		configuration.TaskResourceAttributes = x.TaskResourceAttributes
	case *admin.MatchingAttributes_ClusterResourceAttributes:
		configuration.ClusterResourceAttributes = x.ClusterResourceAttributes
	case *admin.MatchingAttributes_ExecutionQueueAttributes:
		configuration.ExecutionQueueAttributes = x.ExecutionQueueAttributes
	case *admin.MatchingAttributes_ExecutionClusterLabel:
		configuration.ExecutionClusterLabel = x.ExecutionClusterLabel
	case *admin.MatchingAttributes_QualityOfService:
		configuration.QualityOfService = x.QualityOfService
	case *admin.MatchingAttributes_PluginOverrides:
		configuration.PluginOverrides = x.PluginOverrides
	case *admin.MatchingAttributes_WorkflowExecutionConfig:
		configuration.WorkflowExecutionConfig = x.WorkflowExecutionConfig
	case *admin.MatchingAttributes_ClusterAssignment:
		configuration.ClusterAssignment = x.ClusterAssignment
	}
	return configuration
}

func deleteMatchableResourceFromConfiguration(resourceType admin.MatchableResource, configuration *admin.Configuration) *admin.Configuration {
	switch resourceType {
	case admin.MatchableResource_TASK_RESOURCE:
		configuration.TaskResourceAttributes = nil
	case admin.MatchableResource_CLUSTER_RESOURCE:
		configuration.ClusterResourceAttributes = nil
	case admin.MatchableResource_EXECUTION_QUEUE:
		configuration.ExecutionQueueAttributes = nil
	case admin.MatchableResource_EXECUTION_CLUSTER_LABEL:
		configuration.ExecutionClusterLabel = nil
	case admin.MatchableResource_QUALITY_OF_SERVICE_SPECIFICATION:
		configuration.QualityOfService = nil
	case admin.MatchableResource_PLUGIN_OVERRIDE:
		configuration.PluginOverrides = nil
	case admin.MatchableResource_WORKFLOW_EXECUTION_CONFIG:
		configuration.WorkflowExecutionConfig = nil
	case admin.MatchableResource_CLUSTER_ASSIGNMENT:
		configuration.ClusterAssignment = nil
	}
	return configuration
}

func GetMatchingAttributesFromConfiguration(configuration *admin.Configuration, resourceType admin.MatchableResource) (*admin.MatchingAttributes, error) {
	if configuration == nil {
		return nil, attributeNotFoundError
	}
	switch resourceType {
	case admin.MatchableResource_TASK_RESOURCE:
		if configuration.TaskResourceAttributes == nil {
			return nil, attributeNotFoundError
		}
		return &admin.MatchingAttributes{Target: &admin.MatchingAttributes_TaskResourceAttributes{TaskResourceAttributes: configuration.TaskResourceAttributes}}, nil
	case admin.MatchableResource_CLUSTER_RESOURCE:
		if configuration.ClusterResourceAttributes == nil {
			return nil, attributeNotFoundError
		}
		return &admin.MatchingAttributes{Target: &admin.MatchingAttributes_ClusterResourceAttributes{ClusterResourceAttributes: configuration.ClusterResourceAttributes}}, nil
	case admin.MatchableResource_EXECUTION_QUEUE:
		if configuration.ExecutionQueueAttributes == nil {
			return nil, attributeNotFoundError
		}
		return &admin.MatchingAttributes{Target: &admin.MatchingAttributes_ExecutionQueueAttributes{ExecutionQueueAttributes: configuration.ExecutionQueueAttributes}}, nil
	case admin.MatchableResource_EXECUTION_CLUSTER_LABEL:
		if configuration.ExecutionClusterLabel == nil {
			return nil, attributeNotFoundError
		}
		return &admin.MatchingAttributes{Target: &admin.MatchingAttributes_ExecutionClusterLabel{ExecutionClusterLabel: configuration.ExecutionClusterLabel}}, nil
	case admin.MatchableResource_QUALITY_OF_SERVICE_SPECIFICATION:
		if configuration.QualityOfService == nil {
			return nil, attributeNotFoundError
		}
		return &admin.MatchingAttributes{Target: &admin.MatchingAttributes_QualityOfService{QualityOfService: configuration.QualityOfService}}, nil
	case admin.MatchableResource_PLUGIN_OVERRIDE:
		if configuration.PluginOverrides == nil {
			return nil, attributeNotFoundError
		}
		return &admin.MatchingAttributes{Target: &admin.MatchingAttributes_PluginOverrides{PluginOverrides: configuration.PluginOverrides}}, nil
	case admin.MatchableResource_WORKFLOW_EXECUTION_CONFIG:
		if configuration.WorkflowExecutionConfig == nil {
			return nil, attributeNotFoundError
		}
		return &admin.MatchingAttributes{Target: &admin.MatchingAttributes_WorkflowExecutionConfig{WorkflowExecutionConfig: configuration.WorkflowExecutionConfig}}, nil
	case admin.MatchableResource_CLUSTER_ASSIGNMENT:
		if configuration.ClusterAssignment == nil {
			return nil, attributeNotFoundError
		}
		return &admin.MatchingAttributes{Target: &admin.MatchingAttributes_ClusterAssignment{ClusterAssignment: configuration.ClusterAssignment}}, nil
	default:
		return nil, attributeNotFoundError
	}
}

// FromWorkflowAttributesUpdateRequest transform WorkflowAttributesUpdateRequest to ConfigurationUpdateRequest
func FromWorkflowAttributesUpdateRequest(request *admin.WorkflowAttributesUpdateRequest, configuration *admin.Configuration, version string) *admin.ConfigurationUpdateRequest {
	newConfiguration := updateMatchingAttributesInConfiguration(request.Attributes.MatchingAttributes, configuration)
	return &admin.ConfigurationUpdateRequest{
		Id: &admin.ConfigurationID{
			Project:  request.Attributes.Project,
			Domain:   request.Attributes.Domain,
			Workflow: request.Attributes.Workflow,
		},
		VersionToUpdate: version,
		Configuration:   newConfiguration,
	}
}

// FromWorkflowAttributesDeleteRequest transform WorkflowAttributesDeleteRequest to ConfigurationUpdateRequest
func FromWorkflowAttributesDeleteRequest(request *admin.WorkflowAttributesDeleteRequest, configuration *admin.Configuration, version string) *admin.ConfigurationUpdateRequest {
	newConfiguration := deleteMatchableResourceFromConfiguration(request.GetResourceType(), configuration)
	return &admin.ConfigurationUpdateRequest{
		Id: &admin.ConfigurationID{
			Project:  request.Project,
			Domain:   request.Domain,
			Workflow: request.Workflow,
		},
		VersionToUpdate: version,
		Configuration:   newConfiguration,
	}
}

func ToWorkflowAttributesGetResponse(configuration *admin.Configuration, request *admin.WorkflowAttributesGetRequest) (*admin.WorkflowAttributesGetResponse, error) {
	matchingAttributes, err := GetMatchingAttributesFromConfiguration(configuration, request.GetResourceType())
	if err != nil {
		return nil, err
	}
	return &admin.WorkflowAttributesGetResponse{
		Attributes: &admin.WorkflowAttributes{
			Project:            request.Project,
			Domain:             request.Domain,
			Workflow:           request.Workflow,
			MatchingAttributes: matchingAttributes,
		},
	}, nil
}

// FromProjectDomainAttributesUpdateRequest transform ProjectDomainAttributesUpdateRequest to ConfigurationUpdateRequest
func FromProjectDomainAttributesUpdateRequest(request *admin.ProjectDomainAttributesUpdateRequest, configuration *admin.Configuration, version string) *admin.ConfigurationUpdateRequest {
	newConfiguration := updateMatchingAttributesInConfiguration(request.Attributes.MatchingAttributes, configuration)
	return &admin.ConfigurationUpdateRequest{
		Id: &admin.ConfigurationID{
			Project: request.Attributes.Project,
			Domain:  request.Attributes.Domain,
		},
		VersionToUpdate: version,
		Configuration:   newConfiguration,
	}
}

// FromProjectDomainAttributesDeleteRequest transform ProjectDomainAttributesDeleteRequest to ConfigurationUpdateRequest
func FromProjectDomainAttributesDeleteRequest(request *admin.ProjectDomainAttributesDeleteRequest, configuration *admin.Configuration, version string) *admin.ConfigurationUpdateRequest {
	newConfiguration := deleteMatchableResourceFromConfiguration(request.GetResourceType(), configuration)
	return &admin.ConfigurationUpdateRequest{
		Id: &admin.ConfigurationID{
			Project: request.Project,
			Domain:  request.Domain,
		},
		VersionToUpdate: version,
		Configuration:   newConfiguration,
	}
}

// ToProjectDomainAttributesGetResponse transform ConfigurationGetResponse to ProjectDomainAttributesGetResponse
func ToProjectDomainAttributesGetResponse(configuration *admin.Configuration, request *admin.ProjectDomainAttributesGetRequest) (*admin.ProjectDomainAttributesGetResponse, error) {
	matchingAttributes, err := GetMatchingAttributesFromConfiguration(configuration, request.GetResourceType())
	if err != nil {
		return nil, err
	}
	return &admin.ProjectDomainAttributesGetResponse{
		Attributes: &admin.ProjectDomainAttributes{
			Project:            request.Project,
			Domain:             request.Domain,
			MatchingAttributes: matchingAttributes,
		},
	}, nil
}

// FromProjectAttributesUpdateRequest transform ProjectAttributesUpdateRequest to ConfigurationUpdateRequest
func FromProjectAttributesUpdateRequest(request *admin.ProjectAttributesUpdateRequest, configuration *admin.Configuration, version string) *admin.ConfigurationUpdateRequest {
	newConfiguration := updateMatchingAttributesInConfiguration(request.Attributes.MatchingAttributes, configuration)
	return &admin.ConfigurationUpdateRequest{
		Id: &admin.ConfigurationID{
			Project: request.Attributes.Project,
		},
		VersionToUpdate: version,
		Configuration:   newConfiguration,
	}
}

// FromProjectAttributesDeleteRequest transform ProjectAttributesDeleteRequest to ConfigurationUpdateRequest
func FromProjectAttributesDeleteRequest(request *admin.ProjectAttributesDeleteRequest, configuration *admin.Configuration, version string) *admin.ConfigurationUpdateRequest {
	newConfiguration := deleteMatchableResourceFromConfiguration(request.GetResourceType(), configuration)
	return &admin.ConfigurationUpdateRequest{
		Id: &admin.ConfigurationID{
			Project: request.Project,
		},
		VersionToUpdate: version,
		Configuration:   newConfiguration,
	}
}

// ToProjectAttributesGetResponse transform ConfigurationGetResponse to ProjectAttributesGetResponse
func ToProjectAttributesGetResponse(configuration *admin.Configuration, request *admin.ProjectAttributesGetRequest) (*admin.ProjectAttributesGetResponse, error) {
	matchingAttributes, err := GetMatchingAttributesFromConfiguration(configuration, request.GetResourceType())
	if err != nil {
		return nil, err
	}
	return &admin.ProjectAttributesGetResponse{
		Attributes: &admin.ProjectAttributes{
			Project:            request.Project,
			MatchingAttributes: matchingAttributes,
		},
	}, nil
}
