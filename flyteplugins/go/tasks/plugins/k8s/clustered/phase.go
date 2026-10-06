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
	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/gang"
	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/k8s"
	"github.com/flyteorg/flyte/v2/flytestdlib/logger"
	"github.com/flyteorg/flyte/v2/flytestdlib/utils"
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
		if err := utils.UnmarshalStructToPb(taskTemplate.GetCustom(), &spec); err != nil {
			logger.Warningf(ctx, "failed to unmarshal ClusteredTaskSpec: %v", err)
		}
	}

	taskLogs, err := getTaskLogs(ctx, pluginContext, jobSet)
	if err != nil {
		return pluginsCore.PhaseInfoUndefined, err
	}

	occurredAt := time.Now()
	statusDetails, err := utils.MarshalObjToStruct(jobSet.Status)
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
	if condition != nil {
		// Terminal conditions win over everything else; a suspended JobSet is never terminal.
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
	}

	started := hasJobSetStarted(ctx, pluginContext, jobSet)

	// An admission gate (Kueue's JobSet integration) holds the JobSet through
	// spec.suspend: no pods exist while it is held. This must never look like
	// Running, because the executors anchor max_runtime on the first Running and
	// stop queued_timeout there.
	if isSuspended(jobSet, condition) {
		if started {
			return evictedPhaseInfo(ctx, jobSet, &taskInfo, condition), nil
		}
		return holdPhase(ctx, pluginContext, &taskInfo, condition), nil
	}

	if condition != nil && jobsetv1alpha2.JobSetConditionType(condition.Type) == jobsetv1alpha2.JobSetRestarting {
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

	if !started {
		// Pending-pod diagnostics (image pull, unschedulable) and an exhausted restart
		// budget still surface while the gang is forming.
		if phase, ok := maybeFastFailWorker0(ctx, pluginContext, jobSet, &taskInfo, maxRestarts, true); ok {
			return phase, nil
		}
		return pluginsCore.PhaseInfoInitializing(
			occurredAt, pluginsCore.DefaultPhaseVersion, initializingReason(&spec), &taskInfo,
		), nil
	}

	// Started, and no condition we act on (or an unrecognised one): the gang is running.
	if phase, ok := maybeFastFailWorker0(ctx, pluginContext, jobSet, &taskInfo, maxRestarts, true); ok {
		return phase, nil
	}
	return runningPhaseInfo(ctx, pluginContext, &taskInfo, runningReason(jobSet.Status.Restarts)), nil
}

func runningReason(restarts int32) string {
	return fmt.Sprintf("running (restart attempt %d)", restarts)
}

func initializingReason(spec *clusteredpb.ClusteredTaskSpec) string {
	if replicas := spec.GetReplicas(); replicas > 0 {
		return fmt.Sprintf("waiting for all %d workers to be ready (pods scheduling / DNS resolving)", replicas)
	}
	return "waiting for all workers to be ready (pods scheduling / DNS resolving)"
}

// evictionSource names the admission gate that flips spec.suspend on our JobSets.
// Only Kueue's JobSet integration does that today; the plugin itself sets suspend
// only at creation, and only when Kueue is enabled.
const evictionSource = "kueue"

// evictionPolicy decides how a post-start eviction is reported. Plugin config
// (plugins.clustered.kueue.evict-as-system-retry) takes this over once it exists.
var evictionPolicy = gang.Policy{AsSystemRetry: true}

// isSuspended reports whether an admission gate is holding the JobSet. spec.suspend
// is checked as well as the condition because the first poll after creation can run
// before the JobSet controller has written Suspended=True.
func isSuspended(jobSet *jobsetv1alpha2.JobSet, condition *metav1.Condition) bool {
	if jobSet.Spec.Suspend != nil && *jobSet.Spec.Suspend {
		return true
	}
	return condition != nil && jobsetv1alpha2.JobSetConditionType(condition.Type) == jobsetv1alpha2.JobSetSuspended
}

