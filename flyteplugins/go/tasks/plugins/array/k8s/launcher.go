package k8s

import (
	"context"
	"fmt"
	"strconv"

	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/utils"

	"github.com/flyteorg/flyteplugins/go/tasks/plugins/array/errorcollector"

	arrayCore "github.com/flyteorg/flyteplugins/go/tasks/plugins/array/core"

	errors2 "github.com/flyteorg/flytestdlib/errors"

	corev1 "k8s.io/api/core/v1"

	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/core"
)

const (
	ErrBuildPodTemplate       errors2.ErrorCode = "POD_TEMPLATE_FAILED"
	ErrReplaceCmdTemplate     errors2.ErrorCode = "CMD_TEMPLATE_FAILED"
	ErrSubmitJob              errors2.ErrorCode = "SUBMIT_JOB_FAILED"
	ErrGetTaskTypeVersion     errors2.ErrorCode = "GET_TASK_TYPE_VERSION_FAILED"
	JobIndexVarName           string            = "BATCH_JOB_ARRAY_INDEX_VAR_NAME"
	FlyteK8sArrayIndexVarName string            = "FLYTE_K8S_ARRAY_INDEX"
)

var arrayJobEnvVars = []corev1.EnvVar{
	{
		Name:  JobIndexVarName,
		Value: FlyteK8sArrayIndexVarName,
	},
}

func formatSubTaskName(_ context.Context, parentName string, index int, retryAttempt uint64) (subTaskName string) {
	indexStr := strconv.Itoa(index)

	// If the retryAttempt is 0 we do not include it in the pod name. The gives us backwards
	// compatibility in the ability to dynamically transition running map tasks to use subtask retries.
	if retryAttempt == 0 {
		return utils.ConvertToDNS1123SubdomainCompatibleString(fmt.Sprintf("%v-%v", parentName, indexStr))
	}

	retryAttemptStr := strconv.FormatUint(retryAttempt, 10)
	return utils.ConvertToDNS1123SubdomainCompatibleString(fmt.Sprintf("%v-%v-%v", parentName, indexStr, retryAttemptStr))
}

func ApplyPodPolicies(_ context.Context, cfg *Config, pod *corev1.Pod) *corev1.Pod {
	if len(cfg.DefaultScheduler) > 0 {
		pod.Spec.SchedulerName = cfg.DefaultScheduler
	}

	return pod
}

func applyNodeSelectorLabels(_ context.Context, cfg *Config, pod *corev1.Pod) *corev1.Pod {
	if len(cfg.NodeSelector) != 0 {
		pod.Spec.NodeSelector = cfg.NodeSelector
	}

	return pod
}

func applyPodTolerations(_ context.Context, cfg *Config, pod *corev1.Pod) *corev1.Pod {
	if len(cfg.Tolerations) != 0 {
		pod.Spec.Tolerations = cfg.Tolerations
	}

	return pod
}

func TerminateSubTasks(ctx context.Context, tCtx core.TaskExecutionContext, kubeClient core.KubeClient, config *Config,
	currentState *arrayCore.State) error {

	size := currentState.GetExecutionArraySize()
	errs := errorcollector.NewErrorMessageCollector()
	for childIdx := 0; childIdx < size; childIdx++ {
		task := Task{
			ChildIdx: childIdx,
			Config:   config,
			State:    currentState,
		}

		err := task.Abort(ctx, tCtx, kubeClient)
		if err != nil {
			errs.Collect(childIdx, err.Error())
		}
		err = task.Finalize(ctx, tCtx, kubeClient)
		if err != nil {
			errs.Collect(childIdx, err.Error())
		}
	}

	if errs.Length() > 0 {
		return fmt.Errorf(errs.Summary(config.MaxErrorStringLength))
	}

	return nil
}
