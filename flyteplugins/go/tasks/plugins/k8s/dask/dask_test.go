package dask

import (
	"context"
	"testing"
	"time"

	daskAPI "github.com/dask/dask-kubernetes/v2023/dask_kubernetes/operator/go_client/pkg/apis/kubernetes.dask.org/v1"
	"github.com/golang/protobuf/jsonpb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/protobuf/types/known/structpb"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/plugins"
	pluginsCore "github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/core/mocks"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/flytek8s"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/flytek8s/config"
	pluginIOMocks "github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/io/mocks"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/k8s"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/utils"
)

const (
	defaultTestImage             = "image://"
	testNWorkers                 = 10
	testTaskID                   = "some-acceptable-name"
	podTemplateName              = "dask-dummy-pod-template-name"
	defaultServiceAccountName    = "default-service-account"
	defaultNamespace             = "default-namespace"
	podTempaltePriorityClassName = "pod-template-priority-class-name"
)

var (
	testEnvVars = []v1.EnvVar{
		{Name: "Env_Var", Value: "Env_Val"},
	}
	testArgs = []string{
		"execute-dask-task",
	}
	testAnnotations       = map[string]string{"annotation-1": "val1"}
	testLabels            = map[string]string{"label-1": "val1"}
	testPlatformResources = v1.ResourceRequirements{
		Requests: v1.ResourceList{
			v1.ResourceCPU: resource.MustParse("4"),
		},
		Limits: v1.ResourceList{
			v1.ResourceCPU:    resource.MustParse("5"),
			v1.ResourceMemory: resource.MustParse("17G"),
		},
	}
	defaultResources = v1.ResourceRequirements{
		Requests: testPlatformResources.Requests,
		Limits:   testPlatformResources.Requests,
	}
	podTemplate = &v1.PodTemplate{
		ObjectMeta: metav1.ObjectMeta{
			Name: podTemplateName,
		},
		Template: v1.PodTemplateSpec{
			Spec: v1.PodSpec{
				PriorityClassName: podTempaltePriorityClassName,
			},
		},
	}
)

func dummyDaskJob(status daskAPI.JobStatus) *daskAPI.DaskJob {
	return &daskAPI.DaskJob{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "dask-job-name",
			Namespace: defaultNamespace,
		},
		Status: daskAPI.DaskJobStatus{
			ClusterName:      "dask-cluster-name",
			EndTime:          metav1.Time{Time: time.Now()},
			JobRunnerPodName: "job-runner-pod-name",
			JobStatus:        status,
			StartTime:        metav1.Time{Time: time.Now()},
		},
	}
}

func dummpyDaskCustomObj(customImage string, resources *core.Resources) *plugins.DaskJob {
	scheduler := plugins.DaskScheduler{
		Image:     customImage,
		Resources: resources,
	}

	workers := plugins.DaskWorkerGroup{
		NumberOfWorkers: 10,
		Image:           customImage,
		Resources:       resources,
	}

	daskJob := plugins.DaskJob{
		Scheduler: &scheduler,
		Workers:   &workers,
	}
	return &daskJob
}

func dummyDaskTaskTemplate(customImage string, resources *core.Resources, podTemplateName string) *core.TaskTemplate {
	// In a real usecase, resources will always be filled, but might be empty
	if resources == nil {
		resources = &core.Resources{
			Requests: []*core.Resources_ResourceEntry{},
			Limits:   []*core.Resources_ResourceEntry{},
		}
	}

	daskJob := dummpyDaskCustomObj(customImage, resources)
	daskJobJSON, err := utils.MarshalToString(daskJob)
	if err != nil {
		panic(err)
	}

	structObj := structpb.Struct{}
	err = jsonpb.UnmarshalString(daskJobJSON, &structObj)
	if err != nil {
		panic(err)
	}
	var envVars []*core.KeyValuePair
	for _, envVar := range testEnvVars {
		envVars = append(envVars, &core.KeyValuePair{Key: envVar.Name, Value: envVar.Value})
	}
	metadata := &core.TaskMetadata{
		PodTemplateName: podTemplateName,
	}
	return &core.TaskTemplate{
		Id:       &core.Identifier{Name: "test-build-resource"},
		Type:     daskTaskType,
		Metadata: metadata,
		Target: &core.TaskTemplate_Container{
			Container: &core.Container{
				Image: defaultTestImage,
				Args:  testArgs,
				Env:   envVars,
			},
		},
		Custom: &structObj,
	}
}

