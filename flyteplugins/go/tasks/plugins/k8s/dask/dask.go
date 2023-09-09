package dask

import (
	"context"
	"fmt"
	"time"

	daskAPI "github.com/dask/dask-kubernetes/v2023/dask_kubernetes/operator/go_client/pkg/apis/kubernetes.dask.org/v1"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/plugins"
	"github.com/flyteorg/flyteplugins/go/tasks/errors"
	"github.com/flyteorg/flyteplugins/go/tasks/logs"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery"
	pluginsCore "github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/core/template"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/flytek8s"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/flytek8s/config"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/k8s"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/tasklog"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/utils"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	daskTaskType = "dask"
	KindDaskJob  = "DaskJob"
)

type defaults struct {
	Image              string
	JobRunnerContainer v1.Container
	Resources          *v1.ResourceRequirements
	Env                []v1.EnvVar
	Annotations        map[string]string
	IsInterruptible    bool
}

func getDefaults(ctx context.Context, taskCtx pluginsCore.TaskExecutionContext, taskTemplate core.TaskTemplate) (*defaults, error) {
	executionMetadata := taskCtx.TaskExecutionMetadata()

	defaultContainerSpec := taskTemplate.GetContainer()
	if defaultContainerSpec == nil {
		return nil, errors.Errorf(errors.BadTaskSpecification, "task is missing a default container")
	}

	defaultImage := defaultContainerSpec.GetImage()
	if defaultImage == "" {
		return nil, errors.Errorf(errors.BadTaskSpecification, "task is missing a default image")
	}

	var defaultEnvVars []v1.EnvVar
	if taskTemplate.GetContainer().GetEnv() != nil {
		for _, keyValuePair := range taskTemplate.GetContainer().GetEnv() {
			defaultEnvVars = append(defaultEnvVars, v1.EnvVar{Name: keyValuePair.Key, Value: keyValuePair.Value})
		}
	}

	containerResources, err := flytek8s.ToK8sResourceRequirements(defaultContainerSpec.GetResources())
	if err != nil {
		return nil, err
	}

	jobRunnerContainer := v1.Container{
		Name:      "job-runner",
		Image:     defaultImage,
		Args:      defaultContainerSpec.GetArgs(),
		Env:       defaultEnvVars,
		Resources: *containerResources,
	}

	templateParameters := template.Parameters{
		TaskExecMetadata: taskCtx.TaskExecutionMetadata(),
		Inputs:           taskCtx.InputReader(),
		OutputPath:       taskCtx.OutputWriter(),
		Task:             taskCtx.TaskReader(),
	}
	if err = flytek8s.AddFlyteCustomizationsToContainer(ctx, templateParameters,
		flytek8s.ResourceCustomizationModeMergeExistingResources, &jobRunnerContainer); err != nil {

		return nil, err
	}

	return &defaults{
		Image:              defaultImage,
		JobRunnerContainer: jobRunnerContainer,
		Resources:          &jobRunnerContainer.Resources,
		Env:                defaultEnvVars,
		Annotations:        executionMetadata.GetAnnotations(),
		IsInterruptible:    executionMetadata.IsInterruptible(),
	}, nil
}

type daskResourceHandler struct {
}

func (daskResourceHandler) BuildIdentityResource(_ context.Context, _ pluginsCore.TaskExecutionMetadata) (
	client.Object, error) {
	return &daskAPI.DaskJob{
		TypeMeta: metav1.TypeMeta{
			Kind:       KindDaskJob,
			APIVersion: daskAPI.SchemeGroupVersion.String(),
		},
	}, nil
}

