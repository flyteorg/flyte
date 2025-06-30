package mpi

import (
	"context"
	"fmt"
	"reflect"
	"testing"
	"time"

	structpb "github.com/golang/protobuf/ptypes/struct"
	kubeflowv1 "github.com/kubeflow/training-operator/pkg/apis/kubeflow.org/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/plugins"
	kfplugins "github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/plugins/kubeflow"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/logs"
	pluginsCore "github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/core/mocks"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/flytek8s"
	flytek8sConfig "github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/flytek8s/config"
	pluginIOMocks "github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/io/mocks"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/k8s"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/utils"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/plugins/k8s/kfoperators/common"
	stdlibUtils "github.com/flyteorg/flyte/flytestdlib/utils"
)

const testImage = "image://"
const serviceAccount = "mpi_sa"
const mpiID = "the job 1"
const mpiID2 = "the job 2"

var (
	dummyEnvVars = []*core.KeyValuePair{
		{Key: "Env_Var", Value: "Env_Val"},
	}

	testArgs = []string{
		"test-args",
	}

	dummyAnnotations = map[string]string{
		"annotation-key": "annotation-value",
	}
	dummyLabels = map[string]string{
		"label-key": "label-value",
	}

	resourceRequirements = &corev1.ResourceRequirements{
		Limits: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse("1000m"),
			corev1.ResourceMemory: resource.MustParse("1Gi"),
		},
		Requests: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse("100m"),
			corev1.ResourceMemory: resource.MustParse("512Mi"),
		},
	}

	jobName      = "the-job"
	jobNamespace = "mpi-namespace"
)

func dummyMPICustomObj(workers int32, launcher int32, slots int32) *plugins.DistributedMPITrainingTask {
	return &plugins.DistributedMPITrainingTask{
		NumWorkers:          workers,
		NumLauncherReplicas: launcher,
		Slots:               slots,
	}
}

func dummyMPITaskTemplate(id string, args ...interface{}) *core.TaskTemplate {

	var mpiObjJSON string
	var err error

	for _, arg := range args {
		switch t := arg.(type) {
		case *kfplugins.DistributedMPITrainingTask:
			var mpiCustomObj = t
			mpiObjJSON, err = utils.MarshalToString(mpiCustomObj)
		case *plugins.DistributedMPITrainingTask:
			var mpiCustomObj = t
			mpiObjJSON, err = utils.MarshalToString(mpiCustomObj)
		default:
			err = fmt.Errorf("Unknown input type %T", t)
		}
	}

	if err != nil {
		panic(err)
	}

	structObj := structpb.Struct{}

	err = stdlibUtils.UnmarshalStringToPb(mpiObjJSON, &structObj)
	if err != nil {
		panic(err)
	}

	return &core.TaskTemplate{
		Id:   &core.Identifier{Name: id},
		Type: "container",
		Target: &core.TaskTemplate_Container{
			Container: &core.Container{
				Image: testImage,
				Args:  testArgs,
				Env:   dummyEnvVars,
			},
		},
		Custom: &structObj,
	}
}

func dummyMPITaskContext(taskTemplate *core.TaskTemplate, resources *corev1.ResourceRequirements, extendedResources *core.ExtendedResources, pluginState k8s.PluginState) pluginsCore.TaskExecutionContext {
	taskCtx := &mocks.TaskExecutionContext{}
	inputReader := &pluginIOMocks.InputReader{}
	inputReader.EXPECT().GetInputPrefixPath().Return("/input/prefix")
	inputReader.EXPECT().GetInputPath().Return("/input")
	inputReader.EXPECT().Get(mock.Anything).Return(&core.LiteralMap{}, nil)
	taskCtx.EXPECT().InputReader().Return(inputReader)

	outputReader := &pluginIOMocks.OutputWriter{}
	outputReader.EXPECT().GetOutputPath().Return("/data/outputs.pb")
	outputReader.EXPECT().GetOutputPrefixPath().Return("/data/")
	outputReader.EXPECT().GetRawOutputPrefix().Return("")
	outputReader.EXPECT().GetCheckpointPrefix().Return("/checkpoint")
	outputReader.EXPECT().GetPreviousCheckpointsPrefix().Return("/prev")
	taskCtx.EXPECT().OutputWriter().Return(outputReader)

	taskReader := &mocks.TaskReader{}
	taskReader.EXPECT().Read(mock.Anything).Return(taskTemplate, nil)
	taskCtx.EXPECT().TaskReader().Return(taskReader)

	tID := &mocks.TaskExecutionID{}
	tID.EXPECT().GetID().Return(core.TaskExecutionIdentifier{
		NodeExecutionId: &core.NodeExecutionIdentifier{
			ExecutionId: &core.WorkflowExecutionIdentifier{
				Name:    "my_name",
				Project: "my_project",
				Domain:  "my_domain",
			},
		},
	})
	tID.EXPECT().GetGeneratedName().Return("some-acceptable-name")
	tID.On("GetUniqueNodeID").Return("an-unique-id")

	overrides := &mocks.TaskOverrides{}
	overrides.EXPECT().GetResources().Return(resources)
	overrides.EXPECT().GetExtendedResources().Return(extendedResources)
	overrides.EXPECT().GetContainerImage().Return("")
	overrides.EXPECT().GetPodTemplate().Return(nil)

	taskExecutionMetadata := &mocks.TaskExecutionMetadata{}
	taskExecutionMetadata.EXPECT().GetTaskExecutionID().Return(tID)
	taskExecutionMetadata.EXPECT().GetNamespace().Return("test-namespace")
	taskExecutionMetadata.EXPECT().GetAnnotations().Return(dummyAnnotations)
	taskExecutionMetadata.EXPECT().GetLabels().Return(dummyLabels)
	taskExecutionMetadata.EXPECT().GetOwnerReference().Return(v1.OwnerReference{
		Kind: "node",
		Name: "blah",
	})
	taskExecutionMetadata.EXPECT().IsInterruptible().Return(true)
	taskExecutionMetadata.EXPECT().GetOverrides().Return(overrides)
	taskExecutionMetadata.EXPECT().GetK8sServiceAccount().Return(serviceAccount)
	taskExecutionMetadata.EXPECT().GetPlatformResources().Return(&corev1.ResourceRequirements{})
	taskExecutionMetadata.EXPECT().GetEnvironmentVariables().Return(nil)
	taskExecutionMetadata.EXPECT().GetConsoleURL().Return("")
	taskExecutionMetadata.EXPECT().GetOnOOMConfig().Return(nil)
	taskCtx.EXPECT().TaskExecutionMetadata().Return(taskExecutionMetadata)

	pluginStateReaderMock := mocks.PluginStateReader{}
	pluginStateReaderMock.On("Get", mock.AnythingOfType(reflect.TypeOf(&pluginState).String())).Return(
		func(v interface{}) uint8 {
			*(v.(*k8s.PluginState)) = pluginState
			return 0
		},
		func(v interface{}) error {
			return nil
		})

	taskCtx.EXPECT().PluginStateReader().Return(&pluginStateReaderMock)
	return taskCtx
}

