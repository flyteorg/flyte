package pytorch

import (
	"context"
	"fmt"

	"sort"
	"time"

	"github.com/lyft/flyteplugins/go/tasks/pluginmachinery/tasklog"

	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/plugins"
	flyteerr "github.com/lyft/flyteplugins/go/tasks/errors"
	"github.com/lyft/flyteplugins/go/tasks/pluginmachinery"
	"github.com/lyft/flyteplugins/go/tasks/pluginmachinery/flytek8s"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/scheme"

	pluginsCore "github.com/lyft/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/lyft/flyteplugins/go/tasks/pluginmachinery/k8s"
	"github.com/lyft/flyteplugins/go/tasks/pluginmachinery/utils"

	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/lyft/flyteplugins/go/tasks/logs"

	//commonOp "github.com/kubeflow/common/pkg/apis/common/v1" // switch to real 'common' once https://github.com/kubeflow/pytorch-operator/issues/263 resolved
	ptOp "github.com/kubeflow/pytorch-operator/pkg/apis/pytorch/v1"
	commonOp "github.com/kubeflow/tf-operator/pkg/apis/common/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	pytorchTaskType = "pytorch"
)

type pytorchOperatorResourceHandler struct {
}

// Sanity test that the plugin implements method of k8s.Plugin
var _ k8s.Plugin = pytorchOperatorResourceHandler{}

// Defines a func to create a query object (typically just object and type meta portions) that's used to query k8s
// resources.
func (pytorchOperatorResourceHandler) BuildIdentityResource(ctx context.Context, taskCtx pluginsCore.TaskExecutionMetadata) (k8s.Resource, error) {
	return &ptOp.PyTorchJob{
		TypeMeta: metav1.TypeMeta{
			Kind:       ptOp.Kind,
			APIVersion: ptOp.SchemeGroupVersion.String(),
		},
	}, nil
}

// Defines a func to create the full resource object that will be posted to k8s.
func (pytorchOperatorResourceHandler) BuildResource(ctx context.Context, taskCtx pluginsCore.TaskExecutionContext) (k8s.Resource, error) {
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

	podSpec, err := flytek8s.ToK8sPodSpec(ctx, taskCtx.TaskExecutionMetadata(), taskCtx.TaskReader(), taskCtx.InputReader(), taskCtx.OutputWriter())
	if err != nil {
		return nil, flyteerr.Errorf(flyteerr.BadTaskSpecification, "Unable to create pod spec: [%v]", err.Error())
	}

	overrideDefaultContainerName(taskCtx, podSpec)

	workers := pytorchTaskExtraArgs.GetWorkers()

	jobSpec := ptOp.PyTorchJobSpec{
		TTLSecondsAfterFinished: nil,
		PyTorchReplicaSpecs: map[ptOp.PyTorchReplicaType]*commonOp.ReplicaSpec{
			ptOp.PyTorchReplicaTypeMaster: {
				Template: v1.PodTemplateSpec{
					Spec: *podSpec,
				},
				RestartPolicy: commonOp.RestartPolicyNever,
			},
			ptOp.PyTorchReplicaTypeWorker: {
				Replicas: &workers,
				Template: v1.PodTemplateSpec{
					Spec: *podSpec,
				},
				RestartPolicy: commonOp.RestartPolicyNever,
			},
		},
	}

	job := &ptOp.PyTorchJob{
		TypeMeta: metav1.TypeMeta{
			Kind:       ptOp.Kind,
			APIVersion: ptOp.SchemeGroupVersion.String(),
		},
		Spec: jobSpec,
	}

	return job, nil
}

