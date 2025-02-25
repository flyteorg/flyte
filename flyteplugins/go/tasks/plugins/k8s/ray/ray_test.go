package ray

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"
	"time"

	rayv1 "github.com/ray-project/kuberay/ray-operator/apis/ray/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/structpb"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/plugins"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/logs"
	pluginsCore "github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/core/mocks"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/flytek8s"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/flytek8s/config"
	pluginIOMocks "github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/io/mocks"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/k8s"
	mocks2 "github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/k8s/mocks"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/tasklog"
	"github.com/flyteorg/flyte/flytestdlib/utils"
)

const (
	testImage      = "image://"
	serviceAccount = "ray_sa"
)

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

	workerGroupName = "worker-group"
)

func transformRayJobToCustomObj(rayJob *plugins.RayJob) *structpb.Struct {
	structObj, err := utils.MarshalObjToStruct(rayJob)
	if err != nil {
		panic(err)
	}
	return structObj
}

func transformPodSpecToTaskTemplateTarget(podSpec *corev1.PodSpec) *core.TaskTemplate_K8SPod {
	structObj, err := utils.MarshalObjToStruct(&podSpec)
	if err != nil {
		panic(err)
	}
	return &core.TaskTemplate_K8SPod{
		K8SPod: &core.K8SPod{
			PodSpec: structObj,
		},
	}
}

func dummyRayCustomObj() *plugins.RayJob {
	return &plugins.RayJob{
		RayCluster: &plugins.RayCluster{
			HeadGroupSpec:     &plugins.HeadGroupSpec{RayStartParams: map[string]string{"num-cpus": "1"}},
			WorkerGroupSpec:   []*plugins.WorkerGroupSpec{{GroupName: workerGroupName, Replicas: 3, MinReplicas: 3, MaxReplicas: 3}},
			EnableAutoscaling: true,
		},
		ShutdownAfterJobFinishes: true,
		TtlSecondsAfterFinished:  120,
	}
}

func dummyRayTaskTemplate(id string, rayJob *plugins.RayJob) *core.TaskTemplate {
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
		Custom: transformRayJobToCustomObj(rayJob),
	}
}

func dummyRayTaskContext(taskTemplate *core.TaskTemplate, resources *corev1.ResourceRequirements, extendedResources *core.ExtendedResources, containerImage, serviceAccount string) pluginsCore.TaskExecutionContext {
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

	overrides := &mocks.TaskOverrides{}
	overrides.OnGetResources().Return(resources)
	overrides.OnGetExtendedResources().Return(extendedResources)
	overrides.OnGetContainerImage().Return(containerImage)
	overrides.OnGetPodTemplate().Return(nil)

	taskExecutionMetadata := &mocks.TaskExecutionMetadata{}
	taskExecutionMetadata.OnGetTaskExecutionID().Return(tID)
	taskExecutionMetadata.OnGetNamespace().Return("test-namespace")
	taskExecutionMetadata.OnGetAnnotations().Return(map[string]string{"annotation-1": "val1"})
	taskExecutionMetadata.OnGetLabels().Return(map[string]string{"label-1": "val1"})
	taskExecutionMetadata.OnGetOwnerReference().Return(metav1.OwnerReference{
		Kind: "node",
		Name: "blah",
	})
	taskExecutionMetadata.OnIsInterruptible().Return(true)
	taskExecutionMetadata.OnGetOverrides().Return(overrides)
	taskExecutionMetadata.OnGetK8sServiceAccount().Return(serviceAccount)
	taskExecutionMetadata.OnGetPlatformResources().Return(&corev1.ResourceRequirements{})
	taskExecutionMetadata.OnGetSecurityContext().Return(core.SecurityContext{
		RunAs: &core.Identity{K8SServiceAccount: serviceAccount},
	})
	taskExecutionMetadata.OnGetEnvironmentVariables().Return(nil)
	taskExecutionMetadata.OnGetConsoleURL().Return("")
	taskExecutionMetadata.OnGetOOMFailures().Return(0)
	taskCtx.OnTaskExecutionMetadata().Return(taskExecutionMetadata)
	return taskCtx
}

