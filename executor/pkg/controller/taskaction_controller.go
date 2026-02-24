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
	"context"
	"encoding/json"
	"fmt"
	"time"

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
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/workflow"
)

const (
	TaskActionDefaultRequeueDuration = 5 * time.Second
	taskActionFinalizer              = "flyte.org/plugin-finalizer"
)

type K8sEventType string

const (
	FailedUnmarshal     K8sEventType = "FailedUnmarshal"
	FailedPluginResolve K8sEventType = "FailedPluginResolve"
	FailedPluginHandle  K8sEventType = "FailedPluginHandle"
)

// TaskActionReconciler reconciles a TaskAction object
type TaskActionReconciler struct {
	client.Client
	Scheme          *runtime.Scheme
	Recorder        record.EventRecorder
	PluginRegistry  *plugin.Registry
	DataStore       *storage.DataStore
	SecretManager   pluginsCore.SecretManager
	ResourceManager pluginsCore.ResourceManager
	CatalogClient   catalog.AsyncClient
}

// NewTaskActionReconciler creates a new TaskActionReconciler
func NewTaskActionReconciler(
	c client.Client,
	scheme *runtime.Scheme,
	registry *plugin.Registry,
	dataStore *storage.DataStore,
) *TaskActionReconciler {
	return &TaskActionReconciler{
		Client:         c,
		Scheme:         scheme,
		PluginRegistry: registry,
		DataStore:      dataStore,
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

	// Handle deletion
	if !taskAction.DeletionTimestamp.IsZero() {
		return r.handleAbortAndFinalize(ctx, taskAction)
	}

	// Check terminal conditions -- short-circuit
	if isTerminal(taskAction) {
		return ctrl.Result{}, nil
	}

	// Ensure finalizer is present
	if !controllerutil.ContainsFinalizer(taskAction, taskActionFinalizer) {
		controllerutil.AddFinalizer(taskAction, taskActionFinalizer)
		if err := r.Update(ctx, taskAction); err != nil {
			logger.Error(err, "Failed to update TaskAction with finalizer")
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}

	// Resolve plugin from registry
	p, err := r.PluginRegistry.ResolvePlugin(taskAction.Spec.TaskType)
	if err != nil {
		logger.Error(err, "failed to resolve plugin", "taskType", taskAction.Spec.TaskType)
		r.Recorder.Eventf(taskAction, corev1.EventTypeWarning, string(FailedPluginResolve),
			"No plugin found for task type %q: %v", taskAction.Spec.TaskType, err)
		setCondition(taskAction, flyteorgv1.ConditionTypeFailed, metav1.ConditionTrue,
			flyteorgv1.ConditionReasonPluginNotFound, err.Error())
		setCondition(taskAction, flyteorgv1.ConditionTypeProgressing, metav1.ConditionFalse,
			flyteorgv1.ConditionReasonPluginNotFound, err.Error())
		err = r.Status().Update(ctx, taskAction)
		if err != nil {
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

	// Invoke plugin.Handle
	transition, err := p.Handle(ctx, tCtx)
	if err != nil {
		logger.Error(err, "plugin Handle failed", "plugin", p.GetID())
		r.Recorder.Eventf(taskAction, corev1.EventTypeWarning, string(FailedPluginHandle),
			"Plugin %q Handle failed: %v", p.GetID(), err)
		return ctrl.Result{RequeueAfter: TaskActionDefaultRequeueDuration}, nil
	}

	// Map transition phase to TaskAction conditions
	phaseInfo := transition.Info()
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

	taskAction.Status.PluginPhase = phaseInfo.Phase().String()
	taskAction.Status.PluginPhaseVersion = phaseInfo.Version()

	// Update status subresource
	if err := r.Status().Update(ctx, taskAction); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{RequeueAfter: TaskActionDefaultRequeueDuration}, nil
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

	return r.removeFinalizer(ctx, taskAction)
}

func (r *TaskActionReconciler) removeFinalizer(ctx context.Context, taskAction *flyteorgv1.TaskAction) (ctrl.Result, error) {
	controllerutil.RemoveFinalizer(taskAction, taskActionFinalizer)
	if err := r.Update(ctx, taskAction); err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

// mapPhaseToConditions maps a plugin PhaseInfo to TaskAction conditions.
func mapPhaseToConditions(ta *flyteorgv1.TaskAction, info pluginsCore.PhaseInfo) {
	switch info.Phase() {
	case pluginsCore.PhaseNotReady, pluginsCore.PhaseQueued, pluginsCore.PhaseWaitingForResources, pluginsCore.PhaseWaitingForCache:
		setCondition(ta, flyteorgv1.ConditionTypeProgressing, metav1.ConditionTrue,
			flyteorgv1.ConditionReasonQueued, info.Reason())

	case pluginsCore.PhaseInitializing:
		setCondition(ta, flyteorgv1.ConditionTypeProgressing, metav1.ConditionTrue,
			flyteorgv1.ConditionReasonInitializing, info.Reason())

	case pluginsCore.PhaseRunning:
		setCondition(ta, flyteorgv1.ConditionTypeProgressing, metav1.ConditionTrue,
			flyteorgv1.ConditionReasonExecuting, info.Reason())

	case pluginsCore.PhaseSuccess:
		setCondition(ta, flyteorgv1.ConditionTypeProgressing, metav1.ConditionFalse,
			flyteorgv1.ConditionReasonCompleted, "TaskAction has completed")
		setCondition(ta, flyteorgv1.ConditionTypeSucceeded, metav1.ConditionTrue,
			flyteorgv1.ConditionReasonCompleted, "TaskAction completed successfully")

	case pluginsCore.PhasePermanentFailure:
		msg := info.Reason()
		if info.Err() != nil {
			msg = info.Err().GetMessage()
		}
		setCondition(ta, flyteorgv1.ConditionTypeProgressing, metav1.ConditionFalse,
			flyteorgv1.ConditionReasonPermanentFailure, msg)
		setCondition(ta, flyteorgv1.ConditionTypeFailed, metav1.ConditionTrue,
			flyteorgv1.ConditionReasonPermanentFailure, msg)

	case pluginsCore.PhaseRetryableFailure:
		msg := info.Reason()
		if info.Err() != nil {
			msg = info.Err().GetMessage()
		}
		setCondition(ta, flyteorgv1.ConditionTypeProgressing, metav1.ConditionTrue,
			flyteorgv1.ConditionReasonRetryableFailure, msg)

	case pluginsCore.PhaseAborted:
		setCondition(ta, flyteorgv1.ConditionTypeProgressing, metav1.ConditionFalse,
			flyteorgv1.ConditionReasonAborted, "TaskAction was aborted")
		setCondition(ta, flyteorgv1.ConditionTypeFailed, metav1.ConditionTrue,
			flyteorgv1.ConditionReasonAborted, "TaskAction was aborted")
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