func dummyMPIJobResource(mpiResourceHandler mpiOperatorResourceHandler,
	workers int32, launcher int32, slots int32, conditionType kubeflowv1.JobConditionType) *kubeflowv1.MPIJob {
	var jobConditions []kubeflowv1.JobCondition

	now := time.Now()

	jobCreated := kubeflowv1.JobCondition{
		Type:    kubeflowv1.JobCreated,
		Status:  corev1.ConditionTrue,
		Reason:  "MPICreated",
		Message: "MPIJob the-job is created.",
		LastUpdateTime: v1.Time{
			Time: now,
		},
		LastTransitionTime: v1.Time{
			Time: now,
		},
	}
	jobRunningActive := kubeflowv1.JobCondition{
		Type:    kubeflowv1.JobRunning,
		Status:  corev1.ConditionTrue,
		Reason:  "MPIJobRunning",
		Message: "MPIJob the-job is running.",
		LastUpdateTime: v1.Time{
			Time: now.Add(time.Minute),
		},
		LastTransitionTime: v1.Time{
			Time: now.Add(time.Minute),
		},
	}
	jobRunningInactive := *jobRunningActive.DeepCopy()
	jobRunningInactive.Status = corev1.ConditionFalse
	jobSucceeded := kubeflowv1.JobCondition{
		Type:    kubeflowv1.JobSucceeded,
		Status:  corev1.ConditionTrue,
		Reason:  "MPIJobSucceeded",
		Message: "MPIJob the-job is successfully completed.",
		LastUpdateTime: v1.Time{
			Time: now.Add(2 * time.Minute),
		},
		LastTransitionTime: v1.Time{
			Time: now.Add(2 * time.Minute),
		},
	}
	jobFailed := kubeflowv1.JobCondition{
		Type:    kubeflowv1.JobFailed,
		Status:  corev1.ConditionTrue,
		Reason:  "MPIJobFailed",
		Message: "MPIJob the-job is failed.",
		LastUpdateTime: v1.Time{
			Time: now.Add(2 * time.Minute),
		},
		LastTransitionTime: v1.Time{
			Time: now.Add(2 * time.Minute),
		},
	}
	jobRestarting := kubeflowv1.JobCondition{
		Type:    kubeflowv1.JobRestarting,
		Status:  corev1.ConditionTrue,
		Reason:  "MPIJobRestarting",
		Message: "MPIJob the-job is restarting because some replica(s) failed.",
		LastUpdateTime: v1.Time{
			Time: now.Add(3 * time.Minute),
		},
		LastTransitionTime: v1.Time{
			Time: now.Add(3 * time.Minute),
		},
	}

	switch conditionType {
	case kubeflowv1.JobCreated:
		jobConditions = []kubeflowv1.JobCondition{
			jobCreated,
		}
	case kubeflowv1.JobRunning:
		jobConditions = []kubeflowv1.JobCondition{
			jobCreated,
			jobRunningActive,
		}
	case kubeflowv1.JobSucceeded:
		jobConditions = []kubeflowv1.JobCondition{
			jobCreated,
			jobRunningInactive,
			jobSucceeded,
		}
	case kubeflowv1.JobFailed:
		jobConditions = []kubeflowv1.JobCondition{
			jobCreated,
			jobRunningInactive,
			jobFailed,
		}
	case kubeflowv1.JobRestarting:
		jobConditions = []kubeflowv1.JobCondition{
			jobCreated,
			jobRunningInactive,
			jobFailed,
			jobRestarting,
		}
	}

	mpiObj := dummyMPICustomObj(workers, launcher, slots)
	taskTemplate := dummyMPITaskTemplate(mpiID, mpiObj)
	resource, err := mpiResourceHandler.BuildResource(context.TODO(), dummyMPITaskContext(taskTemplate, resourceRequirements, nil, k8s.PluginState{}))
	if err != nil {
		panic(err)
	}

	return &kubeflowv1.MPIJob{
		ObjectMeta: v1.ObjectMeta{
			Name:      jobName,
			Namespace: jobNamespace,
		},
		Spec: resource.(*kubeflowv1.MPIJob).Spec,
		Status: kubeflowv1.JobStatus{
			Conditions:        jobConditions,
			ReplicaStatuses:   nil,
			StartTime:         &v1.Time{Time: time.Now()},
			CompletionTime:    nil,
			LastReconcileTime: nil,
		},
	}
}

