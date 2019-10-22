package mocks

import "github.com/lyft/flyteadmin/pkg/runtime/interfaces"

type MockConfigurationProvider struct {
	applicationConfiguration            interfaces.ApplicationConfiguration
	queueConfiguration                  interfaces.QueueConfiguration
	clusterConfiguration                interfaces.ClusterConfiguration
	taskResourceConfiguration           interfaces.TaskResourceConfiguration
	whitelistConfiguration              interfaces.WhitelistConfiguration
	registrationValidationConfiguration interfaces.RegistrationValidationConfiguration
	clusterResourceConfiguration        interfaces.ClusterResourceConfiguration
	namespaceMappingConfiguration       interfaces.NamespaceMappingConfiguration
}

func (p *MockConfigurationProvider) ApplicationConfiguration() interfaces.ApplicationConfiguration {
	return p.applicationConfiguration
}

func (p *MockConfigurationProvider) QueueConfiguration() interfaces.QueueConfiguration {
	return p.queueConfiguration
}

func (p *MockConfigurationProvider) ClusterConfiguration() interfaces.ClusterConfiguration {
	return p.clusterConfiguration
}

func (p *MockConfigurationProvider) TaskResourceConfiguration() interfaces.TaskResourceConfiguration {
	return p.taskResourceConfiguration
}

func (p *MockConfigurationProvider) WhitelistConfiguration() interfaces.WhitelistConfiguration {
	return p.whitelistConfiguration
}

func (p *MockConfigurationProvider) RegistrationValidationConfiguration() interfaces.RegistrationValidationConfiguration {
	return p.registrationValidationConfiguration
}

func (p *MockConfigurationProvider) AddRegistrationValidationConfiguration(config interfaces.RegistrationValidationConfiguration) {
	p.registrationValidationConfiguration = config
}

func (p *MockConfigurationProvider) ClusterResourceConfiguration() interfaces.ClusterResourceConfiguration {
	return p.clusterResourceConfiguration
}

func (p *MockConfigurationProvider) AddClusterResourceConfiguration(config interfaces.ClusterResourceConfiguration) {
	p.clusterResourceConfiguration = config
}

func (p *MockConfigurationProvider) NamespaceMappingConfiguration() interfaces.NamespaceMappingConfiguration {
	return p.namespaceMappingConfiguration
}

func (p *MockConfigurationProvider) AddNamespaceMappingConfiguration(config interfaces.NamespaceMappingConfiguration) {
	p.namespaceMappingConfiguration = config
}

func NewMockConfigurationProvider(
	applicationConfiguration interfaces.ApplicationConfiguration,
	queueConfiguration interfaces.QueueConfiguration,
	clusterConfiguration interfaces.ClusterConfiguration,
	taskResourceConfiguration interfaces.TaskResourceConfiguration,
	whitelistConfiguration interfaces.WhitelistConfiguration,
	namespaceMappingConfiguration interfaces.NamespaceMappingConfiguration) interfaces.Configuration {
	return &MockConfigurationProvider{
		applicationConfiguration:      applicationConfiguration,
		queueConfiguration:            queueConfiguration,
		clusterConfiguration:          clusterConfiguration,
		taskResourceConfiguration:     taskResourceConfiguration,
		whitelistConfiguration:        whitelistConfiguration,
		namespaceMappingConfiguration: namespaceMappingConfiguration,
	}
}
