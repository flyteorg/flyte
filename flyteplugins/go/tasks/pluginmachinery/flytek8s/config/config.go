package config

import (
	v1 "k8s.io/api/core/v1"

	"github.com/lyft/flyteplugins/go/tasks/config"
)

//go:generate pflags K8sPluginConfig

const k8sPluginConfigSectionKey = "k8s"
const defaultCPURequest = "1000m"
const defaultMemoryRequest = "1024Mi"

var (
	defaultK8sConfig = K8sPluginConfig{
		DefaultAnnotations: map[string]string{
			"cluster-autoscaler.kubernetes.io/safe-to-evict": "false",
		},
	}

	// K8sPluginConfigSection provides a singular top level config section for all plugins.
	// If you are a plugin developer writing a k8s plugin, register your config section as a subsection to this.
	K8sPluginConfigSection = config.MustRegisterSubSection(k8sPluginConfigSectionKey, &defaultK8sConfig)
)

// Top level k8s plugin config.
type K8sPluginConfig struct {
	// Boolean flag that indicates if a finalizer should be injected into every K8s resource launched
	InjectFinalizer bool `json:"inject-finalizer" pflag:",Instructs the plugin to inject a finalizer on startTask and remove it on task termination."`
	// Provide default annotations that should be added to K8s resource
	DefaultAnnotations map[string]string `json:"default-annotations" pflag:"-,Defines a set of default annotations to add to the produced pods."`
	// Provide default labels that should be added to K8s resource
	DefaultLabels map[string]string `json:"default-labels" pflag:"-,Defines a set of default labels to add to the produced pods."`
	// Provide additional environment variable pairs that plugin authors will provide to containers
	DefaultEnvVars map[string]string `json:"default-env-vars" pflag:"-,Additional environment variable that should be injected into every resource"`
	// Provide additional environment variable pairs whose values resolve from the plugin's execution environment.
	DefaultEnvVarsFromEnv map[string]string `json:"default-env-vars-from-env" pflag:"-,Additional environment variable that should be injected into every resource"`
	// Tolerations in the cluster that should be applied for a specific resource
	// Currently we support simple resource based tolerations only
	ResourceTolerations map[v1.ResourceName][]v1.Toleration `json:"resource-tolerations"  pflag:"-,Default tolerations to be applied for resource of type 'key'"`
	// default cpu requests for a container
	DefaultCPURequest string `json:"default-cpus" pflag:",Defines a default value for cpu for containers if not specified."`
	// default memory requests for a container
	DefaultMemoryRequest string `json:"default-memory" pflag:",Defines a default value for memory for containers if not specified."`
	// Tolerations for interruptible k8s pods: These tolerations are added to the pods that can tolerate getting evicted from a node. We
	// can leverage this for better bin-packing and using low-reliability cheaper machines.
	InterruptibleTolerations []v1.Toleration `json:"interruptible-tolerations"  pflag:"-,Tolerations to be applied for interruptible pods"`
	// Node Selector Labels for interruptible pods: Similar to InterruptibleTolerations, these node selector labels are added for pods that can tolerate
	// eviction.
	InterruptibleNodeSelector map[string]string `json:"interruptible-node-selector" pflag:"-,Defines a set of node selector labels to add to the interruptible pods."`
	// Scheduler name.
	SchedulerName string `json:"scheduler-name" pflag:",Defines scheduler name."`
}

// Retrieves the current k8s plugin config or default.
func GetK8sPluginConfig() *K8sPluginConfig {
	pluginsConfig := K8sPluginConfigSection.GetConfig().(*K8sPluginConfig)
	if pluginsConfig.DefaultMemoryRequest == "" {
		pluginsConfig.DefaultMemoryRequest = defaultMemoryRequest
	}
	if pluginsConfig.DefaultCPURequest == "" {
		pluginsConfig.DefaultCPURequest = defaultCPURequest
	}
	return pluginsConfig
}

// [FOR TESTING ONLY] Sets current value for the config.
func SetK8sPluginConfig(cfg *K8sPluginConfig) error {
	return K8sPluginConfigSection.SetConfig(cfg)
}
