package resources

import (
	"github.com/flyteorg/flyte/flyteadmin/pkg/manager/interfaces"
	repoInterface "github.com/flyteorg/flyte/flyteadmin/pkg/repositories/interfaces"
	runtimeInterfaces "github.com/flyteorg/flyte/flyteadmin/pkg/runtime/interfaces"
)

func ConfigureResourceManager(db repoInterface.Repository, config runtimeInterfaces.ApplicationConfiguration, configurationManager interfaces.ConfigurationInterface) interfaces.ResourceInterface {
	if config.GetTopLevelConfig().ResourceAttributesMode == runtimeInterfaces.ResourceAttributesModeConfiguration {
		return NewConfigurationResourceManager(db, config, configurationManager)
	}
	return &ResourceManager{
		db:     db,
		config: config,
	}
}
