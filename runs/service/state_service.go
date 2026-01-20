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
	"github.com/flyteorg/flyte/v2/runs/repository/interfaces"
	"github.com/flyteorg/flyte/v2/runs/repository/models"
)

// StateService implements the StateService gRPC API
type StateService struct {
	repo interfaces.Repository
}

// NewStateService creates a new StateService
func NewStateService(repo interfaces.Repository) *StateService {
	return &StateService{
		repo: repo,
	}
}

// Put handles the unary Put RPC
// Client sends a single PutRequest, server responds with PutResponse
func (s *StateService) Put(ctx context.Context, req *connect.Request[workflow.PutRequest]) (*connect.Response[workflow.PutResponse], error) {
	logger.Infof(ctx, "StateService.Put called")

	msg := req.Msg

	// Validate request
	if msg.ActionId == nil {
		return connect.NewResponse(&workflow.PutResponse{
			ActionId: msg.ActionId,
			Status: &status.Status{
				Code:    int32(connect.CodeInvalidArgument),
				Message: "action_id is required",
			},
		}), nil
	}

	if msg.State == "" {
		return connect.NewResponse(&workflow.PutResponse{
			ActionId: msg.ActionId,
			Status: &status.Status{
				Code:    int32(connect.CodeInvalidArgument),
				Message: "state is required",
			},
		}), nil
	}

	// Validate that state is valid JSON
	var stateObj map[string]interface{}
	if err := json.Unmarshal([]byte(msg.State), &stateObj); err != nil {
		return connect.NewResponse(&workflow.PutResponse{
			ActionId: msg.ActionId,
			Status: &status.Status{
				Code:    int32(connect.CodeInvalidArgument),
				Message: fmt.Sprintf("state must be valid JSON: %v", err),
			},
		}), nil
	}

	// Update action state in database
	if err := s.repo.ActionRepo().UpdateActionState(ctx, msg.ActionId, msg.State); err != nil {
		logger.Warnf(ctx, "Failed to update action state: %v", err)
		return connect.NewResponse(&workflow.PutResponse{
			ActionId: msg.ActionId,
			Status: &status.Status{
				Code:    int32(connect.CodeInternal),
				Message: fmt.Sprintf("failed to update state: %v", err),
			},
		}), nil
	}

	// Send notification
	if err := s.repo.ActionRepo().NotifyStateUpdate(ctx, msg.ActionId); err != nil {
		logger.Warnf(ctx, "Failed to send state update notification: %v", err)
		// Continue anyway - the update was saved
	}

	logger.Infof(ctx, "Updated state for action: %s/%s/%s/%s/%s",
		msg.ActionId.Run.Org, msg.ActionId.Run.Project, msg.ActionId.Run.Domain,
		msg.ActionId.Run.Name, msg.ActionId.Name)

	// Return success response
	return connect.NewResponse(&workflow.PutResponse{
		ActionId: msg.ActionId,
		Status: &status.Status{
			Code:    0, // OK
			Message: "state updated successfully",
		},
	}), nil
}

// Get handles the unary Get RPC
// Client sends a single GetRequest, server responds with GetResponse
func (s *StateService) Get(ctx context.Context, req *connect.Request[workflow.GetRequest]) (*connect.Response[workflow.GetResponse], error) {
	logger.Infof(ctx, "StateService.Get called")

	msg := req.Msg

	// Validate request
	if msg.ActionId == nil {
		return connect.NewResponse(&workflow.GetResponse{
			ActionId: msg.ActionId,
			Status: &status.Status{
				Code:    int32(connect.CodeInvalidArgument),
				Message: "action_id is required",
			},
			State: "",
		}), nil
	}

	// Get action state from database
	state, err := s.repo.ActionRepo().GetActionState(ctx, msg.ActionId)
	if err != nil {
		logger.Warnf(ctx, "Failed to get action state: %v", err)
		return connect.NewResponse(&workflow.GetResponse{
			ActionId: msg.ActionId,
			Status: &status.Status{
				Code:    int32(connect.CodeNotFound),
				Message: fmt.Sprintf("failed to get state: %v", err),
			},
			State: "",
		}), nil
	}

	logger.Infof(ctx, "Retrieved state for action: %s/%s/%s/%s/%s",
		msg.ActionId.Run.Org, msg.ActionId.Run.Project, msg.ActionId.Run.Domain,
		msg.ActionId.Run.Name, msg.ActionId.Name)

	// Return success response
	return connect.NewResponse(&workflow.GetResponse{
		ActionId: msg.ActionId,
		Status: &status.Status{
			Code:    0, // OK
			Message: "state retrieved successfully",
		},
		State: state,
	}), nil
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

	go s.repo.ActionRepo().WatchStateUpdates(ctx, updates, errs)

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
			action, err := s.repo.ActionRepo().GetAction(ctx, actionID)
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
func (s *StateService) getChildActions(ctx context.Context, parentActionID *common.ActionIdentifier) ([]*models.Action, error) {
	// For simplicity, we'll list all actions in the run and filter by parent
	// In a production system, you'd add a more efficient query
	runID := parentActionID.Run
	allActions, _, err := s.repo.ActionRepo().ListActions(ctx, runID, 1000, "")
	if err != nil {
		return nil, err
	}

	var childActions []*models.Action
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
func (s *StateService) actionToUpdate(action *models.Action) *workflow.ActionUpdate {
	update := &workflow.ActionUpdate{
		ActionId: &common.ActionIdentifier{
			Run: &common.RunIdentifier{
				Org:     action.Org,
				Project: action.Project,
				Domain:  action.Domain,
				Name:    action.GetRunName(),
			},
			Name: action.Name,
		},
		Phase: s.stringToPhase(action.Phase),
	}

	// TODO: Extract output URI and error info from ActionDetails JSON
	// This would require unmarshaling action.ActionDetails and extracting the fields

	return update
}

// stringToPhase converts a string phase to a Phase enum
func (s *StateService) stringToPhase(phase string) common.ActionPhase {
	switch phase {
	case "PHASE_QUEUED":
		return common.ActionPhase_ACTION_PHASE_QUEUED
	case "PHASE_INITIALIZING":
		return common.ActionPhase_ACTION_PHASE_INITIALIZING
	case "PHASE_RUNNING":
		return common.ActionPhase_ACTION_PHASE_RUNNING
	case "PHASE_SUCCEEDED":
		return common.ActionPhase_ACTION_PHASE_SUCCEEDED
	case "PHASE_FAILED":
		return common.ActionPhase_ACTION_PHASE_FAILED
	case "PHASE_ABORTED":
		return common.ActionPhase_ACTION_PHASE_ABORTED
	default:
		return common.ActionPhase_ACTION_PHASE_UNSPECIFIED
	}
}
