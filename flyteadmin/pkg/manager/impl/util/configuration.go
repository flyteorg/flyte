package util

import (
	"context"

	"github.com/golang/protobuf/proto"
	"google.golang.org/grpc/codes"

	flyteErrs "github.com/flyteorg/flyte/flyteadmin/pkg/errors"
	"github.com/flyteorg/flyte/flyteadmin/pkg/manager/impl/configurations/plugin"
	runtimeInterfaces "github.com/flyteorg/flyte/flyteadmin/pkg/runtime/interfaces"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flytestdlib/logger"
	"github.com/flyteorg/flyte/flytestdlib/utils"
)

// A special configuration key to represent the global configuration.
var GlobalConfigurationKey = admin.ConfigurationID{
	Global: true,
}

func addConfigurationSource(configuration *admin.Configuration, source admin.AttributesSource) *admin.ConfigurationWithSource {
	configurationWithSource := &admin.ConfigurationWithSource{}
	configurationWithSource.TaskResourceAttributes = &admin.TaskResourceAttributesWithSource{
		Source: source,
		Value:  configuration.TaskResourceAttributes,
	}
	configurationWithSource.ClusterResourceAttributes = &admin.ClusterResourceAttributesWithSource{
		Source: source,
		Value:  configuration.ClusterResourceAttributes,
	}
	configurationWithSource.ExecutionQueueAttributes = &admin.ExecutionQueueAttributesWithSource{
		Source: source,
		Value:  configuration.ExecutionQueueAttributes,
	}
	configurationWithSource.ExecutionClusterLabel = &admin.ExecutionClusterLabelWithSource{
		Source: source,
		Value:  configuration.ExecutionClusterLabel,
	}
	configurationWithSource.QualityOfService = &admin.QualityOfServiceWithSource{
		Source: source,
		Value:  configuration.QualityOfService,
	}
	configurationWithSource.PluginOverrides = &admin.PluginOverridesWithSource{
		Source: source,
		Value:  configuration.PluginOverrides,
	}
	configurationWithSource.WorkflowExecutionConfig = &admin.WorkflowExecutionConfigWithSource{
		Source: source,
		Value:  configuration.WorkflowExecutionConfig,
	}
	configurationWithSource.ClusterAssignment = &admin.ClusterAssignmentWithSource{
		Source: source,
		Value:  configuration.ClusterAssignment,
	}
	configurationWithSource.ExternalResourceAttributes = &admin.ExternalResourceAttributesWithSource{
		Source: source,
		Value:  configuration.ExternalResourceAttributes,
	}
	return configurationWithSource
}

// If current configuration is missing a field, merge the incoming configuration into the current configuration.
func mergeConfigurations(current *admin.ConfigurationWithSource, incoming *admin.ConfigurationWithSource) *admin.ConfigurationWithSource {
	if current.TaskResourceAttributes == nil || current.TaskResourceAttributes.Value == nil {
		current.TaskResourceAttributes = incoming.TaskResourceAttributes
	}
	if current.ClusterResourceAttributes == nil || current.ClusterResourceAttributes.Value == nil {
		current.ClusterResourceAttributes = incoming.ClusterResourceAttributes
	}
	if current.ExecutionQueueAttributes == nil || current.ExecutionQueueAttributes.Value == nil {
		current.ExecutionQueueAttributes = incoming.ExecutionQueueAttributes
	}
	if current.ExecutionClusterLabel == nil || current.ExecutionClusterLabel.Value == nil {
		current.ExecutionClusterLabel = incoming.ExecutionClusterLabel
	}
	if current.QualityOfService == nil || current.QualityOfService.Value == nil {
		current.QualityOfService = incoming.QualityOfService
	}
	if current.PluginOverrides == nil || current.PluginOverrides.Value == nil {
		current.PluginOverrides = incoming.PluginOverrides
	}
	if current.WorkflowExecutionConfig == nil || current.WorkflowExecutionConfig.Value == nil {
		current.WorkflowExecutionConfig = incoming.WorkflowExecutionConfig
	}
	if current.ClusterAssignment == nil || current.ClusterAssignment.Value == nil {
		current.ClusterAssignment = incoming.ClusterAssignment
	}
	if current.ExternalResourceAttributes == nil || current.ExternalResourceAttributes.Value == nil {
		current.ExternalResourceAttributes = incoming.ExternalResourceAttributes
	}
	return current
}

