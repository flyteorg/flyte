package service

import (
	"context"

	"connectrpc.com/connect"

	"github.com/flyteorg/flyte/v2/flytestdlib/logger"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/common"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/task"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/workflow"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/workflow/workflowconnect"
	"github.com/flyteorg/flyte/v2/runs/repository"
)

// RunService implements the RunServiceHandler interface
type RunService struct {
	repo repository.Repository
}

// NewRunService creates a new RunService instance
func NewRunService(repo repository.Repository) *RunService {
	return &RunService{repo: repo}
}

// Ensure we implement the interface
var _ workflowconnect.RunServiceHandler = (*RunService)(nil)

// CreateRun creates a new run
func (s *RunService) CreateRun(
	ctx context.Context,
	req *connect.Request[workflow.CreateRunRequest],
) (*connect.Response[workflow.CreateRunResponse], error) {
	logger.Infof(ctx, "Received CreateRun request")

	// Validate request
	if err := req.Msg.Validate(); err != nil {
		logger.Errorf(ctx, "Invalid CreateRun request: %v", err)
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	// Create run in database
	run, err := s.repo.CreateRun(ctx, req.Msg)
	if err != nil {
		logger.Errorf(ctx, "Failed to create run: %v", err)
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	// Build response (simplified - you'd convert the full Run model)
	resp := &workflow.CreateRunResponse{
		Run: &workflow.Run{
			Action: &workflow.Action{
				Id: &common.ActionIdentifier{
					Run: &common.RunIdentifier{
						Org:     run.Org,
						Project: run.Project,
						Domain:  run.Domain,
						Name:    run.Name,
					},
					Name: run.RootActionName,
				},
			},
		},
	}

	return connect.NewResponse(resp), nil
}

// AbortRun aborts a run
func (s *RunService) AbortRun(
	ctx context.Context,
	req *connect.Request[workflow.AbortRunRequest],
) (*connect.Response[workflow.AbortRunResponse], error) {
	logger.Infof(ctx, "Received AbortRun request for run: %s/%s/%s/%s",
		req.Msg.RunId.Org, req.Msg.RunId.Project, req.Msg.RunId.Domain, req.Msg.RunId.Name)

	// Validate request
	if err := req.Msg.Validate(); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	reason := "User requested abort"
	if req.Msg.Reason != nil {
		reason = *req.Msg.Reason
	}

	// Abort in database
	if err := s.repo.AbortRun(ctx, req.Msg.RunId, reason, nil); err != nil {
		logger.Errorf(ctx, "Failed to abort run: %v", err)
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&workflow.AbortRunResponse{}), nil
}

// GetRunDetails gets detailed information about a run
func (s *RunService) GetRunDetails(
	ctx context.Context,
	req *connect.Request[workflow.GetRunDetailsRequest],
) (*connect.Response[workflow.GetRunDetailsResponse], error) {
	logger.Infof(ctx, "Received GetRunDetails request")

	// Validate request
	if err := req.Msg.Validate(); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	// Get run from database
	run, err := s.repo.GetRun(ctx, req.Msg.RunId)
	if err != nil {
		logger.Errorf(ctx, "Failed to get run: %v", err)
		return nil, connect.NewError(connect.CodeNotFound, err)
	}

	// TODO: Build full RunDetails from the run model
	// For now, return a minimal response
	resp := &workflow.GetRunDetailsResponse{
		Details: &workflow.RunDetails{
			// Would populate this from run model
		},
	}

	logger.Infof(ctx, "Retrieved run details for: %s", run.Name)
	return connect.NewResponse(resp), nil
}

// GetActionDetails gets detailed information about an action
func (s *RunService) GetActionDetails(
	ctx context.Context,
	req *connect.Request[workflow.GetActionDetailsRequest],
) (*connect.Response[workflow.GetActionDetailsResponse], error) {
	logger.Infof(ctx, "Received GetActionDetails request")

	// Validate request
	if err := req.Msg.Validate(); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	// Get action from database
	action, err := s.repo.GetActionWithAttempts(ctx, req.Msg.ActionId)
	if err != nil {
		logger.Errorf(ctx, "Failed to get action: %v", err)
		return nil, connect.NewError(connect.CodeNotFound, err)
	}

	// TODO: Build full ActionDetails from the action model
	resp := &workflow.GetActionDetailsResponse{
		Details: &workflow.ActionDetails{
			// Would populate this from action model
		},
	}

	logger.Infof(ctx, "Retrieved action details for: %s", action.Name)
	return connect.NewResponse(resp), nil
}

// GetActionData gets input and output data for an action
func (s *RunService) GetActionData(
	ctx context.Context,
	req *connect.Request[workflow.GetActionDataRequest],
) (*connect.Response[workflow.GetActionDataResponse], error) {
	logger.Infof(ctx, "Received GetActionData request")

	// Validate request
	if err := req.Msg.Validate(); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	// Get action from database
	action, err := s.repo.GetAction(ctx, req.Msg.ActionId)
	if err != nil {
		logger.Errorf(ctx, "Failed to get action: %v", err)
		return nil, connect.NewError(connect.CodeNotFound, err)
	}

	// Return URIs only (not actual data)
	resp := &workflow.GetActionDataResponse{
		Inputs:  &task.Inputs{}, // Would populate from action.InputURI
		Outputs: &task.Outputs{}, // Would populate from outputs
	}

	logger.Infof(ctx, "Retrieved action data for: %s", action.Name)
	return connect.NewResponse(resp), nil
}

// ListRuns lists runs based on filter criteria
func (s *RunService) ListRuns(
	ctx context.Context,
	req *connect.Request[workflow.ListRunsRequest],
) (*connect.Response[workflow.ListRunsResponse], error) {
	logger.Infof(ctx, "Received ListRuns request")

	// Validate request
	if err := req.Msg.Validate(); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	// List runs from database
	runs, nextToken, err := s.repo.ListRuns(ctx, req.Msg)
	if err != nil {
		logger.Errorf(ctx, "Failed to list runs: %v", err)
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	// Convert to proto format
	protoRuns := make([]*workflow.Run, len(runs))
	for i, run := range runs {
		protoRuns[i] = &workflow.Run{
			Action: &workflow.Action{
				Id: &common.ActionIdentifier{
					Run: &common.RunIdentifier{
						Org:     run.Org,
						Project: run.Project,
						Domain:  run.Domain,
						Name:    run.Name,
					},
					Name: run.RootActionName,
				},
			},
		}
	}

	resp := &workflow.ListRunsResponse{
		Runs:  protoRuns,
		Token: nextToken,
	}

	logger.Infof(ctx, "Listed %d runs", len(runs))
	return connect.NewResponse(resp), nil
}

// ListActions lists actions for a run
func (s *RunService) ListActions(
	ctx context.Context,
	req *connect.Request[workflow.ListActionsRequest],
) (*connect.Response[workflow.ListActionsResponse], error) {
	logger.Infof(ctx, "Received ListActions request")

	// Validate request
	if err := req.Msg.Validate(); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	// List actions from database
	limit := 50
	if req.Msg.Request != nil && req.Msg.Request.Limit > 0 {
		limit = int(req.Msg.Request.Limit)
	}

	actions, nextToken, err := s.repo.ListActions(ctx, req.Msg.RunId, limit, "")
	if err != nil {
		logger.Errorf(ctx, "Failed to list actions: %v", err)
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	// Convert to proto format (simplified)
	protoActions := make([]*workflow.Action, len(actions))
	for i, action := range actions {
		protoActions[i] = &workflow.Action{
			Id: &common.ActionIdentifier{
				Run: &common.RunIdentifier{
					Org:     action.Org,
					Project: action.Project,
					Domain:  action.Domain,
					Name:    action.RunName,
				},
				Name: action.Name,
			},
		}
	}

	resp := &workflow.ListActionsResponse{
		Actions: protoActions,
		Token:   nextToken,
	}

	logger.Infof(ctx, "Listed %d actions", len(actions))
	return connect.NewResponse(resp), nil
}

// AbortAction aborts a specific action
func (s *RunService) AbortAction(
	ctx context.Context,
	req *connect.Request[workflow.AbortActionRequest],
) (*connect.Response[workflow.AbortActionResponse], error) {
	logger.Infof(ctx, "Received AbortAction request")

	// Validate request
	if err := req.Msg.Validate(); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	reason := "User requested abort"
	if req.Msg.Reason != "" {
		reason = req.Msg.Reason
	}

	// Abort in database
	if err := s.repo.AbortAction(ctx, req.Msg.ActionId, reason, nil); err != nil {
		logger.Errorf(ctx, "Failed to abort action: %v", err)
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&workflow.AbortActionResponse{}), nil
}

// Streaming RPCs (simplified implementations)

// WatchRunDetails streams run details updates
func (s *RunService) WatchRunDetails(
	ctx context.Context,
	req *connect.Request[workflow.WatchRunDetailsRequest],
	stream *connect.ServerStream[workflow.WatchRunDetailsResponse],
) error {
	logger.Infof(ctx, "Received WatchRunDetails request")

	// For now, just send initial state and close
	// TODO: Implement actual streaming with polling or database triggers
	run, err := s.repo.GetRun(ctx, req.Msg.RunId)
	if err != nil {
		return connect.NewError(connect.CodeNotFound, err)
	}

	resp := &workflow.WatchRunDetailsResponse{
		Details: &workflow.RunDetails{
			// Would populate from run model
		},
	}

	if err := stream.Send(resp); err != nil {
		return err
	}

	logger.Infof(ctx, "Sent initial run details for: %s", run.Name)

	// Keep connection open and send updates (simplified)
	updates := make(chan *repository.Run)
	errs := make(chan error)

	go s.repo.WatchRunUpdates(ctx, req.Msg.RunId, updates, errs)

	for {
		select {
		case <-ctx.Done():
			return nil
		case err := <-errs:
			return connect.NewError(connect.CodeInternal, err)
		case run := <-updates:
			resp := &workflow.WatchRunDetailsResponse{
				Details: &workflow.RunDetails{
					// Would populate from run
				},
			}
			if err := stream.Send(resp); err != nil {
				return err
			}
			logger.Infof(ctx, "Sent run update for: %s", run.Name)
		}
	}
}

// WatchActionDetails streams action details updates
func (s *RunService) WatchActionDetails(
	ctx context.Context,
	req *connect.Request[workflow.WatchActionDetailsRequest],
	stream *connect.ServerStream[workflow.WatchActionDetailsResponse],
) error {
	logger.Infof(ctx, "Received WatchActionDetails request")

	// Send initial state
	action, err := s.repo.GetActionWithAttempts(ctx, req.Msg.ActionId)
	if err != nil {
		return connect.NewError(connect.CodeNotFound, err)
	}

	resp := &workflow.WatchActionDetailsResponse{
		Details: &workflow.ActionDetails{
			// Would populate from action model
		},
	}

	if err := stream.Send(resp); err != nil {
		return err
	}

	logger.Infof(ctx, "Sent initial action details for: %s", action.Name)

	// TODO: Implement actual streaming
	<-ctx.Done()
	return nil
}

// WatchRuns streams run updates based on filter criteria
func (s *RunService) WatchRuns(
	ctx context.Context,
	req *connect.Request[workflow.WatchRunsRequest],
	stream *connect.ServerStream[workflow.WatchRunsResponse],
) error {
	logger.Infof(ctx, "Received WatchRuns request")

	// TODO: Implement actual streaming with filtering
	<-ctx.Done()
	return nil
}

// WatchActions streams action updates for a run
func (s *RunService) WatchActions(
	ctx context.Context,
	req *connect.Request[workflow.WatchActionsRequest],
	stream *connect.ServerStream[workflow.WatchActionsResponse],
) error {
	logger.Infof(ctx, "Received WatchActions request")

	// TODO: Implement actual streaming
	<-ctx.Done()
	return nil
}

// WatchClusterEvents streams cluster events for an action attempt
func (s *RunService) WatchClusterEvents(
	ctx context.Context,
	req *connect.Request[workflow.WatchClusterEventsRequest],
	stream *connect.ServerStream[workflow.WatchClusterEventsResponse],
) error {
	logger.Infof(ctx, "Received WatchClusterEvents request")

	// Get existing cluster events
	events, err := s.repo.GetClusterEvents(ctx, req.Msg.Id, uint(req.Msg.Attempt))
	if err != nil {
		return connect.NewError(connect.CodeNotFound, err)
	}

	// Send existing events
	// TODO: Convert repository events to proto format
	_ = events

	// TODO: Watch for new events
	<-ctx.Done()
	return nil
}
