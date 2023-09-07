package spark

import (
	pluginsConfig "github.com/flyteorg/flyteplugins/go/tasks/config"
	"github.com/flyteorg/flyteplugins/go/tasks/logs"
)

//go:generate pflags Config --default-var=defaultConfig

var (
	defaultConfig = &Config{
		LogConfig: LogConfig{
			Mixed: logs.LogConfig{
				IsKubernetesEnabled:   true,
				KubernetesTemplateURI: "http://localhost:30082/#!/log/{{ .namespace }}/{{ .podName }}/pod?namespace={{ .namespace }}",
			},
		},
	}

	sparkConfigSection = pluginsConfig.MustRegisterSubSection("spark", defaultConfig)
)

// Spark-specific configs
type Config struct {
	DefaultSparkConfig    map[string]string `json:"spark-config-default" pflag:"-,Key value pairs of default spark configuration that should be applied to every SparkJob"`
	SparkHistoryServerURL string            `json:"spark-history-server-url" pflag:",URL for SparkHistory Server that each job will publish the execution history to."`
	Features              []Feature         `json:"features" pflag:"-,List of optional features supported."`
	LogConfig             LogConfig         `json:"logs" pflag:",Config for log links for spark applications."`
}

type LogConfig struct {
	Mixed   logs.LogConfig `json:"mixed" pflag:",Defines the log config that's not split into user/system."`
	User    logs.LogConfig `json:"user" pflag:",Defines the log config for user logs."`
	System  logs.LogConfig `json:"system" pflag:",Defines the log config for system logs."`
	AllUser logs.LogConfig `json:"all-user" pflag:",All user logs across driver and executors."`
}

// Optional feature with name and corresponding spark-config to use.
type Feature struct {
	Name        string            `json:"name"`
	SparkConfig map[string]string `json:"spark-config"`
}

func GetSparkConfig() *Config {
	return sparkConfigSection.GetConfig().(*Config)
}

// This method should be used for unit testing only
func setSparkConfig(cfg *Config) error {
	return sparkConfigSection.SetConfig(cfg)
}
