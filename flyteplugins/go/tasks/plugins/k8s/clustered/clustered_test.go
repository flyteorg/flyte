package clustered

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/validation"
	k8sscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	jobsetv1alpha2 "sigs.k8s.io/jobset/api/jobset/v1alpha2"
	"sigs.k8s.io/jobset/pkg/util/placement"

	pluginsCore "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/core"
	coreMocks "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/core/mocks"
	pluginIOMocks "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/io/mocks"
	plugink8s "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/k8s"
	k8smocks "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/k8s/mocks"
	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/utils"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/core"
	clusteredpb "github.com/flyteorg/flyte/v2/gen/go/flyteidl2/plugins"
)

const (
	testImage   = "test-image:latest"
	testJobName = "f-abc123"
	testNS      = "my-project-development"
)

// buildTaskTemplate builds a TaskTemplate with the given ClusteredTaskSpec packed into Custom.
func buildTaskTemplate(spec *clusteredpb.ClusteredTaskSpec) *core.TaskTemplate {
	custom, err := utils.MarshalObjToStruct(spec) //nolint:staticcheck
	if err != nil {
		panic(err)
	}
	return &core.TaskTemplate{
		Type:            taskType,
		TaskTypeVersion: 1,
		Target: &core.TaskTemplate_Container{
			Container: &core.Container{
				Image:   testImage,
				Command: []string{"a0"},
				Args:    []string{"a0", "--inputs", "s3://bucket/in"},
			},
		},
		Custom: custom,
	}
}

// dummyTaskCtx builds a minimal task execution context suitable for BuildResource tests.
func dummyTaskCtx(taskTemplate *core.TaskTemplate) *coreMocks.TaskExecutionContext {
	return dummyTaskCtxWithGeneratedName(taskTemplate, testJobName)
}

// dummyTaskCtxWithGeneratedName is dummyTaskCtx with a caller-supplied generated name, used to
// exercise the long composed/nested-name truncation path.
func dummyTaskCtxWithGeneratedName(taskTemplate *core.TaskTemplate, generatedName string) *coreMocks.TaskExecutionContext {
	taskCtx := &coreMocks.TaskExecutionContext{}

	inputReader := &pluginIOMocks.InputReader{}
	inputReader.EXPECT().GetInputPrefixPath().Return("/input/prefix")
	inputReader.EXPECT().GetInputPath().Return("/input")
	inputReader.EXPECT().Get(mock.Anything).Return(&core.LiteralMap{}, nil)
	taskCtx.EXPECT().InputReader().Return(inputReader)

	outputWriter := &pluginIOMocks.OutputWriter{}
	outputWriter.EXPECT().GetOutputPath().Return("/data/outputs.pb")
	outputWriter.EXPECT().GetOutputPrefixPath().Return("/data/")
	outputWriter.EXPECT().GetRawOutputPrefix().Return("")
	outputWriter.EXPECT().GetCheckpointPrefix().Return("/checkpoint")
	outputWriter.EXPECT().GetPreviousCheckpointsPrefix().Return("/prev")
	taskCtx.EXPECT().OutputWriter().Return(outputWriter)

	taskReader := &coreMocks.TaskReader{}
	taskReader.EXPECT().Read(mock.Anything).Return(taskTemplate, nil)
	taskCtx.EXPECT().TaskReader().Return(taskReader)

	tID := &coreMocks.TaskExecutionID{}
	tID.EXPECT().GetID().Return(&core.TaskExecutionIdentifier{
		NodeExecutionId: &core.NodeExecutionIdentifier{
			ExecutionId: &core.WorkflowExecutionIdentifier{
				Name:    "my-exec",
				Project: "my-project",
				Domain:  "development",
			},
		},
	})
	tID.EXPECT().GetGeneratedName().Return(generatedName)
	tID.EXPECT().GetUniqueNodeID().Return("node-id")

	overrides := &coreMocks.TaskOverrides{}
	overrides.EXPECT().GetResources().Return(&corev1.ResourceRequirements{
		Limits: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse("8"),
			corev1.ResourceMemory: resource.MustParse("32Gi"),
		},
	})
	overrides.EXPECT().GetExtendedResources().Return(nil)
	overrides.EXPECT().GetContainerImage().Return("")
	overrides.EXPECT().GetPodTemplate().Return(nil)

	meta := &coreMocks.TaskExecutionMetadata{}
	meta.EXPECT().GetTaskExecutionID().Return(tID)
	meta.EXPECT().GetNamespace().Return(testNS)
	meta.EXPECT().GetAnnotations().Return(map[string]string{"flyte.org/test-annotation": "av"})
	meta.EXPECT().GetLabels().Return(map[string]string{"execution-id": "my-exec", "node-id": "n1"})
	meta.EXPECT().GetOwnerReference().Return(metav1.OwnerReference{Kind: "node", Name: "n1"})
	meta.EXPECT().IsInterruptible().Return(false)
	meta.EXPECT().GetOverrides().Return(overrides)
	meta.EXPECT().GetK8sServiceAccount().Return("")
	meta.EXPECT().GetPlatformResources().Return(&corev1.ResourceRequirements{})
	meta.EXPECT().GetEnvironmentVariables().Return(nil)
	meta.EXPECT().GetConsoleURL().Return("")
	taskCtx.EXPECT().TaskExecutionMetadata().Return(meta)

	return taskCtx
}

// --- BuildResource tests ---

