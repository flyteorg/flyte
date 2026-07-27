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
	"google.golang.org/protobuf/types/known/timestamppb"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/logs"
	pluginsCore "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/core/mocks"
	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/flytek8s"
	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/flytek8s/config"
	pluginIOMocks "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/io/mocks"
	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/k8s"
	k8smocks "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/k8s/mocks"
	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/tasklog"
	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/utils"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/core"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/plugins"
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
	return dummyRayTaskContextInterruptible(taskTemplate, resources, extendedResources, containerImage, serviceAccount, true)
}

func dummyRayTaskContextInterruptible(taskTemplate *core.TaskTemplate, resources *corev1.ResourceRequirements, extendedResources *core.ExtendedResources, containerImage, serviceAccount string, interruptible bool) pluginsCore.TaskExecutionContext {
	taskCtx := &mocks.TaskExecutionContext{}
	inputReader := &pluginIOMocks.InputReader{}
	inputReader.EXPECT().GetInputPrefixPath().Return("/input/prefix")
	inputReader.EXPECT().GetInputPath().Return("/input")
	inputReader.EXPECT().Get(mock.Anything).Return(&core.LiteralMap{}, nil)
	taskCtx.EXPECT().InputReader().Return(inputReader)

	outputReader := &pluginIOMocks.OutputWriter{}
	outputReader.EXPECT().GetOutputPath().Return("/data/outputs.pb")
	outputReader.EXPECT().GetOutputPrefixPath().Return("/data/")
	outputReader.EXPECT().GetRawOutputPrefix().Return("")
	outputReader.EXPECT().GetCheckpointPrefix().Return("/checkpoint")
	outputReader.EXPECT().GetPreviousCheckpointsPrefix().Return("/prev")
	taskCtx.EXPECT().OutputWriter().Return(outputReader)

	taskReader := &mocks.TaskReader{}
	taskReader.EXPECT().Read(mock.Anything).Return(taskTemplate, nil)
	taskCtx.EXPECT().TaskReader().Return(taskReader)

	tID := &mocks.TaskExecutionID{}
	tID.EXPECT().GetID().Return(&core.TaskExecutionIdentifier{
		NodeExecutionId: &core.NodeExecutionIdentifier{
			ExecutionId: &core.WorkflowExecutionIdentifier{
				Name:    "my_name",
				Project: "my_project",
				Domain:  "my_domain",
			},
		},
	})
	tID.EXPECT().GetGeneratedName().Return("some-acceptable-name")

	overrides := &mocks.TaskOverrides{}
	overrides.EXPECT().GetResources().Return(resources)
	overrides.EXPECT().GetExtendedResources().Return(extendedResources)
	overrides.EXPECT().GetContainerImage().Return(containerImage)
	overrides.EXPECT().GetPodTemplate().Return(nil)

	taskExecutionMetadata := &mocks.TaskExecutionMetadata{}
	taskExecutionMetadata.EXPECT().GetTaskExecutionID().Return(tID)
	taskExecutionMetadata.EXPECT().GetNamespace().Return("test-namespace")
	taskExecutionMetadata.EXPECT().GetAnnotations().Return(map[string]string{"annotation-1": "val1"})
	taskExecutionMetadata.EXPECT().GetLabels().Return(map[string]string{"label-1": "val1"})
	taskExecutionMetadata.EXPECT().GetOwnerReference().Return(metav1.OwnerReference{
		Kind: "node",
		Name: "blah",
	})
	taskExecutionMetadata.EXPECT().IsInterruptible().Return(interruptible)
	taskExecutionMetadata.EXPECT().GetOverrides().Return(overrides)
	taskExecutionMetadata.EXPECT().GetK8sServiceAccount().Return(serviceAccount)
	taskExecutionMetadata.EXPECT().GetPlatformResources().Return(&corev1.ResourceRequirements{})
	taskExecutionMetadata.EXPECT().GetSecurityContext().Return(&core.SecurityContext{
		RunAs: &core.Identity{K8SServiceAccount: serviceAccount},
	})
	taskExecutionMetadata.EXPECT().GetEnvironmentVariables().Return(nil)
	taskExecutionMetadata.EXPECT().GetConsoleURL().Return("")
	taskCtx.EXPECT().TaskExecutionMetadata().Return(taskExecutionMetadata)
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

var (
	interruptibleNSR = &corev1.NodeSelectorRequirement{
		Key:      "interruptible",
		Operator: corev1.NodeSelectorOpIn,
		Values:   []string{"true"},
	}
	nonInterruptibleNSR = &corev1.NodeSelectorRequirement{
		Key:      "interruptible",
		Operator: corev1.NodeSelectorOpIn,
		Values:   []string{"false"},
	}
	interruptibleNodeSelector  = map[string]string{"interruptible-node": "true"}
	interruptibleTolerationVal = corev1.Toleration{
		Key:      "interruptible",
		Value:    "true",
		Operator: corev1.TolerationOpEqual,
		Effect:   corev1.TaintEffectNoSchedule,
	}
)

func interruptibleK8sPluginConfig() *config.K8sPluginConfig {
	return &config.K8sPluginConfig{
		InterruptibleNodeSelectorRequirement:    interruptibleNSR,
		NonInterruptibleNodeSelectorRequirement: nonInterruptibleNSR,
		InterruptibleNodeSelector:               interruptibleNodeSelector,
		InterruptibleTolerations:                []corev1.Toleration{interruptibleTolerationVal},
	}
}

// countRequirement returns how many times req appears across all node selector terms.
func countRequirement(terms []corev1.NodeSelectorTerm, req corev1.NodeSelectorRequirement) int {
	count := 0
	for _, term := range terms {
		for _, e := range term.MatchExpressions {
			if reflect.DeepEqual(e, req) {
				count++
			}
		}
	}
	return count
}

// assertEveryTermHasRequirement asserts that every node selector term contains req
// exactly once. This is what guarantees a non-interruptible pod cannot land on a
// spot node via an OR'd alternative term contributed by a custom k8s_pod overlay.
func assertEveryTermHasRequirement(t *testing.T, spec corev1.PodSpec, req corev1.NodeSelectorRequirement) {
	t.Helper()
	require.NotNil(t, spec.Affinity)
	require.NotNil(t, spec.Affinity.NodeAffinity)
	require.NotNil(t, spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution)
	terms := spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms
	require.NotEmpty(t, terms)
	for i, term := range terms {
		assert.Truef(t, containsRequirement(term.MatchExpressions, req),
			"node selector term %d is missing the expected scheduling requirement %v", i, req)
	}
}

func containsRequirement(reqs []corev1.NodeSelectorRequirement, req corev1.NodeSelectorRequirement) bool {
	for _, r := range reqs {
		if reflect.DeepEqual(r, req) {
			return true
		}
	}
	return false
}

func hasToleration(tolerations []corev1.Toleration, tol corev1.Toleration) bool {
	for _, t := range tolerations {
		if reflect.DeepEqual(t, tol) {
			return true
		}
	}
	return false
}

// TestBuildResourceRayInterruptible verifies that the task's interruptible flag is
// reflected on both the head and worker pod templates: interruptible tasks get the
// interruptible node selector requirement / node selector / tolerations, and
// non-interruptible tasks get the non-interruptible requirement and none of the
// interruptible scheduling hints.
func TestBuildResourceRayInterruptible(t *testing.T) {
	for _, interruptible := range []bool{true, false} {
		name := "non-interruptible"
		if interruptible {
			name = "interruptible"
		}
		t.Run(name, func(t *testing.T) {
			require.NoError(t, config.SetK8sPluginConfig(interruptibleK8sPluginConfig()))

			taskTemplate := dummyRayTaskTemplate("ray-id", dummyRayCustomObj())
			taskContext := dummyRayTaskContextInterruptible(taskTemplate, resourceRequirements, nil, "", serviceAccount, interruptible)
			r, err := rayJobResourceHandler{}.BuildResource(context.TODO(), taskContext)
			require.NoError(t, err)
			rayJob, ok := r.(*rayv1.RayJob)
			require.True(t, ok)

			expectedReq := *nonInterruptibleNSR
			if interruptible {
				expectedReq = *interruptibleNSR
			}

			specs := map[string]corev1.PodSpec{
				"head":   rayJob.Spec.RayClusterSpec.HeadGroupSpec.Template.Spec,
				"worker": rayJob.Spec.RayClusterSpec.WorkerGroupSpecs[0].Template.Spec,
			}
			for group, spec := range specs {
				t.Run(group, func(t *testing.T) {
					assertEveryTermHasRequirement(t, spec, expectedReq)
					// Idempotency: the requirement must appear exactly once per pod.
					assert.Equal(t, 1, countRequirement(
						spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms,
						expectedReq), "requirement should be applied exactly once")

					if interruptible {
						assert.Equal(t, "true", spec.NodeSelector["interruptible-node"])
						assert.True(t, hasToleration(spec.Tolerations, interruptibleTolerationVal))
						assert.Equal(t, 1, tolerationCount(spec.Tolerations, interruptibleTolerationVal),
							"interruptible toleration should be applied exactly once")
					} else {
						assert.NotContains(t, spec.NodeSelector, "interruptible-node")
						assert.False(t, hasToleration(spec.Tolerations, interruptibleTolerationVal))
					}
				})
			}
		})
	}
}

func tolerationCount(tolerations []corev1.Toleration, tol corev1.Toleration) int {
	count := 0
	for _, t := range tolerations {
		if reflect.DeepEqual(t, tol) {
			count++
		}
	}
	return count
}

// TestBuildResourceRayInterruptibleCustomAffinity is the regression test for the
// OR-hole: when a worker's custom k8s_pod carries its own node selector term, the
// custom-pod merge (mergo WithAppendSlice) appends it as a new OR'd term. Without
// re-applying interruptible scheduling after the merge, that appended term lacks
// the non-interruptible requirement, letting the pod schedule on a spot node.
func TestBuildResourceRayInterruptibleCustomAffinity(t *testing.T) {
	require.NoError(t, config.SetK8sPluginConfig(interruptibleK8sPluginConfig()))

	customAffinityPod := &core.K8SPod{
		PodSpec: transformStructToStructPB(t, &corev1.PodSpec{
			Containers: []corev1.Container{{Name: "ray-worker"}},
			Affinity: &corev1.Affinity{
				NodeAffinity: &corev1.NodeAffinity{
					RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
						NodeSelectorTerms: []corev1.NodeSelectorTerm{
							{
								MatchExpressions: []corev1.NodeSelectorRequirement{
									{
										Key:      "zone",
										Operator: corev1.NodeSelectorOpIn,
										Values:   []string{"us-east-1a"},
									},
								},
							},
						},
					},
				},
			},
		}),
	}

	rayJobObj := dummyRayCustomObj()
	rayJobObj.RayCluster.WorkerGroupSpec[0].K8SPod = customAffinityPod
	taskTemplate := dummyRayTaskTemplate("ray-id", rayJobObj)
	// interruptible=false: every resulting node selector term (the base term and the
	// appended custom term) must carry the non-interruptible requirement.
	taskContext := dummyRayTaskContextInterruptible(taskTemplate, resourceRequirements, nil, "", serviceAccount, false)

	r, err := rayJobResourceHandler{}.BuildResource(context.TODO(), taskContext)
	require.NoError(t, err)
	rayJob, ok := r.(*rayv1.RayJob)
	require.True(t, ok)

	workerSpec := rayJob.Spec.RayClusterSpec.WorkerGroupSpecs[0].Template.Spec
	terms := workerSpec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms
	// Sanity: the custom term was actually appended as a separate OR'd term.
	require.GreaterOrEqual(t, len(terms), 2, "expected the custom affinity term to be appended as a separate term")
	assertEveryTermHasRequirement(t, workerSpec, *nonInterruptibleNSR)
}

