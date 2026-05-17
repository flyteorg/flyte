package service

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	"golang.org/x/sync/semaphore"

	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/dataproxy"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/dataproxy/dataproxyconnect"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/workflow"
)

const defaultMaxConcurrentStreams = 100

// RunLogsService implements the RunLogsServiceHandler interface.
type RunLogsService struct {
	dataProxyClient dataproxyconnect.DataProxyServiceClient
	sem             *semaphore.Weighted
}

// NewRunLogsService creates a new RunLogsService.
func NewRunLogsService(dataProxyClient dataproxyconnect.DataProxyServiceClient) *RunLogsService {
	return &RunLogsService{
		dataProxyClient: dataProxyClient,
		sem:             semaphore.NewWeighted(defaultMaxConcurrentStreams),
	}
}

// TailLogs streams pod logs for an action attempt by delegating to DataProxyService.
func (s *RunLogsService) TailLogs(ctx context.Context, req *connect.Request[workflow.TailLogsRequest], stream *connect.ServerStream[workflow.TailLogsResponse]) error {
	msg := req.Msg
	if msg.GetActionId() == nil {
		return connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("action_id is required"))
	}
	if !s.sem.TryAcquire(1) {
		return connect.NewError(connect.CodeResourceExhausted, fmt.Errorf("too many concurrent log streams"))
	}
	defer s.sem.Release(1)

	dpStream, err := s.dataProxyClient.TailLogs(ctx, connect.NewRequest(&dataproxy.TailLogsRequest{
		ActionId: msg.GetActionId(),
		Attempt:  msg.GetAttempt(),
	}))
	if err != nil {
		return err
	}
	defer dpStream.Close()

	for dpStream.Receive() {
		dpResp := dpStream.Msg()
		logs := make([]*workflow.TailLogsResponse_Logs, 0, len(dpResp.GetLogs()))
		for _, l := range dpResp.GetLogs() {
			logs = append(logs, &workflow.TailLogsResponse_Logs{
				Lines: l.GetLines(),
			})
		}
		if err := stream.Send(&workflow.TailLogsResponse{Logs: logs}); err != nil {
			return err
		}
	}
	return dpStream.Err()
}