func TestBuildResource_HappyPath(t *testing.T) {
	spec := &clusteredpb.ClusteredTaskSpec{
		Replicas:     4,
		NprocPerNode: 8,
		Runtime: &clusteredpb.Runtime{
			Kind: &clusteredpb.Runtime_Torchrun{
				Torchrun: &clusteredpb.TorchRuntime{
					RdzvBackend: clusteredpb.RdzvBackend_STATIC,
					MaxRestarts: 0,
				},
			},
		},
		FailurePolicy: &clusteredpb.ClusterFailurePolicy{MaxRestarts: 3},
	}
	taskTemplate := buildTaskTemplate(spec)
	taskCtx := dummyTaskCtx(taskTemplate)

	handler := clusteredResourceHandler{}
	obj, err := handler.BuildResource(context.Background(), taskCtx)
	assert.NoError(t, err)
	assert.NotNil(t, obj)

	jobSet, ok := obj.(*jobsetv1alpha2.JobSet)
	assert.True(t, ok, "expected *JobSet")

	assert.Equal(t, testJobName, jobSet.Name)
	assert.Equal(t, testNS, jobSet.Namespace)
	assert.True(t, *jobSet.Spec.Network.EnableDNSHostnames)
	assert.Equal(t, jobsetv1alpha2.OperatorAll, jobSet.Spec.SuccessPolicy.Operator)
	assert.Equal(t, int32(3), jobSet.Spec.FailurePolicy.MaxRestarts)
	assert.Len(t, jobSet.Spec.ReplicatedJobs, 1)
	assert.Equal(t, "workers", jobSet.Spec.ReplicatedJobs[0].Name)
	assert.Equal(t, int32(1), jobSet.Spec.ReplicatedJobs[0].Replicas)

	jobSpec := jobSet.Spec.ReplicatedJobs[0].Template.Spec
	assert.Equal(t, int32(4), *jobSpec.Parallelism)
	assert.Equal(t, int32(4), *jobSpec.Completions)
	assert.Equal(t, batchv1.IndexedCompletion, *jobSpec.CompletionMode)
	assert.Equal(t, int32(0), *jobSpec.BackoffLimit)

	// The node-execution labels/annotations must be propagated onto the pod template so
	// JobSet child pods carry execution-id/node-id; otherwise the node-execution-scoped
	// K8sReader.List in getLogContext returns nothing and no logs reach the UI.
	podMeta := jobSpec.Template.ObjectMeta
	assert.Equal(t, "my-exec", podMeta.Labels["execution-id"])
	assert.Equal(t, "n1", podMeta.Labels["node-id"])
	assert.Equal(t, "av", podMeta.Annotations["flyte.org/test-annotation"])
}

func TestBuildResource_PrimaryContainerPreserved(t *testing.T) {
	// The plugin no longer rewrites container.Command — the SDK does that at
	// serde time (design §3.2 / §3.8). Here we assert the plugin passes the
	// TaskTemplate's container through unchanged and stamps the primary
	// container name onto the JobSet via annotation for status-time recovery.
	spec := &clusteredpb.ClusteredTaskSpec{
		Replicas:     2,
		NprocPerNode: 1,
		Runtime: &clusteredpb.Runtime{
			Kind: &clusteredpb.Runtime_Torchrun{
				Torchrun: &clusteredpb.TorchRuntime{},
			},
		},
	}
	taskTemplate := buildTaskTemplate(spec)
	taskCtx := dummyTaskCtx(taskTemplate)

	handler := clusteredResourceHandler{}
	obj, err := handler.BuildResource(context.Background(), taskCtx)
	assert.NoError(t, err)

	jobSet := obj.(*jobsetv1alpha2.JobSet)
	podSpec := jobSet.Spec.ReplicatedJobs[0].Template.Spec.Template.Spec

	assert.NotEmpty(t, podSpec.Containers)
	primary := &podSpec.Containers[0]

	// Command + args from the TaskTemplate must reach the pod untouched.
	assert.Equal(t, []string{"a0"}, primary.Command)
	assert.Equal(t, []string{"a0", "--inputs", "s3://bucket/in"}, primary.Args)

	// Primary container name must be retrievable from the JobSet at status time.
	assert.Equal(t, primary.Name, jobSet.Annotations[primaryContainerAnnotation])
}

// --- injectTorchRunEnv tests ---

func TestInjectTorchRunEnv_Static(t *testing.T) {
	spec := &clusteredpb.ClusteredTaskSpec{
		Replicas:     4,
		NprocPerNode: 8,
		Runtime: &clusteredpb.Runtime{
			Kind: &clusteredpb.Runtime_Torchrun{
				Torchrun: &clusteredpb.TorchRuntime{RdzvBackend: clusteredpb.RdzvBackend_STATIC},
			},
		},
	}
	container := &corev1.Container{}
	injectTorchRunEnv(container, spec)

	envMap := make(map[string]string)
	for _, e := range container.Env {
		if e.Value != "" {
			envMap[e.Name] = e.Value
		}
	}
	assert.Equal(t, "4", envMap["NNODES"])
	assert.Equal(t, "8", envMap["NPROC_PER_NODE"])
	assert.Equal(t, "29500", envMap["MASTER_PORT"])
	assert.Equal(t, "static", envMap["RDZV_BACKEND"])
	// No failure policy set → budget defaults to 0 (every failure is terminal).
	assert.Equal(t, "0", envMap["JOBSET_MAX_RESTARTS"])

	// Downward API env vars should be present.
	names := make(map[string]bool)
	for _, e := range container.Env {
		names[e.Name] = true
	}
	assert.True(t, names["JOBSET_NAME"])
	assert.True(t, names["JOBSET_RESTART_ATTEMPT"])
	assert.True(t, names["JOBSET_MAX_RESTARTS"])
	assert.True(t, names["POD_NAMESPACE"])
}

func TestInjectTorchRunEnv_MaxRestarts(t *testing.T) {
	spec := &clusteredpb.ClusteredTaskSpec{
		Replicas:      2,
		NprocPerNode:  4,
		FailurePolicy: &clusteredpb.ClusterFailurePolicy{MaxRestarts: 3},
	}
	container := &corev1.Container{}
	injectTorchRunEnv(container, spec)

	for _, e := range container.Env {
		if e.Name == "JOBSET_MAX_RESTARTS" {
			assert.Equal(t, "3", e.Value)
			return
		}
	}
	t.Fatal("JOBSET_MAX_RESTARTS not found")
}

func TestInjectTorchRunEnv_C10D(t *testing.T) {
	spec := &clusteredpb.ClusteredTaskSpec{
		Replicas:     2,
		NprocPerNode: 4,
		Runtime: &clusteredpb.Runtime{
			Kind: &clusteredpb.Runtime_Torchrun{
				Torchrun: &clusteredpb.TorchRuntime{RdzvBackend: clusteredpb.RdzvBackend_C10D},
			},
		},
	}
	container := &corev1.Container{}
	injectTorchRunEnv(container, spec)

	for _, e := range container.Env {
		if e.Name == "RDZV_BACKEND" {
			assert.Equal(t, "c10d", e.Value)
			return
		}
	}
	t.Fatal("RDZV_BACKEND not found")
}

// --- buildFailurePolicy tests ---

func TestBuildFailurePolicy_MaxRestarts(t *testing.T) {
	spec := &clusteredpb.ClusteredTaskSpec{
		FailurePolicy: &clusteredpb.ClusterFailurePolicy{MaxRestarts: 3},
	}
	fp, err := buildFailurePolicy(spec)
	assert.NoError(t, err)
	assert.NotNil(t, fp)
	assert.Equal(t, int32(3), fp.MaxRestarts)
}

