package ray

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	rayv1 "github.com/ray-project/kuberay/ray-operator/apis/ray/v1"
	"gopkg.in/yaml.v2"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/plugins"
	flyteerr "github.com/flyteorg/flyte/flyteplugins/go/tasks/errors"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/logs"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery"
	pluginsCore "github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/flytek8s"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/flytek8s/config"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/k8s"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/tasklog"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/utils"
)

const (
	rayStateMountPath                  = "/tmp/ray"
	defaultRayStateVolName             = "system-ray-state"
	rayTaskType                        = "ray"
	KindRayJob                         = "RayJob"
	IncludeDashboard                   = "include-dashboard"
	NodeIPAddress                      = "node-ip-address"
	DashboardHost                      = "dashboard-host"
	DisableUsageStatsStartParameter    = "disable-usage-stats"
	DisableUsageStatsStartParameterVal = "true"
)

var logTemplateRegexes = struct {
	RayClusterName *regexp.Regexp
	RayJobID       *regexp.Regexp
}{
	tasklog.MustCreateRegex("rayClusterName"),
	tasklog.MustCreateRegex("rayJobID"),
}

type rayJobResourceHandler struct{}

func (rayJobResourceHandler) GetProperties() k8s.PluginProperties {
	return k8s.PluginProperties{}
}

// BuildResource Creates a new ray job resource
func (rayJobResourceHandler) BuildResource(ctx context.Context, taskCtx pluginsCore.TaskExecutionContext) (client.Object, error) {
	taskTemplate, err := taskCtx.TaskReader().Read(ctx)
	if err != nil {
		return nil, flyteerr.Errorf(flyteerr.BadTaskSpecification, "unable to fetch task specification [%v]", err.Error())
	} else if taskTemplate == nil {
		return nil, flyteerr.Errorf(flyteerr.BadTaskSpecification, "nil task specification")
	}

	rayJob := plugins.RayJob{}
	err = utils.UnmarshalStruct(taskTemplate.GetCustom(), &rayJob)
	if err != nil {
		return nil, flyteerr.Errorf(flyteerr.BadTaskSpecification, "invalid TaskSpecification [%v], Err: [%v]", taskTemplate.GetCustom(), err.Error())
	}

	podSpec, objectMeta, primaryContainerName, err := flytek8s.ToK8sPodSpec(ctx, taskCtx)
	if err != nil {
		return nil, flyteerr.Errorf(flyteerr.BadTaskSpecification, "Unable to create pod spec: [%v]", err.Error())
	}

	var primaryContainer *v1.Container
	var primaryContainerIdx int
	for idx, c := range podSpec.Containers {
		if c.Name == primaryContainerName {
			c := c
			primaryContainer = &c
			primaryContainerIdx = idx
			break
		}
	}

	if primaryContainer == nil {
		return nil, flyteerr.Errorf(flyteerr.BadTaskSpecification, "Unable to get primary container from the pod: [%v]", err.Error())
	}

	cfg := GetConfig()

	headNodeRayStartParams := make(map[string]string)
	if rayJob.RayCluster.HeadGroupSpec != nil && rayJob.RayCluster.HeadGroupSpec.RayStartParams != nil {
		headNodeRayStartParams = rayJob.RayCluster.HeadGroupSpec.RayStartParams
	} else if headNode := cfg.Defaults.HeadNode; len(headNode.StartParameters) > 0 {
		headNodeRayStartParams = headNode.StartParameters
	}

	if _, exist := headNodeRayStartParams[IncludeDashboard]; !exist {
		headNodeRayStartParams[IncludeDashboard] = strconv.FormatBool(GetConfig().IncludeDashboard)
	}

	if _, exist := headNodeRayStartParams[NodeIPAddress]; !exist {
		headNodeRayStartParams[NodeIPAddress] = cfg.Defaults.HeadNode.IPAddress
	}

	if _, exist := headNodeRayStartParams[DashboardHost]; !exist {
		headNodeRayStartParams[DashboardHost] = cfg.DashboardHost
	}

	if _, exists := headNodeRayStartParams[DisableUsageStatsStartParameter]; !exists && !cfg.EnableUsageStats {
		headNodeRayStartParams[DisableUsageStatsStartParameter] = DisableUsageStatsStartParameterVal
	}

	podSpec.ServiceAccountName = cfg.ServiceAccount

	headPodSpec := podSpec.DeepCopy()

	rayjob, err := constructRayJob(taskCtx, rayJob, objectMeta, *podSpec, headPodSpec, headNodeRayStartParams, primaryContainerIdx, *primaryContainer)

	return rayjob, err
}

