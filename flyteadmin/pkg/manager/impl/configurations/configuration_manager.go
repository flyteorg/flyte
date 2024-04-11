package configurations

import (
	"context"

	"github.com/flyteorg/flyte/flyteadmin/pkg/common"
	"github.com/flyteorg/flyte/flyteadmin/pkg/errors"
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
	config        runtimeInterfaces.Configuration
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
		GlobalConfiguration: GetGlobalConfigurationFromDocument(document),
	}
	if id.Project != "" {
		response.ProjectConfiguration = GetProjectConfigurationFromDocument(document, id.Project)
	}
	if id.Domain != "" {
		response.ProjectDomainConfiguration = GetProjectDomainConfigurationFromDocument(document, id.Project, id.Domain)
	}
	if id.Workflow != "" {
		response.WorkflowConfiguration = GetWorkflowConfigurationFromDocument(document, id.Project, id.Domain, id.Workflow)
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
	ctx context.Context, request admin.ConfigurationUpdateRequest) (
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

	// Update the document with the new attributes
	updateConfigurationToDocument(document, request.Configuration, request.Id.Project, request.Id.Domain, request.Id.Workflow)

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
		VersionToUpdate:          request.VersionToUpdate,
		NewConfigurationDocument: &newDocument,
	}
	if err := m.db.ConfigurationDocumentRepo().Update(ctx, input); err != nil {
		return nil, err
	}
	return &admin.ConfigurationUpdateResponse{}, nil
}

func collectGlobalConfiguration(config runtimeInterfaces.Configuration) *admin.Configuration {
	// Task resource attributes
	taskResourceAttributesConfig := config.TaskResourceConfiguration()
	defaultCPU := taskResourceAttributesConfig.GetDefaults().CPU
	defaultGPU := taskResourceAttributesConfig.GetDefaults().GPU
	defaultMemory := taskResourceAttributesConfig.GetDefaults().Memory
	defaultEphemeralStorage := taskResourceAttributesConfig.GetDefaults().EphemeralStorage
	limitCPU := taskResourceAttributesConfig.GetLimits().CPU
	limitGPU := taskResourceAttributesConfig.GetLimits().GPU
	limitMemory := taskResourceAttributesConfig.GetLimits().Memory
	limitEphemeralStorage := taskResourceAttributesConfig.GetLimits().EphemeralStorage
	taskResourceAttributes := admin.TaskResourceAttributes{
		Defaults: &admin.TaskResourceSpec{
			Cpu:              defaultCPU.String(),
			Gpu:              defaultGPU.String(),
			Memory:           defaultMemory.String(),
			EphemeralStorage: defaultEphemeralStorage.String(),
		},
		Limits: &admin.TaskResourceSpec{
			Cpu:              limitCPU.String(),
			Gpu:              limitGPU.String(),
			Memory:           limitMemory.String(),
			EphemeralStorage: limitEphemeralStorage.String(),
		},
	}

	// Workflow execution configuration
	topLevelConfig := config.ApplicationConfiguration().GetTopLevelConfig()
	workflowExecutionConfig := admin.WorkflowExecutionConfig{
		MaxParallelism:      topLevelConfig.GetMaxParallelism(),
		SecurityContext:     topLevelConfig.GetSecurityContext(),
		RawOutputDataConfig: topLevelConfig.GetRawOutputDataConfig(),
		Labels:              topLevelConfig.GetLabels(),
		Annotations:         topLevelConfig.GetAnnotations(),
		Interruptible:       topLevelConfig.GetInterruptible(),
		OverwriteCache:      topLevelConfig.GetOverwriteCache(),
		Envs:                topLevelConfig.GetEnvs(),
	}
	executionClusterLabel := admin.ExecutionClusterLabel{
		Value: config.ClusterConfiguration().GetDefaultExecutionLabel(),
	}
	return &admin.Configuration{
		TaskResourceAttributes:    &taskResourceAttributes,
		ClusterResourceAttributes: nil,
		ExecutionQueueAttributes:  nil,
		ExecutionClusterLabel:     &executionClusterLabel,
		QualityOfService:          nil,
		PluginOverrides:           nil,
		WorkflowExecutionConfig:   &workflowExecutionConfig,
		ClusterAssignment:         nil,
	}
}

func NewConfigurationManager(db repositoryInterfaces.Repository, config runtimeInterfaces.Configuration, storageClient *storage.DataStore) interfaces.ConfigurationInterface {
	configurationManager := &ConfigurationManager{
		db:            db,
		config:        config,
		storageClient: storageClient,
		cache:         nil,
	}
	ctx := context.Background()
	activeDocument, err := configurationManager.GetActiveDocument(ctx)
	if err != nil && errors.IsDoesNotExistError(err) {
		generatedVersion, err := GenerateRandomString(10)
		if err != nil {
			panic(err)
		}
		document := &admin.ConfigurationDocument{
			Version:        generatedVersion,
			Configurations: make(map[string]*admin.Configuration),
		}
		updateConfigurationToDocument(document, collectGlobalConfiguration(config), "", "", "")

		updatedDocumentLocation, err := common.OffloadConfigurationDocument(ctx, configurationManager.storageClient, document, generatedVersion)
		if err != nil {
			panic(err)
		}
		err = configurationManager.db.ConfigurationDocumentRepo().Create(ctx, &models.ConfigurationDocument{
			Version:          generatedVersion,
			Active:           true,
			DocumentLocation: updatedDocumentLocation,
		})
		if err != nil {
			panic(err)
		}
		return configurationManager
	}
	if err != nil {
		panic(err)
	}
	_, err = configurationManager.UpdateConfiguration(ctx, admin.ConfigurationUpdateRequest{
		Id:              &admin.ConfigurationID{},
		VersionToUpdate: activeDocument.GetVersion(),
		Configuration:   collectGlobalConfiguration(config),
	})
	if err != nil {
		panic(err)
	}
	return configurationManager
}
