package clustered

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	jobsetv1alpha2 "sigs.k8s.io/jobset/api/jobset/v1alpha2"

	pluginsCore "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/gang"
	plugink8s "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/k8s"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/core"
	clusteredpb "github.com/flyteorg/flyte/v2/gen/go/flyteidl2/plugins"
)

// Tests for the admission-gate (Kueue) paths: a JobSet held through spec.suspend
// must never report Running, a hold never moves the phase backwards, and a
// suspension after the gang had started is a GangEvicted system retry.

// primaryContainerName is the container name the demystify helpers inspect in these fixtures.
const primaryContainerName = "primary"

func twoNodeSpec() *core.TaskTemplate {
	return buildTaskTemplate(&clusteredpb.ClusteredTaskSpec{Replicas: 2, NprocPerNode: 1})
}

func TestGetTaskPhase_SuspendedAtCreation_NoCondition_WaitingForResources(t *testing.T) {
	// The first poll can run before the JobSet controller writes Suspended=True; the
	// spec alone must be enough to recognise the hold.
	js := makeJobSet("", "", true)
	pCtx := dummyPluginCtx(twoNodeSpec(), emptyK8sReader())

	phase, err := clusteredResourceHandler{}.GetTaskPhase(context.Background(), pCtx, js)
	require.NoError(t, err)
	assert.Equal(t, pluginsCore.PhaseWaitingForResources, phase.Phase())
	assert.Contains(t, phase.Reason(), "waiting for gang admission")
	assert.Nil(t, phase.Err())
}

func TestGetTaskPhase_SuspendedCondition_NotStarted_WaitingForResources(t *testing.T) {
	js := makeJobSet(jobsetv1alpha2.JobSetSuspended, metav1.ConditionTrue, true)
	js.Status.Conditions[0].Message = "jobset is suspended"
	js.Status.ReplicatedJobsStatus = []jobsetv1alpha2.ReplicatedJobStatus{
		{Name: workersReplicatedJobName, Suspended: 1},
	}
	pCtx := dummyPluginCtx(twoNodeSpec(), emptyK8sReader())

	phase, err := clusteredResourceHandler{}.GetTaskPhase(context.Background(), pCtx, js)
	require.NoError(t, err)
	assert.Equal(t, pluginsCore.PhaseWaitingForResources, phase.Phase())
	assert.Contains(t, phase.Reason(), "jobset is suspended", "the gate's condition message is surfaced")
}

func TestGetTaskPhase_SuspendedCondition_SpecNotSuspended_StillHeld(t *testing.T) {
	// Condition alone (spec already flipped back) still counts as held until the
	// controller clears it: never Running for a JobSet without pods.
	js := makeJobSet(jobsetv1alpha2.JobSetSuspended, metav1.ConditionTrue, false)
	pCtx := dummyPluginCtx(twoNodeSpec(), emptyK8sReader())

	phase, err := clusteredResourceHandler{}.GetTaskPhase(context.Background(), pCtx, js)
	require.NoError(t, err)
	assert.Equal(t, pluginsCore.PhaseWaitingForResources, phase.Phase())
}

func TestGetTaskPhase_Suspended_PriorInitializing_HoldsInitializing(t *testing.T) {
	// A partial start that the gate released: the plugin had reported Initializing,
	// so it keeps reporting Initializing (a backward phase would be dropped).
	js := makeJobSet(jobsetv1alpha2.JobSetSuspended, metav1.ConditionTrue, true)
	pCtx := dummyPluginCtxWithState(twoNodeSpec(), emptyK8sReader(),
		plugink8s.PluginState{
			Phase:        pluginsCore.PhaseInitializing,
			PhaseVersion: 2,
			Reason:       "waiting for all 2 workers to be ready (pods scheduling / DNS resolving)",
		}, nil)

	phase, err := clusteredResourceHandler{}.GetTaskPhase(context.Background(), pCtx, js)
	require.NoError(t, err)
	assert.Equal(t, pluginsCore.PhaseInitializing, phase.Phase())
	assert.Contains(t, phase.Reason(), "released by admission gate")
	assert.Equal(t, uint32(3), phase.Version(), "reason changed within the same phase, so the version is bumped")
}