func constructRayJob(taskCtx pluginsCore.TaskExecutionContext, rayJob plugins.RayJob, objectMeta *metav1.ObjectMeta, podSpec v1.PodSpec, headPodSpec *v1.PodSpec, headNodeRayStartParams map[string]string, primaryContainerIdx int, primaryContainer v1.Container) (*rayv1.RayJob, error) {
	enableIngress := true
	cfg := GetConfig()
	rayClusterSpec := rayv1.RayClusterSpec{
		HeadGroupSpec: rayv1.HeadGroupSpec{
			Template: buildHeadPodTemplate(
				&headPodSpec.Containers[primaryContainerIdx],
				headPodSpec,
				objectMeta,
				taskCtx,
			),
			ServiceType:    v1.ServiceType(cfg.ServiceType),
			EnableIngress:  &enableIngress,
			RayStartParams: headNodeRayStartParams,
		},
		WorkerGroupSpecs:        []rayv1.WorkerGroupSpec{},
		EnableInTreeAutoscaling: &rayJob.RayCluster.EnableAutoscaling,
	}

	for _, spec := range rayJob.RayCluster.WorkerGroupSpec {
		workerPodSpec := podSpec.DeepCopy()
		workerPodTemplate := buildWorkerPodTemplate(
			&workerPodSpec.Containers[primaryContainerIdx],
			workerPodSpec,
			objectMeta,
			taskCtx,
		)

		workerNodeRayStartParams := make(map[string]string)
		if spec.RayStartParams != nil {
			workerNodeRayStartParams = spec.RayStartParams
		} else if workerNode := cfg.Defaults.WorkerNode; len(workerNode.StartParameters) > 0 {
			workerNodeRayStartParams = workerNode.StartParameters
		}

		if _, exist := workerNodeRayStartParams[NodeIPAddress]; !exist {
			workerNodeRayStartParams[NodeIPAddress] = cfg.Defaults.WorkerNode.IPAddress
		}

		if _, exists := workerNodeRayStartParams[DisableUsageStatsStartParameter]; !exists && !cfg.EnableUsageStats {
			workerNodeRayStartParams[DisableUsageStatsStartParameter] = DisableUsageStatsStartParameterVal
		}

		minReplicas := spec.MinReplicas
		if minReplicas > spec.Replicas {
			minReplicas = spec.Replicas
		}
		maxReplicas := spec.MaxReplicas
		if maxReplicas < spec.Replicas {
			maxReplicas = spec.Replicas
		}

		workerNodeSpec := rayv1.WorkerGroupSpec{
			GroupName:      spec.GroupName,
			MinReplicas:    &minReplicas,
			MaxReplicas:    &maxReplicas,
			Replicas:       &spec.Replicas,
			RayStartParams: workerNodeRayStartParams,
			Template:       workerPodTemplate,
		}

		rayClusterSpec.WorkerGroupSpecs = append(rayClusterSpec.WorkerGroupSpecs, workerNodeSpec)
	}

	serviceAccountName := flytek8s.GetServiceAccountNameFromTaskExecutionMetadata(taskCtx.TaskExecutionMetadata())
	if len(serviceAccountName) == 0 {
		serviceAccountName = cfg.ServiceAccount
	}

	rayClusterSpec.HeadGroupSpec.Template.Spec.ServiceAccountName = serviceAccountName
	for index := range rayClusterSpec.WorkerGroupSpecs {
		rayClusterSpec.WorkerGroupSpecs[index].Template.Spec.ServiceAccountName = serviceAccountName
	}

	shutdownAfterJobFinishes := cfg.ShutdownAfterJobFinishes
	ttlSecondsAfterFinished := &cfg.TTLSecondsAfterFinished
	if rayJob.ShutdownAfterJobFinishes {
		shutdownAfterJobFinishes = true
		ttlSecondsAfterFinished = &rayJob.TtlSecondsAfterFinished
	}

	submitterPodTemplate := buildSubmitterPodTemplate(headPodSpec, objectMeta, taskCtx)

	// TODO: This is for backward compatibility. Remove this block once runtime_env is removed from ray proto.
	var err error
	var runtimeEnvYaml string
	runtimeEnvYaml = rayJob.RuntimeEnvYaml
	// If runtime_env exists but runtime_env_yaml does not, convert runtime_env to runtime_env_yaml
	if rayJob.RuntimeEnv != "" && rayJob.RuntimeEnvYaml == "" {
		runtimeEnvYaml, err = convertBase64RuntimeEnvToYaml(rayJob.RuntimeEnv)
		if err != nil {
			return nil, err
		}
	}

	jobSpec := rayv1.RayJobSpec{
		RayClusterSpec:           &rayClusterSpec,
		Entrypoint:               strings.Join(primaryContainer.Args, " "),
		ShutdownAfterJobFinishes: shutdownAfterJobFinishes,
		TTLSecondsAfterFinished:  *ttlSecondsAfterFinished,
		RuntimeEnvYAML:           runtimeEnvYaml,
		SubmitterPodTemplate:     &submitterPodTemplate,
	}

	return &rayv1.RayJob{
		TypeMeta: metav1.TypeMeta{
			Kind:       KindRayJob,
			APIVersion: rayv1.SchemeGroupVersion.String(),
		},
		Spec:       jobSpec,
		ObjectMeta: *objectMeta,
	}, nil
}

