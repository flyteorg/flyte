package runtime

import (
	"github.com/flyteorg/flyte/flyteadmin/pkg/runtime/interfaces"
	"github.com/flyteorg/flyte/flytestdlib/config"
)

const externalResourcesKey = "externalResources"

var externalResourcesConfig = config.MustRegisterSection(externalResourcesKey, &interfaces.ExternalResourceConfig{})

// ExternalResourcesConfigurationProvider Implementation of an interfaces.ExternalResourceConfiguration
type ExternalResourcesConfigurationProvider struct{}

func (e ExternalResourcesConfigurationProvider) GetConnections() map[string]interfaces.Connection {
	return externalResourcesConfig.GetConfig().(*interfaces.ExternalResourceConfig).Connections
}

func NewExternalResourceConfigurationProvider() interfaces.ExternalResourceConfiguration {
	return &ExternalResourcesConfigurationProvider{}
}