func TestGetTaskPhase_Suspended_PriorWaitingForResources_StaysWaiting(t *testing.T) {
	js := makeJobSet(jobsetv1alpha2.JobSetSuspended, metav1.ConditionTrue, true)
	pCtx := dummyPluginCtxWithState(twoNodeSpec(), emptyK8sReader(),
		plugink8s.PluginState{
			Phase:        pluginsCore.PhaseWaitingForResources,
			PhaseVersion: 0,
			Reason:       "waiting for gang admission: test message",
		}, nil)

	phase, err := clusteredResourceHandler{}.GetTaskPhase(context.Background(), pCtx, js)
	require.NoError(t, err)
	assert.Equal(t, pluginsCore.PhaseWaitingForResources, phase.Phase())
	assert.Equal(t, uint32(0), phase.Version(), "same phase and reason: no spurious version bump")
}

func TestGetTaskPhase_Suspended_AfterRunning_GangEvictedSystemRetry(t *testing.T) {
	// Sticky Running in plugin state means the gang was fully up in this attempt.
	// Kueue re-suspended it (preemption, pods-ready recovery, deactivation): the
	// attempt is over and must be charged as a GangEvicted system retry, never
	// reported as Running or Queued.
	js := makeJobSet(jobsetv1alpha2.JobSetSuspended, metav1.ConditionTrue, true)
	js.Status.Conditions[0].Message = "jobset is suspended"
	pCtx := dummyPluginCtxWithState(twoNodeSpec(), emptyK8sReader(),
		plugink8s.PluginState{Phase: pluginsCore.PhaseRunning, PhaseVersion: 1}, nil)

	phase, err := clusteredResourceHandler{}.GetTaskPhase(context.Background(), pCtx, js)
	require.NoError(t, err)
	assert.Equal(t, pluginsCore.PhaseRetryableFailure, phase.Phase())
	require.NotNil(t, phase.Err())
	assert.Equal(t, gang.CodeGangEvicted, phase.Err().GetCode())
	assert.Equal(t, core.ExecutionError_SYSTEM, phase.Err().GetKind())
	assert.True(t, gang.IsEviction(phase.Err()))
	assert.True(t, phase.CleanupOnFailure())
	assert.Contains(t, phase.Err().GetMessage(), "gang evicted by kueue")
	assert.Contains(t, phase.Err().GetMessage(), "jobset is suspended")
}

func TestGetTaskPhase_Suspended_ReadyStatus_GangEvicted(t *testing.T) {
	// Status alone can also prove the gang was up (Ready counts child Jobs).
	js := makeJobSet("", "", true)
	js.Status.ReplicatedJobsStatus = []jobsetv1alpha2.ReplicatedJobStatus{
		{Name: workersReplicatedJobName, Ready: 1},
	}
	pCtx := dummyPluginCtx(twoNodeSpec(), emptyK8sReader())

	phase, err := clusteredResourceHandler{}.GetTaskPhase(context.Background(), pCtx, js)
	require.NoError(t, err)
	assert.Equal(t, pluginsCore.PhaseRetryableFailure, phase.Phase())
	require.NotNil(t, phase.Err())
	assert.Equal(t, gang.CodeGangEvicted, phase.Err().GetCode())
	assert.Equal(t, core.ExecutionError_SYSTEM, phase.Err().GetKind())
}

func TestGetTaskPhase_Suspended_UserRetryPolicy(t *testing.T) {
	prev := evictionPolicy
	evictionPolicy = gang.Policy{AsSystemRetry: false}
	t.Cleanup(func() { evictionPolicy = prev })

	js := makeJobSet(jobsetv1alpha2.JobSetSuspended, metav1.ConditionTrue, true)
	pCtx := dummyPluginCtxWithState(twoNodeSpec(), emptyK8sReader(),
		plugink8s.PluginState{Phase: pluginsCore.PhaseRunning, PhaseVersion: 1}, nil)

	phase, err := clusteredResourceHandler{}.GetTaskPhase(context.Background(), pCtx, js)
	require.NoError(t, err)
	assert.Equal(t, pluginsCore.PhaseRetryableFailure, phase.Phase())
	require.NotNil(t, phase.Err())
	assert.Equal(t, gang.CodeGangEvicted, phase.Err().GetCode())
	assert.Equal(t, core.ExecutionError_USER, phase.Err().GetKind())
	assert.False(t, gang.IsEviction(phase.Err()), "user-kind evictions are charged to user retries")
}

