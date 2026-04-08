package service

import (
	"context"
	"errors"
	"fmt"

	"database/sql"

	"connectrpc.com/connect"
	"golang.org/x/sync/semaphore"

	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/common"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/core"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/workflow"
	"github.com/flyteorg/flyte/v2/runs/repository/interfaces"

	"github.com/samber/lo"
)

const defaultMaxConcurrentStreams = 100

// LogStreamer abstracts log fetching from different backends.
type LogStreamer interface {
	TailLogs(ctx context.Context, logContext *core.LogContext, stream *connect.ServerStream[workflow.TailLogsResponse]) error
}

// RunLogsService implements the RunLogsServiceHandler interface.
type RunLogsService struct {
	repo     interfaces.Repository
	streamer LogStreamer
	sem      *semaphore.Weighted
}

// NewRunLogsService creates a new RunLogsService.
func NewRunLogsService(repo interfaces.Repository, streamer LogStreamer) *RunLogsService {
	return &RunLogsService{
		repo:     repo,
		streamer: streamer,
		sem:      semaphore.NewWeighted(defaultMaxConcurrentStreams),
	}
}

// TailLogs streams pod logs for an action attempt.
func (s *RunLogsService) TailLogs(ctx context.Context, req *connect.Request[workflow.TailLogsRequest], stream *connect.ServerStream[workflow.TailLogsResponse]) error {
	msg := req.Msg
	if msg.GetActionId() == nil {
		return connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("action_id is required"))
	}
	if !s.sem.TryAcquire(1) {
		return connect.NewError(connect.CodeResourceExhausted, fmt.Errorf("too many concurrent log streams"))
	}
	defer s.sem.Release(1)

	logContext, err := getLogContextForAttempt(ctx, s.repo, msg.GetActionId(), msg.GetAttempt())
	if err != nil {
		return err
	}

	return s.streamer.TailLogs(ctx, logContext, stream)
}

// getLogContextForAttempt fetches the latest event for the given attempt and
// extracts its LogContext. Uses a targeted DB query instead of scanning all events.
func getLogContextForAttempt(ctx context.Context, repo interfaces.Repository, actionID *common.ActionIdentifier, attempt uint32) (*core.LogContext, error) {
	m, err := repo.ActionRepo().GetLatestEventByAttempt(ctx, actionID, attempt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("no event found for action %v attempt %d", actionID, attempt))
		}
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to get event for action %v attempt %d: %w", actionID, attempt, err))
	}

	event, err := m.ToActionEvent()
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to deserialize event: %w", err))
	}

	if event.GetLogContext() == nil {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("no log context found for action %v attempt %d", actionID, attempt))
	}

	return event.GetLogContext(), nil
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
