package impl

import (
	"context"
	"crypto/rand"
	"github.com/flyteorg/flyte/flyteadmin/pkg/repositories/transformers"
	"math/big"

	"github.com/flyteorg/flyte/flyteadmin/pkg/common"
	"github.com/flyteorg/flyte/flyteadmin/pkg/manager/impl/validation"
	"github.com/flyteorg/flyte/flyteadmin/pkg/manager/interfaces"
	repositoryInterfaces "github.com/flyteorg/flyte/flyteadmin/pkg/repositories/interfaces"
	"github.com/flyteorg/flyte/flyteadmin/pkg/repositories/models"
	runtimeInterfaces "github.com/flyteorg/flyte/flyteadmin/pkg/runtime/interfaces"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyte/flytestdlib/storage"
)

type ConfigurationManager struct {
	db            repositoryInterfaces.Repository
	config        runtimeInterfaces.ApplicationConfiguration
	storageClient *storage.DataStore
	cache         *admin.ConfigurationDocument
}

func prepareConfigurationGetResponse(document *admin.ConfigurationDocument, id *admin.ConfigurationID) *admin.ConfigurationGetResponse {
	response := &admin.ConfigurationGetResponse{
		Id: &admin.ConfigurationID{
			Project: id.GetProject(),
			Domain:  id.GetDomain(),
		},
		Version:             document.GetVersion(),
		GlobalConfiguration: getGlobalConfigurationFromDocument(document),
	}
	if id.Project != "" {
		response.ProjectConfiguration = getProjectConfigurationFromDocument(document, id.Project)
	}
	if id.Domain != "" {
		response.ProjectDomainConfiguration = getProjectDomainConfigurationFromDocument(document, id.Project, id.Domain)
	}
	if id.Workflow != "" {
		response.WorkflowConfiguration = getWorkflowConfigurationFromDocument(document, id.Project, id.Domain, id.Workflow)
	}
	return response
}

func (m *ConfigurationManager) GetActiveDocument(ctx context.Context) (*admin.ConfigurationDocument, error) {
	// Get the active configuration
	activeConfiguration, err := m.db.ConfigurationDocumentRepo().GetActive(ctx)
	if err != nil {
		return nil, err
	}

	// If the cache is not nil and the version of the cache is the same as the active override attributes, return the cache
	if m.cache != nil && m.cache.GetVersion() == activeConfiguration.Version {
		return m.cache, nil
	}

	// TODO: Handle the case where the document location is empty
	document := &admin.ConfigurationDocument{}
	if err := m.storageClient.ReadProtobuf(ctx, activeConfiguration.DocumentLocation, document); err != nil {
		return nil, err
	}

	// Cache the document
	m.cache = document

	return document, nil
}

func (m *ConfigurationManager) GetConfiguration(
	ctx context.Context, request admin.ConfigurationGetRequest) (
	*admin.ConfigurationGetResponse, error) {
	// Validate the request
	if err := validation.ValidateConfigurationGetRequest(request); err != nil {
		return nil, err
	}

	// Get the active document
	activeDocument, err := m.GetActiveDocument(ctx)
	if err != nil {
		return nil, err
	}

	// Return the configurations of different scopes
	response := prepareConfigurationGetResponse(activeDocument, request.GetId())
	return response, nil
}

func (m *ConfigurationManager) UpdateConfiguration(
	ctx context.Context, request admin.ConfigurationUpdateRequest, mergeActive bool) (
	*admin.ConfigurationUpdateResponse, error) {
	// Validate the request
	if err := validation.ValidateConfigurationUpdateRequest(request); err != nil {
		return nil, err
	}

	// Get the active override attributes
	document, err := m.GetActiveDocument(ctx)
	if err != nil {
		return nil, err
	}

	// Set the version to update
	var versionToUpdate string
	if request.VersionToUpdate != "" {
		versionToUpdate = request.VersionToUpdate
	} else {
		versionToUpdate = document.Version
	}

	currentConfiguration := getConfigurationFromDocumentWithID(document, request.Id)

	// If mergeActive is true, merge the incoming configuration with the active configuration
	if mergeActive {
		mergedConfiguration := mergeConfiguration(request.Configuration, currentConfiguration)
		request.Configuration = mergedConfiguration
	}

	// Update the document with the new attributes
	updateConfigurationToDocument(document, request.Id.Project, request.Id.Domain, request.Id.Workflow, "", request.Configuration)

	// Generate a new version for the document
	generatedVersion, err := GenerateRandomString(10)
	if err != nil {
		return nil, err
	}
	document.Version = generatedVersion

	// Offload the updated document
	updatedDocumentLocation, err := common.OffloadConfigurationDocument(ctx, m.storageClient, document, generatedVersion)

	if err != nil {
		return nil, err
	}

	newDocument := models.ConfigurationDocument{
		// random generate a string as version
		Version:          generatedVersion,
		Active:           true,
		DocumentLocation: updatedDocumentLocation,
	}

	// Erase the active override attributes and create the new one
	input := &repositoryInterfaces.UpdateConfigurationInput{
		VersionToUpdate:  versionToUpdate,
		NewConfiguration: &newDocument,
	}
	if err := m.db.ConfigurationDocumentRepo().Update(ctx, input); err != nil {
		return nil, err
	}
	return &admin.ConfigurationUpdateResponse{}, nil
}