func TestGetTaskPhase_Suspended_TerminalConditionWins(t *testing.T) {
	// A completed JobSet that (oddly) still carries spec.suspend is Success.
	js := makeJobSet(jobsetv1alpha2.JobSetCompleted, metav1.ConditionTrue, true)
	pCtx := dummyPluginCtx(twoNodeSpec(), emptyK8sReader())

	phase, err := clusteredResourceHandler{}.GetTaskPhase(context.Background(), pCtx, js)
	require.NoError(t, err)
	assert.Equal(t, pluginsCore.PhaseSuccess, phase.Phase())
}

func TestGetTaskPhase_Resumed_NoPods_WaitingForResources(t *testing.T) {
	// After the gate admits the JobSet, the controller writes Suspended=False (reason
	// ResumeJobs). No true condition and no worker pod on a node yet: still waiting.
	js := makeJobSet("", "", false)
	js.Status.Conditions = []metav1.Condition{
		{
			Type:               string(jobsetv1alpha2.JobSetSuspended),
			Status:             metav1.ConditionFalse,
			Reason:             "ResumeJobs",
			LastTransitionTime: metav1.NewTime(time.Now()),
		},
	}
	js.Status.ReplicatedJobsStatus = []jobsetv1alpha2.ReplicatedJobStatus{
		{Name: workersReplicatedJobName, Active: 1},
	}
	pCtx := dummyPluginCtx(twoNodeSpec(), emptyK8sReader())

	phase, err := clusteredResourceHandler{}.GetTaskPhase(context.Background(), pCtx, js)
	require.NoError(t, err)
	assert.Equal(t, pluginsCore.PhaseWaitingForResources, phase.Phase())
	assert.Equal(t, "0 of 2 workers scheduled", phase.Reason())
}

func TestGetTaskPhase_ActiveNotReady_Initializing(t *testing.T) {
	// Active counts child Jobs with any pod: both pods have nodes, one is Ready.
	js := makeJobSet("", "", false)
	js.Status.ReplicatedJobsStatus = []jobsetv1alpha2.ReplicatedJobStatus{
		{Name: workersReplicatedJobName, Active: 1},
	}
	pCtx := dummyPluginCtx(twoNodeSpec(), workerPodsReader(workerPodReady, workerPodScheduled))

	phase, err := clusteredResourceHandler{}.GetTaskPhase(context.Background(), pCtx, js)
	require.NoError(t, err)
	assert.Equal(t, pluginsCore.PhaseInitializing, phase.Phase())
	assert.Equal(t, "1 of 2 workers ready", phase.Reason())
}

func TestGetTaskPhase_WorkerUnscheduled_WaitingForResources(t *testing.T) {
	// Admitted, but one worker has no node: Initializing would claim every pod has one.
	js := makeJobSet("", "", false)
	pCtx := dummyPluginCtx(twoNodeSpec(), workerPodsReader(workerPodReady, workerPodUnscheduled))

	phase, err := clusteredResourceHandler{}.GetTaskPhase(context.Background(), pCtx, js)
	require.NoError(t, err)
	assert.Equal(t, pluginsCore.PhaseWaitingForResources, phase.Phase())
	assert.Equal(t, "1 of 2 workers scheduled (Unschedulable: 0/1 nodes are available: 1 Insufficient memory.)",
		phase.Reason(), "the scheduler's reason for the first unscheduled worker is surfaced")
}

func TestGetTaskPhase_WorkerUnscheduled_LongSchedulerMessageIsCapped(t *testing.T) {
	js := makeJobSet("", "", false)
	pods := workerPods(workerPodReady, workerPodUnscheduled)
	pods[1].Status.Conditions[0].Message = strings.Repeat("0/1 nodes are available: 1 Insufficient memory. ", 20)
	reader := fake.NewClientBuilder().WithScheme(k8sscheme.Scheme).WithObjects(pods[0], pods[1]).Build()
	pCtx := dummyPluginCtx(twoNodeSpec(), reader)

	phase, err := clusteredResourceHandler{}.GetTaskPhase(context.Background(), pCtx, js)
	require.NoError(t, err)
	assert.Equal(t, pluginsCore.PhaseWaitingForResources, phase.Phase())
	assert.True(t, strings.HasSuffix(phase.Reason(), "…)"), "reason %q is not capped", phase.Reason())
	maxLen := len("1 of 2 workers scheduled (Unschedulable: )") + maxSchedulingDetailLen + len("…")
	assert.LessOrEqual(t, len(phase.Reason()), maxLen)
}