func TestBuildResourceRay(t *testing.T) {
	rayJobResourceHandler := rayJobResourceHandler{}
	taskTemplate := dummyRayTaskTemplate("ray-id", dummyRayCustomObj())
	toleration := []corev1.Toleration{{
		Key:      "storage",
		Value:    "dedicated",
		Operator: corev1.TolerationOpExists,
		Effect:   corev1.TaintEffectNoSchedule,
	}}
	err := config.SetK8sPluginConfig(&config.K8sPluginConfig{DefaultTolerations: toleration})
	assert.Nil(t, err)

	rayCtx := dummyRayTaskContext(taskTemplate, resourceRequirements, nil, "", serviceAccount)
	RayResource, err := rayJobResourceHandler.BuildResource(context.TODO(), rayCtx)
	assert.Nil(t, err)

	assert.NotNil(t, RayResource)
	ray, ok := RayResource.(*rayv1.RayJob)
	assert.True(t, ok)

	assert.Equal(t, *ray.Spec.RayClusterSpec.EnableInTreeAutoscaling, true)
	assert.Equal(t, ray.Spec.ShutdownAfterJobFinishes, true)
	assert.Equal(t, ray.Spec.TTLSecondsAfterFinished, int32(120))

	assert.Equal(t, ray.Spec.RayClusterSpec.HeadGroupSpec.Template.Spec.ServiceAccountName, serviceAccount)
	assert.Equal(t, ray.Spec.RayClusterSpec.HeadGroupSpec.RayStartParams,
		map[string]string{
			"dashboard-host": "0.0.0.0", "disable-usage-stats": "true", "include-dashboard": "true",
			"node-ip-address": "$MY_POD_IP", "num-cpus": "1",
		})
	assert.Equal(t, ray.Spec.RayClusterSpec.HeadGroupSpec.Template.Annotations, map[string]string{"annotation-1": "val1"})
	assert.Equal(t, ray.Spec.RayClusterSpec.HeadGroupSpec.Template.Labels, map[string]string{"label-1": "val1"})
	assert.Equal(t, ray.Spec.RayClusterSpec.HeadGroupSpec.Template.Spec.Tolerations, toleration)

	workerReplica := int32(3)
	assert.Equal(t, *ray.Spec.RayClusterSpec.WorkerGroupSpecs[0].Replicas, workerReplica)
	assert.Equal(t, *ray.Spec.RayClusterSpec.WorkerGroupSpecs[0].MinReplicas, workerReplica)
	assert.Equal(t, *ray.Spec.RayClusterSpec.WorkerGroupSpecs[0].MaxReplicas, workerReplica)
	assert.Equal(t, ray.Spec.RayClusterSpec.WorkerGroupSpecs[0].GroupName, workerGroupName)
	assert.Equal(t, ray.Spec.RayClusterSpec.WorkerGroupSpecs[0].Template.Spec.ServiceAccountName, serviceAccount)
	assert.Equal(t, ray.Spec.RayClusterSpec.WorkerGroupSpecs[0].RayStartParams, map[string]string{"disable-usage-stats": "true", "node-ip-address": "$MY_POD_IP"})
	assert.Equal(t, ray.Spec.RayClusterSpec.WorkerGroupSpecs[0].Template.Annotations, map[string]string{"annotation-1": "val1"})
	assert.Equal(t, ray.Spec.RayClusterSpec.WorkerGroupSpecs[0].Template.Labels, map[string]string{"label-1": "val1"})
	assert.Equal(t, ray.Spec.RayClusterSpec.WorkerGroupSpecs[0].Template.Spec.Tolerations, toleration)

	// Make sure the default service account is being used if SA is not provided in the task context
	rayCtx = dummyRayTaskContext(taskTemplate, resourceRequirements, nil, "", "")
	RayResource, err = rayJobResourceHandler.BuildResource(context.TODO(), rayCtx)
	assert.Nil(t, err)
	assert.NotNil(t, RayResource)
	ray, ok = RayResource.(*rayv1.RayJob)
	assert.True(t, ok)
	assert.Equal(t, ray.Spec.RayClusterSpec.HeadGroupSpec.Template.Spec.ServiceAccountName, GetConfig().ServiceAccount)
}

func TestBuildResourceRayContainerImage(t *testing.T) {
	assert.NoError(t, config.SetK8sPluginConfig(&config.K8sPluginConfig{}))

	fixtures := []struct {
		name                   string
		resources              *corev1.ResourceRequirements
		containerImageOverride string
	}{
		{
			"without overrides",
			&corev1.ResourceRequirements{
				Limits: corev1.ResourceList{
					flytek8s.ResourceNvidiaGPU: resource.MustParse("1"),
				},
			},
			"",
		},
		{
			"with overrides",
			&corev1.ResourceRequirements{
				Limits: corev1.ResourceList{
					flytek8s.ResourceNvidiaGPU: resource.MustParse("1"),
				},
			},
			"container-image-override",
		},
	}

	for _, f := range fixtures {
		t.Run(f.name, func(t *testing.T) {
			taskTemplate := dummyRayTaskTemplate("id", dummyRayCustomObj())
			taskContext := dummyRayTaskContext(taskTemplate, f.resources, nil, f.containerImageOverride, serviceAccount)
			rayJobResourceHandler := rayJobResourceHandler{}
			r, err := rayJobResourceHandler.BuildResource(context.TODO(), taskContext)
			assert.Nil(t, err)
			assert.NotNil(t, r)
			rayJob, ok := r.(*rayv1.RayJob)
			assert.True(t, ok)

			var expectedContainerImage string
			if len(f.containerImageOverride) > 0 {
				expectedContainerImage = f.containerImageOverride
			} else {
				expectedContainerImage = testImage
			}

			// Head node
			headNodeSpec := rayJob.Spec.RayClusterSpec.HeadGroupSpec.Template.Spec
			assert.Equal(t, expectedContainerImage, headNodeSpec.Containers[0].Image)

			// Worker node
			workerNodeSpec := rayJob.Spec.RayClusterSpec.WorkerGroupSpecs[0].Template.Spec
			assert.Equal(t, expectedContainerImage, workerNodeSpec.Containers[0].Image)
		})
	}
}

func TestBuildPodTemplate(t *testing.T) {
	taskTemplate := dummyRayTaskTemplate("id", dummyRayCustomObj())
	resources := &corev1.ResourceRequirements{
		Limits: corev1.ResourceList{
			flytek8s.ResourceNvidiaGPU: resource.MustParse("1"),
		},
	}
	taskContext := dummyRayTaskContext(taskTemplate, resources, nil, "container-imag", serviceAccount)
	basePodSpec, objectMeta, _, err := flytek8s.ToK8sPodSpec(context.Background(), taskContext)
	assert.Nil(t, err)

	toleration := []corev1.Toleration{
		{
			Key:      "gpu-node-label",
			Value:    "nvidia-tesla-t4",
			Operator: corev1.TolerationOpEqual,
			Effect:   corev1.TaintEffectNoSchedule,
		},
	}

	customPodSpec :=
		&core.K8SPod{
			PodSpec: transformStructToStructPB(t, &corev1.PodSpec{
				Tolerations: toleration,
			}),
			Metadata: &core.K8SObjectMetadata{
				Labels:      map[string]string{"new-label-1": "val1"},
				Annotations: map[string]string{"new-annotation-1": "val1"},
			},
		}

	headGroupSpec := plugins.HeadGroupSpec{K8SPod: customPodSpec}
	podSpec, err := buildHeadPodTemplate(&basePodSpec.Containers[0], basePodSpec, objectMeta, taskContext, &headGroupSpec)
	assert.Nil(t, err)
	assert.Equal(t, podSpec.Spec.Tolerations, toleration)
	expectedLabels := map[string]string{"label-1": "val1", "new-label-1": "val1"}
	expectedAnnotations := map[string]string{"annotation-1": "val1", "new-annotation-1": "val1"}
	assert.Equal(t, expectedLabels, podSpec.Labels)
	assert.Equal(t, expectedAnnotations, podSpec.Annotations)

	workerGroupSpec := plugins.WorkerGroupSpec{K8SPod: customPodSpec}
	podSpec, err = buildWorkerPodTemplate(&basePodSpec.Containers[0], basePodSpec, objectMeta, taskContext, &workerGroupSpec)
	assert.Nil(t, err)
	assert.Equal(t, toleration, podSpec.Spec.Tolerations)
	assert.Equal(t, expectedLabels, podSpec.Labels)
	assert.Equal(t, expectedAnnotations, podSpec.Annotations)
}

