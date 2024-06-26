package pod

import (
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"path"
	"testing"

	structpb "github.com/golang/protobuf/ptypes/struct"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	errors2 "github.com/flyteorg/flyte/flyteplugins/go/tasks/errors"
	pluginsCore "github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/core"
	pluginsCoreMock "github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/core/mocks"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/flytek8s"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/flytek8s/config"
	pluginsIOMock "github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/io/mocks"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/k8s"
)

const ResourceNvidiaGPU = "nvidia.com/gpu"

var sidecarResourceRequirements = &v1.ResourceRequirements{
	Limits: v1.ResourceList{
		v1.ResourceCPU:              resource.MustParse("2048m"),
		v1.ResourceEphemeralStorage: resource.MustParse("100M"),
		ResourceNvidiaGPU:           resource.MustParse("1"),
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
		Type:   SidecarTaskType,
		Custom: &structObj,
	}
}

func dummySidecarTaskMetadata(resources *v1.ResourceRequirements, extendedResources *core.ExtendedResources) pluginsCore.TaskExecutionMetadata {
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
	taskMetadata.OnGetPlatformResources().Return(&v1.ResourceRequirements{})

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
	tID.On("GetUniqueNodeID").Return("an-unique-id")
	taskMetadata.On("GetTaskExecutionID").Return(tID)

	to := &pluginsCoreMock.TaskOverrides{}
	to.On("GetResources").Return(resources)
	to.On("GetExtendedResources").Return(extendedResources)
	to.On("GetContainerImage").Return("")
	taskMetadata.On("GetOverrides").Return(to)
	taskMetadata.On("GetEnvironmentVariables").Return(nil)
	taskMetadata.On("GetConsoleURL").Return("")

	return taskMetadata
}

