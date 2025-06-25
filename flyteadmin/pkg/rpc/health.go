package rpc

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/service"
)

type HealthService struct {
	ServiceName string
}

func NewHealthService(serviceName string) *HealthService {
	return &HealthService{
		ServiceName: serviceName,
	}
}

func (h HealthService) Check(_ context.Context, req *service.HealthRequest) (*service.HealthResponse, error) {
	if req.GetService() != h.ServiceName {
		return nil, status.Error(codes.NotFound, "unknown service")
	}

	return &service.HealthResponse{Status: service.HealthResponse_SERVING}, nil
}

func (h HealthService) Watch(req *service.HealthRequest, srv service.HealthService_WatchServer) (err error) {
	if req.GetService() != h.ServiceName {
		err = srv.Send(&service.HealthResponse{Status: service.HealthResponse_SERVICE_UNKNOWN})
	} else {
		err = srv.Send(&service.HealthResponse{Status: service.HealthResponse_SERVING})
	}
	if err != nil {
		return err
	}
	for range srv.Context().Done() {
		return nil
	}
	return nil
}