func TestBuildFailurePolicy_Zero(t *testing.T) {
	spec := &clusteredpb.ClusteredTaskSpec{
		FailurePolicy: &clusteredpb.ClusterFailurePolicy{MaxRestarts: 0},
	}
	fp, err := buildFailurePolicy(spec)
	assert.NoError(t, err)
	assert.Nil(t, fp)
}

func TestBuildFailurePolicy_Nil(t *testing.T) {
	spec := &clusteredpb.ClusteredTaskSpec{}
	fp, err := buildFailurePolicy(spec)
	assert.NoError(t, err)
	assert.Nil(t, fp)
}

func TestBuildFailurePolicy_Negative(t *testing.T) {
	spec := &clusteredpb.ClusteredTaskSpec{
		FailurePolicy: &clusteredpb.ClusterFailurePolicy{MaxRestarts: -1},
	}
	fp, err := buildFailurePolicy(spec)
	assert.Error(t, err)
	assert.Nil(t, fp)
}

// --- GetTaskPhase tests ---

func makeJobSet(condType jobsetv1alpha2.JobSetConditionType, status metav1.ConditionStatus, suspend bool) *jobsetv1alpha2.JobSet {
	js := &jobsetv1alpha2.JobSet{
		ObjectMeta: metav1.ObjectMeta{Name: testJobName, Namespace: testNS},
		Spec: jobsetv1alpha2.JobSetSpec{
			Suspend: &suspend,
			ReplicatedJobs: []jobsetv1alpha2.ReplicatedJob{
				{
					Name:     "workers",
					Replicas: 1,
					Template: batchv1.JobTemplateSpec{
						Spec: batchv1.JobSpec{
							Parallelism: func() *int32 { v := int32(2); return &v }(),
						},
					},
				},
			},
		},
	}
	if condType != "" {
		js.Status.Conditions = []metav1.Condition{
			{
				Type:               string(condType),
				Status:             status,
				LastTransitionTime: metav1.NewTime(time.Now()),
				Reason:             "test",
				Message:            "test message",
			},
		}
	}
	return js
}

// emptyK8sReader returns a fake client with no objects, for tests that don't
// exercise pod inspection (getLogContext just yields an empty pod list -> nil LogContext).
func emptyK8sReader() client.Reader {
	return fake.NewClientBuilder().WithScheme(k8sscheme.Scheme).Build()
}

func dummyPluginCtx(taskTemplate *core.TaskTemplate, k8sReader client.Reader) *k8smocks.PluginContext {
	return dummyPluginCtxWithState(taskTemplate, k8sReader, plugink8s.PluginState{}, nil)
}

func dummyPluginCtxWithState(
	taskTemplate *core.TaskTemplate,
	k8sReader client.Reader,
	pluginState plugink8s.PluginState,
	pluginStateErr error,
) *k8smocks.PluginContext {
	pCtx := &k8smocks.PluginContext{}

	taskReader := &coreMocks.TaskReader{}
	taskReader.EXPECT().Read(mock.Anything).Return(taskTemplate, nil)
	pCtx.EXPECT().TaskReader().Return(taskReader)

	pCtx.EXPECT().K8sReader().Return(k8sReader)

	tID := &coreMocks.TaskExecutionID{}
	tID.EXPECT().GetID().Return(&core.TaskExecutionIdentifier{
		NodeExecutionId: &core.NodeExecutionIdentifier{
			ExecutionId: &core.WorkflowExecutionIdentifier{Name: "exec"},
		},
	})
	tID.EXPECT().GetGeneratedName().Return(testJobName)
	tID.EXPECT().GetUniqueNodeID().Return("node-id").Maybe()

	meta := &coreMocks.TaskExecutionMetadata{}
	meta.EXPECT().GetTaskExecutionID().Return(tID)
	pCtx.EXPECT().TaskExecutionMetadata().Return(meta)

	pluginStateReader := &coreMocks.PluginStateReader{}
	pluginStateReader.EXPECT().Get(mock.Anything).RunAndReturn(func(t interface{}) (uint8, error) {
		if pluginStateErr != nil {
			return 0, pluginStateErr
		}
		if s, ok := t.(*plugink8s.PluginState); ok {
			*s = pluginState
		}
		return 0, nil
	})
	pCtx.EXPECT().PluginStateReader().Return(pluginStateReader)

	return pCtx
}

func TestGetTaskPhase_Initializing(t *testing.T) {
	suspend := false
	js := makeJobSet("", "", suspend)

	spec := &clusteredpb.ClusteredTaskSpec{Replicas: 2, NprocPerNode: 1}
	pCtx := dummyPluginCtx(buildTaskTemplate(spec), emptyK8sReader())

	handler := clusteredResourceHandler{}
	phase, err := handler.GetTaskPhase(context.Background(), pCtx, js)
	assert.NoError(t, err)
	assert.Equal(t, pluginsCore.PhaseInitializing, phase.Phase())
}

func TestGetTaskPhase_Success(t *testing.T) {
	js := makeJobSet(jobsetv1alpha2.JobSetCompleted, metav1.ConditionTrue, false)

	spec := &clusteredpb.ClusteredTaskSpec{Replicas: 2, NprocPerNode: 1}
	pCtx := dummyPluginCtx(buildTaskTemplate(spec), emptyK8sReader())

	handler := clusteredResourceHandler{}
	phase, err := handler.GetTaskPhase(context.Background(), pCtx, js)
	assert.NoError(t, err)
	assert.Equal(t, pluginsCore.PhaseSuccess, phase.Phase())
}

func TestGetTaskPhase_Failure(t *testing.T) {
	js := makeJobSet(jobsetv1alpha2.JobSetFailed, metav1.ConditionTrue, false)

	spec := &clusteredpb.ClusteredTaskSpec{Replicas: 2, NprocPerNode: 1}
	pCtx := dummyPluginCtx(buildTaskTemplate(spec), emptyK8sReader())

	handler := clusteredResourceHandler{}
	phase, err := handler.GetTaskPhase(context.Background(), pCtx, js)
	assert.NoError(t, err)
	assert.Equal(t, pluginsCore.PhaseRetryableFailure, phase.Phase())
}

