package dask

import (
	pluginsConfig "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/config"
	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/logs"
)

//go:generate pflags Config --default-var=defaultConfig

var (
	defaultConfig = Config{
		Logs: logs.DefaultConfig,
	}

	configSection = pluginsConfig.MustRegisterSubSection("dask", &defaultConfig)
)

// Config is config for 'dask' plugin
type Config struct {
	Logs logs.LogConfig `json:"logs,omitempty"`

	// ClusterName is the per-cluster identity ("dogfood-1", "dogfood-gcp-dp", etc.)
	// that the leaseworker is running in. When set, the dask plugin starts the
	// scheduler with `--dashboard-prefix`, baked from this name + the task's
	// execution identity, so the Bokeh dashboard's internal links resolve through
	// the union dataproxy reverse-proxy chain. Without it, dashboard tabs render
	// against the page origin's root path and produce empty bodies.
	ClusterName string `json:"clusterName,omitempty" pflag:",K8s cluster identity embedded in the Dask dashboard URL prefix; set to operator's cluster name."`
}

func GetConfig() *Config {
	return configSection.GetConfig().(*Config)
}

func SetConfig(cfg *Config) error {
	return configSection.SetConfig(cfg)
}