func TestGetTaskPhase_TerminatingWorkerNotCounted(t *testing.T) {
	// A pod being deleted (for example from a released gang) is not part of the gang.
	js := makeJobSet("", "", false)
	pods := workerPods(workerPodScheduled, workerPodScheduled)
	now := metav1.Now()
	pods[1].DeletionTimestamp = &now
	pods[1].Finalizers = []string{"test/keep"}
	reader := fake.NewClientBuilder().WithScheme(k8sscheme.Scheme).WithObjects(pods[0], pods[1]).Build()
	pCtx := dummyPluginCtx(twoNodeSpec(), reader)

	phase, err := clusteredResourceHandler{}.GetTaskPhase(context.Background(), pCtx, js)
	require.NoError(t, err)
	assert.Equal(t, pluginsCore.PhaseWaitingForResources, phase.Phase())
	assert.Equal(t, "1 of 2 workers scheduled", phase.Reason())
}

func TestGetTaskPhase_PriorInitializing_Unscheduled_StaysInitializing(t *testing.T) {
	// A re-admitted gang scheduling again after it had reached Initializing: a lower
	// phase would be dropped downstream, so the phase holds and the reason updates.
	js := makeJobSet("", "", false)
	pCtx := dummyPluginCtxWithState(twoNodeSpec(), workerPodsReader(workerPodUnscheduled, workerPodUnscheduled),
		plugink8s.PluginState{Phase: pluginsCore.PhaseInitializing, PhaseVersion: 3, Reason: "1 of 2 workers ready"}, nil)

	phase, err := clusteredResourceHandler{}.GetTaskPhase(context.Background(), pCtx, js)
	require.NoError(t, err)
	assert.Equal(t, pluginsCore.PhaseInitializing, phase.Phase())
	assert.Contains(t, phase.Reason(), "0 of 2 workers scheduled")
	assert.Equal(t, uint32(4), phase.Version(), "reason changed within the same phase, so the version is bumped")
}

func TestGetTaskPhase_NonRank0PendingFailure_FastFails(t *testing.T) {
	// Fatal pending problems are classified on every worker, not only rank 0.
	js := makeJobSet("", "", false)
	pods := workerPods(workerPodScheduled, workerPodScheduled)
	pods[1].Status = imagePullBackOffStatus()
	reader := fake.NewClientBuilder().WithScheme(k8sscheme.Scheme).WithObjects(pods[0], pods[1]).Build()
	pCtx := dummyPluginCtx(twoNodeSpec(), reader)

	phase, err := clusteredResourceHandler{}.GetTaskPhase(context.Background(), pCtx, js)
	require.NoError(t, err)
	assert.True(t, phase.Phase().IsFailure(), "got %s", phase.Phase())
}

func TestGetTaskPhase_PendingImagePull_NotStarted_FastFails(t *testing.T) {
	// Pending-pod diagnostics still surface while the gang is forming, even though a
	// pending-only JobSet no longer counts as started.
	js := makeJobSet("", "", false)
	js.Status.ReplicatedJobsStatus = []jobsetv1alpha2.ReplicatedJobStatus{
		{Name: workersReplicatedJobName, Active: 1},
	}
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      rank0PodName(testJobName) + "-abc12",
			Namespace: testNS,
			Labels:    workerPodLabels(),
		},
		Status: imagePullBackOffStatus(),
	}
	fakeClient := fake.NewClientBuilder().WithScheme(k8sscheme.Scheme).WithObjects(pod).Build()
	pCtx := dummyPluginCtx(twoNodeSpec(), fakeClient)

	phase, err := clusteredResourceHandler{}.GetTaskPhase(context.Background(), pCtx, js)
	require.NoError(t, err)
	assert.True(t, phase.Phase().IsFailure(), "got %s", phase.Phase())
}

// imagePullBackOffStatus is a pending pod that has failed to pull its image for long
// enough that DemystifyPending reports a failure.
func imagePullBackOffStatus() corev1.PodStatus {
	return corev1.PodStatus{
		Phase: corev1.PodPending,
		Conditions: []corev1.PodCondition{
			{
				Type:               corev1.PodReady,
				Status:             corev1.ConditionFalse,
				Reason:             "ContainersNotReady",
				LastTransitionTime: metav1.NewTime(time.Now().Add(-24 * time.Hour)),
			},
		},
		ContainerStatuses: []corev1.ContainerStatus{
			{
				Name:  primaryContainerName,
				Ready: false,
				State: corev1.ContainerState{
					Waiting: &corev1.ContainerStateWaiting{Reason: "ImagePullBackOff", Message: "Back-off pulling image"},
				},
			},
		},
	}
}

