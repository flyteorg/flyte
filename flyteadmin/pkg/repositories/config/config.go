package config

import (
	"github.com/flyteorg/flyte/flytestdlib/config"
)

const SectionKey = "matcheableResources"

type MatcheableResourcesConfig struct {
	ExecutionClusterLabels    ExecutionClusterLabelConfig     `json:"executionClusterLabels" pflag:", Defines the configuration for execution cluster labels."`
	TaskResourceAttributes    TaskResourceAttributesConfig    `json:"taskResourceAttributes" pflag:", Defines the configuration for task resource attributes."`
	ClusterResourceAttributes ClusterResourceAttributesConfig `json:"clusterResourceAttributes" pflag:", Defines the configuration for cluster resource attributes."`
	ExecutionQueueAttributes  ExecutionQueueAttributesConfig  `json:"executionQueueAttributes" pflag:", Defines the configuration for execution queue attributes."`
	PluginOverrides           PluginOverridesConfig           `json:"pluginOverrides" pflag:", Defines the configuration for plugin overrides."`
	WorkflowExecutionConfigs  WorkflowExecutionConfigConfig   `json:"workflowExecutionConfigs" pflag:", Defines the config for workflow execution configs."`
}

type TaskResourceAttributesConfig struct {
	Declarative bool                     `json:"declarative" pflag:", Whether the task resource attributes are managed declaratively, i.e. whether non-listed attributes are deleted."`
	Values      []TaskResourceAttributes `json:"values" pflag:", The list of task resource attributes to be managed."`
}

// Config matching admin.TaskResourceSpec
type TaskResourceSpec struct {
	Cpu              string `json:"cpu" pflag:", Number of CPUs."`
	Gpu              string `json:"gpu" pflag:", Number of GPUs."`
	Memory           string `json:"memory" pflag:", Amount of memory."`
	Storage          string `json:"storage" pflag:", Amount of storage."`
	EphemeralStorage string `json:"ephemeralStorage" pflag:", Amount of ephemeral storage."`
}

type TaskResourceAttributes struct {
	Project  string           `json:"project" pflag:", The project to which the task resource attribute is applied."`
	Domain   string           `json:"domain" pflag:", The domain to which the task resource attribute is applied."`
	Workflow string           `json:"workflow" pflag:", The workflow to which the task resource attribute is applied."`
	Defaults TaskResourceSpec `json:"defaults" pflag:", The default task resource spec."`
	Limits   TaskResourceSpec `json:"limits" pflag:", The task resource spec limits."`
}

type ExecutionClusterLabelConfig struct {
	Declarative bool                    `json:"declarative" pflag:", Whether the execution cluster labels are managed declaratively, i.e. whether non-listed labels are deleted."`
	Values      []ExecutionClusterLabel `json:"values" pflag:", The list of execution cluster labels to be managed."`
}

type ExecutionClusterLabel struct {
	Project  string `json:"project" pflag:", The project to which the execution cluster label is applied."`
	Domain   string `json:"domain" pflag:", The domain to which the execution cluster label is applied."`
	Workflow string `json:"workflow" pflag:", The workflow to which the execution cluster label is applied."`
	Label    string `json:"label" pflag:", The value of the execution cluster label."`
}

type ClusterResourceAttributesConfig struct {
	Declarative bool                        `json:"declarative" pflag:", Whether the cluster resource attributes are managed declaratively, i.e. whether non-listed attributes are deleted."`
	Values      []ClusterResourceAttributes `json:"values" pflag:", The list of cluster resource attributes to be managed."`
}

type ClusterResourceAttributes struct {
	Project    string            `json:"project" pflag:", The project to which the cluster resource attribute is applied."`
	Domain     string            `json:"domain" pflag:", The domain to which the cluster resource attribute is applied."`
	Workflow   string            `json:"workflow" pflag:", The workflow to which the cluster resource attribute is applied."`
	Attributes map[string]string `json:"attributes" pflag:", The cluster resource attributes."`
}

type ExecutionQueueAttributesConfig struct {
	Declarative bool                       `json:"declarative" pflag:", Whether the execution queue attributes are managed declaratively, i.e. whether non-listed attributes are deleted."`
	Values      []ExecutionQueueAttributes `json:"values" pflag:", The list of execution queue attributes to be managed."`
}

type ExecutionQueueAttributes struct {
	Project  string   `json:"project" pflag:", The project to which the execution queue attribute is applied."`
	Domain   string   `json:"domain" pflag:", The domain to which the execution queue attribute is applied."`
	Workflow string   `json:"workflow" pflag:", The workflow to which the execution queue attribute is applied."`
	Tags     []string `json:"tags" pflag:", The tags of the execution queue."`
}

type PluginOverridesConfig struct {
	Declarative bool              `json:"declarative" pflag:", Whether the plugin overrides are managed declaratively, i.e. whether non-listed overrides are deleted."`
	Values      []PluginOverrides `json:"values" pflag:", The list of plugin overrides to be managed."`
}

type PluginOverrides struct {
	Project  string `json:"project" pflag:", The project to which the plugin override is applied."`
	Domain   string `json:"domain" pflag:", The domain to which the plugin override is applied."`
	Workflow string `json:"workflow" pflag:", The workflow to which the plugin override is applied."`

	Overrides []PluginOverride `json:"overrides" pflag:", The list of plugin overrides."`
}

type PluginOverride struct {
	TaskType              string   `json:"task_type" pflag:", Task type for which to apply plugin implementation overrides."`
	PluginId              []string `json:"plugin_id" pflag:", Plugin id(s) to be used in place of the default for the task type."`
	MissingPluginBehavior int      `json:"missing_plugin_behavior" pflag:", Behavior when no specified plugin_id has an associated handler. 0 : FAIL , 1: DEFAULT."`
}

type WorkflowExecutionConfigConfig struct {
	Declarative bool                      `json:"declarative" pflag:", Whether the workflow execution configs are managed declaratively, i.e. whether non-listed configs are deleted."`
	Values      []WorkflowExecutionConfig `json:"values" pflag:", The list of workflow execution configs to be managed."`
}

type WorkflowExecutionConfig struct {
	Project  string `json:"project" pflag:", The project to which the workflow execution config is applied."`
	Domain   string `json:"domain" pflag:", The domain to which the workflow execution config is applied."`
	Workflow string `json:"workflow" pflag:", The workflow to which the workflow execution config is applied."`

	MaxParallelism  int32           `json:"max_parallelism" pflag:", Can be used to control the number of parallel nodes to run within the workflow. This is useful to achieve fairness."`
	SecurityContext SecurityContext `json:"security_context" pflag:", Security context permissions for executions triggered with this matchable attribute."`

	// Please note that only a subset of the configuration is exposed here currently.
}

type SecurityContext struct {
	RunAs Identity `json:"run_as" pflag:", The identity to run the execution as."`
}

type Identity struct {
	K8sServiceAccount string `json:"k8s_service_account" pflag:", The k8s service account to run the execution as."`
}

var defaultMatcheableResourcesConfig = &MatcheableResourcesConfig{}

var matcheableResourcesConfig = config.MustRegisterSection(SectionKey, defaultMatcheableResourcesConfig)

func GetConfig() *MatcheableResourcesConfig {
	return matcheableResourcesConfig.GetConfig().(*MatcheableResourcesConfig)
}