func TestBuildResourceRayExtendedResources(t *testing.T) {
	assert.NoError(t, config.SetK8sPluginConfig(&config.K8sPluginConfig{
		GpuDeviceNodeLabel:                 "gpu-node-label",
		GpuPartitionSizeNodeLabel:          "gpu-partition-size",
		GpuResourceName:                    flytek8s.ResourceNvidiaGPU,
		AddTolerationsForExtendedResources: []string{"nvidia.com/gpu"},
	}))

	params := []struct {
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
						{
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
						{
							Key:      "gpu-node-label",
							Operator: corev1.NodeSelectorOpIn,
							Values:   []string{"nvidia-tesla-a100"},
						},
						{
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

	for _, p := range params {
		t.Run(p.name, func(t *testing.T) {
			taskTemplate := dummyRayTaskTemplate("ray-id", dummyRayCustomObj())
			taskTemplate.ExtendedResources = p.extendedResourcesBase
			taskContext := dummyRayTaskContext(taskTemplate, p.resources, p.extendedResourcesOverride, "", serviceAccount)
			rayJobResourceHandler := rayJobResourceHandler{}
			r, err := rayJobResourceHandler.BuildResource(context.TODO(), taskContext)
			assert.Nil(t, err)
			assert.NotNil(t, r)
			rayJob, ok := r.(*rayv1.RayJob)
			assert.True(t, ok)

			// Head node
			headNodeSpec := rayJob.Spec.RayClusterSpec.HeadGroupSpec.Template.Spec
			assert.EqualValues(
				t,
				p.expectedNsr,
				headNodeSpec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms,
			)
			assert.EqualValues(
				t,
				p.expectedTol,
				headNodeSpec.Tolerations,
			)

			// Worker node
			workerNodeSpec := rayJob.Spec.RayClusterSpec.WorkerGroupSpecs[0].Template.Spec
			assert.EqualValues(
				t,
				p.expectedNsr,
				workerNodeSpec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms,
			)
			assert.EqualValues(
				t,
				p.expectedTol,
				workerNodeSpec.Tolerations,
			)
		})
	}
}

type rayPodAssertions struct {
	resources        *corev1.ResourceRequirements
	runtimeClassName *string
	tolerations      []corev1.Toleration
	affinity         *corev1.Affinity
}

func TestBuildResourceRayCustomK8SPod(t *testing.T) {
	assert.NoError(t, config.SetK8sPluginConfig(&config.K8sPluginConfig{}))

	headResourceEntries := []*core.Resources_ResourceEntry{
		{Name: core.Resources_CPU, Value: "10"},
		{Name: core.Resources_MEMORY, Value: "10Gi"},
		{Name: core.Resources_GPU, Value: "10"},
	}
	headResources := &core.Resources{Requests: headResourceEntries, Limits: headResourceEntries}

	expectedHeadResources, err := flytek8s.ToK8sResourceRequirements(headResources, 0)
	require.NoError(t, err)

	workerResourceEntries := []*core.Resources_ResourceEntry{
		{Name: core.Resources_CPU, Value: "20"},
		{Name: core.Resources_MEMORY, Value: "20Gi"},
		{Name: core.Resources_GPU, Value: "20"},
	}
	workerResources := &core.Resources{Requests: workerResourceEntries, Limits: workerResourceEntries}

	expectedWorkerResources, err := flytek8s.ToK8sResourceRequirements(workerResources, 0)
	require.NoError(t, err)

	nvidiaRuntimeClassName := "nvidia-cdi"

	headTolerations := []corev1.Toleration{
		{
			Key:      "head",
			Operator: corev1.TolerationOpEqual,
			Value:    "true",
			Effect:   corev1.TaintEffectNoSchedule,
		},
	}

	workerTolerations := []corev1.Toleration{
		{
			Key:      "worker",
			Operator: corev1.TolerationOpEqual,
			Value:    "true",
			Effect:   corev1.TaintEffectNoSchedule,
		},
	}

	headPodSpecCustomTolerations := &corev1.PodSpec{
		Tolerations: headTolerations,
	}
	workerPodSpecCustomTolerations := &corev1.PodSpec{
		Tolerations: workerTolerations,
	}

	headAffinity := &corev1.Affinity{
		NodeAffinity: &corev1.NodeAffinity{
			RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
				NodeSelectorTerms: []corev1.NodeSelectorTerm{
					{
						MatchExpressions: []corev1.NodeSelectorRequirement{
							{
								Key:      "node-type",
								Operator: corev1.NodeSelectorOpIn,
								Values:   []string{"head-node"},
							},
						},
					},
				},
			},
		},
	}

	workerAffinity := &corev1.Affinity{
		NodeAffinity: &corev1.NodeAffinity{
			RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
				NodeSelectorTerms: []corev1.NodeSelectorTerm{
					{
						MatchExpressions: []corev1.NodeSelectorRequirement{
							{
								Key:      "node-type",
								Operator: corev1.NodeSelectorOpIn,
								Values:   []string{"worker-node"},
							},
						},
					},
				},
			},
		},
	}

	params := []struct {
		name                string
		headK8SPod          *core.K8SPod
		workerK8SPod        *core.K8SPod
		headPodAssertions   rayPodAssertions
		workerPodAssertions rayPodAssertions
	}{
		{
			name:         "no customizations",
			headK8SPod:   &core.K8SPod{},
			workerK8SPod: &core.K8SPod{},
			headPodAssertions: rayPodAssertions{
				affinity:  &corev1.Affinity{},
				resources: resourceRequirements,
			},
			workerPodAssertions: rayPodAssertions{
				affinity:  &corev1.Affinity{},
				resources: resourceRequirements,
			},
		},
		{
			name: "custom worker and head resources",
			headK8SPod: &core.K8SPod{
				PodSpec: transformStructToStructPB(t, &corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:      "ray-head",
							Resources: *expectedHeadResources,
						},
					},
				}),
			},
			workerK8SPod: &core.K8SPod{
				PodSpec: transformStructToStructPB(t, &corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:      "ray-worker",
							Resources: *expectedWorkerResources,
						},
					},
				}),
			},
			headPodAssertions: rayPodAssertions{
				affinity:  &corev1.Affinity{},
				resources: expectedHeadResources,
			},
			workerPodAssertions: rayPodAssertions{
				affinity:  &corev1.Affinity{},
				resources: expectedWorkerResources,
			},
		},
		{
			name: "custom runtime class name",
			headK8SPod: &core.K8SPod{
				PodSpec: transformStructToStructPB(t, &corev1.PodSpec{
					RuntimeClassName: &nvidiaRuntimeClassName,
				}),
			},
			workerK8SPod: &core.K8SPod{
				PodSpec: transformStructToStructPB(t, &corev1.PodSpec{
					RuntimeClassName: &nvidiaRuntimeClassName,
				}),
			},
			headPodAssertions: rayPodAssertions{
				affinity:         &corev1.Affinity{},
				resources:        resourceRequirements,
				runtimeClassName: &nvidiaRuntimeClassName,
			},
			workerPodAssertions: rayPodAssertions{
				affinity:         &corev1.Affinity{},
				resources:        resourceRequirements,
				runtimeClassName: &nvidiaRuntimeClassName,
			},
		},
		{
			name: "custom tolerations",
			headK8SPod: &core.K8SPod{
				PodSpec: transformStructToStructPB(t, headPodSpecCustomTolerations),
			},
			workerK8SPod: &core.K8SPod{
				PodSpec: transformStructToStructPB(t, workerPodSpecCustomTolerations),
			},
			headPodAssertions: rayPodAssertions{
				affinity:    &corev1.Affinity{},
				resources:   resourceRequirements,
				tolerations: headTolerations,
			},
			workerPodAssertions: rayPodAssertions{
				affinity:    &corev1.Affinity{},
				resources:   resourceRequirements,
				tolerations: workerTolerations,
			},
		},
		{
			name: "custom affinity",
			headK8SPod: &core.K8SPod{
				PodSpec: transformStructToStructPB(t, &corev1.PodSpec{
					Affinity: headAffinity,
				}),
			},
			workerK8SPod: &core.K8SPod{
				PodSpec: transformStructToStructPB(t, &corev1.PodSpec{
					Affinity: workerAffinity,
				}),
			},
			headPodAssertions: rayPodAssertions{
				affinity:  headAffinity,
				resources: resourceRequirements,
			},
			workerPodAssertions: rayPodAssertions{
				affinity:  workerAffinity,
				resources: resourceRequirements,
			},
		},
	}

	for _, p := range params {
		t.Run(p.name, func(t *testing.T) {
			rayJobInput := dummyRayCustomObj()

			if p.headK8SPod != nil {
				rayJobInput.RayCluster.HeadGroupSpec.K8SPod = p.headK8SPod
			}

			if p.workerK8SPod != nil {
				for _, spec := range rayJobInput.GetRayCluster().GetWorkerGroupSpec() {
					spec.K8SPod = p.workerK8SPod
				}
			}

			taskTemplate := dummyRayTaskTemplate("ray-id", rayJobInput)
			taskContext := dummyRayTaskContext(taskTemplate, resourceRequirements, nil, "", serviceAccount)
			rayJobResourceHandler := rayJobResourceHandler{}
			r, err := rayJobResourceHandler.BuildResource(context.TODO(), taskContext)
			assert.Nil(t, err)
			assert.NotNil(t, r)
			rayJob, ok := r.(*rayv1.RayJob)
			assert.True(t, ok)

			headPodSpec := rayJob.Spec.RayClusterSpec.HeadGroupSpec.Template.Spec
			headPodResources := headPodSpec.Containers[0].Resources
			if p.headPodAssertions.resources != nil {
				assert.EqualValues(t,
					*p.headPodAssertions.resources,
					headPodResources,
				)
			}

			assert.EqualValues(t, p.headPodAssertions.runtimeClassName, headPodSpec.RuntimeClassName)
			assert.EqualValues(t, p.headPodAssertions.tolerations, headPodSpec.Tolerations)
			assert.EqualValues(t, p.headPodAssertions.affinity, headPodSpec.Affinity)

			for _, workerGroupSpec := range rayJob.Spec.RayClusterSpec.WorkerGroupSpecs {
				workerPodSpec := workerGroupSpec.Template.Spec
				workerPodResources := workerPodSpec.Containers[0].Resources

				if p.workerPodAssertions.resources != nil {
					assert.EqualValues(t,
						*p.workerPodAssertions.resources,
						workerPodResources,
					)
				}

				assert.EqualValues(t, p.workerPodAssertions.runtimeClassName, workerPodSpec.RuntimeClassName)
				assert.EqualValues(t, p.workerPodAssertions.tolerations, workerPodSpec.Tolerations)
				assert.EqualValues(t, p.workerPodAssertions.affinity, workerPodSpec.Affinity)
			}
		})
	}
}

