package configurations

import (
	"context"
	"sync"

	"google.golang.org/grpc/codes"

	"github.com/flyteorg/flyte/flyteadmin/pkg/common"
	"github.com/flyteorg/flyte/flyteadmin/pkg/errors"
	"github.com/flyteorg/flyte/flyteadmin/pkg/manager/impl/configurations/plugin"
	"github.com/flyteorg/flyte/flyteadmin/pkg/manager/impl/util"
	"github.com/flyteorg/flyte/flyteadmin/pkg/manager/impl/validation"
	"github.com/flyteorg/flyte/flyteadmin/pkg/manager/interfaces"
	repositoryInterfaces "github.com/flyteorg/flyte/flyteadmin/pkg/repositories/interfaces"
	"github.com/flyteorg/flyte/flyteadmin/pkg/repositories/models"
	runtimeInterfaces "github.com/flyteorg/flyte/flyteadmin/pkg/runtime/interfaces"
	"github.com/flyteorg/flyte/flyteadmin/plugins"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyte/flytestdlib/logger"
	"github.com/flyteorg/flyte/flytestdlib/storage"
)

// ConfigurationManager is responsible for managing the project configuration, which includes attributes such as task resource attributes.
type ConfigurationManager struct {
	db             repositoryInterfaces.Repository
	config         runtimeInterfaces.Configuration
	storageClient  *storage.DataStore
	cacheConfigDoc *configurationDocumentCache
	pluginRegistry *plugins.Registry
}

const (
	ShouldBootstrapOrUpdateDefault    = true
	ShouldNotBootstrapOrUpdateDefault = false
)

type documentMode int

const (
	editable documentMode = iota
	readOnly
)

func (m *ConfigurationManager) GetEditableActiveDocument(ctx context.Context) (admin.ConfigurationDocument, error) {
	return m.getActiveDocument(ctx, editable)
}

func (m *ConfigurationManager) GetReadOnlyActiveDocument(ctx context.Context) (admin.ConfigurationDocument, error) {
	return m.getActiveDocument(ctx, readOnly)
}

// Note: This function should not be called directly. Use GetReadOnlyActiveDocument or GetEditableActiveDocument instead.
func (m *ConfigurationManager) getActiveDocument(ctx context.Context, mode documentMode) (admin.ConfigurationDocument, error) {
	// Get the active configuration
	activeMetadata, err := m.db.ConfigurationRepo().GetActive(ctx)
	if err != nil {
		logger.Errorf(ctx, "failed to get active configuration document: %v", err)
		return admin.ConfigurationDocument{}, err
	}

	// If the cache is valid, return the cached document
	cacheDoc, ok := m.getCache(activeMetadata.Version, mode)
	if ok {
		logger.Debugf(ctx, "returning cached configuration document")
		return *cacheDoc, nil
	}

	// Get the active document from the storage
	activeDocument := &admin.ConfigurationDocument{}
	if err := m.storageClient.ReadProtobuf(ctx, activeMetadata.DocumentLocation, activeDocument); err != nil {
		logger.Errorf(ctx, "failed to read configuration document: %v", err)
		return admin.ConfigurationDocument{}, err
	}

	// Since an empty map will be serialized as nil, we need to initialize the map if it is nil
	if activeDocument.Configurations == nil {
		activeDocument.Configurations = make(map[string]*admin.Configuration)
	}

	// Set the cache
	m.setCache(activeDocument)

	return *activeDocument, nil
}

func (m *ConfigurationManager) GetConfiguration(
	ctx context.Context, request admin.ConfigurationGetRequest) (
	*admin.ConfigurationGetResponse, error) {
	// Validate the request
	if err := validation.ValidateConfigurationGetRequest(request); err != nil {
		return nil, err
	}
	if request.Id.Domain != "" && request.Id.Project == "" && request.Id.Workflow == "" {
		logger.Debugf(ctx, "getting default configuration")
		return m.getDefaultConfiguration(ctx, request)
	}
	logger.Debugf(ctx, "getting project domain configuration")
	return m.getProjectDomainConfiguration(ctx, request)
}

