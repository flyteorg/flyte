package runtime

import (
	"github.com/flyteorg/flyte/flyteadmin/pkg/runtime/interfaces"
	"github.com/flyteorg/flyte/flytestdlib/config"
)

const externalResourcesKey = "externalResources"

var externalResourcesConfig = config.MustRegisterSection(externalResourcesKey, &interfaces.ExternalResourceConfig{})

// ExternalResourcesConfigurationProvider Implementation of an interfaces.ExternalResourceConfiguration
type ExternalResourcesConfigurationProvider struct{}

func (e ExternalResourcesConfigurationProvider) GetExternalResource() interfaces.ExternalResource {
	return externalResourcesConfig.GetConfig().(*interfaces.ExternalResourceConfig).ExternalResource
}

func NewExternalResourceConfigurationProvider() interfaces.ExternalResourceConfiguration {
	return &ExternalResourcesConfigurationProvider{}
}
