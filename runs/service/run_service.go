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
	repo        repository.Repository
	queueClient workflowconnect.QueueServiceClient
}

// NewRunService creates a new RunService instance
func NewRunService(repo repository.Repository, queueClient workflowconnect.QueueServiceClient) *RunService {
	return &RunService{
		repo:        repo,
		queueClient: queueClient,
	}
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

	// Enqueue the root action to the queue service
	actionID := &common.ActionIdentifier{
		Run: &common.RunIdentifier{
			Org:     run.Org,
			Project: run.Project,
			Domain:  run.Domain,
			Name:    run.Name,
		},
		Name: run.Name, // For root actions, action name = run name
	}

	// Build EnqueueActionRequest from CreateRunRequest
	enqueueReq := &workflow.EnqueueActionRequest{
		ActionId:      actionID,
		RunSpec:       req.Msg.RunSpec,
		InputUri:      buildInputURI(run),
		RunOutputBase: buildRunOutputBase(run),
	}

	// Set the spec based on the task type in CreateRunRequest
	switch taskSpec := req.Msg.Task.(type) {
	case *workflow.CreateRunRequest_TaskSpec:
		enqueueReq.Spec = &workflow.EnqueueActionRequest_Task{
			Task: &workflow.TaskAction{
				Spec: taskSpec.TaskSpec,
			},
		}
	case *workflow.CreateRunRequest_TaskId:
		enqueueReq.Spec = &workflow.EnqueueActionRequest_Task{
			Task: &workflow.TaskAction{
				Id: taskSpec.TaskId,
			},
		}
	}

	// Call queue service to enqueue the root action
	_, err = s.queueClient.EnqueueAction(ctx, connect.NewRequest(enqueueReq))
	if err != nil {
		logger.Errorf(ctx, "Failed to enqueue root action: %v", err)
		// Note: We don't fail the CreateRun if enqueue fails - the run is already created
		// In production, you might want to mark the run as failed or retry the enqueue
		logger.Warnf(ctx, "Run %s created but failed to enqueue root action", run.Name)
	} else {
		logger.Infof(ctx, "Successfully enqueued root action for run %s", run.Name)
	}

	// Build response (simplified - you'd convert the full Run model)
	resp := &workflow.CreateRunResponse{
		Run: &workflow.Run{
			Action: &workflow.Action{
				Id: actionID,
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
	action, err := s.repo.GetAction(ctx, req.Msg.ActionId)
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
		Inputs:  &task.Inputs{},  // Would populate from action.InputURI
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
					Name: run.Name, // For root actions, action name = run name
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
					Name:    action.GetRunName(),
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
	action, err := s.repo.GetAction(ctx, req.Msg.ActionId)
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

	// Step 1: Send existing runs that match filter
	listReq := s.convertWatchRequestToListRequest(req.Msg)

	runs, _, err := s.repo.ListRuns(ctx, listReq)
	if err != nil {
		logger.Errorf(ctx, "Failed to list runs: %v", err)
		// Continue even if list fails - still watch for new updates
	} else if len(runs) > 0 {
		// Send existing runs
		protoRuns := make([]*workflow.Run, len(runs))
		for i, run := range runs {
			protoRuns[i] = s.convertRunToProto(run)
		}

		if err := stream.Send(&workflow.WatchRunsResponse{
			Runs: protoRuns,
		}); err != nil {
			return err
		}
	}

	// Step 2: Watch for run updates using repository notifications
	// Create channels for receiving updates
	updatesCh := make(chan *repository.Run, 10)
	errsCh := make(chan error, 1)

	// Start watching for updates in a goroutine
	go s.repo.WatchAllRunUpdates(ctx, updatesCh, errsCh)

	for {
		select {
		case <-ctx.Done():
			return nil

		case err := <-errsCh:
			logger.Errorf(ctx, "Error watching runs: %v", err)
			return err

		case run := <-updatesCh:
			// Filter the run based on the watch request criteria
			if !s.runMatchesFilter(run, req.Msg) {
				continue
			}

			// Convert and send the updated run
			protoRun := s.convertRunToProto(run)
			if err := stream.Send(&workflow.WatchRunsResponse{
				Runs: []*workflow.Run{protoRun},
			}); err != nil {
				return err
			}
		}
	}
}

// WatchActions streams action updates for a run
func (s *RunService) WatchActions(
	ctx context.Context,
	req *connect.Request[workflow.WatchActionsRequest],
	stream *connect.ServerStream[workflow.WatchActionsResponse],
) error {
	logger.Infof(ctx, "Received WatchActions request for run: %s", req.Msg.RunId.Name)

	// Step 1: Send existing actions for this run
	actions, _, err := s.repo.ListActions(ctx, req.Msg.RunId, 100, "")
	if err != nil {
		logger.Errorf(ctx, "Failed to list actions: %v", err)
		// Continue even if list fails - still watch for new updates
	} else if len(actions) > 0 {
		// Send existing actions
		enrichedActions := make([]*workflow.EnrichedAction, len(actions))
		for i, action := range actions {
			enrichedActions[i] = s.convertActionToEnrichedProto(action)
		}

		if err := stream.Send(&workflow.WatchActionsResponse{
			EnrichedActions: enrichedActions,
		}); err != nil {
			return err
		}
	}

	// Step 2: Watch for action updates using repository notifications
	// Create channels for receiving updates
	updatesCh := make(chan *repository.Action, 10)
	errsCh := make(chan error, 1)

	// Start watching for updates in a goroutine
	go s.repo.WatchActionUpdates(ctx, req.Msg.RunId, updatesCh, errsCh)

	for {
		select {
		case <-ctx.Done():
			return nil

		case err := <-errsCh:
			logger.Errorf(ctx, "Error watching actions: %v", err)
			return err

		case action := <-updatesCh:
			// Convert and send the updated action
			enrichedAction := s.convertActionToEnrichedProto(action)
			if err := stream.Send(&workflow.WatchActionsResponse{
				EnrichedActions: []*workflow.EnrichedAction{enrichedAction},
			}); err != nil {
				return err
			}
		}
	}
}

// WatchClusterEvents streams cluster events for an action attempt
func (s *RunService) WatchClusterEvents(
	ctx context.Context,
	req *connect.Request[workflow.WatchClusterEventsRequest],
	stream *connect.ServerStream[workflow.WatchClusterEventsResponse],
) error {
	logger.Infof(ctx, "Received WatchClusterEvents request")

	// TODO: Implement cluster events watching
	// Cluster events are now stored in ActionDetails JSON
	// Need to:
	// 1. Get action using s.repo.GetAction(ctx, req.Msg.Id)
	// 2. Unmarshal action.ActionDetails to extract cluster events for the specified attempt
	// 3. Send existing events and watch for updates

	<-ctx.Done()
	return nil
}

// Helper functions

// buildInputURI generates the input URI for the root action
func buildInputURI(run *repository.Run) string {
	// TODO: In production, this should be a real storage path (e.g., s3://bucket/inputs/org/project/domain/run)
	return ""
}

// buildRunOutputBase generates the output base path for the run
func buildRunOutputBase(run *repository.Run) string {
	// TODO: In production, this should be a real storage path (e.g., s3://bucket/outputs/org/project/domain/run)
	return ""
}

// convertRunToProto converts a repository Run to a proto Run
func (s *RunService) convertRunToProto(run *repository.Run) *workflow.Run {
	if run == nil {
		return nil
	}

	// Build the action identifier from the run
	runID := &common.RunIdentifier{
		Org:     run.Org,
		Project: run.Project,
		Domain:  run.Domain,
		Name:    run.Name,
	}

	// Create the root action with status
	action := &workflow.Action{
		Id: &common.ActionIdentifier{
			Run:  runID,
			Name: run.Name, // For root actions, action name = run name
		},
		Metadata: &workflow.ActionMetadata{
			// TODO: Extract from ActionSpec JSON if needed
		},
		Status: &workflow.ActionStatus{
			Phase: workflow.Phase(workflow.Phase_value[run.Phase]),
			// TODO: Extract timestamps, error, etc. from ActionDetails JSON
		},
	}

	return &workflow.Run{
		Action: action,
	}
}

// convertActionToEnrichedProto converts a repository Action to an EnrichedAction proto
func (s *RunService) convertActionToEnrichedProto(action *repository.Action) *workflow.EnrichedAction {
	if action == nil {
		return nil
	}

	// Build the action identifier
	actionID := &common.ActionIdentifier{
		Run: &common.RunIdentifier{
			Org:     action.Org,
			Project: action.Project,
			Domain:  action.Domain,
			Name:    action.GetRunName(),
		},
		Name: action.Name,
	}

	// Build the action status
	actionStatus := &workflow.ActionStatus{
		Phase: workflow.Phase(workflow.Phase_value[action.Phase]),
	}

	// Build the action proto
	actionProto := &workflow.Action{
		Id:     actionID,
		Status: actionStatus,
	}

	return &workflow.EnrichedAction{
		Action:      actionProto,
		MeetsFilter: true, // For now, all actions meet the filter
	}
}

// convertWatchRequestToListRequest converts a WatchRunsRequest to a ListRunsRequest
func (s *RunService) convertWatchRequestToListRequest(req *workflow.WatchRunsRequest) *workflow.ListRunsRequest {
	listReq := &workflow.ListRunsRequest{
		Request: &common.ListRequest{
			Limit: 100,
		},
	}

	// Convert the target filter to the appropriate ListRuns scope
	switch target := req.Target.(type) {
	case *workflow.WatchRunsRequest_Org:
		listReq.ScopeBy = &workflow.ListRunsRequest_Org{
			Org: target.Org,
		}
	case *workflow.WatchRunsRequest_ClusterId:
		// Cluster filtering not directly supported in ListRuns, will filter client-side
		// Could be added to ListRuns if needed
	case *workflow.WatchRunsRequest_ProjectId:
		listReq.ScopeBy = &workflow.ListRunsRequest_ProjectId{
			ProjectId: target.ProjectId,
		}
	case *workflow.WatchRunsRequest_TaskId:
		// Task filtering not directly supported in ListRuns, will filter client-side
		// Could be added to ListRuns if needed
	}

	return listReq
}

// runMatchesFilter checks if a run matches the WatchRunsRequest filter criteria
func (s *RunService) runMatchesFilter(run *repository.Run, req *workflow.WatchRunsRequest) bool {
	if req.Target == nil {
		// No filter, all runs match
		return true
	}

	switch target := req.Target.(type) {
	case *workflow.WatchRunsRequest_Org:
		return run.Org == target.Org

	case *workflow.WatchRunsRequest_ClusterId:
		// TODO: Add cluster field to Run model if needed
		// For now, accept all runs
		return true

	case *workflow.WatchRunsRequest_ProjectId:
		return run.Org == target.ProjectId.Organization &&
			run.Project == target.ProjectId.Name &&
			run.Domain == target.ProjectId.Domain

	case *workflow.WatchRunsRequest_TaskId:
		// TODO: Need to check if the run was triggered by this task
		// This would require storing task_id in the Run model or querying actions
		// For now, accept all runs
		return true

	default:
		// Unknown filter, accept all runs
		return true
	}
}
