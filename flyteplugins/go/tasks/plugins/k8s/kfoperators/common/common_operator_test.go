package common

import (
	"fmt"
	"os"
	"testing"
	"time"

	kubeflowv1 "github.com/kubeflow/training-operator/pkg/apis/kubeflow.org/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/structpb"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/logs"
	pluginsCore "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/core/mocks"
	k8smocks "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/k8s/mocks"
	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/tasklog"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/core"
)

func TestMain(m *testing.M) {
	// All tests should run assuming UTC timezone.
	time.Local = time.UTC
	code := m.Run()
	os.Exit(code)
}

func TestExtractCurrentCondition(t *testing.T) {
	jobCreated := kubeflowv1.JobCondition{
		Type:   kubeflowv1.JobCreated,
		Status: corev1.ConditionTrue,
	}
	jobRunningActive := kubeflowv1.JobCondition{
		Type:   kubeflowv1.JobRunning,
		Status: corev1.ConditionFalse,
	}
	jobConditions := []kubeflowv1.JobCondition{
		jobCreated,
		jobRunningActive,
	}
	currentCondition, err := ExtractCurrentCondition(jobConditions)
	assert.NoError(t, err)
	assert.Equal(t, currentCondition, jobCreated)
	assert.Equal(t, currentCondition, jobCreated)

	jobConditions = nil
	currentCondition, err = ExtractCurrentCondition(jobConditions)
	assert.NoError(t, err)
	assert.Equal(t, currentCondition, kubeflowv1.JobCondition{})

	currentCondition, err = ExtractCurrentCondition(nil)
	assert.NoError(t, err)
	assert.Equal(t, currentCondition, kubeflowv1.JobCondition{})

	jobUnknown := kubeflowv1.JobCondition{Type: "unknown"}
	jobConditions = []kubeflowv1.JobCondition{jobUnknown}
	currentCondition, err = ExtractCurrentCondition(jobConditions)
	assert.Error(t, err)
	assert.Equal(t, currentCondition, kubeflowv1.JobCondition{})
	assert.Equal(t, currentCondition, kubeflowv1.JobCondition{})
	assert.Equal(t, err, fmt.Errorf("found no current condition. Conditions: %+v", jobConditions))
}

func TestGetPhaseInfo(t *testing.T) {
	jobCreating := kubeflowv1.JobCondition{}
	taskPhase, err := GetPhaseInfo(jobCreating, time.Now(), pluginsCore.TaskInfo{})
	assert.NoError(t, err)
	assert.Equal(t, pluginsCore.PhaseQueued, taskPhase.Phase())
	assert.NotNil(t, taskPhase.Info())
	assert.Nil(t, err)

	jobCreated := kubeflowv1.JobCondition{
		Type: kubeflowv1.JobCreated,
	}
	taskPhase, err = GetPhaseInfo(jobCreated, time.Now(), pluginsCore.TaskInfo{})
	assert.NoError(t, err)
	assert.Equal(t, pluginsCore.PhaseQueued, taskPhase.Phase())
	assert.NotNil(t, taskPhase.Info())
	assert.Nil(t, err)

	jobSucceeded := kubeflowv1.JobCondition{
		Type: kubeflowv1.JobSucceeded,
	}
	taskPhase, err = GetPhaseInfo(jobSucceeded, time.Now(), pluginsCore.TaskInfo{})
	assert.NoError(t, err)
	assert.Equal(t, pluginsCore.PhaseSuccess, taskPhase.Phase())
	assert.NotNil(t, taskPhase.Info())
	assert.Nil(t, err)

	jobFailed := kubeflowv1.JobCondition{
		Type: kubeflowv1.JobFailed,
	}
	taskPhase, err = GetPhaseInfo(jobFailed, time.Now(), pluginsCore.TaskInfo{})
	assert.NoError(t, err)
	assert.Equal(t, pluginsCore.PhaseRetryableFailure, taskPhase.Phase())
	assert.NotNil(t, taskPhase.Info())
	assert.Nil(t, err)

	jobRestarting := kubeflowv1.JobCondition{
		Type: kubeflowv1.JobRestarting,
	}
	taskPhase, err = GetPhaseInfo(jobRestarting, time.Now(), pluginsCore.TaskInfo{})
	assert.NoError(t, err)
	assert.Equal(t, pluginsCore.PhaseRunning, taskPhase.Phase())
	assert.NotNil(t, taskPhase.Info())
	assert.Nil(t, err)

	jobRestarting = kubeflowv1.JobCondition{
		Type: kubeflowv1.JobRunning,
	}
	taskPhase, err = GetPhaseInfo(jobRestarting, time.Now(), pluginsCore.TaskInfo{})
	assert.NoError(t, err)
	assert.Equal(t, pluginsCore.PhaseRunning, taskPhase.Phase())
	assert.NotNil(t, taskPhase.Info())
	assert.Nil(t, err)
}

