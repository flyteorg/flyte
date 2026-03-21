package service

import (
	"context"
	"fmt"

	"connectrpc.com/connect"

	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/common"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/core"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/workflow"
	"github.com/flyteorg/flyte/v2/runs/repository/interfaces"

	"github.com/samber/lo"
)

// LogStreamer abstracts log fetching from different backends.
type LogStreamer interface {
	TailLogs(ctx context.Context, logContext *core.LogContext, stream *connect.ServerStream[workflow.TailLogsResponse]) error
}

// RunLogsService implements the RunLogsServiceHandler interface.
type RunLogsService struct {
	repo     interfaces.Repository
	streamer LogStreamer
}

// NewRunLogsService creates a new RunLogsService.
func NewRunLogsService(repo interfaces.Repository, streamer LogStreamer) *RunLogsService {
	return &RunLogsService{
		repo:     repo,
		streamer: streamer,
	}
}

// TailLogs streams pod logs for an action attempt.
func (s *RunLogsService) TailLogs(ctx context.Context, req *connect.Request[workflow.TailLogsRequest], stream *connect.ServerStream[workflow.TailLogsResponse]) error {
	msg := req.Msg
	if msg.GetActionId() == nil {
		return connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("action_id is required"))
	}
	if msg.GetAttempt() == 0 {
		return connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("attempt must be > 0"))
	}

	logContext, err := getLogContextForAttempt(ctx, s.repo, msg.GetActionId(), msg.GetAttempt())
	if err != nil {
		return err
	}

	return s.streamer.TailLogs(ctx, logContext, stream)
}

// getLogContextForAttempt finds the LogContext from action events for the given attempt.
func getLogContextForAttempt(ctx context.Context, repo interfaces.Repository, actionID *common.ActionIdentifier, attempt uint32) (*core.LogContext, error) {
	events, err := repo.ActionRepo().ListEvents(ctx, actionID, 100)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to list events for action %v: %w", actionID, err))
	}

	// Iterate in reverse to find the latest event with a LogContext for this attempt.
	for i := len(events) - 1; i >= 0; i-- {
		m := events[i]
		if m.Attempt != attempt {
			continue
		}
		event, err := m.ToActionEvent()
		if err != nil {
			continue
		}
		if event.GetLogContext() != nil {
			return event.GetLogContext(), nil
		}
	}

	return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("no log context found for action %v attempt %d", actionID, attempt))
}

// getPrimaryPodAndContainer finds the primary pod and container from a LogContext.
func getPrimaryPodAndContainer(logContext *core.LogContext) (*core.PodLogContext, *core.ContainerContext, error) {
	if logContext.GetPrimaryPodName() == "" {
		return nil, nil, fmt.Errorf("primary pod name is empty in log context")
	}

	pod, found := lo.Find(logContext.GetPods(), func(pod *core.PodLogContext) bool {
		return pod.GetPodName() == logContext.GetPrimaryPodName()
	})
	if !found {
		return nil, nil, fmt.Errorf("primary pod %s not found in log context", logContext.GetPrimaryPodName())
	}

	container, found := lo.Find(pod.GetContainers(), func(c *core.ContainerContext) bool {
		return c.GetContainerName() == pod.GetPrimaryContainerName()
	})
	if !found {
		return nil, nil, fmt.Errorf("primary container %s not found in pod %s", pod.GetPrimaryContainerName(), pod.GetPodName())
	}

	return pod, container, nil
}
