package service

import (
	"context"
	"fmt"
	"strings"

	"connectrpc.com/connect"

	executorv1 "github.com/flyteorg/flyte/v2/executor/api/v1"
	"github.com/flyteorg/flyte/v2/flytestdlib/logger"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/actions"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/actions/actionsconnect"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/common"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/workflow"
)

// ActionsService implements the ActionsService gRPC API.
// It consolidates the responsibilities of the queue and state services.
type ActionsService struct {
	client ActionsClientInterface
}

// NewActionsService creates a new ActionsService.
func NewActionsService(client ActionsClientInterface) *ActionsService {
	return &ActionsService{client: client}
}

// Ensure we implement the interface.
var _ actionsconnect.ActionsServiceHandler = (*ActionsService)(nil)

// Enqueue queues a new action for execution.
func (s *ActionsService) Enqueue(
	ctx context.Context,
	req *connect.Request[actions.EnqueueRequest],
) (*connect.Response[actions.EnqueueResponse], error) {
	logger.Infof(ctx, "ActionsService.Enqueue called")

	if err := req.Msg.Validate(); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	if err := s.client.Enqueue(ctx, req.Msg.Action, req.Msg.RunSpec); err != nil {
		logger.Errorf(ctx, "Failed to enqueue action: %v", err)
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&actions.EnqueueResponse{}), nil
}

// GetLatestState returns the latest state of an action.
func (s *ActionsService) GetLatestState(
	ctx context.Context,
	req *connect.Request[actions.GetLatestStateRequest],
) (*connect.Response[actions.GetLatestStateResponse], error) {
	logger.Infof(ctx, "ActionsService.GetLatestState called")

	if err := req.Msg.Validate(); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	state, err := s.client.GetState(ctx, req.Msg.ActionId)
	if err != nil {
		logger.Errorf(ctx, "Failed to get state: %v", err)
		return nil, connect.NewError(connect.CodeNotFound, err)
	}

	return connect.NewResponse(&actions.GetLatestStateResponse{State: state}), nil
}

// WatchForUpdates watches for updates to the state of actions.
func (s *ActionsService) WatchForUpdates(
	ctx context.Context,
	req *connect.Request[actions.WatchForUpdatesRequest],
	stream *connect.ServerStream[actions.WatchForUpdatesResponse],
) error {
	logger.Infof(ctx, "ActionsService.WatchForUpdates stream started")

	var parentActionID *common.ActionIdentifier
	switch filter := req.Msg.Filter.(type) {
	case *actions.WatchForUpdatesRequest_ParentActionId:
		parentActionID = filter.ParentActionId
	default:
		return connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("filter is required"))
	}

	if parentActionID == nil {
		return connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("parent_action_id is required"))
	}

	// Subscribe before listing to avoid missing events between snapshot and watch.
	updateCh := s.client.Subscribe(parentActionID.Name)
	defer s.client.Unsubscribe(parentActionID.Name, updateCh)

	// Send initial state snapshot.
	childActions, err := s.client.ListChildActions(ctx, parentActionID)
	if err != nil {
		return connect.NewError(connect.CodeInternal, fmt.Errorf("failed to list child actions: %w", err))
	}

	for _, action := range childActions {
		resp := &actions.WatchForUpdatesResponse{
			Message: &actions.WatchForUpdatesResponse_ActionUpdate{
				ActionUpdate: taskActionToUpdate(action),
			},
		}
		if err := stream.Send(resp); err != nil {
			return err
		}
	}

	// Send sentinel to signal end of initial snapshot.
	sentinel := &actions.WatchForUpdatesResponse{
		Message: &actions.WatchForUpdatesResponse_ControlMessage{
			ControlMessage: &workflow.ControlMessage{Sentinel: true},
		},
	}
	if err := stream.Send(sentinel); err != nil {
		return err
	}

	logger.Infof(ctx, "Sent initial state (%d actions) and sentinel for parent: %s", len(childActions), parentActionID.Name)

	for {
		select {
		case <-ctx.Done():
			logger.Infof(ctx, "ActionsService.WatchForUpdates stream closed by client")
			return nil

		case update, ok := <-updateCh:
			if !ok {
				logger.Infof(ctx, "Update channel closed")
				return nil
			}

			resp := &actions.WatchForUpdatesResponse{
				Message: &actions.WatchForUpdatesResponse_ActionUpdate{
					ActionUpdate: &workflow.ActionUpdate{
						ActionId:  update.ActionID,
						Phase:     stringToPhase(update.Phase),
						OutputUri: update.OutputUri,
					},
				},
			}
			if err := stream.Send(resp); err != nil {
				return err
			}

			logger.Debugf(ctx, "Sent action update for: %s", update.ActionID.Name)
		}
	}
}

// Update updates the status of an action and saves its serialized state.
func (s *ActionsService) Update(
	ctx context.Context,
	req *connect.Request[actions.UpdateRequest],
) (*connect.Response[actions.UpdateResponse], error) {
	logger.Infof(ctx, "ActionsService.Update called")

	if err := req.Msg.Validate(); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	if err := s.client.PutState(ctx, req.Msg.ActionId, req.Msg.Attempt, req.Msg.Status, req.Msg.State); err != nil {
		logger.Errorf(ctx, "Failed to update action: %v", err)
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	// TODO: forward the update to RunService via an internal stream so it can persist to DB.
	// e.g. s.runServiceClient.RecordActionEvent(ctx, req.Msg.ActionId, req.Msg.Attempt, req.Msg.Status)

	return connect.NewResponse(&actions.UpdateResponse{}), nil
}

// Abort aborts a queued or running action, cascading to descendants.
func (s *ActionsService) Abort(
	ctx context.Context,
	req *connect.Request[actions.AbortRequest],
) (*connect.Response[actions.AbortResponse], error) {
	logger.Infof(ctx, "ActionsService.Abort called")

	if err := req.Msg.Validate(); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	if err := s.client.AbortAction(ctx, req.Msg.ActionId, req.Msg.Reason); err != nil {
		logger.Errorf(ctx, "Failed to abort action: %v", err)
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&actions.AbortResponse{}), nil
}

// taskActionToUpdate converts a TaskAction CR to a workflow.ActionUpdate.
func taskActionToUpdate(action *executorv1.TaskAction) *workflow.ActionUpdate {
	return &workflow.ActionUpdate{
		ActionId: &common.ActionIdentifier{
			Run: &common.RunIdentifier{
				Org:     action.Spec.Org,
				Project: action.Spec.Project,
				Domain:  action.Spec.Domain,
				Name:    action.Spec.RunName,
			},
			Name: action.Spec.ActionName,
		},
		Phase:     getPhaseFromConditions(action),
		OutputUri: actionOutputUri(action.Spec.RunOutputBase, action.Spec.ActionName),
	}
}

func actionOutputUri(runOutputBase, actionName string) string {
	if runOutputBase == "" {
		return ""
	}
	return strings.TrimRight(runOutputBase, "/") + "/" + actionName
}

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