func (p daskResourceHandler) BuildResource(ctx context.Context, taskCtx pluginsCore.TaskExecutionContext) (client.Object, error) {
	taskTemplate, err := taskCtx.TaskReader().Read(ctx)
	if err != nil {
		return nil, errors.Errorf(errors.BadTaskSpecification, "unable to fetch task specification [%v]", err.Error())
	} else if taskTemplate == nil {
		return nil, errors.Errorf(errors.BadTaskSpecification, "nil task specification")
	}
	defaults, err := getDefaults(ctx, taskCtx, *taskTemplate)
	if err != nil {
		return nil, err
	}
	clusterName := taskCtx.TaskExecutionMetadata().GetTaskExecutionID().GetGeneratedName()

	daskJob := plugins.DaskJob{}
	err = utils.UnmarshalStruct(taskTemplate.GetCustom(), &daskJob)
	if err != nil {
		return nil, errors.Wrapf(errors.BadTaskSpecification, err, "invalid TaskSpecification [%v], failed to unmarshal", taskTemplate.GetCustom())
	}

	workerSpec, err := createWorkerSpec(*daskJob.Workers, *defaults)
	if err != nil {
		return nil, err
	}
	schedulerSpec, err := createSchedulerSpec(*daskJob.Scheduler, clusterName, *defaults)
	if err != nil {
		return nil, err
	}
	jobSpec := createJobSpec(*workerSpec, *schedulerSpec, *defaults)

	job := &daskAPI.DaskJob{
		TypeMeta: metav1.TypeMeta{
			Kind:       KindDaskJob,
			APIVersion: daskAPI.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        "will-be-overridden", // Will be overridden by Flyte to `clusterName`
			Annotations: defaults.Annotations,
		},
		Spec: *jobSpec,
	}
	return job, nil
}

func createWorkerSpec(cluster plugins.DaskWorkerGroup, defaults defaults) (*daskAPI.WorkerSpec, error) {
	image := defaults.Image
	if cluster.GetImage() != "" {
		image = cluster.GetImage()
	}

	var err error
	resources := defaults.Resources
	clusterResources := cluster.GetResources()
	if len(clusterResources.Requests) >= 1 || len(clusterResources.Limits) >= 1 {
		resources, err = flytek8s.ToK8sResourceRequirements(cluster.GetResources())
		if err != nil {
			return nil, err
		}
	}
	if resources == nil {
		resources = &v1.ResourceRequirements{}
	}

	workerArgs := []string{
		"dask-worker",
		"--name",
		"$(DASK_WORKER_NAME)",
	}

	// If limits are set, append `--nthreads` and `--memory-limit` as per these docs:
	// https://kubernetes.dask.org/en/latest/kubecluster.html?#best-practices
	if resources != nil && resources.Limits != nil {
		limits := resources.Limits
		if limits.Cpu() != nil {
			cpuCount := fmt.Sprintf("%v", limits.Cpu().Value())
			workerArgs = append(workerArgs, "--nthreads", cpuCount)
		}
		if limits.Memory() != nil {
			memory := limits.Memory().String()
			workerArgs = append(workerArgs, "--memory-limit", memory)
		}
	}

	wokerSpec := v1.PodSpec{
		Affinity: &v1.Affinity{},
		Containers: []v1.Container{
			{
				Name:            "dask-worker",
				Image:           image,
				ImagePullPolicy: v1.PullIfNotPresent,
				Args:            workerArgs,
				Resources:       *resources,
				Env:             defaults.Env,
			},
		},
	}

	if defaults.IsInterruptible {
		wokerSpec.Tolerations = append(wokerSpec.Tolerations, config.GetK8sPluginConfig().InterruptibleTolerations...)
		wokerSpec.NodeSelector = config.GetK8sPluginConfig().InterruptibleNodeSelector
	}
	flytek8s.ApplyInterruptibleNodeSelectorRequirement(defaults.IsInterruptible, wokerSpec.Affinity)

	return &daskAPI.WorkerSpec{
		Replicas: int(cluster.GetNumberOfWorkers()),
		Spec:     wokerSpec,
	}, nil
}

func createSchedulerSpec(cluster plugins.DaskScheduler, clusterName string, defaults defaults) (*daskAPI.SchedulerSpec, error) {
	schedulerImage := defaults.Image
	if cluster.GetImage() != "" {
		schedulerImage = cluster.GetImage()
	}

	var err error
	resources := defaults.Resources

	clusterResources := cluster.GetResources()
	if len(clusterResources.Requests) >= 1 || len(clusterResources.Limits) >= 1 {
		resources, err = flytek8s.ToK8sResourceRequirements(cluster.GetResources())
		if err != nil {
			return nil, err
		}
	}
	if resources == nil {
		resources = &v1.ResourceRequirements{}
	}

	return &daskAPI.SchedulerSpec{
		Spec: v1.PodSpec{
			RestartPolicy: v1.RestartPolicyAlways,
			Containers: []v1.Container{
				{
					Name:      "scheduler",
					Image:     schedulerImage,
					Args:      []string{"dask-scheduler"},
					Resources: *resources,
					Env:       defaults.Env,
					Ports: []v1.ContainerPort{
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
					},
				},
			},
		},
		Service: v1.ServiceSpec{
			Type: v1.ServiceTypeNodePort,
			Selector: map[string]string{
				"dask.org/cluster-name": clusterName,
				"dask.org/component":    "scheduler",
			},
			Ports: []v1.ServicePort{
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
			},
		},
	}, nil
}

