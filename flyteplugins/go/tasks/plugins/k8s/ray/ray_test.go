package ray

import (
	"context"
	"testing"

	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/flytek8s/config"

	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/utils"
	"github.com/golang/protobuf/jsonpb"
	structpb "github.com/golang/protobuf/ptypes/struct"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/plugins"
	pluginsCore "github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/core/mocks"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/flytek8s"
	pluginIOMocks "github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/io/mocks"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/k8s"
	rayv1alpha1 "github.com/ray-project/kuberay/ray-operator/apis/ray/v1alpha1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const testImage = "image://"
const serviceAccount = "ray_sa"
const serviceAccountOverride = "ray_sa_override"
const namespaceOverride = "ray_namespace_override"

var (
	dummyEnvVars = []*core.KeyValuePair{
		{Key: "Env_Var", Value: "Env_Val"},
	}

	testArgs = []string{
		"test-args",
	}

	headResourceOverride = core.Resources{
		Requests: []*core.Resources_ResourceEntry{
			{
				Name:  core.Resources_CPU,
				Value: "1000m",
			},
			{
				Name:  core.Resources_MEMORY,
				Value: "2Gi",
			},
		},
		Limits: []*core.Resources_ResourceEntry{
			{
				Name:  core.Resources_CPU,
				Value: "2000m",
			},
			{
				Name:  core.Resources_MEMORY,
				Value: "4Gi",
			},
		},
	}

	workerResourceOverride = core.Resources{
		Requests: []*core.Resources_ResourceEntry{
			{
				Name:  core.Resources_CPU,
				Value: "5",
			},
			{
				Name:  core.Resources_MEMORY,
				Value: "10G",
			},
		},
		Limits: []*core.Resources_ResourceEntry{
			{
				Name:  core.Resources_CPU,
				Value: "10",
			},
			{
				Name:  core.Resources_MEMORY,
				Value: "20G",
			},
		},
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

func dummyRayCustomObj() *plugins.RayJob {
	return &plugins.RayJob{
		RayCluster: &plugins.RayCluster{
			HeadGroupSpec:   &plugins.HeadGroupSpec{RayStartParams: map[string]string{"num-cpus": "1"}},
			WorkerGroupSpec: []*plugins.WorkerGroupSpec{{GroupName: workerGroupName, Replicas: 3}},
		},
	}
}

func dummyRayCustomObjWithOverrides() *plugins.RayJob {
	return &plugins.RayJob{
		RayCluster: &plugins.RayCluster{
			K8SServiceAccount: serviceAccountOverride,
			Namespace: namespaceOverride,
			HeadGroupSpec:   &plugins.HeadGroupSpec{RayStartParams: map[string]string{"num-cpus": "1"}, Resources: &headResourceOverride},
			WorkerGroupSpec: []*plugins.WorkerGroupSpec{{GroupName: workerGroupName, Replicas: 3, Resources: &workerResourceOverride}},
		},
	}
}

func dummyRayTaskTemplate(id string, rayJobObj *plugins.RayJob) *core.TaskTemplate {

	ptObjJSON, err := utils.MarshalToString(rayJobObj)
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

func dummyRayTaskContext(taskTemplate *core.TaskTemplate) pluginsCore.TaskExecutionContext {
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
	taskExecutionMetadata.OnGetSecurityContext().Return(core.SecurityContext{
		RunAs: &core.Identity{K8SServiceAccount: serviceAccount},
	})
	taskExecutionMetadata.OnGetEnvironmentVariables().Return(nil)
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

	RayResource, err := rayJobResourceHandler.BuildResource(context.TODO(), dummyRayTaskContext(taskTemplate))
	assert.Nil(t, err)

	assert.NotNil(t, RayResource)
	ray, ok := RayResource.(*rayv1alpha1.RayJob)
	assert.True(t, ok)

	headReplica := int32(1)
	assert.Equal(t, &headReplica, ray.Spec.RayClusterSpec.HeadGroupSpec.Replicas)
	assert.Equal(t, serviceAccount, ray.Spec.RayClusterSpec.HeadGroupSpec.Template.Spec.ServiceAccountName)
	assert.Equal(t, map[string]string{"dashboard-host": "0.0.0.0", "include-dashboard": "true", "node-ip-address": "$MY_POD_IP", "num-cpus": "1"},
		ray.Spec.RayClusterSpec.HeadGroupSpec.RayStartParams)
	assert.Equal(t, map[string]string{"annotation-1": "val1"}, ray.Spec.RayClusterSpec.HeadGroupSpec.Template.Annotations)
	assert.Equal(t, map[string]string{"label-1": "val1"}, ray.Spec.RayClusterSpec.HeadGroupSpec.Template.Labels)
	assert.Equal(t, toleration, ray.Spec.RayClusterSpec.HeadGroupSpec.Template.Spec.Tolerations)

	workerReplica := int32(3)
	assert.Equal(t, &workerReplica, ray.Spec.RayClusterSpec.WorkerGroupSpecs[0].Replicas)
	assert.Equal(t, &workerReplica, ray.Spec.RayClusterSpec.WorkerGroupSpecs[0].MinReplicas)
	assert.Equal(t, &workerReplica, ray.Spec.RayClusterSpec.WorkerGroupSpecs[0].MaxReplicas)
	assert.Equal(t, workerGroupName, ray.Spec.RayClusterSpec.WorkerGroupSpecs[0].GroupName)
	assert.Equal(t, serviceAccount, ray.Spec.RayClusterSpec.WorkerGroupSpecs[0].Template.Spec.ServiceAccountName)
	assert.Equal(t, map[string]string{"node-ip-address": "$MY_POD_IP"}, ray.Spec.RayClusterSpec.WorkerGroupSpecs[0].RayStartParams)
	assert.Equal(t, map[string]string{"annotation-1": "val1"}, ray.Spec.RayClusterSpec.WorkerGroupSpecs[0].Template.Annotations)
	assert.Equal(t, map[string]string{"label-1": "val1"}, ray.Spec.RayClusterSpec.WorkerGroupSpecs[0].Template.Labels)
	assert.Equal(t, toleration, ray.Spec.RayClusterSpec.WorkerGroupSpecs[0].Template.Spec.Tolerations)
}

func TestBuildResourceRayWithOverrides(t *testing.T) {
	rayJobResourceHandler := rayJobResourceHandler{}
	taskTemplate := dummyRayTaskTemplate("ray-id", dummyRayCustomObjWithOverrides())
	expectedHeadResources, _ := flytek8s.ToK8sResourceRequirements(&headResourceOverride)
	expectedWorkerResources, _ := flytek8s.ToK8sResourceRequirements(&workerResourceOverride)
	toleration := []corev1.Toleration{{
		Key:      "storage",
		Value:    "dedicated",
		Operator: corev1.TolerationOpExists,
		Effect:   corev1.TaintEffectNoSchedule,
	}}
	err := config.SetK8sPluginConfig(&config.K8sPluginConfig{DefaultTolerations: toleration})
	assert.Nil(t, err)

	RayResource, err := rayJobResourceHandler.BuildResource(context.TODO(), dummyRayTaskContext(taskTemplate))
	assert.Nil(t, err)

	assert.NotNil(t, RayResource)
	ray, ok := RayResource.(*rayv1alpha1.RayJob)
	assert.True(t, ok)

	headReplica := int32(1)
	assert.Equal(t, namespaceOverride, ray.Spec.RayClusterSpec.HeadGroupSpec.Template.ObjectMeta.Namespace)
	assert.Equal(t, &headReplica, ray.Spec.RayClusterSpec.HeadGroupSpec.Replicas)
	assert.Equal(t, serviceAccountOverride, ray.Spec.RayClusterSpec.HeadGroupSpec.Template.Spec.ServiceAccountName)
	assert.Equal(t, map[string]string{"dashboard-host": "0.0.0.0", "include-dashboard": "true", "node-ip-address": "$MY_POD_IP", "num-cpus": "1"},
		ray.Spec.RayClusterSpec.HeadGroupSpec.RayStartParams)
	assert.Equal(t, map[string]string{"annotation-1": "val1"}, ray.Spec.RayClusterSpec.HeadGroupSpec.Template.Annotations)
	assert.Equal(t, map[string]string{"label-1": "val1"}, ray.Spec.RayClusterSpec.HeadGroupSpec.Template.Labels)
	assert.Equal(t, toleration, ray.Spec.RayClusterSpec.HeadGroupSpec.Template.Spec.Tolerations)
	assert.Equal(t, *expectedHeadResources, ray.Spec.RayClusterSpec.HeadGroupSpec.Template.Spec.Containers[0].Resources)

	workerReplica := int32(3)
	assert.Equal(t, namespaceOverride, ray.Spec.RayClusterSpec.WorkerGroupSpecs[0].Template.ObjectMeta.Namespace)
	assert.Equal(t, &workerReplica, ray.Spec.RayClusterSpec.WorkerGroupSpecs[0].Replicas)
	assert.Equal(t, &workerReplica, ray.Spec.RayClusterSpec.WorkerGroupSpecs[0].MinReplicas)
	assert.Equal(t, &workerReplica, ray.Spec.RayClusterSpec.WorkerGroupSpecs[0].MaxReplicas)
	assert.Equal(t, workerGroupName, ray.Spec.RayClusterSpec.WorkerGroupSpecs[0].GroupName)
	assert.Equal(t, serviceAccountOverride, ray.Spec.RayClusterSpec.WorkerGroupSpecs[0].Template.Spec.ServiceAccountName)
	assert.Equal(t, map[string]string{"node-ip-address": "$MY_POD_IP"}, ray.Spec.RayClusterSpec.WorkerGroupSpecs[0].RayStartParams)
	assert.Equal(t, map[string]string{"annotation-1": "val1"}, ray.Spec.RayClusterSpec.WorkerGroupSpecs[0].Template.Annotations)
	assert.Equal(t, map[string]string{"label-1": "val1"}, ray.Spec.RayClusterSpec.WorkerGroupSpecs[0].Template.Labels)
	assert.Equal(t, toleration, ray.Spec.RayClusterSpec.WorkerGroupSpecs[0].Template.Spec.Tolerations)
	assert.Equal(t, *expectedWorkerResources, ray.Spec.RayClusterSpec.WorkerGroupSpecs[0].Template.Spec.Containers[0].Resources)
}

func TestGetPropertiesRay(t *testing.T) {
	rayJobResourceHandler := rayJobResourceHandler{}
	expected := k8s.PluginProperties{}
	assert.Equal(t, expected, rayJobResourceHandler.GetProperties())
}
