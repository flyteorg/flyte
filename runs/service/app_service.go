package service

import (
	"context"
	"errors"

	"connectrpc.com/connect"

	flyteapp "github.com/flyteorg/flyte/v2/gen/go/flyteidl2/app"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/app/appconnect"
)

// AppService is a dummy implementation that returns empty responses for all endpoints.
type AppService struct {
	appconnect.UnimplementedAppServiceHandler
}

func NewAppService() *AppService {
	return &AppService{}
}

var _ appconnect.AppServiceHandler = (*AppService)(nil)

func (s *AppService) Create(
	ctx context.Context,
	req *connect.Request[flyteapp.CreateRequest],
) (*connect.Response[flyteapp.CreateResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("App service is not implemented"))
}

func (s *AppService) Get(
	ctx context.Context,
	req *connect.Request[flyteapp.GetRequest],
) (*connect.Response[flyteapp.GetResponse], error) {
	return connect.NewResponse(&flyteapp.GetResponse{}), nil
}

func (s *AppService) Update(
	ctx context.Context,
	req *connect.Request[flyteapp.UpdateRequest],
) (*connect.Response[flyteapp.UpdateResponse], error) {
	return connect.NewResponse(&flyteapp.UpdateResponse{}), nil
}

func (s *AppService) UpdateStatus(
	ctx context.Context,
	req *connect.Request[flyteapp.UpdateStatusRequest],
) (*connect.Response[flyteapp.UpdateStatusResponse], error) {
	return connect.NewResponse(&flyteapp.UpdateStatusResponse{}), nil
}

func (s *AppService) Delete(
	ctx context.Context,
	req *connect.Request[flyteapp.DeleteRequest],
) (*connect.Response[flyteapp.DeleteResponse], error) {
	return connect.NewResponse(&flyteapp.DeleteResponse{}), nil
}

func (s *AppService) List(
	ctx context.Context,
	req *connect.Request[flyteapp.ListRequest],
) (*connect.Response[flyteapp.ListResponse], error) {
	return connect.NewResponse(&flyteapp.ListResponse{}), nil
}

func (s *AppService) Watch(
	ctx context.Context,
	req *connect.Request[flyteapp.WatchRequest],
	stream *connect.ServerStream[flyteapp.WatchResponse],
) error {
	return nil
}

func (s *AppService) Lease(
	ctx context.Context,
	req *connect.Request[flyteapp.LeaseRequest],
	stream *connect.ServerStream[flyteapp.LeaseResponse],
) error {
	return nil
}
