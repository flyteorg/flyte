package transformers

import (
	flyteErrs "github.com/flyteorg/flyte/flyteadmin/pkg/errors"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"google.golang.org/grpc/codes"
)

var attributeNotFoundError = flyteErrs.NewFlyteAdminErrorf(codes.NotFound, "attribute not found")

// get the real type of matching attributes
func GetConfigurationFromMatchingAttributes(matchingAttributes *admin.MatchingAttributes) *admin.Configuration {
	switch x := matchingAttributes.GetTarget().(type) {
	case *admin.MatchingAttributes_TaskResourceAttributes:
		return &admin.Configuration{TaskResourceAttributes: x.TaskResourceAttributes}
	case *admin.MatchingAttributes_ClusterResourceAttributes:
		return &admin.Configuration{ClusterResourceAttributes: x.ClusterResourceAttributes}
	case *admin.MatchingAttributes_ExecutionQueueAttributes:
		return &admin.Configuration{ExecutionQueueAttributes: x.ExecutionQueueAttributes}
	case *admin.MatchingAttributes_ExecutionClusterLabel:
		return &admin.Configuration{ExecutionClusterLabel: x.ExecutionClusterLabel}
	case *admin.MatchingAttributes_QualityOfService:
		return &admin.Configuration{QualityOfService: x.QualityOfService}
	case *admin.MatchingAttributes_PluginOverrides:
		return &admin.Configuration{PluginOverrides: x.PluginOverrides}
	case *admin.MatchingAttributes_WorkflowExecutionConfig:
		return &admin.Configuration{WorkflowExecutionConfig: x.WorkflowExecutionConfig}
	case *admin.MatchingAttributes_ClusterAssignment:
		return &admin.Configuration{ClusterAssignment: x.ClusterAssignment}
	default:
		return &admin.Configuration{}
	}
}

