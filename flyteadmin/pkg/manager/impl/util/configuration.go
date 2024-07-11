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

func addConfigurationIsMutable(ctx context.Context, configuration *admin.ConfigurationWithSource, projectConfigurationPlugin plugin.ProjectConfigurationPlugin) (*admin.ConfigurationWithSource, error) {
	clusterAssignment := configuration.GetClusterAssignment().GetValue()
	mutableAttributes, err := projectConfigurationPlugin.GetMutableAttributes(ctx, &plugin.GetMutableAttributesInput{ClusterAssignment: clusterAssignment})
	if err != nil {
		return nil, err
	}

	configuration.TaskResourceAttributes.IsMutable = mutableAttributes.Has(admin.MatchableResource_TASK_RESOURCE)
	configuration.ClusterResourceAttributes.IsMutable = mutableAttributes.Has(admin.MatchableResource_CLUSTER_RESOURCE)
	configuration.ExecutionQueueAttributes.IsMutable = mutableAttributes.Has(admin.MatchableResource_EXECUTION_QUEUE)
	configuration.ExecutionClusterLabel.IsMutable = mutableAttributes.Has(admin.MatchableResource_EXECUTION_CLUSTER_LABEL)
	configuration.QualityOfService.IsMutable = mutableAttributes.Has(admin.MatchableResource_QUALITY_OF_SERVICE_SPECIFICATION)
	configuration.PluginOverrides.IsMutable = mutableAttributes.Has(admin.MatchableResource_PLUGIN_OVERRIDE)
	configuration.WorkflowExecutionConfig.IsMutable = mutableAttributes.Has(admin.MatchableResource_WORKFLOW_EXECUTION_CONFIG)
	configuration.ClusterAssignment.IsMutable = mutableAttributes.Has(admin.MatchableResource_CLUSTER_ASSIGNMENT)
	configuration.ExternalResourceAttributes.IsMutable = mutableAttributes.Has(admin.MatchableResource_EXTERNAL_RESOURCE)
	return configuration, nil
}

// GetConfigurationWithSource merges the configuration from higher level to lower level.
// The order of precedence is org-project-domain > org-project > org > domain > global.
// This is because org-project-domain, org-project, and org are users' overrides,
// while domain and global are defaults from the config map.
func GetConfigurationWithSource(ctx context.Context, document *admin.ConfigurationDocument, id *admin.ConfigurationID, projectConfigurationPlugin plugin.ProjectConfigurationPlugin) (*admin.ConfigurationWithSource, error) {
	var err error
	// org-project-domain
	projectDomainConfiguration := admin.Configuration{}
	if id.Project != "" && id.Domain != "" {
		projectDomainConfiguration, err = GetConfigurationFromDocument(ctx, document, &admin.ConfigurationID{
			Org:     id.Org,
			Project: id.Project,
			Domain:  id.Domain,
		})
		if err != nil {
			return nil, err
		}
	}
	// org-project
	projectConfiguration := admin.Configuration{}
	if id.Project != "" {
		projectConfiguration, err = GetConfigurationFromDocument(ctx, document, &admin.ConfigurationID{
			Org:     id.Org,
			Project: id.Project,
		})
		if err != nil {
			return nil, err
		}
	}
	// org
	orgConfiguration, err := GetConfigurationFromDocument(ctx, document, &admin.ConfigurationID{
		Org: id.Org,
	})
	if err != nil {
		return nil, err
	}
	// domain
	domainConfiguration := admin.Configuration{}
	if id.Domain != "" {
		domainConfiguration, err = GetConfigurationFromDocument(ctx, document, &admin.ConfigurationID{
			Domain: id.Domain,
		})
		if err != nil {
			return nil, err
		}
	}
	// global
	globalConfiguration, err := GetConfigurationFromDocument(ctx, document, &GlobalConfigurationKey)
	if err != nil {
		return nil, err
	}
	// Merge configurations from higher level to lower level
	configuration := addConfigurationSource(&projectDomainConfiguration, admin.AttributesSource_PROJECT_DOMAIN)
	configuration = mergeConfigurations(configuration, addConfigurationSource(&projectConfiguration, admin.AttributesSource_PROJECT))
	configuration = mergeConfigurations(configuration, addConfigurationSource(&orgConfiguration, admin.AttributesSource_ORG))
	configuration = mergeConfigurations(configuration, addConfigurationSource(&domainConfiguration, admin.AttributesSource_DOMAIN))
	configuration = mergeConfigurations(configuration, addConfigurationSource(&globalConfiguration, admin.AttributesSource_GLOBAL))

	return addConfigurationIsMutable(ctx, configuration, projectConfigurationPlugin)
}

func GetDefaultConfigurationWithSource(ctx context.Context, document *admin.ConfigurationDocument, domain string, projectConfigurationPlugin plugin.ProjectConfigurationPlugin) (*admin.ConfigurationWithSource, error) {
	domainConfiguration, err := GetConfigurationFromDocument(ctx, document, &admin.ConfigurationID{
		Domain: domain,
	})
	if err != nil {
		return nil, err
	}
	globalConfiguration, err := GetConfigurationFromDocument(ctx, document, &GlobalConfigurationKey)
	if err != nil {
		return nil, err
	}
	configuration := addConfigurationSource(&domainConfiguration, admin.AttributesSource_DOMAIN)
	configuration = mergeConfigurations(configuration, addConfigurationSource(&globalConfiguration, admin.AttributesSource_GLOBAL))

	return addConfigurationIsMutable(ctx, configuration, projectConfigurationPlugin)
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