func TestGetTaskPhase_Running(t *testing.T) {
	suspend := false
	js := makeJobSet("", "", suspend)
	// An active condition with an unrecognized type → falls through to Running.
	js.Status.Conditions = []metav1.Condition{
		{
			Type:               "SomeActiveCondition",
			Status:             metav1.ConditionTrue,
			LastTransitionTime: metav1.NewTime(time.Now()),
		},
	}

	spec := &clusteredpb.ClusteredTaskSpec{Replicas: 2, NprocPerNode: 1}
	pCtx := dummyPluginCtx(buildTaskTemplate(spec), emptyK8sReader())

	handler := clusteredResourceHandler{}
	phase, err := handler.GetTaskPhase(context.Background(), pCtx, js)
	assert.NoError(t, err)
	assert.Equal(t, pluginsCore.PhaseRunning, phase.Phase())
}

// --- fast-fail / maintenance tests ---

func TestGetTaskPhase_FastFail_NoJobsFailed(t *testing.T) {
	// When no jobs have failed in ReplicatedJobsStatus, the fast-fail path is not taken.
	js := makeJobSet("", "", false)
	// Explicitly set workers status with Failed=0.
	js.Status.ReplicatedJobsStatus = []jobsetv1alpha2.ReplicatedJobStatus{
		{Name: "workers", Failed: 0, Active: 2},
	}
	// Add an active condition so the switch falls through to running.
	js.Status.Conditions = []metav1.Condition{
		{
			Type:               "SomeActiveCondition",
			Status:             metav1.ConditionTrue,
			LastTransitionTime: metav1.NewTime(time.Now()),
		},
	}

	spec := &clusteredpb.ClusteredTaskSpec{Replicas: 2, NprocPerNode: 1}
	pCtx := dummyPluginCtx(buildTaskTemplate(spec), emptyK8sReader())

	handler := clusteredResourceHandler{}
	phase, err := handler.GetTaskPhase(context.Background(), pCtx, js)
	assert.NoError(t, err)
	// No pod inspection happens — returns Running.
	assert.Equal(t, pluginsCore.PhaseRunning, phase.Phase())
}

func TestGetTaskPhase_MaintenanceRetry_FlagFalse(t *testing.T) {
	// With RestartOnHostMaintenance=false (default), JobSetFailed always becomes RetryableFailure.
	js := makeJobSet(jobsetv1alpha2.JobSetFailed, metav1.ConditionTrue, false)

	spec := &clusteredpb.ClusteredTaskSpec{
		Replicas:      2,
		NprocPerNode:  1,
		FailurePolicy: &clusteredpb.ClusterFailurePolicy{RestartOnHostMaintenance: false},
	}
	pCtx := dummyPluginCtx(buildTaskTemplate(spec), emptyK8sReader())

	handler := clusteredResourceHandler{}
	phase, err := handler.GetTaskPhase(context.Background(), pCtx, js)
	assert.NoError(t, err)
	// Flag is false → no pod lookup → normal retryable failure.
	assert.Equal(t, pluginsCore.PhaseRetryableFailure, phase.Phase())
}

func TestGetTaskPhase_FastFail_Worker0Failed(t *testing.T) {
	// When Failed>0 for the workers ReplicatedJob, the plugin inspects the rank-0 pod.
	// A pod with a non-zero exit code should surface PhaseRetryableFailure immediately.
	js := makeJobSet("", "", false)
	js.Status.ReplicatedJobsStatus = []jobsetv1alpha2.ReplicatedJobStatus{
		{Name: workersReplicatedJobName, Failed: 1, Active: 1},
	}
	// An active unrecognized condition is required for the switch to fall through to the fast-fail path.
	js.Status.Conditions = []metav1.Condition{
		{Type: "SomeActiveCondition", Status: metav1.ConditionTrue, LastTransitionTime: metav1.NewTime(time.Now())},
	}

	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      rank0PodName(testJobName) + "-abc12",
			Namespace: testNS,
		},
		Status: corev1.PodStatus{
			Phase:  corev1.PodFailed,
			Reason: "Error",
			ContainerStatuses: []corev1.ContainerStatus{
				{
					Name: "primary",
					State: corev1.ContainerState{
						Terminated: &corev1.ContainerStateTerminated{ExitCode: 1, Reason: "Error"},
					},
				},
			},
		},
	}
	fakeClient := fake.NewClientBuilder().WithScheme(k8sscheme.Scheme).WithObjects(pod).Build()

	spec := &clusteredpb.ClusteredTaskSpec{Replicas: 2, NprocPerNode: 1}
	pCtx := dummyPluginCtx(buildTaskTemplate(spec), fakeClient)

	handler := clusteredResourceHandler{}
	phase, err := handler.GetTaskPhase(context.Background(), pCtx, js)
	assert.NoError(t, err)
	assert.Equal(t, pluginsCore.PhaseRetryableFailure, phase.Phase())
	assert.Equal(t, core.ExecutionError_USER, phase.Err().GetKind())
}

func TestGetTaskPhase_MaintenanceRetry_SystemFailure(t *testing.T) {
	// When RestartOnHostMaintenance=true and the rank-0 pod failed due to a node shutdown
	// (system-retryable reason), the plugin returns PhaseRetryableFailure with SYSTEM kind
	// so Flyte retries without consuming the user's max_restarts budget.
	js := makeJobSet(jobsetv1alpha2.JobSetFailed, metav1.ConditionTrue, false)

	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      rank0PodName(testJobName) + "-abc12",
			Namespace: testNS,
		},
		Status: corev1.PodStatus{
			Phase:  corev1.PodFailed,
			Reason: "Shutdown",
		},
	}
	fakeClient := fake.NewClientBuilder().WithScheme(k8sscheme.Scheme).WithObjects(pod).Build()

	spec := &clusteredpb.ClusteredTaskSpec{
		Replicas:      2,
		NprocPerNode:  1,
		FailurePolicy: &clusteredpb.ClusterFailurePolicy{RestartOnHostMaintenance: true},
	}
	pCtx := dummyPluginCtx(buildTaskTemplate(spec), fakeClient)

	handler := clusteredResourceHandler{}
	phase, err := handler.GetTaskPhase(context.Background(), pCtx, js)
	assert.NoError(t, err)
	assert.Equal(t, pluginsCore.PhaseRetryableFailure, phase.Phase())
	assert.Equal(t, core.ExecutionError_SYSTEM, phase.Err().GetKind())
}

