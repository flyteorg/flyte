package service

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/workflow"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/workflow/workflowconnect"
)

// EventsProxyService implements the EventsProxyService gRPC API.
type EventsProxyService struct {
	runClient workflowconnect.InternalRunServiceClient
}

func NewEventsProxyService(runClient workflowconnect.InternalRunServiceClient) *EventsProxyService {
	return &EventsProxyService{runClient: runClient}
}

// Record accepts action events and forwards them synchronously to the run service.
func (s *EventsProxyService) Record(ctx context.Context, req *connect.Request[workflow.RecordRequest]) (*connect.Response[workflow.RecordResponse], error) {
	if s.runClient == nil {
		return nil, connect.NewError(connect.CodeFailedPrecondition, fmt.Errorf("run client is not initialized"))
	}
	if err := req.Msg.Validate(); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	if len(req.Msg.GetEvents()) == 0 {
		return connect.NewResponse(&workflow.RecordResponse{}), nil
	}

	recordEventReq := &workflow.RecordActionEventsRequest{Events: req.Msg.GetEvents()}
	recordActionResp, err := s.runClient.RecordActionEvents(ctx, connect.NewRequest(recordEventReq))
	if err != nil {
		return nil, fmt.Errorf("failed to forward action events to run service: %w", err)
	}
	status := recordActionResp.Msg.GetStatus()
	if status != nil && status.Code != 0 {
		return nil, fmt.Errorf("run service returned non-ok status for action events: code=%d, msg=%s", status.Code, status.Message)
	}

	return connect.NewResponse(&workflow.RecordResponse{}), nil
}