// TestBuildResourceRayDefaultAffinityDilution probes whether a platform-injected
// DefaultAffinity required node-selector term survives a worker's custom k8s_pod
// that carries its own affinity term. Same OR-append mechanism as the interruptible
// bug: if it reproduces, the appended custom term lacks the default-affinity
// requirement, letting the pod escape the cluster's default node constraint.
func TestBuildResourceRayDefaultAffinityDilution(t *testing.T) {
	defaultAffinityReq := corev1.NodeSelectorRequirement{
		Key:      "node-pool",
		Operator: corev1.NodeSelectorOpIn,
		Values:   []string{"flyte-dedicated"},
	}
	require.NoError(t, config.SetK8sPluginConfig(&config.K8sPluginConfig{
		DefaultAffinity: &corev1.Affinity{
			NodeAffinity: &corev1.NodeAffinity{
				RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
					NodeSelectorTerms: []corev1.NodeSelectorTerm{
						{MatchExpressions: []corev1.NodeSelectorRequirement{defaultAffinityReq}},
					},
				},
			},
		},
	}))

	// CPU-only so GPU handling does not pre-seed Affinity (which would skip
	// DefaultAffinity application in UpdatePod).
	cpuOnly := &corev1.ResourceRequirements{
		Limits:   corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("1000m"), corev1.ResourceMemory: resource.MustParse("1Gi")},
		Requests: corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("100m"), corev1.ResourceMemory: resource.MustParse("512Mi")},
	}

	customAffinityPod := &core.K8SPod{
		PodSpec: transformStructToStructPB(t, &corev1.PodSpec{
			Containers: []corev1.Container{{Name: "ray-worker"}},
			Affinity: &corev1.Affinity{
				NodeAffinity: &corev1.NodeAffinity{
					RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
						NodeSelectorTerms: []corev1.NodeSelectorTerm{
							{MatchExpressions: []corev1.NodeSelectorRequirement{
								{Key: "zone", Operator: corev1.NodeSelectorOpIn, Values: []string{"us-east-1a"}},
							}},
						},
					},
				},
			},
		}),
	}

	rayJobObj := dummyRayCustomObj()
	rayJobObj.RayCluster.WorkerGroupSpec[0].K8SPod = customAffinityPod
	taskTemplate := dummyRayTaskTemplate("ray-id", rayJobObj)
	taskContext := dummyRayTaskContext(taskTemplate, cpuOnly, nil, "", serviceAccount)

	r, err := rayJobResourceHandler{}.BuildResource(context.TODO(), taskContext)
	require.NoError(t, err)
	rayJob, ok := r.(*rayv1.RayJob)
	require.True(t, ok)

	workerSpec := rayJob.Spec.RayClusterSpec.WorkerGroupSpecs[0].Template.Spec
	terms := workerSpec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms
	require.GreaterOrEqual(t, len(terms), 2, "expected the custom affinity term to be appended as a separate term")
	// If this fails, DefaultAffinity is diluted (bug). If it passes, no bug.
	assertEveryTermHasRequirement(t, workerSpec, defaultAffinityReq)
}