func TestFindRank0Pod_SuffixedAndDeterministic(t *testing.T) {
	oldFailed := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:              rank0PodName(testJobName) + "-aaaa1",
			Namespace:         testNS,
			CreationTimestamp: metav1.NewTime(time.Now().Add(-2 * time.Minute)),
		},
		Status: corev1.PodStatus{Phase: corev1.PodFailed},
	}
	newRunning := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:              rank0PodName(testJobName) + "-bbbb2",
			Namespace:         testNS,
			CreationTimestamp: metav1.NewTime(time.Now()),
		},
		Status: corev1.PodStatus{Phase: corev1.PodRunning},
	}
	otherPod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      testJobName + "-workers-0-1-ccccc",
			Namespace: testNS,
		},
		Status: corev1.PodStatus{Phase: corev1.PodRunning},
	}
	fakeClient := fake.NewClientBuilder().WithScheme(k8sscheme.Scheme).WithObjects(oldFailed, newRunning, otherPod).Build()

	pCtx := &k8smocks.PluginContext{}
	pCtx.EXPECT().K8sReader().Return(fakeClient)

	js := makeJobSet("", "", false)
	pod := findRank0Pod(context.Background(), pCtx, js)
	assert.NotNil(t, pod)
	assert.Equal(t, newRunning.Name, pod.Name)
}

func TestGetTaskPhase_FastFail_FailedWithBudgetRemainingReturnsRunning(t *testing.T) {
	js := makeJobSet("", "", false)
	js.Spec.FailurePolicy = &jobsetv1alpha2.FailurePolicy{MaxRestarts: 2}
	js.Status.Restarts = 1
	js.Status.ReplicatedJobsStatus = []jobsetv1alpha2.ReplicatedJobStatus{
		{Name: workersReplicatedJobName, Failed: 1, Active: 1},
	}
	js.Status.Conditions = []metav1.Condition{
		{Type: "SomeActiveCondition", Status: metav1.ConditionTrue, LastTransitionTime: metav1.NewTime(time.Now())},
	}

	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: rank0PodName(testJobName) + "-abc12", Namespace: testNS},
		Status: corev1.PodStatus{
			Phase: corev1.PodFailed,
			ContainerStatuses: []corev1.ContainerStatus{
				{Name: "primary", State: corev1.ContainerState{Terminated: &corev1.ContainerStateTerminated{ExitCode: 1}}},
			},
		},
	}
	fakeClient := fake.NewClientBuilder().WithScheme(k8sscheme.Scheme).WithObjects(pod).Build()

	spec := &clusteredpb.ClusteredTaskSpec{
		Replicas:      2,
		NprocPerNode:  1,
		FailurePolicy: &clusteredpb.ClusterFailurePolicy{MaxRestarts: 2},
	}
	pCtx := dummyPluginCtx(buildTaskTemplate(spec), fakeClient)

	handler := clusteredResourceHandler{}
	phase, err := handler.GetTaskPhase(context.Background(), pCtx, js)
	assert.NoError(t, err)
	assert.Equal(t, pluginsCore.PhaseRunning, phase.Phase())
}

func TestGetTaskPhase_FastFail_FailedWithBudgetExhaustedReturnsRetryableFailure(t *testing.T) {
	js := makeJobSet("", "", false)
	js.Spec.FailurePolicy = &jobsetv1alpha2.FailurePolicy{MaxRestarts: 1}
	js.Status.Restarts = 1
	js.Status.ReplicatedJobsStatus = []jobsetv1alpha2.ReplicatedJobStatus{
		{Name: workersReplicatedJobName, Failed: 1, Active: 1},
	}
	js.Status.Conditions = []metav1.Condition{
		{Type: "SomeActiveCondition", Status: metav1.ConditionTrue, LastTransitionTime: metav1.NewTime(time.Now())},
	}

	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: rank0PodName(testJobName) + "-abc12", Namespace: testNS},
		Status: corev1.PodStatus{
			Phase: corev1.PodFailed,
			ContainerStatuses: []corev1.ContainerStatus{
				{Name: "primary", State: corev1.ContainerState{Terminated: &corev1.ContainerStateTerminated{ExitCode: 1}}},
			},
		},
	}
	fakeClient := fake.NewClientBuilder().WithScheme(k8sscheme.Scheme).WithObjects(pod).Build()

	spec := &clusteredpb.ClusteredTaskSpec{
		Replicas:      2,
		NprocPerNode:  1,
		FailurePolicy: &clusteredpb.ClusterFailurePolicy{MaxRestarts: 1},
	}
	pCtx := dummyPluginCtx(buildTaskTemplate(spec), fakeClient)

	handler := clusteredResourceHandler{}
	phase, err := handler.GetTaskPhase(context.Background(), pCtx, js)
	assert.NoError(t, err)
	assert.Equal(t, pluginsCore.PhaseRetryableFailure, phase.Phase())
}

func TestGetTaskPhase_FastFail_PendingImagePullRegardlessBudget(t *testing.T) {
	js := makeJobSet("", "", false)
	js.Spec.FailurePolicy = &jobsetv1alpha2.FailurePolicy{MaxRestarts: 3}
	js.Status.Restarts = 1
	js.Status.Conditions = []metav1.Condition{
		{Type: "SomeActiveCondition", Status: metav1.ConditionTrue, LastTransitionTime: metav1.NewTime(time.Now())},
	}

	oldTransition := metav1.NewTime(time.Now().Add(-24 * time.Hour))
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: rank0PodName(testJobName) + "-abc12", Namespace: testNS},
		Status: corev1.PodStatus{
			Phase: corev1.PodPending,
			Conditions: []corev1.PodCondition{
				{Type: corev1.PodReady, Status: corev1.ConditionFalse, Reason: "ContainersNotReady", LastTransitionTime: oldTransition},
			},
			ContainerStatuses: []corev1.ContainerStatus{
				{
					Name:  "primary",
					Ready: false,
					State: corev1.ContainerState{
						Waiting: &corev1.ContainerStateWaiting{
							Reason:  "ImagePullBackOff",
							Message: "Back-off pulling image",
						},
					},
				},
			},
		},
	}
	fakeClient := fake.NewClientBuilder().WithScheme(k8sscheme.Scheme).WithObjects(pod).Build()

	spec := &clusteredpb.ClusteredTaskSpec{
		Replicas:      2,
		NprocPerNode:  1,
		FailurePolicy: &clusteredpb.ClusterFailurePolicy{MaxRestarts: 3},
	}
	pCtx := dummyPluginCtx(buildTaskTemplate(spec), fakeClient)

	handler := clusteredResourceHandler{}
	phase, err := handler.GetTaskPhase(context.Background(), pCtx, js)
	assert.NoError(t, err)
	assert.Equal(t, pluginsCore.PhaseRetryableFailure, phase.Phase())
}

