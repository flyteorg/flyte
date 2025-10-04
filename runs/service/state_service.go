package service

import (
	"context"
	"encoding/json"
	"fmt"

	"connectrpc.com/connect"
	"google.golang.org/genproto/googleapis/rpc/status"

	"github.com/flyteorg/flyte/v2/flytestdlib/logger"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/common"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/workflow"
	"github.com/flyteorg/flyte/v2/runs/repository"
)

// StateService implements the StateService gRPC API
type StateService struct {
	repo repository.Repository
}

// NewStateService creates a new StateService
func NewStateService(repo repository.Repository) *StateService {
	return &StateService{
		repo: repo,
	}
}

// Put handles the bidirectional streaming Put RPC
// Client sends PutRequest messages, server responds with PutResponse for each
func (s *StateService) Put(ctx context.Context, stream *connect.BidiStream[workflow.PutRequest, workflow.PutResponse]) error {
	logger.Infof(ctx, "StateService.Put stream started")

	for {
		// Receive request from client
		req, err := stream.Receive()
		if err != nil {
			// Stream closed by client
			logger.Infof(ctx, "StateService.Put stream closed: %v", err)
			return nil
		}

		// Validate request
		if req.ActionId == nil {
			resp := &workflow.PutResponse{
				ActionId: req.ActionId,
				Status: &status.Status{
					Code:    int32(connect.CodeInvalidArgument),
					Message: "action_id is required",
				},
			}
			if err := stream.Send(resp); err != nil {
				return err
			}
			continue
		}

		if req.State == "" {
			resp := &workflow.PutResponse{
				ActionId: req.ActionId,
				Status: &status.Status{
					Code:    int32(connect.CodeInvalidArgument),
					Message: "state is required",
				},
			}
			if err := stream.Send(resp); err != nil {
				return err
			}
			continue
		}

		// Validate that state is valid JSON
		var stateObj map[string]interface{}
		if err := json.Unmarshal([]byte(req.State), &stateObj); err != nil {
			resp := &workflow.PutResponse{
				ActionId: req.ActionId,
				Status: &status.Status{
					Code:    int32(connect.CodeInvalidArgument),
					Message: fmt.Sprintf("state must be valid JSON: %v", err),
				},
			}
			if err := stream.Send(resp); err != nil {
				return err
			}
			continue
		}

		// Update action state in database
		if err := s.repo.UpdateActionState(ctx, req.ActionId, req.State); err != nil {
			logger.Warnf(ctx, "Failed to update action state: %v", err)
			resp := &workflow.PutResponse{
				ActionId: req.ActionId,
				Status: &status.Status{
					Code:    int32(connect.CodeInternal),
					Message: fmt.Sprintf("failed to update state: %v", err),
				},
			}
			if err := stream.Send(resp); err != nil {
				return err
			}
			continue
		}

		// Send notification
		if err := s.repo.NotifyStateUpdate(ctx, req.ActionId); err != nil {
			logger.Warnf(ctx, "Failed to send state update notification: %v", err)
			// Continue anyway - the update was saved
		}

		// Send success response
		resp := &workflow.PutResponse{
			ActionId: req.ActionId,
			Status: &status.Status{
				Code:    0, // OK
				Message: "state updated successfully",
			},
		}

		if err := stream.Send(resp); err != nil {
			return err
		}

		logger.Infof(ctx, "Updated state for action: %s/%s/%s/%s/%s",
			req.ActionId.Run.Org, req.ActionId.Run.Project, req.ActionId.Run.Domain,
			req.ActionId.Run.Name, req.ActionId.Name)
	}
}

// Get handles the bidirectional streaming Get RPC
// Client sends GetRequest messages, server responds with GetResponse for each
func (s *StateService) Get(ctx context.Context, stream *connect.BidiStream[workflow.GetRequest, workflow.GetResponse]) error {
	logger.Infof(ctx, "StateService.Get stream started")

	for {
		// Receive request from client
		req, err := stream.Receive()
		if err != nil {
			// Stream closed by client
			logger.Infof(ctx, "StateService.Get stream closed: %v", err)
			return nil
		}

		// Validate request
		if req.ActionId == nil {
			resp := &workflow.GetResponse{
				ActionId: req.ActionId,
				Status: &status.Status{
					Code:    int32(connect.CodeInvalidArgument),
					Message: "action_id is required",
				},
				State: "",
			}
			if err := stream.Send(resp); err != nil {
				return err
			}
			continue
		}

		// Get action state from database
		state, err := s.repo.GetActionState(ctx, req.ActionId)
		if err != nil {
			logger.Warnf(ctx, "Failed to get action state: %v", err)
			resp := &workflow.GetResponse{
				ActionId: req.ActionId,
				Status: &status.Status{
					Code:    int32(connect.CodeNotFound),
					Message: fmt.Sprintf("failed to get state: %v", err),
				},
				State: "",
			}
			if err := stream.Send(resp); err != nil {
				return err
			}
			continue
		}

		// Send success response
		resp := &workflow.GetResponse{
			ActionId: req.ActionId,
			Status: &status.Status{
				Code:    0, // OK
				Message: "state retrieved successfully",
			},
			State: state,
		}

		if err := stream.Send(resp); err != nil {
			return err
		}

		logger.Infof(ctx, "Retrieved state for action: %s/%s/%s/%s/%s",
			req.ActionId.Run.Org, req.ActionId.Run.Project, req.ActionId.Run.Domain,
			req.ActionId.Run.Name, req.ActionId.Name)
	}
}

