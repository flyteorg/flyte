package clustered

import (
	"context"
	"fmt"
	"sort"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	jobsetv1alpha2 "sigs.k8s.io/jobset/api/jobset/v1alpha2"

	pluginsCore "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/flytek8s"
	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/k8s"
	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/utils"
	"github.com/flyteorg/flyte/v2/flytestdlib/logger"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/core"
	clusteredpb "github.com/flyteorg/flyte/v2/gen/go/flyteidl2/plugins"
)

func (clusteredResourceHandler) GetTaskPhase(ctx context.Context, pluginContext k8s.PluginContext, resource client.Object) (pluginsCore.PhaseInfo, error) {
	jobSet, ok := resource.(*jobsetv1alpha2.JobSet)
	if !ok {
		return pluginsCore.PhaseInfoUndefined, fmt.Errorf("unexpected resource type %T", resource)
	}

	// Read spec for failure-policy flags (restart_on_host_maintenance).
	var spec clusteredpb.ClusteredTaskSpec
	if taskTemplate, err := pluginContext.TaskReader().Read(ctx); err == nil && taskTemplate != nil {
		if err := utils.UnmarshalStruct(taskTemplate.GetCustom(), &spec); err != nil { //nolint:staticcheck
			logger.Warningf(ctx, "failed to unmarshal ClusteredTaskSpec: %v", err)
		}
	}

	taskLogs, err := getTaskLogs(ctx, pluginContext, jobSet)
	if err != nil {
		return pluginsCore.PhaseInfoUndefined, err
	}

	occurredAt := time.Now()
	statusDetails, err := utils.MarshalObjToStruct(jobSet.Status) //nolint:staticcheck
	if err != nil {
		logger.Warnf(ctx, "failed to marshal JobSet status for task info: %v", err)
	}
	taskInfo := pluginsCore.TaskInfo{
		Logs:       taskLogs,
		LogContext: getLogContext(ctx, pluginContext, jobSet),
		OccurredAt: &occurredAt,
		CustomInfo: statusDetails,
	}

	condition := extractCurrentCondition(jobSet.Status.Conditions)
	if condition == nil {
		// JobSet exists, no terminal condition yet.
		return pluginsCore.PhaseInfoInitializing(occurredAt, pluginsCore.DefaultPhaseVersion, "pods scheduling / DNS resolving", &taskInfo), nil
	}

	switch jobsetv1alpha2.JobSetConditionType(condition.Type) {
	case jobsetv1alpha2.JobSetCompleted:
		return pluginsCore.PhaseInfoSuccess(&taskInfo), nil

	case jobsetv1alpha2.JobSetFailed:
		if spec.GetFailurePolicy().GetRestartOnHostMaintenance() {
			if phase, ok := maybeSystemRetryOnMaintenance(ctx, pluginContext, jobSet, &taskInfo); ok {
				return phase, nil
			}
		}
		return pluginsCore.PhaseInfoRetryableFailure(condition.Reason, condition.Message, &taskInfo), nil
	}

	// Running: check for fast-fail before reporting PhaseRunning.
	if phase, ok := maybeFastFailWorker0(ctx, pluginContext, jobSet, &taskInfo); ok {
		return phase, nil
	}

	phaseInfo := pluginsCore.PhaseInfoRunning(pluginsCore.DefaultPhaseVersion, &taskInfo)

	if err := k8s.MaybeUpdatePhaseVersionFromPluginContext(&phaseInfo, &pluginContext); err != nil {
		return phaseInfo, err
	}
	return phaseInfo, nil
}

// maybeFastFailWorker0 inspects the rank-0 pod when at least one Job under the "workers"
// ReplicatedJob has failed. Returns (phaseInfo, true) if the pod is in a terminal failed state.
// This surfaces the failure before the JobSet controller sets JobSetFailed, reducing tail latency.
func maybeFastFailWorker0(ctx context.Context, pluginContext k8s.PluginContext, jobSet *jobsetv1alpha2.JobSet, taskInfo *pluginsCore.TaskInfo) (pluginsCore.PhaseInfo, bool) {
	for _, s := range jobSet.Status.ReplicatedJobsStatus {
		if s.Name == workersReplicatedJobName && s.Failed > 0 {
			podName := rank0PodName(jobSet.Name)
			containerName := jobSet.Annotations[primaryContainerAnnotation]
			phase, err := flytek8s.DemystifyFailedOrPendingPod(ctx, pluginContext, *taskInfo, jobSet.Namespace, podName, containerName)
			if err != nil {
				logger.Warnf(ctx, "failed to inspect rank-0 pod for fast-fail: %v", err)
				return pluginsCore.PhaseInfoUndefined, false
			}
			if phase.Phase().IsFailure() {
				return phase, true
			}
		}
	}
	return pluginsCore.PhaseInfoUndefined, false
}

// maybeSystemRetryOnMaintenance inspects the rank-0 pod after a JobSetFailed condition.
// If the pod was evicted due to host maintenance (system-retryable), returns
// PhaseInfoSystemRetryableFailureWithCleanup so Flyte retries without charging user's max_restarts.
// Best-effort: if the pod is already cleaned up, returns (_, false) and the caller falls through.
func maybeSystemRetryOnMaintenance(ctx context.Context, pluginContext k8s.PluginContext, jobSet *jobsetv1alpha2.JobSet, taskInfo *pluginsCore.TaskInfo) (pluginsCore.PhaseInfo, bool) {
	podName := rank0PodName(jobSet.Name)
	containerName := jobSet.Annotations[primaryContainerAnnotation]
	phase, err := flytek8s.DemystifyFailedOrPendingPod(ctx, pluginContext, *taskInfo, jobSet.Namespace, podName, containerName)
	if err != nil {
		logger.Warnf(ctx, "failed to inspect rank-0 pod for maintenance retry: %v", err)
		return pluginsCore.PhaseInfoUndefined, false
	}
	if phase.Phase() == pluginsCore.PhaseRetryableFailure && phase.Err() != nil && phase.Err().GetKind() == core.ExecutionError_SYSTEM {
		return pluginsCore.PhaseInfoSystemRetryableFailureWithCleanup(
			"HostMaintenance", "pod evicted due to host maintenance; retrying without charging max_restarts", taskInfo,
		), true
	}
	return pluginsCore.PhaseInfoUndefined, false
}

// extractCurrentCondition returns the most recently transitioned condition with Status=True, or nil.
// Ported from kfoperators/common/common_operator.go — not imported to avoid the dependency.
func extractCurrentCondition(conditions []metav1.Condition) *metav1.Condition {
	if len(conditions) == 0 {
		return nil
	}
	sorted := make([]metav1.Condition, len(conditions))
	copy(sorted, conditions)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].LastTransitionTime.After(sorted[j].LastTransitionTime.Time)
	})
	for i := range sorted {
		if sorted[i].Status == metav1.ConditionTrue {
			return &sorted[i]
		}
	}
	return nil
}