func TestBuildResourceMPI(t *testing.T) {
	mpiResourceHandler := mpiOperatorResourceHandler{}

	mpiObj := dummyMPICustomObj(100, 50, 1)
	taskTemplate := dummyMPITaskTemplate(mpiID2, mpiObj)

	resource, err := mpiResourceHandler.BuildResource(context.TODO(), dummyMPITaskContext(taskTemplate, resourceRequirements, nil, k8s.PluginState{}))
	assert.NoError(t, err)
	assert.NotNil(t, resource)

	mpiJob, ok := resource.(*kubeflowv1.MPIJob)
	assert.True(t, ok)
	assert.Equal(t, int32(50), *mpiJob.Spec.MPIReplicaSpecs[kubeflowv1.MPIJobReplicaTypeLauncher].Replicas)
	assert.Equal(t, int32(100), *mpiJob.Spec.MPIReplicaSpecs[kubeflowv1.MPIJobReplicaTypeWorker].Replicas)
	assert.Equal(t, int32(1), *mpiJob.Spec.SlotsPerWorker)

	// verify TaskExecutionMetadata labels and annotations are copied to the MPIJob
	for k, v := range dummyAnnotations {
		for _, replicaSpec := range mpiJob.Spec.MPIReplicaSpecs {
			assert.Equal(t, v, replicaSpec.Template.ObjectMeta.Annotations[k])
		}
	}
	for k, v := range dummyLabels {
		for _, replicaSpec := range mpiJob.Spec.MPIReplicaSpecs {
			assert.Equal(t, v, replicaSpec.Template.ObjectMeta.Labels[k])
		}
	}

	for _, replicaSpec := range mpiJob.Spec.MPIReplicaSpecs {
		for _, container := range replicaSpec.Template.Spec.Containers {
			assert.Equal(t, resourceRequirements.Requests, container.Resources.Requests)
			assert.Equal(t, resourceRequirements.Limits, container.Resources.Limits)
		}
	}
}

func TestBuildResourceMPIForWrongInput(t *testing.T) {
	mpiResourceHandler := mpiOperatorResourceHandler{}

	mpiObj := dummyMPICustomObj(0, 0, 1)
	taskTemplate := dummyMPITaskTemplate(mpiID, mpiObj)

	_, err := mpiResourceHandler.BuildResource(context.TODO(), dummyMPITaskContext(taskTemplate, resourceRequirements, nil, k8s.PluginState{}))
	assert.Error(t, err)

	mpiObj = dummyMPICustomObj(1, 1, 1)
	taskTemplate = dummyMPITaskTemplate(mpiID2, mpiObj)

	resource, err := mpiResourceHandler.BuildResource(context.TODO(), dummyMPITaskContext(taskTemplate, resourceRequirements, nil, k8s.PluginState{}))
	app, ok := resource.(*kubeflowv1.MPIJob)
	assert.Nil(t, err)
	assert.Equal(t, true, ok)
	assert.Equal(t, []string{}, app.Spec.MPIReplicaSpecs[kubeflowv1.MPIJobReplicaTypeWorker].Template.Spec.Containers[0].Command)
	assert.Equal(t, []string{}, app.Spec.MPIReplicaSpecs[kubeflowv1.MPIJobReplicaTypeWorker].Template.Spec.Containers[0].Args)
}

