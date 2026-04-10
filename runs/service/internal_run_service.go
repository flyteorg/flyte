package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/common"
	"io"
	"time"

	"connectrpc.com/connect"
	grpcstatus "google.golang.org/genproto/googleapis/rpc/status"
	"google.golang.org/protobuf/proto"

	"github.com/flyteorg/flyte/v2/flytestdlib/logger"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/workflow"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/workflow/workflowconnect"
	"github.com/flyteorg/flyte/v2/runs/repository/models"
)

// Ensure RunService implements the InternalRunService handler interface.
var _ workflowconnect.InternalRunServiceHandler = (*RunService)(nil)

// RecordAction records a new action in the database.
func (s *RunService) RecordAction(
	ctx context.Context,
	req *connect.Request[workflow.RecordActionRequest],
) (*connect.Response[workflow.RecordActionResponse], error) {
	resp := s.recordAction(ctx, req.Msg)
	return connect.NewResponse(resp), nil
}

// RecordActionStream is the bidirectional streaming variant of RecordAction.
func (s *RunService) RecordActionStream(
	ctx context.Context,
	stream *connect.BidiStream[workflow.RecordActionStreamRequest, workflow.RecordActionStreamResponse],
) error {
	for {
		req, err := stream.Receive()
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return err
		}
		resp := s.recordAction(ctx, req.GetRequest())
		if err := stream.Send(&workflow.RecordActionStreamResponse{Response: resp}); err != nil {
			return err
		}
	}
}

func (s *RunService) recordAction(ctx context.Context, req *workflow.RecordActionRequest) *workflow.RecordActionResponse {
	err := s.recordSingleAction(ctx, req)
	st := statusFromError(err)
	return &workflow.RecordActionResponse{
		ActionId: req.GetActionId(),
		Status:   st,
	}
}

