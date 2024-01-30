package logs

import (
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/config"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/tasklog"
)

//go:generate pflags LogConfig --default-var=DefaultConfig

// LogConfig encapsulates plugins' log configs
type LogConfig struct {
	IsCloudwatchEnabled bool `json:"cloudwatch-enabled" pflag:",Enable Cloudwatch Logging"`
	// Deprecated: Please use CloudwatchTemplateURI
	CloudwatchRegion string `json:"cloudwatch-region" pflag:",AWS region in which Cloudwatch logs are stored."`
	// Deprecated: Please use CloudwatchTemplateURI
	CloudwatchLogGroup    string              `json:"cloudwatch-log-group" pflag:",Log group to which streams are associated."`
	CloudwatchTemplateURI tasklog.TemplateURI `json:"cloudwatch-template-uri" pflag:",Template Uri to use when building cloudwatch log links"`

	IsKubernetesEnabled bool `json:"kubernetes-enabled" pflag:",Enable Kubernetes Logging"`
	// Deprecated: Please use KubernetesTemplateURI
	KubernetesURL         string              `json:"kubernetes-url" pflag:",Console URL for Kubernetes logs"`
	KubernetesTemplateURI tasklog.TemplateURI `json:"kubernetes-template-uri" pflag:",Template Uri to use when building kubernetes log links"`

	IsStackDriverEnabled bool `json:"stackdriver-enabled" pflag:",Enable Log-links to stackdriver"`
	// Deprecated: Please use StackDriverTemplateURI
	GCPProjectName string `json:"gcp-project" pflag:",Name of the project in GCP"`
	// Deprecated: Please use StackDriverTemplateURI
	StackdriverLogResourceName string              `json:"stackdriver-logresourcename" pflag:",Name of the logresource in stackdriver"`
	StackDriverTemplateURI     tasklog.TemplateURI `json:"stackdriver-template-uri" pflag:",Template Uri to use when building stackdriver log links"`

	DynamicLogLinks map[string]tasklog.TemplateLogPlugin `json:"dynamic-log-links" pflag:"-,Map of dynamic log links"`

	Templates []tasklog.TemplateLogPlugin `json:"templates" pflag:"-,"`
}

var (
	DefaultConfig = LogConfig{
		IsKubernetesEnabled:   true,
		KubernetesTemplateURI: "http://localhost:30082/#!/log/{{ .namespace }}/{{ .podName }}/pod?namespace={{ .namespace }}",
	}

	logConfigSection = config.MustRegisterSubSection("logs", &DefaultConfig)
)

func GetLogConfig() *LogConfig {
	return logConfigSection.GetConfig().(*LogConfig)
}

// SetLogConfig should be used for unit testing only
func SetLogConfig(logConfig *LogConfig) error {
	return logConfigSection.SetConfig(logConfig)
}
