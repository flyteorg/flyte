package resources

import (
	"context"
	"fmt"
	"sort"

	"google.golang.org/grpc/codes"

	"github.com/flyteorg/flyte/flyteadmin/pkg/errors"
	"github.com/flyteorg/flyte/flyteadmin/pkg/manager/impl/util"
	"github.com/flyteorg/flyte/flyteadmin/pkg/manager/impl/validation"
	"github.com/flyteorg/flyte/flyteadmin/pkg/manager/interfaces"
	repo_interface "github.com/flyteorg/flyte/flyteadmin/pkg/repositories/interfaces"
	"github.com/flyteorg/flyte/flyteadmin/pkg/repositories/transformers"
	runtimeInterfaces "github.com/flyteorg/flyte/flyteadmin/pkg/runtime/interfaces"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyte/flytestdlib/logger"
)

// ConfigurationResourceManager implements the ResourceInterface using the ConfigurationInterface
// to ensure backward compatibility so that we can deprecate the resource manager.
type ConfigurationResourceManager struct {
	db                   repo_interface.Repository
	config               runtimeInterfaces.ApplicationConfiguration
	configurationManager interfaces.ConfigurationInterface
}

type configurationWithVersion struct {
	Configuration *admin.Configuration
	Version       string
}

func (m *ConfigurationResourceManager) getConfigurationWithVersion(ctx context.Context, configurationID *admin.ConfigurationID) (*configurationWithVersion, error) {
	activeDocument, err := m.configurationManager.GetReadOnlyActiveDocument(ctx)
	if err != nil {
		return nil, err
	}

	configuration, err := util.GetConfigurationFromDocument(ctx, activeDocument, configurationID)
	if err != nil {
		return nil, err
	}

	return &configurationWithVersion{
		Configuration: configuration,
		Version:       activeDocument.Version,
	}, nil
}

func (m *ConfigurationResourceManager) ListAll(ctx context.Context, request *admin.ListMatchableAttributesRequest) (*admin.ListMatchableAttributesResponse, error) {
	if err := validation.ValidateListAllMatchableAttributesRequest(request); err != nil {
		return nil, err
	}

	activeDocument, err := m.configurationManager.GetReadOnlyActiveDocument(ctx)
	if err != nil {
		return nil, err
	}

	response := &admin.ListMatchableAttributesResponse{Configurations: make([]*admin.MatchableAttributesConfiguration, 0)}
	for level, configuration := range activeDocument.Configurations {
		id, err := util.DecodeConfigurationDocumentKey(ctx, level)
		if err != nil {
			logger.Errorf(ctx, "Failed to decode document key [%s] with error: %v", level, err)
			return nil, err
		}
		// If the configuration is a default one (domain or global), we can skip it.
		if util.IsDefaultConfigurationID(ctx, id) {
			continue
		}

		attributes, err := transformers.GetMatchingAttributesFromConfiguration(configuration, request.ResourceType)
		if err != nil {
			if errors.IsDoesNotExistError(err) {
				continue
			}
			return nil, err
		}

		response.Configurations = append(response.Configurations, &admin.MatchableAttributesConfiguration{
			Attributes: attributes,
			Project:    id.Project,
			Domain:     id.Domain,
			Workflow:   id.Workflow,
			Org:        id.Org,
		})
	}

	// Sort the response, making sure that the most specific configurations are first
	sort.Slice(response.Configurations, func(i, j int) bool {
		if response.Configurations[i].Workflow != response.Configurations[j].Workflow {
			return response.Configurations[i].Workflow > response.Configurations[j].Workflow
		}
		if response.Configurations[i].Domain != response.Configurations[j].Domain {
			return response.Configurations[i].Domain > response.Configurations[j].Domain
		}
		if response.Configurations[i].Project != response.Configurations[j].Project {
			return response.Configurations[i].Project > response.Configurations[j].Project
		}
		return response.Configurations[i].Org > response.Configurations[j].Org
	})
	return response, nil
}

