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
	"github.com/stretchr/testify/require"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation"
	k8sscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	jobsetv1alpha2 "sigs.k8s.io/jobset/api/jobset/v1alpha2"
	"sigs.k8s.io/jobset/pkg/util/placement"

	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery"
	pluginsCore "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/core"
	coreMocks "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/core/mocks"
	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/encoding"
	pluginIOMocks "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/io/mocks"
	plugink8s "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/k8s"
	k8smocks "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/k8s/mocks"
	"github.com/flyteorg/flyte/v2/flytestdlib/utils"
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
	custom, err := utils.MarshalObjToStruct(spec)
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
func dummyTaskCtx(taskTemplate *core.TaskTemplate, podTemplate *core.K8SPod) *coreMocks.TaskExecutionContext {
	return dummyTaskCtxWithGeneratedName(taskTemplate, testJobName, podTemplate)
}

// dummyTaskCtxWithGeneratedName is dummyTaskCtx with a caller-supplied generated name, used to
// exercise the long composed/nested-name truncation path.
func dummyTaskCtxWithGeneratedName(taskTemplate *core.TaskTemplate, generatedName string, podTemplate *core.K8SPod) *coreMocks.TaskExecutionContext {
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
	// Mirrors the executor's implementation: bound with a hash when over maxLength.
	tID.EXPECT().GetGeneratedNameWith(mock.Anything, mock.Anything).RunAndReturn(func(minLength, maxLength int) (string, error) {
		return encoding.FixedLengthUniqueID(generatedName, maxLength)
	}).Maybe()
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
	overrides.EXPECT().GetPodTemplate().Return(podTemplate)

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
	taskCtx := dummyTaskCtx(taskTemplate, nil)

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
	// Without restart_on_host_maintenance the inner Job carries no podFailurePolicy.
	assert.Nil(t, jobSpec.PodFailurePolicy)

	// The node-execution labels/annotations must be propagated onto the pod template so
	// JobSet child pods carry execution-id/node-id; otherwise the node-execution-scoped
	// K8sReader.List in getLogContext returns nothing and no logs reach the UI.
	podMeta := jobSpec.Template.ObjectMeta
	assert.Equal(t, "my-exec", podMeta.Labels["execution-id"])
	assert.Equal(t, "n1", podMeta.Labels["node-id"])
	assert.Equal(t, "av", podMeta.Annotations["flyte.org/test-annotation"])
}