func dummyDaskTaskContext(taskTemplate *core.TaskTemplate, resources *v1.ResourceRequirements, extendedResources *core.ExtendedResources, isInterruptible bool) pluginsCore.TaskExecutionContext {
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
	taskCtx.On("OutputWriter").Return(outputReader)

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
	tID.On("GetGeneratedName").Return(testTaskID)
	tID.On("GetUniqueNodeID").Return("an-unique-id")

	taskExecutionMetadata := &mocks.TaskExecutionMetadata{}
	taskExecutionMetadata.OnGetTaskExecutionID().Return(tID)
	taskExecutionMetadata.OnGetAnnotations().Return(testAnnotations)
	taskExecutionMetadata.OnGetLabels().Return(testLabels)
	taskExecutionMetadata.OnGetPlatformResources().Return(&testPlatformResources)
	taskExecutionMetadata.OnGetMaxAttempts().Return(uint32(1))
	taskExecutionMetadata.OnIsInterruptible().Return(isInterruptible)
	taskExecutionMetadata.OnGetEnvironmentVariables().Return(nil)
	taskExecutionMetadata.OnGetK8sServiceAccount().Return(defaultServiceAccountName)
	taskExecutionMetadata.OnGetNamespace().Return(defaultNamespace)
	taskExecutionMetadata.OnGetConsoleURL().Return("")
	overrides := &mocks.TaskOverrides{}
	overrides.OnGetResources().Return(resources)
	overrides.OnGetExtendedResources().Return(extendedResources)
	overrides.OnGetContainerImage().Return("")
	taskExecutionMetadata.OnGetOverrides().Return(overrides)
	taskCtx.On("TaskExecutionMetadata").Return(taskExecutionMetadata)
	return taskCtx
}