func TestGetMPIPhaseInfo(t *testing.T) {
	jobCreated := kubeflowv1.JobCondition{
		Type: kubeflowv1.JobCreated,
	}
	taskPhase, err := GetMPIPhaseInfo(jobCreated, time.Now(), pluginsCore.TaskInfo{})
	assert.NoError(t, err)
	assert.Equal(t, pluginsCore.PhaseQueued, taskPhase.Phase())
	assert.NotNil(t, taskPhase.Info())
	assert.Nil(t, err)

	jobSucceeded := kubeflowv1.JobCondition{
		Type: kubeflowv1.JobSucceeded,
	}
	taskPhase, err = GetMPIPhaseInfo(jobSucceeded, time.Now(), pluginsCore.TaskInfo{})
	assert.NoError(t, err)
	assert.Equal(t, pluginsCore.PhaseSuccess, taskPhase.Phase())
	assert.NotNil(t, taskPhase.Info())
	assert.Nil(t, err)

	jobFailed := kubeflowv1.JobCondition{
		Type: kubeflowv1.JobFailed,
	}
	taskPhase, err = GetMPIPhaseInfo(jobFailed, time.Now(), pluginsCore.TaskInfo{})
	assert.NoError(t, err)
	assert.Equal(t, pluginsCore.PhaseRetryableFailure, taskPhase.Phase())
	assert.NotNil(t, taskPhase.Info())
	assert.Nil(t, err)

	jobRestarting := kubeflowv1.JobCondition{
		Type: kubeflowv1.JobRestarting,
	}
	taskPhase, err = GetMPIPhaseInfo(jobRestarting, time.Now(), pluginsCore.TaskInfo{})
	assert.NoError(t, err)
	assert.Equal(t, pluginsCore.PhaseRunning, taskPhase.Phase())
	assert.NotNil(t, taskPhase.Info())
	assert.Nil(t, err)

	jobRestarting = kubeflowv1.JobCondition{
		Type: kubeflowv1.JobRunning,
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

	taskTemplate := dummyTaskTemplate()
	taskCtx := dummyTaskContext()
	mpiJobObjectMeta := meta_v1.ObjectMeta{
		Name:      "test",
		Namespace: "mpi-namespace",
	}
	jobLogs, err := GetLogs(taskCtx, MPITaskType, mpiJobObjectMeta, taskTemplate, false, workers, launcher, 0, 0, kubeflowv1.MPIJobDefaultContainerName)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(jobLogs))
	assert.Equal(t, fmt.Sprintf("k8s.com/#!/log/%s/%s-worker-0/pod?namespace=mpi-namespace", "mpi-namespace", "test"), jobLogs[0].Uri)

	pytorchJobObjectMeta := meta_v1.ObjectMeta{
		Name:      "test",
		Namespace: "pytorch-namespace",
	}
	jobLogs, err = GetLogs(taskCtx, PytorchTaskType, pytorchJobObjectMeta, taskTemplate, true, workers, launcher, 0, 0, kubeflowv1.PyTorchJobDefaultContainerName)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(jobLogs))
	assert.Equal(t, fmt.Sprintf("k8s.com/#!/log/%s/%s-master-0/pod?namespace=pytorch-namespace", "pytorch-namespace", "test"), jobLogs[0].Uri)
	assert.Equal(t, fmt.Sprintf("k8s.com/#!/log/%s/%s-worker-0/pod?namespace=pytorch-namespace", "pytorch-namespace", "test"), jobLogs[1].Uri)

	tensorflowJobObjectMeta := meta_v1.ObjectMeta{
		Name:      "test",
		Namespace: "tensorflow-namespace",
	}
	jobLogs, err = GetLogs(taskCtx, TensorflowTaskType, tensorflowJobObjectMeta, taskTemplate, false, workers, launcher, 1, 0, kubeflowv1.TFJobDefaultContainerName)
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

	taskTemplate := dummyTaskTemplate()
	taskCtx := dummyTaskContext()
	pytorchJobObjectMeta := meta_v1.ObjectMeta{
		Name: "test",
		Namespace: "pytorch-" +
			"namespace",
		CreationTimestamp: meta_v1.Time{
			Time: time.Date(2022, time.January, 1, 12, 0, 0, 0, time.UTC),
		},
	}
	jobLogs, err := GetLogs(taskCtx, PytorchTaskType, pytorchJobObjectMeta, taskTemplate, true, 1, 0, 0, 0, kubeflowv1.PyTorchJobDefaultContainerName)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(jobLogs))
	assert.Equal(t, fmt.Sprintf("https://console.cloud.google.com/logs/query;query=resource.labels.pod_name=%s-master-0&timestamp>%s", "test", "2022-01-01T12:00:00Z"), jobLogs[0].Uri)
	assert.Equal(t, fmt.Sprintf("https://console.cloud.google.com/logs/query;query=resource.labels.pod_name=%s-worker-0&timestamp>%s", "test", "2022-01-01T12:00:00Z"), jobLogs[1].Uri)
}

