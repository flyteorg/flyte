package tensorflow

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

	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/logs"
	pluginsCore "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/core/mocks"
	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/flytek8s"
	flytek8sConfig "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/flytek8s/config"
	pluginIOMocks "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/io/mocks"
	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/k8s"
	k8smocks "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/k8s/mocks"
	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/utils"
	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/plugins/k8s/kfoperators/common"
	stdlibUtils "github.com/flyteorg/flyte/v2/flytestdlib/utils"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/core"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/plugins"
	kfplugins "github.com/flyteorg/flyte/v2/gen/go/flyteidl2/plugins/kubeflow"
)

const testImage = "image://"
const serviceAccount = "tensorflow_sa"

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
			corev1.ResourceCPU:         resource.MustParse("1000m"),
			corev1.ResourceMemory:      resource.MustParse("1Gi"),
			flytek8s.ResourceNvidiaGPU: resource.MustParse("1"),
		},
		Requests: corev1.ResourceList{
			corev1.ResourceCPU:         resource.MustParse("100m"),
			corev1.ResourceMemory:      resource.MustParse("512Mi"),
			flytek8s.ResourceNvidiaGPU: resource.MustParse("1"),
		},
	}

	jobName      = "the-job"
	jobNamespace = "tensorflow-namespace"
)

func dummyTensorFlowCustomObj(workers int32, psReplicas int32, chiefReplicas int32, evaluatorReplicas int32) *plugins.DistributedTensorflowTrainingTask {
	return &plugins.DistributedTensorflowTrainingTask{
		Workers:           workers,
		PsReplicas:        psReplicas,
		ChiefReplicas:     chiefReplicas,
		EvaluatorReplicas: evaluatorReplicas,
	}
}

