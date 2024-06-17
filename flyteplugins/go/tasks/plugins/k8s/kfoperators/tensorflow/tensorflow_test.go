package tensorflow

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/golang/protobuf/jsonpb"
	structpb "github.com/golang/protobuf/ptypes/struct"
	commonOp "github.com/kubeflow/common/pkg/apis/common/v1"
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

	err = jsonpb.UnmarshalString(tfObjJSON, &structObj)
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
	inputReader.OnGetInputPrefixPath().Return("/input/prefix")
	inputReader.OnGetInputPath().Return("/input")
	inputReader.OnGetMatch(mock.Anything).Return(&core.LiteralMap{}, nil)
	taskCtx.OnInputReader().Return(inputReader)

	outputReader := &pluginIOMocks.OutputWriter{}
	outputReader.OnGetOutputPath().Return("/data/outputs.pb")
	outputReader.OnGetOutputPrefixPath().Return("/data/")
	outputReader.OnGetRawOutputPrefix().Return("")
	outputReader.OnGetCheckpointPrefix().Return("/checkpoint")
	outputReader.OnGetPreviousCheckpointsPrefix().Return("/prev")
	taskCtx.OnOutputWriter().Return(outputReader)

	taskReader := &mocks.TaskReader{}
	taskReader.OnReadMatch(mock.Anything).Return(taskTemplate, nil)
	taskCtx.OnTaskReader().Return(taskReader)

	tID := &mocks.TaskExecutionID{}
	tID.OnGetID().Return(core.TaskExecutionIdentifier{
		NodeExecutionId: &core.NodeExecutionIdentifier{
			ExecutionId: &core.WorkflowExecutionIdentifier{
				Name:    "my_name",
				Project: "my_project",
				Domain:  "my_domain",
			},
		},
	})
	tID.OnGetGeneratedName().Return("some-acceptable-name")
	tID.On("GetUniqueNodeID").Return("an-unique-id")

	overrides := &mocks.TaskOverrides{}
	overrides.OnGetResources().Return(resources)
	overrides.OnGetExtendedResources().Return(extendedResources)
	overrides.OnGetContainerImage().Return("")

	taskExecutionMetadata := &mocks.TaskExecutionMetadata{}
	taskExecutionMetadata.OnGetTaskExecutionID().Return(tID)
	taskExecutionMetadata.OnGetNamespace().Return("test-namespace")
	taskExecutionMetadata.OnGetAnnotations().Return(dummyAnnotations)
	taskExecutionMetadata.OnGetLabels().Return(dummyLabels)
	taskExecutionMetadata.OnGetOwnerReference().Return(v1.OwnerReference{
		Kind: "node",
		Name: "blah",
	})
	taskExecutionMetadata.OnIsInterruptible().Return(true)
	taskExecutionMetadata.OnGetOverrides().Return(overrides)
	taskExecutionMetadata.OnGetK8sServiceAccount().Return(serviceAccount)
	taskExecutionMetadata.OnGetPlatformResources().Return(&corev1.ResourceRequirements{})
	taskExecutionMetadata.OnGetEnvironmentVariables().Return(nil)
	taskExecutionMetadata.OnGetConsoleURL().Return("")
	taskCtx.OnTaskExecutionMetadata().Return(taskExecutionMetadata)
	return taskCtx
}