func TestBuildResourceMPIExtendedResources(t *testing.T) {
	assert.NoError(t, flytek8sConfig.SetK8sPluginConfig(&flytek8sConfig.K8sPluginConfig{
		GpuDeviceNodeLabel:                 "gpu-node-label",
		GpuPartitionSizeNodeLabel:          "gpu-partition-size",
		GpuResourceName:                    flytek8s.ResourceNvidiaGPU,
		AddTolerationsForExtendedResources: []string{"nvidia.com/gpu"},
	}))

	fixtures := []struct {
		name                      string
		resources                 *corev1.ResourceRequirements
		extendedResourcesBase     *core.ExtendedResources
		extendedResourcesOverride *core.ExtendedResources
		expectedNsr               []corev1.NodeSelectorTerm
		expectedTol               []corev1.Toleration
	}{
		{
			"without overrides",
			&corev1.ResourceRequirements{
				Limits: corev1.ResourceList{
					flytek8s.ResourceNvidiaGPU: resource.MustParse("1"),
				},
			},
			&core.ExtendedResources{
				GpuAccelerator: &core.GPUAccelerator{
					Device: "nvidia-tesla-t4",
				},
			},
			nil,
			[]corev1.NodeSelectorTerm{
				{
					MatchExpressions: []corev1.NodeSelectorRequirement{
						corev1.NodeSelectorRequirement{
							Key:      "gpu-node-label",
							Operator: corev1.NodeSelectorOpIn,
							Values:   []string{"nvidia-tesla-t4"},
						},
					},
				},
			},
			[]corev1.Toleration{
				{
					Key:      "gpu-node-label",
					Value:    "nvidia-tesla-t4",
					Operator: corev1.TolerationOpEqual,
					Effect:   corev1.TaintEffectNoSchedule,
				},
				{
					Key:      "nvidia.com/gpu",
					Operator: corev1.TolerationOpExists,
					Effect:   corev1.TaintEffectNoSchedule,
				},
			},
		},
		{
			"with overrides",
			&corev1.ResourceRequirements{
				Limits: corev1.ResourceList{
					flytek8s.ResourceNvidiaGPU: resource.MustParse("1"),
				},
			},
			&core.ExtendedResources{
				GpuAccelerator: &core.GPUAccelerator{
					Device: "nvidia-tesla-t4",
				},
			},
			&core.ExtendedResources{
				GpuAccelerator: &core.GPUAccelerator{
					Device: "nvidia-tesla-a100",
					PartitionSizeValue: &core.GPUAccelerator_PartitionSize{
						PartitionSize: "1g.5gb",
					},
				},
			},
			[]corev1.NodeSelectorTerm{
				{
					MatchExpressions: []corev1.NodeSelectorRequirement{
						corev1.NodeSelectorRequirement{
							Key:      "gpu-node-label",
							Operator: corev1.NodeSelectorOpIn,
							Values:   []string{"nvidia-tesla-a100"},
						},
						corev1.NodeSelectorRequirement{
							Key:      "gpu-partition-size",
							Operator: corev1.NodeSelectorOpIn,
							Values:   []string{"1g.5gb"},
						},
					},
				},
			},
			[]corev1.Toleration{
				{
					Key:      "gpu-node-label",
					Value:    "nvidia-tesla-a100",
					Operator: corev1.TolerationOpEqual,
					Effect:   corev1.TaintEffectNoSchedule,
				},
				{
					Key:      "gpu-partition-size",
					Value:    "1g.5gb",
					Operator: corev1.TolerationOpEqual,
					Effect:   corev1.TaintEffectNoSchedule,
				},
				{
					Key:      "nvidia.com/gpu",
					Operator: corev1.TolerationOpExists,
					Effect:   corev1.TaintEffectNoSchedule,
				},
			},
		},
	}

	for _, f := range fixtures {
		t.Run(f.name, func(t *testing.T) {
			mpiObj := dummyMPICustomObj(100, 50, 1)
			taskTemplate := dummyMPITaskTemplate(mpiID2, mpiObj)
			taskTemplate.ExtendedResources = f.extendedResourcesBase
			taskContext := dummyMPITaskContext(taskTemplate, f.resources, f.extendedResourcesOverride, k8s.PluginState{})
			mpiResourceHandler := mpiOperatorResourceHandler{}
			r, err := mpiResourceHandler.BuildResource(context.TODO(), taskContext)
			assert.Nil(t, err)
			assert.NotNil(t, r)
			mpiJob, ok := r.(*kubeflowv1.MPIJob)
			assert.True(t, ok)

			for _, replicaSpec := range mpiJob.Spec.MPIReplicaSpecs {
				assert.EqualValues(
					t,
					f.expectedNsr,
					replicaSpec.Template.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms,
				)
				assert.EqualValues(
					t,
					f.expectedTol,
					replicaSpec.Template.Spec.Tolerations,
				)
			}
		})
	}
}

