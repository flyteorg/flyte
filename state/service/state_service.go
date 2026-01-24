package service

import (
	"context"
	"encoding/json"
	"fmt"

	"connectrpc.com/connect"
	"google.golang.org/genproto/googleapis/rpc/status"

	executorv1 "github.com/flyteorg/flyte/v2/executor/api/v1"
	"github.com/flyteorg/flyte/v2/flytestdlib/logger"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/common"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/workflow"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/workflow/workflowconnect"
)

// StateService implements the StateService gRPC API using Kubernetes TaskAction CRs
type StateService struct {
	k8sClient StateClientInterface
}

// NewStateService creates a new StateService
func NewStateService(k8sClient StateClientInterface) *StateService {
	return &StateService{
		k8sClient: k8sClient,
	}
}

// Ensure we implement the interface
var _ workflowconnect.StateServiceHandler = (*StateService)(nil)

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

	// Update TaskAction state in Kubernetes
	if err := s.k8sClient.PutState(ctx, msg.ActionId, msg.State); err != nil {
		logger.Warnf(ctx, "Failed to update action state: %v", err)
		return connect.NewResponse(&workflow.PutResponse{
			ActionId: msg.ActionId,
			Status: &status.Status{
				Code:    int32(connect.CodeInternal),
				Message: fmt.Sprintf("failed to update state: %v", err),
			},
		}), nil
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

	// Get action state from Kubernetes
	state, err := s.k8sClient.GetState(ctx, msg.ActionId)
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

	// Get all child actions for the parent and send initial state
	childActions, err := s.k8sClient.ListChildActions(ctx, parentActionID)
	if err != nil {
		return connect.NewError(connect.CodeInternal, fmt.Errorf("failed to get child actions: %w", err))
	}

	// Send current state of all actions
	for _, action := range childActions {
		actionUpdate := taskActionToUpdate(action)
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

	// Subscribe to updates
	updateCh := s.k8sClient.Subscribe()
	defer s.k8sClient.Unsubscribe(updateCh)

	for {
		select {
		case <-ctx.Done():
			logger.Infof(ctx, "StateService.Watch stream closed by client")
			return nil

		case update, ok := <-updateCh:
			if !ok {
				logger.Infof(ctx, "Update channel closed")
				return nil
			}

			// Filter for actions that are children of the parent
			if !isChildOf(update.ActionID, parentActionID) {
				continue
			}

			// Convert update to ActionUpdate message
			actionUpdate := &workflow.ActionUpdate{
				ActionId: update.ActionID,
				Phase:    stringToPhase(update.Phase),
			}

			resp := &workflow.WatchResponse{
				Message: &workflow.WatchResponse_ActionUpdate{
					ActionUpdate: actionUpdate,
				},
			}

			if err := stream.Send(resp); err != nil {
				return err
			}

			logger.Debugf(ctx, "Sent action update for: %s", update.ActionID.Name)
		}
	}
}

// taskActionToUpdate converts a TaskAction to an ActionUpdate message
func taskActionToUpdate(action *executorv1.TaskAction) *workflow.ActionUpdate {
	update := &workflow.ActionUpdate{
		ActionId: &common.ActionIdentifier{
			Run: &common.RunIdentifier{
				Org:     action.Spec.Org,
				Project: action.Spec.Project,
				Domain:  action.Spec.Domain,
				Name:    action.Spec.RunName,
			},
			Name: action.Spec.ActionName,
		},
		Phase: getPhaseFromConditions(action),
	}

	return update
}

// isChildOf checks if an action is a child of a parent action
func isChildOf(actionID *common.ActionIdentifier, parentActionID *common.ActionIdentifier) bool {
	// Must be same run
	if actionID.Run.Org != parentActionID.Run.Org ||
		actionID.Run.Project != parentActionID.Run.Project ||
		actionID.Run.Domain != parentActionID.Run.Domain ||
		actionID.Run.Name != parentActionID.Run.Name {
		return false
	}

	// For now, include all actions in the same run
	// A more sophisticated implementation would check the parent-child relationship
	return true
}

// getPhaseFromConditions extracts the phase from TaskAction conditions
func getPhaseFromConditions(taskAction *executorv1.TaskAction) common.ActionPhase {
	for _, cond := range taskAction.Status.Conditions {
		switch cond.Type {
		case string(executorv1.ConditionTypeSucceeded):
			if cond.Status == "True" {
				return common.ActionPhase_ACTION_PHASE_SUCCEEDED
			}
		case string(executorv1.ConditionTypeFailed):
			if cond.Status == "True" {
				return common.ActionPhase_ACTION_PHASE_FAILED
			}
		case string(executorv1.ConditionTypeProgressing):
			if cond.Status == "True" {
				switch cond.Reason {
				case string(executorv1.ConditionReasonQueued):
					return common.ActionPhase_ACTION_PHASE_QUEUED
				case string(executorv1.ConditionReasonInitializing):
					return common.ActionPhase_ACTION_PHASE_INITIALIZING
				case string(executorv1.ConditionReasonExecuting):
					return common.ActionPhase_ACTION_PHASE_RUNNING
				}
			}
		}
	}
	return common.ActionPhase_ACTION_PHASE_UNSPECIFIED
}

// stringToPhase converts a string phase to a Phase enum
func stringToPhase(phase string) common.ActionPhase {
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
