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
	"github.com/flyteorg/flyte/v2/flytestdlib/logger"
	"github.com/flyteorg/flyte/v2/flytestdlib/utils"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/core"
	clusteredpb "github.com/flyteorg/flyte/v2/gen/go/flyteidl2/plugins"
)

// GetTaskPhase maps the JobSet to a phase as a one-way decision tree: read every fact
// once, then ask whether the JobSet is finished, then whether the whole gang has ever
// been up. Under each side, suspended has one meaning: before the gang started an
// admission gate is holding it, after it started the gate took it back.
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

	// Facts, each read once.
	pluginState := readPluginState(ctx, pluginContext)
	condition := extractCurrentCondition(jobSet.Status.Conditions)
	suspended := isSuspended(jobSet, condition)
	started := hasJobSetStarted(jobSet, pluginState)
	pods := currentRestartPods(jobSet, listWorkerPods(ctx, pluginContext, jobSet))
	maxRestarts := getMaxRestarts(jobSet, &spec)

	// 1. Is it over? A suspended JobSet is never terminal.
	switch conditionType(condition) {
	case jobsetv1alpha2.JobSetCompleted:
		return pluginsCore.PhaseInfoSuccess(&taskInfo), nil

	case jobsetv1alpha2.JobSetFailed:
		if spec.GetFailurePolicy().GetRestartOnHostMaintenance() {
			if phase, ok := maybeSystemRetryOnMaintenance(ctx, jobSet, pods, &taskInfo); ok {
				return phase, nil
			}
		}
		return pluginsCore.PhaseInfoRetryableFailure(condition.Reason, condition.Message, &taskInfo), nil
	}

	// 2. Has the whole gang ever been up?
	var phaseInfo pluginsCore.PhaseInfo
	if started {
		phaseInfo = afterStart(ctx, pluginContext, jobSet, condition, suspended, pods, maxRestarts, &taskInfo)
	} else {
		phaseInfo = beforeStart(ctx, pluginContext, jobSet, condition, suspended, pods, maxRestarts, pluginState, &taskInfo)
	}

	// The tree is built so that a non-terminal phase never moves backwards: downstream drops
	// a lower phase and its reason. If one ever does, say so instead of losing it silently.
	if !phaseInfo.Phase().IsTerminal() && !pluginState.Phase.IsTerminal() && phaseInfo.Phase() < pluginState.Phase {
		logger.Warnf(ctx, "JobSet %s/%s: computed phase %s is behind the reported phase %s (%s); it will be dropped",
			jobSet.Namespace, jobSet.Name, phaseInfo.Phase(), pluginState.Phase, phaseInfo.Reason())
	}

	// A new reason within the same phase needs a new version, or its event is dropped
	// as a duplicate of the previous one.
	k8s.MaybeUpdatePhaseVersion(&phaseInfo, &pluginState)
	return phaseInfo, nil
}

// beforeStart reports a gang that has never been fully up in this attempt.
//
// Phases never move backwards within an attempt: the runs service drops a lower phase
// and deduplicates its event, so a gang that falls back after reaching Initializing
// (the gate released a partial start, or a re-admitted gang is scheduling again)
// stays at Initializing and carries the new reason instead.
func beforeStart(
	ctx context.Context,
	pluginContext k8s.PluginContext,
	jobSet *jobsetv1alpha2.JobSet,
	condition *metav1.Condition,
	suspended bool,
	pods []v1.Pod,
	maxRestarts int32,
	pluginState k8s.PluginState,
	taskInfo *pluginsCore.TaskInfo,
) pluginsCore.PhaseInfo {
	reachedInitializing := pluginState.Phase == pluginsCore.PhaseInitializing
	at := *taskInfo.OccurredAt
	waiting := func(reason string) pluginsCore.PhaseInfo {
		if reachedInitializing {
			return pluginsCore.PhaseInfoInitializing(at, pluginsCore.DefaultPhaseVersion, reason, taskInfo)
		}
		return pluginsCore.PhaseInfoWaitingForResourcesInfo(at, pluginsCore.DefaultPhaseVersion, reason, taskInfo)
	}

	// Not admitted yet: the gate holds the JobSet and no pods exist.
	if suspended {
		if phase, ok := queueNotFound(ctx, pluginContext, jobSet, taskInfo); ok {
			return phase
		}
		reason := "waiting for gang admission"
		if reachedInitializing {
			reason = "released by admission gate before all workers were ready; waiting for re-admission"
		}
		return waiting(reason + conditionDetail(condition))
	}

	if phase, ok := fatalFailure(ctx, jobSet, pods, maxRestarts, taskInfo); ok {
		return phase
	}

	expected := expectedWorkers(jobSet)
	scheduled, ready, unscheduledDetail := countWorkers(pods)
	if scheduled < expected {
		return waiting(fmt.Sprintf("%d of %d workers scheduled%s", scheduled, expected, unscheduledDetail))
	}
	return pluginsCore.PhaseInfoInitializing(at, pluginsCore.DefaultPhaseVersion,
		fmt.Sprintf("%d of %d workers ready", ready, expected), taskInfo)
}