func TestGetTaskPhase(t *testing.T) {
	mpiResourceHandler := mpiOperatorResourceHandler{}
	ctx := context.TODO()

	dummyMPIJobResourceCreator := func(conditionType kubeflowv1.JobConditionType) *kubeflowv1.MPIJob {
		return dummyMPIJobResource(mpiResourceHandler, 2, 1, 1, conditionType)
	}

	taskCtx := dummyMPITaskContext(dummyMPITaskTemplate("", dummyMPICustomObj(2, 1, 1)), resourceRequirements, nil, k8s.PluginState{})
	taskPhase, err := mpiResourceHandler.GetTaskPhase(ctx, taskCtx, dummyMPIJobResourceCreator(kubeflowv1.JobCreated))
	assert.NoError(t, err)
	assert.Equal(t, pluginsCore.PhaseQueued, taskPhase.Phase())
	assert.NotNil(t, taskPhase.Info())
	assert.Nil(t, err)

	taskPhase, err = mpiResourceHandler.GetTaskPhase(ctx, taskCtx, dummyMPIJobResourceCreator(kubeflowv1.JobRunning))
	assert.NoError(t, err)
	assert.Equal(t, pluginsCore.PhaseRunning, taskPhase.Phase())
	assert.NotNil(t, taskPhase.Info())
	assert.Nil(t, err)

	taskPhase, err = mpiResourceHandler.GetTaskPhase(ctx, taskCtx, dummyMPIJobResourceCreator(kubeflowv1.JobSucceeded))
	assert.NoError(t, err)
	assert.Equal(t, pluginsCore.PhaseSuccess, taskPhase.Phase())
	assert.NotNil(t, taskPhase.Info())
	assert.Nil(t, err)

	taskPhase, err = mpiResourceHandler.GetTaskPhase(ctx, taskCtx, dummyMPIJobResourceCreator(kubeflowv1.JobFailed))
	assert.NoError(t, err)
	assert.Equal(t, pluginsCore.PhaseRetryableFailure, taskPhase.Phase())
	assert.NotNil(t, taskPhase.Info())
	assert.Nil(t, err)

	taskPhase, err = mpiResourceHandler.GetTaskPhase(ctx, taskCtx, dummyMPIJobResourceCreator(kubeflowv1.JobRestarting))
	assert.NoError(t, err)
	assert.Equal(t, pluginsCore.PhaseRunning, taskPhase.Phase())
	assert.NotNil(t, taskPhase.Info())
	assert.Nil(t, err)

	// Training operator did not modify the job even though it is not suspended
	mpiJob := dummyMPIJobResourceCreator(kubeflowv1.JobCreated)
	mpiJob.CreationTimestamp = v1.Time{Time: time.Now().Add(-time.Hour)}
	mpiJob.Status.StartTime = nil
	taskPhase, err = mpiResourceHandler.GetTaskPhase(ctx, taskCtx, mpiJob)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "kubeflow operator hasn't updated")
	assert.Equal(t, pluginsCore.PhaseInfoUndefined, taskPhase)

	// Training operator did not modify the job because it is suspended
	mpiJobSuspended := dummyMPIJobResourceCreator(kubeflowv1.JobCreated)
	mpiJobSuspended.CreationTimestamp = v1.Time{Time: time.Now().Add(-time.Hour)}
	mpiJobSuspended.Status.StartTime = nil
	suspend := true
	mpiJobSuspended.Spec.RunPolicy.Suspend = &suspend
	taskPhase, err = mpiResourceHandler.GetTaskPhase(ctx, taskCtx, mpiJobSuspended)
	assert.NoError(t, err)
	assert.Equal(t, pluginsCore.PhaseQueued, taskPhase.Phase())
}

func TestGetTaskPhaseIncreasePhaseVersion(t *testing.T) {
	mpiResourceHandler := mpiOperatorResourceHandler{}
	ctx := context.TODO()

	pluginState := k8s.PluginState{
		Phase:        pluginsCore.PhaseQueued,
		PhaseVersion: pluginsCore.DefaultPhaseVersion,
		Reason:       "task submitted to K8s",
	}
	taskCtx := dummyMPITaskContext(dummyMPITaskTemplate("", dummyMPICustomObj(2, 1, 1)), resourceRequirements, nil, pluginState)

	taskPhase, err := mpiResourceHandler.GetTaskPhase(ctx, taskCtx, dummyMPIJobResource(mpiResourceHandler, 2, 1, 1, kubeflowv1.JobCreated))

	assert.NoError(t, err)
	assert.Equal(t, taskPhase.Version(), pluginsCore.DefaultPhaseVersion+1)
}

func TestGetLogs(t *testing.T) {
	assert.NoError(t, logs.SetLogConfig(&logs.LogConfig{
		IsKubernetesEnabled: true,
		KubernetesURL:       "k8s.com",
	}))

	workers := int32(2)
	launcher := int32(1)
	slots := int32(1)

	mpiResourceHandler := mpiOperatorResourceHandler{}
	mpiJob := dummyMPIJobResource(mpiResourceHandler, workers, launcher, slots, kubeflowv1.JobRunning)
	taskTemplate := dummyMPITaskTemplate("", dummyMPICustomObj(workers, launcher, slots))
	taskCtx := dummyMPITaskContext(taskTemplate, resourceRequirements, nil, k8s.PluginState{})
	jobLogs, err := common.GetLogs(taskCtx, common.MPITaskType, mpiJob.ObjectMeta, taskTemplate, false, workers, launcher, 0, 0)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(jobLogs))
	assert.Equal(t, fmt.Sprintf("k8s.com/#!/log/%s/%s-worker-0/pod?namespace=mpi-namespace", jobNamespace, jobName), jobLogs[0].GetUri())
	assert.Equal(t, fmt.Sprintf("k8s.com/#!/log/%s/%s-worker-1/pod?namespace=mpi-namespace", jobNamespace, jobName), jobLogs[1].GetUri())
}

func TestGetProperties(t *testing.T) {
	mpiResourceHandler := mpiOperatorResourceHandler{}
	expected := k8s.PluginProperties{}
	assert.Equal(t, expected, mpiResourceHandler.GetProperties())
}