// Ensure that all attributes have metadata.
func ensureAttributesMetadata(configuration *admin.ConfigurationWithSource) *admin.ConfigurationWithSource {
	if configuration.TaskResourceAttributes.Metadata == nil {
		configuration.TaskResourceAttributes.Metadata = &admin.AttributeMetadata{}
	}
	if configuration.ClusterResourceAttributes.Metadata == nil {
		configuration.ClusterResourceAttributes.Metadata = &admin.AttributeMetadata{}
	}
	if configuration.ExecutionQueueAttributes.Metadata == nil {
		configuration.ExecutionQueueAttributes.Metadata = &admin.AttributeMetadata{}
	}
	if configuration.ExecutionClusterLabel.Metadata == nil {
		configuration.ExecutionClusterLabel.Metadata = &admin.AttributeMetadata{}
	}
	if configuration.QualityOfService.Metadata == nil {
		configuration.QualityOfService.Metadata = &admin.AttributeMetadata{}
	}
	if configuration.PluginOverrides.Metadata == nil {
		configuration.PluginOverrides.Metadata = &admin.AttributeMetadata{}
	}
	if configuration.WorkflowExecutionConfig.Metadata == nil {
		configuration.WorkflowExecutionConfig.Metadata = &admin.AttributeMetadata{}
	}
	if configuration.ClusterAssignment.Metadata == nil {
		configuration.ClusterAssignment.Metadata = &admin.AttributeMetadata{}
	}
	if configuration.ExternalResourceAttributes.Metadata == nil {
		configuration.ExternalResourceAttributes.Metadata = &admin.AttributeMetadata{}
	}
	return configuration
}

func addConfigurationIsMutable(ctx context.Context, configuration *admin.ConfigurationWithSource, id *admin.ConfigurationID, projectConfigurationPlugin plugin.ProjectConfigurationPlugin) (*admin.ConfigurationWithSource, error) {
	clusterAssignment := configuration.GetClusterAssignment().GetValue()
	attributeIsMutable, err := projectConfigurationPlugin.GetAttributeIsMutable(ctx, &plugin.GetAttributeIsMutable{
		ClusterAssignment: clusterAssignment,
		ConfigurationID:   id,
	})
	if err != nil {
		return nil, err
	}

	configuration.TaskResourceAttributes.Metadata.IsMutable = attributeIsMutable[admin.MatchableResource_TASK_RESOURCE]
	configuration.ClusterResourceAttributes.Metadata.IsMutable = attributeIsMutable[admin.MatchableResource_CLUSTER_RESOURCE]
	configuration.ExecutionQueueAttributes.Metadata.IsMutable = attributeIsMutable[admin.MatchableResource_EXECUTION_QUEUE]
	configuration.ExecutionClusterLabel.Metadata.IsMutable = attributeIsMutable[admin.MatchableResource_EXECUTION_CLUSTER_LABEL]
	configuration.QualityOfService.Metadata.IsMutable = attributeIsMutable[admin.MatchableResource_QUALITY_OF_SERVICE_SPECIFICATION]
	configuration.PluginOverrides.Metadata.IsMutable = attributeIsMutable[admin.MatchableResource_PLUGIN_OVERRIDE]
	configuration.WorkflowExecutionConfig.Metadata.IsMutable = attributeIsMutable[admin.MatchableResource_WORKFLOW_EXECUTION_CONFIG]
	configuration.ClusterAssignment.Metadata.IsMutable = attributeIsMutable[admin.MatchableResource_CLUSTER_ASSIGNMENT]
	configuration.ExternalResourceAttributes.Metadata.IsMutable = attributeIsMutable[admin.MatchableResource_EXTERNAL_RESOURCE]
	return configuration, nil
}