func dummyTensorFlowJobResource(tensorflowResourceHandler tensorflowOperatorResourceHandler,
	workers int32, psReplicas int32, chiefReplicas int32, evaluatorReplicas int32, conditionType commonOp.JobConditionType) *kubeflowv1.TFJob {
	var jobConditions []commonOp.JobCondition

	now := time.Now()

	jobCreated := commonOp.JobCondition{
		Type:    commonOp.JobCreated,
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
	jobRunningActive := commonOp.JobCondition{
		Type:    commonOp.JobRunning,
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
	jobSucceeded := commonOp.JobCondition{
		Type:    commonOp.JobSucceeded,
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
	jobFailed := commonOp.JobCondition{
		Type:    commonOp.JobFailed,
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
	jobRestarting := commonOp.JobCondition{
		Type:    commonOp.JobRestarting,
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
	case commonOp.JobCreated:
		jobConditions = []commonOp.JobCondition{
			jobCreated,
		}
	case commonOp.JobRunning:
		jobConditions = []commonOp.JobCondition{
			jobCreated,
			jobRunningActive,
		}
	case commonOp.JobSucceeded:
		jobConditions = []commonOp.JobCondition{
			jobCreated,
			jobRunningInactive,
			jobSucceeded,
		}
	case commonOp.JobFailed:
		jobConditions = []commonOp.JobCondition{
			jobCreated,
			jobRunningInactive,
			jobFailed,
		}
	case commonOp.JobRestarting:
		jobConditions = []commonOp.JobCondition{
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
		Status: commonOp.JobStatus{
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
	v1TaskTemplate := dummyTensorFlowTaskTemplate("v1", &kfplugins.DistributedTensorflowTrainingTask{
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
	})
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

func TestGetTaskPhase(t *testing.T) {
	tensorflowResourceHandler := tensorflowOperatorResourceHandler{}
	ctx := context.TODO()

	dummyTensorFlowJobResourceCreator := func(conditionType commonOp.JobConditionType) *kubeflowv1.TFJob {
		return dummyTensorFlowJobResource(tensorflowResourceHandler, 2, 1, 1, 1, conditionType)
	}

	taskCtx := dummyTensorFlowTaskContext(dummyTensorFlowTaskTemplate("", dummyTensorFlowCustomObj(2, 1, 1, 1)), resourceRequirements, nil)
	taskPhase, err := tensorflowResourceHandler.GetTaskPhase(ctx, taskCtx, dummyTensorFlowJobResourceCreator(commonOp.JobCreated))
	assert.NoError(t, err)
	assert.Equal(t, pluginsCore.PhaseQueued, taskPhase.Phase())
	assert.NotNil(t, taskPhase.Info())
	assert.Nil(t, err)

	taskPhase, err = tensorflowResourceHandler.GetTaskPhase(ctx, taskCtx, dummyTensorFlowJobResourceCreator(commonOp.JobRunning))
	assert.NoError(t, err)
	assert.Equal(t, pluginsCore.PhaseRunning, taskPhase.Phase())
	assert.NotNil(t, taskPhase.Info())
	assert.Nil(t, err)

	taskPhase, err = tensorflowResourceHandler.GetTaskPhase(ctx, taskCtx, dummyTensorFlowJobResourceCreator(commonOp.JobSucceeded))
	assert.NoError(t, err)
	assert.Equal(t, pluginsCore.PhaseSuccess, taskPhase.Phase())
	assert.NotNil(t, taskPhase.Info())
	assert.Nil(t, err)

	taskPhase, err = tensorflowResourceHandler.GetTaskPhase(ctx, taskCtx, dummyTensorFlowJobResourceCreator(commonOp.JobFailed))
	assert.NoError(t, err)
	assert.Equal(t, pluginsCore.PhaseRetryableFailure, taskPhase.Phase())
	assert.NotNil(t, taskPhase.Info())
	assert.Nil(t, err)

	taskPhase, err = tensorflowResourceHandler.GetTaskPhase(ctx, taskCtx, dummyTensorFlowJobResourceCreator(commonOp.JobRestarting))
	assert.NoError(t, err)
	assert.Equal(t, pluginsCore.PhaseRunning, taskPhase.Phase())
	assert.NotNil(t, taskPhase.Info())
	assert.Nil(t, err)
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
	tensorFlowJob := dummyTensorFlowJobResource(tensorflowResourceHandler, workers, psReplicas, chiefReplicas, evaluatorReplicas, commonOp.JobRunning)
	taskCtx := dummyTensorFlowTaskContext(dummyTensorFlowTaskTemplate("", dummyTensorFlowCustomObj(workers, psReplicas, chiefReplicas, evaluatorReplicas)), resourceRequirements, nil)
	jobLogs, err := common.GetLogs(taskCtx, common.TensorflowTaskType, tensorFlowJob.ObjectMeta, false,
		workers, psReplicas, chiefReplicas, evaluatorReplicas)
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
		contains              []commonOp.ReplicaType
		notContains           []commonOp.ReplicaType
	}{
		{"NoWorkers", 1, 1, 0, 1, true, nil, nil},
		{"SingleChief", 1, 0, 1, 0, false,
			[]commonOp.ReplicaType{kubeflowv1.TFJobReplicaTypeChief, kubeflowv1.TFJobReplicaTypeWorker},
			[]commonOp.ReplicaType{kubeflowv1.TFJobReplicaTypePS, kubeflowv1.TFJobReplicaTypeEval}},
		{"SinglePS", 0, 1, 1, 0, false,
			[]commonOp.ReplicaType{kubeflowv1.TFJobReplicaTypePS, kubeflowv1.TFJobReplicaTypeWorker},
			[]commonOp.ReplicaType{kubeflowv1.TFJobReplicaTypeChief, kubeflowv1.TFJobReplicaTypeEval}},
		{"AllContains", 1, 1, 1, 1, false,
			[]commonOp.ReplicaType{kubeflowv1.TFJobReplicaTypePS, kubeflowv1.TFJobReplicaTypeWorker, kubeflowv1.TFJobReplicaTypeChief, kubeflowv1.TFJobReplicaTypeEval},
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
	taskConfig := &kfplugins.DistributedTensorflowTrainingTask{
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
			RestartPolicy: kfplugins.RestartPolicy_RESTART_POLICY_ALWAYS,
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
			RestartPolicy: kfplugins.RestartPolicy_RESTART_POLICY_ALWAYS,
		},
		RunPolicy: &kfplugins.RunPolicy{
			CleanPodPolicy:        kfplugins.CleanPodPolicy_CLEANPOD_POLICY_ALL,
			ActiveDeadlineSeconds: int32(100),
		},
	}

	resourceRequirementsMap := map[commonOp.ReplicaType]*corev1.ResourceRequirements{
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
	assert.Equal(t, commonOp.CleanPodPolicyAll, *tensorflowJob.Spec.RunPolicy.CleanPodPolicy)
	assert.Equal(t, int64(100), *tensorflowJob.Spec.RunPolicy.ActiveDeadlineSeconds)
}

func TestBuildResourceTensorFlowV1WithOnlyWorker(t *testing.T) {
	taskConfig := &kfplugins.DistributedTensorflowTrainingTask{
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
	}

	resourceRequirementsMap := map[commonOp.ReplicaType]*corev1.ResourceRequirements{
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

	taskConfig := &kfplugins.DistributedTensorflowTrainingTask{
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
	}

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