func TestReplicaCounts(t *testing.T) {
	for _, test := range []struct {
		name                 string
		launcherReplicaCount int32
		workerReplicaCount   int32
		expectError          bool
		contains             []kubeflowv1.ReplicaType
		notContains          []kubeflowv1.ReplicaType
	}{
		{"NoWorkers", 1, 0, true, nil, nil},
		{"Minimum One Launcher", 0, 1, false, []kubeflowv1.ReplicaType{kubeflowv1.MPIJobReplicaTypeLauncher, kubeflowv1.MPIJobReplicaTypeWorker}, []kubeflowv1.ReplicaType{}},
		{"Works", 1, 1, false, []kubeflowv1.ReplicaType{kubeflowv1.MPIJobReplicaTypeLauncher, kubeflowv1.MPIJobReplicaTypeWorker}, []kubeflowv1.ReplicaType{}},
	} {
		t.Run(test.name, func(t *testing.T) {
			mpiResourceHandler := mpiOperatorResourceHandler{}

			mpiObj := dummyMPICustomObj(test.workerReplicaCount, test.launcherReplicaCount, 1)
			taskTemplate := dummyMPITaskTemplate(mpiID2, mpiObj)

			resource, err := mpiResourceHandler.BuildResource(context.TODO(), dummyMPITaskContext(taskTemplate, resourceRequirements, nil, k8s.PluginState{}))
			if test.expectError {
				assert.Error(t, err)
				assert.Nil(t, resource)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, resource)

			job, ok := resource.(*kubeflowv1.MPIJob)
			assert.True(t, ok)

			assert.Len(t, job.Spec.MPIReplicaSpecs, len(test.contains))
			for _, replicaType := range test.contains {
				assert.Contains(t, job.Spec.MPIReplicaSpecs, replicaType)
			}
			for _, replicaType := range test.notContains {
				assert.NotContains(t, job.Spec.MPIReplicaSpecs, replicaType)
			}
		})
	}
}

func TestBuildResourceMPIV1(t *testing.T) {
	launcherCommand := []string{"python", "launcher.py"}
	workerCommand := []string{"/usr/sbin/sshd", "/.sshd_config"}
	taskConfigs := []*kfplugins.DistributedMPITrainingTask{
		{
			LauncherReplicas: &kfplugins.DistributedMPITrainingReplicaSpec{
				Image: testImage,
				Resources: &core.Resources{
					Requests: []*core.Resources_ResourceEntry{
						{Name: core.Resources_CPU, Value: "250m"},
						{Name: core.Resources_MEMORY, Value: "250Mi"},
					},
					Limits: []*core.Resources_ResourceEntry{
						{Name: core.Resources_CPU, Value: "500m"},
						{Name: core.Resources_MEMORY, Value: "500Mi"},
					},
				},
				Command: launcherCommand,
			},
			WorkerReplicas: &kfplugins.DistributedMPITrainingReplicaSpec{
				Replicas: 100,
				Resources: &core.Resources{
					Requests: []*core.Resources_ResourceEntry{
						{Name: core.Resources_CPU, Value: "1024m"},
						{Name: core.Resources_MEMORY, Value: "1Gi"},
					},
					Limits: []*core.Resources_ResourceEntry{
						{Name: core.Resources_CPU, Value: "2048m"},
						{Name: core.Resources_MEMORY, Value: "2Gi"},
					},
				},
				Command: workerCommand,
			},
			Slots: int32(1),
		},
		{
			LauncherReplicas: &kfplugins.DistributedMPITrainingReplicaSpec{
				Common: &kfplugins.CommonReplicaSpec{
					Image: testImage,
					Resources: &core.Resources{
						Requests: []*core.Resources_ResourceEntry{
							{Name: core.Resources_CPU, Value: "250m"},
							{Name: core.Resources_MEMORY, Value: "250Mi"},
						},
						Limits: []*core.Resources_ResourceEntry{
							{Name: core.Resources_CPU, Value: "500m"},
							{Name: core.Resources_MEMORY, Value: "500Mi"},
						},
					},
				},
				Command: launcherCommand,
			},
			WorkerReplicas: &kfplugins.DistributedMPITrainingReplicaSpec{
				Common: &kfplugins.CommonReplicaSpec{
					Replicas: 100,
					Resources: &core.Resources{
						Requests: []*core.Resources_ResourceEntry{
							{Name: core.Resources_CPU, Value: "1024m"},
							{Name: core.Resources_MEMORY, Value: "1Gi"},
						},
						Limits: []*core.Resources_ResourceEntry{
							{Name: core.Resources_CPU, Value: "2048m"},
							{Name: core.Resources_MEMORY, Value: "2Gi"},
						},
					},
				},
				Command: workerCommand,
			},
			Slots: int32(1),
		},
	}

	for _, taskConfig := range taskConfigs {
		launcherResourceRequirements := &corev1.ResourceRequirements{
			Requests: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("250m"),
				corev1.ResourceMemory: resource.MustParse("250Mi"),
			},
			Limits: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("500m"),
				corev1.ResourceMemory: resource.MustParse("500Mi"),
			},
		}

		workerResourceRequirements := &corev1.ResourceRequirements{
			Requests: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("1024m"),
				corev1.ResourceMemory: resource.MustParse("1Gi"),
			},
			Limits: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("2048m"),
				corev1.ResourceMemory: resource.MustParse("2Gi"),
			},
		}

		mpiResourceHandler := mpiOperatorResourceHandler{}

		taskTemplate := dummyMPITaskTemplate(mpiID2, taskConfig)
		taskTemplate.TaskTypeVersion = 1

		resource, err := mpiResourceHandler.BuildResource(context.TODO(), dummyMPITaskContext(taskTemplate, resourceRequirements, nil, k8s.PluginState{}))
		assert.NoError(t, err)
		assert.NotNil(t, resource)

		mpiJob, ok := resource.(*kubeflowv1.MPIJob)
		assert.True(t, ok)
		assert.Equal(t, int32(1), *mpiJob.Spec.MPIReplicaSpecs[kubeflowv1.MPIJobReplicaTypeLauncher].Replicas)
		assert.Equal(t, int32(100), *mpiJob.Spec.MPIReplicaSpecs[kubeflowv1.MPIJobReplicaTypeWorker].Replicas)
		assert.Equal(t, int32(1), *mpiJob.Spec.SlotsPerWorker)
		assert.Equal(t, *launcherResourceRequirements, mpiJob.Spec.MPIReplicaSpecs[kubeflowv1.MPIJobReplicaTypeLauncher].Template.Spec.Containers[0].Resources)
		assert.Equal(t, *workerResourceRequirements, mpiJob.Spec.MPIReplicaSpecs[kubeflowv1.MPIJobReplicaTypeWorker].Template.Spec.Containers[0].Resources)
		assert.Equal(t, launcherCommand, mpiJob.Spec.MPIReplicaSpecs[kubeflowv1.MPIJobReplicaTypeLauncher].Template.Spec.Containers[0].Args)
		assert.Equal(t, workerCommand, mpiJob.Spec.MPIReplicaSpecs[kubeflowv1.MPIJobReplicaTypeWorker].Template.Spec.Containers[0].Args)
	}
}