func (m *ConfigurationManager) getProjectDomainConfiguration(
	ctx context.Context, request admin.ConfigurationGetRequest) (
	*admin.ConfigurationGetResponse, error) {
	// Validate the request
	if err := validation.ValidateProjectDomainConfigurationGetRequest(ctx, m.db, m.config.ApplicationConfiguration(), request); err != nil {
		return nil, err
	}

	// Get the active document
	activeDocument, err := m.GetReadOnlyActiveDocument(ctx)
	if err != nil {
		return nil, err
	}

	// Get the configuration with source
	configurationWithSource, err := util.GetConfigurationWithSource(ctx, &activeDocument, request.GetId(), plugins.Get[plugin.ProjectConfigurationPlugin](m.pluginRegistry, plugins.PluginIDProjectConfiguration))
	if err != nil {
		return nil, err
	}

	response := &admin.ConfigurationGetResponse{
		Id:            request.GetId(),
		Version:       activeDocument.GetVersion(),
		Configuration: configurationWithSource,
	}
	return response, nil
}

func (m *ConfigurationManager) getDefaultConfiguration(
	ctx context.Context, request admin.ConfigurationGetRequest) (
	*admin.ConfigurationGetResponse, error) {
	// Validate the request
	if err := validation.ValidateDefaultConfigurationGetRequest(ctx, m.config.ApplicationConfiguration(), request); err != nil {
		return nil, err
	}

	// Get the active document
	activeDocument, err := m.GetReadOnlyActiveDocument(ctx)
	if err != nil {
		return nil, err
	}

	defaultConfigurationWithSource, err := util.GetDefaultConfigurationWithSource(ctx, &activeDocument, request.GetId().GetDomain(), plugins.Get[plugin.ProjectConfigurationPlugin](m.pluginRegistry, plugins.PluginIDProjectConfiguration))
	if err != nil {
		return nil, err
	}

	response := &admin.ConfigurationGetResponse{
		Id:            request.GetId(),
		Version:       activeDocument.GetVersion(),
		Configuration: defaultConfigurationWithSource,
	}
	return response, nil
}

func (m *ConfigurationManager) UpdateWorkflowConfiguration(
	ctx context.Context, request admin.ConfigurationUpdateRequest) (
	*admin.ConfigurationUpdateResponse, error) {
	if err := validation.ValidateWorkflowConfigurationUpdateRequest(ctx, m.db, m.config.ApplicationConfiguration(), request); err != nil {
		return nil, err
	}

	return m.updateConfiguration(ctx, request)
}
func (m *ConfigurationManager) UpdateProjectDomainConfiguration(
	ctx context.Context, request admin.ConfigurationUpdateRequest) (
	*admin.ConfigurationUpdateResponse, error) {
	if err := validation.ValidateProjectDomainConfigurationUpdateRequest(ctx, m.db, m.config.ApplicationConfiguration(), request); err != nil {
		return nil, err
	}

	return m.updateConfiguration(ctx, request)
}
func (m *ConfigurationManager) UpdateProjectConfiguration(
	ctx context.Context, request admin.ConfigurationUpdateRequest) (
	*admin.ConfigurationUpdateResponse, error) {
	if err := validation.ValidateProjectConfigurationUpdateRequest(ctx, m.db, request); err != nil {
		return nil, err
	}

	return m.updateConfiguration(ctx, request)
}

func (m *ConfigurationManager) updateConfiguration(
	ctx context.Context, request admin.ConfigurationUpdateRequest) (
	*admin.ConfigurationUpdateResponse, error) {
	// Validate the request
	if err := validation.ValidateConfigurationUpdateRequest(request); err != nil {
		return nil, err
	}

	// Check if the configuration is editing non-mutable attributes
	projectConfigurationPlugin := plugins.Get[plugin.ProjectConfigurationPlugin](m.pluginRegistry, plugins.PluginIDProjectConfiguration)
	if err := m.canEditAttributes(ctx, request, projectConfigurationPlugin); err != nil {
		return nil, err
	}

	// Get the active document
	activeDocument, err := m.GetEditableActiveDocument(ctx)
	if err != nil {
		return nil, err
	}

	// Fail fast if version to update is not the same as the active version
	if activeDocument.GetVersion() != request.GetVersionToUpdate() {
		logger.Errorf(ctx, "version to update %s is not the same as the active version %s", request.GetVersionToUpdate(), activeDocument.GetVersion())
		return nil, errors.ConfigurationDocumentStaleError
	}

	// Update the document with the new configuration
	updatedDocument, err := util.UpdateConfigurationToDocument(ctx, &activeDocument, request.Configuration, request.Id)
	if err != nil {
		return nil, err
	}

	// Get the metadata for the updated document
	metadata, err := m.getDocumentMetadata(ctx, updatedDocument)
	if err != nil {
		return nil, err
	}

	// Update the metadata to repo
	err = m.updateDocumentMetadata(ctx, updatedDocument, metadata, request.GetVersionToUpdate())
	if err != nil {
		return nil, err
	}

	// Get the configuration with source
	configurationWithSource, err := util.GetConfigurationWithSource(ctx, updatedDocument, request.GetId(), projectConfigurationPlugin)
	if err != nil {
		return nil, err
	}

	return &admin.ConfigurationUpdateResponse{
		Id:            request.GetId(),
		Version:       metadata.Version,
		Configuration: configurationWithSource,
	}, nil
}

