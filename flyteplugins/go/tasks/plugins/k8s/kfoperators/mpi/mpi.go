package mpi

import (
	"context"
	"fmt"
	"time"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/plugins"
	flyteerr "github.com/flyteorg/flyteplugins/go/tasks/errors"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery"
	pluginsCore "github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/flytek8s"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/k8s"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/utils"
	"github.com/flyteorg/flyteplugins/go/tasks/plugins/k8s/kfoperators/common"
	commonKf "github.com/kubeflow/common/pkg/apis/common/v1"
	kubeflowv1 "github.com/kubeflow/training-operator/pkg/apis/kubeflow.org/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type mpiOperatorResourceHandler struct {
}

// Sanity test that the plugin implements method of k8s.Plugin
var _ k8s.Plugin = mpiOperatorResourceHandler{}

func (mpiOperatorResourceHandler) GetProperties() k8s.PluginProperties {
	return k8s.PluginProperties{}
}

// Defines a func to create a query object (typically just object and type meta portions) that's used to query k8s
// resources.
func (mpiOperatorResourceHandler) BuildIdentityResource(ctx context.Context, taskCtx pluginsCore.TaskExecutionMetadata) (client.Object, error) {
	return &kubeflowv1.MPIJob{
		TypeMeta: metav1.TypeMeta{
			Kind:       kubeflowv1.MPIJobKind,
			APIVersion: kubeflowv1.SchemeGroupVersion.String(),
		},
	}, nil
}

// Defines a func to create the full resource object that will be posted to k8s.
func (mpiOperatorResourceHandler) BuildResource(ctx context.Context, taskCtx pluginsCore.TaskExecutionContext) (client.Object, error) {
	taskTemplate, err := taskCtx.TaskReader().Read(ctx)

	if err != nil {
		return nil, flyteerr.Errorf(flyteerr.BadTaskSpecification, "unable to fetch task specification [%v]", err.Error())
	} else if taskTemplate == nil {
		return nil, flyteerr.Errorf(flyteerr.BadTaskSpecification, "nil task specification")
	}

	mpiTaskExtraArgs := plugins.DistributedMPITrainingTask{}
	err = utils.UnmarshalStruct(taskTemplate.GetCustom(), &mpiTaskExtraArgs)
	if err != nil {
		return nil, flyteerr.Errorf(flyteerr.BadTaskSpecification, "invalid TaskSpecification [%v], Err: [%v]", taskTemplate.GetCustom(), err.Error())
	}

	workers := mpiTaskExtraArgs.GetNumWorkers()
	launcherReplicas := mpiTaskExtraArgs.GetNumLauncherReplicas()
	slots := mpiTaskExtraArgs.GetSlots()

	podSpec, err := flytek8s.ToK8sPodSpec(ctx, taskCtx)
	if err != nil {
		return nil, flyteerr.Errorf(flyteerr.BadTaskSpecification, "Unable to create pod spec: [%v]", err.Error())
	}

	// workersPodSpec is deepCopy of podSpec submitted by flyte
	// WorkerPodSpec doesn't need any Argument & command. It will be trigger from launcher pod
	workersPodSpec := podSpec.DeepCopy()

	for k := range workersPodSpec.Containers {
		workersPodSpec.Containers[k].Args = []string{}
		workersPodSpec.Containers[k].Command = []string{}
	}

	if workers == 0 {
		return nil, fmt.Errorf("number of worker should be more then 1 ")
	}
	if launcherReplicas == 0 {
		return nil, fmt.Errorf("number of launch worker should be more then 1")
	}

	jobSpec := kubeflowv1.MPIJobSpec{
		SlotsPerWorker: &slots,
		MPIReplicaSpecs: map[commonKf.ReplicaType]*commonKf.ReplicaSpec{
			kubeflowv1.MPIJobReplicaTypeLauncher: {
				Replicas: &launcherReplicas,
				Template: v1.PodTemplateSpec{
					Spec: *podSpec,
				},
				RestartPolicy: commonKf.RestartPolicyNever,
			},
			kubeflowv1.MPIJobReplicaTypeWorker: {
				Replicas: &workers,
				Template: v1.PodTemplateSpec{
					Spec: *workersPodSpec,
				},
				RestartPolicy: commonKf.RestartPolicyNever,
			},
		},
	}

	job := &kubeflowv1.MPIJob{
		TypeMeta: metav1.TypeMeta{
			Kind:       kubeflowv1.MPIJobKind,
			APIVersion: kubeflowv1.SchemeGroupVersion.String(),
		},
		Spec: jobSpec,
	}

	return job, nil
}

// Analyzes the k8s resource and reports the status as TaskPhase. This call is expected to be relatively fast,
// any operations that might take a long time (limits are configured system-wide) should be offloaded to the
// background.
func (mpiOperatorResourceHandler) GetTaskPhase(_ context.Context, pluginContext k8s.PluginContext, resource client.Object) (pluginsCore.PhaseInfo, error) {
	var numWorkers, numLauncherReplicas *int32
	app, ok := resource.(*kubeflowv1.MPIJob)
	if !ok {
		return pluginsCore.PhaseInfoUndefined, fmt.Errorf("failed to convert resource data type")
	}

	numWorkers = app.Spec.MPIReplicaSpecs[kubeflowv1.MPIJobReplicaTypeWorker].Replicas
	numLauncherReplicas = app.Spec.MPIReplicaSpecs[kubeflowv1.MPIJobReplicaTypeLauncher].Replicas

	taskLogs, err := common.GetLogs(common.MPITaskType, app.Name, app.Namespace,
		*numWorkers, *numLauncherReplicas, 0)
	if err != nil {
		return pluginsCore.PhaseInfoUndefined, err
	}
	currentCondition, err := common.ExtractMPICurrentCondition(app.Status.Conditions)
	if err != nil {
		return pluginsCore.PhaseInfoUndefined, err
	}

	occurredAt := time.Now()
	statusDetails, _ := utils.MarshalObjToStruct(app.Status)
	taskPhaseInfo := pluginsCore.TaskInfo{
		Logs:       taskLogs,
		OccurredAt: &occurredAt,
		CustomInfo: statusDetails,
	}

	return common.GetMPIPhaseInfo(currentCondition, occurredAt, taskPhaseInfo)

}

func init() {
	if err := kubeflowv1.AddToScheme(scheme.Scheme); err != nil {
		panic(err)
	}

	pluginmachinery.PluginRegistry().RegisterK8sPlugin(
		k8s.PluginEntry{
			ID:                  common.MPITaskType,
			RegisteredTaskTypes: []pluginsCore.TaskType{common.MPITaskType},
			ResourceToWatch:     &kubeflowv1.MPIJob{},
			Plugin:              mpiOperatorResourceHandler{},
			IsDefault:           false,
		})
}