func TestBuildResourceMPIV1WithOnlyWorkerReplica(t *testing.T) {
	workerCommand := []string{"/usr/sbin/sshd", "/.sshd_config"}

	taskConfigs := []*kfplugins.DistributedMPITrainingTask{
		{
			WorkerReplicas: &kfplugins.DistributedMPITrainingReplicaSpec{
				Replicas: 100,
				Resources: &core.Resources{
					Requests: []*core.Resources_ResourceEntry{
						{Name: core.Resources_CPU, Value: "1024m"},
						{Name: core.Resources_MEMORY, Value: "1Gi"},
					},
					Limits: []*core.Resources_ResourceEntry{
						{Name: core.Resources_CPU, Value: "2048m"},
						{Name: core.Resources_MEMORY, Value: "2Gi"},
					},
				},
				Command: []string{"/usr/sbin/sshd", "/.sshd_config"},
			},
			Slots: int32(1),
		},
		{
			WorkerReplicas: &kfplugins.DistributedMPITrainingReplicaSpec{
				Common: &kfplugins.CommonReplicaSpec{
					Replicas: 100,
					Resources: &core.Resources{
						Requests: []*core.Resources_ResourceEntry{
							{Name: core.Resources_CPU, Value: "1024m"},
							{Name: core.Resources_MEMORY, Value: "1Gi"},
						},
						Limits: []*core.Resources_ResourceEntry{
							{Name: core.Resources_CPU, Value: "2048m"},
							{Name: core.Resources_MEMORY, Value: "2Gi"},
						},
					},
				},
				Command: []string{"/usr/sbin/sshd", "/.sshd_config"},
			},
			Slots: int32(1),
		},
	}

	for _, taskConfig := range taskConfigs {
		workerResourceRequirements := &corev1.ResourceRequirements{
			Requests: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("1024m"),
				corev1.ResourceMemory: resource.MustParse("1Gi"),
			},
			Limits: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("2048m"),
				corev1.ResourceMemory: resource.MustParse("2Gi"),
			},
		}

		mpiResourceHandler := mpiOperatorResourceHandler{}

		taskTemplate := dummyMPITaskTemplate(mpiID2, taskConfig)
		taskTemplate.TaskTypeVersion = 1

		resource, err := mpiResourceHandler.BuildResource(context.TODO(), dummyMPITaskContext(taskTemplate, resourceRequirements, nil, k8s.PluginState{}))
		assert.NoError(t, err)
		assert.NotNil(t, resource)

		mpiJob, ok := resource.(*kubeflowv1.MPIJob)
		assert.True(t, ok)
		assert.Equal(t, int32(1), *mpiJob.Spec.MPIReplicaSpecs[kubeflowv1.MPIJobReplicaTypeLauncher].Replicas)
		assert.Equal(t, int32(100), *mpiJob.Spec.MPIReplicaSpecs[kubeflowv1.MPIJobReplicaTypeWorker].Replicas)
		assert.Equal(t, int32(1), *mpiJob.Spec.SlotsPerWorker)
		assert.Equal(t, *workerResourceRequirements, mpiJob.Spec.MPIReplicaSpecs[kubeflowv1.MPIJobReplicaTypeWorker].Template.Spec.Containers[0].Resources)
		assert.Equal(t, testArgs, mpiJob.Spec.MPIReplicaSpecs[kubeflowv1.MPIJobReplicaTypeLauncher].Template.Spec.Containers[0].Args)
		assert.Equal(t, workerCommand, mpiJob.Spec.MPIReplicaSpecs[kubeflowv1.MPIJobReplicaTypeWorker].Template.Spec.Containers[0].Args)
	}
}

