package util

import (
	"context"

	"github.com/golang/protobuf/proto"
	"google.golang.org/grpc/codes"

	flyteErrs "github.com/flyteorg/flyte/flyteadmin/pkg/errors"
	runtimeInterfaces "github.com/flyteorg/flyte/flyteadmin/pkg/runtime/interfaces"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyte/flytestdlib/logger"
	"github.com/flyteorg/flyte/flytestdlib/utils"
)

func AddConfigurationSource(configuration *admin.Configuration, source admin.AttributesSource) *admin.ConfigurationWithSource {
	configurationWithSource := &admin.ConfigurationWithSource{}
	if configuration.TaskResourceAttributes != nil {
		configurationWithSource.TaskResourceAttributes = &admin.TaskResourceAttributesWithSource{
			Source: source,
			Value:  configuration.TaskResourceAttributes,
		}
	}
	if configuration.ClusterResourceAttributes != nil {
		configurationWithSource.ClusterResourceAttributes = &admin.ClusterResourceAttributesWithSource{
			Source: source,
			Value:  configuration.ClusterResourceAttributes,
		}
	}
	if configuration.ExecutionQueueAttributes != nil {
		configurationWithSource.ExecutionQueueAttributes = &admin.ExecutionQueueAttributesWithSource{
			Source: source,
			Value:  configuration.ExecutionQueueAttributes,
		}
	}
	if configuration.ExecutionClusterLabel != nil {
		configurationWithSource.ExecutionClusterLabel = &admin.ExecutionClusterLabelWithSource{
			Source: source,
			Value:  configuration.ExecutionClusterLabel,
		}
	}
	if configuration.QualityOfService != nil {
		configurationWithSource.QualityOfService = &admin.QualityOfServiceWithSource{
			Source: source,
			Value:  configuration.QualityOfService,
		}
	}

	if configuration.PluginOverrides != nil {
		configurationWithSource.PluginOverrides = &admin.PluginOverridesWithSource{
			Source: source,
			Value:  configuration.PluginOverrides,
		}
	}
	if configuration.WorkflowExecutionConfig != nil {
		configurationWithSource.WorkflowExecutionConfig = &admin.WorkflowExecutionConfigWithSource{
			Source: source,
			Value:  configuration.WorkflowExecutionConfig,
		}
	}
	if configuration.ClusterAssignment != nil {
		configurationWithSource.ClusterAssignment = &admin.ClusterAssignmentWithSource{
			Source: source,
			Value:  configuration.ClusterAssignment,
		}
	}
	return configurationWithSource
}

// If current configuration is missing a field, merge the incoming configuration into the current configuration.
func MergeConfigurations(current *admin.ConfigurationWithSource, incoming *admin.ConfigurationWithSource) *admin.ConfigurationWithSource {
	if current.TaskResourceAttributes == nil {
		current.TaskResourceAttributes = incoming.TaskResourceAttributes
	}
	if current.ClusterResourceAttributes == nil {
		current.ClusterResourceAttributes = incoming.ClusterResourceAttributes
	}
	if current.ExecutionQueueAttributes == nil {
		current.ExecutionQueueAttributes = incoming.ExecutionQueueAttributes
	}
	if current.ExecutionClusterLabel == nil {
		current.ExecutionClusterLabel = incoming.ExecutionClusterLabel
	}
	if current.QualityOfService == nil {
		current.QualityOfService = incoming.QualityOfService
	}
	if current.PluginOverrides == nil {
		current.PluginOverrides = incoming.PluginOverrides
	}
	if current.WorkflowExecutionConfig == nil {
		current.WorkflowExecutionConfig = incoming.WorkflowExecutionConfig
	}
	if current.ClusterAssignment == nil {
		current.ClusterAssignment = incoming.ClusterAssignment
	}
	return current
}

// Merges the configuration from higher level to lower level.
func GetConfigurationWithSource(ctx context.Context, document *admin.ConfigurationDocument, id *admin.ConfigurationID) (*admin.ConfigurationWithSource, error) {
	projectDomainConfiguration, err := GetConfigurationFromDocument(ctx, document, &admin.ConfigurationID{
		Org:     id.Org,
		Project: id.Project,
		Domain:  id.Domain,
	})
	if err != nil {
		return nil, err
	}
	projectConfiguration, err := GetConfigurationFromDocument(ctx, document, &admin.ConfigurationID{
		Org:     id.Org,
		Project: id.Project,
	})
	if err != nil {
		return nil, err
	}
	domainConfiguration, err := GetConfigurationFromDocument(ctx, document, &admin.ConfigurationID{
		Domain: id.Domain,
	})
	if err != nil {
		return nil, err
	}
	globalConfiguration, err := GetConfigurationFromDocument(ctx, document, &admin.ConfigurationID{})
	if err != nil {
		return nil, err
	}
	configuration := AddConfigurationSource(&projectDomainConfiguration, admin.AttributesSource_PROJECT_DOMAIN)
	configuration = MergeConfigurations(configuration, AddConfigurationSource(&projectConfiguration, admin.AttributesSource_PROJECT))
	configuration = MergeConfigurations(configuration, AddConfigurationSource(&domainConfiguration, admin.AttributesSource_DOMAIN))
	configuration = MergeConfigurations(configuration, AddConfigurationSource(&globalConfiguration, admin.AttributesSource_GLOBAL))
	return configuration, nil
}

func GetDefaultConfigurationWithSource(ctx context.Context, document *admin.ConfigurationDocument, domain string) (*admin.ConfigurationWithSource, error) {
	domainConfiguration, err := GetConfigurationFromDocument(ctx, document, &admin.ConfigurationID{
		Domain: domain,
	})
	if err != nil {
		return nil, err
	}
	globalConfiguration, err := GetConfigurationFromDocument(ctx, document, &admin.ConfigurationID{})
	if err != nil {
		return nil, err
	}
	configuration := AddConfigurationSource(&domainConfiguration, admin.AttributesSource_DOMAIN)
	configuration = MergeConfigurations(configuration, AddConfigurationSource(&globalConfiguration, admin.AttributesSource_GLOBAL))
	return configuration, nil
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
	document, err := UpdateConfigurationToDocument(ctx, document, collectGlobalConfiguration(ctx, config), &admin.ConfigurationID{})
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
	workflowExecutionConfig := config.ApplicationConfiguration().GetTopLevelConfig().GetAsWorkflowExecutionConfig()
	return &admin.Configuration{
		TaskResourceAttributes:    &taskResourceAttributes,
		ClusterResourceAttributes: nil, // handle in domain level
		ExecutionQueueAttributes:  nil,
		ExecutionClusterLabel:     nil,
		QualityOfService:          nil,
		PluginOverrides:           nil,
		WorkflowExecutionConfig:   &workflowExecutionConfig,
		ClusterAssignment:         nil, // handle in domain level
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
