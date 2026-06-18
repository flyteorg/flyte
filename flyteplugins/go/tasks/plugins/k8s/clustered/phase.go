package clustered

import (
	"context"
	"fmt"
	"sort"
	"time"

	v1 "k8s.io/api/core/v1"
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

const (
	// jobSetRestartingConditionType is emitted by newer JobSet controllers, but the pinned v0.5.2 API package
	// predates this constant. Keep this string check until we can bump the dependency.
	jobSetRestartingConditionType = "RestartingJobSet"
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
	maxRestarts := getMaxRestarts(jobSet, &spec)

	condition := extractCurrentCondition(jobSet.Status.Conditions)
	if condition == nil {
		if hasJobSetStarted(ctx, pluginContext, jobSet) {
			if phase, ok := maybeFastFailWorker0(ctx, pluginContext, jobSet, &taskInfo, maxRestarts, true); ok {
				return phase, nil
			}
			return runningPhaseInfo(ctx, pluginContext, &taskInfo, runningReason(jobSet.Status.Restarts)), nil
		}

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

	case jobsetv1alpha2.JobSetConditionType(jobSetRestartingConditionType):
		if phase, ok := maybeFastFailWorker0(ctx, pluginContext, jobSet, &taskInfo, maxRestarts, false); ok {
			return phase, nil
		}
		return runningPhaseInfo(
			ctx,
			pluginContext,
			&taskInfo,
			fmt.Sprintf("restart in progress (attempt %d)", jobSet.Status.Restarts),
		), nil
	}

	if phase, ok := maybeFastFailWorker0(ctx, pluginContext, jobSet, &taskInfo, maxRestarts, true); ok {
		return phase, nil
	}
	return runningPhaseInfo(ctx, pluginContext, &taskInfo, runningReason(jobSet.Status.Restarts)), nil
}

func runningReason(restarts int32) string {
	return fmt.Sprintf("running (restart attempt %d)", restarts)
}

func runningPhaseInfo(
	ctx context.Context,
	pluginContext k8s.PluginContext,
	taskInfo *pluginsCore.TaskInfo,
	reason string,
) pluginsCore.PhaseInfo {
	phaseInfo := pluginsCore.PhaseInfoRunning(pluginsCore.DefaultPhaseVersion, taskInfo)
	if reason != "" {
		phaseInfo.WithReason(reason)
	}
	if err := k8s.MaybeUpdatePhaseVersionFromPluginContext(&phaseInfo, &pluginContext); err != nil {
		logger.Warnf(ctx, "failed to update running phase version from plugin state: %v", err)
	}
	return phaseInfo
}

func getMaxRestarts(jobSet *jobsetv1alpha2.JobSet, spec *clusteredpb.ClusteredTaskSpec) int32 {
	if jobSet.Spec.FailurePolicy != nil {
		return jobSet.Spec.FailurePolicy.MaxRestarts
	}
	return spec.GetFailurePolicy().GetMaxRestarts()
}

func getWorkersStatus(jobSet *jobsetv1alpha2.JobSet) *jobsetv1alpha2.ReplicatedJobStatus {
	for i := range jobSet.Status.ReplicatedJobsStatus {
		if jobSet.Status.ReplicatedJobsStatus[i].Name == workersReplicatedJobName {
			return &jobSet.Status.ReplicatedJobsStatus[i]
		}
	}
	return nil
}

func workersHaveFailures(jobSet *jobsetv1alpha2.JobSet) bool {
	workersStatus := getWorkersStatus(jobSet)
	return workersStatus != nil && workersStatus.Failed > 0
}

func isRestartBudgetExhausted(jobSet *jobsetv1alpha2.JobSet, maxRestarts int32) bool {
	return jobSet.Status.Restarts >= maxRestarts
}

func readPluginState(ctx context.Context, pluginContext k8s.PluginContext) (k8s.PluginState, bool) {
	pluginState := k8s.PluginState{}
	if _, err := pluginContext.PluginStateReader().Get(&pluginState); err != nil {
		logger.Warnf(ctx, "failed to read plugin state: %v", err)
		return pluginState, false
	}
	return pluginState, true
}

func hasJobSetStarted(ctx context.Context, pluginContext k8s.PluginContext, jobSet *jobsetv1alpha2.JobSet) bool {
	if jobSet.Status.Restarts > 0 {
		return true
	}

	if workersStatus := getWorkersStatus(jobSet); workersStatus != nil {
		if workersStatus.Ready > 0 || workersStatus.Active > 0 || workersStatus.Succeeded > 0 {
			return true
		}
		// A non-zero Failed count means rank-0 pods have run, so the JobSet has started
		// regardless of remaining restart budget. Budget gating for surfacing the failure
		// is enforced downstream in maybeFastFailWorker0.
		if workersStatus.Failed > 0 {
			return true
		}
	}

	if pluginState, ok := readPluginState(ctx, pluginContext); ok && pluginState.Phase >= pluginsCore.PhaseRunning {
		return true
	}
	return false
}

// maybeFastFailWorker0 inspects the real rank-0 pod (suffix-tolerant lookup) for pending/failed diagnostics.
// Pending demystification is always evaluated; failed demystification is gated on exhausted restart budget.
func maybeFastFailWorker0(
	ctx context.Context,
	pluginContext k8s.PluginContext,
	jobSet *jobsetv1alpha2.JobSet,
	taskInfo *pluginsCore.TaskInfo,
	maxRestarts int32,
	allowFailedPath bool,
) (pluginsCore.PhaseInfo, bool) {
	pod := findRank0Pod(ctx, pluginContext, jobSet)
	if pod == nil {
		return pluginsCore.PhaseInfoUndefined, false
	}

	if pod.Status.Phase == v1.PodPending {
		phase, err := flytek8s.DemystifyPending(pod.Status, *taskInfo)
		if err != nil {
			logger.Warnf(ctx, "failed to inspect pending rank-0 pod for fast-fail: %v", err)
			return pluginsCore.PhaseInfoUndefined, false
		}
		if phase.Phase().IsFailure() {
			return phase, true
		}
		return pluginsCore.PhaseInfoUndefined, false
	}

	if pod.Status.Phase != v1.PodFailed || !allowFailedPath || !workersHaveFailures(jobSet) || !isRestartBudgetExhausted(jobSet, maxRestarts) {
		return pluginsCore.PhaseInfoUndefined, false
	}

	containerName := jobSet.Annotations[primaryContainerAnnotation]
	phase, err := flytek8s.DemystifyFailure(ctx, pod.Status, *taskInfo, containerName)
	if err != nil {
		logger.Warnf(ctx, "failed to inspect failed rank-0 pod for fast-fail: %v", err)
		return pluginsCore.PhaseInfoUndefined, false
	}
	if phase.Phase().IsFailure() {
		return phase, true
	}
	return pluginsCore.PhaseInfoUndefined, false
}

// maybeSystemRetryOnMaintenance inspects the rank-0 pod after a JobSetFailed condition.
// If the pod was evicted due to host maintenance (system-retryable), returns
// PhaseInfoSystemRetryableFailureWithCleanup so Flyte retries without charging user's max_restarts.
// Best-effort: if the pod is already cleaned up, returns (_, false) and the caller falls through.
func maybeSystemRetryOnMaintenance(ctx context.Context, pluginContext k8s.PluginContext, jobSet *jobsetv1alpha2.JobSet, taskInfo *pluginsCore.TaskInfo) (pluginsCore.PhaseInfo, bool) {
	pod := findRank0Pod(ctx, pluginContext, jobSet)
	if pod == nil {
		return pluginsCore.PhaseInfoUndefined, false
	}

	containerName := jobSet.Annotations[primaryContainerAnnotation]
	var (
		phase pluginsCore.PhaseInfo
		err   error
	)

	switch pod.Status.Phase {
	case v1.PodFailed:
		phase, err = flytek8s.DemystifyFailure(ctx, pod.Status, *taskInfo, containerName)
	case v1.PodPending:
		phase, err = flytek8s.DemystifyPending(pod.Status, *taskInfo)
	default:
		return pluginsCore.PhaseInfoUndefined, false
	}

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