func TestBuildResourceMPIV1ResourceTolerations(t *testing.T) {
	gpuToleration := corev1.Toleration{
		Key:      "nvidia.com/gpu",
		Value:    "present",
		Operator: corev1.TolerationOpEqual,
		Effect:   corev1.TaintEffectNoSchedule,
	}
	assert.NoError(t, flytek8sConfig.SetK8sPluginConfig(&flytek8sConfig.K8sPluginConfig{
		GpuResourceName: flytek8s.ResourceNvidiaGPU,
		ResourceTolerations: map[corev1.ResourceName][]corev1.Toleration{
			flytek8s.ResourceNvidiaGPU: {gpuToleration},
		},
	}))

	taskConfigs := []*kfplugins.DistributedMPITrainingTask{
		{
			LauncherReplicas: &kfplugins.DistributedMPITrainingReplicaSpec{
				Resources: &core.Resources{
					Requests: []*core.Resources_ResourceEntry{
						{Name: core.Resources_CPU, Value: "250m"},
						{Name: core.Resources_MEMORY, Value: "250Mi"},
					},
					Limits: []*core.Resources_ResourceEntry{
						{Name: core.Resources_CPU, Value: "500m"},
						{Name: core.Resources_MEMORY, Value: "500Mi"},
					},
				},
			},
			WorkerReplicas: &kfplugins.DistributedMPITrainingReplicaSpec{
				Replicas: 100,
				Resources: &core.Resources{
					Requests: []*core.Resources_ResourceEntry{
						{Name: core.Resources_CPU, Value: "1024m"},
						{Name: core.Resources_MEMORY, Value: "1Gi"},
						{Name: core.Resources_GPU, Value: "1"},
					},
					Limits: []*core.Resources_ResourceEntry{
						{Name: core.Resources_CPU, Value: "2048m"},
						{Name: core.Resources_MEMORY, Value: "2Gi"},
						{Name: core.Resources_GPU, Value: "1"},
					},
				},
			},
		},
		{
			LauncherReplicas: &kfplugins.DistributedMPITrainingReplicaSpec{
				Common: &kfplugins.CommonReplicaSpec{
					Resources: &core.Resources{
						Requests: []*core.Resources_ResourceEntry{
							{Name: core.Resources_CPU, Value: "250m"},
							{Name: core.Resources_MEMORY, Value: "250Mi"},
						},
						Limits: []*core.Resources_ResourceEntry{
							{Name: core.Resources_CPU, Value: "500m"},
							{Name: core.Resources_MEMORY, Value: "500Mi"},
						},
					},
				},
			},
			WorkerReplicas: &kfplugins.DistributedMPITrainingReplicaSpec{
				Common: &kfplugins.CommonReplicaSpec{
					Replicas: 100,
					Resources: &core.Resources{
						Requests: []*core.Resources_ResourceEntry{
							{Name: core.Resources_CPU, Value: "1024m"},
							{Name: core.Resources_MEMORY, Value: "1Gi"},
							{Name: core.Resources_GPU, Value: "1"},
						},
						Limits: []*core.Resources_ResourceEntry{
							{Name: core.Resources_CPU, Value: "2048m"},
							{Name: core.Resources_MEMORY, Value: "2Gi"},
							{Name: core.Resources_GPU, Value: "1"},
						},
					},
				},
			},
		},
	}

	for _, taskConfig := range taskConfigs {
		mpiResourceHandler := mpiOperatorResourceHandler{}

		taskTemplate := dummyMPITaskTemplate(mpiID2, taskConfig)
		taskTemplate.TaskTypeVersion = 1

		resource, err := mpiResourceHandler.BuildResource(context.TODO(), dummyMPITaskContext(taskTemplate, resourceRequirements, nil, k8s.PluginState{}))
		assert.NoError(t, err)
		assert.NotNil(t, resource)

		mpiJob, ok := resource.(*kubeflowv1.MPIJob)
		assert.True(t, ok)

		assert.NotContains(t, mpiJob.Spec.MPIReplicaSpecs[kubeflowv1.MPIJobReplicaTypeLauncher].Template.Spec.Tolerations, gpuToleration)
		assert.Contains(t, mpiJob.Spec.MPIReplicaSpecs[kubeflowv1.MPIJobReplicaTypeWorker].Template.Spec.Tolerations, gpuToleration)
	}
}

func TestGetReplicaCount(t *testing.T) {
	mpiResourceHandler := mpiOperatorResourceHandler{}
	tfObj := dummyMPICustomObj(1, 1, 0)
	taskTemplate := dummyMPITaskTemplate("the job", tfObj)
	resource, err := mpiResourceHandler.BuildResource(context.TODO(), dummyMPITaskContext(taskTemplate, resourceRequirements, nil, k8s.PluginState{}))
	assert.NoError(t, err)
	assert.NotNil(t, resource)
	MPIJob, ok := resource.(*kubeflowv1.MPIJob)
	assert.True(t, ok)

	assert.NotNil(t, common.GetReplicaCount(MPIJob.Spec.MPIReplicaSpecs, kubeflowv1.MPIJobReplicaTypeWorker))
	assert.NotNil(t, common.GetReplicaCount(MPIJob.Spec.MPIReplicaSpecs, kubeflowv1.MPIJobReplicaTypeLauncher))
}