// afterStart reports a gang that has been fully up at least once in this attempt.
func afterStart(
	ctx context.Context,
	pluginContext k8s.PluginContext,
	jobSet *jobsetv1alpha2.JobSet,
	condition *metav1.Condition,
	suspended bool,
	pods []v1.Pod,
	maxRestarts int32,
	taskInfo *pluginsCore.TaskInfo,
) pluginsCore.PhaseInfo {
	// The gate took the gang back: the JobSet is re-suspended and its pods are gone.
	if suspended {
		return evictedPhaseInfo(ctx, pluginContext, jobSet, taskInfo)
	}

	if conditionType(condition) == jobsetv1alpha2.JobSetRestarting {
		if phase, ok := pendingFailure(ctx, pods, taskInfo); ok {
			return phase
		}
		return runningPhaseInfo(taskInfo, fmt.Sprintf("restart in progress (attempt %d)", jobSet.Status.Restarts))
	}

	if phase, ok := fatalFailure(ctx, jobSet, pods, maxRestarts, taskInfo); ok {
		return phase
	}
	return runningPhaseInfo(taskInfo, runningReason(jobSet.Status.Restarts))
}

func runningReason(restarts int32) string {
	return fmt.Sprintf("running (restart attempt %d)", restarts)
}

func runningPhaseInfo(taskInfo *pluginsCore.TaskInfo, reason string) pluginsCore.PhaseInfo {
	phaseInfo := pluginsCore.PhaseInfoRunning(pluginsCore.DefaultPhaseVersion, taskInfo)
	phaseInfo.WithReason(reason)
	return phaseInfo
}

// evictionSource names the admission gate that suspends JobSets: Kueue's JobSet integration,
// which suspends a JobSet labelled with a queue when it is created and again when it takes
// the gang back. The plugin never suspends a JobSet itself.
const evictionSource = "kueue"

// evictedPhaseInfo reports a gang the gate took back after it had fully started. The attempt
// is over and the JobSet is cleaned up. Kueue's Workload says why: a cause the user owns
// (preemption in the queue they chose, the maximum execution time they set) is a user retry
// that uses one of the task's retries; any other cause, or a Workload that cannot be read, is
// a system retry that does not. The message says which, and how many retries are left.
func evictedPhaseInfo(
	ctx context.Context,
	pluginContext k8s.PluginContext,
	jobSet *jobsetv1alpha2.JobSet,
	taskInfo *pluginsCore.TaskInfo,
) pluginsCore.PhaseInfo {
	wl, err := workloadForJobSet(ctx, pluginContext.K8sReader(), jobSet)
	if err != nil {
		logger.Warnf(ctx, "failed to read the Kueue Workload of JobSet %s/%s; reporting its eviction as a system retry: %v",
			jobSet.Namespace, jobSet.Name, err)
	}
	eviction := classifyEviction(wl)
	if eviction.UserCaused {
		eviction.RetriesLeft = retriesRemaining(pluginContext)
	}
	execErr := eviction.ExecutionError()
	if eviction.UserCaused {
		return pluginsCore.PhaseInfoRetryableFailureWithCleanup(execErr.GetCode(), execErr.GetMessage(), taskInfo)
	}
	return pluginsCore.PhaseInfoSystemRetryableFailureWithCleanup(execErr.GetCode(), execErr.GetMessage(), taskInfo)
}

