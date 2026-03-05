package service

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	"github.com/flyteorg/flyte/v2/flytestdlib/logger"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/workflow"
)

// EventsService implements the EventsService gRPC API.
type EventsService struct {
	eventsWorker *EventsWorker
}

func NewEventService(eventsWorker *EventsWorker) *EventsService {
	return &EventsService{eventsWorker: eventsWorker}
}

// Record accepts action events and enqueues them for async forwarding.
func (s *EventsService) Record(ctx context.Context, req *connect.Request[workflow.RecordRequest]) (*connect.Response[workflow.RecordResponse], error) {
	if s.eventsWorker == nil {
		return nil, connect.NewError(connect.CodeFailedPrecondition, fmt.Errorf("events worker is not initialized"))
	}
	if err := req.Msg.Validate(); err != nil {
		logger.Errorf(ctx, "invalid EventsService.Record request: %v", err)
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	// TODO(nary): implement batch processing for events
	for _, event := range req.Msg.GetEvents() {
		if err := s.eventsWorker.Enqueue(event); err != nil {
			logger.Warnf(ctx, "failed to enqueue action event: %v", err)
			return nil, connect.NewError(connect.CodeResourceExhausted, err)
		}
	}

	return connect.NewResponse(&workflow.RecordResponse{}), nil
}
