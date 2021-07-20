package v1alpha1

import (
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
)

// This contains an OutputLocationPrefix. When running against AWS, this should be something of the form
// s3://my-bucket, or s3://my-bucket/  A sharding string will automatically be appended to this prefix before
// handing off to plugins/tasks. Sharding behavior may change in the future.
// Background available at https://github.com/flyteorg/flyte/issues/211
type RawOutputDataConfig struct {
	*admin.RawOutputDataConfig
}

func (in *RawOutputDataConfig) DeepCopyInto(out *RawOutputDataConfig) {
	*out = *in
}

// This contains workflow-execution specifications and overrides.
type ExecutionConfig struct {
	// Maps individual task types to their alternate (non-default) plugin handlers by name.
	TaskPluginImpls map[string]TaskPluginOverride
	// Can be used to control the number of parallel nodes to run within the workflow. This is useful to achieve fairness.
	MaxParallelism uint32
	// Defines execution behavior for processing nodes.
	RecoveryExecution WorkflowExecutionIdentifier
}

type TaskPluginOverride struct {
	PluginIDs             []string
	MissingPluginBehavior admin.PluginOverride_MissingPluginBehavior
}