func GetConfigurationFromMatchableResource(matchableResource admin.MatchableResource) *admin.Configuration {
	switch matchableResource {
	case admin.MatchableResource_TASK_RESOURCE:
		return &admin.Configuration{TaskResourceAttributes: &admin.TaskResourceAttributes{}}
	case admin.MatchableResource_CLUSTER_RESOURCE:
		return &admin.Configuration{ClusterResourceAttributes: &admin.ClusterResourceAttributes{}}
	case admin.MatchableResource_EXECUTION_QUEUE:
		return &admin.Configuration{ExecutionQueueAttributes: &admin.ExecutionQueueAttributes{}}
	case admin.MatchableResource_EXECUTION_CLUSTER_LABEL:
		return &admin.Configuration{ExecutionClusterLabel: &admin.ExecutionClusterLabel{}}
	case admin.MatchableResource_QUALITY_OF_SERVICE_SPECIFICATION:
		return &admin.Configuration{QualityOfService: &core.QualityOfService{}}
	case admin.MatchableResource_PLUGIN_OVERRIDE:
		return &admin.Configuration{PluginOverrides: &admin.PluginOverrides{}}
	case admin.MatchableResource_WORKFLOW_EXECUTION_CONFIG:
		return &admin.Configuration{WorkflowExecutionConfig: &admin.WorkflowExecutionConfig{}}
	case admin.MatchableResource_CLUSTER_ASSIGNMENT:
		return &admin.Configuration{ClusterAssignment: &admin.ClusterAssignment{}}
	default:
		return &admin.Configuration{}
	}
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
func FromWorkflowAttributesUpdateRequest(request *admin.WorkflowAttributesUpdateRequest) *admin.ConfigurationUpdateRequest {
	return &admin.ConfigurationUpdateRequest{
		Id: &admin.ConfigurationID{
			Project:  request.Attributes.Project,
			Domain:   request.Attributes.Domain,
			Workflow: request.Attributes.Workflow,
		},
		Configuration: GetConfigurationFromMatchingAttributes(request.Attributes.MatchingAttributes),
	}
}

// FromWorkflowAttributesDeleteRequest transform WorkflowAttributesDeleteRequest to ConfigurationUpdateRequest
func FromWorkflowAttributesDeleteRequest(request *admin.WorkflowAttributesDeleteRequest) *admin.ConfigurationUpdateRequest {
	return &admin.ConfigurationUpdateRequest{
		Id: &admin.ConfigurationID{
			Project:  request.Project,
			Domain:   request.Domain,
			Workflow: request.Workflow,
		},
		Configuration: GetConfigurationFromMatchableResource(request.GetResourceType()),
	}
}

// FromWorkflowAttributesGetRequest transform WorkflowAttributesGetRequest to ConfigurationGetRequest
func FromWorkflowAttributesGetRequest(request *admin.WorkflowAttributesGetRequest) *admin.ConfigurationGetRequest {
	return &admin.ConfigurationGetRequest{
		Id: &admin.ConfigurationID{
			Project:  request.Project,
			Domain:   request.Domain,
			Workflow: request.Workflow,
		},
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
func FromProjectDomainAttributesUpdateRequest(request *admin.ProjectDomainAttributesUpdateRequest) *admin.ConfigurationUpdateRequest {
	return &admin.ConfigurationUpdateRequest{
		Id: &admin.ConfigurationID{
			Project: request.Attributes.Project,
			Domain:  request.Attributes.Domain,
		},
		Configuration: GetConfigurationFromMatchingAttributes(request.Attributes.MatchingAttributes),
	}
}

// FromProjectDomainAttributesDeleteRequest transform ProjectDomainAttributesDeleteRequest to ConfigurationUpdateRequest
func FromProjectDomainAttributesDeleteRequest(request *admin.ProjectDomainAttributesDeleteRequest) *admin.ConfigurationUpdateRequest {
	return &admin.ConfigurationUpdateRequest{
		Id: &admin.ConfigurationID{
			Project: request.Project,
			Domain:  request.Domain,
		},
		Configuration: GetConfigurationFromMatchableResource(request.GetResourceType()),
	}
}

// FromProjectDomainAttributesGetRequest transform ProjectDomainAttributesGetRequest to ConfigurationGetRequest
func FromProjectDomainAttributesGetRequest(request *admin.ProjectDomainAttributesGetRequest) *admin.ConfigurationGetRequest {
	return &admin.ConfigurationGetRequest{
		Id: &admin.ConfigurationID{
			Project: request.Project,
			Domain:  request.Domain,
		},
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
func FromProjectAttributesUpdateRequest(request *admin.ProjectAttributesUpdateRequest) *admin.ConfigurationUpdateRequest {
	return &admin.ConfigurationUpdateRequest{
		Id: &admin.ConfigurationID{
			Project: request.Attributes.Project,
		},
		Configuration: GetConfigurationFromMatchingAttributes(request.Attributes.MatchingAttributes),
	}
}

// FromProjectAttributesDeleteRequest transform ProjectAttributesDeleteRequest to ConfigurationUpdateRequest
func FromProjectAttributesDeleteRequest(request *admin.ProjectAttributesDeleteRequest) *admin.ConfigurationUpdateRequest {
	return &admin.ConfigurationUpdateRequest{
		Id: &admin.ConfigurationID{
			Project: request.Project,
		},
		Configuration: GetConfigurationFromMatchableResource(request.GetResourceType()),
	}
}

// FromProjectAttributesGetRequest transform ProjectAttributesGetRequest to ConfigurationGetRequest
func FromProjectAttributesGetRequest(request *admin.ProjectAttributesGetRequest) *admin.ConfigurationGetRequest {
	return &admin.ConfigurationGetRequest{
		Id: &admin.ConfigurationID{
			Project: request.Project,
		},
	}
}

// ToProjectAttributesGetResponse transform ConfigurationGetResponse to ProjectAttributesGetResponse
func ToProjectAttributesGetResponse(configuration *admin.Configuration, request *admin.ProjectAttributesGetRequest) (*admin.ProjectAttributesGetResponse, error) {
	matchingAttributes, err := GetMatchingAttributesFromConfiguration(configuration, request.GetResourceType()
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