func createJobSpec(workerSpec daskAPI.WorkerSpec, schedulerSpec daskAPI.SchedulerSpec, defaults defaults) *daskAPI.DaskJobSpec {
	return &daskAPI.DaskJobSpec{
		Job: daskAPI.JobSpec{
			Spec: v1.PodSpec{
				RestartPolicy: v1.RestartPolicyNever,
				Containers: []v1.Container{
					defaults.JobRunnerContainer,
				},
			},
		},
		Cluster: daskAPI.DaskCluster{
			ObjectMeta: metav1.ObjectMeta{
				Annotations: defaults.Annotations,
			},
			Spec: daskAPI.DaskClusterSpec{
				Worker:    workerSpec,
				Scheduler: schedulerSpec,
			},
		},
	}
}

func (p daskResourceHandler) GetTaskPhase(ctx context.Context, pluginContext k8s.PluginContext, r client.Object) (pluginsCore.PhaseInfo, error) {
	logPlugin, err := logs.InitializeLogPlugins(logs.GetLogConfig())
	if err != nil {
		return pluginsCore.PhaseInfoUndefined, err
	}
	job := r.(*daskAPI.DaskJob)
	status := job.Status.JobStatus
	occurredAt := time.Now()

	info := pluginsCore.TaskInfo{
		OccurredAt: &occurredAt,
	}

	// There is a short period between the `DaskJob` resource being created and `Status.JobStatus` being set by the `dask-operator`.
	// In that period, the `JobStatus` will be an empty string. We're treating this as Initializing/Queuing.
	isQueued := status == "" ||
		status == daskAPI.DaskJobCreated ||
		status == daskAPI.DaskJobClusterCreated

	if !isQueued {
		taskExecID := pluginContext.TaskExecutionMetadata().GetTaskExecutionID().GetID()
		o, err := logPlugin.GetTaskLogs(
			tasklog.Input{
				Namespace:               job.ObjectMeta.Namespace,
				PodName:                 job.Status.JobRunnerPodName,
				LogName:                 "(User logs)",
				TaskExecutionIdentifier: &taskExecID,
			},
		)
		if err != nil {
			return pluginsCore.PhaseInfoUndefined, err
		}
		info.Logs = o.TaskLogs
	}

	switch status {
	case "":
		return pluginsCore.PhaseInfoInitializing(occurredAt, pluginsCore.DefaultPhaseVersion, "unknown", &info), nil
	case daskAPI.DaskJobCreated:
		return pluginsCore.PhaseInfoInitializing(occurredAt, pluginsCore.DefaultPhaseVersion, "job created", &info), nil
	case daskAPI.DaskJobClusterCreated:
		return pluginsCore.PhaseInfoInitializing(occurredAt, pluginsCore.DefaultPhaseVersion, "cluster created", &info), nil
	case daskAPI.DaskJobFailed:
		reason := "Dask Job failed"
		return pluginsCore.PhaseInfoRetryableFailure(errors.DownstreamSystemError, reason, &info), nil
	case daskAPI.DaskJobSuccessful:
		return pluginsCore.PhaseInfoSuccess(&info), nil
	}
	return pluginsCore.PhaseInfoRunning(pluginsCore.DefaultPhaseVersion, &info), nil
}

func (daskResourceHandler) GetProperties() k8s.PluginProperties {
	return k8s.PluginProperties{}
}

func init() {
	if err := daskAPI.AddToScheme(scheme.Scheme); err != nil {
		panic(err)
	}

	pluginmachinery.PluginRegistry().RegisterK8sPlugin(
		k8s.PluginEntry{
			ID:                  daskTaskType,
			RegisteredTaskTypes: []pluginsCore.TaskType{daskTaskType},
			ResourceToWatch:     &daskAPI.DaskJob{},
			Plugin:              daskResourceHandler{},
			IsDefault:           false,
		})
}
