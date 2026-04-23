package service

import (
	"context"

	"connectrpc.com/connect"

	flyteapp "github.com/flyteorg/flyte/v2/gen/go/flyteidl2/app"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/app/appconnect"
)

// AppLogsService is the control plane implementation of AppLogsServiceHandler.
// It proxies the server-streaming TailLogs RPC to the data plane.
type AppLogsService struct {
	appconnect.UnimplementedAppLogsServiceHandler
	internalClient appconnect.AppLogsServiceClient
}

// NewAppLogsService creates a new AppLogsService.
func NewAppLogsService(internalClient appconnect.AppLogsServiceClient) *AppLogsService {
	return &AppLogsService{internalClient: internalClient}
}

var _ appconnect.AppLogsServiceHandler = (*AppLogsService)(nil)

// TailLogs proxies the server-streaming TailLogs RPC to InternalAppLogsService.
func (s *AppLogsService) TailLogs(
	ctx context.Context,
	req *connect.Request[flyteapp.TailLogsRequest],
	stream *connect.ServerStream[flyteapp.TailLogsResponse],
) error {
	clientStream, err := s.internalClient.TailLogs(ctx, req)
	if err != nil {
		return connect.NewError(connect.CodeInternal, err)
	}
	defer clientStream.Close()
	for clientStream.Receive() {
		if err := stream.Send(clientStream.Msg()); err != nil {
			return err
		}
	}
	return clientStream.Err()
}