func convertBase64RuntimeEnvToYaml(s string) (string, error) {
	// Decode from base64
	data, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return "", err
	}

	// Unmarshal JSON
	var obj map[string]interface{}
	err = json.Unmarshal(data, &obj)
	if err != nil {
		return "", err
	}

	// Convert to YAML
	y, err := yaml.Marshal(&obj)
	if err != nil {
		return "", err
	}

	return string(y), nil
}

func injectLogsSidecar(primaryContainer *v1.Container, podSpec *v1.PodSpec) {
	cfg := GetConfig()
	if cfg.LogsSidecar == nil {
		return
	}
	sidecar := cfg.LogsSidecar.DeepCopy()

	// Ray logs integration
	var rayStateVolMount *v1.VolumeMount
	// Look for an existing volume mount on the primary container, mounted at /tmp/ray
	for _, vm := range primaryContainer.VolumeMounts {
		if vm.MountPath == rayStateMountPath {
			vm := vm
			rayStateVolMount = &vm
			break
		}
	}
	// No existing volume mount exists at /tmp/ray. We create a new volume and volume
	// mount and add it to the pod and container specs respectively
	if rayStateVolMount == nil {
		vol := v1.Volume{
			Name: defaultRayStateVolName,
			VolumeSource: v1.VolumeSource{
				EmptyDir: &v1.EmptyDirVolumeSource{},
			},
		}
		podSpec.Volumes = append(podSpec.Volumes, vol)
		volMount := v1.VolumeMount{
			Name:      defaultRayStateVolName,
			MountPath: rayStateMountPath,
		}
		primaryContainer.VolumeMounts = append(primaryContainer.VolumeMounts, volMount)
		rayStateVolMount = &volMount
	}
	// We need to mirror the ray state volume mount into the sidecar as readonly,
	// so that we can read the logs written by the head node.
	readOnlyRayStateVolMount := *rayStateVolMount.DeepCopy()
	readOnlyRayStateVolMount.ReadOnly = true

	// Update volume mounts on sidecar
	// If one already exists with the desired mount path, simply replace it. Otherwise,
	// add it to sidecar's volume mounts.
	foundExistingSidecarVolMount := false
	for idx, vm := range sidecar.VolumeMounts {
		if vm.MountPath == rayStateMountPath {
			foundExistingSidecarVolMount = true
			sidecar.VolumeMounts[idx] = readOnlyRayStateVolMount
		}
	}
	if !foundExistingSidecarVolMount {
		sidecar.VolumeMounts = append(sidecar.VolumeMounts, readOnlyRayStateVolMount)
	}

	// Add sidecar to containers
	podSpec.Containers = append(podSpec.Containers, *sidecar)
}

