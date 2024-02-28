package runtime

import (
	"github.com/flyteorg/flyte/cacheservice/pkg/runtime/configs"
	"github.com/flyteorg/flyte/flytestdlib/config"
)

const cacheservice = "cacheservice"

var cacheserviceConfig = config.MustRegisterSection(cacheservice, &configs.CacheServiceConfig{})

// Defines the interface to return top-level config structs necessary to start up a cacheservice application.
type ApplicationConfiguration interface {
	GetCacheServiceConfig() configs.CacheServiceConfig
}

type ApplicationConfigurationProvider struct{}

func (p *ApplicationConfigurationProvider) GetCacheServiceConfig() configs.CacheServiceConfig {
	return *cacheserviceConfig.GetConfig().(*configs.CacheServiceConfig)
}

func NewApplicationConfigurationProvider() ApplicationConfiguration {
	return &ApplicationConfigurationProvider{}
}