// retriesRemaining is how many of the task's retries are still available for the attempt that
// just failed to use: the attempts the task may make, less the ones it has made, this one
// included.
func retriesRemaining(pluginContext k8s.PluginContext) uint32 {
	meta := pluginContext.TaskExecutionMetadata()
	if meta == nil {
		return 0
	}
	used := meta.GetTaskExecutionID().GetID().GetRetryAttempt() + 1
	if maxAttempts := meta.GetMaxAttempts(); maxAttempts > used {
		return maxAttempts - used
	}
	return 0
}

// codeQueueNotFound is the user error for a task whose Kueue queue label names a LocalQueue
// that does not exist.
const codeQueueNotFound = "KueueLocalQueueNotFound"

// queueNotFound fails a held JobSet whose queue label names a LocalQueue that does not exist:
// Kueue can never admit it, so waiting would only run out the clock. It is the user's error
// and is not retried. The JobSet is cleaned up.
func queueNotFound(
	ctx context.Context,
	pluginContext k8s.PluginContext,
	jobSet *jobsetv1alpha2.JobSet,
	taskInfo *pluginsCore.TaskInfo,
) (pluginsCore.PhaseInfo, bool) {
	if jobSet.Labels[kueueQueueNameLabel] == "" {
		return pluginsCore.PhaseInfoUndefined, false
	}
	wl, err := workloadForJobSet(ctx, pluginContext.K8sReader(), jobSet)
	if err != nil {
		logger.Warnf(ctx, "failed to read the Kueue Workload of JobSet %s/%s: %v", jobSet.Namespace, jobSet.Name, err)
		return pluginsCore.PhaseInfoUndefined, false
	}
	queue, missing := missingLocalQueue(jobSet, wl)
	if !missing {
		return pluginsCore.PhaseInfoUndefined, false
	}
	execErr := &core.ExecutionError{
		Kind: core.ExecutionError_USER,
		Code: codeQueueNotFound,
		Message: fmt.Sprintf("the Kueue queue %q set by the %s label does not exist in namespace %q; "+
			"use one of the queues your platform provides", queue, kueueQueueNameLabel, jobSet.Namespace),
	}
	return pluginsCore.PhaseInfoFailed(pluginsCore.PhasePermanentFailure, execErr, taskInfo).WithCleanupOnFailure(), true
}

// isSuspended reports whether an admission gate is holding the JobSet. spec.suspend
// is checked as well as the condition because the first poll after creation can run
// before the JobSet controller has written Suspended=True.
func isSuspended(jobSet *jobsetv1alpha2.JobSet, condition *metav1.Condition) bool {
	if jobSet.Spec.Suspend != nil && *jobSet.Spec.Suspend {
		return true
	}
	return conditionType(condition) == jobsetv1alpha2.JobSetSuspended
}

func conditionType(condition *metav1.Condition) jobsetv1alpha2.JobSetConditionType {
	if condition == nil {
		return ""
	}
	return jobsetv1alpha2.JobSetConditionType(condition.Type)
}

func conditionDetail(condition *metav1.Condition) string {
	if condition == nil || condition.Message == "" {
		return ""
	}
	return ": " + condition.Message
}

// expectedWorkers is the number of worker pods the gang needs: the workers
// ReplicatedJob's replicas times each Job's parallelism.
func expectedWorkers(jobSet *jobsetv1alpha2.JobSet) int {
	for _, rjob := range jobSet.Spec.ReplicatedJobs {
		if rjob.Name != workersReplicatedJobName {
			continue
		}
		parallelism := int32(1)
		if rjob.Template.Spec.Parallelism != nil {
			parallelism = *rjob.Template.Spec.Parallelism
		}
		return int(rjob.Replicas * parallelism)
	}
	return 0
}

