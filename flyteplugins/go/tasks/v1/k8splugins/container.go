package k8splugins

import (
	"context"
	"time"

	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/core"
	"k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	tasksV1 "github.com/lyft/flyteplugins/go/tasks/v1"
	"github.com/lyft/flyteplugins/go/tasks/v1/errors"
	"github.com/lyft/flyteplugins/go/tasks/v1/events"
	"github.com/lyft/flyteplugins/go/tasks/v1/flytek8s"
	"github.com/lyft/flyteplugins/go/tasks/v1/logs"
	"github.com/lyft/flyteplugins/go/tasks/v1/types"
)

const (
	containerTaskType = "container"
)

func ConvertPodFailureToError(status v1.PodStatus) error {
	reason := errors.TaskFailedUnknownError
	message := "Container/Pod failed. No message received from kubernetes. Could be permissions?"
	if status.Reason != "" {
		reason = status.Reason
	}
	if status.Message != "" {
		message = status.Message
	}
	return errors.Errorf(reason, message)
}

func GetLastTransitionOccurredAt(pod *v1.Pod) metaV1.Time {
	var lastTransitionTime metaV1.Time
	containerStatuses := append(pod.Status.ContainerStatuses, pod.Status.InitContainerStatuses...)
	for _, containerStatus := range containerStatuses {
		if r := containerStatus.LastTerminationState.Running; r != nil {
			if r.StartedAt.Unix() > lastTransitionTime.Unix() {
				lastTransitionTime = r.StartedAt
			}
		} else if r := containerStatus.LastTerminationState.Terminated; r != nil {
			if r.FinishedAt.Unix() > lastTransitionTime.Unix() {
				lastTransitionTime = r.StartedAt
			}
		}
	}

	if lastTransitionTime.IsZero() {
		lastTransitionTime = metaV1.NewTime(time.Now())
	}

	return lastTransitionTime
}

type containerTaskExecutor struct {
}

func (containerTaskExecutor) GetTaskStatus(ctx context.Context, _ types.TaskContext, r flytek8s.K8sResource) (
	types.TaskStatus, *events.TaskEventInfo, error) {

	pod := r.(*v1.Pod)

	var info *events.TaskEventInfo
	if pod.Status.Phase != v1.PodPending && pod.Status.Phase != v1.PodUnknown {
		taskLogs, err := logs.GetLogsForContainerInPod(ctx, pod, 0, " (User)")
		if err != nil {
			return types.TaskStatusUndefined, nil, err
		}

		t := GetLastTransitionOccurredAt(pod).Time
		info = &events.TaskEventInfo{
			Logs:       taskLogs,
			OccurredAt: &t,
		}
	}
	switch pod.Status.Phase {
	case v1.PodSucceeded:
		return types.TaskStatusSucceeded.WithOccurredAt(GetLastTransitionOccurredAt(pod).Time), info, nil
	case v1.PodFailed:
		return types.TaskStatusRetryableFailure(ConvertPodFailureToError(pod.Status)).WithOccurredAt(GetLastTransitionOccurredAt(pod).Time), info, nil
	case v1.PodPending:
		status, err := flytek8s.DemystifyPending(pod.Status)
		return status, info, err
	case v1.PodUnknown:
		return types.TaskStatusUnknown, info, nil
	}
	return types.TaskStatusRunning, info, nil
}

// Creates a new Pod that will Exit on completion. The pods have no retries by design
func (c containerTaskExecutor) BuildResource(ctx context.Context, taskCtx types.TaskContext, task *core.TaskTemplate, inputs *core.LiteralMap) (flytek8s.K8sResource, error) {

	podSpec, err := flytek8s.ToK8sPod(ctx, taskCtx, task.GetContainer(), inputs)
	if err != nil {
		return nil, err
	}

	pod := flytek8s.BuildPodWithSpec(podSpec)
	return pod, nil
}

func (containerTaskExecutor) BuildIdentityResource(_ context.Context, taskCtx types.TaskContext) (flytek8s.K8sResource, error) {
	return flytek8s.BuildIdentityPod(), nil
}

func init() {
	tasksV1.RegisterLoader(func(ctx context.Context) error {
		return tasksV1.K8sRegisterAsDefault(containerTaskType, &v1.Pod{}, flytek8s.DefaultInformerResyncDuration,
			containerTaskExecutor{})
	})
}