func TestBuildResourceDaskHappyPath(t *testing.T) {
	daskResourceHandler := daskResourceHandler{}

	taskTemplate := dummyDaskTaskTemplate("", nil, "")
	taskContext := dummyDaskTaskContext(taskTemplate, &defaultResources, nil, false)
	r, err := daskResourceHandler.BuildResource(context.TODO(), taskContext)
	assert.Nil(t, err)
	assert.NotNil(t, r)
	daskJob, ok := r.(*daskAPI.DaskJob)
	assert.True(t, ok)

	var defaultTolerations []v1.Toleration
	defaultNodeSelector := map[string]string{}
	defaultAffinity := &v1.Affinity{
		NodeAffinity:    nil,
		PodAffinity:     nil,
		PodAntiAffinity: nil,
	}

	// Job
	jobSpec := daskJob.Spec.Job.Spec
	assert.Equal(t, testAnnotations, daskJob.ObjectMeta.GetAnnotations())
	assert.Equal(t, testLabels, daskJob.ObjectMeta.GetLabels())
	assert.Equal(t, v1.RestartPolicyNever, jobSpec.RestartPolicy)
	assert.Equal(t, "job-runner", jobSpec.Containers[0].Name)
	assert.Equal(t, defaultTestImage, jobSpec.Containers[0].Image)
	assert.Equal(t, testArgs, jobSpec.Containers[0].Args)
	assert.Equal(t, defaultResources, jobSpec.Containers[0].Resources)
	assert.Equal(t, defaultTolerations, jobSpec.Tolerations)
	assert.Equal(t, defaultNodeSelector, jobSpec.NodeSelector)
	assert.Equal(t, defaultAffinity, jobSpec.Affinity)

	// Flyte adds more environment variables to the runner
	assert.Contains(t, jobSpec.Containers[0].Env, testEnvVars[0])

	// Cluster
	assert.Equal(t, testAnnotations, daskJob.Spec.Cluster.ObjectMeta.GetAnnotations())
	assert.Equal(t, testLabels, daskJob.Spec.Cluster.ObjectMeta.GetLabels())

	// Scheduler
	schedulerSpec := daskJob.Spec.Cluster.Spec.Scheduler.Spec
	expectedPorts := []v1.ContainerPort{
		{
			Name:          "tcp-comm",
			ContainerPort: 8786,
			Protocol:      "TCP",
		},
		{
			Name:          "dashboard",
			ContainerPort: 8787,
			Protocol:      "TCP",
		},
	}
	assert.Equal(t, v1.RestartPolicyAlways, schedulerSpec.RestartPolicy)
	assert.Equal(t, defaultTestImage, schedulerSpec.Containers[0].Image)
	assert.Equal(t, defaultResources, schedulerSpec.Containers[0].Resources)
	assert.Equal(t, []string{"dask-scheduler"}, schedulerSpec.Containers[0].Args)
	assert.Equal(t, expectedPorts, schedulerSpec.Containers[0].Ports)
	// Flyte adds more environment variables to the scheduler
	assert.Contains(t, schedulerSpec.Containers[0].Env, testEnvVars[0])
	assert.Equal(t, defaultTolerations, schedulerSpec.Tolerations)
	assert.Equal(t, defaultNodeSelector, schedulerSpec.NodeSelector)
	assert.Equal(t, defaultAffinity, schedulerSpec.Affinity)

	schedulerServiceSpec := daskJob.Spec.Cluster.Spec.Scheduler.Service
	expectedSelector := map[string]string{
		"dask.org/cluster-name": testTaskID,
		"dask.org/component":    "scheduler",
	}
	expectedSerivcePorts := []v1.ServicePort{
		{
			Name:       "tcp-comm",
			Protocol:   "TCP",
			Port:       8786,
			TargetPort: intstr.FromString("tcp-comm"),
		},
		{
			Name:       "dashboard",
			Protocol:   "TCP",
			Port:       8787,
			TargetPort: intstr.FromString("dashboard"),
		},
	}
	assert.Equal(t, v1.ServiceTypeNodePort, schedulerServiceSpec.Type)
	assert.Equal(t, expectedSelector, schedulerServiceSpec.Selector)
	assert.Equal(t, expectedSerivcePorts, schedulerServiceSpec.Ports)

	// Default Workers
	workerSpec := daskJob.Spec.Cluster.Spec.Worker.Spec
	assert.Equal(t, testNWorkers, daskJob.Spec.Cluster.Spec.Worker.Replicas)
	assert.Equal(t, "dask-worker", workerSpec.Containers[0].Name)
	assert.Equal(t, defaultTestImage, workerSpec.Containers[0].Image)
	assert.Equal(t, defaultResources, workerSpec.Containers[0].Resources)
	// Flyte adds more environment variables to the worker
	assert.Contains(t, workerSpec.Containers[0].Env, testEnvVars[0])
	assert.Equal(t, defaultTolerations, workerSpec.Tolerations)
	assert.Equal(t, defaultNodeSelector, workerSpec.NodeSelector)
	assert.Equal(t, defaultAffinity, workerSpec.Affinity)
	assert.Equal(t, []string{
		"dask-worker",
		"--name",
		"$(DASK_WORKER_NAME)",
		"--nthreads",
		"4",
		"--memory-limit",
		"1Gi",
	}, workerSpec.Containers[0].Args)
	assert.Equal(t, workerSpec.RestartPolicy, v1.RestartPolicyAlways)
}

func TestBuildResourceDaskCustomImages(t *testing.T) {
	customImage := "customImage"

	daskResourceHandler := daskResourceHandler{}
	taskTemplate := dummyDaskTaskTemplate(customImage, nil, "")
	taskContext := dummyDaskTaskContext(taskTemplate, &v1.ResourceRequirements{}, nil, false)
	r, err := daskResourceHandler.BuildResource(context.TODO(), taskContext)
	assert.Nil(t, err)
	assert.NotNil(t, r)
	daskJob, ok := r.(*daskAPI.DaskJob)
	assert.True(t, ok)

	// Job
	jobSpec := daskJob.Spec.Job.Spec
	assert.Equal(t, defaultTestImage, jobSpec.Containers[0].Image)

	// Scheduler
	schedulerSpec := daskJob.Spec.Cluster.Spec.Scheduler.Spec
	assert.Equal(t, customImage, schedulerSpec.Containers[0].Image)

	// Default Workers
	workerSpec := daskJob.Spec.Cluster.Spec.Worker.Spec
	assert.Equal(t, customImage, workerSpec.Containers[0].Image)
}

