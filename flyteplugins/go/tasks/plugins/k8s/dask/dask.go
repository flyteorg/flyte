package dask

import (
	"context"
	"fmt"
	"strings"
	"time"

	daskAPI "github.com/dask/dask-kubernetes/v2023/dask_kubernetes/operator/go_client/pkg/apis/kubernetes.dask.org/v1"
	"github.com/samber/lo"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/flyteorg/flyte/flyteplugins/go/tasks/errors"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/logs"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery"
	pluginsCore "github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/flytek8s"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/k8s"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/tasklog"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/utils"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/plugins/k8s/pod"
	"github.com/flyteorg/flyte/flytestdlib/logger"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/core"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/plugins"
)

const (
	daskTaskType                             = "dask"
	KindDaskJob                              = "DaskJob"
	defaultDaskJobRunnerPrimaryContainerName = "job-runner"
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

	workerSpec, err := createWorkerSpec(*daskJob.Workers, podSpec, primaryContainerName, taskCtx.TaskExecutionMetadata())
	if err != nil {
		return nil, err
	}

	clusterName := taskCtx.TaskExecutionMetadata().GetTaskExecutionID().GetGeneratedName()
	schedulerSpec, err := createSchedulerSpec(*daskJob.Scheduler, clusterName, nonInterruptiblePodSpec, primaryContainerName, taskCtx.TaskExecutionMetadata())
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

func createWorkerSpec(cluster plugins.DaskWorkerGroup, podSpec *v1.PodSpec, primaryContainerName string,
	teMetadata pluginsCore.TaskExecutionMetadata) (*daskAPI.WorkerSpec, error) {
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
	if len(clusterResources.GetRequests()) >= 1 || len(clusterResources.GetLimits()) >= 1 {
		resources, err = flytek8s.ToK8sResourceRequirements(cluster.GetResources())
		if err != nil {
			return nil, err
		}

		*resources = flytek8s.ApplyK8sResourceOverrides(teMetadata, resources)
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

func createSchedulerSpec(scheduler plugins.DaskScheduler, clusterName string, podSpec *v1.PodSpec, primaryContainerName string,
	teMetadata pluginsCore.TaskExecutionMetadata) (*daskAPI.SchedulerSpec, error) {
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
	if len(schedulerResources.GetRequests()) >= 1 || len(schedulerResources.GetLimits()) >= 1 {
		resources, err = flytek8s.ToK8sResourceRequirements(scheduler.GetResources())
		if err != nil {
			return nil, err
		}

		*resources = flytek8s.ApplyK8sResourceOverrides(teMetadata, resources)
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

	if primaryContainer.ReadinessProbe == nil {
		primaryContainer.ReadinessProbe = &v1.Probe{
			ProbeHandler: v1.ProbeHandler{
				HTTPGet: &v1.HTTPGetAction{
					Port: intstr.FromString("dashboard"),
				},
			},
			InitialDelaySeconds: 5,
			PeriodSeconds:       5,
			FailureThreshold:    30,
		}
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
	logPlugin, err := logs.InitializeLogPlugins(&GetConfig().Logs)
	if err != nil {
		return pluginsCore.PhaseInfoUndefined, err
	}
	job := r.(*daskAPI.DaskJob)
	status := job.Status.JobStatus
	occurredAt := time.Now()

	info := pluginsCore.TaskInfo{
		OccurredAt: &occurredAt,
	}

	taskExecID := pluginContext.TaskExecutionMetadata().GetTaskExecutionID()
	schedulerPodName, err := getDaskSchedulerPodName(ctx, job.Name, pluginContext)
	if err != nil {
		logger.Debug(ctx, "Failed to get dask scheduler pod name. Error: %v", err)
	}
	var enableVscode bool
	if len(job.Spec.Cluster.Spec.Scheduler.Spec.Containers) > 0 {
		enableVscode = flytek8s.IsVscodeEnabled(ctx, job.Spec.Cluster.Spec.Scheduler.Spec.Containers[0].Env)
	}
	input := tasklog.Input{
		Namespace:       job.ObjectMeta.Namespace,
		PodName:         job.Status.JobRunnerPodName,
		TaskExecutionID: taskExecID,
		EnableVscode:    enableVscode,
	}
	input.ExtraTemplateVars = append(
		input.ExtraTemplateVars,
		tasklog.TemplateVar{
			Regex: tasklog.MustCreateRegex("daskSchedulerName"),
			Value: schedulerPodName,
		},
	)

	o, err := logPlugin.GetTaskLogs(input)
	if err != nil {
		return pluginsCore.PhaseInfoUndefined, err
	}
	info.Logs = o.TaskLogs
	info.LogContext = &core.LogContext{
		PrimaryPodName: job.Status.JobRunnerPodName,
		Pods: []*core.PodLogContext{
			{
				Namespace:            job.ObjectMeta.Namespace,
				PodName:              job.Status.JobRunnerPodName,
				PrimaryContainerName: defaultDaskJobRunnerPrimaryContainerName,
				Containers: []*core.ContainerContext{
					{ContainerName: defaultDaskJobRunnerPrimaryContainerName},
				},
			},
		},
	}

	phaseInfo, err := flytek8s.DemystifyFailedOrPendingPod(ctx, pluginContext, info, job.ObjectMeta.Namespace, job.Status.JobRunnerPodName, defaultDaskJobRunnerPrimaryContainerName)
	if err != nil {
		logger.Errorf(ctx, "Failed to demystify pod status for dask job-runner. Error: %v", err)
	}
	if phaseInfo.Phase().IsFailure() {
		// If the job-runner pod is in a failure state, we can fail fast without checking the DaskJob status.
		return phaseInfo, nil
	}
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

	ready, err := isDaskSchedulerReady(ctx, job.Name, pluginContext)
	if err != nil {
		logger.Warnf(ctx, "Failed to determine Dask dashboard readiness. Error: %v", err)
	} else {
		for _, tl := range info.Logs {
			if tl != nil && tl.LinkType == core.TaskLog_DASHBOARD {
				tl.Ready = ready
				if !ready || phaseInfo.Phase() != pluginsCore.PhaseRunning {
					phaseInfo.WithReason("Dask dashboard is not ready")
				} else {
					phaseInfo.WithReason("Dask dashboard is ready")
				}
			} else if tl != nil && tl.LinkType == core.TaskLog_IDE {
				tl.Ready = ready
				if !ready || phaseInfo.Phase() != pluginsCore.PhaseRunning {
					phaseInfo.WithReason("Vscode server is not ready")
				} else {
					phaseInfo.WithReason("Vscode server is ready")
				}
			}
		}
	}

	phaseVersionUpdateErr := k8s.MaybeUpdatePhaseVersionFromPluginContext(&phaseInfo, &pluginContext)
	if phaseVersionUpdateErr != nil {
		return phaseInfo, phaseVersionUpdateErr
	}
	return phaseInfo, nil
}

func getDaskSchedulerPodName(ctx context.Context, daskJobName string, pluginContext k8s.PluginContext) (string, error) {
	podList := &v1.PodList{}
	err := pluginContext.K8sReader().List(ctx, podList)
	if err != nil {
		return "", fmt.Errorf("failed to list dask execution pods. Error: %w", err)
	}
	pods := lo.Filter(podList.Items, func(pod v1.Pod, _ int) bool {
		return strings.HasPrefix(pod.Name, daskJobName) && strings.Contains(pod.Name, "scheduler") && flytek8s.GetPrimaryContainerName(&pod) == "scheduler"
	})
	if len(pods) == 0 {
		return "", fmt.Errorf("no dask scheduler pod found for dask job [%v]", daskJobName)
	}
	return pods[0].Name, nil
}

func isDaskSchedulerReady(ctx context.Context, daskJobName string, pluginContext k8s.PluginContext) (bool, error) {
	podList := &v1.PodList{}
	err := pluginContext.K8sReader().List(ctx, podList)
	if err != nil {
		return false, fmt.Errorf("failed to list dask execution pods. Error: %w", err)
	}
	pods := lo.Filter(podList.Items, func(p v1.Pod, _ int) bool {
		return strings.HasPrefix(p.Name, daskJobName) && strings.Contains(p.Name, "scheduler")
	})
	if len(pods) == 0 {
		return false, nil
	} else if len(pods) == 1 {
		return pod.IsPodReady(&pods[0]), nil
	}

	// More than one dask scheduler pod. Should not happen.
	logger.Debug(ctx, "Cannot determine dask scheduler readiness as more than one dask scheduler pod found")
	return false, fmt.Errorf("more than one dask scheduler pod found for dask job [%v]", daskJobName)
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