func buildHeadPodTemplate(primaryContainer *v1.Container, podSpec *v1.PodSpec, objectMeta *metav1.ObjectMeta, taskCtx pluginsCore.TaskExecutionContext) v1.PodTemplateSpec {
	// Some configs are copy from  https://github.com/ray-project/kuberay/blob/b72e6bdcd9b8c77a9dc6b5da8560910f3a0c3ffd/apiserver/pkg/util/cluster.go#L97
	// They should always be the same, so we could hard code here.
	primaryContainer.Name = "ray-head"

	envs := []v1.EnvVar{
		{
			Name: "MY_POD_IP",
			ValueFrom: &v1.EnvVarSource{
				FieldRef: &v1.ObjectFieldSelector{
					FieldPath: "status.podIP",
				},
			},
		},
	}

	primaryContainer.Args = []string{}

	primaryContainer.Env = append(primaryContainer.Env, envs...)

	ports := []v1.ContainerPort{
		{
			Name:          "redis",
			ContainerPort: 6379,
		},
		{
			Name:          "head",
			ContainerPort: 10001,
		},
		{
			Name:          "dashboard",
			ContainerPort: 8265,
		},
	}

	primaryContainer.Ports = append(primaryContainer.Ports, ports...)

	// Inject a sidecar for capturing and exposing Ray job logs
	injectLogsSidecar(primaryContainer, podSpec)

	podTemplateSpec := v1.PodTemplateSpec{
		Spec:       *podSpec,
		ObjectMeta: *objectMeta,
	}
	cfg := config.GetK8sPluginConfig()
	podTemplateSpec.SetLabels(utils.UnionMaps(cfg.DefaultLabels, podTemplateSpec.GetLabels(), utils.CopyMap(taskCtx.TaskExecutionMetadata().GetLabels())))
	podTemplateSpec.SetAnnotations(utils.UnionMaps(cfg.DefaultAnnotations, podTemplateSpec.GetAnnotations(), utils.CopyMap(taskCtx.TaskExecutionMetadata().GetAnnotations())))
	return podTemplateSpec
}

func buildSubmitterPodTemplate(podSpec *v1.PodSpec, objectMeta *metav1.ObjectMeta, taskCtx pluginsCore.TaskExecutionContext) v1.PodTemplateSpec {
	podTemplateSpec := v1.PodTemplateSpec{
		Spec:       *podSpec,
		ObjectMeta: *objectMeta,
	}
	cfg := config.GetK8sPluginConfig()
	podTemplateSpec.SetLabels(utils.UnionMaps(cfg.DefaultLabels, podTemplateSpec.GetLabels(), utils.CopyMap(taskCtx.TaskExecutionMetadata().GetLabels())))
	podTemplateSpec.SetAnnotations(utils.UnionMaps(cfg.DefaultAnnotations, podTemplateSpec.GetAnnotations(), utils.CopyMap(taskCtx.TaskExecutionMetadata().GetAnnotations())))
	return podTemplateSpec
}

