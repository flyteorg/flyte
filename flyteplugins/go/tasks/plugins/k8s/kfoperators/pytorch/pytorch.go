package pytorch

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

type pytorchOperatorResourceHandler struct {
}

// Sanity test that the plugin implements method of k8s.Plugin
var _ k8s.Plugin = pytorchOperatorResourceHandler{}

func (pytorchOperatorResourceHandler) GetProperties() k8s.PluginProperties {
	return k8s.PluginProperties{}
}

// Defines a func to create a query object (typically just object and type meta portions) that's used to query k8s
// resources.
func (pytorchOperatorResourceHandler) BuildIdentityResource(ctx context.Context, taskCtx pluginsCore.TaskExecutionMetadata) (client.Object, error) {
	return &kubeflowv1.PyTorchJob{
		TypeMeta: metav1.TypeMeta{
			Kind:       kubeflowv1.PytorchJobKind,
			APIVersion: kubeflowv1.SchemeGroupVersion.String(),
		},
	}, nil
}

// Defines a func to create the full resource object that will be posted to k8s.
func (pytorchOperatorResourceHandler) BuildResource(ctx context.Context, taskCtx pluginsCore.TaskExecutionContext) (client.Object, error) {
	taskTemplate, err := taskCtx.TaskReader().Read(ctx)

	if err != nil {
		return nil, flyteerr.Errorf(flyteerr.BadTaskSpecification, "unable to fetch task specification [%v]", err.Error())
	} else if taskTemplate == nil {
		return nil, flyteerr.Errorf(flyteerr.BadTaskSpecification, "nil task specification")
	}

	runPolicy := commonOp.RunPolicy{}
	var elasticPolicy *kubeflowv1.ElasticPolicy

	var masterReplicaSpec, workerReplicaSpec *commonOp.ReplicaSpec

	if taskTemplate.TaskTypeVersion == 0 {
		pytorchTaskExtraArgs := plugins.DistributedPyTorchTrainingTask{}

		err = utils.UnmarshalStruct(taskTemplate.GetCustom(), &pytorchTaskExtraArgs)
		if err != nil {
			return nil, flyteerr.Errorf(flyteerr.BadTaskSpecification, "invalid TaskSpecification [%v], Err: [%v]", taskTemplate.GetCustom(), err.Error())
		}

		replicaSpec, err := common.ToReplicaSpec(ctx, taskCtx, kubeflowv1.PytorchJobDefaultContainerName)
		if err != nil {
			return nil, flyteerr.Errorf(flyteerr.BadTaskSpecification, "Unable to create replica spec: [%v]", err.Error())
		}
		masterReplicaSpec = replicaSpec.DeepCopy()
		masterReplicas := int32(1)
		masterReplicaSpec.Replicas = &masterReplicas
		workerReplicaSpec = replicaSpec.DeepCopy()
		workerReplicas := pytorchTaskExtraArgs.GetWorkers()
		workerReplicaSpec.Replicas = &workerReplicas

		// Set elastic config
		elasticConfig := pytorchTaskExtraArgs.GetElasticConfig()
		if elasticConfig != nil {
			elasticPolicy = ParseElasticConfig(elasticConfig)
		}
	} else if taskTemplate.TaskTypeVersion == 1 {
		kfPytorchTaskExtraArgs := kfplugins.DistributedPyTorchTrainingTask{}

		err = utils.UnmarshalStruct(taskTemplate.GetCustom(), &kfPytorchTaskExtraArgs)
		if err != nil {
			return nil, flyteerr.Errorf(flyteerr.BadTaskSpecification, "invalid TaskSpecification [%v], Err: [%v]", taskTemplate.GetCustom(), err.Error())
		}

		masterReplicaSpec, err = common.ToReplicaSpecWithOverrides(ctx, taskCtx, kfPytorchTaskExtraArgs.GetMasterReplicas(), kubeflowv1.PytorchJobDefaultContainerName, true)
		if err != nil {
			return nil, flyteerr.Errorf(flyteerr.BadTaskSpecification, "Unable to create master replica spec: [%v]", err.Error())
		}

		workerReplicaSpec, err = common.ToReplicaSpecWithOverrides(ctx, taskCtx, kfPytorchTaskExtraArgs.GetWorkerReplicas(), kubeflowv1.PytorchJobDefaultContainerName, false)
		if err != nil {
			return nil, flyteerr.Errorf(flyteerr.BadTaskSpecification, "Unable to create worker replica spec: [%v]", err.Error())
		}

		if kfPytorchTaskExtraArgs.GetRunPolicy() != nil {
			runPolicy = common.ParseRunPolicy(*kfPytorchTaskExtraArgs.GetRunPolicy())
		}
		// Set elastic config
		elasticConfig := kfPytorchTaskExtraArgs.GetElasticConfig()
		if elasticConfig != nil {
			elasticPolicy = ParseElasticConfig(elasticConfig)
		}
	} else {
		return nil, flyteerr.Errorf(flyteerr.BadTaskSpecification,
			"Invalid TaskSpecification, unsupported task template version [%v] key", taskTemplate.TaskTypeVersion)
	}

	if *workerReplicaSpec.Replicas <= 0 {
		return nil, fmt.Errorf("number of workers must be greater than 0")
	}

	jobSpec := kubeflowv1.PyTorchJobSpec{
		PyTorchReplicaSpecs: map[commonOp.ReplicaType]*commonOp.ReplicaSpec{
			kubeflowv1.PyTorchJobReplicaTypeMaster: masterReplicaSpec,
			kubeflowv1.PyTorchJobReplicaTypeWorker: workerReplicaSpec,
		},
		RunPolicy: runPolicy,
	}

	if elasticPolicy != nil {
		jobSpec.ElasticPolicy = elasticPolicy
		// Remove master replica spec if elastic policy is set
		delete(jobSpec.PyTorchReplicaSpecs, kubeflowv1.PyTorchJobReplicaTypeMaster)
	}

	job := &kubeflowv1.PyTorchJob{
		TypeMeta: metav1.TypeMeta{
			Kind:       kubeflowv1.PytorchJobKind,
			APIVersion: kubeflowv1.SchemeGroupVersion.String(),
		},
		Spec: jobSpec,
	}

	return job, nil
}

