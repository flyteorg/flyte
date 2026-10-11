package clustered

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	k8stypes "k8s.io/apimachinery/pkg/types"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	jobsetv1alpha2 "sigs.k8s.io/jobset/api/jobset/v1alpha2"

	pluginsCore "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/core"
	coreMocks "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/core/mocks"
	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/gang"
	plugink8s "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/k8s"
	k8smocks "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/k8s/mocks"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/core"
)

const (
	testJobSetUID    = k8stypes.UID("jobset-uid")
	testQueue        = "team-a"
	preemptedMessage = "Preempted to accommodate a workload (UID: 1234) due to prioritization in the ClusterQueue"
)

// inadmissible is the QuotaReserved condition Kueue writes when it cannot admit a Workload.
func inadmissible(message string) map[string]interface{} {
	return workloadCond(workloadConditionQuotaReserved, "False", quotaReservedInadmissible, message)
}

func workloadCond(conditionType, status, reason, message string) map[string]interface{} {
	return map[string]interface{}{"type": conditionType, "status": status, "reason": reason, "message": message}
}

// workload is a Kueue Workload for the JobSet with the given UID, as Kueue writes it.
func workload(jobSetUID k8stypes.UID, conditions ...map[string]interface{}) *unstructured.Unstructured {
	items := make([]interface{}, 0, len(conditions))
	for _, c := range conditions {
		items = append(items, c)
	}
	return &unstructured.Unstructured{Object: map[string]interface{}{
		"apiVersion": kueueAPIVersion,
		"kind":       "Workload",
		"metadata": map[string]interface{}{
			"name":      "jobset-" + testJobName + "-abcde",
			"namespace": testNS,
			"labels":    map[string]interface{}{kueueJobUIDLabel: string(jobSetUID)},
		},
		"status": map[string]interface{}{"conditions": items},
	}}
}

// readerWith serves the given Kueue objects. It uses its own scheme because the fake client
// registers unknown unstructured kinds into the scheme it is given.
func readerWith(t *testing.T, objects ...client.Object) client.Reader {
	t.Helper()
	scheme := runtime.NewScheme()
	require.NoError(t, clientgoscheme.AddToScheme(scheme))
	return fake.NewClientBuilder().WithScheme(scheme).WithObjects(objects...).Build()
}

// kueuePluginCtx is a plugin context for an attempt of a task with maxAttempts attempts.
func kueuePluginCtx(
	reader client.Reader, state plugink8s.PluginState, retryAttempt, maxAttempts uint32,
) *k8smocks.PluginContext {
	pCtx := &k8smocks.PluginContext{}
	taskReader := &coreMocks.TaskReader{}
	taskReader.EXPECT().Read(mock.Anything).Return(twoNodeSpec(), nil)
	pCtx.EXPECT().TaskReader().Return(taskReader)
	pCtx.EXPECT().K8sReader().Return(reader)

	tID := &coreMocks.TaskExecutionID{}
	tID.EXPECT().GetID().Return(&core.TaskExecutionIdentifier{
		NodeExecutionId: &core.NodeExecutionIdentifier{
			ExecutionId: &core.WorkflowExecutionIdentifier{Name: "exec"},
		},
		RetryAttempt: retryAttempt,
	})
	tID.EXPECT().GetGeneratedName().Return(testJobName)
	tID.EXPECT().GetUniqueNodeID().Return("node-id").Maybe()
	meta := &coreMocks.TaskExecutionMetadata{}
	meta.EXPECT().GetTaskExecutionID().Return(tID)
	meta.EXPECT().GetMaxAttempts().Return(maxAttempts).Maybe()
	pCtx.EXPECT().TaskExecutionMetadata().Return(meta)

	stateReader := &coreMocks.PluginStateReader{}
	stateReader.EXPECT().Get(mock.Anything).RunAndReturn(func(v interface{}) (uint8, error) {
		if s, ok := v.(*plugink8s.PluginState); ok {
			*s = state
		}
		return 0, nil
	})
	pCtx.EXPECT().PluginStateReader().Return(stateReader)
	return pCtx
}