func GetConfiguration(ctx context.Context, document *admin.ConfigurationDocument, id *admin.ConfigurationID, projectConfigurationPlugin plugin.ProjectConfigurationPlugin) (*admin.ConfigurationWithSource, error) {
	return getConfigurationWithSourceAndMetadata(ctx, document, id, projectConfigurationPlugin, false)
}

func GetConfigurationOnlyLowerLevel(ctx context.Context, document *admin.ConfigurationDocument, id *admin.ConfigurationID, projectConfigurationPlugin plugin.ProjectConfigurationPlugin) (*admin.ConfigurationWithSource, error) {
	return getConfigurationWithSourceAndMetadata(ctx, document, id, projectConfigurationPlugin, true)
}

// getConfigurationWithSourceAndMetadata merges the configuration from higher level to lower level.
// The order of precedence is org-project-domain > org-project > org > domain > global.
// This is because org-project-domain, org-project, and org are users' overrides,
// while domain and global are defaults from the config map.
// If onlyLowerLevel is true, the first match will be skipped
// to get the configuration which excludes the specified level.
func getConfigurationWithSourceAndMetadata(ctx context.Context, document *admin.ConfigurationDocument, id *admin.ConfigurationID, projectConfigurationPlugin plugin.ProjectConfigurationPlugin, onlyLowerLevel bool) (*admin.ConfigurationWithSource, error) {
	var configurations []*admin.Configuration
	var configurationLevels []admin.AttributesSource
	matchFoundAtLevel := false
	// Helper function to add configuration to the list
	addConfiguration := func(configurationID *admin.ConfigurationID, configurationLevel admin.AttributesSource) error {
		// Ensure matchFoundAtLevel is set to true before function returns
		defer func() {
			matchFoundAtLevel = true
		}()
		// If mode is onlyLowerLevel, we would like to skip the first match
		if onlyLowerLevel && !matchFoundAtLevel {
			return nil
		}
		configuration, err := GetConfigurationFromDocument(ctx, document, configurationID)
		if err != nil {
			return err
		}
		configurations = append(configurations, &configuration)
		configurationLevels = append(configurationLevels, configurationLevel)
		return nil
	}

	// Org-Project-Domain
	if id.Project != "" && id.Domain != "" {
		if err := addConfiguration(&admin.ConfigurationID{
			Org:     id.Org,
			Project: id.Project,
			Domain:  id.Domain,
		}, admin.AttributesSource_PROJECT_DOMAIN); err != nil {
			return nil, err
		}
	}
	// Org-Project
	if id.Project != "" {
		if err := addConfiguration(&admin.ConfigurationID{
			Org:     id.Org,
			Project: id.Project,
		}, admin.AttributesSource_PROJECT); err != nil {
			return nil, err
		}
	}
	// Org
	if err := addConfiguration(&admin.ConfigurationID{
		Org: id.Org,
	}, admin.AttributesSource_ORG); err != nil {
		return nil, err
	}
	// Domain
	if id.Domain != "" {
		if err := addConfiguration(&admin.ConfigurationID{
			Domain: id.Domain,
		}, admin.AttributesSource_DOMAIN); err != nil {
			return nil, err
		}
	}
	// Global
	if err := addConfiguration(&GlobalConfigurationKey, admin.AttributesSource_GLOBAL); err != nil {
		return nil, err
	}
	// Merge configurations from higher level to lower level
	configuration := &admin.ConfigurationWithSource{}
	for i, conf := range configurations {
		source := configurationLevels[i]
		configuration = mergeConfigurations(configuration, addConfigurationSource(conf, source))
	}
	configuration = ensureAttributesMetadata(configuration)
	return addConfigurationIsMutable(ctx, configuration, id, projectConfigurationPlugin)
}

