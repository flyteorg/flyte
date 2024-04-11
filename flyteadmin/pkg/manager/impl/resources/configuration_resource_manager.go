package resources

import (
	"context"
	"github.com/flyteorg/flyte/flyteadmin/pkg/manager/impl/configurations"
	"github.com/flyteorg/flyte/flyteadmin/pkg/manager/impl/validation"
	"github.com/flyteorg/flyte/flyteadmin/pkg/manager/interfaces"
	repo_interface "github.com/flyteorg/flyte/flyteadmin/pkg/repositories/interfaces"
	"github.com/flyteorg/flyte/flyteadmin/pkg/repositories/transformers"
	runtimeInterfaces "github.com/flyteorg/flyte/flyteadmin/pkg/runtime/interfaces"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
)

type ConfigurationResourceManager struct {
	db                   repo_interface.Repository
	config               runtimeInterfaces.ApplicationConfiguration
	configurationManager interfaces.ConfigurationInterface
}

func (m *ConfigurationResourceManager) ListAll(ctx context.Context, request admin.ListMatchableAttributesRequest) (*admin.ListMatchableAttributesResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (m *ConfigurationResourceManager) GetResource(ctx context.Context, request interfaces.ResourceRequest) (*interfaces.ResourceResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (m *ConfigurationResourceManager) GetProjectAttributes(
	ctx context.Context, request admin.ProjectAttributesGetRequest) (
	*admin.ProjectAttributesGetResponse, error) {
	if err := validation.ValidateProjectAttributesGetRequest(ctx, m.db, request); err != nil {
		return nil, err
	}
	// Get the active document
	activeDocument, err := m.configurationManager.GetActiveDocument(ctx)
	if err != nil {
		return nil, err
	}

	projectConfiguration := configurations.GetProjectConfigurationFromDocument(activeDocument, request.Project)
	projectAttributesGetResponse, err := transformers.ToProjectAttributesGetResponse(projectConfiguration, &request)
	if err != nil {
		return nil, err
	}
	return projectAttributesGetResponse, nil
}

func (m *ConfigurationResourceManager) UpdateProjectAttributes(
	ctx context.Context, request admin.ProjectAttributesUpdateRequest) (
	*admin.ProjectAttributesUpdateResponse, error) {
	var err error
	if _, err = validation.ValidateProjectAttributesUpdateRequest(ctx, m.db, request); err != nil {
		return nil, err
	}

	activeDocument, err := m.configurationManager.GetActiveDocument(ctx)
	if err != nil {
		return nil, err
	}

	configuration := configurations.GetProjectConfigurationFromDocument(activeDocument, request.Attributes.Project)
	configurationUpdateRequest := transformers.FromProjectAttributesUpdateRequest(&request, configuration, activeDocument.Version)
	_, err = m.configurationManager.UpdateConfiguration(ctx, *configurationUpdateRequest)
	if err != nil {
		return nil, err
	}
	return &admin.ProjectAttributesUpdateResponse{}, nil
}

func (m *ConfigurationResourceManager) DeleteProjectAttributes(
	ctx context.Context, request admin.ProjectAttributesDeleteRequest) (
	*admin.ProjectAttributesDeleteResponse, error) {
	if err := validation.ValidateProjectAttributesDeleteRequest(ctx, m.db, request); err != nil {
		return nil, err
	}

	activeDocument, err := m.configurationManager.GetActiveDocument(ctx)
	if err != nil {
		return nil, err
	}

	configuration := configurations.GetProjectConfigurationFromDocument(activeDocument, request.Project)
	configurationUpdateRequest := transformers.FromProjectAttributesDeleteRequest(&request, configuration, activeDocument.Version)
	_, err = m.configurationManager.UpdateConfiguration(ctx, *configurationUpdateRequest)
	if err != nil {
		return nil, err
	}
	return &admin.ProjectAttributesDeleteResponse{}, nil
}

func (m *ConfigurationResourceManager) GetProjectDomainAttributes(
	ctx context.Context, request admin.ProjectDomainAttributesGetRequest) (
	*admin.ProjectDomainAttributesGetResponse, error) {
	if err := validation.ValidateProjectDomainAttributesGetRequest(ctx, m.db, m.config, request); err != nil {
		return nil, err
	}
	// Get the active document
	activeDocument, err := m.configurationManager.GetActiveDocument(ctx)
	if err != nil {
		return nil, err
	}

	projectDomainConfiguration := configurations.GetProjectDomainConfigurationFromDocument(activeDocument, request.Project, request.Domain)
	projectDomainAttributesGetResponse, err := transformers.ToProjectDomainAttributesGetResponse(projectDomainConfiguration, &request)
	if err != nil {
		return nil, err
	}
	return projectDomainAttributesGetResponse, nil
}

func (m *ConfigurationResourceManager) UpdateProjectDomainAttributes(
	ctx context.Context, request admin.ProjectDomainAttributesUpdateRequest) (
	*admin.ProjectDomainAttributesUpdateResponse, error) {
	var err error
	if _, err = validation.ValidateProjectDomainAttributesUpdateRequest(ctx, m.db, m.config, request); err != nil {
		return nil, err
	}

	activeDocument, err := m.configurationManager.GetActiveDocument(ctx)
	if err != nil {
		return nil, err
	}

	configuration := configurations.GetProjectDomainConfigurationFromDocument(activeDocument, request.Attributes.Project, request.Attributes.Domain)
	configurationUpdateRequest := transformers.FromProjectDomainAttributesUpdateRequest(&request, configuration, activeDocument.Version)
	_, err = m.configurationManager.UpdateConfiguration(ctx, *configurationUpdateRequest)
	if err != nil {
		return nil, err
	}
	return &admin.ProjectDomainAttributesUpdateResponse{}, nil
}

func (m *ConfigurationResourceManager) DeleteProjectDomainAttributes(
	ctx context.Context, request admin.ProjectDomainAttributesDeleteRequest) (
	*admin.ProjectDomainAttributesDeleteResponse, error) {
	if err := validation.ValidateProjectDomainAttributesDeleteRequest(ctx, m.db, m.config, request); err != nil {
		return nil, err
	}

	activeDocument, err := m.configurationManager.GetActiveDocument(ctx)
	if err != nil {
		return nil, err
	}

	configuration := configurations.GetProjectDomainConfigurationFromDocument(activeDocument, request.Project, request.Domain)
	configurationDeleteRequest := transformers.FromProjectDomainAttributesDeleteRequest(&request, configuration, activeDocument.Version)
	_, err = m.configurationManager.UpdateConfiguration(ctx, *configurationDeleteRequest)
	if err != nil {
		return nil, err
	}
	return &admin.ProjectDomainAttributesDeleteResponse{}, nil
}

func (m *ConfigurationResourceManager) GetWorkflowAttributes(
	ctx context.Context, request admin.WorkflowAttributesGetRequest) (
	*admin.WorkflowAttributesGetResponse, error) {
	if err := validation.ValidateWorkflowAttributesGetRequest(ctx, m.db, m.config, request); err != nil {
		return nil, err
	}

	// Get the active document
	activeDocument, err := m.configurationManager.GetActiveDocument(ctx)
	if err != nil {
		return nil, err
	}

	workflowConfiguration := configurations.GetWorkflowConfigurationFromDocument(activeDocument, request.Project, request.Domain, request.Workflow)
	workflowAttributesGetResponse, err := transformers.ToWorkflowAttributesGetResponse(workflowConfiguration, &request)
	if err != nil {
		return nil, err
	}
	return workflowAttributesGetResponse, nil
}

func (m *ConfigurationResourceManager) UpdateWorkflowAttributes(
	ctx context.Context, request admin.WorkflowAttributesUpdateRequest) (
	*admin.WorkflowAttributesUpdateResponse, error) {
	var err error
	if _, err = validation.ValidateWorkflowAttributesUpdateRequest(ctx, m.db, m.config, request); err != nil {
		return nil, err
	}

	activeDocument, err := m.configurationManager.GetActiveDocument(ctx)
	if err != nil {
		return nil, err
	}

	configuration := configurations.GetWorkflowConfigurationFromDocument(activeDocument, request.Attributes.Project, request.Attributes.Domain, request.Attributes.Workflow)
	configurationUpdateRequest := transformers.FromWorkflowAttributesUpdateRequest(&request, configuration, activeDocument.Version)
	_, err = m.configurationManager.UpdateConfiguration(ctx, *configurationUpdateRequest)
	if err != nil {
		return nil, err
	}
	return &admin.WorkflowAttributesUpdateResponse{}, nil
}

func (m *ConfigurationResourceManager) DeleteWorkflowAttributes(
	ctx context.Context, request admin.WorkflowAttributesDeleteRequest) (
	*admin.WorkflowAttributesDeleteResponse, error) {
	if err := validation.ValidateWorkflowAttributesDeleteRequest(ctx, m.db, m.config, request); err != nil {
		return nil, err
	}

	activeDocument, err := m.configurationManager.GetActiveDocument(ctx)
	if err != nil {
		return nil, err
	}

	configuration := configurations.GetWorkflowConfigurationFromDocument(activeDocument, request.Project, request.Domain, request.Workflow)
	configurationUpdateRequest := transformers.FromWorkflowAttributesDeleteRequest(&request, configuration, activeDocument.Version)
	_, err = m.configurationManager.UpdateConfiguration(ctx, *configurationUpdateRequest)
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