// countWorkers counts the live worker pods that have been given a node and that are
// Ready. For the first pod without a node it also returns the scheduler's reason,
// formatted as a suffix for the phase reason.
func countWorkers(pods []v1.Pod) (scheduled, ready int, unscheduledDetail string) {
	for i := range pods {
		pod := &pods[i]
		if pod.DeletionTimestamp != nil || !isActivePodPhase(pod.Status.Phase) {
			continue
		}
		if pod.Spec.NodeName == "" {
			if unscheduledDetail == "" {
				unscheduledDetail = schedulingDetail(pod)
			}
			continue
		}
		scheduled++
		if isPodReady(pod) {
			ready++
		}
	}
	return scheduled, ready, unscheduledDetail
}

// maxSchedulingDetailLen bounds the scheduler's message in the phase reason. The
// scheduler appends preemption and claim details that can run to several hundred
// characters; the first part is what tells the user why the pod has no node.
const maxSchedulingDetailLen = 200

func schedulingDetail(pod *v1.Pod) string {
	for _, c := range pod.Status.Conditions {
		if c.Type == v1.PodScheduled && c.Status == v1.ConditionFalse && c.Message != "" {
			message := c.Message
			if len(message) > maxSchedulingDetailLen {
				message = message[:maxSchedulingDetailLen] + "…"
			}
			return fmt.Sprintf(" (%s: %s)", c.Reason, message)
		}
	}
	return ""
}

