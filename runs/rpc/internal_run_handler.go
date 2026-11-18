package rpc

import (
	"context"
	"errors"
	"io"

	"connectrpc.com/connect"
	grpcstatus "google.golang.org/genproto/googleapis/rpc/status"

	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/workflow"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/workflow/workflowconnect"
	"github.com/flyteorg/flyte/v2/runs/service"
)

type internalRunHandler struct{ run *service.RunService }

// NewInternalRunHandler exposes run persistence operations through the internal Connect API.
func NewInternalRunHandler(run *service.RunService) workflowconnect.InternalRunServiceHandler {
	return &internalRunHandler{run: run}
}

// RecordAction records a new action in the database.
func (s *internalRunHandler) RecordAction(
	ctx context.Context,
	req *connect.Request[workflow.RecordActionRequest],
) (*connect.Response[workflow.RecordActionResponse], error) {
	resp := s.recordAction(ctx, req.Msg)
	return connect.NewResponse(resp), nil
}

// RecordActionStream is the bidirectional streaming variant of RecordAction.
func (s *internalRunHandler) RecordActionStream(
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

func (s *internalRunHandler) recordAction(ctx context.Context, req *workflow.RecordActionRequest) *workflow.RecordActionResponse {
	err := s.run.RecordAction(ctx, req)
	st := statusFromError(err)
	return &workflow.RecordActionResponse{
		ActionId: req.GetActionId(),
		Status:   st,
	}
}

// UpdateActionStatus updates the phase of an action.
func (s *internalRunHandler) UpdateActionStatus(
	ctx context.Context,
	req *connect.Request[workflow.UpdateActionStatusRequest],
) (*connect.Response[workflow.UpdateActionStatusResponse], error) {
	resp := s.updateActionStatus(ctx, req.Msg)
	return connect.NewResponse(resp), nil
}

// UpdateActionStatusStream is the bidirectional streaming variant of UpdateActionStatus.
func (s *internalRunHandler) UpdateActionStatusStream(
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

func (s *internalRunHandler) updateActionStatus(ctx context.Context, req *workflow.UpdateActionStatusRequest) *workflow.UpdateActionStatusResponse {
	err := s.run.UpdateActionStatus(ctx, req)
	st := statusFromError(err)
	return &workflow.UpdateActionStatusResponse{
		ActionId: req.GetActionId(),
		Status:   st,
	}
}

// RecordActionEvents records a batch of action events.
func (s *internalRunHandler) RecordActionEvents(
	ctx context.Context,
	req *connect.Request[workflow.RecordActionEventsRequest],
) (*connect.Response[workflow.RecordActionEventsResponse], error) {
	err := s.run.RecordActionEvents(ctx, req.Msg.GetEvents())
	return connect.NewResponse(&workflow.RecordActionEventsResponse{
		Status: statusFromError(err),
	}), nil
}

// RecordActionEventStream is the bidirectional streaming variant of RecordActionEvents.
func (s *internalRunHandler) RecordActionEventStream(
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
		eventErr := s.run.RecordActionEvents(ctx, []*workflow.ActionEvent{event})
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

// statusFromError converts a Go error to a google.rpc.Status proto.
// Returns an OK status (code 0) when err is nil.
func statusFromError(err error) *grpcstatus.Status {
	if err == nil {
		return &grpcstatus.Status{}
	}
	// Unwrap connect errors to get the gRPC code.
	var connectErr *connect.Error
	if errors.As(rpcError(err), &connectErr) {
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