func TestBuildResource_PropagatesPodTemplateMetadataToJobSet(t *testing.T) {
	spec := &clusteredpb.ClusteredTaskSpec{
		Replicas:     2,
		NprocPerNode: 1,
		Runtime: &clusteredpb.Runtime{
			Kind: &clusteredpb.Runtime_Torchrun{
				Torchrun: &clusteredpb.TorchRuntime{},
			},
		},
	}

	podSpec, err := utils.MarshalObjToStruct(corev1.PodSpec{
		Containers: []corev1.Container{
			{
				Name:  "primary",
				Image: testImage,
			},
		},
	})
	require.NoError(t, err)

	podTemplate := &core.K8SPod{
		Metadata: &core.K8SObjectMetadata{
			Labels: map[string]string{
				"kueue.x-k8s.io/queue-name": "test-queue",
				"flyte.org/execution":       "user-provided",
			},
			Annotations: map[string]string{
				"example.org/custom":       "value",
				"flyte.org/task-type":      "user-provided",
				primaryContainerAnnotation: "user-provided",
			},
		},
		PodSpec:              podSpec,
		PrimaryContainerName: "primary",
	}
	taskCtx := dummyTaskCtx(buildTaskTemplate(spec), podTemplate)

	obj, err := clusteredResourceHandler{}.BuildResource(context.Background(), taskCtx)
	require.NoError(t, err)

	jobSet, ok := obj.(*jobsetv1alpha2.JobSet)
	require.True(t, ok, "expected *JobSet")

	podMeta := jobSet.Spec.ReplicatedJobs[0].Template.Spec.Template.ObjectMeta

	// Confirm the fixture reached the existing child pod-template path.
	assert.Equal(t, "test-queue", podMeta.Labels["kueue.x-k8s.io/queue-name"])
	assert.Equal(t, "value", podMeta.Annotations["example.org/custom"])

	// The same metadata must be available to controllers watching the parent JobSet.
	assert.Equal(t, "test-queue", jobSet.Labels["kueue.x-k8s.io/queue-name"])
	assert.Equal(t, "value", jobSet.Annotations["example.org/custom"])

	// Flyte-owned parent metadata must take precedence over user-provided values.
	assert.Equal(t, "my-exec", jobSet.Labels["flyte.org/execution"])
	assert.Equal(t, taskType, jobSet.Annotations["flyte.org/task-type"])
	assert.Equal(t, "primary", jobSet.Annotations[primaryContainerAnnotation])
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
	taskCtx := dummyTaskCtx(taskTemplate, nil)

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

func TestBuildResource_HostMaintenance(t *testing.T) {
	spec := &clusteredpb.ClusteredTaskSpec{
		Replicas:      2,
		NprocPerNode:  1,
		FailurePolicy: &clusteredpb.ClusterFailurePolicy{MaxRestarts: 1, RestartOnHostMaintenance: true},
	}
	taskTemplate := buildTaskTemplate(spec)
	taskCtx := dummyTaskCtx(taskTemplate, nil)

	handler := clusteredResourceHandler{}
	obj, err := handler.BuildResource(context.Background(), taskCtx)
	assert.NoError(t, err)

	jobSet := obj.(*jobsetv1alpha2.JobSet)

	// JobSet failurePolicy carries the free-restart rule.
	require.NotNil(t, jobSet.Spec.FailurePolicy)
	assert.Equal(t, int32(1), jobSet.Spec.FailurePolicy.MaxRestarts)
	require.Len(t, jobSet.Spec.FailurePolicy.Rules, 1)
	assert.Equal(t, jobsetv1alpha2.RestartJobSetAndIgnoreMaxRestarts, jobSet.Spec.FailurePolicy.Rules[0].Action)

	// The inner Job fails with reason PodFailurePolicy on DisruptionTarget so the
	// JobSet rule can distinguish maintenance disruptions from ordinary failures.
	jobSpec := jobSet.Spec.ReplicatedJobs[0].Template.Spec
	require.NotNil(t, jobSpec.PodFailurePolicy)
	require.Len(t, jobSpec.PodFailurePolicy.Rules, 1)
	rule := jobSpec.PodFailurePolicy.Rules[0]
	assert.Equal(t, batchv1.PodFailurePolicyActionFailJob, rule.Action)
	require.Len(t, rule.OnPodConditions, 1)
	assert.Equal(t, corev1.DisruptionTarget, rule.OnPodConditions[0].Type)
	assert.Equal(t, corev1.ConditionTrue, rule.OnPodConditions[0].Status)
	// podFailurePolicy requires restartPolicy Never on the pod template.
	assert.Equal(t, corev1.RestartPolicyNever, jobSpec.Template.Spec.RestartPolicy)
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
	// Without restart_on_host_maintenance no rules are emitted — every restart
	// counts against the budget.
	assert.Empty(t, fp.Rules)
}

func TestBuildFailurePolicy_HostMaintenance(t *testing.T) {
	spec := &clusteredpb.ClusteredTaskSpec{
		FailurePolicy: &clusteredpb.ClusterFailurePolicy{MaxRestarts: 3, RestartOnHostMaintenance: true},
	}
	fp, err := buildFailurePolicy(spec)
	assert.NoError(t, err)
	require.NotNil(t, fp)
	assert.Equal(t, int32(3), fp.MaxRestarts)
	require.Len(t, fp.Rules, 1)
	assert.Equal(t, hostMaintenanceRuleName, fp.Rules[0].Name)
	assert.Equal(t, jobsetv1alpha2.RestartJobSetAndIgnoreMaxRestarts, fp.Rules[0].Action)
	assert.Equal(t, []string{batchv1.JobReasonPodFailurePolicy}, fp.Rules[0].OnJobFailureReasons)
}

func TestBuildFailurePolicy_HostMaintenance_ZeroMaxRestarts(t *testing.T) {
	// max_restarts=0 (the default) must not drop the policy when the flag is set:
	// ordinary failures fail immediately, maintenance disruptions still restart free.
	spec := &clusteredpb.ClusteredTaskSpec{
		FailurePolicy: &clusteredpb.ClusterFailurePolicy{MaxRestarts: 0, RestartOnHostMaintenance: true},
	}
	fp, err := buildFailurePolicy(spec)
	assert.NoError(t, err)
	require.NotNil(t, fp)
	assert.Equal(t, int32(0), fp.MaxRestarts)
	require.Len(t, fp.Rules, 1)
	assert.Equal(t, jobsetv1alpha2.RestartJobSetAndIgnoreMaxRestarts, fp.Rules[0].Action)
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

// workerPodLabels are the labels the JobSet controller puts on testJobName's worker pods.
func workerPodLabels() map[string]string {
	return map[string]string{
		jobsetv1alpha2.JobSetNameKey:        testJobName,
		jobsetv1alpha2.ReplicatedJobNameKey: workersReplicatedJobName,
	}
}

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

	// Every worker has a node, none is Ready yet.
	spec := &clusteredpb.ClusteredTaskSpec{Replicas: 2, NprocPerNode: 1}
	pCtx := dummyPluginCtx(buildTaskTemplate(spec), workerPodsReader(workerPodScheduled, workerPodScheduled))

	handler := clusteredResourceHandler{}
	phase, err := handler.GetTaskPhase(context.Background(), pCtx, js)
	assert.NoError(t, err)
	assert.Equal(t, pluginsCore.PhaseInitializing, phase.Phase())
	assert.Equal(t, "0 of 2 workers ready", phase.Reason())
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

func TestGetTaskPhase_UnknownActiveCondition_NotStarted_WaitingForResources(t *testing.T) {
	suspend := false
	js := makeJobSet("", "", suspend)
	// An active condition with an unrecognized type does not imply Running: with no
	// Ready workers the gang has not started, and with no pods on nodes it is waiting
	// for resources.
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
	assert.Equal(t, pluginsCore.PhaseWaitingForResources, phase.Phase())
	assert.Equal(t, "0 of 2 workers scheduled", phase.Reason())
}

func TestGetTaskPhase_Running_AllWorkersReady(t *testing.T) {
	js := makeJobSet("", "", false)
	// Ready counts child Jobs: 1 means every pod of the single workers Job is up.
	js.Status.ReplicatedJobsStatus = []jobsetv1alpha2.ReplicatedJobStatus{
		{Name: workersReplicatedJobName, Ready: 1, Active: 1},
	}
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
	// Explicitly set workers status with Failed=0. Active counts child Jobs with any
	// pod, so Active alone means pods exist, not that the gang is up.
	js.Status.ReplicatedJobsStatus = []jobsetv1alpha2.ReplicatedJobStatus{
		{Name: "workers", Failed: 0, Active: 2},
	}
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
	// No rank-0 pod to inspect and no worker on a node: the gang is still forming.
	assert.Equal(t, pluginsCore.PhaseWaitingForResources, phase.Phase())
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
			Labels:    workerPodLabels(),
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
			Labels:    workerPodLabels(),
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

func TestGetTaskPhase_FreeRestartsDoNotExhaustBudget(t *testing.T) {
	// Free host-maintenance restarts bump Status.Restarts but not
	// Status.RestartsCountTowardsMax. Even with Restarts well past maxRestarts and a
	// failed rank-0 pod visible, the fast-fail path must not fire while the charged
	// count is within budget — the JobSet controller is still restarting the set.
	js := makeJobSet("", "", false)
	js.Status.Restarts = 3
	js.Status.RestartsCountTowardsMax = 0
	js.Status.ReplicatedJobsStatus = []jobsetv1alpha2.ReplicatedJobStatus{
		{Name: workersReplicatedJobName, Failed: 1, Active: 1},
	}
	js.Status.Conditions = []metav1.Condition{
		{Type: "SomeActiveCondition", Status: metav1.ConditionTrue, LastTransitionTime: metav1.NewTime(time.Now())},
	}

	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      rank0PodName(testJobName) + "-abc12",
			Namespace: testNS,
			Labels:    workerPodLabels(),
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

	spec := &clusteredpb.ClusteredTaskSpec{
		Replicas:      2,
		NprocPerNode:  1,
		FailurePolicy: &clusteredpb.ClusterFailurePolicy{MaxRestarts: 1, RestartOnHostMaintenance: true},
	}
	pCtx := dummyPluginCtx(buildTaskTemplate(spec), fakeClient)

	handler := clusteredResourceHandler{}
	phase, err := handler.GetTaskPhase(context.Background(), pCtx, js)
	assert.NoError(t, err)
	assert.Equal(t, pluginsCore.PhaseRunning, phase.Phase())
}

func TestListWorkerPods_OnlyThisJobSetsWorkers(t *testing.T) {
	oldFailed := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:              rank0PodName(testJobName) + "-aaaa1",
			Namespace:         testNS,
			Labels:            workerPodLabels(),
			CreationTimestamp: metav1.NewTime(time.Now().Add(-2 * time.Minute)),
		},
		Status: corev1.PodStatus{Phase: corev1.PodFailed},
	}
	newRunning := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:              rank0PodName(testJobName) + "-bbbb2",
			Namespace:         testNS,
			Labels:            workerPodLabels(),
			CreationTimestamp: metav1.NewTime(time.Now()),
		},
		Status: corev1.PodStatus{Phase: corev1.PodRunning},
	}
	otherPod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      testJobName + "-workers-0-1-ccccc",
			Namespace: testNS,
			Labels:    workerPodLabels(),
		},
		Status: corev1.PodStatus{Phase: corev1.PodRunning},
	}
	// Same namespace and name prefix, but another JobSet's pod: the label selector excludes it.
	otherJobSet := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      rank0PodName(testJobName) + "-zzzz9",
			Namespace: testNS,
			Labels: map[string]string{
				jobsetv1alpha2.JobSetNameKey:        "another-jobset",
				jobsetv1alpha2.ReplicatedJobNameKey: workersReplicatedJobName,
			},
			CreationTimestamp: metav1.NewTime(time.Now().Add(time.Minute)),
		},
		Status: corev1.PodStatus{Phase: corev1.PodRunning},
	}
	fakeClient := fake.NewClientBuilder().WithScheme(k8sscheme.Scheme).
		WithObjects(oldFailed, newRunning, otherPod, otherJobSet).Build()

	pCtx := &k8smocks.PluginContext{}
	pCtx.EXPECT().K8sReader().Return(fakeClient)

	js := makeJobSet("", "", false)
	pods := listWorkerPods(context.Background(), pCtx, js)
	names := make([]string, 0, len(pods))
	for _, pod := range pods {
		names = append(names, pod.Name)
	}
	assert.ElementsMatch(t, []string{oldFailed.Name, newRunning.Name, otherPod.Name}, names,
		"only this JobSet's worker pods are listed")
}