// Watch handles the server-side streaming Watch RPC
// Streams ActionUpdate messages for actions matching the filter
func (s *StateService) Watch(ctx context.Context, req *connect.Request[workflow.WatchRequest], stream *connect.ServerStream[workflow.WatchResponse]) error {
	logger.Infof(ctx, "StateService.Watch stream started")

	// Get parent action ID from filter
	var parentActionID *common.ActionIdentifier
	switch filter := req.Msg.Filter.(type) {
	case *workflow.WatchRequest_ParentActionId:
		parentActionID = filter.ParentActionId
	default:
		return connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("filter is required"))
	}

	if parentActionID == nil {
		return connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("parent_action_id is required"))
	}

	// Get all child actions for the parent
	childActions, err := s.getChildActions(ctx, parentActionID)
	if err != nil {
		return connect.NewError(connect.CodeInternal, fmt.Errorf("failed to get child actions: %w", err))
	}

	// Send current state of all actions
	for _, action := range childActions {
		actionUpdate := s.actionToUpdate(action)
		resp := &workflow.WatchResponse{
			Message: &workflow.WatchResponse_ActionUpdate{
				ActionUpdate: actionUpdate,
			},
		}

		if err := stream.Send(resp); err != nil {
			return err
		}
	}

	// Send sentinel message to indicate end of initial state
	sentinelResp := &workflow.WatchResponse{
		Message: &workflow.WatchResponse_ControlMessage{
			ControlMessage: &workflow.ControlMessage{
				Sentinel: true,
			},
		},
	}
	if err := stream.Send(sentinelResp); err != nil {
		return err
	}

	logger.Infof(ctx, "Sent initial state (%d actions) and sentinel for parent action: %s", len(childActions), parentActionID.Name)

	// Watch for state updates
	updates := make(chan *common.ActionIdentifier, 100)
	errs := make(chan error, 1)

	go s.repo.WatchStateUpdates(ctx, updates, errs)

	for {
		select {
		case <-ctx.Done():
			logger.Infof(ctx, "StateService.Watch stream closed by client")
			return nil

		case err := <-errs:
			logger.Errorf(ctx, "Error watching state updates: %v", err)
			return connect.NewError(connect.CodeInternal, err)

		case actionID := <-updates:
			// Filter for actions that are children of the parent
			if !s.isChildOf(actionID, parentActionID) {
				continue
			}

			// Get the full action details
			action, err := s.repo.GetAction(ctx, actionID)
			if err != nil {
				logger.Warnf(ctx, "Failed to get action details: %v", err)
				continue
			}

			// Send action update
			actionUpdate := s.actionToUpdate(action)
			resp := &workflow.WatchResponse{
				Message: &workflow.WatchResponse_ActionUpdate{
					ActionUpdate: actionUpdate,
				},
			}

			if err := stream.Send(resp); err != nil {
				return err
			}

			logger.Debugf(ctx, "Sent action update for: %s", actionID.Name)
		}
	}
}

// Helper functions

// getChildActions retrieves all child actions for a parent action
func (s *StateService) getChildActions(ctx context.Context, parentActionID *common.ActionIdentifier) ([]*repository.Action, error) {
	// For simplicity, we'll list all actions in the run and filter by parent
	// In a production system, you'd add a more efficient query
	runID := parentActionID.Run
	allActions, _, err := s.repo.ListActions(ctx, runID, 1000, "")
	if err != nil {
		return nil, err
	}

	var childActions []*repository.Action
	for _, action := range allActions {
		// Include the parent action itself and all its children
		if action.Name == parentActionID.Name ||
			(action.ParentActionName != nil && *action.ParentActionName == parentActionID.Name) {
			childActions = append(childActions, action)
		}
	}

	return childActions, nil
}

// isChildOf checks if an action is a child of a parent action
func (s *StateService) isChildOf(actionID *common.ActionIdentifier, parentActionID *common.ActionIdentifier) bool {
	// Same run
	if actionID.Run.Org != parentActionID.Run.Org ||
		actionID.Run.Project != parentActionID.Run.Project ||
		actionID.Run.Domain != parentActionID.Run.Domain ||
		actionID.Run.Name != parentActionID.Run.Name {
		return false
	}

	// For now, we'll include all actions in the run
	// In production, you'd check the parent relationship
	return true
}

// actionToUpdate converts a repository Action to an ActionUpdate message
func (s *StateService) actionToUpdate(action *repository.Action) *workflow.ActionUpdate {
	update := &workflow.ActionUpdate{
		ActionId: &common.ActionIdentifier{
			Run: &common.RunIdentifier{
				Org:     action.Org,
				Project: action.Project,
				Domain:  action.Domain,
				Name:    action.RunName,
			},
			Name: action.Name,
		},
		Phase: s.stringToPhase(action.Phase),
	}

	// Add output URI if available
	if action.OutputURI != nil {
		update.OutputUri = *action.OutputURI
	}

	// Add error if available
	if action.ErrorKind != nil && action.ErrorMessage != nil {
		// Note: This is simplified - in production, you'd properly construct ExecutionError
		// from the error fields in the action
	}

	return update
}

// stringToPhase converts a string phase to a Phase enum
func (s *StateService) stringToPhase(phase string) workflow.Phase {
	switch phase {
	case "PHASE_QUEUED":
		return workflow.Phase_PHASE_QUEUED
	case "PHASE_INITIALIZING":
		return workflow.Phase_PHASE_INITIALIZING
	case "PHASE_RUNNING":
		return workflow.Phase_PHASE_RUNNING
	case "PHASE_SUCCEEDED":
		return workflow.Phase_PHASE_SUCCEEDED
	case "PHASE_FAILED":
		return workflow.Phase_PHASE_FAILED
	case "PHASE_ABORTED":
		return workflow.Phase_PHASE_ABORTED
	default:
		return workflow.Phase_PHASE_UNSPECIFIED
	}
}
