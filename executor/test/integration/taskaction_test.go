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
	"go.opentelemetry.io/otel/metric/noop"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/durationpb"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/events"
	testingclock "k8s.io/utils/clock/testing"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	actionsk8s "github.com/flyteorg/flyte/v2/actions/k8s"
	flyteorgv1 "github.com/flyteorg/flyte/v2/executor/api/v1"
	"github.com/flyteorg/flyte/v2/executor/pkg/controller"
	pluginsCore "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/core"
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

func taskTemplateBytes(t *testing.T, taskType, image string) []byte {
	t.Helper()
	data, err := proto.Marshal(&core.TaskTemplate{
		Type: taskType,
		Target: &core.TaskTemplate_Container{
			Container: &core.Container{
				Image:   image,
				Command: []string{"echo"},
				Args:    []string{"hello"},
			},
		},
		Metadata: &core.TaskMetadata{
			Runtime: &core.RuntimeMetadata{Type: core.RuntimeMetadata_FLYTE_SDK},
		},
		Interface: &core.TypedInterface{},
	})
	require.NoError(t, err)
	return data
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

func TestQueuedTimeoutReconcile(t *testing.T) {
	const budget = 2 * time.Second
	cases := []struct {
		name                 string
		phase                pluginsCore.Phase
		transitionAt         time.Duration
		delayedStart         bool
		observedInitializing bool
		wantTimedOut         bool
	}{
		{name: "still queued", phase: pluginsCore.PhaseQueued, wantTimedOut: true},
		{name: "controller unavailable at enqueue", delayedStart: true, wantTimedOut: true},
		{name: "initializing before deadline", phase: pluginsCore.PhaseInitializing, transitionAt: time.Second},
		{name: "initializing at deadline", phase: pluginsCore.PhaseInitializing, transitionAt: budget},
		{
			name: "initializing after deadline", phase: pluginsCore.PhaseInitializing,
			transitionAt: budget + time.Second, wantTimedOut: true,
		},
		{
			name: "initialization already observed", phase: pluginsCore.PhaseInitializing,
			transitionAt: time.Second, observedInitializing: true,
		},
		{name: "running before recovery", phase: pluginsCore.PhaseRunning, transitionAt: time.Second},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			runID := &common.RunIdentifier{
				Org: "test-org", Project: "test-project", Domain: "development", Name: uniqueRunName("qt-run"),
			}
			action := newRootTaskAction(runID)
			action.GetTask().GetSpec().GetTaskTemplate().GetMetadata().Timeouts = &core.TimeoutStrategy{
				QueuedTimeout: durationpb.New(budget),
			}
			require.NoError(t, actionsClient.Enqueue(ctx, action, nil))
			key := types.NamespacedName{Name: runID.Name + "-a0", Namespace: "flyte"}
			req := reconcile.Request{NamespacedName: key}
			ta := &flyteorgv1.TaskAction{}
			require.NoError(t, k8sClient.Get(ctx, key, ta))
			enqueuedAt := ta.CreationTimestamp.Time
			fakeClock := testingclock.NewFakeClock(enqueuedAt)
			r := *reconciler
			r.Clock = fakeClock
			if tc.delayedStart {
				fakeClock.Step(budget + time.Second)
			}
			for range 2 {
				_, err := r.Reconcile(ctx, req)
				require.NoError(t, err)
			}
			podKey := types.NamespacedName{Name: key.Name + "-0", Namespace: key.Namespace}
			if !tc.delayedStart {
				require.NoError(t, k8sClient.Get(ctx, key, ta))
				require.Equal(t, pluginsCore.PhaseQueued.String(), ta.Status.PluginPhase)
				pod := &corev1.Pod{}
				require.NoError(t, k8sClient.Get(ctx, podKey, pod))
				occurredAt := metav1.NewTime(enqueuedAt.Add(tc.transitionAt))
				// Envtest has no kubelet. Change only its pod status, leaving the
				// TaskAction and plugin state as the real enqueue and Handle paths wrote them.
				pod.Status.Phase = corev1.PodPending
				pod.Status.Conditions = []corev1.PodCondition{{Type: corev1.PodScheduled, Status: corev1.ConditionFalse}}
				switch tc.phase {
				case pluginsCore.PhaseInitializing:
					pod.Status.Conditions = []corev1.PodCondition{{
						Type: corev1.PodReady, Status: corev1.ConditionFalse, LastTransitionTime: occurredAt,
					}}
					pod.Status.ContainerStatuses = []corev1.ContainerStatus{{
						Name:  pod.Spec.Containers[0].Name,
						State: corev1.ContainerState{Waiting: &corev1.ContainerStateWaiting{Reason: "ContainerCreating"}},
					}}
				case pluginsCore.PhaseRunning:
					pod.Status.Phase = corev1.PodRunning
					pod.Status.ContainerStatuses = []corev1.ContainerStatus{{
						Name:  pod.Spec.Containers[0].Name,
						State: corev1.ContainerState{Running: &corev1.ContainerStateRunning{StartedAt: occurredAt}},
					}}
				}
				require.NoError(t, k8sClient.Status().Update(ctx, pod))
				waitForPodStatus := func() {
					require.Eventually(t, func() bool {
						cached := &corev1.Pod{}
						return pluginClient.Get(ctx, podKey, cached) == nil && cached.ResourceVersion == pod.ResourceVersion
					}, 10*time.Second, 20*time.Millisecond)
				}
				waitForPodStatus()
				if tc.observedInitializing {
					fakeClock.Step(time.Second)
					_, err := r.Reconcile(ctx, req)
					require.NoError(t, err)
					pod.Status.Conditions[0].LastTransitionTime = metav1.NewTime(enqueuedAt.Add(budget + time.Second))
					require.NoError(t, k8sClient.Status().Update(ctx, pod))
					waitForPodStatus()
				}
				fakeClock.SetTime(enqueuedAt.Add(budget + time.Second))
				_, err := r.Reconcile(ctx, req)
				require.NoError(t, err)
			}
			require.NoError(t, k8sClient.Get(ctx, key, ta))
			require.NotNil(t, ta.Status.AttemptQueuedAt)
			assert.Equal(t, enqueuedAt, ta.Status.AttemptQueuedAt.Time)
			t.Logf("created=%s observed=%s phase=%s timeout=%v",
				enqueuedAt, fakeClock.Now(), ta.Status.PluginPhase, ta.Status.TimeoutAt)
			if tc.wantTimedOut {
				require.NotNil(t, ta.Status.ErrorState)
				require.NotNil(t, ta.Status.TimeoutAt)
				assert.Equal(t, controller.TaskExecutionTimedOutCode, ta.Status.ErrorState.Code)
				assert.Contains(t, ta.Status.ErrorState.Message, "queued timeout")
				assert.Equal(t, enqueuedAt.Add(budget), ta.Status.TimeoutAt.Time)
				require.Eventually(t, func() bool {
					return apierrors.IsNotFound(k8sClient.Get(ctx, podKey, &corev1.Pod{}))
				}, 10*time.Second, 20*time.Millisecond)
				require.Eventually(t, func() bool {
					for _, update := range runClient.recordedStatusUpdates() {
						if update.GetActionId().GetName() == runID.Name &&
							update.GetStatus().GetPhase() == common.ActionPhase_ACTION_PHASE_TIMED_OUT {
							return true
						}
					}
					return false
				}, 10*time.Second, 20*time.Millisecond)
			} else {
				assert.Nil(t, ta.Status.TimeoutAt)
				assert.Equal(t, tc.phase.String(), ta.Status.PluginPhase)
				assert.NoError(t, k8sClient.Get(ctx, podKey, &corev1.Pod{}))
			}
		})
	}
}

