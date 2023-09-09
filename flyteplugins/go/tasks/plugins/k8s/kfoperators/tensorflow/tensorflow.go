package tensorflow

import (
	"context"
	"fmt"
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

type tensorflowOperatorResourceHandler struct {
}

// Sanity test that the plugin implements method of k8s.Plugin
var _ k8s.Plugin = tensorflowOperatorResourceHandler{}

func (tensorflowOperatorResourceHandler) GetProperties() k8s.PluginProperties {
	return k8s.PluginProperties{}
}

// Defines a func to create a query object (typically just object and type meta portions) that's used to query k8s
// resources.
func (tensorflowOperatorResourceHandler) BuildIdentityResource(ctx context.Context, taskCtx pluginsCore.TaskExecutionMetadata) (client.Object, error) {
	return &kubeflowv1.TFJob{
		TypeMeta: metav1.TypeMeta{
			Kind:       kubeflowv1.TFJobKind,
			APIVersion: kubeflowv1.SchemeGroupVersion.String(),
		},
	}, nil
}

// Defines a func to create the full resource object that will be posted to k8s.
func (tensorflowOperatorResourceHandler) BuildResource(ctx context.Context, taskCtx pluginsCore.TaskExecutionContext) (client.Object, error) {
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
	common.OverridePrimaryContainerName(podSpec, primaryContainerName, kubeflowv1.TFJobDefaultContainerName)

	replicaSpecMap := map[commonOp.ReplicaType]*common.ReplicaEntry{
		kubeflowv1.TFJobReplicaTypeChief: {
			ReplicaNum:    int32(0),
			PodSpec:       podSpec.DeepCopy(),
			RestartPolicy: commonOp.RestartPolicyNever,
		},
		kubeflowv1.TFJobReplicaTypeWorker: {
			ReplicaNum:    int32(0),
			PodSpec:       podSpec.DeepCopy(),
			RestartPolicy: commonOp.RestartPolicyNever,
		},
		kubeflowv1.TFJobReplicaTypePS: {
			ReplicaNum:    int32(0),
			PodSpec:       podSpec.DeepCopy(),
			RestartPolicy: commonOp.RestartPolicyNever,
		},
	}
	runPolicy := commonOp.RunPolicy{}

	if taskTemplate.TaskTypeVersion == 0 {
		tensorflowTaskExtraArgs := plugins.DistributedTensorflowTrainingTask{}

		err = utils.UnmarshalStruct(taskTemplate.GetCustom(), &tensorflowTaskExtraArgs)
		if err != nil {
			return nil, flyteerr.Errorf(flyteerr.BadTaskSpecification, "invalid TaskSpecification [%v], Err: [%v]", taskTemplate.GetCustom(), err.Error())
		}

		replicaSpecMap[kubeflowv1.TFJobReplicaTypeChief].ReplicaNum = tensorflowTaskExtraArgs.GetChiefReplicas()
		replicaSpecMap[kubeflowv1.TFJobReplicaTypeWorker].ReplicaNum = tensorflowTaskExtraArgs.GetWorkers()
		replicaSpecMap[kubeflowv1.TFJobReplicaTypePS].ReplicaNum = tensorflowTaskExtraArgs.GetPsReplicas()

	} else if taskTemplate.TaskTypeVersion == 1 {
		kfTensorflowTaskExtraArgs := kfplugins.DistributedTensorflowTrainingTask{}

		err = utils.UnmarshalStruct(taskTemplate.GetCustom(), &kfTensorflowTaskExtraArgs)
		if err != nil {
			return nil, flyteerr.Errorf(flyteerr.BadTaskSpecification, "invalid TaskSpecification [%v], Err: [%v]", taskTemplate.GetCustom(), err.Error())
		}

		chiefReplicaSpec := kfTensorflowTaskExtraArgs.GetChiefReplicas()
		if chiefReplicaSpec != nil {
			err := common.OverrideContainerSpec(
				replicaSpecMap[kubeflowv1.TFJobReplicaTypeChief].PodSpec,
				kubeflowv1.TFJobDefaultContainerName,
				chiefReplicaSpec.GetImage(),
				chiefReplicaSpec.GetResources(),
				nil,
			)
			if err != nil {
				return nil, err
			}
			replicaSpecMap[kubeflowv1.TFJobReplicaTypeChief].RestartPolicy = common.ParseRestartPolicy(chiefReplicaSpec.GetRestartPolicy())
			replicaSpecMap[kubeflowv1.TFJobReplicaTypeChief].ReplicaNum = chiefReplicaSpec.GetReplicas()
		}

		workerReplicaSpec := kfTensorflowTaskExtraArgs.GetWorkerReplicas()
		if workerReplicaSpec != nil {
			err := common.OverrideContainerSpec(
				replicaSpecMap[kubeflowv1.MPIJobReplicaTypeWorker].PodSpec,
				kubeflowv1.TFJobDefaultContainerName,
				workerReplicaSpec.GetImage(),
				workerReplicaSpec.GetResources(),
				nil,
			)
			if err != nil {
				return nil, err
			}
			replicaSpecMap[kubeflowv1.TFJobReplicaTypeWorker].RestartPolicy = common.ParseRestartPolicy(workerReplicaSpec.GetRestartPolicy())
			replicaSpecMap[kubeflowv1.TFJobReplicaTypeWorker].ReplicaNum = workerReplicaSpec.GetReplicas()
		}

		psReplicaSpec := kfTensorflowTaskExtraArgs.GetPsReplicas()
		if psReplicaSpec != nil {
			err := common.OverrideContainerSpec(
				replicaSpecMap[kubeflowv1.TFJobReplicaTypePS].PodSpec,
				kubeflowv1.TFJobDefaultContainerName,
				psReplicaSpec.GetImage(),
				psReplicaSpec.GetResources(),
				nil,
			)
			if err != nil {
				return nil, err
			}
			replicaSpecMap[kubeflowv1.TFJobReplicaTypePS].RestartPolicy = common.ParseRestartPolicy(psReplicaSpec.GetRestartPolicy())
			replicaSpecMap[kubeflowv1.TFJobReplicaTypePS].ReplicaNum = psReplicaSpec.GetReplicas()
		}

		if kfTensorflowTaskExtraArgs.GetRunPolicy() != nil {
			runPolicy = common.ParseRunPolicy(*kfTensorflowTaskExtraArgs.GetRunPolicy())
		}

	} else {
		return nil, flyteerr.Errorf(flyteerr.BadTaskSpecification,
			"Invalid TaskSpecification, unsupported task template version [%v] key", taskTemplate.TaskTypeVersion)
	}

	if replicaSpecMap[kubeflowv1.TFJobReplicaTypeWorker].ReplicaNum == 0 {
		return nil, fmt.Errorf("number of worker should be more then 0")
	}

	jobSpec := kubeflowv1.TFJobSpec{
		TFReplicaSpecs: map[commonOp.ReplicaType]*commonOp.ReplicaSpec{},
	}

	for replicaType, replicaEntry := range replicaSpecMap {
		if replicaEntry.ReplicaNum > 0 {
			jobSpec.TFReplicaSpecs[replicaType] = &commonOp.ReplicaSpec{
				Replicas: &replicaEntry.ReplicaNum,
				Template: v1.PodTemplateSpec{
					ObjectMeta: *objectMeta,
					Spec:       *replicaEntry.PodSpec,
				},
				RestartPolicy: replicaEntry.RestartPolicy,
			}
		}
	}

	jobSpec.RunPolicy = runPolicy

	job := &kubeflowv1.TFJob{
		TypeMeta: metav1.TypeMeta{
			Kind:       kubeflowv1.TFJobKind,
			APIVersion: kubeflowv1.SchemeGroupVersion.String(),
		},
		Spec: jobSpec,
	}

	return job, nil
}

// Analyses the k8s resource and reports the status as TaskPhase. This call is expected to be relatively fast,
// any operations that might take a long time (limits are configured system-wide) should be offloaded to the
// background.
func (tensorflowOperatorResourceHandler) GetTaskPhase(_ context.Context, pluginContext k8s.PluginContext, resource client.Object) (pluginsCore.PhaseInfo, error) {
	app := resource.(*kubeflowv1.TFJob)

	workersCount := app.Spec.TFReplicaSpecs[kubeflowv1.TFJobReplicaTypeWorker].Replicas
	psReplicasCount := app.Spec.TFReplicaSpecs[kubeflowv1.TFJobReplicaTypePS].Replicas
	chiefCount := app.Spec.TFReplicaSpecs[kubeflowv1.TFJobReplicaTypeChief].Replicas

	taskLogs, err := common.GetLogs(pluginContext, common.TensorflowTaskType, app.ObjectMeta, false,
		*workersCount, *psReplicasCount, *chiefCount)
	if err != nil {
		return pluginsCore.PhaseInfoUndefined, err
	}

	if app.Status.StartTime == nil && app.CreationTimestamp.Add(common.GetConfig().Timeout.Duration).Before(time.Now()) {
		return pluginsCore.PhaseInfoUndefined, fmt.Errorf("kubeflow operator hasn't updated the tensorflow custom resource since creation time %v", app.CreationTimestamp)
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

	return common.GetPhaseInfo(currentCondition, occurredAt, taskPhaseInfo)
}

func init() {
	if err := kubeflowv1.AddToScheme(scheme.Scheme); err != nil {
		panic(err)
	}

	pluginmachinery.PluginRegistry().RegisterK8sPlugin(
		k8s.PluginEntry{
			ID:                  common.TensorflowTaskType,
			RegisteredTaskTypes: []pluginsCore.TaskType{common.TensorflowTaskType},
			ResourceToWatch:     &kubeflowv1.TFJob{},
			Plugin:              tensorflowOperatorResourceHandler{},
			IsDefault:           false,
		})
}