func TestBuildResourceDaskDefaultResoureRequirements(t *testing.T) {
	flyteWorkflowResources := v1.ResourceRequirements{
		Requests: v1.ResourceList{
			v1.ResourceCPU: resource.MustParse("1"),
		},
		Limits: v1.ResourceList{
			v1.ResourceCPU:    resource.MustParse("2"),
			v1.ResourceMemory: resource.MustParse("2G"),
		},
	}

	daskResourceHandler := daskResourceHandler{}
	taskTemplate := dummyDaskTaskTemplate("", nil, "")
	taskContext := dummyDaskTaskContext(taskTemplate, &flyteWorkflowResources, nil, false)
	r, err := daskResourceHandler.BuildResource(context.TODO(), taskContext)
	assert.Nil(t, err)
	assert.NotNil(t, r)
	daskJob, ok := r.(*daskAPI.DaskJob)
	assert.True(t, ok)

	// Job
	jobSpec := daskJob.Spec.Job.Spec
	assert.Equal(t, flyteWorkflowResources, jobSpec.Containers[0].Resources)

	// Scheduler
	schedulerSpec := daskJob.Spec.Cluster.Spec.Scheduler.Spec
	assert.Equal(t, flyteWorkflowResources, schedulerSpec.Containers[0].Resources)

	// Default Workers
	workerSpec := daskJob.Spec.Cluster.Spec.Worker.Spec
	assert.Equal(t, flyteWorkflowResources, workerSpec.Containers[0].Resources)
	assert.Contains(t, workerSpec.Containers[0].Args, "--nthreads")
	assert.Contains(t, workerSpec.Containers[0].Args, "2")
	assert.Contains(t, workerSpec.Containers[0].Args, "--memory-limit")
	assert.Contains(t, workerSpec.Containers[0].Args, "2G")
}

func TestBuildResourcesDaskCustomResoureRequirements(t *testing.T) {
	protobufResources := core.Resources{
		Requests: []*core.Resources_ResourceEntry{
			{
				Name:  core.Resources_CPU,
				Value: "5",
			},
		},
		Limits: []*core.Resources_ResourceEntry{
			{
				Name:  core.Resources_CPU,
				Value: "10",
			},
			{
				Name:  core.Resources_MEMORY,
				Value: "15G",
			},
		},
	}
	expectedResources, _ := flytek8s.ToK8sResourceRequirements(&protobufResources)

	flyteWorkflowResources := v1.ResourceRequirements{
		Requests: v1.ResourceList{
			v1.ResourceCPU: resource.MustParse("1"),
		},
		Limits: v1.ResourceList{
			v1.ResourceCPU:    resource.MustParse("2"),
			v1.ResourceMemory: resource.MustParse("2G"),
		},
	}

	daskResourceHandler := daskResourceHandler{}
	taskTemplate := dummyDaskTaskTemplate("", &protobufResources, "")
	taskContext := dummyDaskTaskContext(taskTemplate, &flyteWorkflowResources, nil, false)
	r, err := daskResourceHandler.BuildResource(context.TODO(), taskContext)
	assert.Nil(t, err)
	assert.NotNil(t, r)
	daskJob, ok := r.(*daskAPI.DaskJob)
	assert.True(t, ok)

	// Job
	jobSpec := daskJob.Spec.Job.Spec
	assert.Equal(t, flyteWorkflowResources, jobSpec.Containers[0].Resources)

	// Scheduler
	schedulerSpec := daskJob.Spec.Cluster.Spec.Scheduler.Spec
	assert.Equal(t, *expectedResources, schedulerSpec.Containers[0].Resources)

	// Default Workers
	workerSpec := daskJob.Spec.Cluster.Spec.Worker.Spec
	assert.Equal(t, *expectedResources, workerSpec.Containers[0].Resources)
	assert.Contains(t, workerSpec.Containers[0].Args, "--nthreads")
	assert.Contains(t, workerSpec.Containers[0].Args, "10")
	assert.Contains(t, workerSpec.Containers[0].Args, "--memory-limit")
	assert.Contains(t, workerSpec.Containers[0].Args, "15G")
}

