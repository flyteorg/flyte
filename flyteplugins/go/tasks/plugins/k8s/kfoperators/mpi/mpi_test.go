package mpi

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/flyteorg/flyteplugins/go/tasks/plugins/k8s/kfoperators/common"

	"github.com/flyteorg/flyteplugins/go/tasks/logs"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/k8s"
	mpiOp "github.com/kubeflow/common/pkg/apis/common/v1"
	kubeflowv1 "github.com/kubeflow/training-operator/pkg/apis/kubeflow.org/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/stretchr/testify/mock"

	pluginsCore "github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/utils"

	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/core/mocks"

	pluginIOMocks "github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/io/mocks"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/plugins"
	"github.com/golang/protobuf/jsonpb"
	structpb "github.com/golang/protobuf/ptypes/struct"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

func dummyMPITaskTemplate(id string, mpiCustomObj *plugins.DistributedMPITrainingTask) *core.TaskTemplate {

	mpiObjJSON, err := utils.MarshalToString(mpiCustomObj)
	if err != nil {
		panic(err)
	}

	structObj := structpb.Struct{}

	err = jsonpb.UnmarshalString(mpiObjJSON, &structObj)
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

func dummyMPITaskContext(taskTemplate *core.TaskTemplate) pluginsCore.TaskExecutionContext {
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
	taskCtx.OnTaskExecutionMetadata().Return(taskExecutionMetadata)
	return taskCtx
}

func dummyMPIJobResource(mpiResourceHandler mpiOperatorResourceHandler,
	workers int32, launcher int32, slots int32, conditionType mpiOp.JobConditionType) *kubeflowv1.MPIJob {
	var jobConditions []mpiOp.JobCondition

	now := time.Now()

	jobCreated := mpiOp.JobCondition{
		Type:    mpiOp.JobCreated,
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
	jobRunningActive := mpiOp.JobCondition{
		Type:    mpiOp.JobRunning,
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
	jobSucceeded := mpiOp.JobCondition{
		Type:    mpiOp.JobSucceeded,
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
	jobFailed := mpiOp.JobCondition{
		Type:    mpiOp.JobFailed,
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
	jobRestarting := mpiOp.JobCondition{
		Type:    mpiOp.JobRestarting,
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
	case mpiOp.JobCreated:
		jobConditions = []mpiOp.JobCondition{
			jobCreated,
		}
	case mpiOp.JobRunning:
		jobConditions = []mpiOp.JobCondition{
			jobCreated,
			jobRunningActive,
		}
	case mpiOp.JobSucceeded:
		jobConditions = []mpiOp.JobCondition{
			jobCreated,
			jobRunningInactive,
			jobSucceeded,
		}
	case mpiOp.JobFailed:
		jobConditions = []mpiOp.JobCondition{
			jobCreated,
			jobRunningInactive,
			jobFailed,
		}
	case mpiOp.JobRestarting:
		jobConditions = []mpiOp.JobCondition{
			jobCreated,
			jobRunningInactive,
			jobFailed,
			jobRestarting,
		}
	}

	mpiObj := dummyMPICustomObj(workers, launcher, slots)
	taskTemplate := dummyMPITaskTemplate(mpiID, mpiObj)
	resource, err := mpiResourceHandler.BuildResource(context.TODO(), dummyMPITaskContext(taskTemplate))
	if err != nil {
		panic(err)
	}

	return &kubeflowv1.MPIJob{
		ObjectMeta: v1.ObjectMeta{
			Name:      jobName,
			Namespace: jobNamespace,
		},
		Spec: resource.(*kubeflowv1.MPIJob).Spec,
		Status: mpiOp.JobStatus{
			Conditions:        jobConditions,
			ReplicaStatuses:   nil,
			StartTime:         nil,
			CompletionTime:    nil,
			LastReconcileTime: nil,
		},
	}
}

func TestBuildResourceMPI(t *testing.T) {
	mpiResourceHandler := mpiOperatorResourceHandler{}

	mpiObj := dummyMPICustomObj(100, 50, 1)
	taskTemplate := dummyMPITaskTemplate(mpiID2, mpiObj)

	resource, err := mpiResourceHandler.BuildResource(context.TODO(), dummyMPITaskContext(taskTemplate))
	assert.NoError(t, err)
	assert.NotNil(t, resource)

	mpiJob, ok := resource.(*kubeflowv1.MPIJob)
	assert.True(t, ok)
	assert.Equal(t, int32(50), *mpiJob.Spec.MPIReplicaSpecs[kubeflowv1.MPIJobReplicaTypeLauncher].Replicas)
	assert.Equal(t, int32(100), *mpiJob.Spec.MPIReplicaSpecs[kubeflowv1.MPIJobReplicaTypeWorker].Replicas)
	assert.Equal(t, int32(1), *mpiJob.Spec.SlotsPerWorker)

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

	_, err := mpiResourceHandler.BuildResource(context.TODO(), dummyMPITaskContext(taskTemplate))
	assert.Error(t, err)

	mpiObj = dummyMPICustomObj(1, 0, 1)
	taskTemplate = dummyMPITaskTemplate(mpiID2, mpiObj)

	_, err = mpiResourceHandler.BuildResource(context.TODO(), dummyMPITaskContext(taskTemplate))
	assert.Error(t, err)

	mpiObj = dummyMPICustomObj(1, 1, 1)
	taskTemplate = dummyMPITaskTemplate(mpiID2, mpiObj)

	resource, err := mpiResourceHandler.BuildResource(context.TODO(), dummyMPITaskContext(taskTemplate))
	app, ok := resource.(*kubeflowv1.MPIJob)
	assert.Nil(t, err)
	assert.Equal(t, true, ok)
	assert.Equal(t, []string{}, app.Spec.MPIReplicaSpecs[kubeflowv1.MPIJobReplicaTypeWorker].Template.Spec.Containers[0].Command)
	assert.Equal(t, []string{}, app.Spec.MPIReplicaSpecs[kubeflowv1.MPIJobReplicaTypeWorker].Template.Spec.Containers[0].Args)
}

func TestGetTaskPhase(t *testing.T) {
	mpiResourceHandler := mpiOperatorResourceHandler{}
	ctx := context.TODO()

	dummyMPIJobResourceCreator := func(conditionType mpiOp.JobConditionType) *kubeflowv1.MPIJob {
		return dummyMPIJobResource(mpiResourceHandler, 2, 1, 1, conditionType)
	}

	taskPhase, err := mpiResourceHandler.GetTaskPhase(ctx, nil, dummyMPIJobResourceCreator(mpiOp.JobCreated))
	assert.NoError(t, err)
	assert.Equal(t, pluginsCore.PhaseQueued, taskPhase.Phase())
	assert.NotNil(t, taskPhase.Info())
	assert.Nil(t, err)

	taskPhase, err = mpiResourceHandler.GetTaskPhase(ctx, nil, dummyMPIJobResourceCreator(mpiOp.JobRunning))
	assert.NoError(t, err)
	assert.Equal(t, pluginsCore.PhaseRunning, taskPhase.Phase())
	assert.NotNil(t, taskPhase.Info())
	assert.Nil(t, err)

	taskPhase, err = mpiResourceHandler.GetTaskPhase(ctx, nil, dummyMPIJobResourceCreator(mpiOp.JobSucceeded))
	assert.NoError(t, err)
	assert.Equal(t, pluginsCore.PhaseSuccess, taskPhase.Phase())
	assert.NotNil(t, taskPhase.Info())
	assert.Nil(t, err)

	taskPhase, err = mpiResourceHandler.GetTaskPhase(ctx, nil, dummyMPIJobResourceCreator(mpiOp.JobFailed))
	assert.NoError(t, err)
	assert.Equal(t, pluginsCore.PhaseRetryableFailure, taskPhase.Phase())
	assert.NotNil(t, taskPhase.Info())
	assert.Nil(t, err)

	taskPhase, err = mpiResourceHandler.GetTaskPhase(ctx, nil, dummyMPIJobResourceCreator(mpiOp.JobRestarting))
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
	launcher := int32(1)
	slots := int32(1)

	mpiResourceHandler := mpiOperatorResourceHandler{}
	mpiJob := dummyMPIJobResource(mpiResourceHandler, workers, launcher, slots, mpiOp.JobRunning)
	jobLogs, err := common.GetLogs(common.MPITaskType, mpiJob.Name, mpiJob.Namespace, workers, launcher, 0)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(jobLogs))
	assert.Equal(t, fmt.Sprintf("k8s.com/#!/log/%s/%s-worker-0/pod?namespace=mpi-namespace", jobNamespace, jobName), jobLogs[0].Uri)
	assert.Equal(t, fmt.Sprintf("k8s.com/#!/log/%s/%s-worker-1/pod?namespace=mpi-namespace", jobNamespace, jobName), jobLogs[1].Uri)
}

func TestGetProperties(t *testing.T) {
	mpiResourceHandler := mpiOperatorResourceHandler{}
	expected := k8s.PluginProperties{}
	assert.Equal(t, expected, mpiResourceHandler.GetProperties())
}
