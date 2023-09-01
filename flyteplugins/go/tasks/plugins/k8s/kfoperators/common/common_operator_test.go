package common

import (
	"fmt"
	"testing"
	"time"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyteplugins/go/tasks/logs"

	pluginsCore "github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/core/mocks"
	commonOp "github.com/kubeflow/common/pkg/apis/common/v1"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestExtractCurrentCondition(t *testing.T) {
	jobCreated := commonOp.JobCondition{
		Type:   commonOp.JobCreated,
		Status: corev1.ConditionTrue,
	}
	jobRunningActive := commonOp.JobCondition{
		Type:   commonOp.JobRunning,
		Status: corev1.ConditionFalse,
	}
	jobConditions := []commonOp.JobCondition{
		jobCreated,
		jobRunningActive,
	}
	currentCondition, err := ExtractCurrentCondition(jobConditions)
	assert.NoError(t, err)
	assert.Equal(t, currentCondition, jobCreated)

	jobConditions = nil
	currentCondition, err = ExtractCurrentCondition(jobConditions)
	assert.NoError(t, err)
	assert.Equal(t, currentCondition, commonOp.JobCondition{})

	currentCondition, err = ExtractCurrentCondition(nil)
	assert.NoError(t, err)
	assert.Equal(t, currentCondition, commonOp.JobCondition{})

	jobUnknown := commonOp.JobCondition{Type: "unknown"}
	jobConditions = []commonOp.JobCondition{jobUnknown}
	currentCondition, err = ExtractCurrentCondition(jobConditions)
	assert.Error(t, err)
	assert.Equal(t, currentCondition, commonOp.JobCondition{})
	assert.Equal(t, err, fmt.Errorf("found no current condition. Conditions: %+v", jobConditions))
}

func TestGetPhaseInfo(t *testing.T) {
	jobCreating := commonOp.JobCondition{}
	taskPhase, err := GetPhaseInfo(jobCreating, time.Now(), pluginsCore.TaskInfo{})
	assert.NoError(t, err)
	assert.Equal(t, pluginsCore.PhaseQueued, taskPhase.Phase())
	assert.NotNil(t, taskPhase.Info())
	assert.Nil(t, err)

	jobCreated := commonOp.JobCondition{
		Type: commonOp.JobCreated,
	}
	taskPhase, err = GetPhaseInfo(jobCreated, time.Now(), pluginsCore.TaskInfo{})
	assert.NoError(t, err)
	assert.Equal(t, pluginsCore.PhaseQueued, taskPhase.Phase())
	assert.NotNil(t, taskPhase.Info())
	assert.Nil(t, err)

	jobSucceeded := commonOp.JobCondition{
		Type: commonOp.JobSucceeded,
	}
	taskPhase, err = GetPhaseInfo(jobSucceeded, time.Now(), pluginsCore.TaskInfo{})
	assert.NoError(t, err)
	assert.Equal(t, pluginsCore.PhaseSuccess, taskPhase.Phase())
	assert.NotNil(t, taskPhase.Info())
	assert.Nil(t, err)

	jobFailed := commonOp.JobCondition{
		Type: commonOp.JobFailed,
	}
	taskPhase, err = GetPhaseInfo(jobFailed, time.Now(), pluginsCore.TaskInfo{})
	assert.NoError(t, err)
	assert.Equal(t, pluginsCore.PhaseRetryableFailure, taskPhase.Phase())
	assert.NotNil(t, taskPhase.Info())
	assert.Nil(t, err)

	jobRestarting := commonOp.JobCondition{
		Type: commonOp.JobRestarting,
	}
	taskPhase, err = GetPhaseInfo(jobRestarting, time.Now(), pluginsCore.TaskInfo{})
	assert.NoError(t, err)
	assert.Equal(t, pluginsCore.PhaseRunning, taskPhase.Phase())
	assert.NotNil(t, taskPhase.Info())
	assert.Nil(t, err)

	jobRestarting = commonOp.JobCondition{
		Type: commonOp.JobRunning,
	}
	taskPhase, err = GetPhaseInfo(jobRestarting, time.Now(), pluginsCore.TaskInfo{})
	assert.NoError(t, err)
	assert.Equal(t, pluginsCore.PhaseRunning, taskPhase.Phase())
	assert.NotNil(t, taskPhase.Info())
	assert.Nil(t, err)
}

func TestGetMPIPhaseInfo(t *testing.T) {
	jobCreated := commonOp.JobCondition{
		Type: commonOp.JobCreated,
	}
	taskPhase, err := GetMPIPhaseInfo(jobCreated, time.Now(), pluginsCore.TaskInfo{})
	assert.NoError(t, err)
	assert.Equal(t, pluginsCore.PhaseQueued, taskPhase.Phase())
	assert.NotNil(t, taskPhase.Info())
	assert.Nil(t, err)

	jobSucceeded := commonOp.JobCondition{
		Type: commonOp.JobSucceeded,
	}
	taskPhase, err = GetMPIPhaseInfo(jobSucceeded, time.Now(), pluginsCore.TaskInfo{})
	assert.NoError(t, err)
	assert.Equal(t, pluginsCore.PhaseSuccess, taskPhase.Phase())
	assert.NotNil(t, taskPhase.Info())
	assert.Nil(t, err)

	jobFailed := commonOp.JobCondition{
		Type: commonOp.JobFailed,
	}
	taskPhase, err = GetMPIPhaseInfo(jobFailed, time.Now(), pluginsCore.TaskInfo{})
	assert.NoError(t, err)
	assert.Equal(t, pluginsCore.PhaseRetryableFailure, taskPhase.Phase())
	assert.NotNil(t, taskPhase.Info())
	assert.Nil(t, err)

	jobRestarting := commonOp.JobCondition{
		Type: commonOp.JobRestarting,
	}
	taskPhase, err = GetMPIPhaseInfo(jobRestarting, time.Now(), pluginsCore.TaskInfo{})
	assert.NoError(t, err)
	assert.Equal(t, pluginsCore.PhaseRunning, taskPhase.Phase())
	assert.NotNil(t, taskPhase.Info())
	assert.Nil(t, err)

	jobRestarting = commonOp.JobCondition{
		Type: commonOp.JobRunning,
	}
	taskPhase, err = GetMPIPhaseInfo(jobRestarting, time.Now(), pluginsCore.TaskInfo{})
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

	workers := int32(1)
	launcher := int32(1)

	taskCtx := dummyTaskContext()
	mpiJobObjectMeta := meta_v1.ObjectMeta{
		Name:      "test",
		Namespace: "mpi-namespace",
	}
	jobLogs, err := GetLogs(taskCtx, MPITaskType, mpiJobObjectMeta, false, workers, launcher, 0)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(jobLogs))
	assert.Equal(t, fmt.Sprintf("k8s.com/#!/log/%s/%s-worker-0/pod?namespace=mpi-namespace", "mpi-namespace", "test"), jobLogs[0].Uri)

	pytorchJobObjectMeta := meta_v1.ObjectMeta{
		Name:      "test",
		Namespace: "pytorch-namespace",
	}
	jobLogs, err = GetLogs(taskCtx, PytorchTaskType, pytorchJobObjectMeta, true, workers, launcher, 0)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(jobLogs))
	assert.Equal(t, fmt.Sprintf("k8s.com/#!/log/%s/%s-master-0/pod?namespace=pytorch-namespace", "pytorch-namespace", "test"), jobLogs[0].Uri)
	assert.Equal(t, fmt.Sprintf("k8s.com/#!/log/%s/%s-worker-0/pod?namespace=pytorch-namespace", "pytorch-namespace", "test"), jobLogs[1].Uri)

	tensorflowJobObjectMeta := meta_v1.ObjectMeta{
		Name:      "test",
		Namespace: "tensorflow-namespace",
	}
	jobLogs, err = GetLogs(taskCtx, TensorflowTaskType, tensorflowJobObjectMeta, false, workers, launcher, 1)
	assert.NoError(t, err)
	assert.Equal(t, 3, len(jobLogs))
	assert.Equal(t, fmt.Sprintf("k8s.com/#!/log/%s/%s-worker-0/pod?namespace=tensorflow-namespace", "tensorflow-namespace", "test"), jobLogs[0].Uri)
	assert.Equal(t, fmt.Sprintf("k8s.com/#!/log/%s/%s-psReplica-0/pod?namespace=tensorflow-namespace", "tensorflow-namespace", "test"), jobLogs[1].Uri)
	assert.Equal(t, fmt.Sprintf("k8s.com/#!/log/%s/%s-chiefReplica-0/pod?namespace=tensorflow-namespace", "tensorflow-namespace", "test"), jobLogs[2].Uri)

}