func TestBuildResourceDaskInterruptible(t *testing.T) {
	defaultNodeSelector := map[string]string{}
	var defaultAffinity v1.Affinity
	var defaultTolerations []v1.Toleration

	interruptibleNodeSelector := map[string]string{
		"x/interruptible": "true",
	}
	interruptibleNodeSelectorRequirement := &v1.NodeSelectorRequirement{
		Key:      "x/interruptible",
		Operator: v1.NodeSelectorOpIn,
		Values:   []string{"true"},
	}
	interruptibleTolerations := []v1.Toleration{
		{
			Key:      "x/flyte",
			Value:    "interruptible",
			Operator: "Equal",
			Effect:   "NoSchedule",
		},
	}

	assert.NoError(t, config.SetK8sPluginConfig(&config.K8sPluginConfig{
		InterruptibleNodeSelector:            interruptibleNodeSelector,
		InterruptibleNodeSelectorRequirement: interruptibleNodeSelectorRequirement,
		InterruptibleTolerations:             interruptibleTolerations,
	}))

	daskResourceHandler := daskResourceHandler{}

	taskTemplate := dummyDaskTaskTemplate("", nil, "")
	taskContext := dummyDaskTaskContext(taskTemplate, &defaultResources, nil, true)
	r, err := daskResourceHandler.BuildResource(context.TODO(), taskContext)
	assert.Nil(t, err)
	assert.NotNil(t, r)
	daskJob, ok := r.(*daskAPI.DaskJob)
	assert.True(t, ok)

	// Job pod - should not be interruptible
	jobSpec := daskJob.Spec.Job.Spec
	assert.Equal(t, defaultTolerations, jobSpec.Tolerations)
	assert.Equal(t, defaultNodeSelector, jobSpec.NodeSelector)
	assert.Equal(t, &defaultAffinity, jobSpec.Affinity)

	// Scheduler - should not be interruptible
	schedulerSpec := daskJob.Spec.Cluster.Spec.Scheduler.Spec
	assert.Equal(t, defaultTolerations, schedulerSpec.Tolerations)
	assert.Equal(t, defaultNodeSelector, schedulerSpec.NodeSelector)
	assert.Equal(t, &defaultAffinity, schedulerSpec.Affinity)

	// Default Workers - Should be interruptible
	workerSpec := daskJob.Spec.Cluster.Spec.Worker.Spec
	assert.Equal(t, interruptibleTolerations, workerSpec.Tolerations)
	assert.Equal(t, interruptibleNodeSelector, workerSpec.NodeSelector)
	assert.Equal(
		t,
		workerSpec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchExpressions[0],
		*interruptibleNodeSelectorRequirement,
	)
}

func TestBuildResouceDaskUsePodTemplate(t *testing.T) {
	flytek8s.DefaultPodTemplateStore.Store(podTemplate)
	daskResourceHandler := daskResourceHandler{}
	taskTemplate := dummyDaskTaskTemplate("", nil, podTemplateName)
	taskContext := dummyDaskTaskContext(taskTemplate, &defaultResources, nil, false)
	r, err := daskResourceHandler.BuildResource(context.TODO(), taskContext)
	assert.Nil(t, err)
	assert.NotNil(t, r)
	daskJob, ok := r.(*daskAPI.DaskJob)
	assert.True(t, ok)

	assert.Equal(t, podTempaltePriorityClassName, daskJob.Spec.Job.Spec.PriorityClassName)
	assert.Equal(t, podTempaltePriorityClassName, daskJob.Spec.Cluster.Spec.Scheduler.Spec.PriorityClassName)
	assert.Equal(t, podTempaltePriorityClassName, daskJob.Spec.Cluster.Spec.Worker.Spec.PriorityClassName)

	// Cleanup
	flytek8s.DefaultPodTemplateStore.Delete(podTemplate)
}

func TestBuildResourceDaskExtendedResources(t *testing.T) {
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

	for _, f := range fixtures {
		t.Run(f.name, func(t *testing.T) {
			taskTemplate := dummyDaskTaskTemplate("", nil, "")
			taskTemplate.ExtendedResources = f.extendedResourcesBase
			taskContext := dummyDaskTaskContext(taskTemplate, f.resources, f.extendedResourcesOverride, false)
			daskResourceHandler := daskResourceHandler{}
			r, err := daskResourceHandler.BuildResource(context.TODO(), taskContext)
			assert.Nil(t, err)
			assert.NotNil(t, r)
			daskJob, ok := r.(*daskAPI.DaskJob)
			assert.True(t, ok)

			// Job pod
			jobSpec := daskJob.Spec.Job.Spec
			assert.EqualValues(
				t,
				f.expectedNsr,
				jobSpec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms,
			)
			assert.EqualValues(
				t,
				f.expectedTol,
				jobSpec.Tolerations,
			)

			// Scheduler
			schedulerSpec := daskJob.Spec.Cluster.Spec.Scheduler.Spec
			assert.EqualValues(
				t,
				f.expectedNsr,
				schedulerSpec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms,
			)
			assert.EqualValues(
				t,
				f.expectedTol,
				schedulerSpec.Tolerations,
			)

			// Default Workers
			workerSpec := daskJob.Spec.Cluster.Spec.Worker.Spec
			assert.EqualValues(
				t,
				f.expectedNsr,
				workerSpec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms,
			)
			assert.EqualValues(
				t,
				f.expectedTol,
				workerSpec.Tolerations,
			)
		})
	}
}

