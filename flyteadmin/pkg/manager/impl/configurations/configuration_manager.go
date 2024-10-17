package configurations

import (
	"context"
	"sync"

	"github.com/golang/protobuf/proto"
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
	validator      validation.ConfigurationValidator
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
	// Set the configuration ID if it is not set.
	// This is because org can be empty and if it is empty, the ID will be nil.
	if request.Id == nil {
		request.Id = &admin.ConfigurationID{}
	}

	// Validate the request
	if err := m.validator.ValidateGetRequest(ctx, request); err != nil {
		return nil, err
	}
	// Get the active document
	activeDocument, err := m.GetReadOnlyActiveDocument(ctx)
	if err != nil {
		return nil, err
	}
	// Get the configuration
	var configuration *admin.ConfigurationWithSource
	if request.GetOnlyGetLowerLevelConfiguration() {
		configuration, err = util.GetConfigurationOnlyLowerLevel(ctx, &activeDocument, request.GetId(), plugins.Get[plugin.ProjectConfigurationPlugin](m.pluginRegistry, plugins.PluginIDProjectConfiguration))
	} else {
		configuration, err = util.GetConfiguration(ctx, &activeDocument, request.GetId(), plugins.Get[plugin.ProjectConfigurationPlugin](m.pluginRegistry, plugins.PluginIDProjectConfiguration))
	}
	if err != nil {
		return nil, err
	}
	response := &admin.ConfigurationGetResponse{
		Id:            request.GetId(),
		Version:       activeDocument.GetVersion(),
		Configuration: configuration,
	}
	return response, nil
}

func (m *ConfigurationManager) UpdateConfiguration(
	ctx context.Context, request admin.ConfigurationUpdateRequest) (
	*admin.ConfigurationUpdateResponse, error) {
	// Validate the request
	if err := m.validator.ValidateUpdateRequest(ctx, request); err != nil {
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

	// Get the configuration
	configuration, err := util.GetConfiguration(ctx, updatedDocument, request.GetId(), projectConfigurationPlugin)
	if err != nil {
		return nil, err
	}

	return &admin.ConfigurationUpdateResponse{
		Id:            request.GetId(),
		Version:       metadata.Version,
		Configuration: configuration,
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
	attributeIsMutable, err := projectConfigurationPlugin.GetAttributeIsMutable(ctx, &plugin.GetAttributeIsMutable{
		ClusterAssignment: clusterAssignment,
		ConfigurationID:   request.Id,
	})
	if err != nil {
		return err
	}

	// Check if the configuration is editing non-mutable attributes
	var notEditableAttributes []string
	if !attributeIsMutable[admin.MatchableResource_TASK_RESOURCE].GetValue() && !proto.Equal(request.GetConfiguration().GetTaskResourceAttributes(), currentConfiguration.GetTaskResourceAttributes()) {
		notEditableAttributes = append(notEditableAttributes, admin.MatchableResource_name[int32(admin.MatchableResource_TASK_RESOURCE)])
	}
	if !attributeIsMutable[admin.MatchableResource_CLUSTER_RESOURCE].GetValue() && !proto.Equal(request.GetConfiguration().GetClusterResourceAttributes(), currentConfiguration.GetClusterResourceAttributes()) {
		notEditableAttributes = append(notEditableAttributes, admin.MatchableResource_name[int32(admin.MatchableResource_CLUSTER_RESOURCE)])
	}
	if !attributeIsMutable[admin.MatchableResource_EXECUTION_QUEUE].GetValue() && !proto.Equal(request.GetConfiguration().GetExecutionQueueAttributes(), currentConfiguration.GetExecutionQueueAttributes()) {
		notEditableAttributes = append(notEditableAttributes, admin.MatchableResource_name[int32(admin.MatchableResource_EXECUTION_QUEUE)])
	}
	if !attributeIsMutable[admin.MatchableResource_EXECUTION_CLUSTER_LABEL].GetValue() && !proto.Equal(request.GetConfiguration().GetExecutionClusterLabel(), currentConfiguration.GetExecutionClusterLabel()) {
		notEditableAttributes = append(notEditableAttributes, admin.MatchableResource_name[int32(admin.MatchableResource_EXECUTION_CLUSTER_LABEL)])
	}
	if !attributeIsMutable[admin.MatchableResource_QUALITY_OF_SERVICE_SPECIFICATION].GetValue() && !proto.Equal(request.GetConfiguration().GetQualityOfService(), currentConfiguration.GetQualityOfService()) {
		notEditableAttributes = append(notEditableAttributes, admin.MatchableResource_name[int32(admin.MatchableResource_QUALITY_OF_SERVICE_SPECIFICATION)])
	}
	if !attributeIsMutable[admin.MatchableResource_PLUGIN_OVERRIDE].GetValue() && !proto.Equal(request.GetConfiguration().GetPluginOverrides(), currentConfiguration.GetPluginOverrides()) {
		notEditableAttributes = append(notEditableAttributes, admin.MatchableResource_name[int32(admin.MatchableResource_PLUGIN_OVERRIDE)])
	}
	if !attributeIsMutable[admin.MatchableResource_WORKFLOW_EXECUTION_CONFIG].GetValue() && !proto.Equal(request.GetConfiguration().GetWorkflowExecutionConfig(), currentConfiguration.GetWorkflowExecutionConfig()) {
		notEditableAttributes = append(notEditableAttributes, admin.MatchableResource_name[int32(admin.MatchableResource_WORKFLOW_EXECUTION_CONFIG)])
	}
	if !attributeIsMutable[admin.MatchableResource_CLUSTER_ASSIGNMENT].GetValue() && !proto.Equal(request.GetConfiguration().GetClusterAssignment(), currentConfiguration.GetClusterAssignment()) {
		notEditableAttributes = append(notEditableAttributes, admin.MatchableResource_name[int32(admin.MatchableResource_CLUSTER_ASSIGNMENT)])
	}
	if !attributeIsMutable[admin.MatchableResource_EXTERNAL_RESOURCE].GetValue() && !proto.Equal(request.GetConfiguration().GetExternalResourceAttributes(), currentConfiguration.GetExternalResourceAttributes()) {
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
	validator := validation.NewConfigurationValidator(db, config.ApplicationConfiguration())
	configurationManager := &ConfigurationManager{
		db:            db,
		config:        config,
		storageClient: storageClient,
		cacheConfigDoc: &configurationDocumentCache{
			configDoc: nil,
			mutex:     sync.RWMutex{},
		},
		pluginRegistry: pluginRegistry,
		validator:      validator,
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
