package catalog

import (
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/config"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/workqueue"
)

//go:generate pflags Config --default-var=defaultConfig

var cfgSection = config.MustRegisterSubSection("catalogCache", defaultConfig)

type Config struct {
	ReaderWorkqueueConfig workqueue.Config `json:"reader" pflag:",Catalog reader workqueue config. Make sure the index cache must be big enough to accommodate the biggest array task allowed to run on the system."`
	WriterWorkqueueConfig workqueue.Config `json:"writer" pflag:",Catalog writer workqueue config. Make sure the index cache must be big enough to accommodate the biggest array task allowed to run on the system."`
	CacheKey              CacheKeyConfig   `json:"cacheKey" pflag:",Cache key configuration."`
}

type CacheKeyConfig struct {
	EnforceExecutionProjectDomain bool `json:"enforceExecutionProjectDomain" pflag:", Use execution project domain when computing the cache key. This means that even if you reference tasks/launchplans from a different project, cache keys will be computed based on the execution project domain instead."`
}

var defaultConfig = &Config{
	ReaderWorkqueueConfig: workqueue.Config{
		MaxRetries:         3,
		Workers:            10,
		IndexCacheMaxItems: 10000,
	},
	WriterWorkqueueConfig: workqueue.Config{
		MaxRetries:         3,
		Workers:            10,
		IndexCacheMaxItems: 10000,
	},
}

func GetConfig() *Config {
	return cfgSection.GetConfig().(*Config)
}