func TestDefaultStartParameters(t *testing.T) {
	rayJobResourceHandler := rayJobResourceHandler{}
	rayJob := &plugins.RayJob{
		RayCluster: &plugins.RayCluster{
			HeadGroupSpec:     &plugins.HeadGroupSpec{},
			WorkerGroupSpec:   []*plugins.WorkerGroupSpec{{GroupName: workerGroupName, Replicas: 3, MinReplicas: 3, MaxReplicas: 3}},
			EnableAutoscaling: true,
		},
		ShutdownAfterJobFinishes: true,
		TtlSecondsAfterFinished:  120,
	}

	taskTemplate := dummyRayTaskTemplate("ray-id", rayJob)
	toleration := []corev1.Toleration{{
		Key:      "storage",
		Value:    "dedicated",
		Operator: corev1.TolerationOpExists,
		Effect:   corev1.TaintEffectNoSchedule,
	}}
	err := config.SetK8sPluginConfig(&config.K8sPluginConfig{DefaultTolerations: toleration})
	assert.Nil(t, err)

	RayResource, err := rayJobResourceHandler.BuildResource(context.TODO(), dummyRayTaskContext(taskTemplate, resourceRequirements, nil, "", serviceAccount))
	assert.Nil(t, err)

	assert.NotNil(t, RayResource)
	ray, ok := RayResource.(*rayv1.RayJob)
	assert.True(t, ok)

	assert.Equal(t, *ray.Spec.RayClusterSpec.EnableInTreeAutoscaling, true)
	assert.Equal(t, ray.Spec.ShutdownAfterJobFinishes, true)
	assert.Equal(t, ray.Spec.TTLSecondsAfterFinished, int32(120))

	assert.Equal(t, ray.Spec.RayClusterSpec.HeadGroupSpec.Template.Spec.ServiceAccountName, serviceAccount)
	assert.Equal(t, ray.Spec.RayClusterSpec.HeadGroupSpec.RayStartParams,
		map[string]string{
			"dashboard-host": "0.0.0.0", "disable-usage-stats": "true", "include-dashboard": "true",
			"node-ip-address": "$MY_POD_IP",
		})
	assert.Equal(t, ray.Spec.RayClusterSpec.HeadGroupSpec.Template.Annotations, map[string]string{"annotation-1": "val1"})
	assert.Equal(t, ray.Spec.RayClusterSpec.HeadGroupSpec.Template.Labels, map[string]string{"label-1": "val1"})
	assert.Equal(t, ray.Spec.RayClusterSpec.HeadGroupSpec.Template.Spec.Tolerations, toleration)

	workerReplica := int32(3)
	assert.Equal(t, *ray.Spec.RayClusterSpec.WorkerGroupSpecs[0].Replicas, workerReplica)
	assert.Equal(t, *ray.Spec.RayClusterSpec.WorkerGroupSpecs[0].MinReplicas, workerReplica)
	assert.Equal(t, *ray.Spec.RayClusterSpec.WorkerGroupSpecs[0].MaxReplicas, workerReplica)
	assert.Equal(t, ray.Spec.RayClusterSpec.WorkerGroupSpecs[0].GroupName, workerGroupName)
	assert.Equal(t, ray.Spec.RayClusterSpec.WorkerGroupSpecs[0].Template.Spec.ServiceAccountName, serviceAccount)
	assert.Equal(t, ray.Spec.RayClusterSpec.WorkerGroupSpecs[0].RayStartParams, map[string]string{"disable-usage-stats": "true", "node-ip-address": "$MY_POD_IP"})
	assert.Equal(t, ray.Spec.RayClusterSpec.WorkerGroupSpecs[0].Template.Annotations, map[string]string{"annotation-1": "val1"})
	assert.Equal(t, ray.Spec.RayClusterSpec.WorkerGroupSpecs[0].Template.Labels, map[string]string{"label-1": "val1"})
	assert.Equal(t, ray.Spec.RayClusterSpec.WorkerGroupSpecs[0].Template.Spec.Tolerations, toleration)
}

