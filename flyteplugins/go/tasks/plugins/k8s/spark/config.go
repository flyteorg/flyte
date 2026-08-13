package spark

import (
	"time"

	pluginsConfig "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/config"
	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/logs"
	"github.com/flyteorg/flyte/v2/flytestdlib/config"
)

//go:generate pflags Config --default-var=defaultConfig

var (
	defaultConfig = &Config{
		EnablePodTemplate:           true,
		PodTemplateDetectionTimeout: config.Duration{Duration: 30 * time.Second},
		SparkVersion:                minPodTemplateSparkVersion,
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
	EnablePodTemplate     bool              `json:"enable-pod-template" pflag:"-,Pass the full pod spec through as the driver/executor pod template on clusters whose SparkApplication CRD supports it. Disable as a kill switch."`
	// PodTemplateDetectionTimeout bounds the one-shot SparkApplication CRD read behind
	// enable-pod-template. On timeout the probe reports false and the plugin stays on the
	// legacy fields.
	PodTemplateDetectionTimeout config.Duration `json:"pod-template-detection-timeout" pflag:"-,Timeout for the one-time SparkApplication CRD schema read that detects pod template support."`
	// SparkVersion is declared on SparkApplications that carry a pod template. The plugin
	// cannot inspect the image, so this defaults to the minimum the pod template feature
	// requires; set it to the Spark version your images actually ship.
	SparkVersion string `json:"spark-version" pflag:"-,Value for SparkApplication spec.sparkVersion on applications that carry a driver/executor pod template."`
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
func setSparkConfig(cfg *Config) error { //nolint: unused
	return sparkConfigSection.SetConfig(cfg)
}
