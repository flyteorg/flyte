// Package config contains configuration for the flytek8s module - which is global configuration for all Flyte K8s interactions.
// This config is under the subsection `k8s` and registered under the Plugin config
// All K8s based plugins can optionally use the flytek8s module and this configuration allows controlling the defaults
// For example if for every container execution if some default Environment Variables or Annotations should be used, then they can be configured here
// An important configuration is ResourceTolerations that are applied to every container execution that needs some resource on the cluster
package config

import (
	"time"

	"k8s.io/apimachinery/pkg/api/resource"

	config2 "github.com/flyteorg/flytestdlib/config"
	v1 "k8s.io/api/core/v1"

	"github.com/flyteorg/flyteplugins/go/tasks/config"
)

//go:generate pflags K8sPluginConfig --default-var=defaultK8sConfig

const k8sPluginConfigSectionKey = "k8s"

// ResourceNvidiaGPU is the name of the Nvidia GPU resource.
// Copied from: k8s.io/autoscaler/cluster-autoscaler/utils/gpu/gpu.go
const ResourceNvidiaGPU v1.ResourceName = "nvidia.com/gpu"

var defaultCPURequest = resource.MustParse("1000m")
var defaultMemoryRequest = resource.MustParse("1024Mi")

var (
	defaultK8sConfig = K8sPluginConfig{
		DefaultAnnotations: map[string]string{
			"cluster-autoscaler.kubernetes.io/safe-to-evict": "false",
		},
		CoPilot: FlyteCoPilotConfig{
			NamePrefix:           "flyte-copilot-",
			Image:                "cr.flyte.org/flyteorg/flytecopilot:v0.0.15",
			DefaultInputDataPath: "/var/flyte/inputs",
			InputVolumeName:      "flyte-inputs",
			DefaultOutputPath:    "/var/flyte/outputs",
			OutputVolumeName:     "flyte-outputs",
			CPU:                  "500m",
			Memory:               "128Mi",
			StartTimeout: config2.Duration{
				Duration: time.Second * 100,
			},
		},
		DefaultCPURequest:    defaultCPURequest,
		DefaultMemoryRequest: defaultMemoryRequest,
		CreateContainerErrorGracePeriod: config2.Duration{
			Duration: time.Minute * 3,
		},
		ImagePullBackoffGracePeriod: config2.Duration{
			Duration: time.Minute * 3,
		},
		GpuResourceName: ResourceNvidiaGPU,
		DefaultPodTemplateResync: config2.Duration{
			Duration: 30 * time.Second,
		},
	}

	// K8sPluginConfigSection provides a singular top level config section for all plugins.
	// If you are a plugin developer writing a k8s plugin, register your config section as a subsection to this.
	K8sPluginConfigSection = config.MustRegisterSubSection(k8sPluginConfigSectionKey, &defaultK8sConfig)
)

