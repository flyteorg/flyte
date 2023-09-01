package tensorflow

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/flyteorg/flyteplugins/go/tasks/plugins/k8s/kfoperators/common"

	"github.com/flyteorg/flyteplugins/go/tasks/logs"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/flytek8s"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/k8s"
	commonOp "github.com/kubeflow/common/pkg/apis/common/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/stretchr/testify/mock"

	pluginsCore "github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/utils"

	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/core/mocks"

	pluginIOMocks "github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/io/mocks"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/plugins"
	kfplugins "github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/plugins/kubeflow"
	"github.com/golang/protobuf/jsonpb"
	structpb "github.com/golang/protobuf/ptypes/struct"
	kubeflowv1 "github.com/kubeflow/training-operator/pkg/apis/kubeflow.org/v1"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

func dummyTensorFlowCustomObj(workers int32, psReplicas int32, chiefReplicas int32) *plugins.DistributedTensorflowTrainingTask {
	return &plugins.DistributedTensorflowTrainingTask{
		Workers:       workers,
		PsReplicas:    psReplicas,
		ChiefReplicas: chiefReplicas,
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
			err = fmt.Errorf("Unkonw input type %T", t)
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

func dummyTensorFlowTaskContext(taskTemplate *core.TaskTemplate) pluginsCore.TaskExecutionContext {
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

	resources := &mocks.TaskOverrides{}
	resources.OnGetResources().Return(resourceRequirements)

	taskExecutionMetadata := &mocks.TaskExecutionMetadata{}
	taskExecutionMetadata.OnGetTaskExecutionID().Return(tID)
	taskExecutionMetadata.OnGetNamespace().Return("test-namespace")
	taskExecutionMetadata.OnGetAnnotations().Return(map[string]string{"annotation-1": "val1"})
	taskExecutionMetadata.OnGetLabels().Return(map[string]string{"label-1": "val1"})
	taskExecutionMetadata.OnGetOwnerReference().Return(v1.OwnerReference{
		Kind: "node",
		Name: "blah",
	})
	taskExecutionMetadata.OnIsInterruptible().Return(true)
	taskExecutionMetadata.OnGetOverrides().Return(resources)
	taskExecutionMetadata.OnGetK8sServiceAccount().Return(serviceAccount)
	taskExecutionMetadata.OnGetPlatformResources().Return(&corev1.ResourceRequirements{})
	taskExecutionMetadata.OnGetEnvironmentVariables().Return(nil)
	taskCtx.OnTaskExecutionMetadata().Return(taskExecutionMetadata)
	return taskCtx
}

func dummyTensorFlowJobResource(tensorflowResourceHandler tensorflowOperatorResourceHandler,
	workers int32, psReplicas int32, chiefReplicas int32, conditionType commonOp.JobConditionType) *kubeflowv1.TFJob {
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

	tfObj := dummyTensorFlowCustomObj(workers, psReplicas, chiefReplicas)
	taskTemplate := dummyTensorFlowTaskTemplate("the job", tfObj)
	resource, err := tensorflowResourceHandler.BuildResource(context.TODO(), dummyTensorFlowTaskContext(taskTemplate))
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

func TestBuildResourceTensorFlow(t *testing.T) {
	tensorflowResourceHandler := tensorflowOperatorResourceHandler{}

	tfObj := dummyTensorFlowCustomObj(100, 50, 1)
	taskTemplate := dummyTensorFlowTaskTemplate("the job", tfObj)

	resource, err := tensorflowResourceHandler.BuildResource(context.TODO(), dummyTensorFlowTaskContext(taskTemplate))
	assert.NoError(t, err)
	assert.NotNil(t, resource)

	tensorflowJob, ok := resource.(*kubeflowv1.TFJob)
	assert.True(t, ok)
	assert.Equal(t, int32(100), *tensorflowJob.Spec.TFReplicaSpecs[kubeflowv1.TFJobReplicaTypeWorker].Replicas)
	assert.Equal(t, int32(50), *tensorflowJob.Spec.TFReplicaSpecs[kubeflowv1.TFJobReplicaTypePS].Replicas)
	assert.Equal(t, int32(1), *tensorflowJob.Spec.TFReplicaSpecs[kubeflowv1.TFJobReplicaTypeChief].Replicas)

	for _, replicaSpec := range tensorflowJob.Spec.TFReplicaSpecs {
		var hasContainerWithDefaultTensorFlowName = false

		for _, container := range replicaSpec.Template.Spec.Containers {
			if container.Name == kubeflowv1.TFJobDefaultContainerName {
				hasContainerWithDefaultTensorFlowName = true
			}

			assert.Equal(t, resourceRequirements.Requests, container.Resources.Requests)
			assert.Equal(t, resourceRequirements.Limits, container.Resources.Limits)
		}

		assert.True(t, hasContainerWithDefaultTensorFlowName)
	}
}

func TestGetTaskPhase(t *testing.T) {
	tensorflowResourceHandler := tensorflowOperatorResourceHandler{}
	ctx := context.TODO()

	dummyTensorFlowJobResourceCreator := func(conditionType commonOp.JobConditionType) *kubeflowv1.TFJob {
		return dummyTensorFlowJobResource(tensorflowResourceHandler, 2, 1, 1, conditionType)
	}

	taskCtx := dummyTensorFlowTaskContext(dummyTensorFlowTaskTemplate("", dummyTensorFlowCustomObj(2, 1, 1)))
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

	tensorflowResourceHandler := tensorflowOperatorResourceHandler{}
	tensorFlowJob := dummyTensorFlowJobResource(tensorflowResourceHandler, workers, psReplicas, chiefReplicas, commonOp.JobRunning)
	taskCtx := dummyTensorFlowTaskContext(dummyTensorFlowTaskTemplate("", dummyTensorFlowCustomObj(workers, psReplicas, chiefReplicas)))
	jobLogs, err := common.GetLogs(taskCtx, common.TensorflowTaskType, tensorFlowJob.ObjectMeta, false,
		workers, psReplicas, chiefReplicas)
	assert.NoError(t, err)
	assert.Equal(t, 4, len(jobLogs))
	assert.Equal(t, fmt.Sprintf("k8s.com/#!/log/%s/%s-worker-0/pod?namespace=tensorflow-namespace", jobNamespace, jobName), jobLogs[0].Uri)
	assert.Equal(t, fmt.Sprintf("k8s.com/#!/log/%s/%s-worker-1/pod?namespace=tensorflow-namespace", jobNamespace, jobName), jobLogs[1].Uri)
	assert.Equal(t, fmt.Sprintf("k8s.com/#!/log/%s/%s-psReplica-0/pod?namespace=tensorflow-namespace", jobNamespace, jobName), jobLogs[2].Uri)
	assert.Equal(t, fmt.Sprintf("k8s.com/#!/log/%s/%s-chiefReplica-0/pod?namespace=tensorflow-namespace", jobNamespace, jobName), jobLogs[3].Uri)
}

func TestGetProperties(t *testing.T) {
	tensorflowResourceHandler := tensorflowOperatorResourceHandler{}
	expected := k8s.PluginProperties{}
	assert.Equal(t, expected, tensorflowResourceHandler.GetProperties())
}

func TestReplicaCounts(t *testing.T) {
	for _, test := range []struct {
		name               string
		chiefReplicaCount  int32
		psReplicaCount     int32
		workerReplicaCount int32
		expectError        bool
		contains           []commonOp.ReplicaType
		notContains        []commonOp.ReplicaType
	}{
		{"NoWorkers", 1, 1, 0, true, nil, nil},
		{"SingleChief", 1, 0, 1, false,
			[]commonOp.ReplicaType{kubeflowv1.TFJobReplicaTypeChief, kubeflowv1.TFJobReplicaTypeWorker},
			[]commonOp.ReplicaType{kubeflowv1.TFJobReplicaTypePS}},
		{"SinglePS", 0, 1, 1, false,
			[]commonOp.ReplicaType{kubeflowv1.TFJobReplicaTypePS, kubeflowv1.TFJobReplicaTypeWorker},
			[]commonOp.ReplicaType{kubeflowv1.TFJobReplicaTypeChief}},
	} {
		t.Run(test.name, func(t *testing.T) {
			tensorflowResourceHandler := tensorflowOperatorResourceHandler{}

			tfObj := dummyTensorFlowCustomObj(test.workerReplicaCount, test.psReplicaCount, test.chiefReplicaCount)
			taskTemplate := dummyTensorFlowTaskTemplate("the job", tfObj)

			resource, err := tensorflowResourceHandler.BuildResource(context.TODO(), dummyTensorFlowTaskContext(taskTemplate))
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
					{Name: core.Resources_GPU, Value: "1"},
				},
				Limits: []*core.Resources_ResourceEntry{
					{Name: core.Resources_CPU, Value: "2048m"},
					{Name: core.Resources_GPU, Value: "1"},
				},
			},
		},
		PsReplicas: &kfplugins.DistributedTensorflowTrainingReplicaSpec{
			Replicas: 50,
			Resources: &core.Resources{
				Requests: []*core.Resources_ResourceEntry{
					{Name: core.Resources_CPU, Value: "250m"},
				},
				Limits: []*core.Resources_ResourceEntry{
					{Name: core.Resources_CPU, Value: "500m"},
				},
			},
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
				flytek8s.ResourceNvidiaGPU: resource.MustParse("1"),
			},
			Limits: corev1.ResourceList{
				corev1.ResourceCPU:         resource.MustParse("2048m"),
				flytek8s.ResourceNvidiaGPU: resource.MustParse("1"),
			},
		},
		kubeflowv1.TFJobReplicaTypePS: {
			Requests: corev1.ResourceList{
				corev1.ResourceCPU: resource.MustParse("250m"),
			},
			Limits: corev1.ResourceList{
				corev1.ResourceCPU: resource.MustParse("500m"),
			},
		},
	}

	tensorflowResourceHandler := tensorflowOperatorResourceHandler{}

	taskTemplate := dummyTensorFlowTaskTemplate("v1", taskConfig)
	taskTemplate.TaskTypeVersion = 1

	resource, err := tensorflowResourceHandler.BuildResource(context.TODO(), dummyTensorFlowTaskContext(taskTemplate))
	assert.NoError(t, err)
	assert.NotNil(t, resource)

	tensorflowJob, ok := resource.(*kubeflowv1.TFJob)
	assert.True(t, ok)
	assert.Equal(t, int32(100), *tensorflowJob.Spec.TFReplicaSpecs[kubeflowv1.TFJobReplicaTypeWorker].Replicas)
	assert.Equal(t, int32(50), *tensorflowJob.Spec.TFReplicaSpecs[kubeflowv1.TFJobReplicaTypePS].Replicas)
	assert.Equal(t, int32(1), *tensorflowJob.Spec.TFReplicaSpecs[kubeflowv1.TFJobReplicaTypeChief].Replicas)

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
					{Name: core.Resources_GPU, Value: "1"},
				},
				Limits: []*core.Resources_ResourceEntry{
					{Name: core.Resources_CPU, Value: "2048m"},
					{Name: core.Resources_GPU, Value: "1"},
				},
			},
		},
	}

	resourceRequirementsMap := map[commonOp.ReplicaType]*corev1.ResourceRequirements{
		kubeflowv1.TFJobReplicaTypeWorker: {
			Requests: corev1.ResourceList{
				corev1.ResourceCPU:         resource.MustParse("1024m"),
				flytek8s.ResourceNvidiaGPU: resource.MustParse("1"),
			},
			Limits: corev1.ResourceList{
				corev1.ResourceCPU:         resource.MustParse("2048m"),
				flytek8s.ResourceNvidiaGPU: resource.MustParse("1"),
			},
		},
	}

	tensorflowResourceHandler := tensorflowOperatorResourceHandler{}

	taskTemplate := dummyTensorFlowTaskTemplate("v1 with only worker replica", taskConfig)
	taskTemplate.TaskTypeVersion = 1

	resource, err := tensorflowResourceHandler.BuildResource(context.TODO(), dummyTensorFlowTaskContext(taskTemplate))
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
