package namespaceutils

import "github.com/flyteorg/flyte/v2/flytestdlib/config"

//go:generate pflags Config --default-var=defaultConfig

const configSectionKey = "namespaceMapping"

var (
	defaultConfig = &Config{
		CreateNamespaces: true,
	}
	configSection = config.MustRegisterSection(configSectionKey, defaultConfig)
)

// Config maps each project-domain to the Kubernetes namespace its runs execute in.
type Config struct {
	Template         string `json:"template" pflag:",Namespace template for a run's TaskActions and pods such as {{ project }}-{{ domain }}. Empty uses actions.kubernetes.namespace."`
	CreateNamespaces bool   `json:"createNamespaces" pflag:",Create a mapped namespace when it does not exist. When false it must already exist."`
}

// GetConfig returns the namespace mapping configuration.
func GetConfig() *Config {
	return configSection.GetConfig().(*Config)
}