// Analyses the k8s resource and reports the status as TaskPhase. This call is expected to be relatively fast,
// any operations that might take a long time (limits are configured system-wide) should be offloaded to the
// background.
func (pytorchOperatorResourceHandler) GetTaskPhase(_ context.Context, pluginContext k8s.PluginContext, resource k8s.Resource) (pluginsCore.PhaseInfo, error) {
	app := resource.(*ptOp.PyTorchJob)

	workersCount := app.Spec.PyTorchReplicaSpecs[ptOp.PyTorchReplicaTypeWorker].Replicas

	taskLogs, err := getLogs(app, *workersCount)
	if err != nil {
		return pluginsCore.PhaseInfoUndefined, err
	}

	currentCondition, err := extractCurrentCondition(app.Status.Conditions)
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

	switch currentCondition.Type {
	case commonOp.JobCreated:
		return pluginsCore.PhaseInfoQueued(occurredAt, pluginsCore.DefaultPhaseVersion, "JobCreated"), nil
	case commonOp.JobRunning:
		return pluginsCore.PhaseInfoRunning(pluginsCore.DefaultPhaseVersion, &taskPhaseInfo), nil
	case commonOp.JobSucceeded:
		return pluginsCore.PhaseInfoSuccess(&taskPhaseInfo), nil
	case commonOp.JobFailed:
		details := fmt.Sprintf("Job failed:\n\t%v - %v", currentCondition.Reason, currentCondition.Message)
		return pluginsCore.PhaseInfoRetryableFailure(flyteerr.DownstreamSystemError, details, &taskPhaseInfo), nil
	case commonOp.JobRestarting:
		details := fmt.Sprintf("Job failed:\n\t%v - %v", currentCondition.Reason, currentCondition.Message)
		return pluginsCore.PhaseInfoRetryableFailure(flyteerr.RuntimeFailure, details, &taskPhaseInfo), nil
	}

	return pluginsCore.PhaseInfoUndefined, nil
}

func getLogs(app *ptOp.PyTorchJob, workersCount int32) ([]*core.TaskLog, error) {
	p, err := logs.InitializeLogPlugins(logs.GetLogConfig())
	if err != nil {
		return nil, err
	}

	if p == nil {
		return nil, nil
	}

	o, err := p.GetTaskLogs(tasklog.Input{
		PodName:   app.Name + "-master-0",
		Namespace: app.Namespace,
		LogName:   "master",
	})

	if err != nil {
		return nil, err
	}

	taskLogs := make([]*core.TaskLog, 0, 10)
	taskLogs = append(taskLogs, o.TaskLogs...)

	for workerIndex := int32(0); workerIndex < workersCount; workerIndex++ {
		workerLog, err := p.GetTaskLogs(tasklog.Input{
			PodName:   app.Name + fmt.Sprintf("-worker-%d", workerIndex),
			Namespace: app.Namespace,
		})

		if err != nil {
			return nil, err
		}

		taskLogs = append(taskLogs, workerLog.TaskLogs...)
	}

	return taskLogs, nil
}

func extractCurrentCondition(jobConditions []commonOp.JobCondition) (commonOp.JobCondition, error) {
	sort.Slice(jobConditions[:], func(i, j int) bool {
		return jobConditions[i].LastTransitionTime.Time.After(jobConditions[j].LastTransitionTime.Time)
	})

	for _, jc := range jobConditions {
		if jc.Status == v1.ConditionTrue {
			return jc, nil
		}
	}

	return commonOp.JobCondition{}, fmt.Errorf("found no current condition. Conditions: %+v", jobConditions)
}

func overrideDefaultContainerName(taskCtx pluginsCore.TaskExecutionContext, podSpec *v1.PodSpec) {
	// Pytorch operator forces pod to have container named 'pytorch'
	// https://github.com/kubeflow/pytorch-operator/blob/037cd1b18eb77f657f2a4bc8a8334f2a06324b57/pkg/apis/pytorch/validation/validation.go#L54-L62
	// hence we have to override the name set here
	// https://github.com/lyft/flyteplugins/blob/209c52d002b4e6a39be5d175bc1046b7e631c153/go/tasks/pluginmachinery/flytek8s/container_helper.go#L116
	flyteDefaultContainerName := taskCtx.TaskExecutionMetadata().GetTaskExecutionID().GetGeneratedName()
	for idx, c := range podSpec.Containers {
		if c.Name == flyteDefaultContainerName {
			podSpec.Containers[idx].Name = ptOp.DefaultContainerName
			return
		}
	}
}

func init() {
	if err := ptOp.AddToScheme(scheme.Scheme); err != nil {
		panic(err)
	}

	pluginmachinery.PluginRegistry().RegisterK8sPlugin(
		k8s.PluginEntry{
			ID:                  pytorchTaskType,
			RegisteredTaskTypes: []pluginsCore.TaskType{pytorchTaskType},
			ResourceToWatch:     &ptOp.PyTorchJob{},
			Plugin:              pytorchOperatorResourceHandler{},
			IsDefault:           false,
			DefaultForTaskTypes: []pluginsCore.TaskType{pytorchTaskType},
		})
}
