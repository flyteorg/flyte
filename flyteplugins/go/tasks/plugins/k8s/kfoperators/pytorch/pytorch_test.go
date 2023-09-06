package pytorch

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
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	kubeflowv1 "github.com/kubeflow/training-operator/pkg/apis/kubeflow.org/v1"
)

const testImage = "image://"
const testImageMaster = "image://master"
const serviceAccount = "pytorch_sa"

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
	jobNamespace = "pytorch-namespace"
)

func dummyPytorchCustomObj(workers int32) *plugins.DistributedPyTorchTrainingTask {
	return &plugins.DistributedPyTorchTrainingTask{
		Workers: workers,
	}
}

func dummyElasticPytorchCustomObj(workers int32, elasticConfig plugins.ElasticConfig) *plugins.DistributedPyTorchTrainingTask {
	return &plugins.DistributedPyTorchTrainingTask{
		Workers:       workers,
		ElasticConfig: &elasticConfig,
	}
}

func dummyPytorchTaskTemplate(id string, args ...interface{}) *core.TaskTemplate {

	var ptObjJSON string
	var err error

	for _, arg := range args {
		switch t := arg.(type) {
		case *kfplugins.DistributedPyTorchTrainingTask:
			var pytorchCustomObj = t
			ptObjJSON, err = utils.MarshalToString(pytorchCustomObj)
		case *plugins.DistributedPyTorchTrainingTask:
			var pytorchCustomObj = t
			ptObjJSON, err = utils.MarshalToString(pytorchCustomObj)
		default:
			err = fmt.Errorf("Unkonw input type %T", t)
		}
	}

	if err != nil {
		panic(err)
	}

	structObj := structpb.Struct{}

	err = jsonpb.UnmarshalString(ptObjJSON, &structObj)
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

func dummyPytorchTaskContext(taskTemplate *core.TaskTemplate) pluginsCore.TaskExecutionContext {
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
	resources.OnGetResources().Return(&corev1.ResourceRequirements{
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
	})

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

func dummyPytorchJobResource(pytorchResourceHandler pytorchOperatorResourceHandler, workers int32, conditionType commonOp.JobConditionType) *kubeflowv1.PyTorchJob {
	var jobConditions []commonOp.JobCondition

	now := time.Now()

	jobCreated := commonOp.JobCondition{
		Type:    commonOp.JobCreated,
		Status:  corev1.ConditionTrue,
		Reason:  "PyTorchJobCreated",
		Message: "PyTorchJob the-job is created.",
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
		Reason:  "PyTorchJobRunning",
		Message: "PyTorchJob the-job is running.",
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
		Reason:  "PyTorchJobSucceeded",
		Message: "PyTorchJob the-job is successfully completed.",
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
		Reason:  "PyTorchJobFailed",
		Message: "PyTorchJob the-job is failed.",
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
		Reason:  "PyTorchJobRestarting",
		Message: "PyTorchJob the-job is restarting because some replica(s) failed.",
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

	ptObj := dummyPytorchCustomObj(workers)
	taskTemplate := dummyPytorchTaskTemplate("job1", ptObj)
	resource, err := pytorchResourceHandler.BuildResource(context.TODO(), dummyPytorchTaskContext(taskTemplate))
	if err != nil {
		panic(err)
	}

	return &kubeflowv1.PyTorchJob{
		ObjectMeta: v1.ObjectMeta{
			CreationTimestamp: v1.Time{Time: time.Now()},
			Name:              jobName,
			Namespace:         jobNamespace,
		},
		Spec: resource.(*kubeflowv1.PyTorchJob).Spec,
		Status: commonOp.JobStatus{
			Conditions:        jobConditions,
			ReplicaStatuses:   nil,
			StartTime:         nil,
			CompletionTime:    nil,
			LastReconcileTime: nil,
		},
	}
}

func TestBuildResourcePytorchElastic(t *testing.T) {
	pytorchResourceHandler := pytorchOperatorResourceHandler{}

	ptObj := dummyElasticPytorchCustomObj(2, plugins.ElasticConfig{MinReplicas: 1, MaxReplicas: 2, NprocPerNode: 4, RdzvBackend: "c10d"})
	taskTemplate := dummyPytorchTaskTemplate("job2", ptObj)

	resource, err := pytorchResourceHandler.BuildResource(context.TODO(), dummyPytorchTaskContext(taskTemplate))
	assert.NoError(t, err)
	assert.NotNil(t, resource)

	pytorchJob, ok := resource.(*kubeflowv1.PyTorchJob)
	assert.True(t, ok)
	assert.Equal(t, int32(2), *pytorchJob.Spec.PyTorchReplicaSpecs[kubeflowv1.PyTorchJobReplicaTypeWorker].Replicas)
	assert.NotNil(t, pytorchJob.Spec.ElasticPolicy)
	assert.Equal(t, int32(1), *pytorchJob.Spec.ElasticPolicy.MinReplicas)
	assert.Equal(t, int32(2), *pytorchJob.Spec.ElasticPolicy.MaxReplicas)
	assert.Equal(t, int32(4), *pytorchJob.Spec.ElasticPolicy.NProcPerNode)
	assert.Equal(t, kubeflowv1.RDZVBackend("c10d"), *pytorchJob.Spec.ElasticPolicy.RDZVBackend)

	assert.Equal(t, 1, len(pytorchJob.Spec.PyTorchReplicaSpecs))
	assert.Contains(t, pytorchJob.Spec.PyTorchReplicaSpecs, kubeflowv1.PyTorchJobReplicaTypeWorker)

	var hasContainerWithDefaultPytorchName = false

	for _, container := range pytorchJob.Spec.PyTorchReplicaSpecs[kubeflowv1.PyTorchJobReplicaTypeWorker].Template.Spec.Containers {
		if container.Name == kubeflowv1.PytorchJobDefaultContainerName {
			hasContainerWithDefaultPytorchName = true
		}
	}

	assert.True(t, hasContainerWithDefaultPytorchName)
}

func TestBuildResourcePytorch(t *testing.T) {
	pytorchResourceHandler := pytorchOperatorResourceHandler{}

	ptObj := dummyPytorchCustomObj(100)
	taskTemplate := dummyPytorchTaskTemplate("job3", ptObj)

	res, err := pytorchResourceHandler.BuildResource(context.TODO(), dummyPytorchTaskContext(taskTemplate))
	assert.NoError(t, err)
	assert.NotNil(t, res)

	pytorchJob, ok := res.(*kubeflowv1.PyTorchJob)
	assert.True(t, ok)
	assert.Equal(t, int32(100), *pytorchJob.Spec.PyTorchReplicaSpecs[kubeflowv1.PyTorchJobReplicaTypeWorker].Replicas)
	assert.Nil(t, pytorchJob.Spec.ElasticPolicy)

	for _, replicaSpec := range pytorchJob.Spec.PyTorchReplicaSpecs {
		var hasContainerWithDefaultPytorchName = false

		for _, container := range replicaSpec.Template.Spec.Containers {
			if container.Name == kubeflowv1.PytorchJobDefaultContainerName {
				hasContainerWithDefaultPytorchName = true
			}

			assert.Equal(t, resourceRequirements.Requests, container.Resources.Requests, fmt.Sprintf(" container.Resources.Requests [%+v]", container.Resources.Requests.Cpu().String()))
			assert.Equal(t, resourceRequirements.Limits, container.Resources.Limits, fmt.Sprintf(" container.Resources.Limits [%+v]", container.Resources.Limits.Cpu().String()))
		}

		assert.True(t, hasContainerWithDefaultPytorchName)
	}
}

func TestGetTaskPhase(t *testing.T) {
	pytorchResourceHandler := pytorchOperatorResourceHandler{}
	ctx := context.TODO()

	dummyPytorchJobResourceCreator := func(conditionType commonOp.JobConditionType) *kubeflowv1.PyTorchJob {
		return dummyPytorchJobResource(pytorchResourceHandler, 2, conditionType)
	}

	taskCtx := dummyPytorchTaskContext(dummyPytorchTaskTemplate("", dummyPytorchCustomObj(2)))
	taskPhase, err := pytorchResourceHandler.GetTaskPhase(ctx, taskCtx, dummyPytorchJobResourceCreator(commonOp.JobCreated))
	assert.NoError(t, err)
	assert.Equal(t, pluginsCore.PhaseQueued, taskPhase.Phase())
	assert.NotNil(t, taskPhase.Info())
	assert.Nil(t, err)

	taskPhase, err = pytorchResourceHandler.GetTaskPhase(ctx, taskCtx, dummyPytorchJobResourceCreator(commonOp.JobRunning))
	assert.NoError(t, err)
	assert.Equal(t, pluginsCore.PhaseRunning, taskPhase.Phase())
	assert.NotNil(t, taskPhase.Info())
	assert.Nil(t, err)

	taskPhase, err = pytorchResourceHandler.GetTaskPhase(ctx, taskCtx, dummyPytorchJobResourceCreator(commonOp.JobSucceeded))
	assert.NoError(t, err)
	assert.Equal(t, pluginsCore.PhaseSuccess, taskPhase.Phase())
	assert.NotNil(t, taskPhase.Info())
	assert.Nil(t, err)

	taskPhase, err = pytorchResourceHandler.GetTaskPhase(ctx, taskCtx, dummyPytorchJobResourceCreator(commonOp.JobFailed))
	assert.NoError(t, err)
	assert.Equal(t, pluginsCore.PhaseRetryableFailure, taskPhase.Phase())
	assert.NotNil(t, taskPhase.Info())
	assert.Nil(t, err)

	taskPhase, err = pytorchResourceHandler.GetTaskPhase(ctx, taskCtx, dummyPytorchJobResourceCreator(commonOp.JobRestarting))
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

	hasMaster := true
	workers := int32(2)

	pytorchResourceHandler := pytorchOperatorResourceHandler{}
	pytorchJob := dummyPytorchJobResource(pytorchResourceHandler, workers, commonOp.JobRunning)
	taskCtx := dummyPytorchTaskContext(dummyPytorchTaskTemplate("", dummyPytorchCustomObj(workers)))
	jobLogs, err := common.GetLogs(taskCtx, common.PytorchTaskType, pytorchJob.ObjectMeta, hasMaster, workers, 0, 0)
	assert.NoError(t, err)
	assert.Equal(t, 3, len(jobLogs))
	assert.Equal(t, fmt.Sprintf("k8s.com/#!/log/%s/%s-master-0/pod?namespace=pytorch-namespace", jobNamespace, jobName), jobLogs[0].Uri)
	assert.Equal(t, fmt.Sprintf("k8s.com/#!/log/%s/%s-worker-0/pod?namespace=pytorch-namespace", jobNamespace, jobName), jobLogs[1].Uri)
	assert.Equal(t, fmt.Sprintf("k8s.com/#!/log/%s/%s-worker-1/pod?namespace=pytorch-namespace", jobNamespace, jobName), jobLogs[2].Uri)
}

func TestGetLogsElastic(t *testing.T) {
	assert.NoError(t, logs.SetLogConfig(&logs.LogConfig{
		IsKubernetesEnabled: true,
		KubernetesURL:       "k8s.com",
	}))

	hasMaster := false
	workers := int32(2)

	pytorchResourceHandler := pytorchOperatorResourceHandler{}
	pytorchJob := dummyPytorchJobResource(pytorchResourceHandler, workers, commonOp.JobRunning)
	taskCtx := dummyPytorchTaskContext(dummyPytorchTaskTemplate("", dummyPytorchCustomObj(workers)))
	jobLogs, err := common.GetLogs(taskCtx, common.PytorchTaskType, pytorchJob.ObjectMeta, hasMaster, workers, 0, 0)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(jobLogs))
	assert.Equal(t, fmt.Sprintf("k8s.com/#!/log/%s/%s-worker-0/pod?namespace=pytorch-namespace", jobNamespace, jobName), jobLogs[0].Uri)
	assert.Equal(t, fmt.Sprintf("k8s.com/#!/log/%s/%s-worker-1/pod?namespace=pytorch-namespace", jobNamespace, jobName), jobLogs[1].Uri)
}

func TestGetProperties(t *testing.T) {
	pytorchResourceHandler := pytorchOperatorResourceHandler{}
	expected := k8s.PluginProperties{}
	assert.Equal(t, expected, pytorchResourceHandler.GetProperties())
}

func TestReplicaCounts(t *testing.T) {
	for _, test := range []struct {
		name               string
		workerReplicaCount int32
		expectError        bool
		contains           []commonOp.ReplicaType
		notContains        []commonOp.ReplicaType
	}{
		{"NoWorkers", 0, true, nil, nil},
		{"Works", 1, false, []commonOp.ReplicaType{kubeflowv1.PyTorchJobReplicaTypeMaster, kubeflowv1.PyTorchJobReplicaTypeWorker}, []commonOp.ReplicaType{}},
	} {
		t.Run(test.name, func(t *testing.T) {
			pytorchResourceHandler := pytorchOperatorResourceHandler{}

			ptObj := dummyPytorchCustomObj(test.workerReplicaCount)
			taskTemplate := dummyPytorchTaskTemplate("the job", ptObj)

			res, err := pytorchResourceHandler.BuildResource(context.TODO(), dummyPytorchTaskContext(taskTemplate))
			if test.expectError {
				assert.Error(t, err)
				assert.Nil(t, res)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, res)

			job, ok := res.(*kubeflowv1.PyTorchJob)
			assert.True(t, ok)

			assert.Len(t, job.Spec.PyTorchReplicaSpecs, len(test.contains))
			for _, replicaType := range test.contains {
				assert.Contains(t, job.Spec.PyTorchReplicaSpecs, replicaType)
			}
			for _, replicaType := range test.notContains {
				assert.NotContains(t, job.Spec.PyTorchReplicaSpecs, replicaType)
			}
		})
	}
}

func TestBuildResourcePytorchV1(t *testing.T) {
	taskConfig := &kfplugins.DistributedPyTorchTrainingTask{
		MasterReplicas: &kfplugins.DistributedPyTorchTrainingReplicaSpec{
			Image: testImageMaster,
			Resources: &core.Resources{
				Requests: []*core.Resources_ResourceEntry{
					{Name: core.Resources_CPU, Value: "250m"},
				},
				Limits: []*core.Resources_ResourceEntry{
					{Name: core.Resources_CPU, Value: "500m"},
				},
			},
			RestartPolicy: kfplugins.RestartPolicy_RESTART_POLICY_ALWAYS,
		},
		WorkerReplicas: &kfplugins.DistributedPyTorchTrainingReplicaSpec{
			Replicas: 100,
			Resources: &core.Resources{
				Requests: []*core.Resources_ResourceEntry{
					{Name: core.Resources_CPU, Value: "1024m"},
				},
				Limits: []*core.Resources_ResourceEntry{
					{Name: core.Resources_CPU, Value: "2048m"},
				},
			},
		},
	}

	masterResourceRequirements := &corev1.ResourceRequirements{
		Requests: corev1.ResourceList{
			corev1.ResourceCPU: resource.MustParse("250m"),
		},
		Limits: corev1.ResourceList{
			corev1.ResourceCPU: resource.MustParse("500m"),
		},
	}

	workerResourceRequirements := &corev1.ResourceRequirements{
		Requests: corev1.ResourceList{
			corev1.ResourceCPU: resource.MustParse("1024m"),
		},
		Limits: corev1.ResourceList{
			corev1.ResourceCPU: resource.MustParse("2048m"),
		},
	}

	pytorchResourceHandler := pytorchOperatorResourceHandler{}

	taskTemplate := dummyPytorchTaskTemplate("job4", taskConfig)
	taskTemplate.TaskTypeVersion = 1

	res, err := pytorchResourceHandler.BuildResource(context.TODO(), dummyPytorchTaskContext(taskTemplate))
	assert.NoError(t, err)
	assert.NotNil(t, res)

	pytorchJob, ok := res.(*kubeflowv1.PyTorchJob)
	assert.True(t, ok)

	assert.Equal(t, int32(100), *pytorchJob.Spec.PyTorchReplicaSpecs[kubeflowv1.PyTorchJobReplicaTypeWorker].Replicas)
	assert.Nil(t, pytorchJob.Spec.PyTorchReplicaSpecs[kubeflowv1.PyTorchJobReplicaTypeMaster].Replicas)

	assert.Equal(t, testImageMaster, pytorchJob.Spec.PyTorchReplicaSpecs[kubeflowv1.PyTorchJobReplicaTypeMaster].Template.Spec.Containers[0].Image)
	assert.Equal(t, testImage, pytorchJob.Spec.PyTorchReplicaSpecs[kubeflowv1.PyTorchJobReplicaTypeWorker].Template.Spec.Containers[0].Image)

	assert.Equal(t, *masterResourceRequirements, pytorchJob.Spec.PyTorchReplicaSpecs[kubeflowv1.PyTorchJobReplicaTypeMaster].Template.Spec.Containers[0].Resources)
	assert.Equal(t, *workerResourceRequirements, pytorchJob.Spec.PyTorchReplicaSpecs[kubeflowv1.PyTorchJobReplicaTypeWorker].Template.Spec.Containers[0].Resources)

	assert.Equal(t, commonOp.RestartPolicyAlways, pytorchJob.Spec.PyTorchReplicaSpecs[kubeflowv1.PyTorchJobReplicaTypeMaster].RestartPolicy)
	assert.Equal(t, commonOp.RestartPolicyNever, pytorchJob.Spec.PyTorchReplicaSpecs[kubeflowv1.PyTorchJobReplicaTypeWorker].RestartPolicy)

	assert.Nil(t, pytorchJob.Spec.RunPolicy.CleanPodPolicy)
	assert.Nil(t, pytorchJob.Spec.RunPolicy.BackoffLimit)
	assert.Nil(t, pytorchJob.Spec.RunPolicy.TTLSecondsAfterFinished)
	assert.Nil(t, pytorchJob.Spec.RunPolicy.ActiveDeadlineSeconds)

	assert.Nil(t, pytorchJob.Spec.ElasticPolicy)
}

func TestBuildResourcePytorchV1WithRunPolicy(t *testing.T) {
	taskConfig := &kfplugins.DistributedPyTorchTrainingTask{
		WorkerReplicas: &kfplugins.DistributedPyTorchTrainingReplicaSpec{
			Replicas: 100,
		},
		RunPolicy: &kfplugins.RunPolicy{
			CleanPodPolicy:          kfplugins.CleanPodPolicy_CLEANPOD_POLICY_ALL,
			BackoffLimit:            100,
			ActiveDeadlineSeconds:   1000,
			TtlSecondsAfterFinished: 10000,
		},
	}
	pytorchResourceHandler := pytorchOperatorResourceHandler{}

	taskTemplate := dummyPytorchTaskTemplate("job5", taskConfig)
	taskTemplate.TaskTypeVersion = 1

	res, err := pytorchResourceHandler.BuildResource(context.TODO(), dummyPytorchTaskContext(taskTemplate))
	assert.NoError(t, err)
	assert.NotNil(t, res)

	pytorchJob, ok := res.(*kubeflowv1.PyTorchJob)
	assert.True(t, ok)
	assert.Equal(t, int32(100), *pytorchJob.Spec.PyTorchReplicaSpecs[kubeflowv1.PyTorchJobReplicaTypeWorker].Replicas)
	assert.Nil(t, pytorchJob.Spec.PyTorchReplicaSpecs[kubeflowv1.PyTorchJobReplicaTypeMaster].Replicas)
	assert.Equal(t, commonOp.CleanPodPolicyAll, *pytorchJob.Spec.RunPolicy.CleanPodPolicy)
	assert.Equal(t, int32(100), *pytorchJob.Spec.RunPolicy.BackoffLimit)
	assert.Equal(t, int64(1000), *pytorchJob.Spec.RunPolicy.ActiveDeadlineSeconds)
	assert.Equal(t, int32(10000), *pytorchJob.Spec.RunPolicy.TTLSecondsAfterFinished)
}

func TestBuildResourcePytorchV1WithOnlyWorkerSpec(t *testing.T) {
	taskConfig := &kfplugins.DistributedPyTorchTrainingTask{
		WorkerReplicas: &kfplugins.DistributedPyTorchTrainingReplicaSpec{
			Replicas: 100,
			Resources: &core.Resources{
				Requests: []*core.Resources_ResourceEntry{
					{Name: core.Resources_CPU, Value: "1024m"},
				},
				Limits: []*core.Resources_ResourceEntry{
					{Name: core.Resources_CPU, Value: "2048m"},
				},
			},
		},
	}
	// Master Replica should use resource from task override if not set
	taskOverrideResourceRequirements := &corev1.ResourceRequirements{
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

	workerResourceRequirements := &corev1.ResourceRequirements{
		Requests: corev1.ResourceList{
			corev1.ResourceCPU: resource.MustParse("1024m"),
		},
		Limits: corev1.ResourceList{
			corev1.ResourceCPU: resource.MustParse("2048m"),
		},
	}

	pytorchResourceHandler := pytorchOperatorResourceHandler{}

	taskTemplate := dummyPytorchTaskTemplate("job5", taskConfig)
	taskTemplate.TaskTypeVersion = 1

	res, err := pytorchResourceHandler.BuildResource(context.TODO(), dummyPytorchTaskContext(taskTemplate))
	assert.NoError(t, err)
	assert.NotNil(t, res)

	pytorchJob, ok := res.(*kubeflowv1.PyTorchJob)
	assert.True(t, ok)

	assert.Equal(t, int32(100), *pytorchJob.Spec.PyTorchReplicaSpecs[kubeflowv1.PyTorchJobReplicaTypeWorker].Replicas)
	assert.Nil(t, pytorchJob.Spec.PyTorchReplicaSpecs[kubeflowv1.PyTorchJobReplicaTypeMaster].Replicas)

	assert.Equal(t, testImage, pytorchJob.Spec.PyTorchReplicaSpecs[kubeflowv1.PyTorchJobReplicaTypeMaster].Template.Spec.Containers[0].Image)
	assert.Equal(t, testImage, pytorchJob.Spec.PyTorchReplicaSpecs[kubeflowv1.PyTorchJobReplicaTypeWorker].Template.Spec.Containers[0].Image)

	assert.Equal(t, *taskOverrideResourceRequirements, pytorchJob.Spec.PyTorchReplicaSpecs[kubeflowv1.PyTorchJobReplicaTypeMaster].Template.Spec.Containers[0].Resources)
	assert.Equal(t, *workerResourceRequirements, pytorchJob.Spec.PyTorchReplicaSpecs[kubeflowv1.PyTorchJobReplicaTypeWorker].Template.Spec.Containers[0].Resources)

	assert.Equal(t, commonOp.RestartPolicyNever, pytorchJob.Spec.PyTorchReplicaSpecs[kubeflowv1.PyTorchJobReplicaTypeMaster].RestartPolicy)
	assert.Equal(t, commonOp.RestartPolicyNever, pytorchJob.Spec.PyTorchReplicaSpecs[kubeflowv1.PyTorchJobReplicaTypeWorker].RestartPolicy)

	assert.Nil(t, pytorchJob.Spec.ElasticPolicy)
}

func TestBuildResourcePytorchV1WithElastic(t *testing.T) {
	taskConfig := &kfplugins.DistributedPyTorchTrainingTask{
		WorkerReplicas: &kfplugins.DistributedPyTorchTrainingReplicaSpec{
			Replicas: 2,
		},
		ElasticConfig: &kfplugins.ElasticConfig{MinReplicas: 1, MaxReplicas: 2, NprocPerNode: 4, RdzvBackend: "c10d"},
	}
	taskTemplate := dummyPytorchTaskTemplate("job5", taskConfig)
	taskTemplate.TaskTypeVersion = 1

	pytorchResourceHandler := pytorchOperatorResourceHandler{}
	resource, err := pytorchResourceHandler.BuildResource(context.TODO(), dummyPytorchTaskContext(taskTemplate))
	assert.NoError(t, err)
	assert.NotNil(t, resource)

	pytorchJob, ok := resource.(*kubeflowv1.PyTorchJob)
	assert.True(t, ok)
	assert.Equal(t, int32(2), *pytorchJob.Spec.PyTorchReplicaSpecs[kubeflowv1.PyTorchJobReplicaTypeWorker].Replicas)
	assert.NotNil(t, pytorchJob.Spec.ElasticPolicy)
	assert.Equal(t, int32(1), *pytorchJob.Spec.ElasticPolicy.MinReplicas)
	assert.Equal(t, int32(2), *pytorchJob.Spec.ElasticPolicy.MaxReplicas)
	assert.Equal(t, int32(4), *pytorchJob.Spec.ElasticPolicy.NProcPerNode)
	assert.Equal(t, kubeflowv1.RDZVBackend("c10d"), *pytorchJob.Spec.ElasticPolicy.RDZVBackend)

	assert.Equal(t, 1, len(pytorchJob.Spec.PyTorchReplicaSpecs))
	assert.Contains(t, pytorchJob.Spec.PyTorchReplicaSpecs, kubeflowv1.PyTorchJobReplicaTypeWorker)

	var hasContainerWithDefaultPytorchName = false

	for _, container := range pytorchJob.Spec.PyTorchReplicaSpecs[kubeflowv1.PyTorchJobReplicaTypeWorker].Template.Spec.Containers {
		if container.Name == kubeflowv1.PytorchJobDefaultContainerName {
			hasContainerWithDefaultPytorchName = true
		}
	}

	assert.True(t, hasContainerWithDefaultPytorchName)
}

func TestBuildResourcePytorchV1WithZeroWorker(t *testing.T) {
	taskConfig := &kfplugins.DistributedPyTorchTrainingTask{
		WorkerReplicas: &kfplugins.DistributedPyTorchTrainingReplicaSpec{
			Replicas: 0,
		},
	}
	pytorchResourceHandler := pytorchOperatorResourceHandler{}
	taskTemplate := dummyPytorchTaskTemplate("job5", taskConfig)
	taskTemplate.TaskTypeVersion = 1
	_, err := pytorchResourceHandler.BuildResource(context.TODO(), dummyPytorchTaskContext(taskTemplate))
	assert.Error(t, err)
}

func TestParseElasticConfig(t *testing.T) {
	elasticConfig := plugins.ElasticConfig{MinReplicas: 1, MaxReplicas: 2, NprocPerNode: 4, RdzvBackend: "c10d"}
	elasticPolicy := ParseElasticConfig(&elasticConfig)
	assert.Equal(t, int32(1), *elasticPolicy.MinReplicas)
	assert.Equal(t, int32(2), *elasticPolicy.MaxReplicas)
	assert.Equal(t, int32(4), *elasticPolicy.NProcPerNode)
	assert.Equal(t, kubeflowv1.RDZVBackend("c10d"), *elasticPolicy.RDZVBackend)
}