func getDummySidecarTaskContext(taskTemplate *core.TaskTemplate, resources *v1.ResourceRequirements, extendedResources *core.ExtendedResources) pluginsCore.TaskExecutionContext {
	taskCtx := &pluginsCoreMock.TaskExecutionContext{}
	dummyTaskMetadata := dummySidecarTaskMetadata(resources, extendedResources)
	inputReader := &pluginsIOMock.InputReader{}
	inputReader.OnGetInputPrefixPath().Return("test-data-prefix")
	inputReader.OnGetInputPath().Return("test-data-reference")
	inputReader.OnGetMatch(mock.Anything).Return(&core.LiteralMap{}, nil)
	taskCtx.OnInputReader().Return(inputReader)

	outputReader := &pluginsIOMock.OutputWriter{}
	outputReader.OnGetOutputPath().Return("/data/outputs.pb")
	outputReader.OnGetOutputPrefixPath().Return("/data/")
	outputReader.OnGetRawOutputPrefix().Return("")
	outputReader.OnGetCheckpointPrefix().Return("/checkpoint")
	outputReader.OnGetPreviousCheckpointsPrefix().Return("/prev")
	taskCtx.OnOutputWriter().Return(outputReader)

	taskReader := &pluginsCoreMock.TaskReader{}
	taskReader.OnReadMatch(mock.Anything).Return(taskTemplate, nil)
	taskCtx.OnTaskReader().Return(taskReader)

	taskCtx.OnTaskExecutionMetadata().Return(dummyTaskMetadata)

	pluginStateReader := &pluginsCoreMock.PluginStateReader{}
	pluginStateReader.OnGetMatch(mock.Anything).Return(0, nil)
	taskCtx.OnPluginStateReader().Return(pluginStateReader)

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

func checkTolerations(t *testing.T, res client.Object, gpuTol v1.Toleration) {
	// Assert user-specified tolerations don't get overridden
	assert.Len(t, res.(*v1.Pod).Spec.Tolerations, 2)
	for _, tol := range res.(*v1.Pod).Spec.Tolerations {
		if tol.Key == "my toleration key" {
			assert.Equal(t, tol.Value, "my toleration value")
		} else if tol.Key == gpuTol.Key {
			assert.Equal(t, tol, gpuTol)
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
		Type:            SidecarTaskType,
		TaskTypeVersion: 2,
		Config: map[string]string{
			flytek8s.PrimaryContainerKey: "primary container",
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

	tolEphemeralStorage := v1.Toleration{
		Key:      "ephemeral-storage",
		Value:    "dedicated",
		Operator: v1.TolerationOpExists,
		Effect:   v1.TaintEffectNoSchedule,
	}
	assert.NoError(t, config.SetK8sPluginConfig(&config.K8sPluginConfig{
		ResourceTolerations: map[v1.ResourceName][]v1.Toleration{
			v1.ResourceStorage: {tolEphemeralStorage},
			ResourceNvidiaGPU:  {tolGPU},
		},
		DefaultCPURequest:    resource.MustParse("1024m"),
		DefaultMemoryRequest: resource.MustParse("1024Mi"),
		GpuResourceName:      ResourceNvidiaGPU,
	}))
	taskCtx := getDummySidecarTaskContext(&task, sidecarResourceRequirements, nil)
	res, err := DefaultPodPlugin.BuildResource(context.TODO(), taskCtx)
	assert.Nil(t, err)
	assert.EqualValues(t, map[string]string{
		flytek8s.PrimaryContainerKey: "primary container",
		"anno":                       "bar",
	}, res.GetAnnotations())
	assert.EqualValues(t, map[string]string{
		"label": "foo",
	}, res.GetLabels())

	// Assert volumes & volume mounts are preserved
	assert.Len(t, res.(*v1.Pod).Spec.Volumes, 1)
	assert.Equal(t, "dshm", res.(*v1.Pod).Spec.Volumes[0].Name)

	assert.Len(t, res.(*v1.Pod).Spec.Containers[0].VolumeMounts, 1)
	assert.Equal(t, "volume mount", res.(*v1.Pod).Spec.Containers[0].VolumeMounts[0].Name)

	checkTolerations(t, res, tolGPU)

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

	expectedGPURes := resource.MustParse("1")
	assert.Equal(t, expectedGPURes, res.(*v1.Pod).Spec.Containers[0].Resources.Requests[ResourceNvidiaGPU])
	assert.Equal(t, expectedGPURes, res.(*v1.Pod).Spec.Containers[0].Resources.Limits[ResourceNvidiaGPU])
	expectedGPURes = resource.MustParse("2")
	assert.Equal(t, expectedGPURes, res.(*v1.Pod).Spec.Containers[1].Resources.Requests[ResourceNvidiaGPU])
	assert.Equal(t, expectedGPURes, res.(*v1.Pod).Spec.Containers[1].Resources.Limits[ResourceNvidiaGPU])
}

func TestBuildSidecarResource_TaskType2_Invalid_Spec(t *testing.T) {
	task := core.TaskTemplate{
		Type:            SidecarTaskType,
		TaskTypeVersion: 2,
		Config: map[string]string{
			flytek8s.PrimaryContainerKey: "primary container",
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

	taskCtx := getDummySidecarTaskContext(&task, sidecarResourceRequirements, nil)
	_, err := DefaultPodPlugin.BuildResource(context.TODO(), taskCtx)
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
		Type:            SidecarTaskType,
		Custom:          structObj,
		TaskTypeVersion: 1,
		Config: map[string]string{
			flytek8s.PrimaryContainerKey: "primary container",
		},
	}

	tolGPU := v1.Toleration{
		Key:      "flyte/gpu",
		Value:    "dedicated",
		Operator: v1.TolerationOpEqual,
		Effect:   v1.TaintEffectNoSchedule,
	}

	tolEphemeralStorage := v1.Toleration{
		Key:      "ephemeral-storage",
		Value:    "dedicated",
		Operator: v1.TolerationOpExists,
		Effect:   v1.TaintEffectNoSchedule,
	}
	assert.NoError(t, config.SetK8sPluginConfig(&config.K8sPluginConfig{
		ResourceTolerations: map[v1.ResourceName][]v1.Toleration{
			v1.ResourceStorage: {tolEphemeralStorage},
			ResourceNvidiaGPU:  {tolGPU},
		},
		DefaultCPURequest:    resource.MustParse("1024m"),
		DefaultMemoryRequest: resource.MustParse("1024Mi"),
	}))
	taskCtx := getDummySidecarTaskContext(&task, sidecarResourceRequirements, nil)
	res, err := DefaultPodPlugin.BuildResource(context.TODO(), taskCtx)
	assert.Nil(t, err)
	assert.EqualValues(t, map[string]string{
		flytek8s.PrimaryContainerKey: "primary container",
	}, res.GetAnnotations())
	assert.EqualValues(t, map[string]string{}, res.GetLabels())

	// Assert volumes & volume mounts are preserved
	assert.Len(t, res.(*v1.Pod).Spec.Volumes, 1)
	assert.Equal(t, "dshm", res.(*v1.Pod).Spec.Volumes[0].Name)

	assert.Len(t, res.(*v1.Pod).Spec.Containers[0].VolumeMounts, 1)
	assert.Equal(t, "volume mount", res.(*v1.Pod).Spec.Containers[0].VolumeMounts[0].Name)

	checkTolerations(t, res, tolGPU)
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
		Type:            SidecarTaskType,
		Custom:          structObj,
		TaskTypeVersion: 1,
	}

	assert.NoError(t, config.SetK8sPluginConfig(&config.K8sPluginConfig{
		ResourceTolerations: map[v1.ResourceName][]v1.Toleration{
			v1.ResourceStorage: {},
			ResourceNvidiaGPU:  {},
		},
		DefaultCPURequest:    resource.MustParse("1024m"),
		DefaultMemoryRequest: resource.MustParse("1024Mi"),
	}))
	taskCtx := getDummySidecarTaskContext(&task, sidecarResourceRequirements, nil)
	_, err = DefaultPodPlugin.BuildResource(context.TODO(), taskCtx)
	assert.EqualError(t, err, "[BadTaskSpecification] invalid TaskSpecification, config needs to be non-empty and include missing [primary_container_name] key")

	task.Config = map[string]string{
		"foo": "bar",
	}
	taskCtx = getDummySidecarTaskContext(&task, sidecarResourceRequirements, nil)
	_, err = DefaultPodPlugin.BuildResource(context.TODO(), taskCtx)
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
		Type:   SidecarTaskType,
		Custom: &sidecarCustom,
	}

	tolGPU := v1.Toleration{
		Key:      "flyte/gpu",
		Value:    "dedicated",
		Operator: v1.TolerationOpEqual,
		Effect:   v1.TaintEffectNoSchedule,
	}

	tolEphemeralStorage := v1.Toleration{
		Key:      "ephemeral-storage",
		Value:    "dedicated",
		Operator: v1.TolerationOpExists,
		Effect:   v1.TaintEffectNoSchedule,
	}
	assert.NoError(t, config.SetK8sPluginConfig(&config.K8sPluginConfig{
		ResourceTolerations: map[v1.ResourceName][]v1.Toleration{
			v1.ResourceStorage: {tolEphemeralStorage},
			ResourceNvidiaGPU:  {tolGPU},
		},
		DefaultCPURequest:    resource.MustParse("1024m"),
		DefaultMemoryRequest: resource.MustParse("1024Mi"),
	}))
	taskCtx := getDummySidecarTaskContext(&task, sidecarResourceRequirements, nil)
	res, err := DefaultPodPlugin.BuildResource(context.TODO(), taskCtx)
	assert.Nil(t, err)
	assert.EqualValues(t, map[string]string{
		flytek8s.PrimaryContainerKey: "a container",
		"a1":                         "a1",
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

	checkTolerations(t, res, tolGPU)

	// Assert resource requirements are correctly set
	expectedCPURequest := resource.MustParse("2048m")
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

	taskCtx := getDummySidecarTaskContext(task, sidecarResourceRequirements, nil)
	resp, err := DefaultPodPlugin.BuildResource(context.TODO(), taskCtx)
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

	taskCtx := getDummySidecarTaskContext(task, sidecarResourceRequirements, nil)
	_, err := DefaultPodPlugin.BuildResource(context.TODO(), taskCtx)
	assert.True(t, errors.Is(err, errors2.Errorf("BadTaskSpecification", "")))
}

func TestBuildSidecarResource_ExtendedResources(t *testing.T) {
	assert.NoError(t, config.SetK8sPluginConfig(&config.K8sPluginConfig{
		GpuDeviceNodeLabel:        "gpu-node-label",
		GpuPartitionSizeNodeLabel: "gpu-partition-size",
		GpuResourceName:           flytek8s.ResourceNvidiaGPU,
	}))

	fixtures := []struct {
		name                      string
		resources                 *v1.ResourceRequirements
		extendedResourcesBase     *core.ExtendedResources
		extendedResourcesOverride *core.ExtendedResources
		expectedNsr               []v1.NodeSelectorTerm
		expectedTol               []v1.Toleration
	}{
		{
			"without overrides",
			&v1.ResourceRequirements{
				Limits: v1.ResourceList{
					flytek8s.ResourceNvidiaGPU: resource.MustParse("1"),
				},
			},
			&core.ExtendedResources{
				GpuAccelerator: &core.GPUAccelerator{
					Device: "nvidia-tesla-t4",
				},
			},
			nil,
			[]v1.NodeSelectorTerm{
				{
					MatchExpressions: []v1.NodeSelectorRequirement{
						v1.NodeSelectorRequirement{
							Key:      "gpu-node-label",
							Operator: v1.NodeSelectorOpIn,
							Values:   []string{"nvidia-tesla-t4"},
						},
					},
				},
			},
			[]v1.Toleration{
				{
					Key:      "gpu-node-label",
					Value:    "nvidia-tesla-t4",
					Operator: v1.TolerationOpEqual,
					Effect:   v1.TaintEffectNoSchedule,
				},
			},
		},
		{
			"with overrides",
			&v1.ResourceRequirements{
				Limits: v1.ResourceList{
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
			[]v1.NodeSelectorTerm{
				{
					MatchExpressions: []v1.NodeSelectorRequirement{
						v1.NodeSelectorRequirement{
							Key:      "gpu-node-label",
							Operator: v1.NodeSelectorOpIn,
							Values:   []string{"nvidia-tesla-a100"},
						},
						v1.NodeSelectorRequirement{
							Key:      "gpu-partition-size",
							Operator: v1.NodeSelectorOpIn,
							Values:   []string{"1g.5gb"},
						},
					},
				},
			},
			[]v1.Toleration{
				{
					Key:      "gpu-node-label",
					Value:    "nvidia-tesla-a100",
					Operator: v1.TolerationOpEqual,
					Effect:   v1.TaintEffectNoSchedule,
				},
				{
					Key:      "gpu-partition-size",
					Value:    "1g.5gb",
					Operator: v1.TolerationOpEqual,
					Effect:   v1.TaintEffectNoSchedule,
				},
			},
		},
	}

	podSpec := v1.PodSpec{
		Containers: []v1.Container{
			{
				Name: "primary container",
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
	testConfigs := []struct {
		name         string
		taskTemplate core.TaskTemplate
	}{
		{
			"v0",
			*getSidecarTaskTemplateForTest(sidecarJob{
				PrimaryContainerName: podSpec.Containers[0].Name,
				PodSpec:              &podSpec,
			}),
		},
		{
			"v1",
			core.TaskTemplate{
				Type:            SidecarTaskType,
				Custom:          structObj,
				TaskTypeVersion: 1,
				Config: map[string]string{
					flytek8s.PrimaryContainerKey: podSpec.Containers[0].Name,
				},
			},
		},
		{
			"v2",
			core.TaskTemplate{
				Type:            SidecarTaskType,
				TaskTypeVersion: 2,
				Config: map[string]string{
					flytek8s.PrimaryContainerKey: podSpec.Containers[0].Name,
				},
				Target: &core.TaskTemplate_K8SPod{
					K8SPod: &core.K8SPod{
						PodSpec: structObj,
					},
				},
			},
		},
	}

	for _, tCfg := range testConfigs {
		for _, f := range fixtures {
			t.Run(tCfg.name+" "+f.name, func(t *testing.T) {
				taskTemplate := tCfg.taskTemplate
				taskTemplate.ExtendedResources = f.extendedResourcesBase
				taskContext := getDummySidecarTaskContext(&taskTemplate, f.resources, f.extendedResourcesOverride)
				r, err := DefaultPodPlugin.BuildResource(context.TODO(), taskContext)
				assert.Nil(t, err)
				assert.NotNil(t, r)
				pod, ok := r.(*v1.Pod)
				assert.True(t, ok)

				assert.EqualValues(
					t,
					f.expectedNsr,
					pod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms,
				)
				assert.EqualValues(
					t,
					f.expectedTol,
					pod.Spec.Tolerations,
				)
			})
		}
	}
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
			flytek8s.PrimaryContainerKey: "PrimaryContainer",
		})
		taskCtx := getDummySidecarTaskContext(task, sidecarResourceRequirements, nil)
		phaseInfo, err := DefaultPodPlugin.GetTaskPhase(context.TODO(), taskCtx, res)
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
		flytek8s.PrimaryContainerKey: "Primary",
	})
	taskCtx := getDummySidecarTaskContext(&core.TaskTemplate{}, sidecarResourceRequirements, nil)
	phaseInfo, err := DefaultPodPlugin.GetTaskPhase(context.TODO(), taskCtx, res)
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
		flytek8s.PrimaryContainerKey: "Primary",
	})
	taskCtx := getDummySidecarTaskContext(&core.TaskTemplate{}, sidecarResourceRequirements, nil)
	phaseInfo, err := DefaultPodPlugin.GetTaskPhase(context.TODO(), taskCtx, res)
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
		flytek8s.PrimaryContainerKey: "Primary",
	})
	taskCtx := getDummySidecarTaskContext(&core.TaskTemplate{}, sidecarResourceRequirements, nil)
	phaseInfo, err := DefaultPodPlugin.GetTaskPhase(context.TODO(), taskCtx, res)
	assert.Nil(t, err)
	assert.Equal(t, pluginsCore.PhaseRunning, phaseInfo.Phase())
}

func TestDemystifiedSidecarStatus_PrimaryMissing(t *testing.T) {
	res := &v1.Pod{
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Name: "Secondary",
				},
			},
		},
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
		flytek8s.PrimaryContainerKey: "Primary",
	})
	taskCtx := getDummySidecarTaskContext(&core.TaskTemplate{}, sidecarResourceRequirements, nil)
	phaseInfo, err := DefaultPodPlugin.GetTaskPhase(context.TODO(), taskCtx, res)
	assert.Nil(t, err)
	assert.Equal(t, pluginsCore.PhasePermanentFailure, phaseInfo.Phase())
}

func TestDemystifiedSidecarStatus_PrimaryNotExistsYet(t *testing.T) {
	res := &v1.Pod{
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Name: "Primary",
				},
			},
		},
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
		flytek8s.PrimaryContainerKey: "Primary",
	})
	taskCtx := getDummySidecarTaskContext(&core.TaskTemplate{}, sidecarResourceRequirements, nil)
	phaseInfo, err := DefaultPodPlugin.GetTaskPhase(context.TODO(), taskCtx, res)
	assert.Nil(t, err)
	assert.Equal(t, pluginsCore.PhaseRunning, phaseInfo.Phase())
}

func TestGetProperties(t *testing.T) {
	expected := k8s.PluginProperties{}
	assert.Equal(t, expected, DefaultPodPlugin.GetProperties())
}
