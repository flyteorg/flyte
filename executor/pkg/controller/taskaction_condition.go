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
	"fmt"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log"

	flyteorgv1 "github.com/flyteorg/flyte/v2/executor/api/v1"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/workflow"
	"google.golang.org/protobuf/proto"
)

// ConditionTimedOutCode is the ErrorState.Code stamped on condition actions
// that pass their deadline without being signalled.
const ConditionTimedOutCode = "ConditionTimedOut"

// reconcileCondition drives a HITL condition action. Conditions have no plugin
// and no pod; the CR is a pure state machine:
//
//	(created) → Paused → Signaled (Succeeded)  when status.signalValue is set
//	                   → TimedOut (Failed)     when past CreationTimestamp+timeout
//
// The Signal RPC only writes status.signalValue/signalledBy/signalledAt; this
// reconciler is the single writer of status.conditions[]. Abort is CR deletion
// (no finalizer — nothing external to clean up), handled by the shared
// deletion arm before dispatch.
func (r *TaskActionReconciler) reconcileCondition(
	ctx context.Context,
	taskAction *flyteorgv1.TaskAction,
	original *flyteorgv1.TaskAction,
) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// Received signal, transit to Succeeded.
	if len(taskAction.Status.SignalValue) > 0 {
		msg := fmt.Sprintf("condition signalled by %q", taskAction.Status.SignalledBy)
		setCondition(taskAction, flyteorgv1.ConditionTypeProgressing, metav1.ConditionFalse,
			flyteorgv1.ConditionReasonSignaled, msg)
		setCondition(taskAction, flyteorgv1.ConditionTypeSucceeded, metav1.ConditionTrue,
			flyteorgv1.ConditionReasonSignaled, msg)
		appendPhaseHistory(taskAction, string(flyteorgv1.ConditionReasonSignaled), msg)
		return r.finalizeConditionUpdate(ctx, original, taskAction)
	}

	deadline, hasDeadline, err := conditionDeadline(taskAction)
	if err != nil {
		// Malformed conditionSpec — unsignalable, mark terminal like task validation failures.
		logger.Error(err, "invalid conditionSpec")
		msg := err.Error()
		setCondition(taskAction, flyteorgv1.ConditionTypeProgressing, metav1.ConditionFalse,
			flyteorgv1.ConditionReasonInvalidSpec, msg)
		setCondition(taskAction, flyteorgv1.ConditionTypeFailed, metav1.ConditionTrue,
			flyteorgv1.ConditionReasonInvalidSpec, msg)
		taskAction.Status.ErrorState = &flyteorgv1.ErrorState{
			Code: string(flyteorgv1.ConditionReasonInvalidSpec), Kind: "USER", Message: msg,
		}
		return r.finalizeConditionUpdate(ctx, original, taskAction)
	}

	if hasDeadline && !time.Now().Before(deadline) {
		msg := fmt.Sprintf("condition timed out after %s without a signal", time.Since(taskAction.CreationTimestamp.Time).Round(time.Second))
		setCondition(taskAction, flyteorgv1.ConditionTypeProgressing, metav1.ConditionFalse,
			flyteorgv1.ConditionReasonTimedOut, msg)
		setCondition(taskAction, flyteorgv1.ConditionTypeFailed, metav1.ConditionTrue,
			flyteorgv1.ConditionReasonTimedOut, msg)
		taskAction.Status.ErrorState = &flyteorgv1.ErrorState{
			Code: ConditionTimedOutCode, Kind: "USER", Message: msg,
		}
		appendPhaseHistory(taskAction, string(flyteorgv1.ConditionReasonTimedOut), msg)
		return r.finalizeConditionUpdate(ctx, original, taskAction)
	}

	// Not terminal, set as Paused
	setCondition(taskAction, flyteorgv1.ConditionTypeProgressing, metav1.ConditionTrue,
		flyteorgv1.ConditionReasonPaused, "waiting for signal")
	appendPhaseHistory(taskAction, string(flyteorgv1.ConditionReasonPaused), "waiting for signal")
	if err := r.updateConditionStatus(ctx, original, taskAction); err != nil {
		return ctrl.Result{}, err
	}
	if hasDeadline {
		return ctrl.Result{RequeueAfter: time.Until(deadline)}, nil
	}
	return ctrl.Result{}, nil
}

// finalizeConditionUpdate persists a terminal condition status and stamps GC labels.
func (r *TaskActionReconciler) finalizeConditionUpdate(
	ctx context.Context,
	original, taskAction *flyteorgv1.TaskAction,
) (ctrl.Result, error) {
	if err := r.updateConditionStatus(ctx, original, taskAction); err != nil {
		return ctrl.Result{}, err
	}
	if isTerminal(taskAction) {
		if err := r.ensureTerminalLabels(ctx, taskAction); err != nil {
			return ctrl.Result{}, err
		}
	}
	return ctrl.Result{}, nil
}

// updateConditionStatus persists reconciler-owned status fields (conditions,
// phase history, error state) with conflict retry. Unlike updateTaskActionStatus
// it applies fields selectively onto the refetched object: the Signal RPC writes
// signalValue/signalledBy/signalledAt concurrently, and a wholesale
// latest.Status = new.Status would erase a signal that landed after our Get.
func (r *TaskActionReconciler) updateConditionStatus(ctx context.Context, old, new *flyteorgv1.TaskAction) error {
	if !taskActionStatusChanged(old.Status, new.Status) {
		return nil
	}
	if err := r.persistStatusWithRetry(ctx, new, func(latest *flyteorgv1.TaskAction) {
		latest.Status.Conditions = new.Status.Conditions
		latest.Status.PhaseHistory = new.Status.PhaseHistory
		latest.Status.ErrorState = new.Status.ErrorState
	}); err != nil {
		log.FromContext(ctx).Error(err, "failed to update condition status", "name", new.Name)
		return err
	}
	return nil
}

// conditionDeadline returns the timeout deadline derived from the CR
// (CreationTimestamp + conditionSpec.timeout). hasDeadline is false when the
// spec declares no timeout.
func conditionDeadline(taskAction *flyteorgv1.TaskAction) (time.Time, bool, error) {
	if len(taskAction.Spec.ConditionSpec) == 0 {
		return time.Time{}, false, fmt.Errorf("condition action %q has empty conditionSpec", taskAction.Name)
	}
	spec := &workflow.ConditionAction{}
	if err := proto.Unmarshal(taskAction.Spec.ConditionSpec, spec); err != nil {
		return time.Time{}, false, fmt.Errorf("failed to unmarshal conditionSpec: %w", err)
	}
	timeout := spec.GetTimeout().AsDuration()
	if timeout <= 0 {
		return time.Time{}, false, nil
	}
	return taskAction.CreationTimestamp.Add(timeout), true, nil
}
