package dask

import (
	"context"
	"fmt"
	"time"

	daskAPI "github.com/dask/dask-kubernetes/v2023/dask_kubernetes/operator/go_client/pkg/apis/kubernetes.dask.org/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/plugins"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/errors"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/logs"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery"
	pluginsCore "github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/flytek8s"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/k8s"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/tasklog"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/utils"
)

const (
	daskTaskType = "dask"
	KindDaskJob  = "DaskJob"
)

func mergeMapInto(src map[string]string, dst map[string]string) {
	for key, value := range src {
		dst[key] = value
	}
}

func replacePrimaryContainer(spec *v1.PodSpec, primaryContainerName string, container v1.Container) error {
	for i, c := range spec.Containers {
		if c.Name == primaryContainerName {
			spec.Containers[i] = container
			return nil
		}
	}
	return errors.Errorf(errors.BadTaskSpecification, "primary container [%v] not found in pod spec", primaryContainerName)
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

	daskJob := plugins.DaskJob{}
	err = utils.UnmarshalStruct(taskTemplate.GetCustom(), &daskJob)
	if err != nil {
		return nil, errors.Wrapf(errors.BadTaskSpecification, err, "invalid TaskSpecification [%v], failed to unmarshal", taskTemplate.GetCustom())
	}

	podSpec, objectMeta, primaryContainerName, err := flytek8s.ToK8sPodSpec(ctx, taskCtx)
	if err != nil {
		return nil, err
	}
	nonInterruptibleTaskCtx := flytek8s.NewPluginTaskExecutionContext(taskCtx, flytek8s.WithInterruptible(false))
	nonInterruptiblePodSpec, _, _, err := flytek8s.ToK8sPodSpec(ctx, nonInterruptibleTaskCtx)
	if err != nil {
		return nil, err
	}

	// Add labels and annotations to objectMeta as they're not added by ToK8sPodSpec
	mergeMapInto(taskCtx.TaskExecutionMetadata().GetAnnotations(), objectMeta.Annotations)
	mergeMapInto(taskCtx.TaskExecutionMetadata().GetLabels(), objectMeta.Labels)

	workerSpec, err := createWorkerSpec(*daskJob.Workers, podSpec, primaryContainerName)
	if err != nil {
		return nil, err
	}

	clusterName := taskCtx.TaskExecutionMetadata().GetTaskExecutionID().GetGeneratedName()
	schedulerSpec, err := createSchedulerSpec(*daskJob.Scheduler, clusterName, nonInterruptiblePodSpec, primaryContainerName)
	if err != nil {
		return nil, err
	}

	jobSpec, err := createJobSpec(*workerSpec, *schedulerSpec, nonInterruptiblePodSpec, primaryContainerName, objectMeta)
	if err != nil {
		return nil, err
	}

	job := &daskAPI.DaskJob{
		TypeMeta: metav1.TypeMeta{
			Kind:       KindDaskJob,
			APIVersion: daskAPI.SchemeGroupVersion.String(),
		},
		ObjectMeta: *objectMeta,
		Spec:       *jobSpec,
	}
	return job, nil
}

