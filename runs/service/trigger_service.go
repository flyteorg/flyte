package service

import (
	"context"

	"connectrpc.com/connect"

	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/trigger"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/trigger/triggerconnect"
)

// TriggerService is a dummy implementation that returns empty responses for all endpoints.
type TriggerService struct {
	triggerconnect.UnimplementedTriggerServiceHandler
}

func NewTriggerService() *TriggerService {
	return &TriggerService{}
}

var _ triggerconnect.TriggerServiceHandler = (*TriggerService)(nil)

func (s *TriggerService) DeployTrigger(
	ctx context.Context,
	req *connect.Request[trigger.DeployTriggerRequest],
) (*connect.Response[trigger.DeployTriggerResponse], error) {
	return connect.NewResponse(&trigger.DeployTriggerResponse{}), nil
}

func (s *TriggerService) GetTriggerDetails(
	ctx context.Context,
	req *connect.Request[trigger.GetTriggerDetailsRequest],
) (*connect.Response[trigger.GetTriggerDetailsResponse], error) {
	return connect.NewResponse(&trigger.GetTriggerDetailsResponse{}), nil
}

func (s *TriggerService) GetTriggerRevisionDetails(
	ctx context.Context,
	req *connect.Request[trigger.GetTriggerRevisionDetailsRequest],
) (*connect.Response[trigger.GetTriggerRevisionDetailsResponse], error) {
	return connect.NewResponse(&trigger.GetTriggerRevisionDetailsResponse{}), nil
}

func (s *TriggerService) ListTriggers(
	ctx context.Context,
	req *connect.Request[trigger.ListTriggersRequest],
) (*connect.Response[trigger.ListTriggersResponse], error) {
	return connect.NewResponse(&trigger.ListTriggersResponse{}), nil
}

func (s *TriggerService) GetTriggerRevisionHistory(
	ctx context.Context,
	req *connect.Request[trigger.GetTriggerRevisionHistoryRequest],
) (*connect.Response[trigger.GetTriggerRevisionHistoryResponse], error) {
	return connect.NewResponse(&trigger.GetTriggerRevisionHistoryResponse{}), nil
}

func (s *TriggerService) UpdateTriggers(
	ctx context.Context,
	req *connect.Request[trigger.UpdateTriggersRequest],
) (*connect.Response[trigger.UpdateTriggersResponse], error) {
	return connect.NewResponse(&trigger.UpdateTriggersResponse{}), nil
}

func (s *TriggerService) DeleteTriggers(
	ctx context.Context,
	req *connect.Request[trigger.DeleteTriggersRequest],
) (*connect.Response[trigger.DeleteTriggersResponse], error) {
	return connect.NewResponse(&trigger.DeleteTriggersResponse{}), nil
}
