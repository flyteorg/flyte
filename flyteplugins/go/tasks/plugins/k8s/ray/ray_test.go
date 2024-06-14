package ray

import (
	"context"
	"testing"
	"time"

	structpb "github.com/golang/protobuf/ptypes/struct"
	rayv1 "github.com/ray-project/kuberay/ray-operator/apis/ray/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

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
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/utils"
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

func dummyRayTaskContext(taskTemplate *core.TaskTemplate, resources *corev1.ResourceRequirements, extendedResources *core.ExtendedResources, containerImage string) pluginsCore.TaskExecutionContext {
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
	taskExecutionMetadata.OnGetOverrides().Return(overrides)
	taskExecutionMetadata.OnGetK8sServiceAccount().Return(serviceAccount)
	taskExecutionMetadata.OnGetPlatformResources().Return(&corev1.ResourceRequirements{})
	taskExecutionMetadata.OnGetSecurityContext().Return(core.SecurityContext{
		RunAs: &core.Identity{K8SServiceAccount: serviceAccount},
	})
	taskExecutionMetadata.OnGetEnvironmentVariables().Return(nil)
	taskExecutionMetadata.OnGetConsoleURL().Return("")
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

	RayResource, err := rayJobResourceHandler.BuildResource(context.TODO(), dummyRayTaskContext(taskTemplate, resourceRequirements, nil, ""))
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
			taskContext := dummyRayTaskContext(taskTemplate, f.resources, nil, f.containerImageOverride)
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

func TestBuildResourceRayExtendedResources(t *testing.T) {
	assert.NoError(t, config.SetK8sPluginConfig(&config.K8sPluginConfig{
		GpuDeviceNodeLabel:        "gpu-node-label",
		GpuPartitionSizeNodeLabel: "gpu-partition-size",
		GpuResourceName:           flytek8s.ResourceNvidiaGPU,
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
			},
		},
	}

	for _, p := range params {
		t.Run(p.name, func(t *testing.T) {
			taskTemplate := dummyRayTaskTemplate("ray-id", dummyRayCustomObj())
			taskTemplate.ExtendedResources = p.extendedResourcesBase
			taskContext := dummyRayTaskContext(taskTemplate, p.resources, p.extendedResourcesOverride, "")
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

	RayResource, err := rayJobResourceHandler.BuildResource(context.TODO(), dummyRayTaskContext(taskTemplate, resourceRequirements, nil, ""))
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
			taskContext := dummyRayTaskContext(&p.taskTemplate, resourceRequirements, nil, "")
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

func newPluginContext() k8s.PluginContext {
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
	pluginCtx := newPluginContext()

	testCases := []struct {
		rayJobPhase       rayv1.JobDeploymentStatus
		expectedCorePhase pluginsCore.Phase
		expectedError     bool
	}{
		{rayv1.JobDeploymentStatusInitializing, pluginsCore.PhaseInitializing, false},
		{rayv1.JobDeploymentStatusRunning, pluginsCore.PhaseRunning, false},
		{rayv1.JobDeploymentStatusComplete, pluginsCore.PhaseSuccess, false},
		{rayv1.JobDeploymentStatusFailed, pluginsCore.PhasePermanentFailure, false},
		{rayv1.JobDeploymentStatusSuspended, pluginsCore.PhaseUndefined, true},
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

func TestGetEventInfo_LogTemplates(t *testing.T) {
	pluginCtx := newPluginContext()
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
	pluginCtx := newPluginContext()
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
	pluginCtx := newPluginContext()
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
	pluginCtx := newPluginContext()
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