// Note: this function will update the version of a document.
func (m *ConfigurationManager) getDocumentMetadata(ctx context.Context, document *admin.ConfigurationDocument) (*models.ConfigurationDocumentMetadata, error) {
	// Use string digest as version
	version, err := util.GetConfigurationDocumentStringDigest(ctx, *document)
	if err != nil {
		logger.Errorf(ctx, "failed to generate version for document %+v:, error: %v", document.Configurations, err)
		return nil, err
	}
	document.Version = version
	logger.Debugf(ctx, "generated version %s for document %+v", document.Version, document)

	// Offload the updated document
	documentLocation, err := common.OffloadConfigurationDocument(ctx, m.storageClient, document, document.Version)
	if err != nil {
		logger.Errorf(ctx, "failed to offload configuration document: %+v, version: %s, error: %v", document, document.Version, err)
		return nil, err
	}

	return &models.ConfigurationDocumentMetadata{
		Version:          document.Version,
		Active:           true,
		DocumentLocation: documentLocation,
	}, nil
}

func (m *ConfigurationManager) createDocumentMetadata(ctx context.Context, document *admin.ConfigurationDocument, metadata *models.ConfigurationDocumentMetadata) error {
	err := m.db.ConfigurationRepo().Create(ctx, metadata)
	if err != nil {
		logger.Errorf(ctx, "failed to create configuration document metadata: %+v, error: %v", metadata, err)
		return err
	}
	m.setCache(document)
	return nil
}

func (m *ConfigurationManager) updateDocumentMetadata(ctx context.Context, document *admin.ConfigurationDocument, metadata *models.ConfigurationDocumentMetadata, versionToUpdate string) error {
	err := m.db.ConfigurationRepo().Update(ctx, &repositoryInterfaces.UpdateConfigurationInput{
		VersionToUpdate:          versionToUpdate,
		NewConfigurationMetadata: metadata,
	})
	if err != nil {
		logger.Errorf(ctx, "failed to update configuration document metadata: %+v, error: %v", metadata, err)
		return err
	}
	m.setCache(document)
	return nil
}

func (m *ConfigurationManager) bootstrapOrUpdateDefaultConfigurationDocument(ctx context.Context, config runtimeInterfaces.Configuration) error {
	document, err := m.GetEditableActiveDocument(ctx)
	if err != nil && errors.IsDoesNotExistError(err) {
		document = admin.ConfigurationDocument{
			Configurations: make(map[string]*admin.Configuration),
		}
		updatedDocument, err := util.UpdateDefaultConfigurationToDocument(ctx, &document, config)
		if err != nil {
			logger.Errorf(ctx, "failed to update default configuration to document: %v", err)
			return err
		}

		metadata, err := m.getDocumentMetadata(ctx, updatedDocument)
		if err != nil {
			return err
		}
		return m.createDocumentMetadata(ctx, updatedDocument, metadata)
	}
	if err != nil {
		return err
	}

	versionToUpdate := document.Version
	updatedDocument, err := util.UpdateDefaultConfigurationToDocument(ctx, &document, config)
	if err != nil {
		logger.Errorf(ctx, "failed to update default configuration to document: %v", err)
		return err
	}

	metadata, err := m.getDocumentMetadata(ctx, updatedDocument)
	if err != nil {
		return err
	}

	return m.updateDocumentMetadata(ctx, updatedDocument, metadata, versionToUpdate)
}