func TestGetLogsTemplateUri(t *testing.T) {
	assert.NoError(t, logs.SetLogConfig(&logs.LogConfig{
		IsStackDriverEnabled:   true,
		StackDriverTemplateURI: "https://console.cloud.google.com/logs/query;query=resource.labels.pod_name={{.podName}}&timestamp>{{.podRFC3339StartTime}}",
	}))

	taskCtx := dummyTaskContext()
	pytorchJobObjectMeta := meta_v1.ObjectMeta{
		Name:      "test",
		Namespace: "pytorch-namespace",
		CreationTimestamp: meta_v1.Time{
			Time: time.Date(2022, time.January, 1, 12, 0, 0, 0, time.UTC),
		},
	}
	jobLogs, err := GetLogs(taskCtx, PytorchTaskType, pytorchJobObjectMeta, true, 1, 0, 0)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(jobLogs))
	assert.Equal(t, fmt.Sprintf("https://console.cloud.google.com/logs/query;query=resource.labels.pod_name=%s-master-0&timestamp>%s", "test", "2022-01-01T12:00:00Z"), jobLogs[0].Uri)
	assert.Equal(t, fmt.Sprintf("https://console.cloud.google.com/logs/query;query=resource.labels.pod_name=%s-worker-0&timestamp>%s", "test", "2022-01-01T12:00:00Z"), jobLogs[1].Uri)
}