func buildWorkerPodTemplate(primaryContainer *v1.Container, podSpec *v1.PodSpec, objectMetadata *metav1.ObjectMeta, taskCtx pluginsCore.TaskExecutionContext) v1.PodTemplateSpec {
	// Some configs are copy from  https://github.com/ray-project/kuberay/blob/b72e6bdcd9b8c77a9dc6b5da8560910f3a0c3ffd/apiserver/pkg/util/cluster.go#L185
	// They should always be the same, so we could hard code here.

	primaryContainer.Name = "ray-worker"

	primaryContainer.Args = []string{}

	envs := []v1.EnvVar{
		{
			Name:  "RAY_DISABLE_DOCKER_CPU_WARNING",
			Value: "1",
		},
		{
			Name:  "TYPE",
			Value: "worker",
		},
		{
			Name: "CPU_REQUEST",
			ValueFrom: &v1.EnvVarSource{
				ResourceFieldRef: &v1.ResourceFieldSelector{
					ContainerName: "ray-worker",
					Resource:      "requests.cpu",
				},
			},
		},
		{
			Name: "CPU_LIMITS",
			ValueFrom: &v1.EnvVarSource{
				ResourceFieldRef: &v1.ResourceFieldSelector{
					ContainerName: "ray-worker",
					Resource:      "limits.cpu",
				},
			},
		},
		{
			Name: "MEMORY_REQUESTS",
			ValueFrom: &v1.EnvVarSource{
				ResourceFieldRef: &v1.ResourceFieldSelector{
					ContainerName: "ray-worker",
					Resource:      "requests.cpu",
				},
			},
		},
		{
			Name: "MEMORY_LIMITS",
			ValueFrom: &v1.EnvVarSource{
				ResourceFieldRef: &v1.ResourceFieldSelector{
					ContainerName: "ray-worker",
					Resource:      "limits.cpu",
				},
			},
		},
		{
			Name: "MY_POD_NAME",
			ValueFrom: &v1.EnvVarSource{
				FieldRef: &v1.ObjectFieldSelector{
					FieldPath: "metadata.name",
				},
			},
		},
		{
			Name: "MY_POD_IP",
			ValueFrom: &v1.EnvVarSource{
				FieldRef: &v1.ObjectFieldSelector{
					FieldPath: "status.podIP",
				},
			},
		},
	}

	primaryContainer.Env = append(primaryContainer.Env, envs...)

	primaryContainer.Lifecycle = &v1.Lifecycle{
		PreStop: &v1.LifecycleHandler{
			Exec: &v1.ExecAction{
				Command: []string{
					"/bin/sh", "-c", "ray stop",
				},
			},
		},
	}

	ports := []v1.ContainerPort{
		{
			Name:          "redis",
			ContainerPort: 6379,
		},
		{
			Name:          "head",
			ContainerPort: 10001,
		},
		{
			Name:          "dashboard",
			ContainerPort: 8265,
		},
	}
	primaryContainer.Ports = append(primaryContainer.Ports, ports...)

	podTemplateSpec := v1.PodTemplateSpec{
		Spec:       *podSpec,
		ObjectMeta: *objectMetadata,
	}
	podTemplateSpec.SetLabels(utils.UnionMaps(podTemplateSpec.GetLabels(), utils.CopyMap(taskCtx.TaskExecutionMetadata().GetLabels())))
	podTemplateSpec.SetAnnotations(utils.UnionMaps(podTemplateSpec.GetAnnotations(), utils.CopyMap(taskCtx.TaskExecutionMetadata().GetAnnotations())))
	return podTemplateSpec
}

func (rayJobResourceHandler) BuildIdentityResource(ctx context.Context, taskCtx pluginsCore.TaskExecutionMetadata) (client.Object, error) {
	return &rayv1.RayJob{
		TypeMeta: metav1.TypeMeta{
			Kind:       KindRayJob,
			APIVersion: rayv1.SchemeGroupVersion.String(),
		},
	}, nil
}

func getEventInfoForRayJob(logConfig logs.LogConfig, pluginContext k8s.PluginContext, rayJob *rayv1.RayJob) (*pluginsCore.TaskInfo, error) {
	logPlugin, err := logs.InitializeLogPlugins(&logConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize log plugins. Error: %w", err)
	}

	var taskLogs []*core.TaskLog

	taskExecID := pluginContext.TaskExecutionMetadata().GetTaskExecutionID()
	input := tasklog.Input{
		Namespace:         rayJob.Namespace,
		TaskExecutionID:   taskExecID,
		ExtraTemplateVars: []tasklog.TemplateVar{},
	}
	if rayJob.Status.JobId != "" {
		input.ExtraTemplateVars = append(
			input.ExtraTemplateVars,
			tasklog.TemplateVar{
				Regex: logTemplateRegexes.RayJobID,
				Value: rayJob.Status.JobId,
			},
		)
	}
	if rayJob.Status.RayClusterName != "" {
		input.ExtraTemplateVars = append(
			input.ExtraTemplateVars,
			tasklog.TemplateVar{
				Regex: logTemplateRegexes.RayClusterName,
				Value: rayJob.Status.RayClusterName,
			},
		)
	}

	// TODO: Retrieve the name of head pod from rayJob.status, and add it to task logs
	// RayJob CRD does not include the name of the worker or head pod for now
	logOutput, err := logPlugin.GetTaskLogs(input)
	if err != nil {
		return nil, fmt.Errorf("failed to generate task logs. Error: %w", err)
	}
	taskLogs = append(taskLogs, logOutput.TaskLogs...)

	// Handling for Ray Dashboard
	dashboardURLTemplate := GetConfig().DashboardURLTemplate
	if dashboardURLTemplate != nil &&
		rayJob.Status.DashboardURL != "" &&
		rayJob.Status.JobStatus == rayv1.JobStatusRunning {
		dashboardURLOutput, err := dashboardURLTemplate.GetTaskLogs(input)
		if err != nil {
			return nil, fmt.Errorf("failed to generate Ray dashboard link. Error: %w", err)
		}
		taskLogs = append(taskLogs, dashboardURLOutput.TaskLogs...)
	}

	return &pluginsCore.TaskInfo{Logs: taskLogs}, nil
}