func TestBuildResourceRay_DisablesLogNoiseEnv(t *testing.T) {
	rayJobResourceHandler := rayJobResourceHandler{}
	taskTemplate := dummyRayTaskTemplate("ray-id", dummyRayCustomObj())
	rayCtx := dummyRayTaskContext(taskTemplate, resourceRequirements, nil, "", serviceAccount)

	rayResource, err := rayJobResourceHandler.BuildResource(context.TODO(), rayCtx)
	require.NoError(t, err)
	ray, ok := rayResource.(*rayv1.RayJob)
	require.True(t, ok)

	// envOf returns the named container's env vars as name->value. The log-noise vars must land
	// on the Ray container itself, not an injected sidecar, so select by container name rather
	// than scanning every container in the pod.
	envOf := func(containers []corev1.Container, containerName string) map[string]string {
		for _, c := range containers {
			if c.Name == containerName {
				env := make(map[string]string, len(c.Env))
				for _, e := range c.Env {
					env[e.Name] = e.Value
				}
				return env
			}
		}
		return nil
	}

	require.NotEmpty(t, ray.Spec.RayClusterSpec.WorkerGroupSpecs, "expected at least one worker group")
	headEnv := envOf(ray.Spec.RayClusterSpec.HeadGroupSpec.Template.Spec.Containers, RayHeadContainerName)
	workerEnv := envOf(ray.Spec.RayClusterSpec.WorkerGroupSpecs[0].Template.Spec.Containers, "ray-worker")
	require.NotNil(t, headEnv, "head pod is missing the %s container", RayHeadContainerName)
	require.NotNil(t, workerEnv, "worker pod is missing the ray-worker container")

	for _, env := range []struct{ name, value string }{
		{"RAY_COLOR_PREFIX", "0"},
		{"RAY_DATA_DISABLE_PROGRESS_BARS", "1"},
	} {
		assert.Equal(t, env.value, headEnv[env.name], "head container must set %s", env.name)
		assert.Equal(t, env.value, workerEnv[env.name], "worker container must set %s", env.name)
	}
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

func TestBuildResourceRayGroupExtendedResources(t *testing.T) {
	assert.NoError(t, config.SetK8sPluginConfig(&config.K8sPluginConfig{
		GpuDeviceNodeLabel:                 "gpu-node-label",
		GpuPartitionSizeNodeLabel:          "gpu-partition-size",
		GpuResourceName:                    flytek8s.ResourceNvidiaGPU,
		AddTolerationsForExtendedResources: []string{"nvidia.com/gpu"},
	}))

	resources := &corev1.ResourceRequirements{
		Limits: corev1.ResourceList{
			flytek8s.ResourceNvidiaGPU: resource.MustParse("1"),
		},
	}

	// The head and worker groups request different extended resources via their own
	// extended_resources fields. No task-level extended resources are set, so each
	// group's pod spec should reflect only its own configuration.
	rayJobObj := dummyRayCustomObj()
	rayJobObj.RayCluster.HeadGroupSpec.ExtendedResources = &core.ExtendedResources{
		GpuAccelerator: &core.GPUAccelerator{
			Device: "nvidia-tesla-t4",
		},
	}
	rayJobObj.RayCluster.WorkerGroupSpec[0].ExtendedResources = &core.ExtendedResources{
		GpuAccelerator: &core.GPUAccelerator{
			Device: "nvidia-tesla-a100",
		},
		SharedMemory: &core.SharedMemory{
			MountName: "dshm",
			MountPath: "/dev/shm",
			SizeLimit: "1Gi",
		},
	}

	taskTemplate := dummyRayTaskTemplate("ray-id", rayJobObj)
	taskContext := dummyRayTaskContext(taskTemplate, resources, nil, "", serviceAccount)
	rayJobResourceHandler := rayJobResourceHandler{}
	r, err := rayJobResourceHandler.BuildResource(context.TODO(), taskContext)
	assert.NoError(t, err)
	assert.NotNil(t, r)
	rayJob, ok := r.(*rayv1.RayJob)
	assert.True(t, ok)

	gpuNodeSelectorTerms := func(device string) []corev1.NodeSelectorTerm {
		return []corev1.NodeSelectorTerm{
			{
				MatchExpressions: []corev1.NodeSelectorRequirement{
					{
						Key:      "gpu-node-label",
						Operator: corev1.NodeSelectorOpIn,
						Values:   []string{device},
					},
				},
			},
		}
	}
	// Each group's pod is built via ToK8sPodSpec, which applies the GPU device
	// toleration before the nvidia.com/gpu extended-resource toleration.
	gpuTolerations := func(device string) []corev1.Toleration {
		return []corev1.Toleration{
			{
				Key:      "gpu-node-label",
				Value:    device,
				Operator: corev1.TolerationOpEqual,
				Effect:   corev1.TaintEffectNoSchedule,
			},
			{
				Key:      "nvidia.com/gpu",
				Operator: corev1.TolerationOpExists,
				Effect:   corev1.TaintEffectNoSchedule,
			},
		}
	}

	// Head node uses the head group's extended resources (t4, no shared memory).
	headNodeSpec := rayJob.Spec.RayClusterSpec.HeadGroupSpec.Template.Spec
	assert.EqualValues(t, gpuNodeSelectorTerms("nvidia-tesla-t4"),
		headNodeSpec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms)
	assert.EqualValues(t, gpuTolerations("nvidia-tesla-t4"), headNodeSpec.Tolerations)
	assert.False(t, hasVolume(headNodeSpec.Volumes, "dshm"), "head pod should not have the worker's shared memory volume")

	// Worker node uses the worker group's extended resources (a100 + shared memory).
	workerNodeSpec := rayJob.Spec.RayClusterSpec.WorkerGroupSpecs[0].Template.Spec
	assert.EqualValues(t, gpuNodeSelectorTerms("nvidia-tesla-a100"),
		workerNodeSpec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms)
	assert.EqualValues(t, gpuTolerations("nvidia-tesla-a100"), workerNodeSpec.Tolerations)
	assert.True(t, hasVolume(workerNodeSpec.Volumes, "dshm"), "worker pod should have the shared memory volume")

	var workerContainer *corev1.Container
	for i := range workerNodeSpec.Containers {
		if workerNodeSpec.Containers[i].Name == "ray-worker" {
			workerContainer = &workerNodeSpec.Containers[i]
		}
	}
	assert.NotNil(t, workerContainer)
	hasSharedMemoryMount := false
	for _, vm := range workerContainer.VolumeMounts {
		if vm.Name == "dshm" && vm.MountPath == "/dev/shm" {
			hasSharedMemoryMount = true
		}
	}
	assert.True(t, hasSharedMemoryMount, "worker container should mount the shared memory volume")
}

func hasVolume(volumes []corev1.Volume, name string) bool {
	for _, v := range volumes {
		if v.Name == name {
			return true
		}
	}
	return false
}

// TestBuildResourceRayGroupExtendedResourcesOverride verifies that when both task-level
// and group-level extended resources are set, the group-level GPU accelerator overrides
// the task-level one (rather than producing two conflicting node selector requirements
// for the same GPU label, which would make the pod unschedulable). Groups that do not
// set their own extended_resources fall back to the task-level configuration.
func TestBuildResourceRayGroupExtendedResourcesOverride(t *testing.T) {
	assert.NoError(t, config.SetK8sPluginConfig(&config.K8sPluginConfig{
		GpuDeviceNodeLabel:                 "gpu-node-label",
		GpuResourceName:                    flytek8s.ResourceNvidiaGPU,
		AddTolerationsForExtendedResources: []string{"nvidia.com/gpu"},
	}))

	resources := &corev1.ResourceRequirements{
		Limits: corev1.ResourceList{
			flytek8s.ResourceNvidiaGPU: resource.MustParse("1"),
		},
	}

	// Only the worker group overrides the task-level GPU device; the head group inherits
	// the task-level device.
	rayJobObj := dummyRayCustomObj()
	rayJobObj.RayCluster.WorkerGroupSpec[0].ExtendedResources = &core.ExtendedResources{
		GpuAccelerator: &core.GPUAccelerator{
			Device: "nvidia-tesla-a100",
		},
	}

	taskTemplate := dummyRayTaskTemplate("ray-id", rayJobObj)
	// Task-level extended resources request a different (default) GPU device.
	taskTemplate.ExtendedResources = &core.ExtendedResources{
		GpuAccelerator: &core.GPUAccelerator{
			Device: "nvidia-tesla-t4",
		},
	}
	taskContext := dummyRayTaskContext(taskTemplate, resources, nil, "", serviceAccount)
	rayJobResourceHandler := rayJobResourceHandler{}
	r, err := rayJobResourceHandler.BuildResource(context.TODO(), taskContext)
	assert.NoError(t, err)
	assert.NotNil(t, r)
	rayJob, ok := r.(*rayv1.RayJob)
	assert.True(t, ok)

	gpuNodeSelectorTerms := func(device string) []corev1.NodeSelectorTerm {
		return []corev1.NodeSelectorTerm{
			{
				MatchExpressions: []corev1.NodeSelectorRequirement{
					{
						Key:      "gpu-node-label",
						Operator: corev1.NodeSelectorOpIn,
						Values:   []string{device},
					},
				},
			},
		}
	}

	// Head node inherits the task-level GPU device (t4) since it sets no extended resources.
	headNodeSpec := rayJob.Spec.RayClusterSpec.HeadGroupSpec.Template.Spec
	assert.EqualValues(t, gpuNodeSelectorTerms("nvidia-tesla-t4"),
		headNodeSpec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms)

	// Worker node's group-level GPU device (a100) cleanly overrides the task-level device
	// (t4); there is a single node selector requirement, not two conflicting ones.
	workerNodeSpec := rayJob.Spec.RayClusterSpec.WorkerGroupSpecs[0].Template.Spec
	assert.EqualValues(t, gpuNodeSelectorTerms("nvidia-tesla-a100"),
		workerNodeSpec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms)
}

// TestBuildResourceRayGroupK8SPodGPU is a regression test for the case where the GPU
// resource is requested only through a group's custom k8s_pod (not the task container).
// ToK8sPodSpec applies GPU node selectors before the custom pod spec is merged, so the
// builders must re-apply them post-merge or the device toleration and node affinity are
// silently dropped (while the generic nvidia.com/gpu toleration, which is computed
// post-merge, still appears).
func TestBuildResourceRayGroupK8SPodGPU(t *testing.T) {
	assert.NoError(t, config.SetK8sPluginConfig(&config.K8sPluginConfig{
		GpuDeviceNodeLabel:                 "gpu-node-label",
		GpuResourceName:                    flytek8s.ResourceNvidiaGPU,
		AddTolerationsForExtendedResources: []string{"nvidia.com/gpu"},
	}))

	// Task-level resources do not include any GPU.
	cpuOnlyResources := &corev1.ResourceRequirements{
		Limits: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse("1000m"),
			corev1.ResourceMemory: resource.MustParse("1Gi"),
		},
		Requests: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse("100m"),
			corev1.ResourceMemory: resource.MustParse("512Mi"),
		},
	}

	gpuResources := corev1.ResourceRequirements{
		Limits: corev1.ResourceList{
			flytek8s.ResourceNvidiaGPU: resource.MustParse("1"),
		},
		Requests: corev1.ResourceList{
			flytek8s.ResourceNvidiaGPU: resource.MustParse("1"),
		},
	}
	workerGPUPod := &core.K8SPod{
		PodSpec: transformStructToStructPB(t, &corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:      "ray-worker",
					Resources: gpuResources,
				},
			},
		}),
	}

	expectedNodeSelectorTerms := []corev1.NodeSelectorTerm{
		{
			MatchExpressions: []corev1.NodeSelectorRequirement{
				{
					Key:      "gpu-node-label",
					Operator: corev1.NodeSelectorOpIn,
					Values:   []string{"nvidia-tesla-t4"},
				},
			},
		},
	}
	expectedTolerations := []corev1.Toleration{
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
	}

	params := []struct {
		name                   string
		taskExtendedResources  *core.ExtendedResources
		groupExtendedResources *core.ExtendedResources
	}{
		{
			name: "accelerator from group extended resources",
			groupExtendedResources: &core.ExtendedResources{
				GpuAccelerator: &core.GPUAccelerator{Device: "nvidia-tesla-t4"},
			},
		},
		{
			name: "accelerator from task extended resources",
			taskExtendedResources: &core.ExtendedResources{
				GpuAccelerator: &core.GPUAccelerator{Device: "nvidia-tesla-t4"},
			},
		},
	}

	for _, p := range params {
		t.Run(p.name, func(t *testing.T) {
			rayJobObj := dummyRayCustomObj()
			rayJobObj.RayCluster.WorkerGroupSpec[0].K8SPod = workerGPUPod
			rayJobObj.RayCluster.WorkerGroupSpec[0].ExtendedResources = p.groupExtendedResources

			taskTemplate := dummyRayTaskTemplate("ray-id", rayJobObj)
			taskTemplate.ExtendedResources = p.taskExtendedResources
			taskContext := dummyRayTaskContext(taskTemplate, cpuOnlyResources, nil, "", serviceAccount)
			r, err := rayJobResourceHandler{}.BuildResource(context.TODO(), taskContext)
			assert.NoError(t, err)
			rayJob, ok := r.(*rayv1.RayJob)
			assert.True(t, ok)

			// The worker pod's GPU comes only from the group's k8s_pod, so the device
			// node affinity and toleration must be applied after that merge.
			workerNodeSpec := rayJob.Spec.RayClusterSpec.WorkerGroupSpecs[0].Template.Spec
			assert.NotNil(t, workerNodeSpec.Affinity)
			assert.NotNil(t, workerNodeSpec.Affinity.NodeAffinity)
			assert.EqualValues(t, expectedNodeSelectorTerms,
				workerNodeSpec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms)
			assert.EqualValues(t, expectedTolerations, workerNodeSpec.Tolerations)

			// The head pod requests no GPU, so it must not pick up GPU scheduling
			// constraints. (ToK8sPodSpec always initializes an empty Affinity via
			// ApplyInterruptibleNodeAffinity, so only NodeAffinity is asserted.)
			headNodeSpec := rayJob.Spec.RayClusterSpec.HeadGroupSpec.Template.Spec
			assert.Nil(t, headNodeSpec.Affinity.NodeAffinity)
			assert.Empty(t, headNodeSpec.Tolerations)
		})
	}
}