func TestDeletingTerminalTaskActionDoesNotEmitAborted(t *testing.T) {
	eventsClient := &recordingEventsClient{}
	terminalReconciler := controller.NewTaskActionReconciler(
		k8sClient, scheme.Scheme, pluginRegistry, dataStore,
		eventsClient, "", noop.NewMeterProvider(), nil,
	)
	terminalReconciler.Recorder = events.NewFakeRecorder(10)

	name := fmt.Sprintf("td-%d", time.Now().UnixNano()%1_000_000_000)
	crKey := types.NamespacedName{Name: name, Namespace: "default"}
	ta := &flyteorgv1.TaskAction{
		ObjectMeta: metav1.ObjectMeta{
			Name:       name,
			Namespace:  "default",
			Finalizers: []string{"flyte.org/plugin-finalizer"},
		},
		Spec: flyteorgv1.TaskActionSpec{
			RunName:       name,
			Project:       "test-project",
			Domain:        "development",
			ActionName:    name,
			InputURI:      "/tmp/input",
			RunOutputBase: "/tmp/output",
			TaskType:      "python",
			TaskTemplate:  taskTemplateBytes(t, "python", "python:3.11"),
		},
	}
	require.NoError(t, k8sClient.Create(ctx, ta))
	t.Cleanup(func() {
		current := &flyteorgv1.TaskAction{}
		if err := k8sClient.Get(ctx, crKey, current); err == nil {
			current.Finalizers = nil
			_ = k8sClient.Update(ctx, current)
			_ = k8sClient.Delete(ctx, current)
		}
	})

	ta.Status.Conditions = []metav1.Condition{{
		Type:               string(flyteorgv1.ConditionTypeSucceeded),
		Status:             metav1.ConditionTrue,
		Reason:             string(flyteorgv1.ConditionReasonCompleted),
		Message:            "TaskAction completed successfully",
		LastTransitionTime: metav1.Now(),
	}}
	require.NoError(t, k8sClient.Status().Update(ctx, ta))
	require.NoError(t, k8sClient.Delete(ctx, ta))

	_, err := terminalReconciler.Reconcile(ctx, reconcile.Request{NamespacedName: crKey})
	require.NoError(t, err)

	getErr := k8sClient.Get(ctx, crKey, &flyteorgv1.TaskAction{})
	require.Error(t, getErr)
	assert.Contains(t, getErr.Error(), "not found")

	for _, event := range eventsClient.RecordedEvents() {
		assert.NotEqual(t, common.ActionPhase_ACTION_PHASE_ABORTED, event.GetPhase())
	}
}
