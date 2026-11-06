package service

import (
	"context"
	"errors"
	"io"

	"connectrpc.com/connect/v2"

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
	req *flyteapp.TailLogsRequest,
	stream appconnect.AppLogsServiceTailLogsServerStream,
) error {
	clientStream, err := s.internalClient.TailLogs(ctx, req)
	if err != nil {
		return connect.NewError(connect.CodeInternal, err.Error()).WithCause(err)
	}
	defer clientStream.Close() //nolint:errcheck
	var streamErr error
	for {
		msg, err2 := clientStream.Receive()
		if err2 != nil {
			streamErr = err2
			break
		}
		if err := stream.Send(msg); err != nil {
			return err
		}
	}
	if errors.Is(streamErr, io.EOF) {
		streamErr = nil
	}
	return streamErr
}