func TestBuildResourceRayCustomK8SPod(t *testing.T) {
	assert.NoError(t, config.SetK8sPluginConfig(&config.K8sPluginConfig{}))

	headResourceEntries := []*core.Resources_ResourceEntry{
		{Name: core.Resources_CPU, Value: "10"},
		{Name: core.Resources_MEMORY, Value: "10Gi"},
		{Name: core.Resources_GPU, Value: "10"},
	}
	headResources := &core.Resources{Requests: headResourceEntries, Limits: headResourceEntries}

	expectedHeadResources, err := flytek8s.ToK8sResourceRequirements(headResources)
	require.NoError(t, err)
	// Add nvidia.com/gpu from task resources since mergeCustomPodSpec only replaces resources
	expectedHeadResources.Limits[flytek8s.ResourceNvidiaGPU] = resource.MustParse("1")
	expectedHeadResources.Requests[flytek8s.ResourceNvidiaGPU] = resource.MustParse("1")

	workerResourceEntries := []*core.Resources_ResourceEntry{
		{Name: core.Resources_CPU, Value: "20"},
		{Name: core.Resources_MEMORY, Value: "20Gi"},
		{Name: core.Resources_GPU, Value: "20"},
	}
	workerResources := &core.Resources{Requests: workerResourceEntries, Limits: workerResourceEntries}

	expectedWorkerResources, err := flytek8s.ToK8sResourceRequirements(workerResources)
	require.NoError(t, err)
	// Add nvidia.com/gpu from task resources since mergeCustomPodSpec only replaces resources
	expectedWorkerResources.Limits[flytek8s.ResourceNvidiaGPU] = resource.MustParse("1")
	expectedWorkerResources.Requests[flytek8s.ResourceNvidiaGPU] = resource.MustParse("1")

	nvidiaRuntimeClassName := "nvidia-cdi"

	headPodSpecCustomResources := &corev1.PodSpec{
		Containers: []corev1.Container{
			{
				Name:      "ray-head",
				Resources: *expectedHeadResources,
			},
		},
	}
	workerPodSpecCustomResources := &corev1.PodSpec{
		Containers: []corev1.Container{
			{
				Name:      "ray-worker",
				Resources: *expectedWorkerResources,
			},
		},
	}

	headPodSpecCustomRuntimeClass := &corev1.PodSpec{
		RuntimeClassName: &nvidiaRuntimeClassName,
	}
	workerPodSpecCustomRuntimeClass := &corev1.PodSpec{
		RuntimeClassName: &nvidiaRuntimeClassName,
	}

	params := []struct {
		name                              string
		taskResources                     *corev1.ResourceRequirements
		headK8SPod                        *core.K8SPod
		workerK8SPod                      *core.K8SPod
		expectedSubmitterResources        *corev1.ResourceRequirements
		expectedHeadResources             *corev1.ResourceRequirements
		expectedWorkerResources           *corev1.ResourceRequirements
		expectedSubmitterRuntimeClassName *string
		expectedHeadRuntimeClassName      *string
		expectedWorkerRuntimeClassName    *string
	}{
		{
			name:                       "task resources",
			taskResources:              resourceRequirements,
			expectedSubmitterResources: &submitterDefaultResourceRequirements,
			expectedHeadResources:      resourceRequirements,
			expectedWorkerResources:    resourceRequirements,
		},
		{
			name:          "custom worker and head resources",
			taskResources: resourceRequirements,
			headK8SPod: &core.K8SPod{
				PodSpec: transformStructToStructPB(t, headPodSpecCustomResources),
			},
			workerK8SPod: &core.K8SPod{
				PodSpec: transformStructToStructPB(t, workerPodSpecCustomResources),
			},
			expectedSubmitterResources: &submitterDefaultResourceRequirements,
			expectedHeadResources:      expectedHeadResources,
			expectedWorkerResources:    expectedWorkerResources,
		},
		{
			name:                       "custom runtime class name",
			taskResources:              resourceRequirements,
			expectedSubmitterResources: &submitterDefaultResourceRequirements,
			expectedHeadResources:      resourceRequirements,
			expectedWorkerResources:    resourceRequirements,
			headK8SPod: &core.K8SPod{
				PodSpec: transformStructToStructPB(t, headPodSpecCustomRuntimeClass),
			},
			workerK8SPod: &core.K8SPod{
				PodSpec: transformStructToStructPB(t, workerPodSpecCustomRuntimeClass),
			},
			expectedHeadRuntimeClassName:   &nvidiaRuntimeClassName,
			expectedWorkerRuntimeClassName: &nvidiaRuntimeClassName,
		},
	}

	for _, p := range params {
		t.Run(p.name, func(t *testing.T) {
			rayJobInput := dummyRayCustomObj()

			if p.headK8SPod != nil {
				rayJobInput.RayCluster.HeadGroupSpec.K8SPod = p.headK8SPod
			}

			if p.workerK8SPod != nil {
				for _, spec := range rayJobInput.RayCluster.WorkerGroupSpec {
					spec.K8SPod = p.workerK8SPod
				}
			}

			taskTemplate := dummyRayTaskTemplate("ray-id", rayJobInput)
			taskContext := dummyRayTaskContext(taskTemplate, p.taskResources, nil, "", serviceAccount)
			rayJobResourceHandler := rayJobResourceHandler{}
			r, err := rayJobResourceHandler.BuildResource(context.TODO(), taskContext)
			assert.Nil(t, err)
			assert.NotNil(t, r)
			rayJob, ok := r.(*rayv1.RayJob)
			assert.True(t, ok)

			submitterPodResources := rayJob.Spec.SubmitterPodTemplate.Spec.Containers[0].Resources
			assert.EqualValues(t,
				p.expectedSubmitterResources,
				&submitterPodResources,
			)

			headPodSpec := rayJob.Spec.RayClusterSpec.HeadGroupSpec.Template.Spec
			headPodResources := headPodSpec.Containers[0].Resources
			assert.EqualValues(t,
				p.expectedHeadResources,
				&headPodResources,
			)

			assert.EqualValues(t, p.expectedHeadRuntimeClassName, headPodSpec.RuntimeClassName)

			for _, workerGroupSpec := range rayJob.Spec.RayClusterSpec.WorkerGroupSpecs {
				workerPodSpec := workerGroupSpec.Template.Spec
				workerPodResources := workerPodSpec.Containers[0].Resources
				assert.EqualValues(t,
					p.expectedWorkerResources,
					&workerPodResources,
				)
				assert.EqualValues(t, p.expectedWorkerRuntimeClassName, workerPodSpec.RuntimeClassName)
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
		taskTemplate *core.TaskTemplate
		// primaryContainerName string
		logsSidecarCfg                       *corev1.Container
		expectedVolumes                      []corev1.Volume
		expectedPrimaryContainerVolumeMounts []corev1.VolumeMount
		expectedLogsSidecarVolumeMounts      []corev1.VolumeMount
	}{
		{
			"container target",
			&core.TaskTemplate{
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
			&core.TaskTemplate{
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
			&core.TaskTemplate{
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
			&core.TaskTemplate{
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

	for i := range params {
		p := params[i]
		t.Run(p.name, func(t *testing.T) {
			assert.NoError(t, SetConfig(&Config{
				LogsSidecar: p.logsSidecarCfg,
			}))
			taskContext := dummyRayTaskContext(p.taskTemplate, resourceRequirements, nil, "", serviceAccount)
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
				if cnt.Name == RayHeadContainerName {
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

func newPluginContext(pluginState k8s.PluginState) *k8smocks.PluginContext {
	plg := &k8smocks.PluginContext{}

	taskExecID := &mocks.TaskExecutionID{}
	taskExecID.EXPECT().GetID().Return(&core.TaskExecutionIdentifier{
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
	taskExecID.EXPECT().GetUniqueNodeID().Return("unique-node")
	taskExecID.EXPECT().GetGeneratedName().Return("generated-name")

	tskCtx := &mocks.TaskExecutionMetadata{}
	tskCtx.EXPECT().GetTaskExecutionID().Return(taskExecID)
	plg.EXPECT().TaskExecutionMetadata().Return(tskCtx)

	pluginStateReaderMock := mocks.PluginStateReader{}
	pluginStateReaderMock.EXPECT().Get(mock.AnythingOfType(reflect.TypeOf(&pluginState).String())).RunAndReturn(
		func(v interface{}) (uint8, error) {
			*(v.(*k8s.PluginState)) = pluginState
			return 0, nil
		})

	plg.EXPECT().PluginStateReader().Return(&pluginStateReaderMock)

	taskReader := &mocks.TaskReader{}
	taskReader.EXPECT().Read(mock.Anything).Return(&core.TaskTemplate{}, nil)
	plg.EXPECT().TaskReader().Return(taskReader)

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
	pluginCtx := rayPluginContext(k8s.PluginState{})

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

	startTime := time.Date(2024, 0, 0, 0, 0, 0, 0, time.UTC)
	endTime := startTime.Add(time.Hour)
	podName, contName, initCont := "ray-clust-ray-head", "ray-head", "init"
	logCtx := &core.LogContext{
		PrimaryPodName: podName,
		Pods: []*core.PodLogContext{
			{
				Namespace:            "ns",
				PodName:              podName,
				PrimaryContainerName: contName,
				Containers: []*core.ContainerContext{
					{
						ContainerName: contName,
						Process: &core.ContainerContext_ProcessContext{
							ContainerStartTime: timestamppb.New(startTime),
							ContainerEndTime:   timestamppb.New(endTime),
						},
					},
				},
				InitContainers: []*core.ContainerContext{
					{
						ContainerName: initCont,
						Process: &core.ContainerContext_ProcessContext{
							ContainerStartTime: timestamppb.New(startTime),
							ContainerEndTime:   timestamppb.New(endTime),
						},
					},
				},
			},
		},
	}
	for _, tc := range testCases {
		t.Run("TestGetTaskPhase_"+string(tc.rayJobPhase), func(t *testing.T) {
			startTime := metav1.NewTime(time.Now())
			rayObject := &rayv1.RayJob{
				Spec: rayv1.RayJobSpec{
					RayClusterSpec: &rayv1.RayClusterSpec{
						HeadGroupSpec: rayv1.HeadGroupSpec{
							Template: corev1.PodTemplateSpec{
								Spec: corev1.PodSpec{
									Containers: []corev1.Container{
										{
											Name:  "ray-head",
											Image: "rayproject/ray:latest",
											Env:   []corev1.EnvVar{},
										},
									},
								},
							},
						},
					},
				},
				Status: rayv1.RayJobStatus{
					JobDeploymentStatus: tc.rayJobPhase,
					RayClusterName:      "ray-clust",
					StartTime:           &startTime,
				},
			}
			phaseInfo, err := rayJobResourceHandler.GetTaskPhase(ctx, pluginCtx, rayObject)
			if tc.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedCorePhase.String(), phaseInfo.Phase().String())
				assert.Equal(t, logCtx, phaseInfo.Info().LogContext)
			}
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
	pluginCtx := rayPluginContext(pluginState)

	rayObject := &rayv1.RayJob{}
	rayObject.Status.JobDeploymentStatus = rayv1.JobDeploymentStatusInitializing
	phaseInfo, err := rayJobResourceHandler.GetTaskPhase(ctx, pluginCtx, rayObject)

	assert.NoError(t, err)
	assert.Equal(t, phaseInfo.Version(), pluginsCore.DefaultPhaseVersion+1)
}

func TestGetEventInfo_LogTemplates(t *testing.T) {
	pluginCtx := rayPluginContext(k8s.PluginState{})
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
					Name:  "namespace",
					Uri:   "http://test/test-namespace",
					Ready: true,
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
					Name:  "taskExecID",
					Uri:   "http://test/projects/my-execution-project/domains/my-execution-domain/executions/my-execution-name/nodeId/unique-node/taskId/my-task-name/attempt/1",
					Ready: true,
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
					Name:  "ray cluster name",
					Uri:   "http://test/test-namespace/ray-cluster",
					Ready: true,
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
					Name:  "ray job ID",
					Uri:   "http://test/test-namespace/ray-job-1",
					Ready: true,
				},
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.name, func(t *testing.T) {
			ti, err := getEventInfoForRayJob(
				context.TODO(),
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
	pluginCtx := rayPluginContext(k8s.PluginState{})
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
					Name:  "namespace",
					Uri:   "http://test/test-namespace",
					Ready: true,
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
					Name:  "taskExecID",
					Uri:   "http://test/projects/my-execution-project/domains/my-execution-domain/executions/my-execution-name/nodeId/unique-node/taskId/my-task-name/attempt/1",
					Ready: true,
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
					Name:  "ray cluster name",
					Uri:   "http://test/test-namespace/ray-cluster",
					Ready: true,
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
					Name:  "ray job ID",
					Uri:   "http://test/test-namespace/ray-job-1",
					Ready: true,
				},
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.name, func(t *testing.T) {
			ti, err := getEventInfoForRayJob(
				context.TODO(),
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
	pluginCtx := rayPluginContext(k8s.PluginState{})
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
					Name:     "Ray Dashboard",
					Uri:      "http://test/generated-name",
					LinkType: core.TaskLog_DASHBOARD,
					Ready:    true,
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

	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.name, func(t *testing.T) {
			assert.NoError(t, SetConfig(&Config{DashboardURLTemplate: &tc.dashboardURLTemplate}))
			ti, err := getEventInfoForRayJob(context.TODO(), logs.LogConfig{}, pluginCtx, &tc.rayJob)
			assert.NoError(t, err)
			assert.Equal(t, tc.expectedTaskLogs, ti.Logs)
		})
	}
}

func TestGetEventInfo_DashboardURL_V1(t *testing.T) {
	pluginCtx := rayPluginContext(k8s.PluginState{})
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
					Name:     "Ray Dashboard",
					Uri:      "http://test/generated-name",
					LinkType: core.TaskLog_DASHBOARD,
					Ready:    true,
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

	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.name, func(t *testing.T) {
			assert.NoError(t, SetConfig(&Config{DashboardURLTemplate: &tc.dashboardURLTemplate}))
			ti, err := getEventInfoForRayJob(context.TODO(), logs.LogConfig{}, pluginCtx, &tc.rayJob)
			assert.NoError(t, err)
			assert.Equal(t, tc.expectedTaskLogs, ti.Logs)
		})
	}
}

func TestBuildResourceRaySubmitterPodAffinity(t *testing.T) {
	// Create a default affinity to be set in the K8s plugin config
	defaultAffinity := &corev1.Affinity{
		NodeAffinity: &corev1.NodeAffinity{
			RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
				NodeSelectorTerms: []corev1.NodeSelectorTerm{
					{
						MatchExpressions: []corev1.NodeSelectorRequirement{
							{
								Key:      "node-type",
								Operator: corev1.NodeSelectorOpIn,
								Values:   []string{"ray-submitter"},
							},
						},
					},
				},
			},
		},
	}

	assert.NoError(t, config.SetK8sPluginConfig(&config.K8sPluginConfig{
		DefaultAffinity: defaultAffinity,
	}))

	taskTemplate := dummyRayTaskTemplate("ray-id", dummyRayCustomObj())
	taskContext := dummyRayTaskContext(taskTemplate, resourceRequirements, nil, "", serviceAccount)
	rayJobResourceHandler := rayJobResourceHandler{}
	r, err := rayJobResourceHandler.BuildResource(context.TODO(), taskContext)
	assert.Nil(t, err)
	assert.NotNil(t, r)
	rayJob, ok := r.(*rayv1.RayJob)
	assert.True(t, ok)

	// Verify the submitter pod template has the default affinity from K8s plugin config
	submitterPodAffinity := rayJob.Spec.SubmitterPodTemplate.Spec.Affinity
	assert.NotNil(t, submitterPodAffinity, "submitter pod should have affinity set")
	assert.EqualValues(t, defaultAffinity, submitterPodAffinity, "submitter pod affinity should match default affinity from config")
}

func TestGetPropertiesRay(t *testing.T) {
	rayJobResourceHandler := rayJobResourceHandler{}
	maxLength := 47
	expected := k8s.PluginProperties{GeneratedNameMaxLength: &maxLength}
	assert.Equal(t, expected, rayJobResourceHandler.GetProperties())
}

func rayPluginContext(pluginState k8s.PluginState) *k8smocks.PluginContext {
	pluginCtx := newPluginContext(pluginState)
	startTime := time.Date(2024, 0, 0, 0, 0, 0, 0, time.UTC)
	endTime := startTime.Add(time.Hour)
	podName, contName, initCont := "ray-clust-ray-head", "ray-head", "init"
	podList := []runtime.Object{
		&corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{Namespace: "ns", Name: "initializing ignored pod"},
			Status:     corev1.PodStatus{Phase: corev1.PodPending},
		},
		&corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{Namespace: "ns", Name: podName},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{Name: contName},
				},
			},
			Status: corev1.PodStatus{
				Phase: corev1.PodRunning,
				ContainerStatuses: []corev1.ContainerStatus{
					{
						Name: contName,
						State: corev1.ContainerState{
							Terminated: &corev1.ContainerStateTerminated{
								StartedAt:  metav1.Time{Time: startTime},
								FinishedAt: metav1.Time{Time: endTime},
							},
						},
					},
				},
				InitContainerStatuses: []corev1.ContainerStatus{
					{
						Name: initCont,
						State: corev1.ContainerState{
							Terminated: &corev1.ContainerStateTerminated{
								StartedAt:  metav1.Time{Time: startTime},
								FinishedAt: metav1.Time{Time: endTime},
							},
						},
					},
				},
			},
		},
	}
	reader := fake.NewFakeClient(podList...)
	pluginCtx.EXPECT().K8sReader().Return(reader)
	return pluginCtx
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

func rayPluginContextWithPods(pluginState k8s.PluginState, pods ...runtime.Object) *k8smocks.PluginContext {
	pluginCtx := newPluginContext(pluginState)
	reader := fake.NewFakeClient(pods...)
	pluginCtx.EXPECT().K8sReader().Return(reader)
	return pluginCtx
}

func TestGetTaskPhaseWithFailedPod(t *testing.T) {
	rayJobResourceHandler := rayJobResourceHandler{}
	ctx := context.Background()

	rayJobName := "test-rayjob"
	rayClusterName := "test-raycluster"
	// Create a failed head pod - name must match pattern used in GetTaskPhase: {RayClusterName}-head
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      rayClusterName + "-head",
			Namespace: "ns",
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name: RayHeadContainerName,
				},
			},
		},
		Status: corev1.PodStatus{
			Phase: corev1.PodFailed,
			ContainerStatuses: []corev1.ContainerStatus{
				{
					Name: RayHeadContainerName,
					State: corev1.ContainerState{
						Terminated: &corev1.ContainerStateTerminated{
							ExitCode: 1,
							Reason:   "Error",
							Message:  "Container failed",
						},
					},
				},
			},
		},
	}

	pluginCtx := rayPluginContextWithPods(k8s.PluginState{}, pod)

	startTime := metav1.NewTime(time.Now())
	rayObject := &rayv1.RayJob{
		ObjectMeta: metav1.ObjectMeta{
			Name:      rayJobName,
			Namespace: "ns",
		},
		Spec: rayv1.RayJobSpec{
			RayClusterSpec: &rayv1.RayClusterSpec{
				HeadGroupSpec: rayv1.HeadGroupSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Name:  RayHeadContainerName,
									Image: "rayproject/ray:latest",
									Env:   []corev1.EnvVar{},
								},
							},
						},
					},
				},
			},
		},
		Status: rayv1.RayJobStatus{
			JobDeploymentStatus: rayv1.JobDeploymentStatusRunning,
			RayClusterName:      rayClusterName,
			StartTime:           &startTime,
		},
	}

	// Even though RayJob status is running, should return failure due to pod status
	phaseInfo, err := rayJobResourceHandler.GetTaskPhase(ctx, pluginCtx, rayObject)
	assert.NoError(t, err)
	assert.True(t, phaseInfo.Phase().IsFailure())
}

func TestGetTaskPhaseContainerNameConstant(t *testing.T) {
	rayJobResourceHandler := rayJobResourceHandler{}
	ctx := context.Background()
	pluginCtx := rayPluginContext(k8s.PluginState{})

	startTime := metav1.NewTime(time.Now())
	rayObject := &rayv1.RayJob{
		Spec: rayv1.RayJobSpec{
			RayClusterSpec: &rayv1.RayClusterSpec{
				HeadGroupSpec: rayv1.HeadGroupSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Name:  RayHeadContainerName,
									Image: "rayproject/ray:latest",
									Env:   []corev1.EnvVar{},
								},
							},
						},
					},
				},
			},
		},
		Status: rayv1.RayJobStatus{
			JobDeploymentStatus: rayv1.JobDeploymentStatusComplete,
			RayClusterName:      "ray-clust",
			StartTime:           &startTime,
		},
	}

	phaseInfo, err := rayJobResourceHandler.GetTaskPhase(ctx, pluginCtx, rayObject)
	assert.NoError(t, err)
	assert.NotNil(t, phaseInfo.Info())
	assert.NotNil(t, phaseInfo.Info().LogContext)

	// Verify the constant is used for head container names
	assert.Equal(t, 1, len(phaseInfo.Info().LogContext.Pods))
	headPodLogContext := phaseInfo.Info().LogContext.Pods[0]
	assert.Equal(t, RayHeadContainerName, headPodLogContext.PrimaryContainerName)
	assert.Equal(t, 1, len(headPodLogContext.Containers))
	assert.Equal(t, RayHeadContainerName, headPodLogContext.Containers[0].ContainerName)
}