// failedWorkerPod is a worker pod that failed with the given pod reason (for example
// "Shutdown" for a host-maintenance eviction) or container exit, terminating at finishedAt.
// It belongs to restart round 0.
func failedWorkerPod(name, podReason, containerReason string, finishedAt time.Time) *corev1.Pod {
	labels := workerPodLabels()
	labels[restartAttemptLabel] = "0"
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: testNS, Labels: labels},
		Status: corev1.PodStatus{
			Phase:  corev1.PodFailed,
			Reason: podReason,
			ContainerStatuses: []corev1.ContainerStatus{{
				Name: "primary",
				State: corev1.ContainerState{Terminated: &corev1.ContainerStateTerminated{
					ExitCode:   1,
					Reason:     containerReason,
					FinishedAt: metav1.NewTime(finishedAt),
				}},
			}},
		},
	}
}

func runningWorkerPod(name, restartRound string) *corev1.Pod {
	labels := workerPodLabels()
	labels[restartAttemptLabel] = restartRound
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: testNS, Labels: labels},
		Spec:       corev1.PodSpec{NodeName: testNodeName},
		Status: corev1.PodStatus{
			Phase:      corev1.PodRunning,
			Conditions: []corev1.PodCondition{{Type: corev1.PodReady, Status: corev1.ConditionTrue}},
		},
	}
}