func (m *ConfigurationResourceManager) GetResource(ctx context.Context, request interfaces.ResourceRequest) (*interfaces.ResourceResponse, error) {
	activeDocument, err := m.configurationManager.GetReadOnlyActiveDocument(ctx)
	if err != nil {
		return nil, err
	}

	// Check org-project-domain-workflow
	configuration, err := util.GetConfigurationFromDocument(ctx, activeDocument, &admin.ConfigurationID{
		Org:      request.Org,
		Project:  request.Project,
		Domain:   request.Domain,
		Workflow: request.Workflow,
	})
	if err != nil {
		return nil, err
	}

	attributes, err := transformers.GetMatchingAttributesFromConfiguration(configuration, request.ResourceType)
	if err != nil && !errors.IsDoesNotExistError(err) {
		return nil, err
	}
	if err == nil {
		return &interfaces.ResourceResponse{
			Org:          request.Org,
			Project:      request.Project,
			Domain:       request.Domain,
			Workflow:     request.Workflow,
			ResourceType: request.ResourceType.String(),
			Attributes:   attributes,
		}, nil
	}

	// Check org-project-domain
	configuration, err = util.GetConfigurationFromDocument(ctx, activeDocument, &admin.ConfigurationID{
		Org:     request.Org,
		Project: request.Project,
		Domain:  request.Domain,
	})
	if err != nil {
		return nil, err
	}

	attributes, err = transformers.GetMatchingAttributesFromConfiguration(configuration, request.ResourceType)
	if err != nil && !errors.IsDoesNotExistError(err) {
		return nil, err
	}
	if err == nil {
		return &interfaces.ResourceResponse{
			Org:          request.Org,
			Project:      request.Project,
			Domain:       request.Domain,
			ResourceType: request.ResourceType.String(),
			Attributes:   attributes,
		}, nil
	}

	// Check org-project
	configuration, err = util.GetConfigurationFromDocument(ctx, activeDocument, &admin.ConfigurationID{
		Org:     request.Org,
		Project: request.Project,
	})
	if err != nil {
		return nil, err
	}

	attributes, err = transformers.GetMatchingAttributesFromConfiguration(configuration, request.ResourceType)
	if err != nil && !errors.IsDoesNotExistError(err) {
		return nil, err
	}
	if err == nil {
		return &interfaces.ResourceResponse{
			Org:          request.Org,
			Project:      request.Project,
			ResourceType: request.ResourceType.String(),
			Attributes:   attributes,
		}, nil
	}

	// Check org
	configuration, err = util.GetConfigurationFromDocument(ctx, activeDocument, &admin.ConfigurationID{
		Org: request.Org,
	})
	if err != nil {
		return nil, err
	}

	attributes, err = transformers.GetMatchingAttributesFromConfiguration(configuration, request.ResourceType)
	if err != nil && !errors.IsDoesNotExistError(err) {
		return nil, err
	}
	if err == nil {
		return &interfaces.ResourceResponse{
			Org:          request.Org,
			ResourceType: request.ResourceType.String(),
			Attributes:   attributes,
		}, nil
	}

	return nil, errors.NewFlyteAdminErrorf(codes.NotFound, fmt.Sprintf("Resource [{Project:%s Domain:%s Workflow:%s LaunchPlan:%s ResourceType:%s Org:}] not found", request.Project, request.Domain, request.Workflow, request.LaunchPlan, request.ResourceType.String()))
}

func (m *ConfigurationResourceManager) UpdateOrgAttributes(ctx context.Context, request *admin.OrgAttributesUpdateRequest) (*admin.OrgAttributesUpdateResponse, error) {
	var err error
	if _, err = validation.ValidateOrgAttributesUpdateRequest(ctx, request); err != nil {
		return nil, err
	}

	configurationWithVersion, err := m.getConfigurationWithVersion(ctx, &admin.ConfigurationID{
		Org: request.Attributes.Org,
	})
	if err != nil {
		return nil, err
	}

	configurationUpdateRequest, err := transformers.FromOrgAttributesUpdateRequest(request, configurationWithVersion.Configuration, configurationWithVersion.Version)
	if err != nil {
		logger.Debugf(ctx, "Failed to transform org attributes update request into configuration update request with error: %v", err)
		return nil, err
	}

	_, err = m.configurationManager.UpdateConfiguration(ctx, configurationUpdateRequest)
	if err != nil {
		return nil, err
	}
	return &admin.OrgAttributesUpdateResponse{}, nil
}

func (m *ConfigurationResourceManager) GetOrgAttributes(ctx context.Context, request *admin.OrgAttributesGetRequest) (*admin.OrgAttributesGetResponse, error) {
	if err := validation.ValidateOrgAttributesGetRequest(ctx, request); err != nil {
		return nil, err
	}

	response, err := m.GetResource(ctx, interfaces.ResourceRequest{
		Org:          request.Org,
		ResourceType: request.ResourceType,
	})
	if err != nil {
		return nil, err
	}

	return &admin.OrgAttributesGetResponse{
		Attributes: &admin.OrgAttributes{
			Org:                response.Org,
			MatchingAttributes: response.Attributes,
		},
	}, nil
}