func TestGetTaskPhase_NoCondition_ZeroBudgetFailureFastFails(t *testing.T) {
	// maxRestarts == 0 and a worker has failed, but the JobSet controller has not yet
	// written any condition. hasJobSetStarted must still treat this as started so the
	// failure is surfaced via maybeFastFailWorker0 instead of falling back to Initializing.
	js := makeJobSet("", "", false)
	js.Spec.FailurePolicy = &jobsetv1alpha2.FailurePolicy{MaxRestarts: 0}
	js.Status.ReplicatedJobsStatus = []jobsetv1alpha2.ReplicatedJobStatus{
		{Name: workersReplicatedJobName, Failed: 1},
	}

	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: rank0PodName(testJobName) + "-abc12", Namespace: testNS},
		Status: corev1.PodStatus{
			Phase: corev1.PodFailed,
			ContainerStatuses: []corev1.ContainerStatus{
				{Name: "primary", State: corev1.ContainerState{Terminated: &corev1.ContainerStateTerminated{ExitCode: 1}}},
			},
		},
	}
	fakeClient := fake.NewClientBuilder().WithScheme(k8sscheme.Scheme).WithObjects(pod).Build()

	spec := &clusteredpb.ClusteredTaskSpec{
		Replicas:      2,
		NprocPerNode:  1,
		FailurePolicy: &clusteredpb.ClusterFailurePolicy{MaxRestarts: 0},
	}
	pCtx := dummyPluginCtx(buildTaskTemplate(spec), fakeClient)

	handler := clusteredResourceHandler{}
	phase, err := handler.GetTaskPhase(context.Background(), pCtx, js)
	assert.NoError(t, err)
	assert.Equal(t, pluginsCore.PhaseRetryableFailure, phase.Phase())
}

func TestGetTaskPhase_RestartingCondition_ReportsRunningWithAttempt(t *testing.T) {
	js := makeJobSet("", "", false)
	js.Spec.FailurePolicy = &jobsetv1alpha2.FailurePolicy{MaxRestarts: 1}
	js.Status.Restarts = 1
	js.Status.ReplicatedJobsStatus = []jobsetv1alpha2.ReplicatedJobStatus{
		{Name: workersReplicatedJobName, Failed: 1},
	}
	js.Status.Conditions = []metav1.Condition{
		{Type: jobSetRestartingConditionType, Status: metav1.ConditionTrue, LastTransitionTime: metav1.NewTime(time.Now())},
	}

	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: rank0PodName(testJobName) + "-abc12", Namespace: testNS},
		Status: corev1.PodStatus{
			Phase: corev1.PodFailed,
			ContainerStatuses: []corev1.ContainerStatus{
				{Name: "primary", State: corev1.ContainerState{Terminated: &corev1.ContainerStateTerminated{ExitCode: 1}}},
			},
		},
	}
	fakeClient := fake.NewClientBuilder().WithScheme(k8sscheme.Scheme).WithObjects(pod).Build()
	pCtx := dummyPluginCtx(buildTaskTemplate(&clusteredpb.ClusteredTaskSpec{Replicas: 2, NprocPerNode: 1}), fakeClient)

	handler := clusteredResourceHandler{}
	phase, err := handler.GetTaskPhase(context.Background(), pCtx, js)
	assert.NoError(t, err)
	assert.Equal(t, pluginsCore.PhaseRunning, phase.Phase())
	assert.Contains(t, phase.Reason(), "restart in progress (attempt 1)")
}

func TestGetTaskPhase_NoTrueConditionWithRestarts_ReportsRunning(t *testing.T) {
	js := makeJobSet("", "", false)
	js.Status.Restarts = 1
	js.Status.Conditions = []metav1.Condition{
		{Type: string(jobsetv1alpha2.JobSetSuspended), Status: metav1.ConditionFalse, LastTransitionTime: metav1.NewTime(time.Now())},
		{Type: string(jobsetv1alpha2.JobSetCompleted), Status: metav1.ConditionFalse, LastTransitionTime: metav1.NewTime(time.Now())},
	}

	pCtx := dummyPluginCtx(buildTaskTemplate(&clusteredpb.ClusteredTaskSpec{Replicas: 2, NprocPerNode: 1}), emptyK8sReader())

	handler := clusteredResourceHandler{}
	phase, err := handler.GetTaskPhase(context.Background(), pCtx, js)
	assert.NoError(t, err)
	assert.Equal(t, pluginsCore.PhaseRunning, phase.Phase())
	assert.Contains(t, phase.Reason(), "restart attempt 1")
}

func TestGetTaskPhase_NoConditionWithPriorRunningState_ReportsRunning(t *testing.T) {
	js := makeJobSet("", "", false)
	pCtx := dummyPluginCtxWithState(
		buildTaskTemplate(&clusteredpb.ClusteredTaskSpec{Replicas: 2, NprocPerNode: 1}),
		emptyK8sReader(),
		plugink8s.PluginState{Phase: pluginsCore.PhaseRunning, PhaseVersion: 1},
		nil,
	)

	handler := clusteredResourceHandler{}
	phase, err := handler.GetTaskPhase(context.Background(), pCtx, js)
	assert.NoError(t, err)
	assert.Equal(t, pluginsCore.PhaseRunning, phase.Phase())
}

func TestGetTaskPhase_NoTrueCondition_StateReadErrorFallsBackToStatus(t *testing.T) {
	js := makeJobSet("", "", false)
	js.Status.Restarts = 1
	js.Status.Conditions = []metav1.Condition{
		{Type: string(jobsetv1alpha2.JobSetCompleted), Status: metav1.ConditionFalse, LastTransitionTime: metav1.NewTime(time.Now())},
	}

	pCtx := dummyPluginCtxWithState(
		buildTaskTemplate(&clusteredpb.ClusteredTaskSpec{Replicas: 2, NprocPerNode: 1}),
		emptyK8sReader(),
		plugink8s.PluginState{},
		errors.New("state read failed"),
	)

	handler := clusteredResourceHandler{}
	phase, err := handler.GetTaskPhase(context.Background(), pCtx, js)
	assert.NoError(t, err)
	assert.Equal(t, pluginsCore.PhaseRunning, phase.Phase())
}