func dummyPodSpec() v1.PodSpec {
	return v1.PodSpec{
		Containers: []v1.Container{
			{
				Name: "primary container",
				Args: []string{"pyflyte-execute", "--task-module", "tests.flytekit.unit.sdk.tasks.test_sidecar_tasks", "--task-name", "simple_sidecar_task", "--inputs", "{{.input}}", "--output-prefix", "{{.outputPrefix}}"},
				Resources: v1.ResourceRequirements{
					Limits: v1.ResourceList{
						"cpu":    resource.MustParse("2"),
						"memory": resource.MustParse("200Mi"),
						"gpu":    resource.MustParse("1"),
					},
					Requests: v1.ResourceList{
						"cpu":    resource.MustParse("1"),
						"memory": resource.MustParse("100Mi"),
						"gpu":    resource.MustParse("1"),
					},
				},
				VolumeMounts: []v1.VolumeMount{
					{
						Name: "volume mount",
					},
				},
			},
			{
				Name: "secondary container",
				Resources: v1.ResourceRequirements{
					Limits: v1.ResourceList{
						"gpu": resource.MustParse("2"),
					},
					Requests: v1.ResourceList{
						"gpu": resource.MustParse("2"),
					},
				},
			},
		},
		Volumes: []v1.Volume{
			{
				Name: "dshm",
			},
		},
		Tolerations: []v1.Toleration{
			{
				Key:   "my toleration key",
				Value: "my toleration value",
			},
		},
	}
}

