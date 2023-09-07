package runtime

import (
	"github.com/flyteorg/flyteadmin/pkg/runtime/interfaces"
)

// Implementation of an interfaces.Configuration
type ConfigurationProvider struct {
	applicationConfiguration            interfaces.ApplicationConfiguration
	queueConfiguration                  interfaces.QueueConfiguration
	clusterConfiguration                interfaces.ClusterConfiguration
	taskResourceConfiguration           interfaces.TaskResourceConfiguration
	whitelistConfiguration              interfaces.WhitelistConfiguration
	registrationValidationConfiguration interfaces.RegistrationValidationConfiguration
	clusterResourceConfiguration        interfaces.ClusterResourceConfiguration
	namespaceMappingConfiguration       interfaces.NamespaceMappingConfiguration
	qualityOfServiceConfiguration       interfaces.QualityOfServiceConfiguration
	clusterPoolAssignmentConfiguration  interfaces.ClusterPoolAssignmentConfiguration
}

func (p *ConfigurationProvider) ApplicationConfiguration() interfaces.ApplicationConfiguration {
	return p.applicationConfiguration
}

func (p *ConfigurationProvider) QueueConfiguration() interfaces.QueueConfiguration {
	return p.queueConfiguration
}

func (p *ConfigurationProvider) ClusterConfiguration() interfaces.ClusterConfiguration {
	return p.clusterConfiguration
}

func (p *ConfigurationProvider) TaskResourceConfiguration() interfaces.TaskResourceConfiguration {
	return p.taskResourceConfiguration
}

func (p *ConfigurationProvider) WhitelistConfiguration() interfaces.WhitelistConfiguration {
	return p.whitelistConfiguration
}

func (p *ConfigurationProvider) RegistrationValidationConfiguration() interfaces.RegistrationValidationConfiguration {
	return p.registrationValidationConfiguration
}

func (p *ConfigurationProvider) ClusterResourceConfiguration() interfaces.ClusterResourceConfiguration {
	return p.clusterResourceConfiguration
}

func (p *ConfigurationProvider) NamespaceMappingConfiguration() interfaces.NamespaceMappingConfiguration {
	return p.namespaceMappingConfiguration
}

func (p *ConfigurationProvider) QualityOfServiceConfiguration() interfaces.QualityOfServiceConfiguration {
	return p.qualityOfServiceConfiguration
}

func (p *ConfigurationProvider) ClusterPoolAssignmentConfiguration() interfaces.ClusterPoolAssignmentConfiguration {
	return p.clusterPoolAssignmentConfiguration
}

func NewConfigurationProvider() interfaces.Configuration {
	return &ConfigurationProvider{
		applicationConfiguration:            NewApplicationConfigurationProvider(),
		queueConfiguration:                  NewQueueConfigurationProvider(),
		clusterConfiguration:                NewClusterConfigurationProvider(),
		taskResourceConfiguration:           NewTaskResourceProvider(),
		whitelistConfiguration:              NewWhitelistConfigurationProvider(),
		registrationValidationConfiguration: NewRegistrationValidationProvider(),
		clusterResourceConfiguration:        NewClusterResourceConfigurationProvider(),
		namespaceMappingConfiguration:       NewNamespaceMappingConfigurationProvider(),
		qualityOfServiceConfiguration:       NewQualityOfServiceConfigProvider(),
		clusterPoolAssignmentConfiguration:  NewClusterPoolAssignmentConfigurationProvider(),
	}
}