func TestInjectLogsSidecar(t *testing.T) {
	rayJobObj := transformRayJobToCustomObj(dummyRayCustomObj())
	params := []struct {
		name         string
		taskTemplate core.TaskTemplate
		// primaryContainerName string
		logsSidecarCfg                       *corev1.Container
		expectedVolumes                      []corev1.Volume
		expectedPrimaryContainerVolumeMounts []corev1.VolumeMount
		expectedLogsSidecarVolumeMounts      []corev1.VolumeMount
	}{
		{
			"container target",
			core.TaskTemplate{
				Id: &core.Identifier{Name: "ray-id"},
				Target: &core.TaskTemplate_Container{
					Container: &core.Container{
						Image: testImage,
						Args:  testArgs,
					},
				},
				Custom: rayJobObj,
			},
			&corev1.Container{
				Name:  "logs-sidecar",
				Image: "test-image",
			},
			[]corev1.Volume{
				{
					Name: "system-ray-state",
					VolumeSource: corev1.VolumeSource{
						EmptyDir: &corev1.EmptyDirVolumeSource{},
					},
				},
			},
			[]corev1.VolumeMount{
				{
					Name:      "system-ray-state",
					MountPath: "/tmp/ray",
				},
			},
			[]corev1.VolumeMount{
				{
					Name:      "system-ray-state",
					MountPath: "/tmp/ray",
					ReadOnly:  true,
				},
			},
		},
		{
			"container target with no sidecar",
			core.TaskTemplate{
				Id: &core.Identifier{Name: "ray-id"},
				Target: &core.TaskTemplate_Container{
					Container: &core.Container{
						Image: testImage,
						Args:  testArgs,
					},
				},
				Custom: rayJobObj,
			},
			nil,
			nil,
			nil,
			nil,
		},
		{
			"pod target",
			core.TaskTemplate{
				Id: &core.Identifier{Name: "ray-id"},
				Target: transformPodSpecToTaskTemplateTarget(&corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "main",
							Image: "primary-image",
						},
					},
				}),
				Custom: rayJobObj,
				Config: map[string]string{
					flytek8s.PrimaryContainerKey: "main",
				},
			},
			&corev1.Container{
				Name:  "logs-sidecar",
				Image: "test-image",
			},
			[]corev1.Volume{
				{
					Name: "system-ray-state",
					VolumeSource: corev1.VolumeSource{
						EmptyDir: &corev1.EmptyDirVolumeSource{},
					},
				},
			},
			[]corev1.VolumeMount{
				{
					Name:      "system-ray-state",
					MountPath: "/tmp/ray",
				},
			},
			[]corev1.VolumeMount{
				{
					Name:      "system-ray-state",
					MountPath: "/tmp/ray",
					ReadOnly:  true,
				},
			},
		},
		{
			"pod target with existing ray state volume",
			core.TaskTemplate{
				Id: &core.Identifier{Name: "ray-id"},
				Target: transformPodSpecToTaskTemplateTarget(&corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "main",
							Image: "primary-image",
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "test-vol",
									MountPath: "/tmp/ray",
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "test-vol",
							VolumeSource: corev1.VolumeSource{
								EmptyDir: &corev1.EmptyDirVolumeSource{},
							},
						},
					},
				}),
				Custom: rayJobObj,
				Config: map[string]string{
					flytek8s.PrimaryContainerKey: "main",
				},
			},
			&corev1.Container{
				Name:  "logs-sidecar",
				Image: "test-image",
			},
			[]corev1.Volume{
				{
					Name: "test-vol",
					VolumeSource: corev1.VolumeSource{
						EmptyDir: &corev1.EmptyDirVolumeSource{},
					},
				},
			},
			[]corev1.VolumeMount{
				{
					Name:      "test-vol",
					MountPath: "/tmp/ray",
				},
			},
			[]corev1.VolumeMount{
				{
					Name:      "test-vol",
					MountPath: "/tmp/ray",
					ReadOnly:  true,
				},
			},
		},
	}

	for _, p := range params {
		t.Run(p.name, func(t *testing.T) {
			assert.NoError(t, SetConfig(&Config{
				LogsSidecar: p.logsSidecarCfg,
			}))
			taskContext := dummyRayTaskContext(&p.taskTemplate, resourceRequirements, nil, "", serviceAccount)
			rayJobResourceHandler := rayJobResourceHandler{}
			r, err := rayJobResourceHandler.BuildResource(context.TODO(), taskContext)
			assert.Nil(t, err)
			assert.NotNil(t, r)
			rayJob, ok := r.(*rayv1.RayJob)
			assert.True(t, ok)

			headPodSpec := rayJob.Spec.RayClusterSpec.HeadGroupSpec.Template.Spec

			// Check volumes
			assert.EqualValues(t, p.expectedVolumes, headPodSpec.Volumes)

			// Check containers and respective volume mounts
			foundPrimaryContainer := false
			foundLogsSidecar := false
			for _, cnt := range headPodSpec.Containers {
				if cnt.Name == "ray-head" {
					foundPrimaryContainer = true
					assert.EqualValues(
						t,
						p.expectedPrimaryContainerVolumeMounts,
						cnt.VolumeMounts,
					)
				}
				if p.logsSidecarCfg != nil && cnt.Name == p.logsSidecarCfg.Name {
					foundLogsSidecar = true
					assert.EqualValues(
						t,
						p.expectedLogsSidecarVolumeMounts,
						cnt.VolumeMounts,
					)
				}
			}
			assert.Equal(t, true, foundPrimaryContainer)
			assert.Equal(t, p.logsSidecarCfg != nil, foundLogsSidecar)
		})
	}
}