func (s *RunService) recordSingleAction(ctx context.Context, req *workflow.RecordActionRequest) error {
	actionID := req.GetActionId()

	// Build the base action model with fields common to all action types.
	action := models.NewActionModel(actionID)
	if req.GetParent() != "" {
		action.ParentActionName = sql.NullString{String: req.GetParent(), Valid: true}
	}

	action.ActionGroup = sql.NullString{String: req.GetGroup(), Valid: req.GetGroup() != ""}
	action.ActionDetails = []byte("{}")
	action.Attempts = 1

	// Build the ActionSpec proto (stored as bytes for later deserialization).
	spec := &workflow.ActionSpec{
		ActionId: actionID,
		InputUri: req.GetInputUri(),
		Group:    req.GetGroup(),
	}
	if req.GetParent() != "" {
		parent := req.GetParent()
		spec.ParentActionName = &parent
	}

	// Build RunInfo for storage URIs and task spec digest.
	info := &workflow.RunInfo{
		InputsUri: req.GetInputUri(),
	}

	switch v := req.GetSpec().(type) {
	case *workflow.RecordActionRequest_Task:
		taskAction := v.Task
		spec.Spec = &workflow.ActionSpec_Task{Task: taskAction}

		action.ActionType = int32(workflow.ActionType_ACTION_TYPE_TASK)

		taskID := taskAction.GetId()
		if taskID != nil {
			action.TaskProject = sql.NullString{String: taskID.GetProject(), Valid: taskID.GetProject() != ""}
			action.TaskDomain = sql.NullString{String: taskID.GetDomain(), Valid: taskID.GetDomain() != ""}
			action.TaskName = sql.NullString{String: taskID.GetName(), Valid: taskID.GetName() != ""}
			action.TaskVersion = sql.NullString{String: taskID.GetVersion(), Valid: taskID.GetVersion() != ""}
			action.FunctionName = taskID.GetName()
			action.TaskShortName = sql.NullString{String: taskID.GetName(), Valid: taskID.GetName() != ""}
		}
		if taskSpec := taskAction.GetSpec(); taskSpec != nil {
			action.TaskType = taskSpec.GetTaskTemplate().GetType()
			if taskSpec.GetShortName() != "" {
				action.TaskShortName = sql.NullString{String: taskSpec.GetShortName(), Valid: true}
			}
			if env := taskSpec.GetEnvironment(); env != nil && env.GetName() != "" {
				action.EnvironmentName = sql.NullString{String: env.GetName(), Valid: true}
			}

			taskSpecModel, err := models.NewTaskSpecModel(ctx, taskSpec)
			if err != nil {
				logger.Warnf(ctx, "RecordAction: failed to create task spec model: %v", err)
				return connect.NewError(connect.CodeInternal, err)
			}
			if err := s.repo.TaskRepo().CreateTaskSpec(ctx, taskSpecModel); err != nil {
				logger.Warnf(ctx, "RecordAction: failed to store task spec: %v", err)
				// Non-fatal: continue without digest
			} else {
				info.TaskSpecDigest = taskSpecModel.Digest
			}
		}

	case *workflow.RecordActionRequest_Trace:
		spec.Spec = &workflow.ActionSpec_Trace{Trace: v.Trace}

		action.ActionType = int32(workflow.ActionType_ACTION_TYPE_TRACE)
		action.TaskName = sql.NullString{String: v.Trace.GetName(), Valid: v.Trace.GetName() != ""}
		action.FunctionName = v.Trace.GetName()
		action.Phase = int32(common.ActionPhase_ACTION_PHASE_SUCCEEDED)

		// Use the trace's start_time as created_at so that parent traces
		// (which start earlier but are written to DB later) sort before their
		// children when ordered by created_at ASC.
		if st := v.Trace.GetStartTime(); st != nil {
			action.CreatedAt = st.AsTime()
		}
		if et := v.Trace.GetEndTime(); et != nil {
			action.EndedAt = sql.NullTime{Time: et.AsTime(), Valid: true}
		}

		if traceSpec := v.Trace.GetSpec(); traceSpec != nil {
			traceSpecModel, err := models.NewTaskSpecModelFromTraceSpec(ctx, traceSpec)
			if err != nil {
				logger.Warnf(ctx, "RecordAction: failed to create trace spec model: %v", err)
			} else if traceSpecModel != nil {
				if err := s.repo.TaskRepo().CreateTaskSpec(ctx, traceSpecModel); err != nil {
					logger.Warnf(ctx, "RecordAction: failed to store trace spec: %v", err)
				} else {
					info.TaskSpecDigest = traceSpecModel.Digest
				}
			}
		}
		if v.Trace.GetOutputs().GetOutputUri() != "" {
			info.OutputsUri = v.Trace.GetOutputs().GetOutputUri()
		}

	default:
		return connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("unsupported action spec type: %T", req.GetSpec()))
	}

	// Marshal ActionSpec proto into bytes for storage.
	actionSpecBytes, err := proto.Marshal(spec)
	if err != nil {
		logger.Warnf(ctx, "RecordAction: failed to marshal action spec: %v", err)
		return connect.NewError(connect.CodeInternal, err)
	}
	action.ActionSpec = actionSpecBytes

	// Marshal RunInfo proto into bytes for storage.
	detailedInfo, err := proto.Marshal(info)
	if err != nil {
		logger.Warnf(ctx, "RecordAction: failed to marshal run info: %v", err)
		return connect.NewError(connect.CodeInternal, err)
	}
	action.DetailedInfo = detailedInfo

	if _, err := s.repo.ActionRepo().CreateAction(ctx, action, false); err != nil {
		logger.Warnf(ctx, "RecordAction: failed to create action %s: %v", actionID.GetName(), err)
		return connect.NewError(connect.CodeInternal, err)
	}
	return nil
}

// UpdateActionStatus updates the phase of an action.
func (s *RunService) UpdateActionStatus(
	ctx context.Context,
	req *connect.Request[workflow.UpdateActionStatusRequest],
) (*connect.Response[workflow.UpdateActionStatusResponse], error) {
	resp := s.updateActionStatus(ctx, req.Msg)
	return connect.NewResponse(resp), nil
}

// UpdateActionStatusStream is the bidirectional streaming variant of UpdateActionStatus.
func (s *RunService) UpdateActionStatusStream(
	ctx context.Context,
	stream *connect.BidiStream[workflow.UpdateActionStatusStreamRequest, workflow.UpdateActionStatusStreamResponse],
) error {
	for {
		req, err := stream.Receive()
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return err
		}
		resp := s.updateActionStatus(ctx, req.GetRequest())
		if err := stream.Send(&workflow.UpdateActionStatusStreamResponse{
			Response: resp,
			Nonce:    req.GetNonce(),
		}); err != nil {
			return err
		}
	}
}

