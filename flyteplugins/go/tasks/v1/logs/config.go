package logs

import "github.com/lyft/flyteplugins/go/tasks/v1/config"

//go:generate pflags LogConfig

// Log plugins configs
type LogConfig struct {
	IsCloudwatchEnabled bool   `json:"cloudwatch-enabled" pflag:",Enable Cloudwatch Logging"`
	CloudwatchRegion    string `json:"cloudwatch-region" pflag:",AWS region in which Cloudwatch logs are stored."`
	CloudwatchLogGroup  string `json:"cloudwatch-log-group" pflag:",Log group to which streams are associated."`

	IsKubernetesEnabled bool   `json:"kubernetes-enabled" pflag:",Enable Kubernetes Logging"`
	KubernetesURL       string `json:"kubernetes-url" pflag:",Console URL for Kubernetes logs"`

	IsStackDriverEnabled       bool   `json:"stackdriver-enabled" pflag:",Enable Log-links to stackdriver"`
	GCPProjectName             string `json:"gcp-project" pflag:",Name of the project in GCP"`
	StackdriverLogResourceName string `json:"stackdriver-logresourcename" pflag:",Name of the logresource in stackdriver"`
}

var (
	logConfigSection = config.MustRegisterSubSection("logs", &LogConfig{})
)

func GetLogConfig() *LogConfig {
	return logConfigSection.GetConfig().(*LogConfig)
}

// This method should be used for unit testing only
func SetLogConfig(logConfig *LogConfig) error {
	return logConfigSection.SetConfig(logConfig)
}