// For backward compatibility
func (m *ConfigurationManager) GetProjectAttributes(
	ctx context.Context, request admin.ProjectAttributesGetRequest) (
	*admin.ProjectAttributesGetResponse, error) {
	if err := validation.ValidateProjectAttributesGetRequest(ctx, m.db, request); err != nil {
		return nil, err
	}
	// Get the active document
	activeDocument, err := m.GetActiveDocument(ctx)
	if err != nil {
		return nil, err
	}

	projectConfiguration := getProjectConfigurationFromDocument(activeDocument, request.Project)
	projectAttributesGetResponse, err := transformers.ToProjectAttributesGetResponse(projectConfiguration, &request)
	if err != nil {
		return nil, err
	}
	return projectAttributesGetResponse, nil
}

func (m *ConfigurationManager) UpdateProjectAttributes(
	ctx context.Context, request admin.ProjectAttributesUpdateRequest) (
	*admin.ProjectAttributesUpdateResponse, error) {
	var err error
	if _, err = validation.ValidateProjectAttributesUpdateRequest(ctx, m.db, request); err != nil {
		return nil, err
	}
	configurationUpdateRequest := transformers.FromProjectAttributesUpdateRequest(&request)
	_, err = m.UpdateConfiguration(ctx, *configurationUpdateRequest, true)
	if err != nil {
		return nil, err
	}
	return &admin.ProjectAttributesUpdateResponse{}, nil
}

func (m *ConfigurationManager) DeleteProjectAttributes(
	ctx context.Context, request admin.ProjectAttributesDeleteRequest) (
	*admin.ProjectAttributesDeleteResponse, error) {
	if err := validation.ValidateProjectAttributesDeleteRequest(ctx, m.db, request); err != nil {
		return nil, err
	}
	configurationUpdateRequest := transformers.FromProjectAttributesDeleteRequest(&request)
	_, err := m.UpdateConfiguration(ctx, *configurationUpdateRequest, true)
	if err != nil {
		return nil, err
	}
	return &admin.ProjectAttributesDeleteResponse{}, nil
}

func (m *ConfigurationManager) GetProjectDomainAttributes(
	ctx context.Context, request admin.ProjectDomainAttributesGetRequest) (
	*admin.ProjectDomainAttributesGetResponse, error) {
	if err := validation.ValidateProjectDomainAttributesGetRequest(ctx, m.db, m.config, request); err != nil {
		return nil, err
	}
	// Get the active document
	activeDocument, err := m.GetActiveDocument(ctx)
	if err != nil {
		return nil, err
	}

	projectDomainConfiguration := getProjectDomainConfigurationFromDocument(activeDocument, request.Project, request.Domain)
	projectDomainAttributesGetResponse, err := transformers.ToProjectDomainAttributesGetResponse(projectDomainConfiguration, &request)
	if err != nil {
		return nil, err
	}
	return projectDomainAttributesGetResponse, nil
}

func (m *ConfigurationManager) UpdateProjectDomainAttributes(
	ctx context.Context, request admin.ProjectDomainAttributesUpdateRequest) (
	*admin.ProjectDomainAttributesUpdateResponse, error) {
	var err error
	if _, err = validation.ValidateProjectDomainAttributesUpdateRequest(ctx, m.db, m.config, request); err != nil {
		return nil, err
	}
	configurationUpdateRequest := transformers.FromProjectDomainAttributesUpdateRequest(&request)
	_, err = m.UpdateConfiguration(ctx, *configurationUpdateRequest, true)
	if err != nil {
		return nil, err
	}
	return &admin.ProjectDomainAttributesUpdateResponse{}, nil
}

