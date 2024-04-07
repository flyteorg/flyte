package tensorflow

import (
	"context"
	"fmt"
	"time"

	commonOp "github.com/kubeflow/common/pkg/apis/common/v1"
	kubeflowv1 "github.com/kubeflow/training-operator/pkg/apis/kubeflow.org/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/plugins"
	kfplugins "github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/plugins/kubeflow"
	flyteerr "github.com/flyteorg/flyte/flyteplugins/go/tasks/errors"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery"
	pluginsCore "github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/k8s"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/utils"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/plugins/k8s/kfoperators/common"
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

	replicaSpecMap := make(map[commonOp.ReplicaType]*commonOp.ReplicaSpec)
	runPolicy := commonOp.RunPolicy{}

	if taskTemplate.TaskTypeVersion == 0 {
		tensorflowTaskExtraArgs := plugins.DistributedTensorflowTrainingTask{}

		err = utils.UnmarshalStruct(taskTemplate.GetCustom(), &tensorflowTaskExtraArgs)
		if err != nil {
			return nil, flyteerr.Errorf(flyteerr.BadTaskSpecification, "invalid TaskSpecification [%v], Err: [%v]", taskTemplate.GetCustom(), err.Error())
		}

		replicaSpec, err := common.ToReplicaSpec(ctx, taskCtx, kubeflowv1.TFJobDefaultContainerName)
		if err != nil {
			return nil, flyteerr.Errorf(flyteerr.BadTaskSpecification, "Unable to create replica spec: [%v]", err.Error())
		}

		replicaNumMap := map[commonOp.ReplicaType]int32{
			kubeflowv1.TFJobReplicaTypeChief:  tensorflowTaskExtraArgs.GetChiefReplicas(),
			kubeflowv1.TFJobReplicaTypeWorker: tensorflowTaskExtraArgs.GetWorkers(),
			kubeflowv1.TFJobReplicaTypePS:     tensorflowTaskExtraArgs.GetPsReplicas(),
			kubeflowv1.TFJobReplicaTypeEval:   tensorflowTaskExtraArgs.GetEvaluatorReplicas(),
		}
		for t, r := range replicaNumMap {
			rs := replicaSpec.DeepCopy()
			replicas := r
			if replicas > 0 {
				rs.Replicas = &replicas
				replicaSpecMap[t] = rs
			}
		}

	} else if taskTemplate.TaskTypeVersion == 1 {
		kfTensorflowTaskExtraArgs := kfplugins.DistributedTensorflowTrainingTask{}

		err = utils.UnmarshalStruct(taskTemplate.GetCustom(), &kfTensorflowTaskExtraArgs)
		if err != nil {
			return nil, flyteerr.Errorf(flyteerr.BadTaskSpecification, "invalid TaskSpecification [%v], Err: [%v]", taskTemplate.GetCustom(), err.Error())
		}

		replicaSpecCfgMap := map[commonOp.ReplicaType]*kfplugins.DistributedTensorflowTrainingReplicaSpec{
			kubeflowv1.TFJobReplicaTypeChief:  kfTensorflowTaskExtraArgs.GetChiefReplicas(),
			kubeflowv1.TFJobReplicaTypeWorker: kfTensorflowTaskExtraArgs.GetWorkerReplicas(),
			kubeflowv1.TFJobReplicaTypePS:     kfTensorflowTaskExtraArgs.GetPsReplicas(),
			kubeflowv1.TFJobReplicaTypeEval:   kfTensorflowTaskExtraArgs.GetEvaluatorReplicas(),
		}
		for t, cfg := range replicaSpecCfgMap {
			// Short circuit if replica set has no replicas to avoid unnecessarily
			// generating pod specs
			if cfg.GetReplicas() <= 0 {
				continue
			}
			rs, err := common.ToReplicaSpecWithOverrides(ctx, taskCtx, cfg, kubeflowv1.TFJobDefaultContainerName, false)
			if err != nil {
				return nil, flyteerr.Errorf(flyteerr.BadTaskSpecification, "Unable to create replica spec: [%v]", err.Error())
			}
			replicaSpecMap[t] = rs
		}

		if kfTensorflowTaskExtraArgs.GetRunPolicy() != nil {
			runPolicy = common.ParseRunPolicy(*kfTensorflowTaskExtraArgs.GetRunPolicy())
		}

	} else {
		return nil, flyteerr.Errorf(flyteerr.BadTaskSpecification,
			"Invalid TaskSpecification, unsupported task template version [%v] key", taskTemplate.TaskTypeVersion)
	}

	if v, ok := replicaSpecMap[kubeflowv1.TFJobReplicaTypeWorker]; !ok || *v.Replicas <= 0 {
		return nil, fmt.Errorf("number of workers must be greater than 0")
	}

	jobSpec := kubeflowv1.TFJobSpec{
		TFReplicaSpecs: replicaSpecMap,
		RunPolicy:      runPolicy,
	}

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
	app, ok := resource.(*kubeflowv1.TFJob)
	if !ok {
		return pluginsCore.PhaseInfoUndefined, fmt.Errorf("failed to convert resource data type")
	}

	workersCount := common.GetReplicaCount(app.Spec.TFReplicaSpecs, kubeflowv1.TFJobReplicaTypeWorker)
	psReplicasCount := common.GetReplicaCount(app.Spec.TFReplicaSpecs, kubeflowv1.TFJobReplicaTypePS)
	chiefCount := common.GetReplicaCount(app.Spec.TFReplicaSpecs, kubeflowv1.TFJobReplicaTypeChief)
	evaluatorReplicasCount := common.GetReplicaCount(app.Spec.TFReplicaSpecs, kubeflowv1.TFJobReplicaTypeEval)

	taskLogs, err := common.GetLogs(pluginContext, common.TensorflowTaskType, app.ObjectMeta, false,
		*workersCount, *psReplicasCount, *chiefCount, *evaluatorReplicasCount)
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

	phaseInfo, err := common.GetPhaseInfo(currentCondition, occurredAt, taskPhaseInfo)

	phaseVersionUpdateErr := k8s.MaybeUpdatePhaseVersionFromPluginContext(&phaseInfo, &pluginContext)
	if phaseVersionUpdateErr != nil {
		return phaseInfo, phaseVersionUpdateErr
	}

	return phaseInfo, err
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