func createWorkerSpec(cluster plugins.DaskWorkerGroup, podSpec *v1.PodSpec, primaryContainerName string) (*daskAPI.WorkerSpec, error) {
	workerPodSpec := podSpec.DeepCopy()
	primaryContainer, err := flytek8s.GetContainer(workerPodSpec, primaryContainerName)
	if err != nil {
		return nil, err
	}
	primaryContainer.Name = "dask-worker"

	// Set custom image if present
	if cluster.GetImage() != "" {
		primaryContainer.Image = cluster.GetImage()
	}

	// Set custom resources
	resources := &primaryContainer.Resources
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
	primaryContainer.Resources = *resources

	// Set custom args
	workerArgs := []string{
		"dask-worker",
		"--name",
		"$(DASK_WORKER_NAME)",
	}
	// If limits are set, append `--nthreads` and `--memory-limit` as per these docs:
	// https://kubernetes.dask.org/en/latest/kubecluster.html?#best-practices
	if resources.Limits != nil {
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
	primaryContainer.Args = workerArgs

	err = replacePrimaryContainer(workerPodSpec, primaryContainerName, *primaryContainer)
	if err != nil {
		return nil, err
	}

	// All workers are created as k8s deployment and must have a restart policy of Always
	workerPodSpec.RestartPolicy = v1.RestartPolicyAlways

	return &daskAPI.WorkerSpec{
		Replicas: int(cluster.GetNumberOfWorkers()),
		Spec:     *workerPodSpec,
	}, nil
}

func createSchedulerSpec(scheduler plugins.DaskScheduler, clusterName string, podSpec *v1.PodSpec, primaryContainerName string) (*daskAPI.SchedulerSpec, error) {
	schedulerPodSpec := podSpec.DeepCopy()
	primaryContainer, err := flytek8s.GetContainer(schedulerPodSpec, primaryContainerName)
	if err != nil {
		return nil, err
	}
	primaryContainer.Name = "scheduler"

	// Override image if applicable
	if scheduler.GetImage() != "" {
		primaryContainer.Image = scheduler.GetImage()
	}

	// Override resources if applicable
	resources := &primaryContainer.Resources
	schedulerResources := scheduler.GetResources()
	if len(schedulerResources.Requests) >= 1 || len(schedulerResources.Limits) >= 1 {
		resources, err = flytek8s.ToK8sResourceRequirements(scheduler.GetResources())
		if err != nil {
			return nil, err
		}
	}
	primaryContainer.Resources = *resources

	// Override args
	primaryContainer.Args = []string{"dask-scheduler"}

	// Add ports
	primaryContainer.Ports = []v1.ContainerPort{
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

	schedulerPodSpec.RestartPolicy = v1.RestartPolicyAlways

	// Set primary container
	err = replacePrimaryContainer(schedulerPodSpec, primaryContainerName, *primaryContainer)
	if err != nil {
		return nil, err
	}

	return &daskAPI.SchedulerSpec{
		Spec: *schedulerPodSpec,
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

func createJobSpec(workerSpec daskAPI.WorkerSpec, schedulerSpec daskAPI.SchedulerSpec, podSpec *v1.PodSpec, primaryContainerName string, objectMeta *metav1.ObjectMeta) (*daskAPI.DaskJobSpec, error) {
	jobPodSpec := podSpec.DeepCopy()
	jobPodSpec.RestartPolicy = v1.RestartPolicyNever

	primaryContainer, err := flytek8s.GetContainer(jobPodSpec, primaryContainerName)
	if err != nil {
		return nil, err
	}
	primaryContainer.Name = "job-runner"

	err = replacePrimaryContainer(jobPodSpec, primaryContainerName, *primaryContainer)
	if err != nil {
		return nil, err
	}

	return &daskAPI.DaskJobSpec{
		Job: daskAPI.JobSpec{
			Spec: *jobPodSpec,
		},
		Cluster: daskAPI.DaskCluster{
			ObjectMeta: *objectMeta,
			Spec: daskAPI.DaskClusterSpec{
				Worker:    workerSpec,
				Scheduler: schedulerSpec,
			},
		},
	}, nil
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
	// isQueued := status == "" ||
	// 	status == daskAPI.DaskJobCreated ||
	// 	status == daskAPI.DaskJobClusterCreated

	taskExecID := pluginContext.TaskExecutionMetadata().GetTaskExecutionID()
	o, err := logPlugin.GetTaskLogs(
		tasklog.Input{
			Namespace:       job.ObjectMeta.Namespace,
			PodName:         job.Status.JobRunnerPodName,
			LogName:         "(User logs)",
			TaskExecutionID: taskExecID,
		},
	)
	if err != nil {
		return pluginsCore.PhaseInfoUndefined, err
	}
	info.Logs = o.TaskLogs

	var phaseInfo pluginsCore.PhaseInfo

	switch status {
	case "":
		phaseInfo = pluginsCore.PhaseInfoInitializing(occurredAt, pluginsCore.DefaultPhaseVersion, "unknown", &info)
	case daskAPI.DaskJobCreated:
		phaseInfo = pluginsCore.PhaseInfoInitializing(occurredAt, pluginsCore.DefaultPhaseVersion, "job created", &info)
	case daskAPI.DaskJobClusterCreated:
		phaseInfo = pluginsCore.PhaseInfoInitializing(occurredAt, pluginsCore.DefaultPhaseVersion, "cluster created", &info)
	case daskAPI.DaskJobFailed:
		reason := "Dask Job failed"
		phaseInfo = pluginsCore.PhaseInfoRetryableFailure(errors.DownstreamSystemError, reason, &info)
	case daskAPI.DaskJobSuccessful:
		phaseInfo = pluginsCore.PhaseInfoSuccess(&info)
	default:
		phaseInfo = pluginsCore.PhaseInfoRunning(pluginsCore.DefaultPhaseVersion, &info)
	}

	phaseVersionUpdateErr := k8s.MaybeUpdatePhaseVersionFromPluginContext(&phaseInfo, &pluginContext)
	if phaseVersionUpdateErr != nil {
		return pluginsCore.PhaseInfoUndefined, phaseVersionUpdateErr
	}

	return phaseInfo, nil
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