func (m *ConfigurationResourceManager) DeleteOrgAttributes(ctx context.Context, request *admin.OrgAttributesDeleteRequest) (*admin.OrgAttributesDeleteResponse, error) {
	if err := validation.ValidateOrgAttributesDeleteRequest(ctx, request); err != nil {
		return nil, err
	}

	configurationWithVersion, err := m.getConfigurationWithVersion(ctx, &admin.ConfigurationID{
		Org: request.Org,
	})
	if err != nil {
		return nil, err
	}

	configurationUpdateRequest, err := transformers.FromOrgAttributesDeleteRequest(request, configurationWithVersion.Configuration, configurationWithVersion.Version)
	if err != nil {
		logger.Debugf(ctx, "Failed to transform org attributes delete request into configuration update request with error: %v", err)
		return nil, err
	}

	_, err = m.configurationManager.UpdateConfiguration(ctx, configurationUpdateRequest)
	if err != nil {
		return nil, err
	}
	return &admin.OrgAttributesDeleteResponse{}, nil
}

func (m *ConfigurationResourceManager) GetProjectAttributesBase(
	ctx context.Context, request *admin.ProjectAttributesGetRequest) (
	*admin.ProjectAttributesGetResponse, error) {
	if err := validation.ValidateProjectAttributesGetRequest(ctx, m.db, request); err != nil {
		return nil, err
	}

	response, err := m.GetResource(ctx, interfaces.ResourceRequest{
		Org:          request.Org,
		Project:      request.Project,
		ResourceType: request.ResourceType,
	})
	if err != nil {
		return nil, err
	}

	return &admin.ProjectAttributesGetResponse{
		Attributes: &admin.ProjectAttributes{
			Org:                response.Org,
			Project:            response.Project,
			MatchingAttributes: response.Attributes,
		},
	}, nil
}

// For backward compatibility, GetProjectAttributes combines the call to the database to get the Project level settings with
// Admin server level configuration.
func (m *ConfigurationResourceManager) GetProjectAttributes(
	ctx context.Context, request *admin.ProjectAttributesGetRequest) (
	*admin.ProjectAttributesGetResponse, error) {
	// Get global workflow execution config
	globalConfigurationWithVersion, err := m.getConfigurationWithVersion(ctx, &util.GlobalConfigurationKey)
	if err != nil {
		return nil, err
	}

	_globalWorkflowExecutionConfig, err := transformers.GetMatchingAttributesFromConfiguration(globalConfigurationWithVersion.Configuration, admin.MatchableResource_WORKFLOW_EXECUTION_CONFIG)
	if err != nil {
		logger.Debugf(ctx, "Failed to get global workflow execution config with error: %v", err)
		return nil, err
	}
	globalWorkflowExecutionConfig := _globalWorkflowExecutionConfig.GetWorkflowExecutionConfig()
	getResponse, err := m.GetProjectAttributesBase(ctx, request)
	if err != nil {
		ec, ok := err.(errors.FlyteAdminError)
		if ok && ec.Code() == codes.NotFound && request.ResourceType == admin.MatchableResource_WORKFLOW_EXECUTION_CONFIG {
			return &admin.ProjectAttributesGetResponse{
				Attributes: &admin.ProjectAttributes{
					Org:     request.Org,
					Project: request.Project,
					MatchingAttributes: &admin.MatchingAttributes{
						Target: &admin.MatchingAttributes_WorkflowExecutionConfig{
							WorkflowExecutionConfig: globalWorkflowExecutionConfig,
						},
					},
				},
			}, nil
		}
		return nil, err
	}

	responseAttributes := getResponse.Attributes.GetMatchingAttributes().GetWorkflowExecutionConfig()
	if responseAttributes != nil {
		logger.Warningf(ctx, "Merging response %s with defaults %s", responseAttributes, globalWorkflowExecutionConfig)
		tmp := util.MergeIntoExecConfig(responseAttributes, globalWorkflowExecutionConfig)
		responseAttributes = tmp
		return &admin.ProjectAttributesGetResponse{
			Attributes: &admin.ProjectAttributes{
				Org:     request.Org,
				Project: request.Project,
				MatchingAttributes: &admin.MatchingAttributes{
					Target: &admin.MatchingAttributes_WorkflowExecutionConfig{
						WorkflowExecutionConfig: responseAttributes,
					},
				},
			},
		}, nil
	}

	return getResponse, nil
}