// suspendedJobSet is a JobSet Kueue holds or took back, in queue testQueue.
func suspendedJobSet() *jobsetv1alpha2.JobSet {
	js := makeJobSet(jobsetv1alpha2.JobSetSuspended, metav1.ConditionTrue, true)
	js.UID = testJobSetUID
	js.Labels = map[string]string{kueueQueueNameLabel: testQueue}
	return js
}

var startedState = plugink8s.PluginState{Phase: pluginsCore.PhaseRunning, PhaseVersion: 1}

func TestGetTaskPhase_Evicted_ClassifiedByWorkloadReason(t *testing.T) {
	tests := []struct {
		name         string
		workload     *unstructured.Unstructured
		retryAttempt uint32
		maxAttempts  uint32
		wantKind     core.ExecutionError_ErrorKind
		wantInMsg    []string
	}{
		{
			name: "preemption is the user's and uses a retry",
			workload: workload(testJobSetUID,
				workloadCond(workloadConditionEvicted, "True", evictedByPreemption, preemptedMessage),
				workloadCond(workloadConditionPreempted, "True", "InClusterQueue", preemptedMessage)),
			maxAttempts: 3,
			wantKind:    core.ExecutionError_USER,
			wantInMsg: []string{"gang evicted by kueue (Preempted, InClusterQueue)", preemptedMessage,
				"uses one of the task's retries (1 left after it)"},
		},
		{
			name: "preemption on the last attempt says the task fails",
			workload: workload(testJobSetUID,
				workloadCond(workloadConditionEvicted, "True", evictedByPreemption, preemptedMessage)),
			retryAttempt: 2,
			maxAttempts:  3,
			wantKind:     core.ExecutionError_USER,
			wantInMsg:    []string{"none are left, so the task fails"},
		},
		{
			name: "the user's maximum execution time is the user's",
			workload: workload(testJobSetUID,
				workloadCond(workloadConditionEvicted, "True", evictedByMaximumExecutionTime,
					"exceeding the maximum execution time")),
			maxAttempts: 2,
			wantKind:    core.ExecutionError_USER,
			wantInMsg:   []string{"(" + evictedByMaximumExecutionTime + ")", "uses the task's last retry"},
		},
		{
			name: "an operator deactivating it is a system retry",
			workload: workload(testJobSetUID,
				workloadCond(workloadConditionEvicted, "True", "Deactivated", "The workload is deactivated")),
			maxAttempts: 3,
			wantKind:    core.ExecutionError_SYSTEM,
			wantInMsg:   []string{"(Deactivated): The workload is deactivated", "does not count against the task's retries"},
		},
		{
			name: "a stopped queue is a system retry",
			workload: workload(testJobSetUID,
				workloadCond(workloadConditionEvicted, "True", "ClusterQueueStopped", "The ClusterQueue is stopped")),
			wantKind:  core.ExecutionError_SYSTEM,
			wantInMsg: []string{"(ClusterQueueStopped)"},
		},
		{
			name: "node failures are a system retry",
			workload: workload(testJobSetUID,
				workloadCond(workloadConditionEvicted, "True", "NodeFailures", "node lost")),
			wantKind: core.ExecutionError_SYSTEM,
		},
		{
			name:      "no Workload for this JobSet is a system retry",
			workload:  workload("another-jobset", workloadCond(workloadConditionEvicted, "True", evictedByPreemption, "")),
			wantKind:  core.ExecutionError_SYSTEM,
			wantInMsg: []string{"gang evicted by kueue; this does not count against the task's retries"},
		},
		{
			name: "an Evicted condition that is not true is a system retry",
			workload: workload(testJobSetUID,
				workloadCond(workloadConditionEvicted, "False", evictedByPreemption, "")),
			wantKind: core.ExecutionError_SYSTEM,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pCtx := kueuePluginCtx(readerWith(t, tt.workload), startedState, tt.retryAttempt, tt.maxAttempts)

			phase, err := clusteredResourceHandler{}.GetTaskPhase(context.Background(), pCtx, suspendedJobSet())
			require.NoError(t, err)

			assert.Equal(t, pluginsCore.PhaseRetryableFailure, phase.Phase())
			require.NotNil(t, phase.Err())
			assert.Equal(t, gang.CodeGangEvicted, phase.Err().GetCode())
			assert.Equal(t, tt.wantKind, phase.Err().GetKind())
			assert.True(t, phase.CleanupOnFailure(), "the evicted JobSet is always cleaned up")
			for _, want := range tt.wantInMsg {
				assert.Contains(t, phase.Err().GetMessage(), want)
			}
		})
	}
}

