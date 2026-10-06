package clustered

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sscheme "k8s.io/client-go/kubernetes/scheme"
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

func TestGetTaskPhase_Resumed_NoPods_Initializing(t *testing.T) {
	// After the gate releases the JobSet, the controller writes Suspended=False
	// (reason ResumeJobs). No true condition and no Ready workers: Initializing.
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
	assert.Equal(t, pluginsCore.PhaseInitializing, phase.Phase())
}

func TestGetTaskPhase_ActiveNotReady_Initializing(t *testing.T) {
	js := makeJobSet("", "", false)
	js.Status.ReplicatedJobsStatus = []jobsetv1alpha2.ReplicatedJobStatus{
		{Name: workersReplicatedJobName, Active: 1},
	}
	pCtx := dummyPluginCtx(twoNodeSpec(), emptyK8sReader())

	phase, err := clusteredResourceHandler{}.GetTaskPhase(context.Background(), pCtx, js)
	require.NoError(t, err)
	assert.Equal(t, pluginsCore.PhaseInitializing, phase.Phase())
}

func TestGetTaskPhase_PendingImagePull_NotStarted_FastFails(t *testing.T) {
	// Pending-pod diagnostics still surface while the gang is forming, even though a
	// pending-only JobSet no longer counts as started.
	js := makeJobSet("", "", false)
	js.Status.ReplicatedJobsStatus = []jobsetv1alpha2.ReplicatedJobStatus{
		{Name: workersReplicatedJobName, Active: 1},
	}
	oldTransition := metav1.NewTime(time.Now().Add(-24 * time.Hour))
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: rank0PodName(testJobName) + "-abc12", Namespace: testNS},
		Status: corev1.PodStatus{
			Phase: corev1.PodPending,
			Conditions: []corev1.PodCondition{
				{
					Type:               corev1.PodReady,
					Status:             corev1.ConditionFalse,
					Reason:             "ContainersNotReady",
					LastTransitionTime: oldTransition,
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
		},
	}
	fakeClient := fake.NewClientBuilder().WithScheme(k8sscheme.Scheme).WithObjects(pod).Build()
	pCtx := dummyPluginCtx(twoNodeSpec(), fakeClient)

	phase, err := clusteredResourceHandler{}.GetTaskPhase(context.Background(), pCtx, js)
	require.NoError(t, err)
	assert.True(t, phase.Phase().IsFailure(), "got %s", phase.Phase())
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
			pCtx := dummyPluginCtxWithState(twoNodeSpec(), emptyK8sReader(), tt.state, nil)
			assert.Equal(t, tt.want, hasJobSetStarted(context.Background(), pCtx, js))
		})
	}
}