func dummyTensorFlowTaskTemplate(id string, args ...interface{}) *core.TaskTemplate {

	var tfObjJSON string
	var err error

	for _, arg := range args {
		switch t := arg.(type) {
		case *kfplugins.DistributedTensorflowTrainingTask:
			var tensorflowCustomObj = t
			tfObjJSON, err = utils.MarshalToString(tensorflowCustomObj)
		case *plugins.DistributedTensorflowTrainingTask:
			var tensorflowCustomObj = t
			tfObjJSON, err = utils.MarshalToString(tensorflowCustomObj)
		default:
			err = fmt.Errorf("Unknown input type %T", t)
		}
	}

	if err != nil {
		panic(err)
	}

	structObj := structpb.Struct{}

	err = stdlibUtils.UnmarshalStringToPb(tfObjJSON, &structObj)
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

func dummyTensorFlowTaskContext(taskTemplate *core.TaskTemplate, resources *corev1.ResourceRequirements, extendedResources *core.ExtendedResources) pluginsCore.TaskExecutionContext {
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
	tID.EXPECT().GetUniqueNodeID().Return("an-unique-id")

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
	taskCtx.EXPECT().TaskExecutionMetadata().Return(taskExecutionMetadata)
	pluginStateReaderMock := mocks.PluginStateReader{}
	pluginStateReaderMock.EXPECT().Get(mock.AnythingOfType(reflect.TypeOf(&k8s.PluginState{}).String())).RunAndReturn(
		func(v interface{}) (uint8, error) {
			*(v.(*k8s.PluginState)) = k8s.PluginState{}
			return 0, nil
		})

	taskCtx.EXPECT().PluginStateReader().Return(&pluginStateReaderMock)
	return taskCtx
}

func dummyTensorFlowPluginContext(taskTemplate *core.TaskTemplate, resources *corev1.ResourceRequirements, extendedResources *core.ExtendedResources, pluginState k8s.PluginState) *k8smocks.PluginContext {
	pCtx := &k8smocks.PluginContext{}
	inputReader := &pluginIOMocks.InputReader{}
	inputReader.EXPECT().GetInputPrefixPath().Return("/input/prefix")
	inputReader.EXPECT().GetInputPath().Return("/input")
	inputReader.EXPECT().Get(mock.Anything).Return(&core.LiteralMap{}, nil)
	pCtx.EXPECT().InputReader().Return(inputReader)

	outputReader := &pluginIOMocks.OutputWriter{}
	outputReader.EXPECT().GetOutputPath().Return("/data/outputs.pb")
	outputReader.EXPECT().GetOutputPrefixPath().Return("/data/")
	outputReader.EXPECT().GetRawOutputPrefix().Return("")
	outputReader.EXPECT().GetCheckpointPrefix().Return("/checkpoint")
	outputReader.EXPECT().GetPreviousCheckpointsPrefix().Return("/prev")
	pCtx.EXPECT().OutputWriter().Return(outputReader)

	taskReader := &mocks.TaskReader{}
	taskReader.EXPECT().Read(mock.Anything).Return(taskTemplate, nil)
	pCtx.EXPECT().TaskReader().Return(taskReader)

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
	tID.EXPECT().GetUniqueNodeID().Return("an-unique-id")

	overrides := &mocks.TaskOverrides{}
	overrides.EXPECT().GetResources().Return(resources)
	overrides.EXPECT().GetExtendedResources().Return(extendedResources)
	overrides.EXPECT().GetContainerImage().Return("")

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
	pCtx.EXPECT().TaskExecutionMetadata().Return(taskExecutionMetadata)

	pluginStateReaderMock := mocks.PluginStateReader{}
	pluginStateReaderMock.EXPECT().Get(mock.AnythingOfType(reflect.TypeOf(&pluginState).String())).RunAndReturn(
		func(v interface{}) (uint8, error) {
			*(v.(*k8s.PluginState)) = pluginState
			return 0, nil
		})
	pCtx.EXPECT().PluginStateReader().Return(&pluginStateReaderMock)

	return pCtx
}

func dummyTensorFlowJobResource(tensorflowResourceHandler tensorflowOperatorResourceHandler,
	workers int32, psReplicas int32, chiefReplicas int32, evaluatorReplicas int32, conditionType kubeflowv1.JobConditionType) *kubeflowv1.TFJob {
	var jobConditions []kubeflowv1.JobCondition

	now := time.Now()

	jobCreated := kubeflowv1.JobCondition{
		Type:    kubeflowv1.JobCreated,
		Status:  corev1.ConditionTrue,
		Reason:  "TensorFlowJobCreated",
		Message: "TensorFlowJob the-job is created.",
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
		Reason:  "TensorFlowJobRunning",
		Message: "TensorFlowJob the-job is running.",
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
		Reason:  "TensorFlowJobSucceeded",
		Message: "TensorFlowJob the-job is successfully completed.",
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
		Reason:  "TensorFlowJobFailed",
		Message: "TensorFlowJob the-job is failed.",
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
		Reason:  "TensorFlowJobRestarting",
		Message: "TensorFlowJob the-job is restarting because some replica(s) failed.",
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

	tfObj := dummyTensorFlowCustomObj(workers, psReplicas, chiefReplicas, evaluatorReplicas)
	taskTemplate := dummyTensorFlowTaskTemplate("the job", tfObj)
	resource, err := tensorflowResourceHandler.BuildResource(context.TODO(), dummyTensorFlowTaskContext(taskTemplate, resourceRequirements, nil))
	if err != nil {
		panic(err)
	}

	return &kubeflowv1.TFJob{
		ObjectMeta: v1.ObjectMeta{
			Name:      jobName,
			Namespace: jobNamespace,
		},
		Spec: resource.(*kubeflowv1.TFJob).Spec,
		Status: kubeflowv1.JobStatus{
			Conditions:        jobConditions,
			ReplicaStatuses:   nil,
			StartTime:         &v1.Time{Time: time.Now()},
			CompletionTime:    nil,
			LastReconcileTime: nil,
		},
	}
}

func TestGetReplicaCount(t *testing.T) {
	tensorflowResourceHandler := tensorflowOperatorResourceHandler{}
	tfObj := dummyTensorFlowCustomObj(1, 0, 0, 0)
	taskTemplate := dummyTensorFlowTaskTemplate("the job", tfObj)
	resource, err := tensorflowResourceHandler.BuildResource(context.TODO(), dummyTensorFlowTaskContext(taskTemplate, resourceRequirements, nil))
	assert.NoError(t, err)
	assert.NotNil(t, resource)
	tensorflowJob, ok := resource.(*kubeflowv1.TFJob)
	assert.True(t, ok)

	assert.NotNil(t, common.GetReplicaCount(tensorflowJob.Spec.TFReplicaSpecs, kubeflowv1.TFJobReplicaTypeWorker))
	assert.NotNil(t, common.GetReplicaCount(tensorflowJob.Spec.TFReplicaSpecs, kubeflowv1.TFJobReplicaTypePS))
	assert.NotNil(t, common.GetReplicaCount(tensorflowJob.Spec.TFReplicaSpecs, kubeflowv1.TFJobReplicaTypeChief))
	assert.NotNil(t, common.GetReplicaCount(tensorflowJob.Spec.TFReplicaSpecs, kubeflowv1.TFJobReplicaTypeEval))
}

func TestBuildResourceTensorFlow(t *testing.T) {
	tensorflowResourceHandler := tensorflowOperatorResourceHandler{}

	tfObj := dummyTensorFlowCustomObj(100, 50, 1, 1)
	taskTemplate := dummyTensorFlowTaskTemplate("the job", tfObj)

	resource, err := tensorflowResourceHandler.BuildResource(context.TODO(), dummyTensorFlowTaskContext(taskTemplate, resourceRequirements, nil))
	assert.NoError(t, err)
	assert.NotNil(t, resource)

	tensorflowJob, ok := resource.(*kubeflowv1.TFJob)
	assert.True(t, ok)
	assert.Equal(t, int32(100), *tensorflowJob.Spec.TFReplicaSpecs[kubeflowv1.TFJobReplicaTypeWorker].Replicas)
	assert.Equal(t, int32(50), *tensorflowJob.Spec.TFReplicaSpecs[kubeflowv1.TFJobReplicaTypePS].Replicas)
	assert.Equal(t, int32(1), *tensorflowJob.Spec.TFReplicaSpecs[kubeflowv1.TFJobReplicaTypeChief].Replicas)
	assert.Equal(t, int32(1), *tensorflowJob.Spec.TFReplicaSpecs[kubeflowv1.TFJobReplicaTypeEval].Replicas)

	// verify TaskExecutionMetadata labels and annotations are copied to the TensorFlowJob
	for k, v := range dummyAnnotations {
		for _, replicaSpec := range tensorflowJob.Spec.TFReplicaSpecs {
			assert.Equal(t, v, replicaSpec.Template.ObjectMeta.Annotations[k])
		}
	}
	for k, v := range dummyLabels {
		for _, replicaSpec := range tensorflowJob.Spec.TFReplicaSpecs {
			assert.Equal(t, v, replicaSpec.Template.ObjectMeta.Labels[k])
		}
	}

	for _, replicaSpec := range tensorflowJob.Spec.TFReplicaSpecs {
		var hasContainerWithDefaultTensorFlowName = false
		podSpec := replicaSpec.Template.Spec
		for _, container := range podSpec.Containers {
			if container.Name == kubeflowv1.TFJobDefaultContainerName {
				hasContainerWithDefaultTensorFlowName = true
			}

			assert.Equal(t, resourceRequirements.Requests, container.Resources.Requests)
			assert.Equal(t, resourceRequirements.Limits, container.Resources.Limits)
		}

		assert.True(t, hasContainerWithDefaultTensorFlowName)
	}
}

func TestBuildResourceTensorFlowExtendedResources(t *testing.T) {
	assert.NoError(t, flytek8sConfig.SetK8sPluginConfig(&flytek8sConfig.K8sPluginConfig{
		GpuDeviceNodeLabel:        "gpu-node-label",
		GpuPartitionSizeNodeLabel: "gpu-partition-size",
		GpuResourceName:           flytek8s.ResourceNvidiaGPU,
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
			},
		},
	}

	v0TaskTemplate := dummyTensorFlowTaskTemplate("v0", dummyTensorFlowCustomObj(100, 50, 1, 1))
	v1TaskTemplates := []*core.TaskTemplate{
		dummyTensorFlowTaskTemplate("v1", &kfplugins.DistributedTensorflowTrainingTask{
			ChiefReplicas: &kfplugins.DistributedTensorflowTrainingReplicaSpec{
				Replicas: 1,
			},
			WorkerReplicas: &kfplugins.DistributedTensorflowTrainingReplicaSpec{
				Replicas: 100,
			},
			PsReplicas: &kfplugins.DistributedTensorflowTrainingReplicaSpec{
				Replicas: 50,
			},
			EvaluatorReplicas: &kfplugins.DistributedTensorflowTrainingReplicaSpec{
				Replicas: 1,
			},
		}),
		dummyTensorFlowTaskTemplate("v1", &kfplugins.DistributedTensorflowTrainingTask{
			ChiefReplicas: &kfplugins.DistributedTensorflowTrainingReplicaSpec{
				Common: &plugins.CommonReplicaSpec{
					Replicas: 1,
				},
			},
			WorkerReplicas: &kfplugins.DistributedTensorflowTrainingReplicaSpec{
				Common: &plugins.CommonReplicaSpec{
					Replicas: 100,
				},
			},
			PsReplicas: &kfplugins.DistributedTensorflowTrainingReplicaSpec{
				Common: &plugins.CommonReplicaSpec{
					Replicas: 50,
				},
			},
			EvaluatorReplicas: &kfplugins.DistributedTensorflowTrainingReplicaSpec{
				Common: &plugins.CommonReplicaSpec{
					Replicas: 1,
				},
			},
		}),
	}
	for _, v1TaskTemplate := range v1TaskTemplates {
		v1TaskTemplate.TaskTypeVersion = 1
		testConfigs := []struct {
			name         string
			taskTemplate *core.TaskTemplate
		}{
			{"v0", v0TaskTemplate},
			{"v1", v1TaskTemplate},
		}

		for _, tCfg := range testConfigs {
			for _, f := range fixtures {
				t.Run(tCfg.name+" "+f.name, func(t *testing.T) {
					taskTemplate := *tCfg.taskTemplate
					taskTemplate.ExtendedResources = f.extendedResourcesBase
					tensorflowResourceHandler := tensorflowOperatorResourceHandler{}
					taskContext := dummyTensorFlowTaskContext(&taskTemplate, f.resources, f.extendedResourcesOverride)
					r, err := tensorflowResourceHandler.BuildResource(context.TODO(), taskContext)
					assert.NoError(t, err)
					assert.NotNil(t, r)
					tensorflowJob, ok := r.(*kubeflowv1.TFJob)
					assert.True(t, ok)

					for _, replicaSpec := range tensorflowJob.Spec.TFReplicaSpecs {
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
	}
}

func TestGetTaskPhase(t *testing.T) {
	tensorflowResourceHandler := tensorflowOperatorResourceHandler{}
	ctx := context.TODO()

	dummyTensorFlowJobResourceCreator := func(conditionType kubeflowv1.JobConditionType) *kubeflowv1.TFJob {
		return dummyTensorFlowJobResource(tensorflowResourceHandler, 2, 1, 1, 1, conditionType)
	}

	pluginContext := dummyTensorFlowPluginContext(dummyTensorFlowTaskTemplate("", dummyTensorFlowCustomObj(2, 1, 1, 1)), resourceRequirements, nil, k8s.PluginState{})
	taskPhase, err := tensorflowResourceHandler.GetTaskPhase(ctx, pluginContext, dummyTensorFlowJobResourceCreator(kubeflowv1.JobCreated))
	assert.NoError(t, err)
	assert.Equal(t, pluginsCore.PhaseQueued, taskPhase.Phase())
	assert.NotNil(t, taskPhase.Info())
	assert.Nil(t, err)

	taskPhase, err = tensorflowResourceHandler.GetTaskPhase(ctx, pluginContext, dummyTensorFlowJobResourceCreator(kubeflowv1.JobRunning))
	assert.NoError(t, err)
	assert.Equal(t, pluginsCore.PhaseRunning, taskPhase.Phase())
	assert.NotNil(t, taskPhase.Info())
	assert.Nil(t, err)

	taskPhase, err = tensorflowResourceHandler.GetTaskPhase(ctx, pluginContext, dummyTensorFlowJobResourceCreator(kubeflowv1.JobSucceeded))
	assert.NoError(t, err)
	assert.Equal(t, pluginsCore.PhaseSuccess, taskPhase.Phase())
	assert.NotNil(t, taskPhase.Info())
	assert.Nil(t, err)

	taskPhase, err = tensorflowResourceHandler.GetTaskPhase(ctx, pluginContext, dummyTensorFlowJobResourceCreator(kubeflowv1.JobFailed))
	assert.NoError(t, err)
	assert.Equal(t, pluginsCore.PhaseRetryableFailure, taskPhase.Phase())
	assert.NotNil(t, taskPhase.Info())
	assert.Nil(t, err)

	taskPhase, err = tensorflowResourceHandler.GetTaskPhase(ctx, pluginContext, dummyTensorFlowJobResourceCreator(kubeflowv1.JobRestarting))
	assert.NoError(t, err)
	assert.Equal(t, pluginsCore.PhaseRunning, taskPhase.Phase())
	assert.NotNil(t, taskPhase.Info())
	assert.Nil(t, err)
}

func TestGetTaskPhaseIncreasePhaseVersion(t *testing.T) {
	tensorflowResourceHandler := tensorflowOperatorResourceHandler{}
	ctx := context.TODO()

	pluginState := k8s.PluginState{
		Phase:        pluginsCore.PhaseQueued,
		PhaseVersion: pluginsCore.DefaultPhaseVersion,
		Reason:       "task submitted to K8s",
	}
	pluginContext := dummyTensorFlowPluginContext(dummyTensorFlowTaskTemplate("", dummyTensorFlowCustomObj(2, 1, 1, 1)), resourceRequirements, nil, pluginState)

	taskPhase, err := tensorflowResourceHandler.GetTaskPhase(ctx, pluginContext, dummyTensorFlowJobResource(tensorflowResourceHandler, 2, 1, 1, 1, kubeflowv1.JobCreated))

	assert.NoError(t, err)
	assert.Equal(t, taskPhase.Version(), pluginsCore.DefaultPhaseVersion+1)
}

func TestGetLogs(t *testing.T) {
	assert.NoError(t, logs.SetLogConfig(&logs.LogConfig{
		IsKubernetesEnabled: true,
		KubernetesURL:       "k8s.com",
	}))

	workers := int32(2)
	psReplicas := int32(1)
	chiefReplicas := int32(1)
	evaluatorReplicas := int32(1)

	tensorflowResourceHandler := tensorflowOperatorResourceHandler{}
	tensorFlowJob := dummyTensorFlowJobResource(tensorflowResourceHandler, workers, psReplicas, chiefReplicas, evaluatorReplicas, kubeflowv1.JobRunning)
	taskTemplate := dummyTensorFlowTaskTemplate("", dummyTensorFlowCustomObj(workers, psReplicas, chiefReplicas, evaluatorReplicas))
	pluginContext := dummyTensorFlowPluginContext(taskTemplate, resourceRequirements, nil, k8s.PluginState{})
	jobLogs, err := common.GetLogs(pluginContext, common.TensorflowTaskType, tensorFlowJob.ObjectMeta, taskTemplate, false,
		workers, psReplicas, chiefReplicas, evaluatorReplicas, kubeflowv1.TFJobDefaultContainerName)

	assert.NoError(t, err)
	assert.Equal(t, 5, len(jobLogs))
	assert.Equal(t, fmt.Sprintf("k8s.com/#!/log/%s/%s-worker-0/pod?namespace=tensorflow-namespace", jobNamespace, jobName), jobLogs[0].Uri)
	assert.Equal(t, fmt.Sprintf("k8s.com/#!/log/%s/%s-worker-1/pod?namespace=tensorflow-namespace", jobNamespace, jobName), jobLogs[1].Uri)
	assert.Equal(t, fmt.Sprintf("k8s.com/#!/log/%s/%s-psReplica-0/pod?namespace=tensorflow-namespace", jobNamespace, jobName), jobLogs[2].Uri)
	assert.Equal(t, fmt.Sprintf("k8s.com/#!/log/%s/%s-chiefReplica-0/pod?namespace=tensorflow-namespace", jobNamespace, jobName), jobLogs[3].Uri)
	assert.Equal(t, fmt.Sprintf("k8s.com/#!/log/%s/%s-evaluatorReplica-0/pod?namespace=tensorflow-namespace", jobNamespace, jobName), jobLogs[4].Uri)
}

func TestGetProperties(t *testing.T) {
	tensorflowResourceHandler := tensorflowOperatorResourceHandler{}
	expected := k8s.PluginProperties{}
	assert.Equal(t, expected, tensorflowResourceHandler.GetProperties())
}

func TestReplicaCounts(t *testing.T) {
	for _, test := range []struct {
		name                  string
		chiefReplicaCount     int32
		psReplicaCount        int32
		workerReplicaCount    int32
		evaluatorReplicaCount int32
		expectError           bool
		contains              []kubeflowv1.ReplicaType
		notContains           []kubeflowv1.ReplicaType
	}{
		{"NoWorkers", 1, 1, 0, 1, true, nil, nil},
		{"SingleChief", 1, 0, 1, 0, false,
			[]kubeflowv1.ReplicaType{kubeflowv1.TFJobReplicaTypeChief, kubeflowv1.TFJobReplicaTypeWorker},
			[]kubeflowv1.ReplicaType{kubeflowv1.TFJobReplicaTypePS, kubeflowv1.TFJobReplicaTypeEval}},
		{"SinglePS", 0, 1, 1, 0, false,
			[]kubeflowv1.ReplicaType{kubeflowv1.TFJobReplicaTypePS, kubeflowv1.TFJobReplicaTypeWorker},
			[]kubeflowv1.ReplicaType{kubeflowv1.TFJobReplicaTypeChief, kubeflowv1.TFJobReplicaTypeEval}},
		{"AllContains", 1, 1, 1, 1, false,
			[]kubeflowv1.ReplicaType{kubeflowv1.TFJobReplicaTypePS, kubeflowv1.TFJobReplicaTypeWorker, kubeflowv1.TFJobReplicaTypeChief, kubeflowv1.TFJobReplicaTypeEval},
			nil,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			tensorflowResourceHandler := tensorflowOperatorResourceHandler{}

			tfObj := dummyTensorFlowCustomObj(test.workerReplicaCount, test.psReplicaCount, test.chiefReplicaCount, test.evaluatorReplicaCount)
			taskTemplate := dummyTensorFlowTaskTemplate("the job", tfObj)

			resource, err := tensorflowResourceHandler.BuildResource(context.TODO(), dummyTensorFlowTaskContext(taskTemplate, resourceRequirements, nil))
			if test.expectError {
				assert.Error(t, err)
				assert.Nil(t, resource)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, resource)

			job, ok := resource.(*kubeflowv1.TFJob)
			assert.True(t, ok)

			assert.Len(t, job.Spec.TFReplicaSpecs, len(test.contains))
			for _, replicaType := range test.contains {
				assert.Contains(t, job.Spec.TFReplicaSpecs, replicaType)
			}
			for _, replicaType := range test.notContains {
				assert.NotContains(t, job.Spec.TFReplicaSpecs, replicaType)
			}
		})
	}
}

func TestBuildResourceTensorFlowV1(t *testing.T) {
	taskConfigs := []*kfplugins.DistributedTensorflowTrainingTask{
		{
			ChiefReplicas: &kfplugins.DistributedTensorflowTrainingReplicaSpec{
				Replicas: 1,
				Image:    testImage,
				Resources: &core.Resources{
					Requests: []*core.Resources_ResourceEntry{
						{Name: core.Resources_CPU, Value: "250m"},
						{Name: core.Resources_MEMORY, Value: "1Gi"},
					},
					Limits: []*core.Resources_ResourceEntry{
						{Name: core.Resources_CPU, Value: "500m"},
						{Name: core.Resources_MEMORY, Value: "2Gi"},
					},
				},
				RestartPolicy: plugins.RestartPolicy_RESTART_POLICY_ALWAYS,
			},
			WorkerReplicas: &kfplugins.DistributedTensorflowTrainingReplicaSpec{
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
			PsReplicas: &kfplugins.DistributedTensorflowTrainingReplicaSpec{
				Replicas: 50,
				Resources: &core.Resources{
					Requests: []*core.Resources_ResourceEntry{
						{Name: core.Resources_CPU, Value: "250m"},
						{Name: core.Resources_MEMORY, Value: "1Gi"},
					},
					Limits: []*core.Resources_ResourceEntry{
						{Name: core.Resources_CPU, Value: "500m"},
						{Name: core.Resources_MEMORY, Value: "2Gi"},
					},
				},
			},
			EvaluatorReplicas: &kfplugins.DistributedTensorflowTrainingReplicaSpec{
				Replicas: 1,
				Image:    testImage,
				Resources: &core.Resources{
					Requests: []*core.Resources_ResourceEntry{
						{Name: core.Resources_CPU, Value: "250m"},
						{Name: core.Resources_MEMORY, Value: "1Gi"},
					},
					Limits: []*core.Resources_ResourceEntry{
						{Name: core.Resources_CPU, Value: "500m"},
						{Name: core.Resources_MEMORY, Value: "2Gi"},
					},
				},
				RestartPolicy: plugins.RestartPolicy_RESTART_POLICY_ALWAYS,
			},
			RunPolicy: &kfplugins.RunPolicy{
				CleanPodPolicy:        kfplugins.CleanPodPolicy_CLEANPOD_POLICY_ALL,
				ActiveDeadlineSeconds: int32(100),
			},
		},
		{
			ChiefReplicas: &kfplugins.DistributedTensorflowTrainingReplicaSpec{
				Common: &plugins.CommonReplicaSpec{
					Replicas: 1,
					Image:    testImage,
					Resources: &core.Resources{
						Requests: []*core.Resources_ResourceEntry{
							{Name: core.Resources_CPU, Value: "250m"},
							{Name: core.Resources_MEMORY, Value: "1Gi"},
						},
						Limits: []*core.Resources_ResourceEntry{
							{Name: core.Resources_CPU, Value: "500m"},
							{Name: core.Resources_MEMORY, Value: "2Gi"},
						},
					},
					RestartPolicy: plugins.RestartPolicy_RESTART_POLICY_ALWAYS,
				},
			},
			WorkerReplicas: &kfplugins.DistributedTensorflowTrainingReplicaSpec{
				Common: &plugins.CommonReplicaSpec{
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
			PsReplicas: &kfplugins.DistributedTensorflowTrainingReplicaSpec{
				Common: &plugins.CommonReplicaSpec{
					Replicas: 50,
					Resources: &core.Resources{
						Requests: []*core.Resources_ResourceEntry{
							{Name: core.Resources_CPU, Value: "250m"},
							{Name: core.Resources_MEMORY, Value: "1Gi"},
						},
						Limits: []*core.Resources_ResourceEntry{
							{Name: core.Resources_CPU, Value: "500m"},
							{Name: core.Resources_MEMORY, Value: "2Gi"},
						},
					},
				},
			},
			EvaluatorReplicas: &kfplugins.DistributedTensorflowTrainingReplicaSpec{
				Common: &plugins.CommonReplicaSpec{
					Replicas: 1,
					Image:    testImage,
					Resources: &core.Resources{
						Requests: []*core.Resources_ResourceEntry{
							{Name: core.Resources_CPU, Value: "250m"},
							{Name: core.Resources_MEMORY, Value: "1Gi"},
						},
						Limits: []*core.Resources_ResourceEntry{
							{Name: core.Resources_CPU, Value: "500m"},
							{Name: core.Resources_MEMORY, Value: "2Gi"},
						},
					},
					RestartPolicy: plugins.RestartPolicy_RESTART_POLICY_ALWAYS,
				},
			},
			RunPolicy: &kfplugins.RunPolicy{
				CleanPodPolicy:        kfplugins.CleanPodPolicy_CLEANPOD_POLICY_ALL,
				ActiveDeadlineSeconds: int32(100),
			},
		},
	}
	for _, taskConfig := range taskConfigs {

		resourceRequirementsMap := map[kubeflowv1.ReplicaType]*corev1.ResourceRequirements{
			kubeflowv1.TFJobReplicaTypeChief: {
				Requests: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("250m"),
					corev1.ResourceMemory: resource.MustParse("1Gi"),
				},
				Limits: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("500m"),
					corev1.ResourceMemory: resource.MustParse("2Gi"),
				},
			},
			kubeflowv1.TFJobReplicaTypeWorker: {
				Requests: corev1.ResourceList{
					corev1.ResourceCPU:         resource.MustParse("1024m"),
					corev1.ResourceMemory:      resource.MustParse("1Gi"),
					flytek8s.ResourceNvidiaGPU: resource.MustParse("1"),
				},
				Limits: corev1.ResourceList{
					corev1.ResourceCPU:         resource.MustParse("2048m"),
					corev1.ResourceMemory:      resource.MustParse("2Gi"),
					flytek8s.ResourceNvidiaGPU: resource.MustParse("1"),
				},
			},
			kubeflowv1.TFJobReplicaTypePS: {
				Requests: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("250m"),
					corev1.ResourceMemory: resource.MustParse("1Gi"),
				},
				Limits: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("500m"),
					corev1.ResourceMemory: resource.MustParse("2Gi"),
				},
			},
			kubeflowv1.TFJobReplicaTypeEval: {
				Requests: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("250m"),
					corev1.ResourceMemory: resource.MustParse("1Gi"),
				},
				Limits: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("500m"),
					corev1.ResourceMemory: resource.MustParse("2Gi"),
				},
			},
		}

		tensorflowResourceHandler := tensorflowOperatorResourceHandler{}

		taskTemplate := dummyTensorFlowTaskTemplate("v1", taskConfig)
		taskTemplate.TaskTypeVersion = 1

		resource, err := tensorflowResourceHandler.BuildResource(context.TODO(), dummyTensorFlowTaskContext(taskTemplate, resourceRequirements, nil))
		assert.NoError(t, err)
		assert.NotNil(t, resource)

		tensorflowJob, ok := resource.(*kubeflowv1.TFJob)
		assert.True(t, ok)
		assert.Equal(t, int32(100), *tensorflowJob.Spec.TFReplicaSpecs[kubeflowv1.TFJobReplicaTypeWorker].Replicas)
		assert.Equal(t, int32(50), *tensorflowJob.Spec.TFReplicaSpecs[kubeflowv1.TFJobReplicaTypePS].Replicas)
		assert.Equal(t, int32(1), *tensorflowJob.Spec.TFReplicaSpecs[kubeflowv1.TFJobReplicaTypeChief].Replicas)
		assert.Equal(t, int32(1), *tensorflowJob.Spec.TFReplicaSpecs[kubeflowv1.TFJobReplicaTypeEval].Replicas)

		for replicaType, replicaSpec := range tensorflowJob.Spec.TFReplicaSpecs {
			var hasContainerWithDefaultTensorFlowName = false

			for _, container := range replicaSpec.Template.Spec.Containers {
				if container.Name == kubeflowv1.TFJobDefaultContainerName {
					hasContainerWithDefaultTensorFlowName = true
					assert.Equal(t, *resourceRequirementsMap[replicaType], container.Resources)
				}
			}

			assert.True(t, hasContainerWithDefaultTensorFlowName)
		}
		assert.Equal(t, kubeflowv1.CleanPodPolicyAll, *tensorflowJob.Spec.RunPolicy.CleanPodPolicy)
		assert.Equal(t, int64(100), *tensorflowJob.Spec.RunPolicy.ActiveDeadlineSeconds)
	}
}

func TestBuildResourceTensorFlowV1WithOnlyWorker(t *testing.T) {
	taskConfigs := []*kfplugins.DistributedTensorflowTrainingTask{
		{
			WorkerReplicas: &kfplugins.DistributedTensorflowTrainingReplicaSpec{
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
			WorkerReplicas: &kfplugins.DistributedTensorflowTrainingReplicaSpec{
				Common: &plugins.CommonReplicaSpec{
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
		resourceRequirementsMap := map[kubeflowv1.ReplicaType]*corev1.ResourceRequirements{
			kubeflowv1.TFJobReplicaTypeWorker: {
				Requests: corev1.ResourceList{
					corev1.ResourceCPU:         resource.MustParse("1024m"),
					corev1.ResourceMemory:      resource.MustParse("1Gi"),
					flytek8s.ResourceNvidiaGPU: resource.MustParse("1"),
				},
				Limits: corev1.ResourceList{
					corev1.ResourceCPU:         resource.MustParse("2048m"),
					corev1.ResourceMemory:      resource.MustParse("2Gi"),
					flytek8s.ResourceNvidiaGPU: resource.MustParse("1"),
				},
			},
		}

		tensorflowResourceHandler := tensorflowOperatorResourceHandler{}

		taskTemplate := dummyTensorFlowTaskTemplate("v1 with only worker replica", taskConfig)
		taskTemplate.TaskTypeVersion = 1

		resource, err := tensorflowResourceHandler.BuildResource(context.TODO(), dummyTensorFlowTaskContext(taskTemplate, resourceRequirements, nil))
		assert.NoError(t, err)
		assert.NotNil(t, resource)

		tensorflowJob, ok := resource.(*kubeflowv1.TFJob)
		assert.True(t, ok)
		assert.Equal(t, int32(100), *tensorflowJob.Spec.TFReplicaSpecs[kubeflowv1.TFJobReplicaTypeWorker].Replicas)
		assert.Nil(t, tensorflowJob.Spec.TFReplicaSpecs[kubeflowv1.TFJobReplicaTypeChief])
		assert.Nil(t, tensorflowJob.Spec.TFReplicaSpecs[kubeflowv1.TFJobReplicaTypePS])

		for replicaType, replicaSpec := range tensorflowJob.Spec.TFReplicaSpecs {
			var hasContainerWithDefaultTensorFlowName = false

			for _, container := range replicaSpec.Template.Spec.Containers {
				if container.Name == kubeflowv1.TFJobDefaultContainerName {
					hasContainerWithDefaultTensorFlowName = true
					assert.Equal(t, *resourceRequirementsMap[replicaType], container.Resources)
				}
			}

			assert.True(t, hasContainerWithDefaultTensorFlowName)
		}
	}
}

func TestBuildResourceTensorFlowV1ResourceTolerations(t *testing.T) {
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

	taskConfigs := []*kfplugins.DistributedTensorflowTrainingTask{
		{
			ChiefReplicas: &kfplugins.DistributedTensorflowTrainingReplicaSpec{
				Replicas: 1,
				Resources: &core.Resources{
					Requests: []*core.Resources_ResourceEntry{
						{Name: core.Resources_CPU, Value: "250m"},
						{Name: core.Resources_MEMORY, Value: "1Gi"},
					},
					Limits: []*core.Resources_ResourceEntry{
						{Name: core.Resources_CPU, Value: "500m"},
						{Name: core.Resources_MEMORY, Value: "2Gi"},
					},
				},
			},
			WorkerReplicas: &kfplugins.DistributedTensorflowTrainingReplicaSpec{
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
			ChiefReplicas: &kfplugins.DistributedTensorflowTrainingReplicaSpec{
				Common: &plugins.CommonReplicaSpec{
					Replicas: 1,
					Resources: &core.Resources{
						Requests: []*core.Resources_ResourceEntry{
							{Name: core.Resources_CPU, Value: "250m"},
							{Name: core.Resources_MEMORY, Value: "1Gi"},
						},
						Limits: []*core.Resources_ResourceEntry{
							{Name: core.Resources_CPU, Value: "500m"},
							{Name: core.Resources_MEMORY, Value: "2Gi"},
						},
					},
				},
			},
			WorkerReplicas: &kfplugins.DistributedTensorflowTrainingReplicaSpec{
				Common: &plugins.CommonReplicaSpec{
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

		tensorflowResourceHandler := tensorflowOperatorResourceHandler{}

		taskTemplate := dummyTensorFlowTaskTemplate("v1", taskConfig)
		taskTemplate.TaskTypeVersion = 1

		resource, err := tensorflowResourceHandler.BuildResource(context.TODO(), dummyTensorFlowTaskContext(taskTemplate, resourceRequirements, nil))
		assert.NoError(t, err)
		assert.NotNil(t, resource)

		tensorflowJob, ok := resource.(*kubeflowv1.TFJob)
		assert.True(t, ok)

		assert.NotContains(t, tensorflowJob.Spec.TFReplicaSpecs[kubeflowv1.TFJobReplicaTypeChief].Template.Spec.Tolerations, gpuToleration)
		assert.Contains(t, tensorflowJob.Spec.TFReplicaSpecs[kubeflowv1.TFJobReplicaTypeWorker].Template.Spec.Tolerations, gpuToleration)
	}
}

func TestIsTerminal(t *testing.T) {
	tensorflowResourceHandler := tensorflowOperatorResourceHandler{}
	ctx := context.Background()

	tests := []struct {
		name           string
		conditionType  kubeflowv1.JobConditionType
		expectedResult bool
	}{
		{"Succeeded", kubeflowv1.JobSucceeded, true},
		{"Failed", kubeflowv1.JobFailed, true},
		{"Created", kubeflowv1.JobCreated, false},
		{"Running", kubeflowv1.JobRunning, false},
		{"Restarting", kubeflowv1.JobRestarting, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a simple job with only the condition we want to test
			job := &kubeflowv1.TFJob{
				Status: kubeflowv1.JobStatus{
					Conditions: []kubeflowv1.JobCondition{
						{
							Type:   tt.conditionType,
							Status: corev1.ConditionTrue,
						},
					},
				},
			}
			result, err := tensorflowResourceHandler.IsTerminal(ctx, job)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedResult, result)
		})
	}
}

func TestIsTerminal_WrongResourceType(t *testing.T) {
	tensorflowResourceHandler := tensorflowOperatorResourceHandler{}
	ctx := context.Background()

	wrongResource := &corev1.ConfigMap{}
	result, err := tensorflowResourceHandler.IsTerminal(ctx, wrongResource)
	assert.Error(t, err)
	assert.False(t, result)
	assert.Contains(t, err.Error(), "unexpected resource type")
}

func TestGetCompletionTime(t *testing.T) {
	tensorflowResourceHandler := tensorflowOperatorResourceHandler{}

	now := time.Now().Truncate(time.Second)
	earlier := now.Add(-1 * time.Hour)
	evenEarlier := now.Add(-2 * time.Hour)

	tests := []struct {
		name         string
		job          *kubeflowv1.TFJob
		expectedTime time.Time
	}{
		{
			name: "uses CompletionTime",
			job: &kubeflowv1.TFJob{
				ObjectMeta: v1.ObjectMeta{
					CreationTimestamp: v1.NewTime(evenEarlier),
				},
				Status: kubeflowv1.JobStatus{
					CompletionTime: &v1.Time{Time: now},
					StartTime:      &v1.Time{Time: earlier},
				},
			},
			expectedTime: now,
		},
		{
			name: "falls back to condition LastTransitionTime",
			job: &kubeflowv1.TFJob{
				ObjectMeta: v1.ObjectMeta{
					CreationTimestamp: v1.NewTime(evenEarlier),
				},
				Status: kubeflowv1.JobStatus{
					StartTime: &v1.Time{Time: earlier},
					Conditions: []kubeflowv1.JobCondition{
						{
							Type:               kubeflowv1.JobSucceeded,
							Status:             corev1.ConditionTrue,
							LastTransitionTime: v1.NewTime(now),
						},
					},
				},
			},
			expectedTime: now,
		},
		{
			name: "falls back to StartTime",
			job: &kubeflowv1.TFJob{
				ObjectMeta: v1.ObjectMeta{
					CreationTimestamp: v1.NewTime(evenEarlier),
				},
				Status: kubeflowv1.JobStatus{
					StartTime: &v1.Time{Time: now},
				},
			},
			expectedTime: now,
		},
		{
			name: "falls back to CreationTimestamp",
			job: &kubeflowv1.TFJob{
				ObjectMeta: v1.ObjectMeta{
					CreationTimestamp: v1.NewTime(now),
				},
				Status: kubeflowv1.JobStatus{},
			},
			expectedTime: now,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tensorflowResourceHandler.GetCompletionTime(tt.job)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedTime.Unix(), result.Unix())
		})
	}
}

func TestGetCompletionTime_WrongResourceType(t *testing.T) {
	tensorflowResourceHandler := tensorflowOperatorResourceHandler{}

	wrongResource := &corev1.ConfigMap{}
	result, err := tensorflowResourceHandler.GetCompletionTime(wrongResource)
	assert.Error(t, err)
	assert.True(t, result.IsZero())
	assert.Contains(t, err.Error(), "unexpected resource type")
}
