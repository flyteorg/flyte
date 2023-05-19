package ray

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/plugins"
	"github.com/flyteorg/flyteplugins/go/tasks/logs"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery"
	pluginsCore "github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/flytek8s"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/k8s"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/utils"
	rayv1alpha1 "github.com/ray-project/kuberay/ray-operator/apis/ray/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"

	v1 "k8s.io/api/core/v1"

	flyteerr "github.com/flyteorg/flyteplugins/go/tasks/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	rayTaskType      = "ray"
	KindRayJob       = "RayJob"
	IncludeDashboard = "include-dashboard"
	NodeIPAddress    = "node-ip-address"
	DashboardHost    = "dashboard-host"
)

type rayJobResourceHandler struct {
}

func (rayJobResourceHandler) GetProperties() k8s.PluginProperties {
	return k8s.PluginProperties{}
}

// BuildResource Creates a new ray job resource.
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

	var container v1.Container
	found := false
	for _, c := range podSpec.Containers {
		if c.Name == primaryContainerName {
			container = c
			found = true
			break
		}
	}

	if !found {
		return nil, flyteerr.Errorf(flyteerr.BadTaskSpecification, "Unable to get primary container from the pod: [%v]", err.Error())
	}

	headReplicas := int32(1)
	headNodeRayStartParams := make(map[string]string)
	if rayJob.RayCluster.HeadGroupSpec != nil && rayJob.RayCluster.HeadGroupSpec.RayStartParams != nil {
		headNodeRayStartParams = rayJob.RayCluster.HeadGroupSpec.RayStartParams
	}
	if _, exist := headNodeRayStartParams[IncludeDashboard]; !exist {
		headNodeRayStartParams[IncludeDashboard] = strconv.FormatBool(GetConfig().IncludeDashboard)
	}
	if _, exist := headNodeRayStartParams[NodeIPAddress]; !exist {
		headNodeRayStartParams[NodeIPAddress] = GetConfig().NodeIPAddress
	}
	if _, exist := headNodeRayStartParams[DashboardHost]; !exist {
		headNodeRayStartParams[DashboardHost] = GetConfig().DashboardHost
	}

	enableIngress := true
	rayClusterSpec := rayv1alpha1.RayClusterSpec{
		HeadGroupSpec: rayv1alpha1.HeadGroupSpec{
			Template:       buildHeadPodTemplate(&container, podSpec, objectMeta, taskCtx),
			ServiceType:    v1.ServiceType(GetConfig().ServiceType),
			Replicas:       &headReplicas,
			EnableIngress:  &enableIngress,
			RayStartParams: headNodeRayStartParams,
		},
		WorkerGroupSpecs: []rayv1alpha1.WorkerGroupSpec{},
	}

	for _, spec := range rayJob.RayCluster.WorkerGroupSpec {
		workerPodTemplate := buildWorkerPodTemplate(&container, podSpec, objectMeta, taskCtx)

		minReplicas := spec.Replicas
		maxReplicas := spec.Replicas
		if spec.MinReplicas != 0 {
			minReplicas = spec.MinReplicas
		}
		if spec.MaxReplicas != 0 {
			maxReplicas = spec.MaxReplicas
		}

		workerNodeRayStartParams := make(map[string]string)
		if spec.RayStartParams != nil {
			workerNodeRayStartParams = spec.RayStartParams
		}
		if _, exist := workerNodeRayStartParams[NodeIPAddress]; !exist {
			workerNodeRayStartParams[NodeIPAddress] = GetConfig().NodeIPAddress
		}

		workerNodeSpec := rayv1alpha1.WorkerGroupSpec{
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

	rayClusterSpec.HeadGroupSpec.Template.Spec.ServiceAccountName = serviceAccountName
	for index := range rayClusterSpec.WorkerGroupSpecs {
		rayClusterSpec.WorkerGroupSpecs[index].Template.Spec.ServiceAccountName = serviceAccountName
	}

	jobSpec := rayv1alpha1.RayJobSpec{
		RayClusterSpec:           rayClusterSpec,
		Entrypoint:               strings.Join(container.Args, " "),
		ShutdownAfterJobFinishes: GetConfig().ShutdownAfterJobFinishes,
		TTLSecondsAfterFinished:  &GetConfig().TTLSecondsAfterFinished,
		RuntimeEnv:               rayJob.RuntimeEnv,
	}

	rayJobObject := rayv1alpha1.RayJob{
		TypeMeta: metav1.TypeMeta{
			Kind:       KindRayJob,
			APIVersion: rayv1alpha1.SchemeGroupVersion.String(),
		},
		Spec:       jobSpec,
		ObjectMeta: *objectMeta,
	}

	return &rayJobObject, nil
}

func buildHeadPodTemplate(container *v1.Container, podSpec *v1.PodSpec, objectMeta *metav1.ObjectMeta, taskCtx pluginsCore.TaskExecutionContext) v1.PodTemplateSpec {
	// Some configs are copy from  https://github.com/ray-project/kuberay/blob/b72e6bdcd9b8c77a9dc6b5da8560910f3a0c3ffd/apiserver/pkg/util/cluster.go#L97
	// They should always be the same, so we could hard code here.
	primaryContainer := container.DeepCopy()
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

	headPodSpec := podSpec.DeepCopy()

	headPodSpec.Containers = []v1.Container{*primaryContainer}

	podTemplateSpec := v1.PodTemplateSpec{
		Spec:       *headPodSpec,
		ObjectMeta: *objectMeta,
	}
	podTemplateSpec.SetLabels(utils.UnionMaps(podTemplateSpec.GetLabels(), utils.CopyMap(taskCtx.TaskExecutionMetadata().GetLabels())))
	podTemplateSpec.SetAnnotations(utils.UnionMaps(podTemplateSpec.GetAnnotations(), utils.CopyMap(taskCtx.TaskExecutionMetadata().GetAnnotations())))
	return podTemplateSpec
}

func buildWorkerPodTemplate(container *v1.Container, podSpec *v1.PodSpec, objectMetadata *metav1.ObjectMeta, taskCtx pluginsCore.TaskExecutionContext) v1.PodTemplateSpec {
	// Some configs are copy from  https://github.com/ray-project/kuberay/blob/b72e6bdcd9b8c77a9dc6b5da8560910f3a0c3ffd/apiserver/pkg/util/cluster.go#L185
	// They should always be the same, so we could hard code here.
	initContainers := []v1.Container{
		{
			Name:  "init-myservice",
			Image: "busybox:1.28",
			Command: []string{
				"sh",
				"-c",
				"until nslookup $RAY_IP.$(cat /var/run/secrets/kubernetes.io/serviceaccount/namespace).svc.cluster.local; do echo waiting for myservice; sleep 2; done",
			},
			Resources: container.Resources,
		},
	}
	primaryContainer := container.DeepCopy()
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

	workerPodSpec := podSpec.DeepCopy()
	workerPodSpec.Containers = []v1.Container{*primaryContainer}
	workerPodSpec.InitContainers = initContainers

	podTemplateSpec := v1.PodTemplateSpec{
		Spec:       *workerPodSpec,
		ObjectMeta: *objectMetadata,
	}
	podTemplateSpec.SetLabels(utils.UnionMaps(podTemplateSpec.GetLabels(), utils.CopyMap(taskCtx.TaskExecutionMetadata().GetLabels())))
	podTemplateSpec.SetAnnotations(utils.UnionMaps(podTemplateSpec.GetAnnotations(), utils.CopyMap(taskCtx.TaskExecutionMetadata().GetAnnotations())))
	return podTemplateSpec
}

func (rayJobResourceHandler) BuildIdentityResource(ctx context.Context, taskCtx pluginsCore.TaskExecutionMetadata) (client.Object, error) {
	return &rayv1alpha1.RayJob{
		TypeMeta: metav1.TypeMeta{
			Kind:       KindRayJob,
			APIVersion: rayv1alpha1.SchemeGroupVersion.String(),
		},
	}, nil
}

func getEventInfoForRayJob() (*pluginsCore.TaskInfo, error) {
	taskLogs := make([]*core.TaskLog, 0, 3)
	logPlugin, err := logs.InitializeLogPlugins(logs.GetLogConfig())

	if err != nil {
		return nil, err
	}

	if logPlugin == nil {
		return nil, nil
	}

	// TODO: Retrieve the name of head pod from rayJob.status, and add it to task logs
	// RayJob CRD does not include the name of the worker or head pod for now

	// TODO: Add ray Dashboard URI to task logs

	return &pluginsCore.TaskInfo{
		Logs: taskLogs,
	}, nil
}

func (rayJobResourceHandler) GetTaskPhase(ctx context.Context, pluginContext k8s.PluginContext, resource client.Object) (pluginsCore.PhaseInfo, error) {
	rayJob := resource.(*rayv1alpha1.RayJob)
	info, err := getEventInfoForRayJob()
	if err != nil {
		return pluginsCore.PhaseInfoUndefined, err
	}
	switch rayJob.Status.JobStatus {
	case rayv1alpha1.JobStatusPending:
		return pluginsCore.PhaseInfoNotReady(time.Now(), pluginsCore.DefaultPhaseVersion, "job is pending"), nil
	case rayv1alpha1.JobStatusFailed:
		reason := fmt.Sprintf("Failed to create Ray job: %s", rayJob.Name)
		return pluginsCore.PhaseInfoFailure(flyteerr.TaskFailedWithError, reason, info), nil
	case rayv1alpha1.JobStatusSucceeded:
		return pluginsCore.PhaseInfoSuccess(info), nil
	case rayv1alpha1.JobStatusRunning:
		return pluginsCore.PhaseInfoRunning(pluginsCore.DefaultPhaseVersion, info), nil
	}
	return pluginsCore.PhaseInfoQueued(time.Now(), pluginsCore.DefaultPhaseVersion, "JobCreated"), nil
}

func init() {
	if err := rayv1alpha1.AddToScheme(scheme.Scheme); err != nil {
		panic(err)
	}

	pluginmachinery.PluginRegistry().RegisterK8sPlugin(
		k8s.PluginEntry{
			ID:                  rayTaskType,
			RegisteredTaskTypes: []pluginsCore.TaskType{rayTaskType},
			ResourceToWatch:     &rayv1alpha1.RayJob{},
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
