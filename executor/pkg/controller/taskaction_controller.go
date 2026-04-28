/*
Copyright 2025.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"time"

	"k8s.io/client-go/util/retry"

	"connectrpc.com/connect"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	flyteorgv1 "github.com/flyteorg/flyte/v2/executor/api/v1"
	"github.com/flyteorg/flyte/v2/executor/pkg/plugin"
	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/catalog"
	pluginsCore "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/flyteorg/flyte/v2/flytestdlib/storage"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/common"
	core "github.com/flyteorg/flyte/v2/gen/go/flyteidl2/core"
	task "github.com/flyteorg/flyte/v2/gen/go/flyteidl2/task"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/workflow"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/workflow/workflowconnect"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	TaskActionDefaultRequeueDuration = 10 * time.Second
	taskActionFinalizer              = "flyte.org/plugin-finalizer"

	// DefaultMaxSystemFailures bounds consecutive system errors before the
	// TaskAction is forced to PermanentFailure.
	DefaultMaxSystemFailures uint32 = 3

	// MaxSystemFailuresExceededCode is the ExecutionError.Code stamped on
	// TaskActions that hit the system-failure ceiling.
	MaxSystemFailuresExceededCode = "MaxSystemFailuresExceeded"

	// LabelTerminationStatus marks a TaskAction as terminated for GC discovery.
	LabelTerminationStatus = "flyte.org/termination-status"
	// LabelCompletedTime records the UTC time (minute precision) when the TaskAction became terminal.
	LabelCompletedTime = "flyte.org/completed-time"
	// LabelValueTerminated is the value for LabelTerminationStatus.
	LabelValueTerminated = "terminated"
	// labelTimeFormat is the time format used for the completed-time label (lexicographically ordered, minute precision).
	labelTimeFormat = "2006-01-02.15-04"
)

type K8sEventType string

const (
	FailedUnmarshal     K8sEventType = "FailedUnmarshal"
	FailedValidation    K8sEventType = "FailedValidation"
	FailedPluginResolve K8sEventType = "FailedPluginResolve"
	FailedPluginHandle  K8sEventType = "FailedPluginHandle"
)

// TaskActionReconciler reconciles a TaskAction object
type TaskActionReconciler struct {
	client.Client
	Scheme            *runtime.Scheme
	Recorder          record.EventRecorder
	PluginRegistry    *plugin.Registry
	DataStore         *storage.DataStore
	SecretManager     pluginsCore.SecretManager
	ResourceManager   pluginsCore.ResourceManager
	CatalogClient     catalog.AsyncClient
	Catalog           catalog.Client
	eventsClient      workflowconnect.EventsProxyServiceClient
	cluster           string
	MaxSystemFailures uint32
}

// isSystemRetryableFailure reports whether the plugin transition is a
// PhaseRetryableFailure with kind=SYSTEM (as produced by PhaseInfoSystemRetryableFailure).
func isSystemRetryableFailure(phaseInfo pluginsCore.PhaseInfo) bool {
	if phaseInfo.Phase() != pluginsCore.PhaseRetryableFailure {
		return false
	}
	execErr := phaseInfo.Err()
	return execErr != nil && execErr.GetKind() == core.ExecutionError_SYSTEM
}

func systemErrorFromPhaseInfo(phaseInfo pluginsCore.PhaseInfo) error {
	if execErr := phaseInfo.Err(); execErr != nil {
		return fmt.Errorf("[%s] %s", execErr.GetCode(), execErr.GetMessage())
	}
	return fmt.Errorf("system retryable failure")
}

func (r *TaskActionReconciler) maxSystemFailures() uint32 {
	if r.MaxSystemFailures == 0 {
		return DefaultMaxSystemFailures
	}
	return r.MaxSystemFailures
}

// resetPluginResource aborts any in-flight plugin resource and clears persisted
// plugin state so the next reconcile starts fresh from PluginPhaseNotStarted.
func (r *TaskActionReconciler) resetPluginResource(
	ctx context.Context,
	taskAction *flyteorgv1.TaskAction,
	p pluginsCore.Plugin,
	tCtx pluginsCore.TaskExecutionContext,
) {
	if abortErr := p.Abort(ctx, tCtx); abortErr != nil {
		log.FromContext(ctx).Error(abortErr, "failed to abort during system-error reset", "plugin", p.GetID())
	}
	taskAction.Status.PluginState = nil
	taskAction.Status.PluginStateVersion = 0
}

// recordSystemError increments Status.SystemFailures and either requeues for
// another attempt or, once the configured threshold is exceeded, converts the
// TaskAction to a permanent failure. It does not touch the underlying plugin
// resource — callers that observed a failure phase should invoke
// resetPluginResource first.
func (r *TaskActionReconciler) recordSystemError(
	ctx context.Context,
	taskAction *flyteorgv1.TaskAction,
	original *flyteorgv1.TaskAction,
	pluginID string,
	handleErr error,
) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Error(handleErr, "system error from plugin", "plugin", pluginID)
	if r.Recorder != nil {
		r.Recorder.Eventf(taskAction, corev1.EventTypeWarning, string(FailedPluginHandle),
			"Plugin %q system error: %v", pluginID, handleErr)
	}

	taskAction.Status.SystemFailures++
	if taskAction.Status.SystemFailures > r.maxSystemFailures() {
		execErr := &core.ExecutionError{
			Kind:    core.ExecutionError_SYSTEM,
			Code:    MaxSystemFailuresExceededCode,
			Message: fmt.Sprintf("plugin %q failed with system error %d times: %v", pluginID, taskAction.Status.SystemFailures, handleErr),
		}
		return r.finalizePermanentFailure(ctx, taskAction, original, execErr)
	}

	if taskActionStatusChanged(original.Status, taskAction.Status) {
		if updErr := r.Status().Update(ctx, taskAction); updErr != nil {
			logger.Error(updErr, "failed to persist SystemFailures counter")
		}
	}
	return ctrl.Result{RequeueAfter: TaskActionDefaultRequeueDuration}, nil
}

// finalizePermanentFailure converts the TaskAction to a terminal PermanentFailure
// with the given ExecutionError and stamps GC labels.
func (r *TaskActionReconciler) finalizePermanentFailure(
	ctx context.Context,
	taskAction *flyteorgv1.TaskAction,
	original *flyteorgv1.TaskAction,
	execErr *core.ExecutionError,
) (ctrl.Result, error) {
	phaseInfo := pluginsCore.PhaseInfoFailed(pluginsCore.PhasePermanentFailure, execErr, nil)
	mapPhaseToConditions(taskAction, phaseInfo)
	taskAction.Status.PluginPhase = phaseInfo.Phase().String()
	taskAction.Status.PluginPhaseVersion = phaseInfo.Version()
	if updErr := r.updateTaskActionStatus(ctx, original, taskAction, phaseInfo); updErr != nil {
		return ctrl.Result{}, updErr
	}
	if isTerminal(taskAction) {
		if labelErr := r.ensureTerminalLabels(ctx, taskAction); labelErr != nil {
			return ctrl.Result{}, labelErr
		}
	}
	return ctrl.Result{}, nil
}

// NewTaskActionReconciler creates a new TaskActionReconciler
func NewTaskActionReconciler(
	c client.Client,
	scheme *runtime.Scheme,
	registry *plugin.Registry,
	dataStore *storage.DataStore,
	eventsClient workflowconnect.EventsProxyServiceClient,
	cluster string,
) *TaskActionReconciler {
	return &TaskActionReconciler{
		Client:         c,
		Scheme:         scheme,
		PluginRegistry: registry,
		DataStore:      dataStore,
		eventsClient:   eventsClient,
		cluster:        cluster,
	}
}

// +kubebuilder:rbac:groups=flyte.org,resources=taskactions,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=flyte.org,resources=taskactions/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=flyte.org,resources=taskactions/finalizers,verbs=update
// +kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *TaskActionReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// Fetch the TaskAction instance
	taskAction := &flyteorgv1.TaskAction{}
	if err := r.Get(ctx, req.NamespacedName, taskAction); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Please do NOT modify `originalTaskActionInstance` in the following code. This is for checking
	// if the TaskAction instance changes
	originalTaskActionInstance := taskAction.DeepCopy()

	// Handle deletion
	if !taskAction.DeletionTimestamp.IsZero() {
		return r.handleAbortAndFinalize(ctx, taskAction)
	}

	// Check terminal conditions -- short-circuit
	if isTerminal(taskAction) {
		if err := r.ensureTerminalLabels(ctx, taskAction); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}

	// Validate spec fields and resolve plugin before adding the finalizer
	// If either fails, the resource is marked terminal and not requeued — no finalizer to clean up
	p, reason, err := validateTaskAction(taskAction, r.PluginRegistry)
	if err != nil {
		logger.Error(err, "TaskAction validation failed")
		eventType := FailedValidation
		if reason == flyteorgv1.ConditionReasonPluginNotFound {
			eventType = FailedPluginResolve
		}
		r.Recorder.Eventf(taskAction, corev1.EventTypeWarning, string(eventType), "%v", err)
		setCondition(taskAction, flyteorgv1.ConditionTypeFailed, metav1.ConditionTrue, reason, err.Error())
		setCondition(taskAction, flyteorgv1.ConditionTypeProgressing, metav1.ConditionFalse, reason, err.Error())
		_ = r.Status().Update(ctx, taskAction)
		return ctrl.Result{}, nil // terminal — do not requeue
	}

	// Ensure finalizer is present (once validation passes)
	if !controllerutil.ContainsFinalizer(taskAction, taskActionFinalizer) {
		controllerutil.AddFinalizer(taskAction, taskActionFinalizer)
		if err := r.Update(ctx, taskAction); err != nil {
			logger.Error(err, "Failed to update TaskAction with finalizer")
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}

	// Build PluginStateManager from persisted state
	stateMgr := plugin.NewPluginStateManager(
		taskAction.Status.PluginState,
		taskAction.Status.PluginStateVersion,
	)

	// Build TaskExecutionContext
	tCtx, err := plugin.NewTaskExecutionContext(
		taskAction,
		r.DataStore,
		stateMgr,
		r.SecretManager,
		r.ResourceManager,
		r.CatalogClient,
	)
	if err != nil {
		logger.Error(err, "failed to build task execution context")
		return ctrl.Result{RequeueAfter: TaskActionDefaultRequeueDuration}, nil
	}

	// cacheShortCircuited is true when cache handling already decided the outcome,
	// either via cache hit or waiting on the reservation owner.
	var cacheShortCircuited bool
	transition, cacheShortCircuited, err := r.evaluateCacheBeforeExecution(ctx, taskAction, tCtx)
	if err != nil {
		logger.Error(err, "cache pre-execution handling failed")
		return ctrl.Result{RequeueAfter: TaskActionDefaultRequeueDuration}, nil
	}
	// Even when cache handling short-circuits execution, we still continue through the
	// shared reconcile tail below so the derived transition updates conditions, status,
	// and emitted action events in the same way as the normal plugin path.

	// Invoke plugin.Handle only when cache handling did not short-circuit execution.
	if !cacheShortCircuited {
		transition, err = p.Handle(ctx, tCtx)
		if err != nil {
			return r.recordSystemError(ctx, taskAction, originalTaskActionInstance, p.GetID(), err)
		}
	}

	if transition, err = r.finalizeCacheAfterExecution(ctx, taskAction, tCtx, transition, cacheShortCircuited); err != nil {
		logger.Error(err, "cache post-execution handling failed")
		return ctrl.Result{RequeueAfter: TaskActionDefaultRequeueDuration}, nil
	}

	// Map transition phase to TaskAction conditions
	phaseInfo := transition.Info()

	if !cacheShortCircuited && isSystemRetryableFailure(phaseInfo) {
		r.resetPluginResource(ctx, taskAction, p, tCtx)
		return r.recordSystemError(ctx, taskAction, originalTaskActionInstance, p.GetID(), systemErrorFromPhaseInfo(phaseInfo))
	}

	// Reset the consecutive system-failure counter on any non-system-error
	// transition so transient blips earlier in the lifecycle don't accumulate
	// toward the permanent-failure threshold.
	taskAction.Status.SystemFailures = 0

	// In-place task restart on USER-kind retryable failure: bump Status.Attempts
	// and relaunch the pod under the same TaskAction.
	var restartAttempts uint32
	if !cacheShortCircuited && phaseInfo.Phase() == pluginsCore.PhaseRetryableFailure {
		currentAttempts := observedAttempts(taskAction)
		maxAttempts := tCtx.TaskExecutionMetadata().GetMaxAttempts()

		if currentAttempts < maxAttempts {
			// Abort (delete) the current pod before incrementing attempts.
			// tCtx was built with the current attempt number so Abort targets the right pod.
			if abortErr := p.Abort(ctx, tCtx); abortErr != nil {
				logger.Error(abortErr, "failed to abort pod during in-place restart")
			}
			// Track the new attempt count; applied to Status.Attempts after the stateMgr block.
			restartAttempts = currentAttempts + 1
			logger.Info("restarting task in-place", "attempt", currentAttempts+1, "maxAttempts", maxAttempts)
			// Override the transition to Queued so the TaskAction stays non-terminal.
			transition = pluginsCore.DoTransition(pluginsCore.PhaseInfoQueued(time.Now(), pluginsCore.DefaultPhaseVersion, "restarting task"))
			phaseInfo = transition.Info()
		} else {
			// All retries exhausted — convert to a permanent (terminal) failure.
			execErr := phaseInfo.Err()
			if execErr == nil {
				execErr = &core.ExecutionError{
					Kind:    core.ExecutionError_USER,
					Code:    "MaxRetriesExceeded",
					Message: fmt.Sprintf("task failed after %d attempt(s)", currentAttempts),
				}
			}
			transition = pluginsCore.DoTransition(pluginsCore.PhaseInfoFailed(pluginsCore.PhasePermanentFailure, execErr, phaseInfo.Info()))
			phaseInfo = transition.Info()
		}
	}
	mapPhaseToConditions(taskAction, phaseInfo)

	// Update StateJSON for observability
	actionSpec, _ := taskAction.Spec.GetActionSpec()
	if actionSpec != nil {
		taskAction.Status.StateJSON = createStateJSON(actionSpec, phaseInfo.Phase().String())
	}

	// Persist new PluginState
	if newBytes, newVersion, written := stateMgr.GetNewState(); written {
		taskAction.Status.PluginState = newBytes
		taskAction.Status.PluginStateVersion = newVersion
	}

	// If an in-place restart was triggered, increment attempts and clear plugin state so the
	// next reconcile starts fresh with PluginPhaseNotStarted and creates a new pod.
	if restartAttempts > 0 {
		taskAction.Status.Attempts = restartAttempts
		taskAction.Status.PluginState = nil
		taskAction.Status.PluginStateVersion = 0
	}

	taskAction.Status.PluginPhase = phaseInfo.Phase().String()
	taskAction.Status.PluginPhaseVersion = phaseInfo.Version()
	taskAction.Status.Attempts = observedAttempts(taskAction)
	taskAction.Status.CacheStatus = observedCacheStatus(phaseInfo.Info())

	if err := r.updateTaskActionStatus(ctx, originalTaskActionInstance, taskAction, phaseInfo); err != nil {
		return ctrl.Result{}, err
	}

	// If the TaskAction just became terminal, stamp GC labels
	if isTerminal(taskAction) {
		if err := r.ensureTerminalLabels(ctx, taskAction); err != nil {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{RequeueAfter: TaskActionDefaultRequeueDuration}, nil
}

// ensureTerminalLabels adds GC-related labels to a terminal TaskAction if not already present.
// This is idempotent — if the labels are already set, it's a no-op.
// Uses a MergeFrom patch instead of a full Update to reduce conflict surface with concurrent reconciles.
func (r *TaskActionReconciler) ensureTerminalLabels(ctx context.Context, taskAction *flyteorgv1.TaskAction) error {
	labels := taskAction.GetLabels()
	if labels != nil && labels[LabelTerminationStatus] == LabelValueTerminated && labels[LabelCompletedTime] != "" {
		return nil // already labeled
	}

	patch := client.MergeFrom(taskAction.DeepCopy())

	if labels == nil {
		labels = make(map[string]string)
	}
	labels[LabelTerminationStatus] = LabelValueTerminated
	labels[LabelCompletedTime] = terminalTransitionTime(taskAction).Format(labelTimeFormat)
	taskAction.SetLabels(labels)

	if err := r.Patch(ctx, taskAction, patch); err != nil {
		log.FromContext(ctx).Error(err, "failed to set terminal labels on TaskAction")
		return err
	}
	return nil
}

// handleAbortAndFinalize handles the deletion of a TaskAction by aborting and finalizing the plugin.
func (r *TaskActionReconciler) handleAbortAndFinalize(ctx context.Context, taskAction *flyteorgv1.TaskAction) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	if !controllerutil.ContainsFinalizer(taskAction, taskActionFinalizer) {
		return ctrl.Result{}, nil
	}

	p, err := r.PluginRegistry.ResolvePlugin(taskAction.Spec.TaskType)
	if err != nil {
		logger.Info("Cannot resolve plugin for abort/finalize, removing finalizer", "error", err)
		return r.removeFinalizer(ctx, taskAction)
	}

	stateMgr := plugin.NewPluginStateManager(
		taskAction.Status.PluginState,
		taskAction.Status.PluginStateVersion,
	)

	tCtx, err := plugin.NewTaskExecutionContext(
		taskAction, r.DataStore, stateMgr, r.SecretManager, r.ResourceManager, r.CatalogClient,
	)
	if err != nil {
		logger.Error(err, "failed to build context for abort/finalize")
		r.Recorder.Eventf(taskAction, corev1.EventTypeWarning, "FinalizationSkipped",
			"Could not build task execution context; skipping Abort/Finalize. Underlying resources may need manual cleanup: %v", err)
		return r.removeFinalizer(ctx, taskAction)
	}

	if err := p.Abort(ctx, tCtx); err != nil {
		logger.Error(err, "plugin Abort failed, will retry")
		return ctrl.Result{RequeueAfter: TaskActionDefaultRequeueDuration}, nil
	}

	if err := p.Finalize(ctx, tCtx); err != nil {
		logger.Error(err, "plugin Finalize failed, will retry")
		return ctrl.Result{RequeueAfter: TaskActionDefaultRequeueDuration}, nil
	}

	if cacheCfg, ok, err := buildTaskCacheConfig(ctx, taskAction, tCtx); err != nil {
		logger.Error(err, "failed to build cache config for finalization cleanup")
	} else if ok {
		if err := r.releaseCacheReservation(ctx, cacheCfg); err != nil {
			logger.Error(err, "failed to release cache reservation during finalization cleanup")
		}
	}

	abortTime := time.Now()
	abortPhaseInfo := pluginsCore.PhaseInfoAborted(abortTime, pluginsCore.DefaultPhaseVersion, "aborted")
	actionEvent := r.buildActionEvent(ctx, taskAction, abortPhaseInfo)
	// buildActionEvent derives UpdatedTime from PhaseHistory, which doesn't include the
	// abort transition. Override it so mergeEvents uses the actual abort time as end_time.
	actionEvent.UpdatedTime = timestamppb.New(abortTime)
	if _, err := r.eventsClient.Record(ctx, connect.NewRequest(&workflow.RecordRequest{
		Events: []*workflow.ActionEvent{actionEvent},
	})); err != nil {
		logger.Error(err, "failed to emit abort event, will retry")
		return ctrl.Result{RequeueAfter: TaskActionDefaultRequeueDuration}, nil
	}

	return r.removeFinalizer(ctx, taskAction)
}

func (r *TaskActionReconciler) removeFinalizer(ctx context.Context, taskAction *flyteorgv1.TaskAction) (ctrl.Result, error) {
	controllerutil.RemoveFinalizer(taskAction, taskActionFinalizer)
	if err := r.Update(ctx, taskAction); err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

// updateTaskActionStatus updates the TaskAction status only when the status has changed,
// avoiding unnecessary API calls for unchanged state.
func (r *TaskActionReconciler) updateTaskActionStatus(
	ctx context.Context,
	oldTaskAction, newTaskAction *flyteorgv1.TaskAction,
	phaseInfo pluginsCore.PhaseInfo,
) error {
	logger := log.FromContext(ctx)

	if !taskActionStatusChanged(oldTaskAction.Status, newTaskAction.Status) {
		return nil
	}

	actionEvent := r.buildActionEvent(ctx, newTaskAction, phaseInfo)
	if _, err := r.eventsClient.Record(ctx, connect.NewRequest(&workflow.RecordRequest{
		Events: []*workflow.ActionEvent{actionEvent},
	})); err != nil {
		r.Recorder.Eventf(
			newTaskAction,
			corev1.EventTypeWarning,
			"ActionEventPublishFailed",
			"Failed to persist action event %q: %v",
			actionEvent.GetId().GetName(),
			err,
		)
		logger.Error(err, "failed to persist action event", "action", actionEvent.GetId().GetName())
		return err
	}

	// The retry.RetryOnConflict will refetch the k8s resource to get the latest resource version
	// This will resovle the conflict error caused by k8s optimistic lock when 2 reconcile loops updating the same CRD
	if err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		latest := &flyteorgv1.TaskAction{}
		if getErr := r.Get(ctx, client.ObjectKeyFromObject(newTaskAction), latest); getErr != nil {
			return getErr
		}
		latest.Status = newTaskAction.Status
		return r.Status().Update(ctx, latest)
	}); err != nil {
		logger.Error(err, "Error updating status", "name", oldTaskAction.Name, "error", err, "TaskAction", newTaskAction)
		return err
	}

	return nil
}

func (r *TaskActionReconciler) buildActionEvent(
	ctx context.Context,
	taskAction *flyteorgv1.TaskAction,
	phaseInfo pluginsCore.PhaseInfo,
) *workflow.ActionEvent {
	actionID := &common.ActionIdentifier{
		Run: &common.RunIdentifier{
			Project: taskAction.Spec.Project,
			Domain:  taskAction.Spec.Domain,
			Name:    taskAction.Spec.RunName,
		},
		Name: taskAction.Spec.ActionName,
	}

	info := phaseInfo.Info()
	updatedTime := updatedTimestamp(taskAction.Status.PhaseHistory)

	event := &workflow.ActionEvent{
		Id:            actionID,
		Attempt:       observedAttempts(taskAction),
		Phase:         phaseToActionPhase(phaseInfo.Phase()),
		Version:       phaseInfo.Version(),
		UpdatedTime:   updatedTime,
		ErrorInfo:     toActionErrorInfo(phaseInfo.Err()),
		Cluster:       r.cluster,
		Outputs:       outputRefs(ctx, taskAction),
		ClusterEvents: toClusterEvents(phaseInfo, updatedTime),
		ReportedTime:  timestamppb.New(time.Now()),
	}

	if info != nil {
		event.LogInfo = info.Logs
		event.LogContext = info.LogContext
	}
	event.CacheStatus = observedCacheStatus(info)

	return event
}

func observedAttempts(taskAction *flyteorgv1.TaskAction) uint32 {
	if taskAction.Status.Attempts > 0 {
		return taskAction.Status.Attempts
	}
	// if attempts is not set, default to 1
	return 1
}

func observedCacheStatus(info *pluginsCore.TaskInfo) core.CatalogCacheStatus {
	if info == nil {
		return core.CatalogCacheStatus_CACHE_DISABLED
	}
	return cacheStatusFromExternalResources(info.ExternalResources)
}

func updatedTimestamp(history []flyteorgv1.PhaseTransition) *timestamppb.Timestamp {
	if n := len(history); n > 0 {
		return timestamppb.New(history[n-1].OccurredAt.Time)
	}
	return timestamppb.Now()
}

func outputRefs(ctx context.Context, taskAction *flyteorgv1.TaskAction) *task.OutputReferences {
	if taskAction.Spec.RunOutputBase == "" {
		return nil
	}
	attempt := observedAttempts(taskAction)
	prefix, err := plugin.ComputeActionOutputPath(ctx, taskAction.Namespace, taskAction.Name, taskAction.Spec.RunOutputBase, taskAction.Spec.ActionName, attempt)
	if err != nil {
		return nil
	}
	base := strings.TrimRight(string(prefix), "/")
	return &task.OutputReferences{
		OutputUri: base + "/outputs.pb",
		ReportUri: base + "/report.html",
	}
}

func phaseToActionPhase(phase pluginsCore.Phase) common.ActionPhase {
	switch phase {
	case pluginsCore.PhaseNotReady, pluginsCore.PhaseQueued:
		return common.ActionPhase_ACTION_PHASE_QUEUED
	case pluginsCore.PhaseWaitingForResources, pluginsCore.PhaseWaitingForCache:
		return common.ActionPhase_ACTION_PHASE_WAITING_FOR_RESOURCES
	case pluginsCore.PhaseInitializing:
		return common.ActionPhase_ACTION_PHASE_INITIALIZING
	case pluginsCore.PhaseRunning:
		return common.ActionPhase_ACTION_PHASE_RUNNING
	case pluginsCore.PhaseSuccess:
		return common.ActionPhase_ACTION_PHASE_SUCCEEDED
	case pluginsCore.PhaseRetryableFailure, pluginsCore.PhasePermanentFailure:
		return common.ActionPhase_ACTION_PHASE_FAILED
	case pluginsCore.PhaseAborted:
		return common.ActionPhase_ACTION_PHASE_ABORTED
	default:
		return common.ActionPhase_ACTION_PHASE_UNSPECIFIED
	}
}

func toActionErrorInfo(err *core.ExecutionError) *workflow.ErrorInfo {
	if err == nil {
		return nil
	}
	out := &workflow.ErrorInfo{
		Message: err.GetMessage(),
		Kind:    workflow.ErrorInfo_KIND_UNSPECIFIED,
	}
	switch err.GetKind() {
	case core.ExecutionError_USER:
		out.Kind = workflow.ErrorInfo_KIND_USER
	case core.ExecutionError_SYSTEM:
		out.Kind = workflow.ErrorInfo_KIND_SYSTEM
	}
	return out
}

// errorStateFromExecError converts a plugin core.ExecutionError into the
// CR-persisted ErrorState.
func errorStateFromExecError(err *core.ExecutionError) *flyteorgv1.ErrorState {
	if err == nil {
		return nil
	}
	kind := ""
	switch err.GetKind() {
	case core.ExecutionError_USER:
		kind = "USER"
	case core.ExecutionError_SYSTEM:
		kind = "SYSTEM"
	}
	return &flyteorgv1.ErrorState{
		Code:    err.GetCode(),
		Kind:    kind,
		Message: err.GetMessage(),
	}
}

func toClusterEvents(phaseInfo pluginsCore.PhaseInfo, fallbackTime *timestamppb.Timestamp) []*workflow.ClusterEvent {
	info := phaseInfo.Info()
	if phaseInfo.Reason() == "" && (info == nil || len(info.AdditionalReasons) == 0) {
		return nil
	}

	out := []*workflow.ClusterEvent{}
	if phaseInfo.Reason() != "" {
		e := &workflow.ClusterEvent{
			Message: phaseInfo.Reason(),
		}
		if info != nil && info.OccurredAt != nil {
			e.OccurredAt = timestamppb.New(*info.OccurredAt)
		} else {
			e.OccurredAt = fallbackTime
		}
		out = append(out, e)
	}

	if info == nil {
		return out
	}

	for _, reason := range info.AdditionalReasons {
		e := &workflow.ClusterEvent{
			Message: reason.Reason,
		}
		if reason.OccurredAt != nil {
			e.OccurredAt = timestamppb.New(*reason.OccurredAt)
		} else {
			e.OccurredAt = fallbackTime
		}
		out = append(out, e)
	}
	return out
}

func cacheStatusFromExternalResources(resources []*pluginsCore.ExternalResource) core.CatalogCacheStatus {
	for _, resource := range resources {
		if resource == nil {
			continue
		}
		// Return the first explicit cache status signal.
		if resource.CacheStatus != core.CatalogCacheStatus_CACHE_DISABLED {
			return resource.CacheStatus
		}
	}
	return core.CatalogCacheStatus_CACHE_DISABLED
}

// taskActionStatusChanged reports whether any status field has changed between old and new,
// covering plugin phase, state, state version, observability JSON, conditions, and phase history.
func taskActionStatusChanged(oldStatus, newStatus flyteorgv1.TaskActionStatus) bool {
	if oldStatus.StateJSON != newStatus.StateJSON ||
		oldStatus.PluginStateVersion != newStatus.PluginStateVersion ||
		oldStatus.PluginPhase != newStatus.PluginPhase ||
		oldStatus.PluginPhaseVersion != newStatus.PluginPhaseVersion ||
		oldStatus.Attempts != newStatus.Attempts ||
		oldStatus.SystemFailures != newStatus.SystemFailures ||
		oldStatus.CacheStatus != newStatus.CacheStatus {
		return true
	}

	if !bytes.Equal(oldStatus.PluginState, newStatus.PluginState) {
		return true
	}

	if len(oldStatus.PhaseHistory) != len(newStatus.PhaseHistory) {
		return true
	}

	return !reflect.DeepEqual(oldStatus.Conditions, newStatus.Conditions)
}

// mapPhaseToConditions maps a plugin PhaseInfo to TaskAction conditions.
func mapPhaseToConditions(ta *flyteorgv1.TaskAction, info pluginsCore.PhaseInfo) {
	var phaseName string
	var msg string

	switch info.Phase() {
	case pluginsCore.PhaseNotReady, pluginsCore.PhaseQueued, pluginsCore.PhaseWaitingForResources, pluginsCore.PhaseWaitingForCache:
		phaseName = string(flyteorgv1.ConditionReasonQueued)
		msg = info.Reason()
		setCondition(ta, flyteorgv1.ConditionTypeProgressing, metav1.ConditionTrue,
			flyteorgv1.ConditionReasonQueued, msg)

	case pluginsCore.PhaseInitializing:
		phaseName = string(flyteorgv1.ConditionReasonInitializing)
		msg = info.Reason()
		setCondition(ta, flyteorgv1.ConditionTypeProgressing, metav1.ConditionTrue,
			flyteorgv1.ConditionReasonInitializing, msg)

	case pluginsCore.PhaseRunning:
		phaseName = string(flyteorgv1.ConditionReasonExecuting)
		msg = info.Reason()
		setCondition(ta, flyteorgv1.ConditionTypeProgressing, metav1.ConditionTrue,
			flyteorgv1.ConditionReasonExecuting, msg)

	case pluginsCore.PhaseSuccess:
		phaseName = string(flyteorgv1.ConditionReasonCompleted)
		msg = "TaskAction completed successfully"
		setCondition(ta, flyteorgv1.ConditionTypeProgressing, metav1.ConditionFalse,
			flyteorgv1.ConditionReasonCompleted, "TaskAction has completed")
		setCondition(ta, flyteorgv1.ConditionTypeSucceeded, metav1.ConditionTrue,
			flyteorgv1.ConditionReasonCompleted, msg)

	case pluginsCore.PhasePermanentFailure:
		phaseName = string(flyteorgv1.ConditionReasonPermanentFailure)
		msg = info.Reason()
		if info.Err() != nil {
			msg = info.Err().GetMessage()
			ta.Status.ErrorState = errorStateFromExecError(info.Err())
		}
		setCondition(ta, flyteorgv1.ConditionTypeProgressing, metav1.ConditionFalse,
			flyteorgv1.ConditionReasonPermanentFailure, msg)
		setCondition(ta, flyteorgv1.ConditionTypeFailed, metav1.ConditionTrue,
			flyteorgv1.ConditionReasonPermanentFailure, msg)

	case pluginsCore.PhaseRetryableFailure:
		phaseName = string(flyteorgv1.ConditionReasonRetryableFailure)
		msg = info.Reason()
		if info.Err() != nil {
			msg = info.Err().GetMessage()
			ta.Status.ErrorState = errorStateFromExecError(info.Err())
		}
		setCondition(ta, flyteorgv1.ConditionTypeProgressing, metav1.ConditionTrue,
			flyteorgv1.ConditionReasonRetryableFailure, msg)

	case pluginsCore.PhaseAborted:
		phaseName = string(flyteorgv1.ConditionReasonAborted)
		msg = "TaskAction was aborted"
		setCondition(ta, flyteorgv1.ConditionTypeProgressing, metav1.ConditionFalse,
			flyteorgv1.ConditionReasonAborted, msg)
		setCondition(ta, flyteorgv1.ConditionTypeFailed, metav1.ConditionTrue,
			flyteorgv1.ConditionReasonAborted, msg)
	}

	// Append to PhaseHistory if this is a new phase (dedup by checking last entry).
	if phaseName != "" {
		n := len(ta.Status.PhaseHistory)
		if n == 0 || ta.Status.PhaseHistory[n-1].Phase != phaseName {
			ta.Status.PhaseHistory = append(ta.Status.PhaseHistory, flyteorgv1.PhaseTransition{
				Phase:      phaseName,
				OccurredAt: metav1.Now(),
				Message:    msg,
			})
		}
	}
}

// isTerminal returns true if the TaskAction has reached a terminal condition.
func isTerminal(ta *flyteorgv1.TaskAction) bool {
	for _, cond := range ta.Status.Conditions {
		if cond.Type == string(flyteorgv1.ConditionTypeSucceeded) && cond.Status == metav1.ConditionTrue {
			return true
		}
		if cond.Type == string(flyteorgv1.ConditionTypeFailed) && cond.Status == metav1.ConditionTrue {
			return true
		}
	}
	return false
}

// terminalTransitionTime returns the LastTransitionTime from the terminal condition
// (Succeeded or Failed). Falls back to time.Now().UTC() if no transition time is found.
func terminalTransitionTime(ta *flyteorgv1.TaskAction) time.Time {
	for _, cond := range ta.Status.Conditions {
		if cond.Status != metav1.ConditionTrue {
			continue
		}
		if cond.Type == string(flyteorgv1.ConditionTypeSucceeded) || cond.Type == string(flyteorgv1.ConditionTypeFailed) {
			if !cond.LastTransitionTime.IsZero() {
				return cond.LastTransitionTime.UTC()
			}
			break
		}
	}
	return time.Now().UTC()
}

// createStateJSON creates a simplified state JSON for observability.
func createStateJSON(actionSpec *workflow.ActionSpec, phase string) string {
	state := map[string]interface{}{
		"phase":     phase,
		"actionId":  fmt.Sprintf("%s/%s", actionSpec.ActionId.Run.Name, actionSpec.ActionId.Name),
		"timestamp": time.Now().Format(time.RFC3339),
	}

	stateBytes, err := json.Marshal(state)
	if err != nil {
		return "{}"
	}
	return string(stateBytes)
}

// SetupWithManager sets up the controller with the Manager.
func (r *TaskActionReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&flyteorgv1.TaskAction{}).
		Owns(&corev1.Pod{}).
		Named("taskaction").
		Complete(r)
}

// pluginResolver is satisfied by *plugin.Registry and allows mocking in tests.
type pluginResolver interface {
	ResolvePlugin(taskType string) (pluginsCore.Plugin, error)
}

// validateTaskAction checks that all required spec fields are populated and that a plugin
// is registered for the given task type. Both checks happen before the finalizer is added,
// so a failure here leaves the resource finalizer-free and trivially deletable.
func validateTaskAction(taskAction *flyteorgv1.TaskAction, registry pluginResolver) (pluginsCore.Plugin, flyteorgv1.TaskActionConditionReason, error) {
	var missing []string
	if taskAction.Spec.RunName == "" {
		missing = append(missing, "runName")
	}
	if taskAction.Spec.Project == "" {
		missing = append(missing, "project")
	}
	if taskAction.Spec.Domain == "" {
		missing = append(missing, "domain")
	}
	if taskAction.Spec.ActionName == "" {
		missing = append(missing, "actionName")
	}
	if taskAction.Spec.TaskType == "" {
		missing = append(missing, "taskType")
	}
	if len(taskAction.Spec.TaskTemplate) == 0 {
		missing = append(missing, "taskTemplate")
	}
	if taskAction.Spec.InputURI == "" {
		missing = append(missing, "inputUri")
	}
	if taskAction.Spec.RunOutputBase == "" {
		missing = append(missing, "runOutputBase")
	}
	if len(missing) > 0 {
		return nil, flyteorgv1.ConditionReasonInvalidSpec,
			fmt.Errorf("required spec fields are empty: %v", missing)
	}

	p, err := registry.ResolvePlugin(taskAction.Spec.TaskType)
	if err != nil {
		return nil, flyteorgv1.ConditionReasonPluginNotFound,
			fmt.Errorf("no plugin found for task type %q: %w", taskAction.Spec.TaskType, err)
	}

	return p, "", nil
}

// setCondition sets or updates a condition on the TaskAction.
func setCondition(taskAction *flyteorgv1.TaskAction, conditionType flyteorgv1.TaskActionConditionType, status metav1.ConditionStatus, reason flyteorgv1.TaskActionConditionReason, message string) {
	condition := metav1.Condition{
		Type:    string(conditionType),
		Status:  status,
		Reason:  string(reason),
		Message: message,
	}
	meta.SetStatusCondition(&taskAction.Status.Conditions, condition)
}