func TestGetTaskPhase_Held_MissingLocalQueue_FailsAsUserError(t *testing.T) {
	wl := workload(testJobSetUID,
		inadmissible("LocalQueue "+testQueue+" doesn't exist"))
	pCtx := kueuePluginCtx(readerWith(t, wl), plugink8s.PluginState{}, 0, 3)

	phase, err := clusteredResourceHandler{}.GetTaskPhase(context.Background(), pCtx, suspendedJobSet())
	require.NoError(t, err)

	assert.Equal(t, pluginsCore.PhasePermanentFailure, phase.Phase(), "not retried: the queue will not appear")
	require.NotNil(t, phase.Err())
	assert.Equal(t, codeQueueNotFound, phase.Err().GetCode())
	assert.Equal(t, core.ExecutionError_USER, phase.Err().GetKind())
	assert.Contains(t, phase.Err().GetMessage(), `"`+testQueue+`"`)
	assert.True(t, phase.CleanupOnFailure())
}

func TestGetTaskPhase_Held_KeepsWaiting(t *testing.T) {
	tests := []struct {
		name     string
		workload *unstructured.Unstructured
		noLabel  bool
	}{
		{
			name: "an inactive queue is the platform's, so the gang keeps waiting",
			workload: workload(testJobSetUID,
				inadmissible("LocalQueue "+testQueue+" is inactive")),
		},
		{
			name: "a missing queue other than the one the JobSet names is not this JobSet's",
			workload: workload(testJobSetUID,
				inadmissible("LocalQueue other doesn't exist")),
		},
		{
			name: "waiting for quota",
			workload: workload(testJobSetUID,
				workloadCond(workloadConditionQuotaReserved, "False", "Pending", "couldn't assign flavors to pod set workers")),
		},
		{
			name:     "no Workload yet",
			workload: workload("another-jobset"),
		},
		{
			name:    "no queue label",
			noLabel: true,
			workload: workload(testJobSetUID,
				inadmissible("LocalQueue "+testQueue+" doesn't exist")),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			js := suspendedJobSet()
			if tt.noLabel {
				js.Labels = nil
			}
			pCtx := kueuePluginCtx(readerWith(t, tt.workload), plugink8s.PluginState{}, 0, 3)

			phase, err := clusteredResourceHandler{}.GetTaskPhase(context.Background(), pCtx, js)
			require.NoError(t, err)
			assert.Equal(t, pluginsCore.PhaseWaitingForResources, phase.Phase())
			assert.Contains(t, phase.Reason(), "waiting for gang admission")
		})
	}
}

func TestWorkloadForJobSet_NewestWins(t *testing.T) {
	older := workload(testJobSetUID)
	older.SetName("older")
	older.SetCreationTimestamp(metav1.Unix(100, 0))
	newer := workload(testJobSetUID)
	newer.SetName("newer")
	newer.SetCreationTimestamp(metav1.Unix(200, 0))

	wl, err := workloadForJobSet(context.Background(), readerWith(t, older, newer), suspendedJobSet())
	require.NoError(t, err)
	require.NotNil(t, wl)
	assert.Equal(t, "newer", wl.GetName())

	wl, err = workloadForJobSet(context.Background(), nil, suspendedJobSet())
	require.NoError(t, err)
	assert.Nil(t, wl, "no reader means no Workload")
}

func TestClassifyEviction_Unreadable(t *testing.T) {
	assert.Equal(t, gang.Eviction{Source: evictionSource}, classifyEviction(nil))
	malformed := workload(testJobSetUID)
	malformed.Object["status"] = "not a map"
	assert.False(t, classifyEviction(malformed).UserCaused)
}
