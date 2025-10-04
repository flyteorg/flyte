package service

import (
	"context"

	"connectrpc.com/connect"

	"github.com/flyteorg/flyte/v2/flytestdlib/logger"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/workflow"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/workflow/workflowconnect"
	"github.com/flyteorg/flyte/v2/queue/repository"
)

// QueueService implements the QueueServiceHandler interface
type QueueService struct {
	repo repository.Repository
}

// NewQueueService creates a new QueueService instance
func NewQueueService(repo repository.Repository) *QueueService {
	return &QueueService{repo: repo}
}

// Ensure we implement the interface
var _ workflowconnect.QueueServiceHandler = (*QueueService)(nil)

// EnqueueAction queues a new action for execution
func (s *QueueService) EnqueueAction(
	ctx context.Context,
	req *connect.Request[workflow.EnqueueActionRequest],
) (*connect.Response[workflow.EnqueueActionResponse], error) {
	logger.Infof(ctx, "Received EnqueueAction request for action: %s/%s/%s/%s/%s",
		req.Msg.ActionId.Run.Org,
		req.Msg.ActionId.Run.Project,
		req.Msg.ActionId.Run.Domain,
		req.Msg.ActionId.Run.Name,
		req.Msg.ActionId.Name)

	// Validate request using buf validate
	if err := req.Msg.Validate(); err != nil {
		logger.Errorf(ctx, "Invalid EnqueueAction request: %v", err)
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	// Persist to database
	if err := s.repo.EnqueueAction(ctx, req.Msg); err != nil {
		logger.Errorf(ctx, "Failed to enqueue action: %v", err)
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&workflow.EnqueueActionResponse{}), nil
}

// AbortQueuedRun aborts a queued run
func (s *QueueService) AbortQueuedRun(
	ctx context.Context,
	req *connect.Request[workflow.AbortQueuedRunRequest],
) (*connect.Response[workflow.AbortQueuedRunResponse], error) {
	logger.Infof(ctx, "Received AbortQueuedRun request for run: %s/%s/%s/%s",
		req.Msg.RunId.Org,
		req.Msg.RunId.Project,
		req.Msg.RunId.Domain,
		req.Msg.RunId.Name)

	// Validate request
	if err := req.Msg.Validate(); err != nil {
		logger.Errorf(ctx, "Invalid AbortQueuedRun request: %v", err)
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	// Get reason or use default
	reason := "User requested abort"
	if req.Msg.Reason != nil {
		reason = *req.Msg.Reason
	}

	// Abort in database
	if err := s.repo.AbortQueuedRun(ctx, req.Msg.RunId, reason); err != nil {
		logger.Errorf(ctx, "Failed to abort queued run: %v", err)
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&workflow.AbortQueuedRunResponse{}), nil
}

// AbortQueuedAction aborts a queued action
func (s *QueueService) AbortQueuedAction(
	ctx context.Context,
	req *connect.Request[workflow.AbortQueuedActionRequest],
) (*connect.Response[workflow.AbortQueuedActionResponse], error) {
	logger.Infof(ctx, "Received AbortQueuedAction request for action: %s/%s/%s/%s/%s",
		req.Msg.ActionId.Run.Org,
		req.Msg.ActionId.Run.Project,
		req.Msg.ActionId.Run.Domain,
		req.Msg.ActionId.Run.Name,
		req.Msg.ActionId.Name)

	// Validate request
	if err := req.Msg.Validate(); err != nil {
		logger.Errorf(ctx, "Invalid AbortQueuedAction request: %v", err)
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	// Get reason or use default
	reason := "User requested abort"
	if req.Msg.Reason != nil {
		reason = *req.Msg.Reason
	}

	// Abort in database
	if err := s.repo.AbortQueuedAction(ctx, req.Msg.ActionId, reason); err != nil {
		logger.Errorf(ctx, "Failed to abort queued action: %v", err)
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&workflow.AbortQueuedActionResponse{}), nil
}
