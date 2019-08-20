package config

import (
	"k8s.io/api/core/v1"

	"github.com/lyft/flyteplugins/go/tasks/v1/config"
)

const k8sPluginConfigSectionKey = "k8s"
const defaultCpuRequest = "1000m"
const defaultMemoryRequest = "1024Mi"

var (
	defaultK8sConfig = K8sPluginConfig{
		DefaultAnnotations: map[string]string{
			"cluster-autoscaler.kubernetes.io/safe-to-evict": "false",
		},
	}

	// Top level k8s plugin config section. If you are a plugin developer writing a k8s plugin,
	// register your config section as a subsection to this.
	K8sPluginConfigSection = config.MustRegisterSubSection(k8sPluginConfigSectionKey, &defaultK8sConfig)
)

// Top level k8s plugin config.
type K8sPluginConfig struct {
	// Boolean flag that indicates if a finalizer should be injected into every K8s resource launched
	InjectFinalizer bool `json:"inject-finalizer" pflag:",Instructs the plugin to inject a finalizer on startTask and remove it on task termination."`
	// Provide default annotations that should be added to K8s resource
	DefaultAnnotations map[string]string `json:"default-annotations" pflag:",Defines a set of default annotations to add to the produced pods."`
	// Provide default labels that should be added to K8s resource
	DefaultLabels map[string]string `json:"default-labels" pflag:",Defines a set of default labels to add to the produced pods."`
	// Provide additional environment variable pairs that plugin authors will provide to containers
	DefaultEnvVars map[string]string `json:"default-env-vars" pflag:",Additional environment variable that should be injected into every resource"`
	// Tolerations in the cluster that should be applied for a specific resource
	// Currently we support simple resource based tolerations only
	ResourceTolerations map[v1.ResourceName][]v1.Toleration `json:"resource-tolerations"`
	// default cpu requests for a container
	DefaultCpuRequest string `json:"default-cpus" pflag:",Defines a default value for cpu for containers if not specified."`
	// default memory requests for a container
	DefaultMemoryRequest string `json:"default-memory" pflag:",Defines a default value for memory for containers if not specified."`
}

// Retrieves the current k8s plugin config or default.
func GetK8sPluginConfig() *K8sPluginConfig {
	pluginsConfig := K8sPluginConfigSection.GetConfig().(*K8sPluginConfig)
	if pluginsConfig.DefaultMemoryRequest == "" {
		pluginsConfig.DefaultMemoryRequest = defaultMemoryRequest
	}
	if pluginsConfig.DefaultCpuRequest == "" {
		pluginsConfig.DefaultCpuRequest = defaultCpuRequest
	}
	return pluginsConfig
}

// [FOR TESTING ONLY] Sets current value for the config.
func SetK8sPluginConfig(cfg *K8sPluginConfig) error {
	return K8sPluginConfigSection.SetConfig(cfg)
}