func (s *RunService) updateActionStatus(ctx context.Context, req *workflow.UpdateActionStatusRequest) *workflow.UpdateActionStatusResponse {
	err := s.updateSingleActionStatus(ctx, req)
	st := statusFromError(err)
	return &workflow.UpdateActionStatusResponse{
		ActionId: req.GetActionId(),
		Status:   st,
	}
}

func (s *RunService) updateSingleActionStatus(ctx context.Context, req *workflow.UpdateActionStatusRequest) error {
	actionStatus := req.GetStatus()
	var endTime *time.Time
	if actionStatus.GetEndTime() != nil {
		t := actionStatus.GetEndTime().AsTime()
		endTime = &t
	} else if IsTerminalPhase(actionStatus.GetPhase()) {
		// If no end time is provided but the phase is terminal, use now.
		t := time.Now()
		endTime = &t
	}

	if err := s.repo.ActionRepo().UpdateActionPhase(
		ctx,
		req.GetActionId(),
		actionStatus.GetPhase(),
		actionStatus.GetAttempts(),
		actionStatus.GetCacheStatus(),
		endTime,
	); err != nil {
		logger.Warnf(ctx, "UpdateActionStatus: failed to update action %s: %v", req.GetActionId().GetName(), err)
		return connect.NewError(connect.CodeInternal, err)
	}
	return nil
}

// RecordActionEvents records a batch of action events.
func (s *RunService) RecordActionEvents(
	ctx context.Context,
	req *connect.Request[workflow.RecordActionEventsRequest],
) (*connect.Response[workflow.RecordActionEventsResponse], error) {
	err := s.recordEvents(ctx, req.Msg.GetEvents())
	return connect.NewResponse(&workflow.RecordActionEventsResponse{
		Status: statusFromError(err),
	}), nil
}

// RecordActionEventStream is the bidirectional streaming variant of RecordActionEvents.
func (s *RunService) RecordActionEventStream(
	ctx context.Context,
	stream *connect.BidiStream[workflow.RecordActionEventStreamRequest, workflow.RecordActionEventStreamResponse],
) error {
	for {
		req, err := stream.Receive()
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return err
		}
		event := req.GetRequest().GetEvent()
		eventErr := s.recordEvents(ctx, []*workflow.ActionEvent{event})
		if err := stream.Send(&workflow.RecordActionEventStreamResponse{
			Response: &workflow.RecordActionEventResponse{
				ActionId: event.GetId(),
				Status:   statusFromError(eventErr),
			},
			Nonce: req.GetNonce(),
		}); err != nil {
			return err
		}
	}
}

// recordEvents inserts ActionEvents into the dedicated table and updates the action phase.
func (s *RunService) recordEvents(ctx context.Context, events []*workflow.ActionEvent) error {
	eventModels := make([]*models.ActionEvent, 0, len(events))
	for _, event := range events {
		m, err := models.NewActionEventModel(event)
		if err != nil {
			logger.Warnf(ctx, "RecordActionEvents: failed to build event model for %s: %v", event.GetId().GetName(), err)
			return connect.NewError(connect.CodeInternal, err)
		}
		eventModels = append(eventModels, m)
	}
	if err := s.repo.ActionRepo().InsertEvents(ctx, eventModels); err != nil {
		logger.Warnf(ctx, "RecordActionEvents: failed to insert events: %v", err)
		return connect.NewError(connect.CodeInternal, err)
	}
	return nil
}

// statusFromError converts a Go error to a google.rpc.Status proto.
// Returns an OK status (code 0) when err is nil.
func statusFromError(err error) *grpcstatus.Status {
	if err == nil {
		return &grpcstatus.Status{}
	}
	// Unwrap connect errors to get the gRPC code.
	var connectErr *connect.Error
	if errors.As(err, &connectErr) {
		return &grpcstatus.Status{
			Code:    int32(connectErr.Code()),
			Message: connectErr.Message(),
		}
	}
	return &grpcstatus.Status{
		Code:    int32(connect.CodeInternal),
		Message: err.Error(),
	}
}