// testNodeName is the node the scheduled worker pods in these fixtures run on.
const testNodeName = "node-a"

// workerPodState is how far a worker pod has got.
type workerPodState int

const (
	workerPodUnscheduled workerPodState = iota // pending, no node
	workerPodScheduled                         // on a node, not Ready
	workerPodReady                             // on a node and Ready
)

// workerPods builds testJobName's worker pods, in index order (index 0 is rank 0).
func workerPods(states ...workerPodState) []*corev1.Pod {
	pods := make([]*corev1.Pod, 0, len(states))
	for i, state := range states {
		pod := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("%s-%s-0-%d-abc%02d", testJobName, workersReplicatedJobName, i, i),
				Namespace: testNS,
				Labels:    workerPodLabels(),
			},
			Status: corev1.PodStatus{Phase: corev1.PodPending},
		}
		switch state {
		case workerPodUnscheduled:
			pod.Status.Conditions = []corev1.PodCondition{{
				Type:    corev1.PodScheduled,
				Status:  corev1.ConditionFalse,
				Reason:  corev1.PodReasonUnschedulable,
				Message: "0/1 nodes are available: 1 Insufficient memory.",
			}}
		case workerPodScheduled:
			pod.Spec.NodeName = testNodeName
		case workerPodReady:
			pod.Spec.NodeName = testNodeName
			pod.Status.Phase = corev1.PodRunning
			pod.Status.Conditions = []corev1.PodCondition{{Type: corev1.PodReady, Status: corev1.ConditionTrue}}
		}
		pods = append(pods, pod)
	}
	return pods
}

// workerPodsReader serves workerPods(states...) to the plugin.
func workerPodsReader(states ...workerPodState) client.Reader {
	builder := fake.NewClientBuilder().WithScheme(k8sscheme.Scheme)
	for _, pod := range workerPods(states...) {
		builder = builder.WithObjects(pod)
	}
	return builder.Build()
}

// workers returns a JobSet status whose workers ReplicatedJob has the given counts.
func workers(status jobsetv1alpha2.ReplicatedJobStatus) jobsetv1alpha2.JobSetStatus {
	status.Name = workersReplicatedJobName
	return jobsetv1alpha2.JobSetStatus{ReplicatedJobsStatus: []jobsetv1alpha2.ReplicatedJobStatus{status}}
}

func TestHasJobSetStarted_Table(t *testing.T) {
	tests := []struct {
		name   string
		status jobsetv1alpha2.JobSetStatus
		state  plugink8s.PluginState
		want   bool
	}{
		{name: "empty", want: false},
		{name: "active only is not started", status: workers(jobsetv1alpha2.ReplicatedJobStatus{Active: 1}), want: false},
		{name: "failed only is not started", status: workers(jobsetv1alpha2.ReplicatedJobStatus{Failed: 1}), want: false},
		{name: "suspended only is not started", status: workers(jobsetv1alpha2.ReplicatedJobStatus{Suspended: 1})},
		{name: "ready is started", status: workers(jobsetv1alpha2.ReplicatedJobStatus{Ready: 1}), want: true},
		{name: "succeeded is started", status: workers(jobsetv1alpha2.ReplicatedJobStatus{Succeeded: 1}), want: true},
		{name: "restarts is started", status: jobsetv1alpha2.JobSetStatus{Restarts: 1}, want: true},
		{name: "sticky running state is started", state: plugink8s.PluginState{Phase: pluginsCore.PhaseRunning}, want: true},
		{
			name:  "initializing state is not started",
			state: plugink8s.PluginState{Phase: pluginsCore.PhaseInitializing},
			want:  false,
		},
		{
			name: "other replicated job ready does not count",
			status: jobsetv1alpha2.JobSetStatus{
				ReplicatedJobsStatus: []jobsetv1alpha2.ReplicatedJobStatus{{Name: "other", Ready: 1}},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			js := makeJobSet("", "", false)
			js.Status = tt.status
			assert.Equal(t, tt.want, hasJobSetStarted(js, tt.state))
		})
	}
}
