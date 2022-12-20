// Interface for configurable values used in entity registration and validation
package runtime

import (
	"github.com/flyteorg/flyteadmin/pkg/runtime/interfaces"
	"github.com/flyteorg/flytestdlib/config"
)

const registration = "registration"

var registrationValidationConfig = config.MustRegisterSection(registration, &interfaces.RegistrationValidationConfig{
	MaxWorkflowNodes: 100,
})

// Implementation of an interfaces.TaskResourceConfiguration
type RegistrationValidationProvider struct{}

func (p *RegistrationValidationProvider) GetWorkflowNodeLimit() int {
	return registrationValidationConfig.GetConfig().(*interfaces.RegistrationValidationConfig).MaxWorkflowNodes
}

func (p *RegistrationValidationProvider) GetMaxLabelEntries() int {
	return registrationValidationConfig.GetConfig().(*interfaces.RegistrationValidationConfig).MaxLabelEntries
}

func (p *RegistrationValidationProvider) GetMaxAnnotationEntries() int {
	return registrationValidationConfig.GetConfig().(*interfaces.RegistrationValidationConfig).MaxAnnotationEntries
}

func (p *RegistrationValidationProvider) GetWorkflowSizeLimit() string {
	return registrationValidationConfig.GetConfig().(*interfaces.RegistrationValidationConfig).WorkflowSizeLimit
}

func NewRegistrationValidationProvider() interfaces.RegistrationValidationConfiguration {
	return &RegistrationValidationProvider{}
}
