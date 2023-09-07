package mpi

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/plugins"
	kfplugins "github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/plugins/kubeflow"

	flyteerr "github.com/flyteorg/flyteplugins/go/tasks/errors"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery"
	pluginsCore "github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/flytek8s"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/k8s"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/utils"
	"github.com/flyteorg/flyteplugins/go/tasks/plugins/k8s/kfoperators/common"
	commonOp "github.com/kubeflow/common/pkg/apis/common/v1"
	kubeflowv1 "github.com/kubeflow/training-operator/pkg/apis/kubeflow.org/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const workerSpecCommandKey = "worker_spec_command"

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

	podSpec, objectMeta, primaryContainerName, err := flytek8s.ToK8sPodSpec(ctx, taskCtx)
	if err != nil {
		return nil, flyteerr.Errorf(flyteerr.BadTaskSpecification, "Unable to create pod spec: [%v]", err.Error())
	}
	common.OverridePrimaryContainerName(podSpec, primaryContainerName, kubeflowv1.MPIJobDefaultContainerName)

	var launcherReplica = common.ReplicaEntry{
		ReplicaNum:    int32(1),
		PodSpec:       podSpec.DeepCopy(),
		RestartPolicy: commonOp.RestartPolicyNever,
	}
	var workerReplica = common.ReplicaEntry{
		ReplicaNum:    int32(0),
		PodSpec:       podSpec.DeepCopy(),
		RestartPolicy: commonOp.RestartPolicyNever,
	}
	slots := int32(1)
	runPolicy := commonOp.RunPolicy{}

	if taskTemplate.TaskTypeVersion == 0 {
		mpiTaskExtraArgs := plugins.DistributedMPITrainingTask{}
		err = utils.UnmarshalStruct(taskTemplate.GetCustom(), &mpiTaskExtraArgs)
		if err != nil {
			return nil, flyteerr.Errorf(flyteerr.BadTaskSpecification, "invalid TaskSpecification [%v], Err: [%v]", taskTemplate.GetCustom(), err.Error())
		}

		workerReplica.ReplicaNum = mpiTaskExtraArgs.GetNumWorkers()
		launcherReplica.ReplicaNum = mpiTaskExtraArgs.GetNumLauncherReplicas()
		slots = mpiTaskExtraArgs.GetSlots()

		// V1 requires passing worker command as template config parameter
		taskTemplateConfig := taskTemplate.GetConfig()
		workerSpecCommand := []string{}
		if val, ok := taskTemplateConfig[workerSpecCommandKey]; ok {
			workerSpecCommand = strings.Split(val, " ")
		}

		for k := range workerReplica.PodSpec.Containers {
			if workerReplica.PodSpec.Containers[k].Name == kubeflowv1.MPIJobDefaultContainerName {
				workerReplica.PodSpec.Containers[k].Args = workerSpecCommand
				workerReplica.PodSpec.Containers[k].Command = []string{}
			}
		}

	} else if taskTemplate.TaskTypeVersion == 1 {
		kfMPITaskExtraArgs := kfplugins.DistributedMPITrainingTask{}

		err = utils.UnmarshalStruct(taskTemplate.GetCustom(), &kfMPITaskExtraArgs)
		if err != nil {
			return nil, flyteerr.Errorf(flyteerr.BadTaskSpecification, "invalid TaskSpecification [%v], Err: [%v]", taskTemplate.GetCustom(), err.Error())
		}

		launcherReplicaSpec := kfMPITaskExtraArgs.GetLauncherReplicas()
		if launcherReplicaSpec != nil {
			// flyte commands will be passed as args to the container
			err = common.OverrideContainerSpec(
				launcherReplica.PodSpec,
				kubeflowv1.MPIJobDefaultContainerName,
				launcherReplicaSpec.GetImage(),
				launcherReplicaSpec.GetResources(),
				launcherReplicaSpec.GetCommand(),
			)
			if err != nil {
				return nil, err
			}
			launcherReplica.RestartPolicy = common.ParseRestartPolicy(launcherReplicaSpec.GetRestartPolicy())
		}

		workerReplicaSpec := kfMPITaskExtraArgs.GetWorkerReplicas()
		if workerReplicaSpec != nil {
			err = common.OverrideContainerSpec(
				workerReplica.PodSpec,
				kubeflowv1.MPIJobDefaultContainerName,
				workerReplicaSpec.GetImage(),
				workerReplicaSpec.GetResources(),
				workerReplicaSpec.GetCommand(),
			)
			if err != nil {
				return nil, err
			}
			workerReplica.RestartPolicy = common.ParseRestartPolicy(workerReplicaSpec.GetRestartPolicy())
			workerReplica.ReplicaNum = workerReplicaSpec.GetReplicas()
		}

		if kfMPITaskExtraArgs.GetRunPolicy() != nil {
			runPolicy = common.ParseRunPolicy(*kfMPITaskExtraArgs.GetRunPolicy())
		}

	} else {
		return nil, flyteerr.Errorf(flyteerr.BadTaskSpecification,
			"Invalid TaskSpecification, unsupported task template version [%v] key", taskTemplate.TaskTypeVersion)
	}

	if workerReplica.ReplicaNum == 0 {
		return nil, fmt.Errorf("number of worker should be more then 0")
	}
	if launcherReplica.ReplicaNum == 0 {
		return nil, fmt.Errorf("number of launch worker should be more then 0")
	}

	jobSpec := kubeflowv1.MPIJobSpec{
		SlotsPerWorker: &slots,
		RunPolicy:      runPolicy,
		MPIReplicaSpecs: map[commonOp.ReplicaType]*commonOp.ReplicaSpec{
			kubeflowv1.MPIJobReplicaTypeLauncher: {
				Replicas: &launcherReplica.ReplicaNum,
				Template: v1.PodTemplateSpec{
					ObjectMeta: *objectMeta,
					Spec:       *launcherReplica.PodSpec,
				},
				RestartPolicy: launcherReplica.RestartPolicy,
			},
			kubeflowv1.MPIJobReplicaTypeWorker: {
				Replicas: &workerReplica.ReplicaNum,
				Template: v1.PodTemplateSpec{
					ObjectMeta: *objectMeta,
					Spec:       *workerReplica.PodSpec,
				},
				RestartPolicy: workerReplica.RestartPolicy,
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

	taskLogs, err := common.GetLogs(pluginContext, common.MPITaskType, app.ObjectMeta, false,
		*numWorkers, *numLauncherReplicas, 0)
	if err != nil {
		return pluginsCore.PhaseInfoUndefined, err
	}
	if app.Status.StartTime == nil && app.CreationTimestamp.Add(common.GetConfig().Timeout.Duration).Before(time.Now()) {
		return pluginsCore.PhaseInfoUndefined, fmt.Errorf("kubeflow operator hasn't updated the mpi custom resource since creation time %v", app.CreationTimestamp)
	}
	currentCondition, err := common.ExtractCurrentCondition(app.Status.Conditions)
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