// K8sPluginConfig should be used to configure per-pod defaults for the entire platform. This allows adding global defaults
// for pods that are being launched. For example, default annotations, labels, if a finalizer should be injected,
// if taints/tolerations should be used for certain resource types etc.
type K8sPluginConfig struct {
	// InjectFinalizer is a boolean flag that indicates if a finalizer should be injected into every K8s resource launched
	InjectFinalizer bool `json:"inject-finalizer" pflag:",Instructs the plugin to inject a finalizer on startTask and remove it on task termination."`

	// -------------------------------------------------------------------------------------------------------------
	// Default Configurations to be applied to all Pods launched by Flyte. These are always applied to every Pod.
	// Thus if a Pod is interruptible, it will have the default + interruptible tolerations

	// Provide default annotations that should be added to K8s resource
	DefaultAnnotations map[string]string `json:"default-annotations" pflag:"-,Defines a set of default annotations to add to the produced pods."`
	// Provide default labels that should be added to K8s resource
	DefaultLabels map[string]string `json:"default-labels" pflag:"-,Defines a set of default labels to add to the produced pods."`
	// Provide additional environment variable pairs that plugin authors will provide to containers
	DefaultEnvVars map[string]string `json:"default-env-vars" pflag:"-,Additional environment variable that should be injected into every resource"`
	// Provide additional environment variable pairs whose values resolve from the plugin's execution environment.
	DefaultEnvVarsFromEnv map[string]string `json:"default-env-vars-from-env" pflag:"-,Additional environment variable that should be injected into every resource"`

	// default cpu requests for a container
	DefaultCPURequest resource.Quantity `json:"default-cpus" pflag:",Defines a default value for cpu for containers if not specified."`
	// default memory requests for a container
	DefaultMemoryRequest resource.Quantity `json:"default-memory" pflag:",Defines a default value for memory for containers if not specified."`

	// Default Tolerations that will be added to every Pod that is created by Flyte. These can be used in heterogenous clusters, where one wishes to keep all pods created by Flyte on a separate
	// set of nodes.
	DefaultTolerations []v1.Toleration `json:"default-tolerations"  pflag:"-,Tolerations to be applied for every node that is launched by Flyte. Useful in non dedicated flyte clusters"`
	// Default Node Selector Labels for pods. These NodeSelector labels are added to all pods, created by Flyte, unless they are marked as interruptible (default of interruptible are different).
	DefaultNodeSelector map[string]string `json:"default-node-selector" pflag:"-,Defines a set of node selector labels to add to the all pods launched by Flyte. Useful in non dedicated Flyte clusters"`
	// Default Affinity that is applied to every pod that Flyte launches
	DefaultAffinity *v1.Affinity `json:"default-affinity,omitempty" pflag:"-,Defines default Affinity to be added for every Pod launched by Flyte. Useful in non dedicated Flyte clusters"`

	// Default scheduler that should be used for all pods or CRD that accept Scheduler name.
	SchedulerName string `json:"scheduler-name" pflag:",Defines scheduler name."`

	// -----------------------------------------------------------------
	// Special tolerations and node selector for Interruptible tasks. This allows scheduling interruptible tasks onto specific hardward

	// Tolerations for interruptible k8s pods: These tolerations are added to the pods that can tolerate getting evicted from a node. We
	// can leverage this for better bin-packing and using low-reliability cheaper machines.
	InterruptibleTolerations []v1.Toleration `json:"interruptible-tolerations"  pflag:"-,Tolerations to be applied for interruptible pods"`
	// Node Selector Labels for interruptible pods: Similar to InterruptibleTolerations, these node selector labels are added for pods that can tolerate
	// eviction.
	// Deprecated: Please use InterruptibleNodeSelectorRequirement/NonInterruptibleNodeSelectorRequirement
	InterruptibleNodeSelector map[string]string `json:"interruptible-node-selector" pflag:"-,Defines a set of node selector labels to add to the interruptible pods."`
	// Node Selector Requirements to be added to interruptible and non-interruptible
	// pods respectively
	InterruptibleNodeSelectorRequirement    *v1.NodeSelectorRequirement `json:"interruptible-node-selector-requirement" pflag:"-,Node selector requirement to add to interruptible pods"`
	NonInterruptibleNodeSelectorRequirement *v1.NodeSelectorRequirement `json:"non-interruptible-node-selector-requirement" pflag:"-,Node selector requirement to add to non-interruptible pods"`

	// ----------------------------------------------------------------------
	// Specific tolerations that are added for certain resources. Useful for maintaining gpu resources separate in the cluster

	// Tolerations in the cluster that should be applied for a specific resource
	// Currently we support simple resource based tolerations only
	ResourceTolerations map[v1.ResourceName][]v1.Toleration `json:"resource-tolerations"  pflag:"-,Default tolerations to be applied for resource of type 'key'"`

	// Flyte CoPilot Configuration
	CoPilot FlyteCoPilotConfig `json:"co-pilot" pflag:",Co-Pilot Configuration"`

	// DeleteResourceOnFinalize instructs the system to delete the resource on finalize. This ensures that no resources
	// are kept around (potentially consuming cluster resources). This, however, will cause k8s log links to expire as
	// soon as the resource is finalized.
	DeleteResourceOnFinalize bool `json:"delete-resource-on-finalize" pflag:",Instructs the system to delete the resource upon successful execution of a k8s pod rather than have the k8s garbage collector clean it up. This ensures that no resources are kept around (potentially consuming cluster resources). This, however, will cause k8s log links to expire as soon as the resource is finalized."`

	// Time to wait for transient CreateContainerError errors to be resolved. If the
	// error persists past this grace period, it will be inferred to be a permanent
	// one, and the corresponding task marked as failed
	CreateContainerErrorGracePeriod config2.Duration `json:"create-container-error-grace-period" pflag:"-,Time to wait for transient CreateContainerError errors to be resolved."`

	// Time to wait for transient ImagePullBackoff errors to be resolved. If the
	// error persists past this grace period, it will be inferred to be a permanent
	// one, and the corresponding task marked as failed
	ImagePullBackoffGracePeriod config2.Duration `json:"image-pull-backoff-grace-period" pflag:"-,Time to wait for transient ImagePullBackoff errors to be resolved."`

	// The name of the GPU resource to use when the task resource requests GPUs.
	GpuResourceName v1.ResourceName `json:"gpu-resource-name" pflag:"-,The name of the GPU resource to use when the task resource requests GPUs."`

	// DefaultPodSecurityContext provides a default pod security context that should be applied for every pod that is launched by FlytePropeller. This may not be applicable to all plugins. For
	// downstream plugins - i.e. TensorflowOperators may not support setting this, but Spark does.
	DefaultPodSecurityContext *v1.PodSecurityContext `json:"default-pod-security-context" pflag:"-,Optionally specify any default pod security context that should be applied to every Pod launched by FlytePropeller."`

	// DefaultSecurityContext provides a default container security context that should be applied for the primary container launched and created by FlytePropeller. This may not be applicable to all plugins. For
	//	// downstream plugins - i.e. TensorflowOperators may not support setting this, but Spark does.
	DefaultSecurityContext *v1.SecurityContext `json:"default-security-context" pflag:"-,Optionally specify a default security context that should be applied to every container launched/created by FlytePropeller. This will not be applied to plugins that do not support it or to user supplied containers in pod tasks."`

	// EnableHostNetworkingPod is a binary switch to enable `hostNetwork: true` for all pods launched by Flyte.
	// Refer to - https://kubernetes.io/docs/concepts/policy/pod-security-policy/#host-namespaces.
	// As a follow up, the default pod configurations will now be adjusted using podTemplates per namespace
	EnableHostNetworkingPod *bool `json:"enable-host-networking-pod" pflag:"-,If true, will schedule all pods with hostNetwork: true."`

	// DefaultPodDNSConfig provides a default pod DNS config that that should be applied for the primary container launched and created by FlytePropeller. This may not be applicable to all plugins. For
	//	// downstream plugins - i.e. TensorflowOperators may not support setting this.
	DefaultPodDNSConfig *v1.PodDNSConfig `json:"default-pod-dns-config" pflag:"-,Optionally specify a default DNS config that should be applied to every Pod launched by FlytePropeller."`

	// DefaultPodTemplateName that serves as the base PodTemplate for all k8s pods (including
	// individual containers) that are creating by FlytePropeller.
	DefaultPodTemplateName string `json:"default-pod-template-name" pflag:",Name of the PodTemplate to use as the base for all k8s pods created by FlytePropeller."`

	// DefaultPodTemplateResync defines the frequency at which the k8s informer resyncs the default
	// pod template resources.
	DefaultPodTemplateResync config2.Duration `json:"default-pod-template-resync" pflag:",Frequency of resyncing default pod templates"`
}

