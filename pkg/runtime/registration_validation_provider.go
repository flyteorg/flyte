// Interface for configurable values used in entity registration and validation
package runtime

import (
	"context"

	"github.com/flyteorg/flyteadmin/pkg/runtime/interfaces"
	"github.com/flyteorg/flytestdlib/config"
	"github.com/flyteorg/flytestdlib/logger"
)

const registration = "registration"

var registrationValidationConfig = config.MustRegisterSection(registration, &interfaces.RegistrationValidationConfig{})

// Implementation of an interfaces.TaskResourceConfiguration
type RegistrationValidationProvider struct{}

func (p *RegistrationValidationProvider) GetWorkflowNodeLimit() int {
	if registrationValidationConfig != nil {
		return registrationValidationConfig.GetConfig().(*interfaces.RegistrationValidationConfig).MaxWorkflowNodes
	}
	logger.Warning(context.Background(), "failed to find max workflow node values in config. Returning 0")
	return 0
}

func (p *RegistrationValidationProvider) GetMaxLabelEntries() int {
	if registrationValidationConfig != nil {
		return registrationValidationConfig.GetConfig().(*interfaces.RegistrationValidationConfig).MaxLabelEntries
	}
	logger.Warning(context.Background(), "failed to find max label entries in config. Returning 0")
	return 0
}

func (p *RegistrationValidationProvider) GetMaxAnnotationEntries() int {
	if registrationValidationConfig != nil {
		return registrationValidationConfig.GetConfig().(*interfaces.RegistrationValidationConfig).MaxAnnotationEntries
	}
	logger.Warning(context.Background(), "failed to find max annotation entries in config. Returning 0")
	return 0
}

func (p *RegistrationValidationProvider) GetWorkflowSizeLimit() string {
	if registrationValidationConfig != nil {
		return registrationValidationConfig.GetConfig().(*interfaces.RegistrationValidationConfig).WorkflowSizeLimit
	}
	logger.Warning(context.Background(), "failed to findworkflow size limit in config. Returning ''")
	return ""
}

func NewRegistrationValidationProvider() interfaces.RegistrationValidationConfiguration {
	return &RegistrationValidationProvider{}
}
