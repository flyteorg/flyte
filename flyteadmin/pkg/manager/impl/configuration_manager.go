package impl

import (
	"context"
	"crypto/rand"
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
	config        runtimeInterfaces.Configuration
	storageClient *storage.DataStore
	cache         *admin.ConfigurationDocument
}

func (m *ConfigurationManager) GetConfiguration(
	ctx context.Context, request admin.ConfigurationGetRequest) (
	*admin.ConfigurationGetResponse, error) {
	// Validate the request
	if err := validation.ValidateConfigurationGetRequest(request); err != nil {
		return nil, err
	}

	// Get the active override attributes
	activeConfiguration, err := m.db.ConfigurationDocumentRepo().GetActive(ctx)
	if err != nil {
		return nil, err
	}

	// If the cache is not nil and the version of the cache is the same as the active override attributes, return the cache
	if m.cache != nil && m.cache.GetVersion() == activeConfiguration.Version {
		return &admin.ConfigurationGetResponse{
			Id: &admin.ProjectID{
				Project: request.GetId().GetProject(),
				Domain:  request.GetId().GetDomain(),
			},
			Version:                    m.cache.GetVersion(),
			GlobalConfiguration:        getGlobalConfigurationFromDocument(m.cache),
			ProjectConfiguration:       getProjectConfigurationFromDocument(m.cache, request.Id.Project),
			ProjectDomainConfiguration: getProjectDomainConfigurationFromDocument(m.cache, request.Id.Project, request.Id.Domain),
		}, nil
	}

	// TODO: Handle the case where the document location is empty
	document := &admin.ConfigurationDocument{}
	if err := m.storageClient.ReadProtobuf(ctx, activeConfiguration.DocumentLocation, document); err != nil {
		return nil, err
	}

	// Cache the document
	m.cache = document

	// Return the override attributes of different scopes
	return &admin.ConfigurationGetResponse{
		Id: &admin.ProjectID{
			Project: request.GetId().GetProject(),
			Domain:  request.GetId().GetDomain(),
		},
		Version:                    document.GetVersion(),
		GlobalConfiguration:        getGlobalConfigurationFromDocument(m.cache),
		ProjectConfiguration:       getProjectConfigurationFromDocument(m.cache, request.Id.Project),
		ProjectDomainConfiguration: getProjectDomainConfigurationFromDocument(m.cache, request.Id.Project, request.Id.Domain),
	}, nil
}

func (m *ConfigurationManager) UpdateConfiguration(
	ctx context.Context, request admin.ConfigurationUpdateRequest, mergeActive bool) (
	*admin.ConfigurationUpdateResponse, error) {
	// Validate the request
	if err := validation.ValidateConfigurationUpdateRequest(request); err != nil {
		return nil, err
	}

	// Get the active override attributes
	activeConfigurationDocument, err := m.db.ConfigurationDocumentRepo().GetActive(ctx)
	if err != nil {
		return nil, err
	}

	// TODO: Handle the case where the document location is empty
	// TODO: Use cache
	document := &admin.ConfigurationDocument{}
	if err := m.storageClient.ReadProtobuf(ctx, activeConfigurationDocument.DocumentLocation, document); err != nil {
		return nil, err
	}

	// If mergeActive is true, merge the incoming configuration with the active configuration
	if mergeActive {
		currentConfiguration := getConfigurationFromDocument(document, request.Id.Project, request.Id.Domain, "", "")
		if currentConfiguration == nil {
			currentConfiguration = &admin.Configuration{}
		}
		mergedConfiguration := mergeConfiguration(request.Configuration, currentConfiguration)
		request.Configuration = mergedConfiguration
	}

	// Update the document with the new attributes
	updateConfigurationToDocument(document, request.Id.Project, request.Id.Domain, "", "", request.Configuration)

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

	newConfiguration := models.ConfigurationDocument{
		// random generate a string as version
		Version:          generatedVersion,
		Active:           true,
		DocumentLocation: updatedDocumentLocation,
	}

	// Erase the active override attributes and create the new one
	var versionToUpdate string
	if request.VersionToUpdate != "" {
		versionToUpdate = request.VersionToUpdate
	} else {
		versionToUpdate = activeConfigurationDocument.Version
	}
	input := &repositoryInterfaces.UpdateConfigurationInput{
		VersionToUpdate:  versionToUpdate,
		NewConfiguration: &newConfiguration,
	}
	if err := m.db.ConfigurationDocumentRepo().Update(ctx, input); err != nil {
		return nil, err
	}
	return &admin.ConfigurationUpdateResponse{}, nil
}

// Get configuration document for update
func (m *ConfigurationManager) GetUpdateInput(ctx context.Context, inputConfiguration *admin.Configuration, project, domain, workflow, launchPlan string) (*repositoryInterfaces.UpdateConfigurationInput, error) {
	// Get the active configuration document
	activeConfigurationDocumentModel, err := m.db.ConfigurationDocumentRepo().GetActive(ctx)
	if err != nil {
		return nil, err
	}
	var activeConfigurationDocument *admin.ConfigurationDocument
	// If the cache is not nil and the version of the cache is the same as the active override attributes, use the cache
	if m.cache != nil && m.cache.GetVersion() == activeConfigurationDocument.Version {
		activeConfigurationDocument = m.cache
	} else {
		// TODO: Handle the case where the document location is empty
		if err := m.storageClient.ReadProtobuf(ctx, activeConfigurationDocumentModel.DocumentLocation, activeConfigurationDocument); err != nil {
			return nil, err
		}
	}

	currentConfiguration := getConfigurationFromDocument(activeConfigurationDocument, project, domain, workflow, launchPlan)
	if currentConfiguration == nil {
		currentConfiguration = &admin.Configuration{}
	}
	mergedConfiguration := mergeConfiguration(inputConfiguration, currentConfiguration)

	// Update the document with the new attributes
	updateConfigurationToDocument(activeConfigurationDocument, project, domain, workflow, launchPlan, mergedConfiguration)

	// Generate a new version for the document
	generatedVersion, err := GenerateRandomString(10)
	if err != nil {
		return nil, err
	}
	activeConfigurationDocument.Version = generatedVersion

	// Offload the updated document
	updatedDocumentLocation, err := common.OffloadConfigurationDocument(ctx, m.storageClient, activeConfigurationDocument, generatedVersion)

	if err != nil {
		return nil, err
	}

	newConfiguration := models.ConfigurationDocument{
		// random generate a string as version
		Version:          generatedVersion,
		Active:           true,
		DocumentLocation: updatedDocumentLocation,
	}

	return &repositoryInterfaces.UpdateConfigurationInput{
		VersionToUpdate:  activeConfigurationDocument.Version,
		NewConfiguration: &newConfiguration,
	}, nil
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
	documentKey := encodeDocumentKey(project, domain, "", "")
	document.Configurations[documentKey] = attributes
}

// This function is used to encode the document key based on the org, project, domain, workflow, and launch plan.
func encodeDocumentKey(project, domain, workflow, launchPlan string) string {
	// TODO: This is a temporary solution to encode the document key. We need to come up with a better solution.
	return project + "/" + domain + "/" + workflow + "/" + launchPlan
}

func NewConfigurationManager(db repositoryInterfaces.Repository, config runtimeInterfaces.Configuration, storageClient *storage.DataStore) interfaces.ConfigurationInterface {
	return &ConfigurationManager{
		db:            db,
		config:        config,
		storageClient: storageClient,
		cache:         nil,
	}
}