func GetConfigurationFromDocument(ctx context.Context, document *admin.ConfigurationDocument, id *admin.ConfigurationID) (admin.Configuration, error) {
	logger.Debugf(ctx, "Getting configuration for org: %s, project: %s, domain: %s, workflow: %s", id.Org, id.Project, id.Domain, id.Workflow)
	documentKey, err := EncodeConfigurationDocumentKey(ctx, id)
	if err != nil {
		logger.Errorf(ctx, "Failed to encode document key: %+v, error: %v", id, err)
		return admin.Configuration{}, err
	}
	if _, ok := document.Configurations[documentKey]; !ok {
		return admin.Configuration{}, nil
	}
	return *document.Configurations[documentKey], nil
}

func UpdateConfigurationToDocument(ctx context.Context, document *admin.ConfigurationDocument, configuration *admin.Configuration, id *admin.ConfigurationID) (*admin.ConfigurationDocument, error) {
	logger.Debugf(ctx, "Updating configuration for org: %s, project: %s, domain: %s, workflow: %s", id.Org, id.Project, id.Domain, id.Workflow)
	documentKey, err := EncodeConfigurationDocumentKey(ctx, id)
	if err != nil {
		logger.Errorf(ctx, "Failed to encode document key: %+v, error: %v", id, err)
		return nil, err
	}
	logger.Debugf(ctx, "Existing configuration: %+v", document.Configurations[documentKey])
	logger.Debugf(ctx, "Incoming configuration: %+v", configuration)
	if proto.Equal(configuration, &admin.Configuration{}) {
		// Remove configuration if it is empty to ensure digest consistency
		delete(document.Configurations, documentKey)
	} else {
		document.Configurations[documentKey] = configuration
	}
	return document, nil
}

func EncodeConfigurationDocumentKey(ctx context.Context, id *admin.ConfigurationID) (string, error) {
	key, err := utils.MarshalPbToString(id)
	if err != nil {
		logger.Errorf(ctx, "Failed to marshal configuration ID: %+v to string, error:, %v", id, err)
		return "", flyteErrs.NewFlyteAdminErrorf(codes.Internal, "failed to marshal document key")
	}
	logger.Debugf(ctx, "Encoded configuration id: %v to document key: %s", id, key)
	return key, nil
}

func DecodeConfigurationDocumentKey(ctx context.Context, key string) (*admin.ConfigurationID, error) {
	id := &admin.ConfigurationID{}
	err := utils.UnmarshalStringToPb(key, id)
	if err != nil {
		logger.Errorf(ctx, "Failed to unmarshal configuration ID from string: %s, error: %v", key, err)
		return nil, flyteErrs.NewFlyteAdminErrorf(codes.Internal, "failed to unmarshal document key")
	}
	logger.Debugf(ctx, "Decoded document key: %s to configuration id: %v", key, id)
	return id, nil
}

func UpdateDefaultConfigurationToDocument(ctx context.Context, document *admin.ConfigurationDocument, config runtimeInterfaces.Configuration) (*admin.ConfigurationDocument, error) {
	// update global configuration
	document, err := UpdateConfigurationToDocument(ctx, document, collectGlobalConfiguration(ctx, config), &GlobalConfigurationKey)
	if err != nil {
		return nil, err
	}

	// update domain configuration
	domains := config.ApplicationConfiguration().GetDomainsConfig()
	for _, domain := range *domains {
		document, err = UpdateConfigurationToDocument(ctx, document, collectDomainConfiguration(ctx, config, domain.ID), &admin.ConfigurationID{Domain: domain.ID})
		if err != nil {
			return nil, err
		}
	}

	return document, nil
}