func TestGetTaskPhase_MaintenanceRetry_NonRank0Worker(t *testing.T) {
	// The drained node held rank 1, not rank 0: still a free system retry.
	js := makeJobSet(jobsetv1alpha2.JobSetFailed, metav1.ConditionTrue, false)
	now := time.Now()
	rank0 := runningWorkerPod(rank0PodName(testJobName)+"-abc12", "0")
	rank1 := failedWorkerPod(testJobName+"-workers-0-1-def34", "Shutdown", "", now)
	reader := fake.NewClientBuilder().WithScheme(k8sscheme.Scheme).WithObjects(rank0, rank1).Build()

	spec := &clusteredpb.ClusteredTaskSpec{
		Replicas:      2,
		NprocPerNode:  1,
		FailurePolicy: &clusteredpb.ClusterFailurePolicy{RestartOnHostMaintenance: true},
	}
	pCtx := dummyPluginCtx(buildTaskTemplate(spec), reader)
	phase, err := clusteredResourceHandler{}.GetTaskPhase(context.Background(), pCtx, js)
	require.NoError(t, err)
	assert.Equal(t, pluginsCore.PhaseRetryableFailure, phase.Phase())
	require.NotNil(t, phase.Err())
	assert.Equal(t, "HostMaintenance", phase.Err().GetCode())
	assert.Equal(t, core.ExecutionError_SYSTEM, phase.Err().GetKind())
	assert.Contains(t, phase.Err().GetMessage(), rank1.Name)
}

func TestGetTaskPhase_MaintenanceRetry_TeardownAfterUserCrash_NotMaintenance(t *testing.T) {
	// Rank 1 crashed (exit 1); the Job controller then tore rank 0 down, which exits
	// with SIGKILL and on its own looks like a system failure. The earliest failure is
	// the cause, so this is the user's crash, not host maintenance.
	js := makeJobSet(jobsetv1alpha2.JobSetFailed, metav1.ConditionTrue, false)
	js.Annotations = map[string]string{primaryContainerAnnotation: "primary"}
	now := time.Now()
	rank1 := failedWorkerPod(testJobName+"-workers-0-1-def34", "", "Error", now)
	rank0 := failedWorkerPod(rank0PodName(testJobName)+"-abc12", "", "Error", now.Add(5*time.Second))
	rank0.Status.ContainerStatuses[0].State.Terminated.ExitCode = 137
	reader := fake.NewClientBuilder().WithScheme(k8sscheme.Scheme).WithObjects(rank0, rank1).Build()

	spec := &clusteredpb.ClusteredTaskSpec{
		Replicas:      2,
		NprocPerNode:  1,
		FailurePolicy: &clusteredpb.ClusterFailurePolicy{RestartOnHostMaintenance: true},
	}
	pCtx := dummyPluginCtx(buildTaskTemplate(spec), reader)
	phase, err := clusteredResourceHandler{}.GetTaskPhase(context.Background(), pCtx, js)
	require.NoError(t, err)
	assert.Equal(t, pluginsCore.PhaseRetryableFailure, phase.Phase())
	require.NotNil(t, phase.Err())
	assert.Equal(t, core.ExecutionError_USER, phase.Err().GetKind())
	assert.NotEqual(t, "HostMaintenance", phase.Err().GetCode())
}