func TestGetTaskPhase_LogContext(t *testing.T) {
	const primaryContainer = "primary"
	const sidecarContainer = "sidecar"

	// mkPod builds a realistic JobSet child pod: a primary container plus a sidecar,
	// with matching container statuses so BuildPodLogContext produces real container
	// contexts. Pending pods carry no statuses.
	mkPod := func(name string, phase corev1.PodPhase) *corev1.Pod {
		pod := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: testNS,
			},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{{Name: primaryContainer}, {Name: sidecarContainer}},
			},
			Status: corev1.PodStatus{Phase: phase},
		}
		if phase == corev1.PodRunning {
			running := corev1.ContainerState{Running: &corev1.ContainerStateRunning{StartedAt: metav1.NewTime(time.Now())}}
			pod.Status.ContainerStatuses = []corev1.ContainerStatus{
				{Name: primaryContainer, State: running},
				{Name: sidecarContainer, State: running},
			}
		}
		return pod
	}

	// jobSet annotates the authoritative primary container name at build time.
	makeRunningJobSet := func() *jobsetv1alpha2.JobSet {
		js := makeJobSet("", "", false)
		js.Annotations = map[string]string{primaryContainerAnnotation: primaryContainer}
		js.Status.Conditions = []metav1.Condition{
			{Type: "SomeActiveCondition", Status: metav1.ConditionTrue, LastTransitionTime: metav1.NewTime(time.Now())},
		}
		return js
	}

	// Real JobSet pods carry a random suffix after the "<jobset>-workers-<job>-<idx>" stem.
	rank0 := rank0PodName(testJobName) + "-x1y2z"
	rank1 := testJobName + "-workers-0-1-a9b8c"
	rank2 := testJobName + "-workers-0-2-pppp"

	t.Run("primary pod and container resolved from live pods", func(t *testing.T) {
		js := makeRunningJobSet()
		fakeClient := fake.NewClientBuilder().WithScheme(k8sscheme.Scheme).
			WithObjects(
				mkPod(rank0, corev1.PodRunning),
				mkPod(rank1, corev1.PodRunning),
				mkPod(rank2, corev1.PodPending),
			).Build()

		spec := &clusteredpb.ClusteredTaskSpec{Replicas: 2, NprocPerNode: 1}
		pCtx := dummyPluginCtx(buildTaskTemplate(spec), fakeClient)

		handler := clusteredResourceHandler{}
		phase, err := handler.GetTaskPhase(context.Background(), pCtx, js)
		assert.NoError(t, err)
		assert.Equal(t, pluginsCore.PhaseRunning, phase.Phase())

		lc := phase.Info().LogContext
		assert.NotNil(t, lc)
		assert.Equal(t, rank0, lc.PrimaryPodName)
		// Pending pod is excluded → only the two running pods remain.
		assert.Len(t, lc.Pods, 2)
		names := []string{lc.Pods[0].GetPodName(), lc.Pods[1].GetPodName()}
		assert.Contains(t, names, rank0)
		assert.Contains(t, names, rank1)

		// Each pod's primary container comes from the JobSet annotation (not the
		// sidecar / first container), and container contexts are populated.
		for _, p := range lc.Pods {
			assert.Equal(t, primaryContainer, p.GetPrimaryContainerName())
			assert.GreaterOrEqual(t, len(p.GetContainers()), 1)
		}
	})

	t.Run("primary falls back when rank-0 pod is pending", func(t *testing.T) {
		js := makeRunningJobSet()
		fakeClient := fake.NewClientBuilder().WithScheme(k8sscheme.Scheme).
			WithObjects(
				mkPod(rank0, corev1.PodPending),
				mkPod(rank1, corev1.PodRunning),
			).Build()

		spec := &clusteredpb.ClusteredTaskSpec{Replicas: 2, NprocPerNode: 1}
		pCtx := dummyPluginCtx(buildTaskTemplate(spec), fakeClient)

		handler := clusteredResourceHandler{}
		phase, err := handler.GetTaskPhase(context.Background(), pCtx, js)
		assert.NoError(t, err)

		lc := phase.Info().LogContext
		assert.NotNil(t, lc)
		// rank-0 is pending and excluded → PrimaryPodName must still reference an
		// included pod so downstream log streaming can resolve it.
		assert.Len(t, lc.Pods, 1)
		assert.Equal(t, rank1, lc.PrimaryPodName)
		assert.Equal(t, lc.Pods[0].GetPodName(), lc.PrimaryPodName)
	})
}

// --- IsTerminal / GetCompletionTime ---

func TestIsTerminal(t *testing.T) {
	handler := clusteredResourceHandler{}

	js := makeJobSet(jobsetv1alpha2.JobSetCompleted, metav1.ConditionTrue, false)
	ok, err := handler.IsTerminal(context.Background(), js)
	assert.NoError(t, err)
	assert.True(t, ok)

	js2 := makeJobSet("", "", false)
	ok2, err := handler.IsTerminal(context.Background(), js2)
	assert.NoError(t, err)
	assert.False(t, ok2)
}

func TestGetCompletionTime(t *testing.T) {
	handler := clusteredResourceHandler{}
	js := makeJobSet(jobsetv1alpha2.JobSetCompleted, metav1.ConditionTrue, false)
	ts, err := handler.GetCompletionTime(js)
	assert.NoError(t, err)
	assert.False(t, ts.IsZero())
}

// --- buildJobSetName tests ---

// longestPodName reproduces the worst-case pod name JobSet's admission webhook validates:
// "<jobSetName>-<replicatedJob>-<jobIdx>-<podIdx>-<5-char random suffix>" (jobIdx is always
// 0 here since the single ReplicatedJob has Replicas=1; podIdx maxes at replicas-1).
func longestPodName(jobSetName string, replicas int32) string {
	maxPodIdx := strconv.Itoa(int(replicas - 1))
	return placement.GenPodName(jobSetName, workersReplicatedJobName, "0", maxPodIdx) + "-abcde"
}

func TestBuildJobSetName_ShortNameUnchanged(t *testing.T) {
	name := buildJobSetName("f-abc123")
	assert.Equal(t, "f-abc123", name)
}

func TestBuildJobSetName_SanitizesInvalidChars(t *testing.T) {
	// Uppercase/underscore aren't DNS-1123 subdomain compatible; they must be normalized.
	name := buildJobSetName("My_Run")
	assert.Empty(t, validation.IsDNS1035Label(name))
	assert.NotContains(t, name, "_")
	assert.Equal(t, strings.ToLower(name), name)
}

