package rpc

import (
	"context"

	"connectrpc.com/connect"

	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/trigger"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/trigger/triggerconnect"
	"github.com/flyteorg/flyte/v2/runs/service"
)

type triggerHandler struct{ triggers service.Trigger }

// NewTriggerHandler exposes the trigger service through the Connect API.
func NewTriggerHandler(triggers service.Trigger) triggerconnect.TriggerServiceHandler {
	return &triggerHandler{triggers: triggers}
}

func (h *triggerHandler) DeployTrigger(ctx context.Context, req *connect.Request[trigger.DeployTriggerRequest]) (*connect.Response[trigger.DeployTriggerResponse], error) {
	return rpcResponse(h.triggers.DeployTrigger(ctx, req.Msg))
}

func (h *triggerHandler) GetTriggerDetails(ctx context.Context, req *connect.Request[trigger.GetTriggerDetailsRequest]) (*connect.Response[trigger.GetTriggerDetailsResponse], error) {
	return rpcResponse(h.triggers.GetTriggerDetails(ctx, req.Msg))
}

func (h *triggerHandler) GetTriggerRevisionDetails(ctx context.Context, req *connect.Request[trigger.GetTriggerRevisionDetailsRequest]) (*connect.Response[trigger.GetTriggerRevisionDetailsResponse], error) {
	return rpcResponse(h.triggers.GetTriggerRevisionDetails(ctx, req.Msg))
}

func (h *triggerHandler) ListTriggers(ctx context.Context, req *connect.Request[trigger.ListTriggersRequest]) (*connect.Response[trigger.ListTriggersResponse], error) {
	return rpcResponse(h.triggers.ListTriggers(ctx, req.Msg))
}

func (h *triggerHandler) GetTriggerRevisionHistory(ctx context.Context, req *connect.Request[trigger.GetTriggerRevisionHistoryRequest]) (*connect.Response[trigger.GetTriggerRevisionHistoryResponse], error) {
	return rpcResponse(h.triggers.GetTriggerRevisionHistory(ctx, req.Msg))
}

func (h *triggerHandler) UpdateTriggers(ctx context.Context, req *connect.Request[trigger.UpdateTriggersRequest]) (*connect.Response[trigger.UpdateTriggersResponse], error) {
	return rpcResponse(h.triggers.UpdateTriggers(ctx, req.Msg))
}

func (h *triggerHandler) DeleteTriggers(ctx context.Context, req *connect.Request[trigger.DeleteTriggersRequest]) (*connect.Response[trigger.DeleteTriggersResponse], error) {
	return rpcResponse(h.triggers.DeleteTriggers(ctx, req.Msg))
}
