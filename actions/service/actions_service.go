package service

import (
	"context"
	"errors"
	"fmt"

	"connectrpc.com/connect/v2"

	"github.com/flyteorg/flyte/v2/actions/k8s"
	executorv1 "github.com/flyteorg/flyte/v2/executor/api/v1"
	"github.com/flyteorg/flyte/v2/flytestdlib/logger"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/actions"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/actions/actionsconnect"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/common"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/core"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/workflow"
)

// ActionsService implements the ActionsService gRPC API.
type ActionsService struct {
	client ActionsClientInterface
}

// Signal delivers a value to a paused condition action.
func (s *ActionsService) Signal(ctx context.Context, req *actions.SignalRequest) (*actions.SignalResponse, error) {
	logger.Infof(ctx, "ActionsService.Signal called")

	if err := req.Validate(); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err.Error()).WithCause(err)
	}

	if err := s.client.Signal(ctx, req.ActionId, req.Value, principalSubject(req.GetSignalledBy())); err != nil {
		logger.Errorf(ctx, "Failed to signal action: %v", err)
		return nil, toConnectError(err)
	}

	return &actions.SignalResponse{}, nil
}

// principalSubject extracts a storable subject string from the caller identity.
func principalSubject(id *common.EnrichedIdentity) string {
	if s := id.GetUser().GetId().GetSubject(); s != "" {
		return s
	}
	return id.GetApplication().GetId().GetSubject()
}

// toConnectError preserves typed errors from the client (e.g. NotFound,
// InvalidArgument) and wraps everything else as Internal.
func toConnectError(err error) error {
	var connectErr *connect.Error
	if errors.As(err, &connectErr) {
		return err
	}
	return connect.NewError(connect.CodeInternal, err.Error()).WithCause(err)
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
	req *actions.EnqueueRequest,
) (*actions.EnqueueResponse, error) {
	logger.Infof(ctx, "ActionsService.Enqueue called")

	if err := req.Validate(); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err.Error()).WithCause(err)
	}

	if err := s.client.Enqueue(ctx, req.Action, req.RunSpec); err != nil {
		logger.Errorf(ctx, "Failed to enqueue action: %v", err)
		return nil, toConnectError(err)
	}

	return &actions.EnqueueResponse{}, nil
}

// WatchForUpdates watches for updates to the state of actions.
func (s *ActionsService) WatchForUpdates(
	ctx context.Context,
	req *actions.WatchForUpdatesRequest,
	stream actionsconnect.ActionsServiceWatchForUpdatesServerStream,
) error {
	logger.Infof(ctx, "ActionsService.WatchForUpdates stream started")

	var parentActionID *common.ActionIdentifier
	switch filter := req.Filter.(type) {
	case *actions.WatchForUpdatesRequest_ParentActionId:
		parentActionID = filter.ParentActionId
	default:
		return connect.Errorf(connect.CodeInvalidArgument, "filter is required")
	}

	if parentActionID == nil {
		return connect.Errorf(connect.CodeInvalidArgument, "parent_action_id is required")
	}

	// Subscribe before listing to avoid missing events between snapshot and watch.
	updateCh := s.client.Subscribe(parentActionID.GetRun().GetName(), parentActionID.Name)
	defer s.client.Unsubscribe(parentActionID.GetRun().GetName(), parentActionID.Name, updateCh)

	// Send initial state snapshot.
	childActions, err := s.client.ListChildActions(ctx, parentActionID)
	if err != nil {
		return connect.NewError(connect.CodeInternal, fmt.Errorf("failed to list child actions: %w", err).Error())
	}

	for _, action := range childActions {
		resp := &actions.WatchForUpdatesResponse{
			Message: &actions.WatchForUpdatesResponse_ActionUpdate{
				ActionUpdate: taskActionToUpdate(ctx, action),
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

			au := &workflow.ActionUpdate{
				ActionId:  update.ActionID,
				Phase:     update.Phase,
				OutputUri: update.OutputUri,
				Value:     update.SignalValue,
			}
			if (update.Phase == common.ActionPhase_ACTION_PHASE_FAILED ||
				update.Phase == common.ActionPhase_ACTION_PHASE_TIMED_OUT) &&
				update.ErrorState != nil {
				au.Error = errorStateToExecutionError(update.ErrorState)
			}
			resp := &actions.WatchForUpdatesResponse{
				Message: &actions.WatchForUpdatesResponse_ActionUpdate{
					ActionUpdate: au,
				},
			}
			if err := stream.Send(resp); err != nil {
				return err
			}

			logger.Debugf(ctx, "Sent action update for: %s", update.ActionID.Name)
		}
	}
}

// Update updates the status of an action.
func (s *ActionsService) Update(
	ctx context.Context,
	req *actions.UpdateRequest,
) (*actions.UpdateResponse, error) {
	logger.Infof(ctx, "ActionsService.Update called")

	if err := req.Validate(); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err.Error()).WithCause(err)
	}

	if err := s.client.PutStatus(ctx, req.ActionId, req.Attempt, req.Status); err != nil {
		logger.Errorf(ctx, "Failed to update action: %v", err)
		return nil, connect.NewError(connect.CodeInternal, err.Error()).WithCause(err)
	}

	return &actions.UpdateResponse{}, nil
}

// Abort aborts a queued or running action, cascading to descendants.
func (s *ActionsService) Abort(
	ctx context.Context,
	req *actions.AbortRequest,
) (*actions.AbortResponse, error) {
	logger.Infof(ctx, "ActionsService.Abort called")

	if err := req.Validate(); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err.Error()).WithCause(err)
	}

	if err := s.client.AbortAction(ctx, req.ActionId, req.Reason); err != nil {
		logger.Errorf(ctx, "Failed to abort action: %v", err)
		return nil, connect.NewError(connect.CodeInternal, err.Error()).WithCause(err)
	}

	return &actions.AbortResponse{}, nil
}

// taskActionToUpdate converts a TaskAction CR to a workflow.ActionUpdate.
// The signal value matters here for the recovery case: a parent that
// reconnects after its condition already resolved gets the value from the
// snapshot rather than a live event.
func taskActionToUpdate(ctx context.Context, action *executorv1.TaskAction) *workflow.ActionUpdate {
	phase := k8s.GetPhaseFromConditions(action)
	update := &workflow.ActionUpdate{
		ActionId: &common.ActionIdentifier{
			Run: &common.RunIdentifier{
				Project: action.Spec.Project,
				Domain:  action.Spec.Domain,
				Name:    action.Spec.RunName,
			},
			Name: action.Spec.ActionName,
		},
		Phase:     phase,
		OutputUri: k8s.BuildOutputUri(ctx, action),
		Value:     k8s.SignalValueFromStatus(ctx, action),
	}
	if (phase == common.ActionPhase_ACTION_PHASE_FAILED ||
		phase == common.ActionPhase_ACTION_PHASE_TIMED_OUT) &&
		action.Status.ErrorState != nil {
		update.Error = errorStateToExecutionError(action.Status.ErrorState)
	}
	return update
}

func errorStateToExecutionError(es *executorv1.ErrorState) *core.ExecutionError {
	kind := core.ExecutionError_UNKNOWN
	switch es.Kind {
	case "USER":
		kind = core.ExecutionError_USER
	case "SYSTEM":
		kind = core.ExecutionError_SYSTEM
	}
	return &core.ExecutionError{
		Code:    es.Code,
		Kind:    kind,
		Message: es.Message,
	}
}