func newPluginContext(pluginState k8s.PluginState) k8s.PluginContext {
	plg := &mocks2.PluginContext{}

	taskExecID := &mocks.TaskExecutionID{}
	taskExecID.OnGetID().Return(core.TaskExecutionIdentifier{
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
		RetryAttempt: 1,
	})
	taskExecID.OnGetUniqueNodeID().Return("unique-node")
	taskExecID.OnGetGeneratedName().Return("generated-name")

	tskCtx := &mocks.TaskExecutionMetadata{}
	tskCtx.OnGetTaskExecutionID().Return(taskExecID)
	plg.OnTaskExecutionMetadata().Return(tskCtx)

	pluginStateReaderMock := mocks.PluginStateReader{}
	pluginStateReaderMock.On("Get", mock.AnythingOfType(reflect.TypeOf(&pluginState).String())).Return(
		func(v interface{}) uint8 {
			*(v.(*k8s.PluginState)) = pluginState
			return 0
		},
		func(v interface{}) error {
			return nil
		})

	plg.OnPluginStateReader().Return(&pluginStateReaderMock)

	return plg
}

func init() {
	f := defaultConfig
	f.Logs = logs.LogConfig{
		IsKubernetesEnabled: true,
	}

	if err := SetConfig(&f); err != nil {
		panic(err)
	}
}

func TestGetTaskPhase(t *testing.T) {
	ctx := context.Background()
	rayJobResourceHandler := rayJobResourceHandler{}
	pluginCtx := newPluginContext(k8s.PluginState{})

	testCases := []struct {
		rayJobPhase       rayv1.JobDeploymentStatus
		expectedCorePhase pluginsCore.Phase
		expectedError     bool
	}{
		{rayv1.JobDeploymentStatusInitializing, pluginsCore.PhaseInitializing, false},
		{rayv1.JobDeploymentStatusRunning, pluginsCore.PhaseRunning, false},
		{rayv1.JobDeploymentStatusComplete, pluginsCore.PhaseSuccess, false},
		{rayv1.JobDeploymentStatusFailed, pluginsCore.PhasePermanentFailure, false},
		{rayv1.JobDeploymentStatusSuspended, pluginsCore.PhaseQueued, false},
		{rayv1.JobDeploymentStatusSuspending, pluginsCore.PhaseQueued, false},
	}

	for _, tc := range testCases {
		t.Run("TestGetTaskPhase_"+string(tc.rayJobPhase), func(t *testing.T) {
			rayObject := &rayv1.RayJob{}
			rayObject.Status.JobDeploymentStatus = tc.rayJobPhase
			startTime := metav1.NewTime(time.Now())
			rayObject.Status.StartTime = &startTime
			phaseInfo, err := rayJobResourceHandler.GetTaskPhase(ctx, pluginCtx, rayObject)
			if tc.expectedError {
				assert.Error(t, err)
			} else {
				assert.Nil(t, err)
			}
			assert.Equal(t, tc.expectedCorePhase.String(), phaseInfo.Phase().String())
		})
	}
}