func TestGetTaskPhase_MaintenanceRetry_EvictionBeforeTeardown(t *testing.T) {
	// Rank 1 was killed with its node: the kubelet recorded no container status, only
	// the Failed phase, the Shutdown reason and a DisruptionTarget condition. Rank 0 was
	// torn down afterwards. The eviction is the earliest failure, so it is maintenance.
	js := makeJobSet(jobsetv1alpha2.JobSetFailed, metav1.ConditionTrue, false)
	js.Annotations = map[string]string{primaryContainerAnnotation: "primary"}
	now := time.Now()
	labels := workerPodLabels()
	labels[restartAttemptLabel] = "0"
	rank1 := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:              testJobName + "-workers-0-1-def34",
			Namespace:         testNS,
			Labels:            labels,
			CreationTimestamp: metav1.NewTime(now.Add(-time.Hour)),
		},
		Status: corev1.PodStatus{
			Phase:  corev1.PodFailed,
			Reason: "Shutdown",
			Conditions: []corev1.PodCondition{{
				Type:               corev1.DisruptionTarget,
				Status:             corev1.ConditionTrue,
				Reason:             "TerminationByKubelet",
				LastTransitionTime: metav1.NewTime(now),
			}},
		},
	}
	rank0 := failedWorkerPod(rank0PodName(testJobName)+"-abc12", "", "Error", now.Add(5*time.Second))
	rank0.Status.ContainerStatuses[0].State.Terminated.ExitCode = 137
	reader := fake.NewClientBuilder().WithScheme(k8sscheme.Scheme).WithObjects(rank0, rank1).Build()

	spec := &clusteredpb.ClusteredTaskSpec{
		Replicas:      2,
		NprocPerNode:  1,
		FailurePolicy: &clusteredpb.ClusterFailurePolicy{RestartOnHostMaintenance: true},
	}
	pCtx := dummyPluginCtx(buildTaskTemplate(spec), reader)
	phase, err := clusteredResourceHandler{}.GetTaskPhase(context.Background(), pCtx, js)
	require.NoError(t, err)
	require.NotNil(t, phase.Err())
	assert.Equal(t, "HostMaintenance", phase.Err().GetCode())
	assert.Equal(t, core.ExecutionError_SYSTEM, phase.Err().GetKind())
	assert.Contains(t, phase.Err().GetMessage(), rank1.Name)
}

func TestFirstFailedPod(t *testing.T) {
	now := time.Now()
	terminated := func(name string, at time.Time) corev1.Pod {
		return *failedWorkerPod(name, "", "Error", at)
	}
	conditionOnly := func(name string, at time.Time) corev1.Pod {
		return corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{Name: name, CreationTimestamp: metav1.NewTime(now.Add(-time.Hour))},
			Status: corev1.PodStatus{
				Phase: corev1.PodFailed,
				Conditions: []corev1.PodCondition{
					{
						Type:               corev1.PodReady,
						Status:             corev1.ConditionFalse,
						LastTransitionTime: metav1.NewTime(at.Add(-time.Minute)),
					},
					{
						Type:               corev1.DisruptionTarget,
						Status:             corev1.ConditionTrue,
						LastTransitionTime: metav1.NewTime(at),
					},
				},
			},
		}
	}
	bare := func(name string, created time.Time) corev1.Pod {
		return corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{Name: name, CreationTimestamp: metav1.NewTime(created)},
			Status:     corev1.PodStatus{Phase: corev1.PodFailed},
		}
	}
	running := corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: "running"},
		Status:     corev1.PodStatus{Phase: corev1.PodRunning},
	}

	tests := []struct {
		name string
		pods []corev1.Pod
		want string
	}{
		{name: "none failed", pods: []corev1.Pod{running}, want: ""},
		{
			name: "earliest container termination wins",
			pods: []corev1.Pod{terminated("b", now), terminated("a", now.Add(-time.Second)), running},
			want: "a",
		},
		{
			name: "no container status falls back to the last condition change",
			pods: []corev1.Pod{terminated("b", now), conditionOnly("a", now.Add(-time.Second))},
			want: "a",
		},
		{
			name: "no status at all falls back to creation time",
			pods: []corev1.Pod{terminated("b", now), bare("a", now.Add(-time.Second))},
			want: "a",
		},
		{
			name: "equal times break ties by name",
			pods: []corev1.Pod{terminated("b", now), terminated("a", now)},
			want: "a",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := firstFailedPod(tt.pods)
			if tt.want == "" {
				assert.Nil(t, got)
				return
			}
			require.NotNil(t, got)
			assert.Equal(t, tt.want, got.Name)
		})
	}
}

