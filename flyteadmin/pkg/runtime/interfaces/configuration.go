package interfaces

//go:generate mockery -name=Configuration -output=../mocks -case=underscore

// Interface for getting parsed values from a configuration file
type Configuration interface {
	ApplicationConfiguration() ApplicationConfiguration
	QueueConfiguration() QueueConfiguration
	ClusterConfiguration() ClusterConfiguration
	TaskResourceConfiguration() TaskResourceConfiguration
	WhitelistConfiguration() WhitelistConfiguration
	RegistrationValidationConfiguration() RegistrationValidationConfiguration
	ClusterResourceConfiguration() ClusterResourceConfiguration
	NamespaceMappingConfiguration() NamespaceMappingConfiguration
	QualityOfServiceConfiguration() QualityOfServiceConfiguration
	ClusterPoolAssignmentConfiguration() ClusterPoolAssignmentConfiguration
}
