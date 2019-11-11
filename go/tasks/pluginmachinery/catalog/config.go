package catalog

import (
	"github.com/lyft/flyteplugins/go/tasks/config"
	"github.com/lyft/flyteplugins/go/tasks/pluginmachinery/workqueue"
)

//go:generate pflags Config --default-var=defaultConfig

var cfgSection = config.MustRegisterSubSection("catalogCache", defaultConfig)

type Config struct {
	ReaderWorkqueueConfig workqueue.Config
	WriterWorkqueueConfig workqueue.Config
}

var defaultConfig = &Config{
	ReaderWorkqueueConfig: workqueue.Config{
		MaxRetries:         3,
		Workers:            10,
		IndexCacheMaxItems: 1000,
	},
	WriterWorkqueueConfig: workqueue.Config{
		MaxRetries:         3,
		Workers:            10,
		IndexCacheMaxItems: 1000,
	},
}

func GetConfig() *Config {
	return cfgSection.GetConfig().(*Config)
}