func TestGetTaskPhase_BudgetExhausted_NonRank0FailureFastFails(t *testing.T) {
	// The gang was running, rank 1 crashed with no restarts left, and the JobSet
	// controller has not written Failed yet: rank 1's failure is surfaced.
	js := makeJobSet("", "", false)
	js.Spec.FailurePolicy = &jobsetv1alpha2.FailurePolicy{MaxRestarts: 0}
	js.Status.ReplicatedJobsStatus = []jobsetv1alpha2.ReplicatedJobStatus{{Name: workersReplicatedJobName, Failed: 1}}
	rank0 := runningWorkerPod(rank0PodName(testJobName)+"-abc12", "0")
	rank1 := failedWorkerPod(testJobName+"-workers-0-1-def34", "", "Error", time.Now())
	reader := fake.NewClientBuilder().WithScheme(k8sscheme.Scheme).WithObjects(rank0, rank1).Build()
	pCtx := dummyPluginCtxWithState(twoNodeSpecWithRestarts(0), reader,
		plugink8s.PluginState{Phase: pluginsCore.PhaseRunning, PhaseVersion: 1}, nil)

	phase, err := clusteredResourceHandler{}.GetTaskPhase(context.Background(), pCtx, js)
	require.NoError(t, err)
	assert.True(t, phase.Phase().IsFailure(), "got %s", phase.Phase())
}

func TestGetTaskPhase_BudgetExhausted_EarliestFailureWins(t *testing.T) {
	// Rank 1 failed first (out of memory); rank 0 failed afterwards when the gang was torn
	// down. The root cause is the earliest failure.
	js := makeJobSet("", "", false)
	js.Spec.FailurePolicy = &jobsetv1alpha2.FailurePolicy{MaxRestarts: 0}
	js.Status.ReplicatedJobsStatus = []jobsetv1alpha2.ReplicatedJobStatus{{Name: workersReplicatedJobName, Failed: 2}}
	now := time.Now()
	rank0 := failedWorkerPod(rank0PodName(testJobName)+"-abc12", "", "Error", now)
	rank1 := failedWorkerPod(testJobName+"-workers-0-1-def34", "", "OOMKilled", now.Add(-10*time.Second))
	reader := fake.NewClientBuilder().WithScheme(k8sscheme.Scheme).WithObjects(rank0, rank1).Build()
	pCtx := dummyPluginCtxWithState(twoNodeSpecWithRestarts(0), reader,
		plugink8s.PluginState{Phase: pluginsCore.PhaseRunning, PhaseVersion: 1}, nil)

	phase, err := clusteredResourceHandler{}.GetTaskPhase(context.Background(), pCtx, js)
	require.NoError(t, err)
	require.True(t, phase.Phase().IsFailure(), "got %s", phase.Phase())
	assert.Equal(t, "OOMKilled", phase.Err().GetCode())
}

func TestGetTaskPhase_PreviousRestartRoundFailureIgnored(t *testing.T) {
	// A failed pod from restart round 0 lingers while round 1 runs. With the restart
	// budget counted as used it must not be mistaken for a current failure.
	js := makeJobSet("", "", false)
	js.Spec.FailurePolicy = &jobsetv1alpha2.FailurePolicy{MaxRestarts: 1}
	js.Status.Restarts = 1
	js.Status.RestartsCountTowardsMax = 1
	js.Status.ReplicatedJobsStatus = []jobsetv1alpha2.ReplicatedJobStatus{
		{Name: workersReplicatedJobName, Failed: 1, Ready: 1},
	}
	stale := failedWorkerPod(testJobName+"-workers-0-1-old11", "", "Error", time.Now().Add(-time.Minute))
	rank0 := runningWorkerPod(rank0PodName(testJobName)+"-new22", "1")
	rank1 := runningWorkerPod(testJobName+"-workers-0-1-new33", "1")
	reader := fake.NewClientBuilder().WithScheme(k8sscheme.Scheme).WithObjects(stale, rank0, rank1).Build()

	phase, err := clusteredResourceHandler{}.GetTaskPhase(context.Background(),
		dummyPluginCtx(twoNodeSpecWithRestarts(1), reader), js)
	require.NoError(t, err)
	assert.Equal(t, pluginsCore.PhaseRunning, phase.Phase())
}

