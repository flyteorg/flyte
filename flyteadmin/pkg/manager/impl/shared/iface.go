package shared

import (
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/golang/protobuf/ptypes/wrappers"
)

// WorkflowExecutionConfigInterface is used as common interface for capturing the common behavior catering to the needs
// of fetching the WorkflowExecutionConfig across LaunchPlanSpec, ExecutionCreateRequest
// MatchableResource_WORKFLOW_EXECUTION_CONFIG and ApplicationConfig
type WorkflowExecutionConfigInterface interface {
	// GetMaxParallelism Can be used to control the number of parallel nodes to run within the workflow. This is useful to achieve fairness.
	GetMaxParallelism() int32
	// GetRawOutputDataConfig Encapsulates user settings pertaining to offloaded data (i.e. Blobs, Schema, query data, etc.).
	GetRawOutputDataConfig() *admin.RawOutputDataConfig
	// GetSecurityContext Indicates security context permissions for executions triggered with this matchable attribute.
	GetSecurityContext() *core.SecurityContext
	// GetAnnotations Custom annotations to be applied to a triggered execution resource.
	GetAnnotations() *admin.Annotations
	// GetLabels Custom labels to be applied to a triggered execution resource.
	GetLabels() *admin.Labels
	// GetInterruptible indicates a workflow should be flagged as interruptible for a single execution. If omitted, the workflow's default is used.
	GetInterruptible() *wrappers.BoolValue
	// GetOverwriteCache indicates a workflow should skip all its cached results and re-compute its output, overwriting any already stored data.
	GetOverwriteCache() bool
	// GetEnvs defines environment variables to be set for the execution.
	GetEnvs() *admin.Envs
}
