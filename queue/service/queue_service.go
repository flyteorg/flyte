package service

import (
	"context"

	"connectrpc.com/connect"

	"github.com/flyteorg/flyte/v2/flytestdlib/logger"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/workflow"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/workflow/workflowconnect"
	"github.com/flyteorg/flyte/v2/queue/k8s"
)

// QueueService implements the QueueServiceHandler interface
type QueueService struct {
	k8sClient QueueClientInterface
}

// NewQueueService creates a new QueueService instance
func NewQueueService(k8sClient *k8s.QueueClient) *QueueService {
	return &QueueService{k8sClient: k8sClient}
}

// NewQueueServiceWithClient creates a new QueueService with a custom client implementation
// This is useful for testing
func NewQueueServiceWithClient(client QueueClientInterface) *QueueService {
	return &QueueService{k8sClient: client}
}

// Ensure we implement the interface
var _ workflowconnect.QueueServiceHandler = (*QueueService)(nil)

// EnqueueAction creates a TaskAction CR in Kubernetes
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

	// Create TaskAction CR in Kubernetes
	if err := s.k8sClient.EnqueueAction(ctx, req.Msg); err != nil {
		logger.Errorf(ctx, "Failed to create TaskAction CR: %v", err)
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&workflow.EnqueueActionResponse{}), nil
}

// AbortQueuedRun deletes all TaskAction CRs for a run
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

	// Delete all TaskAction CRs for this run
	if err := s.k8sClient.AbortQueuedRun(ctx, req.Msg.RunId, req.Msg.Reason); err != nil {
		logger.Errorf(ctx, "Failed to abort queued run: %v", err)
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&workflow.AbortQueuedRunResponse{}), nil
}

// AbortQueuedAction deletes a specific TaskAction CR
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

	// Delete the TaskAction CR
	if err := s.k8sClient.AbortQueuedAction(ctx, req.Msg.ActionId, req.Msg.Reason); err != nil {
		logger.Errorf(ctx, "Failed to abort queued action: %v", err)
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&workflow.AbortQueuedActionResponse{}), nil
}
