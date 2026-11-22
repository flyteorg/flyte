package rpc

import (
	"context"

	"connectrpc.com/connect"

	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/common"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/workflow"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/workflow/workflowconnect"
	"github.com/flyteorg/flyte/v2/runs/config"
	"github.com/flyteorg/flyte/v2/runs/service"
)

// NewRunHandler exposes a Run through the public Connect API.
func NewRunHandler(run service.Run, authServerBaseURL string, trustHeaders bool, identityHeaders config.IdentityHeadersConfig) workflowconnect.RunServiceHandler {
	return &runHandler{
		run:             run,
		enricher:        newIdentityEnricher(authServerBaseURL),
		trustHeaders:    trustHeaders,
		identityHeaders: identityHeaders,
	}
}

type runHandler struct {
	run             service.Run
	enricher        *identityEnricher
	trustHeaders    bool
	identityHeaders config.IdentityHeadersConfig
}

func (h *runHandler) WatchGroups(ctx context.Context, req *connect.Request[workflow.WatchGroupsRequest], stream *connect.ServerStream[workflow.WatchGroupsResponse]) error {
	return rpcError(h.run.WatchGroups(ctx, req.Msg, stream.Send))
}

func (h *runHandler) CreateRun(ctx context.Context, req *connect.Request[workflow.CreateRunRequest]) (*connect.Response[workflow.CreateRunResponse], error) {
	var executedBy *common.EnrichedIdentity
	if h.trustHeaders {
		executedBy = resolveIdentity(ctx, req.Header(), h.identityHeaders, h.enricher)
	}
	return rpcResponse(h.run.CreateRun(ctx, req.Msg, executedBy))
}

func (h *runHandler) AbortRun(ctx context.Context, req *connect.Request[workflow.AbortRunRequest]) (*connect.Response[workflow.AbortRunResponse], error) {
	return rpcResponse(h.run.AbortRun(ctx, req.Msg))
}

func (h *runHandler) GetRunDetails(ctx context.Context, req *connect.Request[workflow.GetRunDetailsRequest]) (*connect.Response[workflow.GetRunDetailsResponse], error) {
	return rpcResponse(h.run.GetRunDetails(ctx, req.Msg))
}

func (h *runHandler) GetActionDetails(ctx context.Context, req *connect.Request[workflow.GetActionDetailsRequest]) (*connect.Response[workflow.GetActionDetailsResponse], error) {
	return rpcResponse(h.run.GetActionDetails(ctx, req.Msg))
}

func (h *runHandler) GetActionData(ctx context.Context, req *connect.Request[workflow.GetActionDataRequest]) (*connect.Response[workflow.GetActionDataResponse], error) {
	return rpcResponse(h.run.GetActionData(ctx, req.Msg))
}

func (h *runHandler) ListRuns(ctx context.Context, req *connect.Request[workflow.ListRunsRequest]) (*connect.Response[workflow.ListRunsResponse], error) {
	return rpcResponse(h.run.ListRuns(ctx, req.Msg))
}

func (h *runHandler) ListActions(ctx context.Context, req *connect.Request[workflow.ListActionsRequest]) (*connect.Response[workflow.ListActionsResponse], error) {
	return rpcResponse(h.run.ListActions(ctx, req.Msg))
}

func (h *runHandler) GetActionDataURIs(ctx context.Context, req *connect.Request[workflow.GetActionDataURIsRequest]) (*connect.Response[workflow.GetActionDataURIsResponse], error) {
	return rpcResponse(h.run.GetActionDataURIs(ctx, req.Msg))
}

func (h *runHandler) GetActionLogContext(ctx context.Context, req *connect.Request[workflow.GetActionLogContextRequest]) (*connect.Response[workflow.GetActionLogContextResponse], error) {
	return rpcResponse(h.run.GetActionLogContext(ctx, req.Msg))
}

func (h *runHandler) AbortAction(ctx context.Context, req *connect.Request[workflow.AbortActionRequest]) (*connect.Response[workflow.AbortActionResponse], error) {
	return rpcResponse(h.run.AbortAction(ctx, req.Msg))
}

func (h *runHandler) SignalEvent(ctx context.Context, req *connect.Request[workflow.SignalEventRequest]) (*connect.Response[workflow.SignalEventResponse], error) {
	var signalledBy *common.EnrichedIdentity
	if h.trustHeaders {
		signalledBy = identityFromHeaders(req.Header(), h.identityHeaders)
	}
	return rpcResponse(h.run.SignalEvent(ctx, req.Msg, signalledBy))
}

func (h *runHandler) WatchRunDetails(ctx context.Context, req *connect.Request[workflow.WatchRunDetailsRequest], stream *connect.ServerStream[workflow.WatchRunDetailsResponse]) error {
	return rpcError(h.run.WatchRunDetails(ctx, req.Msg, stream.Send))
}

func (h *runHandler) WatchActionDetails(ctx context.Context, req *connect.Request[workflow.WatchActionDetailsRequest], stream *connect.ServerStream[workflow.WatchActionDetailsResponse]) error {
	return rpcError(h.run.WatchActionDetails(ctx, req.Msg, stream.Send))
}

func (h *runHandler) WatchRuns(ctx context.Context, req *connect.Request[workflow.WatchRunsRequest], stream *connect.ServerStream[workflow.WatchRunsResponse]) error {
	return rpcError(h.run.WatchRuns(ctx, req.Msg, stream.Send))
}

func (h *runHandler) WatchActions(ctx context.Context, req *connect.Request[workflow.WatchActionsRequest], stream *connect.ServerStream[workflow.WatchActionsResponse]) error {
	return rpcError(h.run.WatchActions(ctx, req.Msg, stream.Send))
}

func (h *runHandler) WatchClusterEvents(ctx context.Context, req *connect.Request[workflow.WatchClusterEventsRequest], stream *connect.ServerStream[workflow.WatchClusterEventsResponse]) error {
	return rpcError(h.run.WatchClusterEvents(ctx, req.Msg, stream.Send))
}
