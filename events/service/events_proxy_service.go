package service

import (
	"context"

	"connectrpc.com/connect/v2"

	"github.com/flyteorg/flyte/v2/flytestdlib/logger"
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
func (s *EventsProxyService) Record(ctx context.Context, req *workflow.RecordRequest) (*workflow.RecordResponse, error) {
	if s.runClient == nil {
		return nil, connect.Errorf(connect.CodeFailedPrecondition, "run client is not initialized")
	}
	if err := req.Validate(); err != nil {
		logger.Errorf(ctx, "invalid EventsProxyService.Record request: %v", err)
		return nil, connect.NewError(connect.CodeInvalidArgument, err.Error()).WithCause(err)
	}
	if len(req.GetEvents()) == 0 {
		return &workflow.RecordResponse{}, nil
	}

	recordEventReq := &workflow.RecordActionEventsRequest{Events: req.GetEvents()}
	recordActionResp, err := s.runClient.RecordActionEvents(ctx, recordEventReq)
	if err != nil {
		logger.Warnf(ctx, "failed to forward action events to run service: %v", err)
		return nil, connect.NewError(connect.CodeInternal, err.Error()).WithCause(err)
	}
	status := recordActionResp.GetStatus()
	if status != nil && status.Code != 0 {
		logger.Warnf(ctx, "run service returned non-ok status for action events: code=%d, msg=%s", status.Code, status.Message)
		return nil, connect.Errorf(connect.CodeInternal, "run service failed: %s", status.Message)
	}

	return &workflow.RecordResponse{}, nil
}
