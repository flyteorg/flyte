/*
Copyright 2026.

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

package integration

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	actionsk8s "github.com/flyteorg/flyte/v2/actions/k8s"
	flyteorgv1 "github.com/flyteorg/flyte/v2/executor/api/v1"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/actions"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/common"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/core"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/task"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/workflow"
)

// uniqueRunName returns a DNS-compliant run name unique per call, so repeated
// runs (-count>1) against the same apiserver never collide on CR names.
func uniqueRunName(prefix string) string {
	return fmt.Sprintf("%s-%d", prefix, time.Now().UnixNano())
}

// newRootTaskAction builds a root task action (action name == run name, CR
// named <run>-a0) whose template the real pod plugin can handle.
func newRootTaskAction(runID *common.RunIdentifier) *actions.Action {
	return &actions.Action{
		ActionId:      &common.ActionIdentifier{Run: runID, Name: runID.Name},
		InputUri:      "/tmp/inputs.pb",
		RunOutputBase: "/tmp/outputs",
		Spec: &actions.Action_Task{
			Task: &workflow.TaskAction{
				Spec: &task.TaskSpec{
					TaskTemplate: &core.TaskTemplate{
						Type: "python",
						Target: &core.TaskTemplate_Container{
							Container: &core.Container{
								Image:   "python:3.11",
								Command: []string{"echo"},
								Args:    []string{"hello"},
							},
						},
						Metadata: &core.TaskMetadata{
							Runtime: &core.RuntimeMetadata{Type: core.RuntimeMetadata_FLYTE_SDK},
						},
						Interface: &core.TypedInterface{},
					},
				},
			},
		},
	}
}

// TestEnqueueReconcile drives the CRD handoff end-to-end on a real apiserver:
// ActionsClient.Enqueue creates a TaskAction CR, the TaskActionReconciler
// accepts it (validation passes, finalizer added, phase written to status),
// and the watch path forwards RecordAction/UpdateActionStatus to the internal
// run service.
func TestEnqueueReconcile(t *testing.T) {
	runID := &common.RunIdentifier{
		Org:     "test-org",
		Project: "test-project",
		Domain:  "development",
		Name:    uniqueRunName("enq-run"),
	}
	require.NoError(t, actionsClient.Enqueue(ctx, newRootTaskAction(runID), nil))

	// Enqueue must produce a CR whose spec passes the reconciler's validation
	// (this is the schema contract between the two halves).
	crKey := types.NamespacedName{Name: runID.Name + "-a0", Namespace: "flyte"}
	ta := &flyteorgv1.TaskAction{}
	require.NoError(t, k8sClient.Get(ctx, crKey, ta))
	assert.Equal(t, runID.Name, ta.Spec.RunName)
	assert.Equal(t, runID.Project, ta.Spec.Project)
	assert.Equal(t, runID.Domain, ta.Spec.Domain)
	assert.Equal(t, runID.Name, ta.Spec.ActionName)
	assert.Equal(t, "python", ta.Spec.TaskType)
	assert.NotEmpty(t, ta.Spec.TaskTemplate)
	assert.Equal(t, "/tmp/inputs.pb", ta.Spec.InputURI)
	assert.Equal(t, "/tmp/outputs", ta.Spec.RunOutputBase)

	// First reconcile adds the finalizer.
	req := reconcile.Request{NamespacedName: crKey}
	_, err := reconciler.Reconcile(ctx, req)
	require.NoError(t, err)
	require.NoError(t, k8sClient.Get(ctx, crKey, ta))
	assert.NotEmpty(t, ta.Finalizers, "reconciler should add a finalizer")

	// Second reconcile drives the plugin Handle path and writes phase conditions.
	_, err = reconciler.Reconcile(ctx, req)
	require.NoError(t, err)
	require.NoError(t, k8sClient.Get(ctx, crKey, ta))
	assert.NotEmpty(t, ta.Status.Conditions, "reconciler should write phase conditions to status")

	// The ActionsClient watcher must forward the handoff back to the internal
	// run service: RecordAction on CR creation, UpdateActionStatus once the
	// reconciler writes a meaningful phase.
	isOurs := func(id *common.ActionIdentifier) bool {
		return id.GetRun().GetName() == runID.Name && id.GetName() == runID.Name
	}
	require.Eventually(t, func() bool {
		for _, r := range runClient.recordedActions() {
			if isOurs(r.GetActionId()) {
				return true
			}
		}
		return false
	}, 15*time.Second, 200*time.Millisecond, "RecordAction should reach the internal run service")

	require.Eventually(t, func() bool {
		for _, u := range runClient.recordedStatusUpdates() {
			if isOurs(u.GetActionId()) && u.GetStatus().GetPhase() != common.ActionPhase_ACTION_PHASE_UNSPECIFIED {
				return true
			}
		}
		return false
	}, 15*time.Second, 200*time.Millisecond, "UpdateActionStatus with a meaningful phase should reach the internal run service")
}

// TestWatchPropagation verifies the informer half of the handoff: when the
// reconciler writes phase conditions to a TaskAction's status, a Subscribe
// channel on the ActionsClient delivers an ActionUpdate carrying that phase.
func TestWatchPropagation(t *testing.T) {
	runID := &common.RunIdentifier{
		Org:     "test-org",
		Project: "test-project",
		Domain:  "development",
		Name:    uniqueRunName("watch-run"),
	}

	// Subscribe before enqueueing so no update can be missed. Root actions
	// have an empty parent action name.
	ch := actionsClient.Subscribe(runID.Name, "")
	defer actionsClient.Unsubscribe(runID.Name, "", ch)

	require.NoError(t, actionsClient.Enqueue(ctx, newRootTaskAction(runID), nil))

	// Drive the executor: two reconciles take the CR to a plugin-reported phase.
	crKey := types.NamespacedName{Name: runID.Name + "-a0", Namespace: "flyte"}
	req := reconcile.Request{NamespacedName: crKey}
	for range 2 {
		_, err := reconciler.Reconcile(ctx, req)
		require.NoError(t, err)
	}

	// The phase the reconciler actually persisted is what subscribers must see.
	ta := &flyteorgv1.TaskAction{}
	require.NoError(t, k8sClient.Get(ctx, crKey, ta))
	expectedPhase := actionsk8s.GetPhaseFromConditions(ta)
	require.NotEqual(t, common.ActionPhase_ACTION_PHASE_UNSPECIFIED, expectedPhase,
		"reconciler should have written a meaningful phase")

	// Earlier events (Added, finalizer update) may deliver updates with an
	// unspecified phase first; wait for the one carrying the reconciled phase.
	deadline := time.After(15 * time.Second)
	for {
		select {
		case update, ok := <-ch:
			require.True(t, ok, "subscription channel closed unexpectedly")
			if update.Phase != expectedPhase {
				continue
			}
			assert.Equal(t, runID.Name, update.ActionID.GetRun().GetName())
			assert.Equal(t, runID.Name, update.ActionID.GetName())
			assert.Equal(t, "python", update.TaskType)
			return
		case <-deadline:
			t.Fatalf("timed out waiting for ActionUpdate with phase %v", expectedPhase)
		}
	}
}

// TestImagePullFailureDeletesPod verifies that a task pod stuck pulling its image is
// deleted when the action fails, so it cannot start once the image becomes available.
func TestImagePullFailureDeletesPod(t *testing.T) {
	runID := &common.RunIdentifier{
		Org:     "test-org",
		Project: "test-project",
		Domain:  "development",
		Name:    uniqueRunName("pull-fail"),
	}
	require.NoError(t, actionsClient.Enqueue(ctx, newRootTaskAction(runID), nil))

	key := types.NamespacedName{Name: runID.Name + "-a0", Namespace: "flyte"}
	req := reconcile.Request{NamespacedName: key}
	ta := &flyteorgv1.TaskAction{}
	pod := &corev1.Pod{}
	require.Eventually(t, func() bool {
		if _, err := reconciler.Reconcile(ctx, req); err != nil {
			return false
		}
		if err := k8sClient.Get(ctx, key, ta); err != nil {
			return false
		}
		pods := &corev1.PodList{}
		if err := k8sClient.List(ctx, pods, client.InNamespace(key.Namespace)); err != nil {
			return false
		}
		for i := range pods.Items {
			for _, ref := range pods.Items[i].OwnerReferences {
				if ref.UID == ta.UID {
					*pod = pods.Items[i]
					return true
				}
			}
		}
		return false
	}, 15*time.Second, 200*time.Millisecond, "the pod plugin should create the task pod")

	stuckSince := metav1.NewTime(time.Now().Add(-time.Hour))
	pod.Status = corev1.PodStatus{
		Phase: corev1.PodPending,
		Conditions: []corev1.PodCondition{
			{Type: corev1.PodScheduled, Status: corev1.ConditionTrue, LastTransitionTime: stuckSince},
			{Type: corev1.PodReady, Status: corev1.ConditionFalse, Reason: "ContainersNotReady", LastTransitionTime: stuckSince},
		},
		ContainerStatuses: []corev1.ContainerStatus{{
			Name:  pod.Spec.Containers[0].Name,
			Image: pod.Spec.Containers[0].Image,
			State: corev1.ContainerState{Waiting: &corev1.ContainerStateWaiting{
				Reason:  "ImagePullBackOff",
				Message: "Back-off pulling image",
			}},
		}},
	}
	require.NoError(t, k8sClient.Status().Update(ctx, pod))

	require.Eventually(t, func() bool {
		if _, err := reconciler.Reconcile(ctx, req); err != nil {
			return false
		}
		if err := k8sClient.Get(ctx, key, ta); err != nil {
			return false
		}
		return apimeta.IsStatusConditionTrue(ta.Status.Conditions, string(flyteorgv1.ConditionTypeFailed))
	}, 15*time.Second, 200*time.Millisecond, "the action should fail once the image pull grace period has passed")

	err := k8sClient.Get(ctx, client.ObjectKeyFromObject(pod), &corev1.Pod{})
	t.Logf("pod %s after the action failed: %v", pod.Name, err)
	assert.True(t, apierrors.IsNotFound(err), "the failed action's pod %s should be deleted, got %v", pod.Name, err)
}