// Interface for unified elastic config handling across plugin version v0 and v1. This interface should
// always be aligned with the ElasticConfig defined in flyteidl.
type ElasticConfig interface {
	GetMinReplicas() int32
	GetMaxReplicas() int32
	GetNprocPerNode() int32
	GetMaxRestarts() int32
	GetRdzvBackend() string
}

// To support parsing elastic config from both v0 and v1 of kubeflow pytorch idl
func ParseElasticConfig(elasticConfig ElasticConfig) *kubeflowv1.ElasticPolicy {
	minReplicas := elasticConfig.GetMinReplicas()
	maxReplicas := elasticConfig.GetMaxReplicas()
	nProcPerNode := elasticConfig.GetNprocPerNode()
	maxRestarts := elasticConfig.GetMaxRestarts()
	rdzvBackend := kubeflowv1.RDZVBackend(elasticConfig.GetRdzvBackend())
	return &kubeflowv1.ElasticPolicy{
		MinReplicas:  &minReplicas,
		MaxReplicas:  &maxReplicas,
		RDZVBackend:  &rdzvBackend,
		NProcPerNode: &nProcPerNode,
		MaxRestarts:  &maxRestarts,
	}
}

// Analyses the k8s resource and reports the status as TaskPhase. This call is expected to be relatively fast,
// any operations that might take a long time (limits are configured system-wide) should be offloaded to the
// background.
func (pytorchOperatorResourceHandler) GetTaskPhase(_ context.Context, pluginContext k8s.PluginContext, resource client.Object) (pluginsCore.PhaseInfo, error) {
	app, ok := resource.(*kubeflowv1.PyTorchJob)
	if !ok {
		return pluginsCore.PhaseInfoUndefined, fmt.Errorf("failed to convert resource data type")
	}

	// Elastic PytorchJobs don't use master replicas
	hasMaster := false
	if _, ok := app.Spec.PyTorchReplicaSpecs[kubeflowv1.PyTorchJobReplicaTypeMaster]; ok {
		hasMaster = true
	}

	workersCount := common.GetReplicaCount(app.Spec.PyTorchReplicaSpecs, kubeflowv1.PyTorchJobReplicaTypeWorker)

	taskLogs, err := common.GetLogs(pluginContext, common.PytorchTaskType, app.ObjectMeta, hasMaster, *workersCount, 0, 0, 0)
	if err != nil {
		return pluginsCore.PhaseInfoUndefined, err
	}

	if app.Status.StartTime == nil && app.CreationTimestamp.Add(common.GetConfig().Timeout.Duration).Before(time.Now()) {
		return pluginsCore.PhaseInfoUndefined, fmt.Errorf("kubeflow operator hasn't updated the pytorch custom resource since creation time %v", app.CreationTimestamp)
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
		return pluginsCore.PhaseInfoUndefined, phaseVersionUpdateErr
	}

	return phaseInfo, err
}

func init() {
	if err := kubeflowv1.AddToScheme(scheme.Scheme); err != nil {
		panic(err)
	}

	pluginmachinery.PluginRegistry().RegisterK8sPlugin(
		k8s.PluginEntry{
			ID:                  common.PytorchTaskType,
			RegisteredTaskTypes: []pluginsCore.TaskType{common.PytorchTaskType},
			ResourceToWatch:     &kubeflowv1.PyTorchJob{},
			Plugin:              pytorchOperatorResourceHandler{},
			IsDefault:           false,
		})
}
