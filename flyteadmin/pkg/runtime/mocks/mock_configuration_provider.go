package mocks

import (
	"github.com/flyteorg/flyteadmin/pkg/runtime/interfaces"
	ifaceMocks "github.com/flyteorg/flyteadmin/pkg/runtime/interfaces/mocks"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
)

type MockConfigurationProvider struct {
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

func (p *MockConfigurationProvider) QualityOfServiceConfiguration() interfaces.QualityOfServiceConfiguration {
	return p.qualityOfServiceConfiguration
}

func (p *MockConfigurationProvider) AddQualityOfServiceConfiguration(config interfaces.QualityOfServiceConfiguration) {
	p.qualityOfServiceConfiguration = config
}

func (p *MockConfigurationProvider) ClusterPoolAssignmentConfiguration() interfaces.ClusterPoolAssignmentConfiguration {
	return p.clusterPoolAssignmentConfiguration
}

func (p *MockConfigurationProvider) AddClusterPoolAssignmentConfiguration(cfg interfaces.ClusterPoolAssignmentConfiguration) {
	p.clusterPoolAssignmentConfiguration = cfg
}

func NewMockConfigurationProvider(
	applicationConfiguration interfaces.ApplicationConfiguration,
	queueConfiguration interfaces.QueueConfiguration,
	clusterConfiguration interfaces.ClusterConfiguration,
	taskResourceConfiguration interfaces.TaskResourceConfiguration,
	whitelistConfiguration interfaces.WhitelistConfiguration,
	namespaceMappingConfiguration interfaces.NamespaceMappingConfiguration) interfaces.Configuration {

	mockQualityOfServiceConfiguration := &ifaceMocks.QualityOfServiceConfiguration{}
	mockQualityOfServiceConfiguration.OnGetDefaultTiers().Return(make(map[string]core.QualityOfService_Tier))
	mockQualityOfServiceConfiguration.OnGetTierExecutionValues().Return(make(map[core.QualityOfService_Tier]core.QualityOfServiceSpec))

	mockClusterPoolAssignmentConfiguration := &ifaceMocks.ClusterPoolAssignmentConfiguration{}
	mockClusterPoolAssignmentConfiguration.OnGetClusterPoolAssignments().Return(make(map[string]interfaces.ClusterPoolAssignment))

	return &MockConfigurationProvider{
		applicationConfiguration:           applicationConfiguration,
		queueConfiguration:                 queueConfiguration,
		clusterConfiguration:               clusterConfiguration,
		taskResourceConfiguration:          taskResourceConfiguration,
		whitelistConfiguration:             whitelistConfiguration,
		namespaceMappingConfiguration:      namespaceMappingConfiguration,
		qualityOfServiceConfiguration:      mockQualityOfServiceConfiguration,
		clusterPoolAssignmentConfiguration: mockClusterPoolAssignmentConfiguration,
	}
}