func (m *ConfigurationResourceManager) UpdateProjectAttributes(
	ctx context.Context, request *admin.ProjectAttributesUpdateRequest) (
	*admin.ProjectAttributesUpdateResponse, error) {
	var err error
	if _, err = validation.ValidateProjectAttributesUpdateRequest(ctx, m.db, request); err != nil {
		return nil, err
	}

	configurationWithVersion, err := m.getConfigurationWithVersion(ctx, &admin.ConfigurationID{
		Org:     request.Attributes.Org,
		Project: request.Attributes.Project,
	})
	if err != nil {
		return nil, err
	}

	configurationUpdateRequest, err := transformers.FromProjectAttributesUpdateRequest(request, configurationWithVersion.Configuration, configurationWithVersion.Version)
	if err != nil {
		logger.Debugf(ctx, "Failed to transform project attributes update request into configuration update request with error: %v", err)
		return nil, err
	}

	_, err = m.configurationManager.UpdateConfiguration(ctx, configurationUpdateRequest)
	if err != nil {
		return nil, err
	}
	return &admin.ProjectAttributesUpdateResponse{}, nil
}

func (m *ConfigurationResourceManager) DeleteProjectAttributes(
	ctx context.Context, request *admin.ProjectAttributesDeleteRequest) (
	*admin.ProjectAttributesDeleteResponse, error) {
	if err := validation.ValidateProjectAttributesDeleteRequest(ctx, m.db, request); err != nil {
		return nil, err
	}

	configurationWithVersion, err := m.getConfigurationWithVersion(ctx, &admin.ConfigurationID{
		Org:     request.Org,
		Project: request.Project,
	})
	if err != nil {
		return nil, err
	}

	configurationUpdateRequest, err := transformers.FromProjectAttributesDeleteRequest(request, configurationWithVersion.Configuration, configurationWithVersion.Version)
	if err != nil {
		logger.Debugf(ctx, "Failed to transform project attributes delete request into configuration update request with error: %v", err)
		return nil, err
	}

	_, err = m.configurationManager.UpdateConfiguration(ctx, configurationUpdateRequest)
	if err != nil {
		return nil, err
	}
	return &admin.ProjectAttributesDeleteResponse{}, nil
}

func (m *ConfigurationResourceManager) GetProjectDomainAttributes(
	ctx context.Context, request *admin.ProjectDomainAttributesGetRequest) (
	*admin.ProjectDomainAttributesGetResponse, error) {
	if err := validation.ValidateProjectDomainAttributesGetRequest(ctx, m.db, m.config, request); err != nil {
		return nil, err
	}

	response, err := m.GetResource(ctx, interfaces.ResourceRequest{
		Org:          request.Org,
		Project:      request.Project,
		Domain:       request.Domain,
		ResourceType: request.ResourceType,
	})
	if err != nil {
		return nil, err
	}

	return &admin.ProjectDomainAttributesGetResponse{
		Attributes: &admin.ProjectDomainAttributes{
			Org:                response.Org,
			Project:            response.Project,
			Domain:             response.Domain,
			MatchingAttributes: response.Attributes,
		},
	}, nil
}

func (m *ConfigurationResourceManager) UpdateProjectDomainAttributes(
	ctx context.Context, request *admin.ProjectDomainAttributesUpdateRequest) (
	*admin.ProjectDomainAttributesUpdateResponse, error) {
	var err error
	if _, err = validation.ValidateProjectDomainAttributesUpdateRequest(ctx, m.db, m.config, request); err != nil {
		return nil, err
	}

	configurationWithVersion, err := m.getConfigurationWithVersion(ctx, &admin.ConfigurationID{
		Org:     request.Attributes.Org,
		Project: request.Attributes.Project,
		Domain:  request.Attributes.Domain,
	})
	if err != nil {
		return nil, err
	}

	configurationUpdateRequest, err := transformers.FromProjectDomainAttributesUpdateRequest(request, configurationWithVersion.Configuration, configurationWithVersion.Version)
	if err != nil {
		logger.Debugf(ctx, "Failed to transform project domain attributes update request into configuration update request with error: %v", err)
		return nil, err
	}

	_, err = m.configurationManager.UpdateConfiguration(ctx, configurationUpdateRequest)
	if err != nil {
		return nil, err
	}
	return &admin.ProjectDomainAttributesUpdateResponse{}, nil
}