func (plugin rayJobResourceHandler) GetTaskPhase(ctx context.Context, pluginContext k8s.PluginContext, resource client.Object) (pluginsCore.PhaseInfo, error) {
	rayJob := resource.(*rayv1.RayJob)
	info, err := getEventInfoForRayJob(GetConfig().Logs, pluginContext, rayJob)
	if err != nil {
		return pluginsCore.PhaseInfoUndefined, err
	}

	if len(rayJob.Status.JobDeploymentStatus) == 0 {
		return pluginsCore.PhaseInfoQueuedWithTaskInfo(time.Now(), pluginsCore.DefaultPhaseVersion, "Scheduling", info), nil
	}

	var phaseInfo pluginsCore.PhaseInfo

	// KubeRay creates a Ray cluster first, and then submits a Ray job to the cluster
	switch rayJob.Status.JobDeploymentStatus {
	case rayv1.JobDeploymentStatusInitializing:
		phaseInfo, err = pluginsCore.PhaseInfoInitializing(rayJob.CreationTimestamp.Time, pluginsCore.DefaultPhaseVersion, "cluster is creating", info), nil
	case rayv1.JobDeploymentStatusRunning:
		phaseInfo, err = pluginsCore.PhaseInfoRunning(pluginsCore.DefaultPhaseVersion, info), nil
	case rayv1.JobDeploymentStatusComplete:
		phaseInfo, err = pluginsCore.PhaseInfoSuccess(info), nil
	case rayv1.JobDeploymentStatusFailed:
		failInfo := fmt.Sprintf("Failed to run Ray job %s with error: [%s] %s", rayJob.Name, rayJob.Status.Reason, rayJob.Status.Message)
		phaseInfo, err = pluginsCore.PhaseInfoFailure(flyteerr.TaskFailedWithError, failInfo, info), nil
	default:
		// We already handle all known deployment status, so this should never happen unless a future version of ray
		// introduced a new job status.
		phaseInfo, err = pluginsCore.PhaseInfoUndefined, fmt.Errorf("unknown job deployment status: %s", rayJob.Status.JobDeploymentStatus)
	}

	phaseVersionUpdateErr := k8s.MaybeUpdatePhaseVersionFromPluginContext(&phaseInfo, &pluginContext)
	if phaseVersionUpdateErr != nil {
		return phaseInfo, phaseVersionUpdateErr
	}

	return phaseInfo, err
}

func init() {
	if err := rayv1.AddToScheme(scheme.Scheme); err != nil {
		panic(err)
	}

	pluginmachinery.PluginRegistry().RegisterK8sPlugin(
		k8s.PluginEntry{
			ID:                  rayTaskType,
			RegisteredTaskTypes: []pluginsCore.TaskType{rayTaskType},
			ResourceToWatch:     &rayv1.RayJob{},
			Plugin:              rayJobResourceHandler{},
			IsDefault:           false,
			CustomKubeClient: func(ctx context.Context) (pluginsCore.KubeClient, error) {
				remoteConfig := GetConfig().RemoteClusterConfig
				if !remoteConfig.Enabled {
					// use controller-runtime KubeClient
					return nil, nil
				}

				kubeConfig, err := k8s.KubeClientConfig(remoteConfig.Endpoint, remoteConfig.Auth)
				if err != nil {
					return nil, err
				}

				return k8s.NewDefaultKubeClient(kubeConfig)
			},
		})
}