// FlyteCoPilotConfig specifies configuration for the Flyte CoPilot system. FlyteCoPilot, allows running flytekit-less containers
// in K8s, where the IO is managed by the FlyteCoPilot sidecar process.
type FlyteCoPilotConfig struct {
	// Co-pilot sidecar container name
	NamePrefix string `json:"name" pflag:",Flyte co-pilot sidecar container name prefix. (additional bits will be added after this)"`
	// Docker image FQN where co-pilot binary is installed
	Image string `json:"image" pflag:",Flyte co-pilot Docker Image FQN"`
	// Default Input Path for every task execution that uses co-pilot. This is used only if a input path is not provided by the user and inputs are required for the task
	DefaultInputDataPath string `json:"default-input-path" pflag:",Default path where the volume should be mounted"`
	// Default Output Path for every task execution that uses co-pilot. This is used only if a output path is not provided by the user and outputs are required for the task
	DefaultOutputPath string `json:"default-output-path" pflag:",Default path where the volume should be mounted"`
	// Name of the input volume
	InputVolumeName string `json:"input-vol-name" pflag:",Name of the data volume that is created for storing inputs"`
	// Name of the output volume
	OutputVolumeName string `json:"output-vol-name" pflag:",Name of the data volume that is created for storing outputs"`
	// Time for which the sidecar container should wait after starting up, for the primary process to appear. If it does not show up in this time
	// the process will be assumed to be dead or in a terminal condition and will trigger an abort.
	StartTimeout config2.Duration `json:"start-timeout" pflag:"-,Time for which the sidecar should wait on startup before assuming the primary container to have failed startup."`
	// Resources for CoPilot Containers
	CPU     string `json:"cpu" pflag:",Used to set cpu for co-pilot containers"`
	Memory  string `json:"memory" pflag:",Used to set memory for co-pilot containers"`
	Storage string `json:"storage" pflag:",Default storage limit for individual inputs / outputs"`
}

// GetK8sPluginConfig retrieves the current k8s plugin config or default.
func GetK8sPluginConfig() *K8sPluginConfig {
	return K8sPluginConfigSection.GetConfig().(*K8sPluginConfig)
}

// SetK8sPluginConfig should be used for TESTING ONLY, It Sets current value for the config.
func SetK8sPluginConfig(cfg *K8sPluginConfig) error {
	return K8sPluginConfigSection.SetConfig(cfg)
}
