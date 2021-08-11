package sidecar

import (
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"path"
	"testing"

	"sigs.k8s.io/controller-runtime/pkg/client"

	errors2 "github.com/flyteorg/flyteplugins/go/tasks/errors"

	"github.com/flyteorg/flytestdlib/storage"
	"github.com/stretchr/testify/mock"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	structpb "github.com/golang/protobuf/ptypes/struct"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"

	pluginsCore "github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/core"
	pluginsCoreMock "github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/core/mocks"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/flytek8s/config"
	pluginsIOMock "github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/io/mocks"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/k8s"
)

const ResourceNvidiaGPU = "nvidia.com/gpu"

var resourceRequirements = &v1.ResourceRequirements{
	Limits: v1.ResourceList{
		v1.ResourceCPU:              resource.MustParse("2048m"),
		v1.ResourceEphemeralStorage: resource.MustParse("100M"),
	},
}

func getSidecarTaskTemplateForTest(sideCarJob sidecarJob) *core.TaskTemplate {
	sidecarJSON, err := json.Marshal(&sideCarJob)
	if err != nil {
		panic(err)
	}
	structObj := structpb.Struct{}
	err = json.Unmarshal(sidecarJSON, &structObj)
	if err != nil {
		panic(err)
	}
	return &core.TaskTemplate{
		Custom: &structObj,
	}
}

func dummyContainerTaskMetadata(resources *v1.ResourceRequirements) pluginsCore.TaskExecutionMetadata {
	taskMetadata := &pluginsCoreMock.TaskExecutionMetadata{}
	taskMetadata.On("GetNamespace").Return("test-namespace")
	taskMetadata.On("GetAnnotations").Return(map[string]string{"annotation-1": "val1"})

	taskMetadata.On("GetLabels").Return(map[string]string{"label-1": "val1"})
	taskMetadata.On("GetOwnerReference").Return(metav1.OwnerReference{
		Kind: "node",
		Name: "blah",
	})
	taskMetadata.On("IsInterruptible").Return(true)
	taskMetadata.On("GetSecurityContext").Return(core.SecurityContext{})
	taskMetadata.On("GetK8sServiceAccount").Return("service-account")
	taskMetadata.On("GetOwnerID").Return(types.NamespacedName{
		Namespace: "test-namespace",
		Name:      "test-owner-name",
	})

	tID := &pluginsCoreMock.TaskExecutionID{}
	tID.On("GetID").Return(core.TaskExecutionIdentifier{
		NodeExecutionId: &core.NodeExecutionIdentifier{
			ExecutionId: &core.WorkflowExecutionIdentifier{
				Name:    "my_name",
				Project: "my_project",
				Domain:  "my_domain",
			},
		},
	})
	tID.On("GetGeneratedName").Return("my_project:my_domain:my_name")
	taskMetadata.On("GetTaskExecutionID").Return(tID)

	to := &pluginsCoreMock.TaskOverrides{}
	to.On("GetResources").Return(resources)
	taskMetadata.On("GetOverrides").Return(to)

	return taskMetadata
}

func getDummySidecarTaskContext(taskTemplate *core.TaskTemplate, resources *v1.ResourceRequirements) pluginsCore.TaskExecutionContext {
	taskCtx := &pluginsCoreMock.TaskExecutionContext{}
	dummyTaskMetadata := dummyContainerTaskMetadata(resources)
	inputReader := &pluginsIOMock.InputReader{}
	inputReader.On("GetInputPrefixPath").Return(storage.DataReference("test-data-prefix"))
	inputReader.On("GetInputPath").Return(storage.DataReference("test-data-reference"))
	inputReader.On("Get", mock.Anything).Return(&core.LiteralMap{}, nil)
	taskCtx.On("InputReader").Return(inputReader)

	outputReader := &pluginsIOMock.OutputWriter{}
	outputReader.On("GetOutputPath").Return(storage.DataReference("/data/outputs.pb"))
	outputReader.On("GetOutputPrefixPath").Return(storage.DataReference("/data/"))
	outputReader.On("GetRawOutputPrefix").Return(storage.DataReference(""))
	taskCtx.On("OutputWriter").Return(outputReader)

	taskReader := &pluginsCoreMock.TaskReader{}
	taskReader.On("Read", mock.Anything).Return(taskTemplate, nil)
	taskCtx.On("TaskReader").Return(taskReader)

	taskCtx.On("TaskExecutionMetadata").Return(dummyTaskMetadata)

	return taskCtx
}