func TestIsTerminal(t *testing.T) {
	rayJobResourceHandler := rayJobResourceHandler{}
	ctx := context.Background()

	tests := []struct {
		name           string
		status         rayv1.JobDeploymentStatus
		expectedResult bool
	}{
		{"Complete", rayv1.JobDeploymentStatusComplete, true},
		{"Failed", rayv1.JobDeploymentStatusFailed, true},
		{"Running", rayv1.JobDeploymentStatusRunning, false},
		{"Initializing", rayv1.JobDeploymentStatusInitializing, false},
		{"Suspended", rayv1.JobDeploymentStatusSuspended, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			job := &rayv1.RayJob{
				Status: rayv1.RayJobStatus{
					JobDeploymentStatus: tt.status,
				},
			}
			result, err := rayJobResourceHandler.IsTerminal(ctx, job)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedResult, result)
		})
	}
}

func TestIsTerminal_WrongResourceType(t *testing.T) {
	rayJobResourceHandler := rayJobResourceHandler{}
	ctx := context.Background()

	wrongResource := &corev1.ConfigMap{}
	result, err := rayJobResourceHandler.IsTerminal(ctx, wrongResource)
	assert.Error(t, err)
	assert.False(t, result)
	assert.Contains(t, err.Error(), "unexpected resource type")
}

