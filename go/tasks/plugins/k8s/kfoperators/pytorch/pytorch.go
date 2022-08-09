package pytorch

import (
	"context"
	"time"

	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/flyteorg/flyteplugins/go/tasks/plugins/k8s/kfoperators/common"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/plugins"
	flyteerr "github.com/flyteorg/flyteplugins/go/tasks/errors"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/flytek8s"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/scheme"

	pluginsCore "github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/k8s"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/utils"

	commonOp "github.com/kubeflow/common/pkg/apis/common/v1"
	kubeflowv1 "github.com/kubeflow/training-operator/pkg/apis/kubeflow.org/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

	pytorchTaskExtraArgs := plugins.DistributedPyTorchTrainingTask{}
	err = utils.UnmarshalStruct(taskTemplate.GetCustom(), &pytorchTaskExtraArgs)
	if err != nil {
		return nil, flyteerr.Errorf(flyteerr.BadTaskSpecification, "invalid TaskSpecification [%v], Err: [%v]", taskTemplate.GetCustom(), err.Error())
	}

	podSpec, err := flytek8s.ToK8sPodSpec(ctx, taskCtx)
	if err != nil {
		return nil, flyteerr.Errorf(flyteerr.BadTaskSpecification, "Unable to create pod spec: [%v]", err.Error())
	}

	common.OverrideDefaultContainerName(taskCtx, podSpec, kubeflowv1.PytorchJobDefaultContainerName)

	workers := pytorchTaskExtraArgs.GetWorkers()

	jobSpec := kubeflowv1.PyTorchJobSpec{
		PyTorchReplicaSpecs: map[commonOp.ReplicaType]*commonOp.ReplicaSpec{
			kubeflowv1.PyTorchJobReplicaTypeMaster: {
				Template: v1.PodTemplateSpec{
					Spec: *podSpec,
				},
				RestartPolicy: commonOp.RestartPolicyNever,
			},
			kubeflowv1.PyTorchJobReplicaTypeWorker: {
				Replicas: &workers,
				Template: v1.PodTemplateSpec{
					Spec: *podSpec,
				},
				RestartPolicy: commonOp.RestartPolicyNever,
			},
		},
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

// Analyses the k8s resource and reports the status as TaskPhase. This call is expected to be relatively fast,
// any operations that might take a long time (limits are configured system-wide) should be offloaded to the
// background.
func (pytorchOperatorResourceHandler) GetTaskPhase(_ context.Context, pluginContext k8s.PluginContext, resource client.Object) (pluginsCore.PhaseInfo, error) {
	app := resource.(*kubeflowv1.PyTorchJob)

	workersCount := app.Spec.PyTorchReplicaSpecs[kubeflowv1.PyTorchJobReplicaTypeWorker].Replicas

	taskLogs, err := common.GetLogs(common.PytorchTaskType, app.Name, app.Namespace, *workersCount, 0, 0)
	if err != nil {
		return pluginsCore.PhaseInfoUndefined, err
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
			ID:                  common.PytorchTaskType,
			RegisteredTaskTypes: []pluginsCore.TaskType{common.PytorchTaskType},
			ResourceToWatch:     &kubeflowv1.PyTorchJob{},
			Plugin:              pytorchOperatorResourceHandler{},
			IsDefault:           false,
		})
}