func (m *ConfigurationManager) canEditAttributes(ctx context.Context, request admin.ConfigurationUpdateRequest, projectConfigurationPlugin plugin.ProjectConfigurationPlugin) error {
	// Check if the configuration is editing non-mutable attributes
	document, err := m.GetReadOnlyActiveDocument(ctx)
	if err != nil {
		return err
	}

	currentConfiguration, err := util.GetConfigurationFromDocument(ctx, &document, request.Id)
	if err != nil {
		return err
	}
	// If the new configuration contains cluster assignment, we should use that to determine mutability
	clusterAssignment := currentConfiguration.GetClusterAssignment()
	if request.Configuration.ClusterAssignment != nil {
		clusterAssignment = request.Configuration.ClusterAssignment
	}
	mutableAttributes, err := projectConfigurationPlugin.GetMutableAttributes(ctx, &plugin.GetMutableAttributesInput{ClusterAssignment: clusterAssignment})
	if err != nil {
		return err
	}
	var notEditableAttributes []string
	if !mutableAttributes.Has(admin.MatchableResource_TASK_RESOURCE) && request.Configuration.TaskResourceAttributes != nil {
		notEditableAttributes = append(notEditableAttributes, admin.MatchableResource_name[int32(admin.MatchableResource_TASK_RESOURCE)])
	}
	if !mutableAttributes.Has(admin.MatchableResource_CLUSTER_RESOURCE) && request.Configuration.ClusterResourceAttributes != nil {
		notEditableAttributes = append(notEditableAttributes, admin.MatchableResource_name[int32(admin.MatchableResource_CLUSTER_RESOURCE)])
	}
	if !mutableAttributes.Has(admin.MatchableResource_EXECUTION_QUEUE) && request.Configuration.ExecutionQueueAttributes != nil {
		notEditableAttributes = append(notEditableAttributes, admin.MatchableResource_name[int32(admin.MatchableResource_EXECUTION_QUEUE)])
	}
	if !mutableAttributes.Has(admin.MatchableResource_EXECUTION_CLUSTER_LABEL) && request.Configuration.ExecutionClusterLabel != nil {
		notEditableAttributes = append(notEditableAttributes, admin.MatchableResource_name[int32(admin.MatchableResource_EXECUTION_CLUSTER_LABEL)])
	}
	if !mutableAttributes.Has(admin.MatchableResource_QUALITY_OF_SERVICE_SPECIFICATION) && request.Configuration.QualityOfService != nil {
		notEditableAttributes = append(notEditableAttributes, admin.MatchableResource_name[int32(admin.MatchableResource_QUALITY_OF_SERVICE_SPECIFICATION)])
	}
	if !mutableAttributes.Has(admin.MatchableResource_PLUGIN_OVERRIDE) && request.Configuration.PluginOverrides != nil {
		notEditableAttributes = append(notEditableAttributes, admin.MatchableResource_name[int32(admin.MatchableResource_PLUGIN_OVERRIDE)])
	}
	if !mutableAttributes.Has(admin.MatchableResource_WORKFLOW_EXECUTION_CONFIG) && request.Configuration.WorkflowExecutionConfig != nil {
		notEditableAttributes = append(notEditableAttributes, admin.MatchableResource_name[int32(admin.MatchableResource_WORKFLOW_EXECUTION_CONFIG)])
	}
	if !mutableAttributes.Has(admin.MatchableResource_CLUSTER_ASSIGNMENT) && request.Configuration.ClusterAssignment != nil {
		notEditableAttributes = append(notEditableAttributes, admin.MatchableResource_name[int32(admin.MatchableResource_CLUSTER_ASSIGNMENT)])
	}
	if !mutableAttributes.Has(admin.MatchableResource_EXTERNAL_RESOURCE) && request.Configuration.ExternalResourceAttributes != nil {
		notEditableAttributes = append(notEditableAttributes, admin.MatchableResource_name[int32(admin.MatchableResource_EXTERNAL_RESOURCE)])

	}
	if len(notEditableAttributes) > 0 {
		logger.Debugf(ctx, "Not editable attributes: %v", notEditableAttributes)
		return errors.NewFlyteAdminErrorf(codes.InvalidArgument, "attributes not editable: %v", notEditableAttributes)
	}
	return nil
}

// In the context of a flyteadmin, we expect to mutate the global configuration based on the values in configmap. At this scenario, we expect bootstrap to be true.
// In the context of a cluster resource controller, we expect to read the configuration from the database and not mutate it. At this scenario, we expect bootstrap to be false.
func NewConfigurationManager(ctx context.Context, db repositoryInterfaces.Repository, config runtimeInterfaces.Configuration, storageClient *storage.DataStore, pluginRegistry *plugins.Registry, bootstrapOrUpdateDefault bool) (interfaces.ConfigurationInterface, error) {
	configurationManager := &ConfigurationManager{
		db:            db,
		config:        config,
		storageClient: storageClient,
		cacheConfigDoc: &configurationDocumentCache{
			configDoc: nil,
			mutex:     sync.RWMutex{},
		},
		pluginRegistry: pluginRegistry,
	}
	if bootstrapOrUpdateDefault {
		logger.Debug(ctx, "Bootstrapping or updating default configuration document")
		err := configurationManager.bootstrapOrUpdateDefaultConfigurationDocument(ctx, config)
		if err != nil {
			return nil, err
		}
	}
	return configurationManager, nil
}
