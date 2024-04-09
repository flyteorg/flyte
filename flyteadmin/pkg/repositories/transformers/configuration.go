package transformers

import (
	"github.com/flyteorg/flyte/flyteadmin/pkg/repositories/models"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
)

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

// transform ProjectDomainAttributesUpdateRequest to ConfigurationUpdateRequest
func FromProjectDomainAttributesUpdateRequest(request *admin.ProjectDomainAttributesUpdateRequest) *admin.ConfigurationUpdateRequest {
	return &admin.ConfigurationUpdateRequest{
		Id: &admin.ConfigurationID{
			Project: request.Attributes.Project,
			Domain:  request.Attributes.Domain,
		},
		Configuration: GetConfigurationFromMatchingAttributes(request.Attributes.MatchingAttributes),
	}
}

// transform ProjectDomainAttributesDeleteRequest to ConfigurationUpdateRequest
func FromProjectDomainAttributesDeleteRequest(request *admin.ProjectDomainAttributesDeleteRequest) *admin.ConfigurationUpdateRequest {
	return &admin.ConfigurationUpdateRequest{
		Id: &admin.ConfigurationID{
			Project: request.Project,
			Domain:  request.Domain,
		},
		Configuration: &admin.Configuration{},
	}
}

// transform ProjectAttributesUpdateRequest to ConfigurationUpdateRequest
func FromProjectAttributesUpdateRequest(request *admin.ProjectAttributesUpdateRequest) *admin.ConfigurationUpdateRequest {
	return &admin.ConfigurationUpdateRequest{
		Id: &admin.ConfigurationID{
			Project: request.Attributes.Project,
		},
		Configuration: GetConfigurationFromMatchingAttributes(request.Attributes.MatchingAttributes),
	}
}

// transform models.Resource to admin.ConfigurationDocument
func FromResourceModelToConfiguration(resource models.Resource) *admin.Configuration {
	matchableResource, err := FromResourceModelToMatchableAttributes(resource)
	if err != nil {
		return &admin.Configuration{}
	}
	configuration := GetConfigurationFromMatchingAttributes(matchableResource.GetAttributes())
	return configuration
}
