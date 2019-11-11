package k8s

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/lyft/flyteplugins/go/tasks/plugins/array/errorcollector"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	arrayCore "github.com/lyft/flyteplugins/go/tasks/plugins/array/core"

	arraystatus2 "github.com/lyft/flyteplugins/go/tasks/plugins/array/arraystatus"
	errors2 "github.com/lyft/flytestdlib/errors"

	"github.com/lyft/flytestdlib/logger"

	"github.com/lyft/flyteplugins/go/tasks/pluginmachinery/utils"

	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"

	"github.com/lyft/flyteplugins/go/tasks/pluginmachinery/core"
)

const (
	ErrBuildPodTemplate       errors2.ErrorCode = "POD_TEMPLATE_FAILED"
	ErrReplaceCmdTemplate     errors2.ErrorCode = "CMD_TEMPLATE_FAILED"
	ErrSubmitJob              errors2.ErrorCode = "SUBMIT_JOB_FAILED"
	JobIndexVarName           string            = "BATCH_JOB_ARRAY_INDEX_VAR_NAME"
	FlyteK8sArrayIndexVarName string            = "FLYTE_K8S_ARRAY_INDEX"
)

var arrayJobEnvVars = []corev1.EnvVar{
	{
		Name:  JobIndexVarName,
		Value: FlyteK8sArrayIndexVarName,
	},
}

func formatSubTaskName(_ context.Context, parentName, suffix string) (subTaskName string) {
	return fmt.Sprintf("%v-%v", parentName, suffix)
}

func ApplyPodPolicies(_ context.Context, cfg *Config, pod *corev1.Pod) *corev1.Pod {
	if len(cfg.DefaultScheduler) > 0 {
		pod.Spec.SchedulerName = cfg.DefaultScheduler
	}

	return pod
}

// Launches subtasks
func LaunchSubTasks(ctx context.Context, tCtx core.TaskExecutionContext, kubeClient core.KubeClient,
	config *Config, currentState *arrayCore.State) (newState *arrayCore.State, err error) {

	if int64(currentState.GetExecutionArraySize()) > config.MaxArrayJobSize {
		ee := fmt.Errorf("array size > max allowed. Requested [%v]. Allowed [%v]", currentState.GetExecutionArraySize(), config.MaxArrayJobSize)
		logger.Info(ctx, ee)
		currentState = currentState.SetPhase(arrayCore.PhasePermanentFailure, 0).SetReason(ee.Error())
		return currentState, nil
	}

	podTemplate, _, err := FlyteArrayJobToK8sPodTemplate(ctx, tCtx)
	if err != nil {
		return currentState, errors2.Wrapf(ErrBuildPodTemplate, err, "Failed to convert task template to a pod template for task")
	}

	var command []string
	if len(podTemplate.Spec.Containers) > 0 {
		command = append(podTemplate.Spec.Containers[0].Command, podTemplate.Spec.Containers[0].Args...)
		podTemplate.Spec.Containers[0].Args = []string{}
	}

	size := currentState.GetExecutionArraySize()
	// TODO: Respect parallelism param
	for i := 0; i < size; i++ {
		pod := podTemplate.DeepCopy()
		indexStr := strconv.Itoa(i)
		pod.Name = formatSubTaskName(ctx, tCtx.TaskExecutionMetadata().GetTaskExecutionID().GetGeneratedName(), indexStr)
		pod.Spec.Containers[0].Env = append(pod.Spec.Containers[0].Env, corev1.EnvVar{
			Name:  FlyteK8sArrayIndexVarName,
			Value: indexStr,
		})

		pod.Spec.Containers[0].Env = append(pod.Spec.Containers[0].Env, arrayJobEnvVars...)

		pod.Spec.Containers[0].Command, err = utils.ReplaceTemplateCommandArgs(ctx, command, arrayJobInputReader{tCtx.InputReader()}, tCtx.OutputWriter())
		if err != nil {
			return currentState, errors2.Wrapf(ErrReplaceCmdTemplate, err, "Failed to replace cmd args")
		}

		pod = ApplyPodPolicies(ctx, config, pod)

		err = kubeClient.GetClient().Create(ctx, pod)
		if err != nil && !k8serrors.IsAlreadyExists(err) {
			if k8serrors.IsForbidden(err) {
				if strings.Contains(err.Error(), "exceeded quota") {
					// TODO: Quota errors are retried forever, it would be good to have support for backoff strategy.
					logger.Infof(ctx, "Failed to launch job, resource quota exceeded. Err: %v", err)
					currentState = currentState.SetPhase(arrayCore.PhaseWaitingForResources, 0).SetReason("Not enough resources to launch job.")
				} else {
					currentState = currentState.SetPhase(arrayCore.PhaseRetryableFailure, 0).SetReason("Failed to launch job.")
				}

				currentState = currentState.SetReason(err.Error())
				return currentState, nil
			}

			return currentState, errors2.Wrapf(ErrSubmitJob, err, "Failed to submit job")
		}
	}

	logger.Infof(ctx, "Successfully submitted Job(s) with Prefix:[%v], Count:[%v]", tCtx.TaskExecutionMetadata().GetTaskExecutionID().GetGeneratedName(), size)

	arrayStatus := arraystatus2.ArrayStatus{
		Summary:  arraystatus2.ArraySummary{},
		Detailed: arrayCore.NewPhasesCompactArray(uint(size)),
	}

	currentState.SetPhase(arrayCore.PhaseCheckingSubTaskExecutions, 0).SetReason("Job launched.")
	currentState.SetArrayStatus(arrayStatus)

	return currentState, nil
}

func TerminateSubTasks(ctx context.Context, tMeta core.TaskExecutionMetadata, kubeClient core.KubeClient,
	errsMaxLength int, currentState *arrayCore.State) error {

	size := currentState.GetExecutionArraySize()
	errs := errorcollector.NewErrorMessageCollector()
	for i := 0; i < size; i++ {
		indexStr := strconv.Itoa(i)
		podName := formatSubTaskName(ctx, tMeta.GetTaskExecutionID().GetGeneratedName(), indexStr)
		pod := &corev1.Pod{
			TypeMeta: metav1.TypeMeta{
				Kind:       PodKind,
				APIVersion: metav1.SchemeGroupVersion.String(),
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      podName,
				Namespace: tMeta.GetNamespace(),
			},
		}

		err := kubeClient.GetClient().Delete(ctx, pod)
		if err != nil {
			if k8serrors.IsNotFound(err) {
				continue
			}

			errs.Collect(i, err.Error())
		}
	}

	if errs.Length() > 0 {
		return fmt.Errorf(errs.Summary(errsMaxLength))
	}

	return nil
}