func TestGetTaskPhase_Started_NonRank0PendingFailure_FastFails(t *testing.T) {
	// After a restart, a non-rank-0 worker that can never start fails the task.
	js := makeJobSet("", "", false)
	js.Status.Restarts = 1
	rank0 := runningWorkerPod(rank0PodName(testJobName)+"-abc12", "1")
	rank1 := runningWorkerPod(testJobName+"-workers-0-1-def34", "1")
	rank1.Status = imagePullBackOffStatus()
	reader := fake.NewClientBuilder().WithScheme(k8sscheme.Scheme).WithObjects(rank0, rank1).Build()

	phase, err := clusteredResourceHandler{}.GetTaskPhase(context.Background(),
		dummyPluginCtx(twoNodeSpecWithRestarts(1), reader), js)
	require.NoError(t, err)
	assert.True(t, phase.Phase().IsFailure(), "got %s", phase.Phase())
}

func twoNodeSpecWithRestarts(maxRestarts int32) *core.TaskTemplate {
	return buildTaskTemplate(&clusteredpb.ClusteredTaskSpec{
		Replicas:      2,
		NprocPerNode:  1,
		FailurePolicy: &clusteredpb.ClusterFailurePolicy{MaxRestarts: maxRestarts},
	})
}

func TestGetTaskPhase_FastFail_FailedWithBudgetRemainingReturnsRunning(t *testing.T) {
	js := makeJobSet("", "", false)
	js.Spec.FailurePolicy = &jobsetv1alpha2.FailurePolicy{MaxRestarts: 2}
	js.Status.Restarts = 1
	js.Status.RestartsCountTowardsMax = 1
	js.Status.ReplicatedJobsStatus = []jobsetv1alpha2.ReplicatedJobStatus{
		{Name: workersReplicatedJobName, Failed: 1, Active: 1},
	}
	js.Status.Conditions = []metav1.Condition{
		{Type: "SomeActiveCondition", Status: metav1.ConditionTrue, LastTransitionTime: metav1.NewTime(time.Now())},
	}

	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      rank0PodName(testJobName) + "-abc12",
			Namespace: testNS,
			Labels:    workerPodLabels(),
		},
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
	// Ordinary (non-maintenance) restarts are charged: the controller bumps both counters.
	js.Status.RestartsCountTowardsMax = 1
	js.Status.ReplicatedJobsStatus = []jobsetv1alpha2.ReplicatedJobStatus{
		{Name: workersReplicatedJobName, Failed: 1, Active: 1},
	}
	js.Status.Conditions = []metav1.Condition{
		{Type: "SomeActiveCondition", Status: metav1.ConditionTrue, LastTransitionTime: metav1.NewTime(time.Now())},
	}

	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      rank0PodName(testJobName) + "-abc12",
			Namespace: testNS,
			Labels:    workerPodLabels(),
		},
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
		ObjectMeta: metav1.ObjectMeta{
			Name:      rank0PodName(testJobName) + "-abc12",
			Namespace: testNS,
			Labels:    workerPodLabels(),
		},
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
		ObjectMeta: metav1.ObjectMeta{
			Name:      rank0PodName(testJobName) + "-abc12",
			Namespace: testNS,
			Labels:    workerPodLabels(),
		},
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
		{Type: string(jobsetv1alpha2.JobSetRestarting), Status: metav1.ConditionTrue, LastTransitionTime: metav1.NewTime(time.Now())},
	}

	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      rank0PodName(testJobName) + "-abc12",
			Namespace: testNS,
			Labels:    workerPodLabels(),
		},
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
				Labels:    workerPodLabels(),
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

	// jobSet annotates the authoritative primary container name at build time. Ready=1
	// on the workers Job marks the gang as fully up, which is what makes it Running.
	makeRunningJobSet := func() *jobsetv1alpha2.JobSet {
		js := makeJobSet("", "", false)
		js.Annotations = map[string]string{primaryContainerAnnotation: primaryContainer}
		js.Status.ReplicatedJobsStatus = []jobsetv1alpha2.ReplicatedJobStatus{
			{Name: workersReplicatedJobName, Ready: 1, Active: 1},
		}
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

// --- naming tests ---

// longestPodName reproduces the worst-case pod name JobSet's admission webhook validates:
// "<jobSetName>-<replicatedJob>-<jobIdx>-<podIdx>-<5-char random suffix>" (jobIdx is always
// 0 here since the single ReplicatedJob has Replicas=1; podIdx maxes at replicas-1).
func longestPodName(jobSetName string, replicas int32) string {
	maxPodIdx := strconv.Itoa(int(replicas - 1))
	return placement.GenPodName(jobSetName, workersReplicatedJobName, "0", maxPodIdx) + "-abcde"
}

// managerStampedName mirrors the name the plugin manager stamps on both the create and
// lookup paths: GetGeneratedNameWith(0, GeneratedNameMaxLength) as implemented by the
// executor (bound with a hash when over the max length).
func managerStampedName(t *testing.T, generatedName string) string {
	name, err := encoding.FixedLengthUniqueID(generatedName, generatedNameMaxLength)
	assert.NoError(t, err)
	return name
}

// TestGeneratedNameMaxLength_BoundsPodNames guards the advertised bound: every
// manager-stamped name within GeneratedNameMaxLength must keep the worst-case derived
// pod name a valid DNS-1035 label within the 63-char limit, across every supported
// replica count (the bound is independent of the replica count so create and lookup
// agree without a task template).
func TestGeneratedNameMaxLength_BoundsPodNames(t *testing.T) {
	props := clusteredResourceHandler{}.GetProperties()
	if assert.NotNil(t, props.GeneratedNameMaxLength) {
		assert.Equal(t, generatedNameMaxLength, *props.GeneratedNameMaxLength)
	}

	// Run names are validated as DNS-1035 labels at creation, so generated names are
	// always label-compatible; only their length varies.
	for _, generated := range []string{
		"f-abc123", // short
		"g" + strings.Repeat("a", generatedNameMaxLength-1), // exactly at the bound
		strings.Repeat("composed-subtask-", 8) + "tail-0",   // long composed/nested name
	} {
		name := managerStampedName(t, generated)
		assert.LessOrEqual(t, len(name), generatedNameMaxLength)
		assert.Empty(t, validation.IsDNS1035Label(name), "stamped name %q (from %q) is not a valid DNS-1035 label", name, generated)
		for _, replicas := range []int32{1, 128, 10000, maxReplicasForNaming} {
			podName := longestPodName(name, replicas)
			assert.LessOrEqual(t, len(podName), dns1035LabelMaxLength, "pod name %q (%d chars) exceeds limit for replicas=%d", podName, len(podName), replicas)
			assert.Empty(t, validation.IsDNS1035Label(podName), "pod name %q invalid for replicas=%d", podName, replicas)
		}
	}
}

// TestBuildResource_NameMatchesManagerStamp guards the create/lookup invariant under the
// GeneratedNameMaxLength mechanism: the manager stamps the same name on the built object
// and the identity object, and BuildResource must derive that identical name for the pod
// subdomain (the headless service and JobSet name must match for pod DNS to resolve).
// BuildIdentityResource leaves the name empty — the manager owns naming.
func TestBuildResource_NameMatchesManagerStamp(t *testing.T) {
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
	taskCtx := dummyTaskCtxWithGeneratedName(buildTaskTemplate(spec), longGeneratedName, nil)
	handler := clusteredResourceHandler{}

	created, err := handler.BuildResource(context.Background(), taskCtx)
	assert.NoError(t, err)

	expected := managerStampedName(t, longGeneratedName)
	assert.NotEqual(t, longGeneratedName, expected, "long name should have been truncated")

	jobSet, ok := created.(*jobsetv1alpha2.JobSet)
	assert.True(t, ok, "expected *JobSet")
	assert.Equal(t, expected, jobSet.Name)
	assert.Equal(t, expected, jobSet.Spec.ReplicatedJobs[0].Template.Spec.Template.Spec.Subdomain)

	identity, err := handler.BuildIdentityResource(context.Background(), taskCtx.TaskExecutionMetadata())
	assert.NoError(t, err)
	assert.Empty(t, identity.GetName(), "identity name is stamped by the plugin manager, not the plugin")
}

// TestBuildResource_ReplicasExceedNamingBudget verifies BuildResource fails fast with a
// spec error when replicas exceeds the pod-index budget generatedNameMaxLength reserves
// for; past that bound the derived pod names could exceed 63 chars and be rejected by the
// webhook.
func TestBuildResource_ReplicasExceedNamingBudget(t *testing.T) {
	spec := &clusteredpb.ClusteredTaskSpec{
		Replicas:     maxReplicasForNaming + 1,
		NprocPerNode: 1,
		Runtime: &clusteredpb.Runtime{
			Kind: &clusteredpb.Runtime_Torchrun{Torchrun: &clusteredpb.TorchRuntime{}},
		},
	}
	taskCtx := dummyTaskCtx(buildTaskTemplate(spec), nil)
	handler := clusteredResourceHandler{}

	_, err := handler.BuildResource(context.Background(), taskCtx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "replicas must be <=")
}

// TestSchemeRegistration guards the contract every k8s plugin owes its host binary: the
// CRD it watches must be reachable through PluginRegistry().GetSchemeRegisters(), which is
// how the executor builds its scheme (executor/setup.go). Hosts that drain the registry
// otherwise fail at Create with "no kind is registered for the type ... in scheme".
func TestSchemeRegistration(t *testing.T) {
	s := runtime.NewScheme()
	found := false
	for _, reg := range pluginmachinery.PluginRegistry().GetSchemeRegisters() {
		if reg.ID == taskType {
			found = true
			require.NoError(t, reg.AddToScheme(s))
		}
	}
	require.True(t, found, "clustered-task did not register an AddToScheme")

	gvks, _, err := s.ObjectKinds(&jobsetv1alpha2.JobSet{})
	require.NoError(t, err)
	assert.NotEmpty(t, gvks)
}