func TestGetPropertiesDask(t *testing.T) {
	daskResourceHandler := daskResourceHandler{}
	expected := k8s.PluginProperties{}
	assert.Equal(t, expected, daskResourceHandler.GetProperties())
}

func TestBuildIdentityResourceDask(t *testing.T) {
	daskResourceHandler := daskResourceHandler{}
	expected := &daskAPI.DaskJob{
		TypeMeta: metav1.TypeMeta{
			Kind:       KindDaskJob,
			APIVersion: daskAPI.SchemeGroupVersion.String(),
		},
	}

	taskTemplate := dummyDaskTaskTemplate("", nil, "")
	taskContext := dummyDaskTaskContext(taskTemplate, &v1.ResourceRequirements{}, nil, false)
	identityResources, err := daskResourceHandler.BuildIdentityResource(context.TODO(), taskContext.TaskExecutionMetadata())
	if err != nil {
		panic(err)
	}
	assert.Equal(t, expected, identityResources)
}

func TestGetTaskPhaseDask(t *testing.T) {
	daskResourceHandler := daskResourceHandler{}
	ctx := context.TODO()

	taskTemplate := dummyDaskTaskTemplate("", nil, "")
	taskCtx := dummyDaskTaskContext(taskTemplate, &v1.ResourceRequirements{}, nil, false)

	taskPhase, err := daskResourceHandler.GetTaskPhase(ctx, taskCtx, dummyDaskJob(""))
	assert.NoError(t, err)
	assert.Equal(t, taskPhase.Phase(), pluginsCore.PhaseInitializing)
	assert.NotNil(t, taskPhase.Info())
	assert.Nil(t, taskPhase.Info().Logs)
	assert.Nil(t, err)

	taskPhase, err = daskResourceHandler.GetTaskPhase(ctx, taskCtx, dummyDaskJob(daskAPI.DaskJobCreated))
	assert.NoError(t, err)
	assert.Equal(t, taskPhase.Phase(), pluginsCore.PhaseInitializing)
	assert.NotNil(t, taskPhase.Info())
	assert.Nil(t, taskPhase.Info().Logs)
	assert.Nil(t, err)

	taskPhase, err = daskResourceHandler.GetTaskPhase(ctx, taskCtx, dummyDaskJob(daskAPI.DaskJobClusterCreated))
	assert.NoError(t, err)
	assert.Equal(t, taskPhase.Phase(), pluginsCore.PhaseInitializing)
	assert.NotNil(t, taskPhase.Info())
	assert.Nil(t, taskPhase.Info().Logs)
	assert.Nil(t, err)

	taskPhase, err = daskResourceHandler.GetTaskPhase(ctx, taskCtx, dummyDaskJob(daskAPI.DaskJobRunning))
	assert.NoError(t, err)
	assert.Equal(t, taskPhase.Phase(), pluginsCore.PhaseRunning)
	assert.NotNil(t, taskPhase.Info())
	assert.NotNil(t, taskPhase.Info().Logs)
	assert.Nil(t, err)

	taskPhase, err = daskResourceHandler.GetTaskPhase(ctx, taskCtx, dummyDaskJob(daskAPI.DaskJobSuccessful))
	assert.NoError(t, err)
	assert.Equal(t, taskPhase.Phase(), pluginsCore.PhaseSuccess)
	assert.NotNil(t, taskPhase.Info())
	assert.NotNil(t, taskPhase.Info().Logs)
	assert.Nil(t, err)

	taskPhase, err = daskResourceHandler.GetTaskPhase(ctx, taskCtx, dummyDaskJob(daskAPI.DaskJobFailed))
	assert.NoError(t, err)
	assert.Equal(t, taskPhase.Phase(), pluginsCore.PhaseRetryableFailure)
	assert.NotNil(t, taskPhase.Info())
	assert.NotNil(t, taskPhase.Info().Logs)
	assert.Nil(t, err)
}