func getPodSpec() v1.PodSpec {
	return v1.PodSpec{
		Containers: []v1.Container{
			{
				Name: "primary container",
				Args: []string{"pyflyte-execute", "--task-module", "tests.flytekit.unit.sdk.tasks.test_sidecar_tasks", "--task-name", "simple_sidecar_task", "--inputs", "{{.input}}", "--output-prefix", "{{.outputPrefix}}"},
				Resources: v1.ResourceRequirements{
					Limits: v1.ResourceList{
						"cpu":    resource.MustParse("2"),
						"memory": resource.MustParse("200Mi"),
					},
					Requests: v1.ResourceList{
						"cpu":    resource.MustParse("1"),
						"memory": resource.MustParse("100Mi"),
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

func checkUserTolerations(t *testing.T, res client.Object) {
	// Assert user-specified tolerations don't get overridden
	assert.Len(t, res.(*v1.Pod).Spec.Tolerations, 1)
	for _, tol := range res.(*v1.Pod).Spec.Tolerations {
		if tol.Key == "my toleration key" {
			assert.Equal(t, tol.Value, "my toleration value")
		} else {
			t.Fatalf("unexpected toleration [%+v]", tol)
		}
	}
}

func TestBuildSidecarResource_TaskType2(t *testing.T) {
	podSpec := getPodSpec()

	b, err := json.Marshal(podSpec)
	if err != nil {
		t.Fatal(err)
	}

	structObj := &structpb.Struct{}
	if err := json.Unmarshal(b, structObj); err != nil {
		t.Fatal(err)
	}

	task := core.TaskTemplate{
		TaskTypeVersion: 2,
		Config: map[string]string{
			primaryContainerKey: "primary container",
		},
		Target: &core.TaskTemplate_K8SPod{
			K8SPod: &core.K8SPod{
				PodSpec: structObj,
				Metadata: &core.K8SObjectMetadata{
					Labels: map[string]string{
						"label": "foo",
					},
					Annotations: map[string]string{
						"anno": "bar",
					},
				},
			},
		},
	}

	tolGPU := v1.Toleration{
		Key:      "flyte/gpu",
		Value:    "dedicated",
		Operator: v1.TolerationOpEqual,
		Effect:   v1.TaintEffectNoSchedule,
	}

	tolStorage := v1.Toleration{
		Key:      "storage",
		Value:    "dedicated",
		Operator: v1.TolerationOpExists,
		Effect:   v1.TaintEffectNoSchedule,
	}
	assert.NoError(t, config.SetK8sPluginConfig(&config.K8sPluginConfig{
		ResourceTolerations: map[v1.ResourceName][]v1.Toleration{
			v1.ResourceStorage: {tolStorage},
			ResourceNvidiaGPU:  {tolGPU},
		},
		DefaultCPURequest:    "1024m",
		DefaultMemoryRequest: "1024Mi",
	}))
	handler := &sidecarResourceHandler{}
	taskCtx := getDummySidecarTaskContext(&task, resourceRequirements)
	res, err := handler.BuildResource(context.TODO(), taskCtx)
	assert.Nil(t, err)
	assert.EqualValues(t, map[string]string{
		primaryContainerKey: "primary container",
		"anno":              "bar",
	}, res.GetAnnotations())
	assert.EqualValues(t, map[string]string{
		"label": "foo",
	}, res.GetLabels())

	// Assert volumes & volume mounts are preserved
	assert.Len(t, res.(*v1.Pod).Spec.Volumes, 1)
	assert.Equal(t, "dshm", res.(*v1.Pod).Spec.Volumes[0].Name)

	assert.Len(t, res.(*v1.Pod).Spec.Containers[0].VolumeMounts, 1)
	assert.Equal(t, "volume mount", res.(*v1.Pod).Spec.Containers[0].VolumeMounts[0].Name)
	checkUserTolerations(t, res)

	// Assert resource requirements are correctly set
	expectedCPURequest := resource.MustParse("1")
	assert.Equal(t, expectedCPURequest.Value(), res.(*v1.Pod).Spec.Containers[0].Resources.Requests.Cpu().Value())
	expectedMemRequest := resource.MustParse("100Mi")
	assert.Equal(t, expectedMemRequest.Value(), res.(*v1.Pod).Spec.Containers[0].Resources.Requests.Memory().Value())
	expectedCPULimit := resource.MustParse("2048m")
	assert.Equal(t, expectedCPULimit.Value(), res.(*v1.Pod).Spec.Containers[0].Resources.Limits.Cpu().Value())
	expectedMemLimit := resource.MustParse("200Mi")
	assert.Equal(t, expectedMemLimit.Value(), res.(*v1.Pod).Spec.Containers[0].Resources.Limits.Memory().Value())
	expectedEphemeralStorageLimit := resource.MustParse("100M")
	assert.Equal(t, expectedEphemeralStorageLimit.Value(), res.(*v1.Pod).Spec.Containers[0].Resources.Limits.StorageEphemeral().Value())
}

func TestBuildSidecarResource_TaskType2_Invalid_Spec(t *testing.T) {
	task := core.TaskTemplate{
		TaskTypeVersion: 2,
		Config: map[string]string{
			primaryContainerKey: "primary container",
		},
		Target: &core.TaskTemplate_K8SPod{
			K8SPod: &core.K8SPod{
				Metadata: &core.K8SObjectMetadata{
					Labels: map[string]string{
						"label": "foo",
					},
					Annotations: map[string]string{
						"anno": "bar",
					},
				},
			},
		},
	}

	handler := &sidecarResourceHandler{}
	taskCtx := getDummySidecarTaskContext(&task, resourceRequirements)
	_, err := handler.BuildResource(context.TODO(), taskCtx)
	assert.EqualError(t, err, "[BadTaskSpecification] Pod tasks with task type version > 1 should specify their target as a K8sPod with a defined pod spec")
}

func TestBuildSidecarResource_TaskType1(t *testing.T) {
	podSpec := getPodSpec()

	b, err := json.Marshal(podSpec)
	if err != nil {
		t.Fatal(err)
	}

	structObj := &structpb.Struct{}
	if err := json.Unmarshal(b, structObj); err != nil {
		t.Fatal(err)
	}

	task := core.TaskTemplate{
		Custom:          structObj,
		TaskTypeVersion: 1,
		Config: map[string]string{
			primaryContainerKey: "primary container",
		},
	}

	tolGPU := v1.Toleration{
		Key:      "flyte/gpu",
		Value:    "dedicated",
		Operator: v1.TolerationOpEqual,
		Effect:   v1.TaintEffectNoSchedule,
	}

	tolStorage := v1.Toleration{
		Key:      "storage",
		Value:    "dedicated",
		Operator: v1.TolerationOpExists,
		Effect:   v1.TaintEffectNoSchedule,
	}
	assert.NoError(t, config.SetK8sPluginConfig(&config.K8sPluginConfig{
		ResourceTolerations: map[v1.ResourceName][]v1.Toleration{
			v1.ResourceStorage: {tolStorage},
			ResourceNvidiaGPU:  {tolGPU},
		},
		DefaultCPURequest:    "1024m",
		DefaultMemoryRequest: "1024Mi",
	}))
	handler := &sidecarResourceHandler{}
	taskCtx := getDummySidecarTaskContext(&task, resourceRequirements)
	res, err := handler.BuildResource(context.TODO(), taskCtx)
	assert.Nil(t, err)
	assert.EqualValues(t, map[string]string{
		primaryContainerKey: "primary container",
	}, res.GetAnnotations())
	assert.EqualValues(t, map[string]string{}, res.GetLabels())

	// Assert volumes & volume mounts are preserved
	assert.Len(t, res.(*v1.Pod).Spec.Volumes, 1)
	assert.Equal(t, "dshm", res.(*v1.Pod).Spec.Volumes[0].Name)

	assert.Len(t, res.(*v1.Pod).Spec.Containers[0].VolumeMounts, 1)
	assert.Equal(t, "volume mount", res.(*v1.Pod).Spec.Containers[0].VolumeMounts[0].Name)

	checkUserTolerations(t, res)

	// Assert resource requirements are correctly set
	expectedCPURequest := resource.MustParse("1")
	assert.Equal(t, expectedCPURequest.Value(), res.(*v1.Pod).Spec.Containers[0].Resources.Requests.Cpu().Value())
	expectedMemRequest := resource.MustParse("100Mi")
	assert.Equal(t, expectedMemRequest.Value(), res.(*v1.Pod).Spec.Containers[0].Resources.Requests.Memory().Value())
	expectedCPULimit := resource.MustParse("2048m")
	assert.Equal(t, expectedCPULimit.Value(), res.(*v1.Pod).Spec.Containers[0].Resources.Limits.Cpu().Value())
	expectedMemLimit := resource.MustParse("200Mi")
	assert.Equal(t, expectedMemLimit.Value(), res.(*v1.Pod).Spec.Containers[0].Resources.Limits.Memory().Value())
	expectedEphemeralStorageLimit := resource.MustParse("100M")
	assert.Equal(t, expectedEphemeralStorageLimit.Value(), res.(*v1.Pod).Spec.Containers[0].Resources.Limits.StorageEphemeral().Value())
}

func TestBuildSideResource_TaskType1_InvalidSpec(t *testing.T) {
	podSpec := v1.PodSpec{
		Containers: []v1.Container{
			{
				Name: "primary container",
			},
			{
				Name: "secondary container",
			},
		},
	}

	b, err := json.Marshal(podSpec)
	if err != nil {
		t.Fatal(err)
	}

	structObj := &structpb.Struct{}
	if err := json.Unmarshal(b, structObj); err != nil {
		t.Fatal(err)
	}

	task := core.TaskTemplate{
		Custom:          structObj,
		TaskTypeVersion: 1,
	}

	assert.NoError(t, config.SetK8sPluginConfig(&config.K8sPluginConfig{
		ResourceTolerations: map[v1.ResourceName][]v1.Toleration{
			v1.ResourceStorage: {},
			ResourceNvidiaGPU:  {},
		},
		DefaultCPURequest:    "1024m",
		DefaultMemoryRequest: "1024Mi",
	}))
	handler := &sidecarResourceHandler{}
	taskCtx := getDummySidecarTaskContext(&task, resourceRequirements)
	_, err = handler.BuildResource(context.TODO(), taskCtx)
	assert.EqualError(t, err, "[BadTaskSpecification] invalid TaskSpecification, config needs to be non-empty and include missing [primary_container_name] key")

	task.Config = map[string]string{
		"foo": "bar",
	}
	taskCtx = getDummySidecarTaskContext(&task, resourceRequirements)
	_, err = handler.BuildResource(context.TODO(), taskCtx)
	assert.EqualError(t, err, "[BadTaskSpecification] invalid TaskSpecification, config missing [primary_container_name] key in [map[foo:bar]]")

}

func TestBuildSidecarResource(t *testing.T) {
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	sidecarCustomJSON, err := ioutil.ReadFile(path.Join(dir, "testdata", "sidecar_custom"))
	if err != nil {
		t.Fatal(sidecarCustomJSON)
	}
	sidecarCustom := structpb.Struct{}
	if err := json.Unmarshal(sidecarCustomJSON, &sidecarCustom); err != nil {
		t.Fatal(err)
	}
	task := core.TaskTemplate{
		Custom: &sidecarCustom,
	}

	tolGPU := v1.Toleration{
		Key:      "flyte/gpu",
		Value:    "dedicated",
		Operator: v1.TolerationOpEqual,
		Effect:   v1.TaintEffectNoSchedule,
	}

	tolStorage := v1.Toleration{
		Key:      "storage",
		Value:    "dedicated",
		Operator: v1.TolerationOpExists,
		Effect:   v1.TaintEffectNoSchedule,
	}
	assert.NoError(t, config.SetK8sPluginConfig(&config.K8sPluginConfig{
		ResourceTolerations: map[v1.ResourceName][]v1.Toleration{
			v1.ResourceStorage: {tolStorage},
			ResourceNvidiaGPU:  {tolGPU},
		},
		DefaultCPURequest:    "1024m",
		DefaultMemoryRequest: "1024Mi",
	}))
	handler := &sidecarResourceHandler{}
	taskCtx := getDummySidecarTaskContext(&task, resourceRequirements)
	res, err := handler.BuildResource(context.TODO(), taskCtx)
	assert.Nil(t, err)
	assert.EqualValues(t, map[string]string{
		primaryContainerKey: "a container",
		"a1":                "a1",
	}, res.GetAnnotations())

	assert.EqualValues(t, map[string]string{
		"b1": "b1",
	}, res.GetLabels())

	// Assert volumes & volume mounts are preserved
	assert.Len(t, res.(*v1.Pod).Spec.Volumes, 1)
	assert.Equal(t, "dshm", res.(*v1.Pod).Spec.Volumes[0].Name)

	assert.Len(t, res.(*v1.Pod).Spec.Containers[0].VolumeMounts, 1)
	assert.Equal(t, "volume mount", res.(*v1.Pod).Spec.Containers[0].VolumeMounts[0].Name)

	assert.Equal(t, "service-account", res.(*v1.Pod).Spec.ServiceAccountName)
	// Assert user-specified tolerations don't get overridden
	assert.Len(t, res.(*v1.Pod).Spec.Tolerations, 1)
	for _, tol := range res.(*v1.Pod).Spec.Tolerations {
		if tol.Key == "flyte/gpu" {
			assert.Equal(t, tol.Value, "dedicated")
			assert.Equal(t, tol.Operator, v1.TolerationOperator("Equal"))
			assert.Equal(t, tol.Effect, v1.TaintEffect("NoSchedule"))
		} else if tol.Key == "my toleration key" {
			assert.Equal(t, tol.Value, "my toleration value")
		} else {
			t.Fatalf("unexpected toleration [%+v]", tol)
		}
	}

	// Assert resource requirements are correctly set
	expectedCPURequest := resource.MustParse("1024m")
	assert.Equal(t, expectedCPURequest.Value(), res.(*v1.Pod).Spec.Containers[0].Resources.Requests.Cpu().Value())
	expectedMemRequest := resource.MustParse("1024Mi")
	assert.Equal(t, expectedMemRequest.Value(), res.(*v1.Pod).Spec.Containers[0].Resources.Requests.Memory().Value())
	expectedCPULimit := resource.MustParse("2048m")
	assert.Equal(t, expectedCPULimit.Value(), res.(*v1.Pod).Spec.Containers[0].Resources.Limits.Cpu().Value())
	expectedMemLimit := resource.MustParse("1024Mi")
	assert.Equal(t, expectedMemLimit.Value(), res.(*v1.Pod).Spec.Containers[0].Resources.Limits.Memory().Value())
	expectedEphemeralStorageLimit := resource.MustParse("100M")
	assert.Equal(t, expectedEphemeralStorageLimit.Value(), res.(*v1.Pod).Spec.Containers[0].Resources.Limits.StorageEphemeral().Value())
}

func TestBuildSidecarReosurceMissingAnnotationsAndLabels(t *testing.T) {
	sideCarJob := sidecarJob{
		PrimaryContainerName: "PrimaryContainer",
		PodSpec: &v1.PodSpec{
			Containers: []v1.Container{
				{
					Name: "PrimaryContainer",
				},
			},
		},
	}

	task := getSidecarTaskTemplateForTest(sideCarJob)

	handler := &sidecarResourceHandler{}
	taskCtx := getDummySidecarTaskContext(task, resourceRequirements)
	resp, err := handler.BuildResource(context.TODO(), taskCtx)
	assert.NoError(t, err)
	assert.EqualValues(t, map[string]string{}, resp.GetLabels())
	assert.EqualValues(t, map[string]string{"primary_container_name": "PrimaryContainer"}, resp.GetAnnotations())
}

func TestBuildSidecarResourceMissingPrimary(t *testing.T) {
	sideCarJob := sidecarJob{
		PrimaryContainerName: "PrimaryContainer",
		PodSpec: &v1.PodSpec{
			Containers: []v1.Container{
				{
					Name: "SecondaryContainer",
				},
			},
		},
	}

	task := getSidecarTaskTemplateForTest(sideCarJob)

	handler := &sidecarResourceHandler{}
	taskCtx := getDummySidecarTaskContext(task, resourceRequirements)
	_, err := handler.BuildResource(context.TODO(), taskCtx)
	assert.True(t, errors.Is(err, errors2.Errorf("BadTaskSpecification", "")))
}

func TestGetTaskSidecarStatus(t *testing.T) {
	sideCarJob := sidecarJob{
		PrimaryContainerName: "PrimaryContainer",
		PodSpec: &v1.PodSpec{
			Containers: []v1.Container{
				{
					Name: "PrimaryContainer",
				},
			},
		},
	}

	task := getSidecarTaskTemplateForTest(sideCarJob)

	var testCases = map[v1.PodPhase]pluginsCore.Phase{
		v1.PodSucceeded:           pluginsCore.PhaseSuccess,
		v1.PodFailed:              pluginsCore.PhaseRetryableFailure,
		v1.PodReasonUnschedulable: pluginsCore.PhaseQueued,
		v1.PodUnknown:             pluginsCore.PhaseUndefined,
	}

	for podPhase, expectedTaskPhase := range testCases {
		res := &v1.Pod{
			Status: v1.PodStatus{
				Phase: podPhase,
			},
		}
		res.SetAnnotations(map[string]string{
			primaryContainerKey: "PrimaryContainer",
		})
		handler := &sidecarResourceHandler{}
		taskCtx := getDummySidecarTaskContext(task, resourceRequirements)
		phaseInfo, err := handler.GetTaskPhase(context.TODO(), taskCtx, res)
		assert.Nil(t, err)
		assert.Equal(t, expectedTaskPhase, phaseInfo.Phase(),
			"Expected [%v] got [%v] instead, for podPhase [%v]", expectedTaskPhase, phaseInfo.Phase(), podPhase)
	}
}

func TestDemystifiedSidecarStatus_PrimaryFailed(t *testing.T) {
	res := &v1.Pod{
		Status: v1.PodStatus{
			Phase: v1.PodRunning,
			ContainerStatuses: []v1.ContainerStatus{
				{
					Name: "Primary",
					State: v1.ContainerState{
						Terminated: &v1.ContainerStateTerminated{
							ExitCode: 1,
						},
					},
				},
			},
		},
	}
	res.SetAnnotations(map[string]string{
		primaryContainerKey: "Primary",
	})
	handler := &sidecarResourceHandler{}
	taskCtx := getDummySidecarTaskContext(&core.TaskTemplate{}, resourceRequirements)
	phaseInfo, err := handler.GetTaskPhase(context.TODO(), taskCtx, res)
	assert.Nil(t, err)
	assert.Equal(t, pluginsCore.PhaseRetryableFailure, phaseInfo.Phase())
}

func TestDemystifiedSidecarStatus_PrimarySucceeded(t *testing.T) {
	res := &v1.Pod{
		Status: v1.PodStatus{
			Phase: v1.PodRunning,
			ContainerStatuses: []v1.ContainerStatus{
				{
					Name: "Primary",
					State: v1.ContainerState{
						Terminated: &v1.ContainerStateTerminated{
							ExitCode: 0,
						},
					},
				},
			},
		},
	}
	res.SetAnnotations(map[string]string{
		primaryContainerKey: "Primary",
	})
	handler := &sidecarResourceHandler{}
	taskCtx := getDummySidecarTaskContext(&core.TaskTemplate{}, resourceRequirements)
	phaseInfo, err := handler.GetTaskPhase(context.TODO(), taskCtx, res)
	assert.Nil(t, err)
	assert.Equal(t, pluginsCore.PhaseSuccess, phaseInfo.Phase())
}

func TestDemystifiedSidecarStatus_PrimaryRunning(t *testing.T) {
	res := &v1.Pod{
		Status: v1.PodStatus{
			Phase: v1.PodRunning,
			ContainerStatuses: []v1.ContainerStatus{
				{
					Name: "Primary",
					State: v1.ContainerState{
						Waiting: &v1.ContainerStateWaiting{
							Reason: "stay patient",
						},
					},
				},
			},
		},
	}
	res.SetAnnotations(map[string]string{
		primaryContainerKey: "Primary",
	})
	handler := &sidecarResourceHandler{}
	taskCtx := getDummySidecarTaskContext(&core.TaskTemplate{}, resourceRequirements)
	phaseInfo, err := handler.GetTaskPhase(context.TODO(), taskCtx, res)
	assert.Nil(t, err)
	assert.Equal(t, pluginsCore.PhaseRunning, phaseInfo.Phase())
}

func TestDemystifiedSidecarStatus_PrimaryMissing(t *testing.T) {
	res := &v1.Pod{
		Status: v1.PodStatus{
			Phase: v1.PodRunning,
			ContainerStatuses: []v1.ContainerStatus{
				{
					Name: "Secondary",
				},
			},
		},
	}
	res.SetAnnotations(map[string]string{
		primaryContainerKey: "Primary",
	})
	handler := &sidecarResourceHandler{}
	taskCtx := getDummySidecarTaskContext(&core.TaskTemplate{}, resourceRequirements)
	phaseInfo, err := handler.GetTaskPhase(context.TODO(), taskCtx, res)
	assert.Nil(t, err)
	assert.Equal(t, pluginsCore.PhasePermanentFailure, phaseInfo.Phase())
}

func TestGetProperties(t *testing.T) {
	handler := &sidecarResourceHandler{}
	expected := k8s.PluginProperties{}
	assert.Equal(t, expected, handler.GetProperties())
}