func (m *ConfigurationResourceManager) DeleteProjectDomainAttributes(
	ctx context.Context, request *admin.ProjectDomainAttributesDeleteRequest) (
	*admin.ProjectDomainAttributesDeleteResponse, error) {
	if err := validation.ValidateProjectDomainAttributesDeleteRequest(ctx, m.db, m.config, request); err != nil {
		return nil, err
	}

	configurationWithVersion, err := m.getConfigurationWithVersion(ctx, &admin.ConfigurationID{
		Org:     request.Org,
		Project: request.Project,
		Domain:  request.Domain,
	})
	if err != nil {
		return nil, err
	}

	configurationUpdateRequest, err := transformers.FromProjectDomainAttributesDeleteRequest(request, configurationWithVersion.Configuration, configurationWithVersion.Version)
	if err != nil {
		logger.Debugf(ctx, "Failed to transform project domain attributes delete request into configuration update request with error: %v", err)
		return nil, err
	}

	_, err = m.configurationManager.UpdateConfiguration(ctx, configurationUpdateRequest)
	if err != nil {
		return nil, err
	}
	return &admin.ProjectDomainAttributesDeleteResponse{}, nil
}

func (m *ConfigurationResourceManager) GetWorkflowAttributes(
	ctx context.Context, request *admin.WorkflowAttributesGetRequest) (
	*admin.WorkflowAttributesGetResponse, error) {
	if err := validation.ValidateWorkflowAttributesGetRequest(ctx, m.db, m.config, request); err != nil {
		return nil, err
	}

	response, err := m.GetResource(ctx, interfaces.ResourceRequest{
		Org:          request.Org,
		Project:      request.Project,
		Domain:       request.Domain,
		Workflow:     request.Workflow,
		ResourceType: request.ResourceType,
	})
	if err != nil {
		return nil, err
	}

	return &admin.WorkflowAttributesGetResponse{
		Attributes: &admin.WorkflowAttributes{
			Project:            response.Project,
			Domain:             response.Domain,
			Workflow:           response.Workflow,
			Org:                response.Org,
			MatchingAttributes: response.Attributes,
		},
	}, nil
}

func (m *ConfigurationResourceManager) UpdateWorkflowAttributes(
	ctx context.Context, request *admin.WorkflowAttributesUpdateRequest) (
	*admin.WorkflowAttributesUpdateResponse, error) {
	var err error
	if _, err = validation.ValidateWorkflowAttributesUpdateRequest(ctx, m.db, m.config, request); err != nil {
		return nil, err
	}

	configurationWithVersion, err := m.getConfigurationWithVersion(ctx, &admin.ConfigurationID{
		Org:      request.Attributes.Org,
		Project:  request.Attributes.Project,
		Domain:   request.Attributes.Domain,
		Workflow: request.Attributes.Workflow,
	})
	if err != nil {
		return nil, err
	}

	configurationUpdateRequest, err := transformers.FromWorkflowAttributesUpdateRequest(request, configurationWithVersion.Configuration, configurationWithVersion.Version)
	if err != nil {
		logger.Debugf(ctx, "Failed to transform workflow attributes update request into configuration update request with error: %v", err)
		return nil, err
	}

	_, err = m.configurationManager.UpdateConfiguration(ctx, configurationUpdateRequest)
	if err != nil {
		return nil, err
	}
	return &admin.WorkflowAttributesUpdateResponse{}, nil
}

func (m *ConfigurationResourceManager) DeleteWorkflowAttributes(
	ctx context.Context, request *admin.WorkflowAttributesDeleteRequest) (
	*admin.WorkflowAttributesDeleteResponse, error) {
	if err := validation.ValidateWorkflowAttributesDeleteRequest(ctx, m.db, m.config, request); err != nil {
		return nil, err
	}

	configurationWithVersion, err := m.getConfigurationWithVersion(ctx, &admin.ConfigurationID{
		Org:      request.Org,
		Project:  request.Project,
		Domain:   request.Domain,
		Workflow: request.Workflow,
	})
	if err != nil {
		return nil, err
	}

	configurationUpdateRequest, err := transformers.FromWorkflowAttributesDeleteRequest(request, configurationWithVersion.Configuration, configurationWithVersion.Version)
	if err != nil {
		logger.Debugf(ctx, "Failed to transform workflow attributes delete request into configuration update request with error: %v", err)
		return nil, err
	}

	_, err = m.configurationManager.UpdateConfiguration(ctx, configurationUpdateRequest)
	if err != nil {
		return nil, err
	}
	return &admin.WorkflowAttributesDeleteResponse{}, nil
}

func NewConfigurationResourceManager(db repo_interface.Repository, config runtimeInterfaces.ApplicationConfiguration, configurationManager interfaces.ConfigurationInterface) interfaces.ResourceInterface {
	return &ConfigurationResourceManager{
		db:                   db,
		config:               config,
		configurationManager: configurationManager,
	}
}
