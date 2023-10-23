package processor

import (
	"context"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/artifact"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/event"
	"github.com/flyteorg/flyte/flytestdlib/logger"
)

// ServiceCallHandler will take events and call the grpc endpoints directly. The service should most likely be local.
type ServiceCallHandler struct {
	service artifact.ArtifactRegistryServer
}

func (s *ServiceCallHandler) HandleEventExecStart(ctx context.Context, start *event.CloudEventExecutionStart) error {
	//TODO implement me
	panic("implement me")
}

func (s *ServiceCallHandler) HandleEventWorkflowExec(ctx context.Context, execution *event.CloudEventWorkflowExecution) error {
	//TODO implement me
	panic("implement me")
}

func (s *ServiceCallHandler) HandleEventTaskExec(ctx context.Context, execution *event.CloudEventTaskExecution) error {
	//TODO implement me
	panic("implement me")
}

func (s *ServiceCallHandler) HandleEventNodeExec(ctx context.Context, execution *event.CloudEventNodeExecution) error {
	//TODO implement me
	panic("implement me")
}

func NewServiceCallHandler(ctx context.Context, svc artifact.ArtifactRegistryServer) EventsHandlerInterface {
	logger.Infof(ctx, "Creating new service call handler")
	return &ServiceCallHandler{
		service: svc,
	}
}
