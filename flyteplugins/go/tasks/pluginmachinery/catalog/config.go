package catalog

import (
	"github.com/flyteorg/flyteplugins/go/tasks/config"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/workqueue"
)

//go:generate pflags Config --default-var=defaultConfig

var cfgSection = config.MustRegisterSubSection("catalogCache", defaultConfig)

type Config struct {
	ReaderWorkqueueConfig workqueue.Config `json:"reader" pflag:",Catalog reader workqueue config. Make sure the index cache must be big enough to accommodate the biggest array task allowed to run on the system."`
	WriterWorkqueueConfig workqueue.Config `json:"writer" pflag:",Catalog writer workqueue config. Make sure the index cache must be big enough to accommodate the biggest array task allowed to run on the system."`
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