func TestGetTaskPhaseIncreasePhaseVersion(t *testing.T) {
	rayJobResourceHandler := rayJobResourceHandler{}

	ctx := context.TODO()

	pluginState := k8s.PluginState{
		Phase:        pluginsCore.PhaseInitializing,
		PhaseVersion: pluginsCore.DefaultPhaseVersion,
		Reason:       "task submitted to K8s",
	}
	pluginCtx := newPluginContext(pluginState)

	rayObject := &rayv1.RayJob{}
	rayObject.Status.JobDeploymentStatus = rayv1.JobDeploymentStatusInitializing
	phaseInfo, err := rayJobResourceHandler.GetTaskPhase(ctx, pluginCtx, rayObject)

	assert.NoError(t, err)
	assert.Equal(t, phaseInfo.Version(), pluginsCore.DefaultPhaseVersion+1)
}

func TestGetEventInfo_LogTemplates(t *testing.T) {
	pluginCtx := newPluginContext(k8s.PluginState{})
	testCases := []struct {
		name             string
		rayJob           rayv1.RayJob
		logPlugin        tasklog.TemplateLogPlugin
		expectedTaskLogs []*core.TaskLog
	}{
		{
			name: "namespace",
			rayJob: rayv1.RayJob{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test-namespace",
				},
			},
			logPlugin: tasklog.TemplateLogPlugin{
				DisplayName:  "namespace",
				TemplateURIs: []tasklog.TemplateURI{"http://test/{{ .namespace }}"},
			},
			expectedTaskLogs: []*core.TaskLog{
				{
					Name: "namespace",
					Uri:  "http://test/test-namespace",
				},
			},
		},
		{
			name:   "task execution ID",
			rayJob: rayv1.RayJob{},
			logPlugin: tasklog.TemplateLogPlugin{
				DisplayName: "taskExecID",
				TemplateURIs: []tasklog.TemplateURI{
					"http://test/projects/{{ .executionProject }}/domains/{{ .executionDomain }}/executions/{{ .executionName }}/nodeId/{{ .nodeID }}/taskId/{{ .taskID }}/attempt/{{ .taskRetryAttempt }}",
				},
			},
			expectedTaskLogs: []*core.TaskLog{
				{
					Name: "taskExecID",
					Uri:  "http://test/projects/my-execution-project/domains/my-execution-domain/executions/my-execution-name/nodeId/unique-node/taskId/my-task-name/attempt/1",
				},
			},
		},
		{
			name: "ray cluster name",
			rayJob: rayv1.RayJob{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test-namespace",
				},
				Status: rayv1.RayJobStatus{
					RayClusterName: "ray-cluster",
				},
			},
			logPlugin: tasklog.TemplateLogPlugin{
				DisplayName:  "ray cluster name",
				TemplateURIs: []tasklog.TemplateURI{"http://test/{{ .namespace }}/{{ .rayClusterName }}"},
			},
			expectedTaskLogs: []*core.TaskLog{
				{
					Name: "ray cluster name",
					Uri:  "http://test/test-namespace/ray-cluster",
				},
			},
		},
		{
			name: "ray job ID",
			rayJob: rayv1.RayJob{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test-namespace",
				},
				Status: rayv1.RayJobStatus{
					JobId: "ray-job-1",
				},
			},
			logPlugin: tasklog.TemplateLogPlugin{
				DisplayName:  "ray job ID",
				TemplateURIs: []tasklog.TemplateURI{"http://test/{{ .namespace }}/{{ .rayJobID }}"},
			},
			expectedTaskLogs: []*core.TaskLog{
				{
					Name: "ray job ID",
					Uri:  "http://test/test-namespace/ray-job-1",
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ti, err := getEventInfoForRayJob(
				logs.LogConfig{Templates: []tasklog.TemplateLogPlugin{tc.logPlugin}},
				pluginCtx,
				&tc.rayJob,
			)
			assert.NoError(t, err)
			assert.Equal(t, tc.expectedTaskLogs, ti.Logs)
		})
	}
}