func isPodReady(pod *v1.Pod) bool {
	for _, c := range pod.Status.Conditions {
		if c.Type == v1.PodReady {
			return c.Status == v1.ConditionTrue
		}
	}
	return false
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

// readPluginState returns the state stored by the last poll, or the zero state (phase
// Undefined) if it cannot be read.
func readPluginState(ctx context.Context, pluginContext k8s.PluginContext) k8s.PluginState {
	pluginState := k8s.PluginState{}
	if _, err := pluginContext.PluginStateReader().Get(&pluginState); err != nil {
		logger.Warnf(ctx, "failed to read plugin state: %v", err)
		return k8s.PluginState{}
	}
	return pluginState
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
// failure, by which point the gang had started (the JobSet controller increments
// Restarts and sets Restarting=True in the same status update). A Failed count alone
// does not mean started; the failure still surfaces on the not-started path.
func hasJobSetStarted(jobSet *jobsetv1alpha2.JobSet, pluginState k8s.PluginState) bool {
	if jobSet.Status.Restarts > 0 {
		return true
	}
	if workersStatus := getWorkersStatus(jobSet); workersStatus != nil {
		if workersStatus.Ready > 0 || workersStatus.Succeeded > 0 {
			return true
		}
	}
	return pluginState.Phase >= pluginsCore.PhaseRunning
}

// fatalFailure surfaces problems the gang will not recover from by waiting: a fatal
// pending state on any worker, or a failed worker once the JobSet has no restarts left.
func fatalFailure(
	ctx context.Context,
	jobSet *jobsetv1alpha2.JobSet,
	pods []v1.Pod,
	maxRestarts int32,
	taskInfo *pluginsCore.TaskInfo,
) (pluginsCore.PhaseInfo, bool) {
	if phase, ok := pendingFailure(ctx, pods, taskInfo); ok {
		return phase, true
	}
	return failedWithBudgetExhausted(ctx, jobSet, pods, maxRestarts, taskInfo)
}

// pendingFailure reports the first pending worker whose state is fatal (for example an
// image that cannot be pulled), as classified by DemystifyPending.
func pendingFailure(ctx context.Context, pods []v1.Pod, taskInfo *pluginsCore.TaskInfo) (pluginsCore.PhaseInfo, bool) {
	for i := range pods {
		pod := &pods[i]
		if pod.Status.Phase != v1.PodPending {
			continue
		}
		phase, err := flytek8s.DemystifyPending(pod.Status, *taskInfo)
		if err != nil {
			logger.Warnf(ctx, "failed to inspect pending pod %s for fast-fail: %v", pod.Name, err)
			continue
		}
		if phase.Phase().IsFailure() {
			return phase, true
		}
	}
	return pluginsCore.PhaseInfoUndefined, false
}

// failedWithBudgetExhausted reports the first worker to fail once the JobSet has no
// restarts left, so the failure is surfaced before the JobSet controller writes Failed.
// The earliest failure is the root cause; the ranks that failed after it were torn down
// because of it.
func failedWithBudgetExhausted(
	ctx context.Context,
	jobSet *jobsetv1alpha2.JobSet,
	pods []v1.Pod,
	maxRestarts int32,
	taskInfo *pluginsCore.TaskInfo,
) (pluginsCore.PhaseInfo, bool) {
	if !workersHaveFailures(jobSet) || !isRestartBudgetExhausted(jobSet, maxRestarts) {
		return pluginsCore.PhaseInfoUndefined, false
	}
	pod := firstFailedPod(pods)
	if pod == nil {
		return pluginsCore.PhaseInfoUndefined, false
	}

	containerName := jobSet.Annotations[primaryContainerAnnotation]
	phase, err := flytek8s.DemystifyFailure(ctx, pod.Status, *taskInfo, containerName)
	if err != nil {
		logger.Warnf(ctx, "failed to inspect failed pod %s for fast-fail: %v", pod.Name, err)
		return pluginsCore.PhaseInfoUndefined, false
	}
	if phase.Phase().IsFailure() {
		return phase, true
	}
	return pluginsCore.PhaseInfoUndefined, false
}

// maybeSystemRetryOnMaintenance inspects the workers after a JobSetFailed condition. If
// the gang went down because a worker was evicted for host maintenance (system-retryable),
// returns PhaseInfoSystemRetryableFailureWithCleanup so Flyte retries without charging
// user's max_restarts.
//
// Only the earliest failure is classified: it is the cause, and the workers that failed
// after it were torn down because of it. A torn-down worker is SIGKILLed, which on its
// own would be classified as a system failure and turn a user's crash into a free retry.
// Best-effort: if the pods are already cleaned up, returns (_, false) and the caller falls through.
func maybeSystemRetryOnMaintenance(
	ctx context.Context,
	jobSet *jobsetv1alpha2.JobSet,
	pods []v1.Pod,
	taskInfo *pluginsCore.TaskInfo,
) (pluginsCore.PhaseInfo, bool) {
	containerName := jobSet.Annotations[primaryContainerAnnotation]
	if pod := firstFailedPod(pods); pod != nil {
		phase, err := flytek8s.DemystifyFailure(ctx, pod.Status, *taskInfo, containerName)
		if err != nil {
			logger.Warnf(ctx, "failed to inspect failed pod %s for maintenance retry: %v", pod.Name, err)
			return pluginsCore.PhaseInfoUndefined, false
		}
		return maintenanceRetry(phase, pod, taskInfo)
	}

	// No worker failed: a pending worker can still carry a system-level reason.
	for i := range pods {
		pod := &pods[i]
		if pod.Status.Phase != v1.PodPending {
			continue
		}
		phase, err := flytek8s.DemystifyPending(pod.Status, *taskInfo)
		if err != nil {
			logger.Warnf(ctx, "failed to inspect pending pod %s for maintenance retry: %v", pod.Name, err)
			continue
		}
		if retry, ok := maintenanceRetry(phase, pod, taskInfo); ok {
			return retry, true
		}
	}
	return pluginsCore.PhaseInfoUndefined, false
}

// maintenanceRetry turns a system-retryable pod failure into the HostMaintenance retry.
func maintenanceRetry(
	phase pluginsCore.PhaseInfo,
	pod *v1.Pod,
	taskInfo *pluginsCore.TaskInfo,
) (pluginsCore.PhaseInfo, bool) {
	if phase.Phase() != pluginsCore.PhaseRetryableFailure || phase.Err() == nil ||
		phase.Err().GetKind() != core.ExecutionError_SYSTEM {
		return pluginsCore.PhaseInfoUndefined, false
	}
	return pluginsCore.PhaseInfoSystemRetryableFailureWithCleanup(
		"HostMaintenance",
		fmt.Sprintf("pod %s evicted due to host maintenance; retrying without charging max_restarts", pod.Name),
		taskInfo,
	), true
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