func TestBuildJobSetName_CoercesToDNS1035Label(t *testing.T) {
	// The JobSet name becomes a component of the child pod names, which the webhook
	// validates as DNS-1035 labels: no dots, must start with a letter. A generated
	// name with a dot or a leading digit must not leak into the derived names.
	for _, generated := range []string{"my.run-a0-0", "9run-a0-0", "0.1.2", "svc.default.a0"} {
		name := buildJobSetName(generated)
		assert.Empty(t, validation.IsDNS1035Label(name), "jobset name %q (from %q) is not a valid DNS-1035 label", name, generated)
		assert.NotContains(t, name, ".", "jobset name %q retained a dot", name)
		podName := longestPodName(name, maxReplicasForNaming)
		assert.Empty(t, validation.IsDNS1035Label(podName), "derived pod name %q invalid (from %q)", podName, generated)
	}
}

func TestBuildJobSetName_LongNameFitsPodLimit(t *testing.T) {
	// A composed/nested task can produce a generated name well beyond the budget.
	long := strings.Repeat("composed-subtask-", 8) + "tail" // ~140 chars
	name := buildJobSetName(long)

	// The name is derived from the generated name alone (no replica count), so a single
	// bounded name must keep the longest pod name valid across every supported replica
	// count, including the worst case the truncation reserves for. The name itself must
	// be a valid DNS-1035 label, since it's embedded in the derived Job/Pod names.
	assert.Empty(t, validation.IsDNS1035Label(name), "jobset name not a valid DNS-1035 label")
	for _, replicas := range []int32{1, 4, 16, 128, 1024, 10000, maxReplicasForNaming} {
		podName := longestPodName(name, replicas)
		assert.LessOrEqual(t, len(podName), dns1035LabelMaxLength, "pod name %q (%d chars) exceeds limit for replicas=%d", podName, len(podName), replicas)
		assert.Empty(t, validation.IsDNS1035Label(podName), "pod name %q invalid for replicas=%d", podName, replicas)
	}
}

func TestBuildJobSetName_LongNamesAreDistinct(t *testing.T) {
	// Two different long names that share a prefix must not collide after truncation.
	a := buildJobSetName(strings.Repeat("a", 60) + "-one")
	b := buildJobSetName(strings.Repeat("a", 60) + "-two")
	assert.NotEqual(t, a, b)
}

// TestGeneratedNameMaxLength_BoundsSourceName guards the source-of-truth bound: plugin
// managers may stamp GetGeneratedName() directly onto the JobSet, overwriting the name
// BuildResource chose, so the plugin advertises GeneratedNameMaxLength and every name
// within that bound must (a) keep the worst-case derived pod name within the 63-char
// limit and (b) pass through buildJobSetName unchanged, so a manager-stamped name and
// the plugin-built name are identical.
func TestGeneratedNameMaxLength_BoundsSourceName(t *testing.T) {
	props := clusteredResourceHandler{}.GetProperties()
	if assert.NotNil(t, props.GeneratedNameMaxLength) {
		assert.Equal(t, generatedNameMaxLength, *props.GeneratedNameMaxLength)
	}

	// A generated name at exactly the advertised bound must survive untouched and
	// still fit the pod-name budget at the worst-case replica count.
	atBound := "g" + strings.Repeat("a", generatedNameMaxLength-1)
	assert.Equal(t, atBound, buildJobSetName(atBound))
	podName := longestPodName(atBound, maxReplicasForNaming)
	assert.LessOrEqual(t, len(podName), dns1035LabelMaxLength, "pod name %q (%d chars) exceeds limit", podName, len(podName))
	assert.Empty(t, validation.IsDNS1035Label(podName), "pod name %q invalid", podName)
}

func TestBuildJobSetName_NoTrailingSeparator(t *testing.T) {
	// Truncation must not leave a trailing '-' or '.', which would be invalid.
	name := buildJobSetName(strings.Repeat("x", 50) + "." + strings.Repeat("y", 50))
	assert.NotEqual(t, "-", name[len(name)-1:])
	assert.NotEqual(t, ".", name[len(name)-1:])
	assert.Empty(t, validation.IsDNS1035Label(name))
}

// TestBuildResourceAndIdentityNameMatch guards the create/lookup invariant: BuildResource
// (create) and BuildIdentityResource (lookup/abort) must name the JobSet identically, even
// for a long composed/nested generated name, or the plugin manager can't find the object it
// created. The shared name must also keep the longest derived pod name within 63 chars.
func TestBuildResourceAndIdentityNameMatch(t *testing.T) {
	longGeneratedName := strings.Repeat("composed-subtask-", 8) + "tail-0" // ~140 chars
	spec := &clusteredpb.ClusteredTaskSpec{
		Replicas:     4,
		NprocPerNode: 8,
		Runtime: &clusteredpb.Runtime{
			Kind: &clusteredpb.Runtime_Torchrun{
				Torchrun: &clusteredpb.TorchRuntime{RdzvBackend: clusteredpb.RdzvBackend_STATIC},
			},
		},
	}
	taskCtx := dummyTaskCtxWithGeneratedName(buildTaskTemplate(spec), longGeneratedName)
	handler := clusteredResourceHandler{}

	created, err := handler.BuildResource(context.Background(), taskCtx)
	assert.NoError(t, err)

	identity, err := handler.BuildIdentityResource(context.Background(), taskCtx.TaskExecutionMetadata())
	assert.NoError(t, err)

	assert.Equal(t, created.GetName(), identity.GetName(), "create and lookup names diverge")
	assert.NotEqual(t, longGeneratedName, created.GetName(), "long name should have been truncated")
	assert.Empty(t, validation.IsDNS1035Label(created.GetName()))
	podName := longestPodName(created.GetName(), maxReplicasForNaming)
	assert.LessOrEqual(t, len(podName), dns1035LabelMaxLength)
}

// TestBuildResource_ReplicasExceedNamingBudget verifies BuildResource fails fast with a
// spec error when replicas exceeds the pod-index budget buildJobSetName reserves for; past
// that bound the derived pod names could exceed 63 chars and be rejected by the webhook.
func TestBuildResource_ReplicasExceedNamingBudget(t *testing.T) {
	spec := &clusteredpb.ClusteredTaskSpec{
		Replicas:     maxReplicasForNaming + 1,
		NprocPerNode: 1,
		Runtime: &clusteredpb.Runtime{
			Kind: &clusteredpb.Runtime_Torchrun{Torchrun: &clusteredpb.TorchRuntime{}},
		},
	}
	taskCtx := dummyTaskCtx(buildTaskTemplate(spec))
	handler := clusteredResourceHandler{}

	_, err := handler.BuildResource(context.Background(), taskCtx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "replicas must be <=")
}