func TestGetCompletionTime(t *testing.T) {
	rayJobResourceHandler := rayJobResourceHandler{}

	now := time.Now().Truncate(time.Second)
	earlier := now.Add(-1 * time.Hour)
	evenEarlier := now.Add(-2 * time.Hour)

	tests := []struct {
		name         string
		job          *rayv1.RayJob
		expectedTime time.Time
	}{
		{
			name: "uses EndTime",
			job: &rayv1.RayJob{
				ObjectMeta: metav1.ObjectMeta{
					CreationTimestamp: metav1.NewTime(evenEarlier),
				},
				Status: rayv1.RayJobStatus{
					EndTime:   &metav1.Time{Time: now},
					StartTime: &metav1.Time{Time: earlier},
				},
			},
			expectedTime: now,
		},
		{
			name: "falls back to StartTime",
			job: &rayv1.RayJob{
				ObjectMeta: metav1.ObjectMeta{
					CreationTimestamp: metav1.NewTime(evenEarlier),
				},
				Status: rayv1.RayJobStatus{
					StartTime: &metav1.Time{Time: now},
				},
			},
			expectedTime: now,
		},
		{
			name: "falls back to CreationTimestamp",
			job: &rayv1.RayJob{
				ObjectMeta: metav1.ObjectMeta{
					CreationTimestamp: metav1.NewTime(now),
				},
				Status: rayv1.RayJobStatus{},
			},
			expectedTime: now,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := rayJobResourceHandler.GetCompletionTime(tt.job)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedTime.Unix(), result.Unix())
		})
	}
}

func TestGetCompletionTime_WrongResourceType(t *testing.T) {
	rayJobResourceHandler := rayJobResourceHandler{}

	wrongResource := &corev1.ConfigMap{}
	result, err := rayJobResourceHandler.GetCompletionTime(wrongResource)
	assert.Error(t, err)
	assert.True(t, result.IsZero())
	assert.Contains(t, err.Error(), "unexpected resource type")
}
