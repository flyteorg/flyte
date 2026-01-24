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
	"sigs.k8s.io/controller-runtime/pkg/log"

	flyteorgv1 "github.com/flyteorg/flyte/v2/executor/api/v1"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/workflow"
)

type K8sEventType string

const (
	TaskActionDefaultRequeueDuration              = 5 * time.Second
	FailedUnmarshal                  K8sEventType = "FailedUnmarshal"
)

// TaskActionReconciler reconciles a TaskAction object
type TaskActionReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

// NewTaskActionReconciler creates a new TaskActionReconciler
func NewTaskActionReconciler(c client.Client, scheme *runtime.Scheme) *TaskActionReconciler {
	return &TaskActionReconciler{
		Client: c,
		Scheme: scheme,
	}
}

// +kubebuilder:rbac:groups=flyte.org,resources=taskactions,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=flyte.org,resources=taskactions/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=flyte.org,resources=taskactions/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *TaskActionReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// Fetch the TaskAction instance
	taskAction := &flyteorgv1.TaskAction{}
	if err := r.Get(ctx, req.NamespacedName, taskAction); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Get the ActionSpec from the TaskAction
	actionSpec, err := taskAction.Spec.GetActionSpec()
	if err != nil {
		r.Recorder.Eventf(taskAction, corev1.EventTypeWarning, string(FailedUnmarshal), "Failed to unmarshal ActionSpec %s/%s: %v", taskAction.Namespace, taskAction.Name, err)
		return ctrl.Result{}, err
	}

	// TODO (haytham): Remove when we add real code that executes plugins. For now this is here so that watchers can see
	//  things transition between states.
	time.Sleep(2 * time.Second)

	// Check terminal conditions first
	succeededCond := findConditionByType(taskAction.Status.Conditions, flyteorgv1.ConditionTypeSucceeded)
	if succeededCond != nil && succeededCond.Status == metav1.ConditionTrue {
		logger.Info("TaskAction already succeeded",
			"name", taskAction.Name, "action", actionSpec.ActionId.Name)
		return ctrl.Result{}, nil
	}

	failedCond := findConditionByType(taskAction.Status.Conditions, flyteorgv1.ConditionTypeFailed)
	if failedCond != nil && failedCond.Status == metav1.ConditionTrue {
		logger.Info("TaskAction already failed",
			"name", taskAction.Name, "action", actionSpec.ActionId.Name)
		return ctrl.Result{}, nil
	}

	// Sequential condition evaluation for Progressing
	progressingCond := findConditionByType(taskAction.Status.Conditions, flyteorgv1.ConditionTypeProgressing)
	if progressingCond != nil && progressingCond.Status == metav1.ConditionTrue {
		// Check the Reason to determine sub-state
		if progressingCond.Reason == string(flyteorgv1.ConditionReasonExecuting) {
			// Executing to Succeeded
			logger.Info("TaskAction is executing, transitioning to Succeeded",
				"name", taskAction.Name, "action", actionSpec.ActionId.Name)

			setCondition(taskAction, flyteorgv1.ConditionTypeProgressing, metav1.ConditionFalse,
				flyteorgv1.ConditionReasonCompleted, "TaskAction has completed")
			setCondition(taskAction, flyteorgv1.ConditionTypeSucceeded, metav1.ConditionTrue,
				flyteorgv1.ConditionReasonCompleted, "TaskAction completed successfully")

			taskAction.Status.StateJSON = createStateJSON(actionSpec, "Succeeded")

			if err := r.Status().Update(ctx, taskAction); err != nil {
				return ctrl.Result{}, err
			}
			return ctrl.Result{}, nil
		}

		if progressingCond.Reason == string(flyteorgv1.ConditionReasonInitializing) {
			// Initializing to Executing
			logger.Info("TaskAction is initializing, transitioning to Executing",
				"name", taskAction.Name, "action", actionSpec.ActionId.Name)

			setCondition(taskAction, flyteorgv1.ConditionTypeProgressing, metav1.ConditionTrue,
				flyteorgv1.ConditionReasonExecuting, "TaskAction is executing")

			taskAction.Status.StateJSON = createStateJSON(actionSpec, "Running")

			if err := r.Status().Update(ctx, taskAction); err != nil {
				return ctrl.Result{}, err
			}
			return ctrl.Result{RequeueAfter: 5 * time.Second}, nil
		}

		if progressingCond.Reason == string(flyteorgv1.ConditionReasonQueued) {
			// Queued to Initializing
			logger.Info("TaskAction is queued, transitioning to Initializing",
				"name", taskAction.Name, "action", actionSpec.ActionId.Name)

			setCondition(taskAction, flyteorgv1.ConditionTypeProgressing, metav1.ConditionTrue,
				flyteorgv1.ConditionReasonInitializing, "TaskAction is being initialized")

			taskAction.Status.StateJSON = createStateJSON(actionSpec, "Initializing")

			if err := r.Status().Update(ctx, taskAction); err != nil {
				return ctrl.Result{}, err
			}
			return ctrl.Result{RequeueAfter: 5 * time.Second}, nil
		}
	}

	// No conditions exist, this is the first reconcile
	// Set Condition to Queued
	logger.Info("New TaskAction, setting Progressing condition",
		"name", taskAction.Name, "action", actionSpec.ActionId.Name)

	setCondition(taskAction, flyteorgv1.ConditionTypeProgressing, metav1.ConditionTrue,
		flyteorgv1.ConditionReasonQueued, "TaskAction is queued and waiting for resources")

	taskAction.Status.StateJSON = createStateJSON(actionSpec, "Queued")

	if err := r.Status().Update(ctx, taskAction); err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{RequeueAfter: TaskActionDefaultRequeueDuration}, nil
}

// createStateJSON creates a simplified NodeStatus JSON representation
func createStateJSON(actionSpec *workflow.ActionSpec, phase string) string {
	// Create a simplified state object
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
		Named("taskaction").
		Complete(r)
}

// findConditionByType finds a condition by type in the conditions list
// Returns nil if not found
func findConditionByType(conditions []metav1.Condition, condType flyteorgv1.TaskActionConditionType) *metav1.Condition {
	for i := range conditions {
		if conditions[i].Type == string(condType) {
			return &conditions[i]
		}
	}
	return nil
}

// setCondition sets or updates a condition on the TaskAction
func setCondition(taskAction *flyteorgv1.TaskAction, conditionType flyteorgv1.TaskActionConditionType, status metav1.ConditionStatus, reason flyteorgv1.TaskActionConditionReason, message string) {
	condition := metav1.Condition{
		Type:    string(conditionType),
		Status:  status,
		Reason:  string(reason),
		Message: message,
	}
	meta.SetStatusCondition(&taskAction.Status.Conditions, condition)
}
