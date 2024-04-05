package transformers

import "github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"

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

func GetConfigurationFromMatchableResource(matchableResource *admin.MatchableResource) *admin.Configuration {
	switch matchableResource {
	case admin.MatchableResource_TASK_RESOURCE:
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

// transform ProjectDomainAttributesUpdateRequest to ConfigurationUpdateRequest
func FromProjectDomainAttributesUpdateRequest(request *admin.ProjectDomainAttributesUpdateRequest) *admin.ConfigurationUpdateRequest {
	return &admin.ConfigurationUpdateRequest{
		Id: &admin.ProjectID{
			Project: request.Attributes.Project,
			Domain:  request.Attributes.Domain,
		},
		Configuration: GetConfigurationFromMatchingAttributes(request.Attributes.MatchingAttributes),
	}
}

// transform ProjectDomainAttributesDeleteRequest to ConfigurationUpdateRequest
func FromProjectDomainAttributesDeleteRequest(request *admin.ProjectDomainAttributesDeleteRequest) *admin.ConfigurationUpdateRequest {
	return &admin.ConfigurationUpdateRequest{
		Id: &admin.ProjectID{
			Project: request.Project,
			Domain:  request.Domain,
		},
		Configuration: &admin.Configuration{},
	}
}

// transform ProjectAttributesUpdateRequest to ConfigurationUpdateRequest
func FromProjectAttributesUpdateRequest(request *admin.ProjectAttributesUpdateRequest) *admin.ConfigurationUpdateRequest {
	return &admin.ConfigurationUpdateRequest{
		Id: &admin.ProjectID{
			Project: request.Attributes.Project,
		},
		Configuration: GetConfigurationFromMatchingAttributes(request.Attributes.MatchingAttributes),
	}
}