func (m *ConfigurationManager) DeleteProjectDomainAttributes(
	ctx context.Context, request admin.ProjectDomainAttributesDeleteRequest) (
	*admin.ProjectDomainAttributesDeleteResponse, error) {
	if err := validation.ValidateProjectDomainAttributesDeleteRequest(ctx, m.db, m.config, request); err != nil {
		return nil, err
	}
	configurationDeleteRequest := transformers.FromProjectDomainAttributesDeleteRequest(&request)
	_, err := m.UpdateConfiguration(ctx, *configurationDeleteRequest, true)
	if err != nil {
		return nil, err
	}
	return &admin.ProjectDomainAttributesDeleteResponse{}, nil
}

func (m *ConfigurationManager) GetWorkflowAttributes(
	ctx context.Context, request admin.WorkflowAttributesGetRequest) (
	*admin.WorkflowAttributesGetResponse, error) {
	if err := validation.ValidateWorkflowAttributesGetRequest(ctx, m.db, m.config, request); err != nil {
		return nil, err
	}

	// Get the active document
	activeDocument, err := m.GetActiveDocument(ctx)
	if err != nil {
		return nil, err
	}

	workflowConfiguration := getWorkflowConfigurationFromDocument(activeDocument, request.Project, request.Domain, request.Workflow)
	workflowAttributesGetResponse, err := transformers.ToWorkflowAttributesGetResponse(workflowConfiguration, &request)
	if err != nil {
		return nil, err
	}
	return workflowAttributesGetResponse, nil
}

func (m *ConfigurationManager) UpdateWorkflowAttributes(
	ctx context.Context, request admin.WorkflowAttributesUpdateRequest) (
	*admin.WorkflowAttributesUpdateResponse, error) {
	var err error
	if _, err = validation.ValidateWorkflowAttributesUpdateRequest(ctx, m.db, m.config, request); err != nil {
		return nil, err
	}
	configurationUpdateRequest := transformers.FromWorkflowAttributesUpdateRequest(&request)
	_, err = m.UpdateConfiguration(ctx, *configurationUpdateRequest, true)
	if err != nil {
		return nil, err
	}
	return &admin.WorkflowAttributesUpdateResponse{}, nil
}

func (m *ConfigurationManager) DeleteWorkflowAttributes(
	ctx context.Context, request admin.WorkflowAttributesDeleteRequest) (
	*admin.WorkflowAttributesDeleteResponse, error) {
	if err := validation.ValidateWorkflowAttributesDeleteRequest(ctx, m.db, m.config, request); err != nil {
		return nil, err
	}
	configurationUpdateRequest := transformers.FromWorkflowAttributesDeleteRequest(&request)
	_, err := m.UpdateConfiguration(ctx, *configurationUpdateRequest, true)
	if err != nil {
		return nil, err
	}
	return &admin.WorkflowAttributesDeleteResponse{}, nil
}

// Merge two configurations
func mergeConfiguration(incomingConfiguration, currentConfiguration *admin.Configuration) *admin.Configuration {
	var mergedConfiguration admin.Configuration
	// task resource attributes
	if incomingConfiguration.GetTaskResourceAttributes() != nil {
		mergedConfiguration.TaskResourceAttributes = incomingConfiguration.GetTaskResourceAttributes()
	} else {
		mergedConfiguration.TaskResourceAttributes = currentConfiguration.GetTaskResourceAttributes()
	}
	// cluster resource attributes
	if incomingConfiguration.GetClusterResourceAttributes() != nil {
		mergedConfiguration.ClusterResourceAttributes = incomingConfiguration.GetClusterResourceAttributes()
	} else {
		mergedConfiguration.ClusterResourceAttributes = currentConfiguration.GetClusterResourceAttributes()
	}
	// execution queue attributes
	if incomingConfiguration.GetExecutionQueueAttributes() != nil {
		mergedConfiguration.ExecutionQueueAttributes = incomingConfiguration.GetExecutionQueueAttributes()
	} else {
		mergedConfiguration.ExecutionQueueAttributes = currentConfiguration.GetExecutionQueueAttributes()
	}
	// execution cluster label
	if incomingConfiguration.GetExecutionClusterLabel() != nil {
		mergedConfiguration.ExecutionClusterLabel = incomingConfiguration.GetExecutionClusterLabel()
	} else {
		mergedConfiguration.ExecutionClusterLabel = currentConfiguration.GetExecutionClusterLabel()
	}
	// quality of service
	if incomingConfiguration.GetQualityOfService() != nil {
		mergedConfiguration.QualityOfService = incomingConfiguration.GetQualityOfService()
	} else {
		mergedConfiguration.QualityOfService = currentConfiguration.GetQualityOfService()
	}
	// plugin overrides
	if incomingConfiguration.GetPluginOverrides() != nil {
		mergedConfiguration.PluginOverrides = incomingConfiguration.GetPluginOverrides()
	} else {
		mergedConfiguration.PluginOverrides = currentConfiguration.GetPluginOverrides()
	}
	// workflow execution config
	if incomingConfiguration.GetWorkflowExecutionConfig() != nil {
		mergedConfiguration.WorkflowExecutionConfig = incomingConfiguration.GetWorkflowExecutionConfig()
	} else {
		mergedConfiguration.WorkflowExecutionConfig = currentConfiguration.GetWorkflowExecutionConfig()
	}
	// cluster assignment
	if incomingConfiguration.GetClusterAssignment() != nil {
		mergedConfiguration.ClusterAssignment = incomingConfiguration.GetClusterAssignment()
	} else {
		mergedConfiguration.ClusterAssignment = currentConfiguration.GetClusterAssignment()
	}
	return &mergedConfiguration
}