func TestGetEventInfo_LogTemplates_V1(t *testing.T) {
	pluginCtx := newPluginContext(k8s.PluginState{})
	testCases := []struct {
		name             string
		rayJob           rayv1.RayJob
		logPlugin        tasklog.TemplateLogPlugin
		expectedTaskLogs []*core.TaskLog
	}{
		{
			name: "namespace",
			rayJob: rayv1.RayJob{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test-namespace",
				},
			},
			logPlugin: tasklog.TemplateLogPlugin{
				DisplayName:  "namespace",
				TemplateURIs: []tasklog.TemplateURI{"http://test/{{ .namespace }}"},
			},
			expectedTaskLogs: []*core.TaskLog{
				{
					Name: "namespace",
					Uri:  "http://test/test-namespace",
				},
			},
		},
		{
			name:   "task execution ID",
			rayJob: rayv1.RayJob{},
			logPlugin: tasklog.TemplateLogPlugin{
				DisplayName: "taskExecID",
				TemplateURIs: []tasklog.TemplateURI{
					"http://test/projects/{{ .executionProject }}/domains/{{ .executionDomain }}/executions/{{ .executionName }}/nodeId/{{ .nodeID }}/taskId/{{ .taskID }}/attempt/{{ .taskRetryAttempt }}",
				},
			},
			expectedTaskLogs: []*core.TaskLog{
				{
					Name: "taskExecID",
					Uri:  "http://test/projects/my-execution-project/domains/my-execution-domain/executions/my-execution-name/nodeId/unique-node/taskId/my-task-name/attempt/1",
				},
			},
		},
		{
			name: "ray cluster name",
			rayJob: rayv1.RayJob{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test-namespace",
				},
				Status: rayv1.RayJobStatus{
					RayClusterName: "ray-cluster",
				},
			},
			logPlugin: tasklog.TemplateLogPlugin{
				DisplayName:  "ray cluster name",
				TemplateURIs: []tasklog.TemplateURI{"http://test/{{ .namespace }}/{{ .rayClusterName }}"},
			},
			expectedTaskLogs: []*core.TaskLog{
				{
					Name: "ray cluster name",
					Uri:  "http://test/test-namespace/ray-cluster",
				},
			},
		},
		{
			name: "ray job ID",
			rayJob: rayv1.RayJob{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test-namespace",
				},
				Status: rayv1.RayJobStatus{
					JobId: "ray-job-1",
				},
			},
			logPlugin: tasklog.TemplateLogPlugin{
				DisplayName:  "ray job ID",
				TemplateURIs: []tasklog.TemplateURI{"http://test/{{ .namespace }}/{{ .rayJobID }}"},
			},
			expectedTaskLogs: []*core.TaskLog{
				{
					Name: "ray job ID",
					Uri:  "http://test/test-namespace/ray-job-1",
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ti, err := getEventInfoForRayJob(
				logs.LogConfig{Templates: []tasklog.TemplateLogPlugin{tc.logPlugin}},
				pluginCtx,
				&tc.rayJob,
			)
			assert.NoError(t, err)
			assert.Equal(t, tc.expectedTaskLogs, ti.Logs)
		})
	}
}

func TestGetEventInfo_DashboardURL(t *testing.T) {
	pluginCtx := newPluginContext(k8s.PluginState{})
	testCases := []struct {
		name                 string
		rayJob               rayv1.RayJob
		dashboardURLTemplate tasklog.TemplateLogPlugin
		expectedTaskLogs     []*core.TaskLog
	}{
		{
			name: "dashboard URL displayed",
			rayJob: rayv1.RayJob{
				Status: rayv1.RayJobStatus{
					DashboardURL: "exists",
					JobStatus:    rayv1.JobStatusRunning,
				},
			},
			dashboardURLTemplate: tasklog.TemplateLogPlugin{
				DisplayName:  "Ray Dashboard",
				TemplateURIs: []tasklog.TemplateURI{"http://test/{{.generatedName}}"},
			},
			expectedTaskLogs: []*core.TaskLog{
				{
					Name: "Ray Dashboard",
					Uri:  "http://test/generated-name",
				},
			},
		},
		{
			name: "dashboard URL is not displayed",
			rayJob: rayv1.RayJob{
				Status: rayv1.RayJobStatus{
					JobStatus: rayv1.JobStatusPending,
				},
			},
			dashboardURLTemplate: tasklog.TemplateLogPlugin{
				DisplayName:  "dummy",
				TemplateURIs: []tasklog.TemplateURI{"http://dummy"},
			},
			expectedTaskLogs: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.NoError(t, SetConfig(&Config{DashboardURLTemplate: &tc.dashboardURLTemplate}))
			ti, err := getEventInfoForRayJob(logs.LogConfig{}, pluginCtx, &tc.rayJob)
			assert.NoError(t, err)
			assert.Equal(t, tc.expectedTaskLogs, ti.Logs)
		})
	}
}

func TestGetEventInfo_DashboardURL_V1(t *testing.T) {
	pluginCtx := newPluginContext(k8s.PluginState{})
	testCases := []struct {
		name                 string
		rayJob               rayv1.RayJob
		dashboardURLTemplate tasklog.TemplateLogPlugin
		expectedTaskLogs     []*core.TaskLog
	}{
		{
			name: "dashboard URL displayed",
			rayJob: rayv1.RayJob{
				Status: rayv1.RayJobStatus{
					DashboardURL: "exists",
					JobStatus:    rayv1.JobStatusRunning,
				},
			},
			dashboardURLTemplate: tasklog.TemplateLogPlugin{
				DisplayName:  "Ray Dashboard",
				TemplateURIs: []tasklog.TemplateURI{"http://test/{{.generatedName}}"},
			},
			expectedTaskLogs: []*core.TaskLog{
				{
					Name: "Ray Dashboard",
					Uri:  "http://test/generated-name",
				},
			},
		},
		{
			name: "dashboard URL is not displayed",
			rayJob: rayv1.RayJob{
				Status: rayv1.RayJobStatus{
					JobStatus: rayv1.JobStatusPending,
				},
			},
			dashboardURLTemplate: tasklog.TemplateLogPlugin{
				DisplayName:  "dummy",
				TemplateURIs: []tasklog.TemplateURI{"http://dummy"},
			},
			expectedTaskLogs: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.NoError(t, SetConfig(&Config{DashboardURLTemplate: &tc.dashboardURLTemplate}))
			ti, err := getEventInfoForRayJob(logs.LogConfig{}, pluginCtx, &tc.rayJob)
			assert.NoError(t, err)
			assert.Equal(t, tc.expectedTaskLogs, ti.Logs)
		})
	}
}

func TestGetPropertiesRay(t *testing.T) {
	rayJobResourceHandler := rayJobResourceHandler{}
	expected := k8s.PluginProperties{}
	assert.Equal(t, expected, rayJobResourceHandler.GetProperties())
}

func transformStructToStructPB(t *testing.T, obj interface{}) *structpb.Struct {
	data, err := json.Marshal(obj)
	assert.Nil(t, err)
	podSpecMap := make(map[string]interface{})
	err = json.Unmarshal(data, &podSpecMap)
	assert.Nil(t, err)
	s, err := structpb.NewStruct(podSpecMap)
	assert.Nil(t, err)
	return s
}