func TestGetLogsDynamic(t *testing.T) {
	dynamicLinks := map[string]tasklog.TemplateLogPlugin{
		"test-dynamic-link": {
			TemplateURIs: []string{"https://some-service.com/{{.taskConfig.dynamicParam}}"},
		},
	}

	assert.NoError(t, logs.SetLogConfig(&logs.LogConfig{
		DynamicLogLinks: dynamicLinks,
	}))

	taskTemplate := dummyTaskTemplate()
	taskTemplate.Config = map[string]string{
		"link_type":    "test-dynamic-link",
		"dynamicParam": "dynamic-value",
	}
	taskCtx := dummyTaskContext()
	pytorchJobObjectMeta := meta_v1.ObjectMeta{
		Name: "test",
		Namespace: "pytorch-" +
			"namespace",
		CreationTimestamp: meta_v1.Time{
			Time: time.Date(2022, time.January, 1, 12, 0, 0, 0, time.UTC),
		},
	}
	jobLogs, err := GetLogs(taskCtx, PytorchTaskType, pytorchJobObjectMeta, taskTemplate, true, 1, 0, 0, 0, kubeflowv1.PyTorchJobDefaultContainerName)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(jobLogs))
	assert.Equal(t, "https://some-service.com/dynamic-value", jobLogs[0].GetUri())
}

func dummyPodSpec() v1.PodSpec {
	return v1.PodSpec{
		Containers: []v1.Container{
			{
				Name:  "primary container",
				Args:  []string{"pyflyte-execute", "--task-module", "tests.flytekit.unit.sdk.tasks.test_sidecar_tasks", "--task-name", "simple_sidecar_task", "--inputs", "{{.input}}", "--output-prefix", "{{.outputPrefix}}"},
				Image: "dummy-image",
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
		[]string{"python", "-m", "run.py"},
	)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(podSpec.Containers))
	assert.Equal(t, "testing-image", podSpec.Containers[0].Image)
	assert.Equal(t, []string{"python", "-m", "run.py"}, podSpec.Containers[0].Args)
}

func TestOverrideContainerSpecEmptyFields(t *testing.T) {
	podSpec := dummyPodSpec()
	err := OverrideContainerSpec(&podSpec, "primary container", "", []string{})
	assert.NoError(t, err)
	assert.Equal(t, 2, len(podSpec.Containers))
	assert.Equal(t, "dummy-image", podSpec.Containers[0].Image)
	assert.Equal(t, []string{"pyflyte-execute", "--task-module", "tests.flytekit.unit.sdk.tasks.test_sidecar_tasks", "--task-name", "simple_sidecar_task", "--inputs", "{{.input}}", "--output-prefix", "{{.outputPrefix}}"}, podSpec.Containers[0].Args)
}

func dummyTaskTemplate() *core.TaskTemplate {
	id := "dummy-id"

	testImage := "dummy-image"

	structObj := structpb.Struct{}

	return &core.TaskTemplate{
		Id:   &core.Identifier{Name: id},
		Type: "container",
		Target: &core.TaskTemplate_Container{
			Container: &core.Container{
				Image: testImage,
			},
		},
		Custom: &structObj,
	}
}

func dummyTaskContext() *k8smocks.PluginContext {
	pCtx := &k8smocks.PluginContext{}

	tID := &mocks.TaskExecutionID{}
	tID.EXPECT().GetID().Return(core.TaskExecutionIdentifier{
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
	tID.EXPECT().GetGeneratedName().Return("some-acceptable-name")
	tID.EXPECT().GetUniqueNodeID().Return("an-unique-id")

	taskExecutionMetadata := &mocks.TaskExecutionMetadata{}
	taskExecutionMetadata.EXPECT().GetTaskExecutionID().Return(tID)
	pCtx.On("TaskExecutionMetadata").Return(taskExecutionMetadata)
	return pCtx
}