func TestOverrideContainerSpec(t *testing.T) {
	podSpec := dummyPodSpec()
	err := OverrideContainerSpec(
		&podSpec, "primary container", "testing-image",
		&core.Resources{
			Requests: []*core.Resources_ResourceEntry{
				{Name: core.Resources_CPU, Value: "250m"},
			},
			Limits: []*core.Resources_ResourceEntry{
				{Name: core.Resources_CPU, Value: "500m"},
			},
		},
		[]string{"python", "-m", "run.py"},
	)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(podSpec.Containers))
	assert.Equal(t, "testing-image", podSpec.Containers[0].Image)
	assert.NotNil(t, podSpec.Containers[0].Resources.Limits)
	assert.NotNil(t, podSpec.Containers[0].Resources.Requests)
	// verify resources not overridden if empty resources
	assert.True(t, podSpec.Containers[0].Resources.Requests.Cpu().Equal(resource.MustParse("250m")))
	assert.True(t, podSpec.Containers[0].Resources.Limits.Cpu().Equal(resource.MustParse("500m")))
	assert.Equal(t, []string{"python", "-m", "run.py"}, podSpec.Containers[0].Args)
}

func TestOverrideContainerSpecEmptyFields(t *testing.T) {
	podSpec := dummyPodSpec()
	err := OverrideContainerSpec(&podSpec, "primary container", "", &core.Resources{}, []string{})
	assert.NoError(t, err)
	assert.Equal(t, 2, len(podSpec.Containers))
	assert.NotNil(t, podSpec.Containers[0].Resources.Limits)
	assert.NotNil(t, podSpec.Containers[0].Resources.Requests)
	// verify resources not overridden if empty resources
	assert.True(t, podSpec.Containers[0].Resources.Requests.Cpu().Equal(resource.MustParse("1")))
	assert.True(t, podSpec.Containers[0].Resources.Requests.Memory().Equal(resource.MustParse("100Mi")))
	assert.True(t, podSpec.Containers[0].Resources.Limits.Cpu().Equal(resource.MustParse("2")))
	assert.True(t, podSpec.Containers[0].Resources.Limits.Memory().Equal(resource.MustParse("200Mi")))
}

func TestOverrideContainerNilResources(t *testing.T) {
	podSpec := dummyPodSpec()
	podSpecCopy := podSpec.DeepCopy()

	err := OverrideContainerSpec(&podSpec, "primary container", "", nil, []string{})
	assert.NoError(t, err)
	assert.Equal(t, 2, len(podSpec.Containers))
	assert.Equal(t, podSpec.Containers[0].Resources.Limits, podSpecCopy.Containers[0].Resources.Limits)
	assert.Equal(t, podSpec.Containers[0].Resources.Requests, podSpecCopy.Containers[0].Resources.Requests)
}

func dummyTaskContext() pluginsCore.TaskExecutionContext {
	taskCtx := &mocks.TaskExecutionContext{}

	tID := &mocks.TaskExecutionID{}
	tID.OnGetID().Return(core.TaskExecutionIdentifier{
		TaskId: &core.Identifier{
			ResourceType: core.ResourceType_TASK,
			Name:         "my-task-name",
			Project:      "my-task-project",
			Domain:       "my-task-domain",
			Version:      "1",
		},
		NodeExecutionId: &core.NodeExecutionIdentifier{
			ExecutionId: &core.WorkflowExecutionIdentifier{
				Name:    "my-execution-name",
				Project: "my-execution-project",
				Domain:  "my-execution-domain",
			},
		},
		RetryAttempt: 0,
	})

	taskExecutionMetadata := &mocks.TaskExecutionMetadata{}
	taskExecutionMetadata.OnGetTaskExecutionID().Return(tID)
	taskCtx.OnTaskExecutionMetadata().Return(taskExecutionMetadata)
	return taskCtx
}