func collectGlobalConfiguration(ctx context.Context, config runtimeInterfaces.Configuration) *admin.Configuration {
	logger.Debug(ctx, "Collecting global configuration")
	// Task resource attributes
	taskResourceAttributesConfig := config.TaskResourceConfiguration()
	defaults := taskResourceAttributesConfig.GetDefaults()
	limits := taskResourceAttributesConfig.GetLimits()
	taskResourceAttributes := admin.TaskResourceAttributes{
		Defaults: ToAdminProtoTaskResourceSpec(&defaults),
		Limits:   ToAdminProtoTaskResourceSpec(&limits),
	}

	// Workflow execution configuration
	workflowExecutionConfig := config.ApplicationConfiguration().GetTopLevelConfig().GetAsWorkflowExecutionConfig()

	// External resource attributes
	externalResourceAttributesConfig := config.ExternalResourceConfiguration()
	connections := make(map[string]*core.Connection)
	for key, connection := range externalResourceAttributesConfig.GetConnections() {
		connections[key] = &core.Connection{
			TaskType: connection.TaskType,
			Secrets:  connection.Secrets,
			Configs:  connection.Configs,
		}
	}
	ExternalResourceAttributes := admin.ExternalResourceAttributes{
		Connections: connections,
	}

	return &admin.Configuration{
		TaskResourceAttributes:     &taskResourceAttributes,
		ClusterResourceAttributes:  nil, // handle in domain level
		ExecutionQueueAttributes:   nil,
		ExecutionClusterLabel:      nil,
		QualityOfService:           nil,
		PluginOverrides:            nil,
		WorkflowExecutionConfig:    &workflowExecutionConfig,
		ClusterAssignment:          nil, // handle in domain level
		ExternalResourceAttributes: &ExternalResourceAttributes,
	}
}

func collectDomainConfiguration(ctx context.Context, config runtimeInterfaces.Configuration, domain string) *admin.Configuration {
	logger.Debugf(ctx, "Collecting domain configuration for domain %s", domain)
	// Cluster resource attributes
	customTemplateData := config.ClusterResourceConfiguration().GetCustomTemplateData()
	var clusterResourceAttributes *admin.ClusterResourceAttributes
	if _, ok := customTemplateData[domain]; ok {
		customTemplateDataInDomain := customTemplateData[domain]
		attributes := make(map[string]string)
		for key, value := range customTemplateDataInDomain {
			attributes[key] = value.Value
		}
		clusterResourceAttributes = &admin.ClusterResourceAttributes{
			Attributes: attributes,
		}
	} else {
		logger.Debugf(ctx, "No custom template data found for domain %s", domain)
		clusterResourceAttributes = nil
	}

	// Cluster assignment
	clusterPoolAssignments := config.ClusterPoolAssignmentConfiguration().GetClusterPoolAssignments()
	var clusterPoolAssignment string
	if _, ok := clusterPoolAssignments[domain]; ok {
		clusterPoolAssignment = clusterPoolAssignments[domain].Pool
	} else {
		logger.Errorf(ctx, "No cluster pool assignment found for domain %s", domain)
		clusterPoolAssignment = ""
	}

	return &admin.Configuration{
		TaskResourceAttributes:    nil,
		ClusterResourceAttributes: clusterResourceAttributes,
		ExecutionQueueAttributes:  nil,
		ExecutionClusterLabel:     nil,
		QualityOfService:          nil,
		PluginOverrides:           nil,
		WorkflowExecutionConfig:   nil,
		ClusterAssignment: &admin.ClusterAssignment{
			ClusterPoolName: clusterPoolAssignment,
		},
	}
}

// A default configuration is either a global configuration or a domain configuration.
// Both of them are from the config map.
func IsDefaultConfigurationID(ctx context.Context, configurationID *admin.ConfigurationID) bool {
	// Is a global configuration
	if proto.Equal(configurationID, &GlobalConfigurationKey) {
		return true
	}
	// Is a domain configuration
	return configurationID.Workflow == "" && configurationID.Project == "" && configurationID.Domain != "" && configurationID.Org == ""
}