// holdPhase reports a gang the gate is holding before it ever fully started.
//
// The phase never moves backwards. Once the plugin has reported Initializing for a
// partial start that the gate then released, it keeps reporting Initializing while
// the gate requeues the gang in place: phases are monotonic within an attempt for
// everything downstream of the plugin, so a lower phase would only be dropped.
func holdPhase(
	ctx context.Context,
	pluginContext k8s.PluginContext,
	taskInfo *pluginsCore.TaskInfo,
	condition *metav1.Condition,
) pluginsCore.PhaseInfo {
	occurredAt := time.Now()
	if taskInfo.OccurredAt != nil {
		occurredAt = *taskInfo.OccurredAt
	}
	detail := ""
	if condition != nil && condition.Message != "" {
		detail = ": " + condition.Message
	}

	var phaseInfo pluginsCore.PhaseInfo
	pluginState, ok := readPluginState(ctx, pluginContext)
	if ok && pluginState.Phase >= pluginsCore.PhaseInitializing && pluginState.Phase < pluginsCore.PhaseSuccess {
		phaseInfo = pluginsCore.PhaseInfoInitializing(occurredAt, pluginsCore.DefaultPhaseVersion,
			"released by admission gate before all workers were ready; waiting for re-admission"+detail, taskInfo)
	} else {
		phaseInfo = pluginsCore.PhaseInfoWaitingForResourcesInfo(occurredAt, pluginsCore.DefaultPhaseVersion,
			"waiting for gang admission"+detail, taskInfo)
	}
	if err := k8s.MaybeUpdatePhaseVersionFromPluginContext(&phaseInfo, &pluginContext); err != nil {
		logger.Warnf(ctx, "failed to update hold phase version from plugin state: %v", err)
	}
	return phaseInfo
}

// evictedPhaseInfo reports a gang the gate revoked after it had fully started. The
// JobSet is re-suspended and its pods are gone, so the attempt is over: the gang
// package decides whether that is charged as a system or a user retry.
func evictedPhaseInfo(
	ctx context.Context,
	jobSet *jobsetv1alpha2.JobSet,
	taskInfo *pluginsCore.TaskInfo,
	condition *metav1.Condition,
) pluginsCore.PhaseInfo {
	occurredAt := time.Now()
	if taskInfo.OccurredAt != nil {
		occurredAt = *taskInfo.OccurredAt
	}
	message := "JobSet suspended after the gang had started; admission was revoked"
	if condition != nil && condition.Message != "" {
		message += " (" + condition.Message + ")"
	}

	action, execErr := gang.Decide(gang.Eviction{
		Source:     evictionSource,
		Reason:     gang.ReasonUnknown,
		Message:    message,
		Started:    true,
		OccurredAt: occurredAt,
	}, evictionPolicy)
	switch action {
	case gang.UserRetry:
		return pluginsCore.PhaseInfoRetryableFailureWithCleanup(execErr.GetCode(), execErr.GetMessage(), taskInfo)
	case gang.SystemRetry:
		return pluginsCore.PhaseInfoSystemRetryableFailureWithCleanup(execErr.GetCode(), execErr.GetMessage(), taskInfo)
	default:
		// Decide never holds a started gang; never report Running for a JobSet with no pods.
		logger.Warnf(ctx, "unexpected gang action %s for started JobSet %s/%s", action, jobSet.Namespace, jobSet.Name)
		return pluginsCore.PhaseInfoSystemRetryableFailureWithCleanup(gang.CodeGangEvicted, message, taskInfo)
	}
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
	// Status.Restarts counts every whole-set restart, including free host-maintenance
	// restarts (RestartJobSetAndIgnoreMaxRestarts). Only RestartsCountTowardsMax is
	// charged against maxRestarts, so compare that — otherwise free restarts would
	// falsely trip the fast-fail path while the JobSet controller is still restarting.
	return jobSet.Status.RestartsCountTowardsMax >= maxRestarts
}

func readPluginState(ctx context.Context, pluginContext k8s.PluginContext) (k8s.PluginState, bool) {
	pluginState := k8s.PluginState{}
	if _, err := pluginContext.PluginStateReader().Get(&pluginState); err != nil {
		logger.Warnf(ctx, "failed to read plugin state: %v", err)
		return pluginState, false
	}
	return pluginState, true
}

// hasJobSetStarted reports whether the whole gang has been up at least once in this
// attempt.
//
// JobSet's ReplicatedJobStatus counts child Jobs, not pods: with one Job of N pods,
// Active becomes 1 as soon as any pod exists and Ready becomes 1 only when all N are
// up. Ready (or Succeeded) is therefore the signal, never Active. The stored plugin
// phase is sticky: once Running has been reported the gang counts as started for
// the rest of the attempt, which is what turns a later suspension into an eviction
// instead of a hold; executors clear plugin state when they relaunch. Restarts > 0
// keeps the existing restart semantics: a whole-set restart needs a prior child
// failure, by which point the gang had started. A Failed count alone does not mean
// started; the failure still surfaces through maybeFastFailWorker0 on the
// not-started path.
func hasJobSetStarted(ctx context.Context, pluginContext k8s.PluginContext, jobSet *jobsetv1alpha2.JobSet) bool {
	if jobSet.Status.Restarts > 0 {
		return true
	}

	if workersStatus := getWorkersStatus(jobSet); workersStatus != nil {
		if workersStatus.Ready > 0 || workersStatus.Succeeded > 0 {
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