func getGlobalConfigurationFromDocument(document *admin.ConfigurationDocument) *admin.Configuration {
	return getConfigurationFromDocument(document, "", "", "", "")
}

func getProjectConfigurationFromDocument(document *admin.ConfigurationDocument, project string) *admin.Configuration {
	return getConfigurationFromDocument(document, project, "", "", "")
}

func getProjectDomainConfigurationFromDocument(document *admin.ConfigurationDocument, project, domain string) *admin.Configuration {
	return getConfigurationFromDocument(document, project, domain, "", "")
}

func getWorkflowConfigurationFromDocument(document *admin.ConfigurationDocument, project, domain, workflow string) *admin.Configuration {
	return getConfigurationFromDocument(document, project, domain, workflow, "")
}

func getConfigurationFromDocumentWithID(document *admin.ConfigurationDocument, id *admin.ConfigurationID) *admin.Configuration {
	return getConfigurationFromDocument(document, id.Project, id.Domain, id.Workflow, "")
}

// TODO: Check if this function is implemented already
func GenerateRandomString(length int) (string, error) {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	var result string
	charsetLength := big.NewInt(int64(len(charset)))
	for i := 0; i < length; i++ {
		randomNumber, err := rand.Int(rand.Reader, charsetLength)
		if err != nil {
			return "", err // Return the error if there was a problem generating the random number.
		}
		result += string(charset[randomNumber.Int64()])
	}
	return result, nil
}

func getConfigurationFromGetResponse(response *admin.ConfigurationGetResponse, id *admin.ConfigurationID) *admin.Configuration {
	if id.Workflow != "" {
		return response.WorkflowConfiguration
	}
	if id.Domain != "" {
		return response.ProjectDomainConfiguration
	}
	if id.Project != "" {
		return response.ProjectConfiguration
	}
	return response.GlobalConfiguration
}

// This function is used to get the attributes of a document based on the project, and domain.
func getConfigurationFromDocument(document *admin.ConfigurationDocument, project, domain, workflow, launchPlan string) *admin.Configuration {
	documentKey := encodeDocumentKey(project, domain, workflow, launchPlan)
	if attributes, ok := document.Configurations[documentKey]; ok {
		return attributes
	}
	return nil
}

// This function is used to update the attributes of a document based on the project, and domain.
func updateConfigurationToDocument(document *admin.ConfigurationDocument, project, domain, workflow, launchPlan string, attributes *admin.Configuration) {
	documentKey := encodeDocumentKey(project, domain, workflow, launchPlan)
	document.Configurations[documentKey] = attributes
}

// This function is used to encode the document key based on the org, project, domain, workflow, and launch plan.
func encodeDocumentKey(project, domain, workflow, launchPlan string) string {
	// TODO: This is a temporary solution to encode the document key. We need to come up with a better solution.
	return project + "/" + domain + "/" + workflow + "/" + launchPlan
}

func NewConfigurationManager(db repositoryInterfaces.Repository, config runtimeInterfaces.ApplicationConfiguration, storageClient *storage.DataStore) interfaces.ConfigurationInterface {
	return &ConfigurationManager{
		db:            db,
		config:        config,
		storageClient: storageClient,
		cache:         nil,
	}
}
