package events

import (
	"context"
	"fmt"
	"time"

	grpcRetry "github.com/grpc-ecosystem/go-grpc-middleware/retry"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel/propagation"
	"golang.org/x/time/rate"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"

	admin2 "github.com/flyteorg/flyte/flyteidl/clients/go/admin"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/event"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/service"
	"github.com/flyteorg/flyte/flytepropeller/events/errors"
	"github.com/flyteorg/flyte/flytestdlib/fastcheck"
	"github.com/flyteorg/flyte/flytestdlib/logger"
	"github.com/flyteorg/flyte/flytestdlib/otelutils"
	"github.com/flyteorg/flyte/flytestdlib/promutils"
)

type adminEventSink struct {
	adminClient service.AdminServiceClient
	filter      fastcheck.Filter
	rateLimiter *rate.Limiter
	cfg         *Config
}

// Constructs a new EventSink that sends events to FlyteAdmin through gRPC
func NewAdminEventSink(ctx context.Context, adminClient service.AdminServiceClient, config *Config, filter fastcheck.Filter) (EventSink, error) {
	rateLimiter := rate.NewLimiter(rate.Limit(config.Rate), config.Capacity)

	eventSink := &adminEventSink{
		adminClient: adminClient,
		filter:      filter,
		rateLimiter: rateLimiter,
		cfg:         config,
	}

	logger.Infof(ctx, "Created new AdminEventSink to Admin service")
	return eventSink, nil
}

// Sends events to the FlyteAdmin service through gRPC
func (s *adminEventSink) Sink(ctx context.Context, message proto.Message) error {
	logger.Debugf(ctx, "AdminEventSink received a new event %+v", message)

	// Short-circuit if event has already been sent
	id, err := IDFromMessage(message)
	if err != nil {
		return fmt.Errorf("Failed to parse message id [%+v]", message)
	}

	if s.filter.Contains(ctx, id) {
		logger.Debugf(ctx, "event '%s' has already been sent", string(id))
		return &errors.EventError{
			Code:    errors.AlreadyExists,
			Cause:   fmt.Errorf("event has already been sent"),
			Message: "Event Already Exists",
		}
	}

	// Validate submission with rate limiter and send admin event
	if s.rateLimiter.Allow() {
		switch eventMessage := message.(type) {
		case *event.WorkflowExecutionEvent:
			request := &admin.WorkflowExecutionEventRequest{
				Event: eventMessage,
			}

			_, err := s.adminClient.CreateWorkflowEvent(ctx, request)
			if err != nil {
				return errors.WrapError(err)
			}
		case *event.NodeExecutionEvent:
			request := &admin.NodeExecutionEventRequest{
				Event: eventMessage,
			}

			_, err := s.adminClient.CreateNodeEvent(ctx, request)
			if err != nil {
				return errors.WrapError(err)
			}
		case *event.TaskExecutionEvent:
			request := &admin.TaskExecutionEventRequest{
				Event: eventMessage,
			}

			_, err := s.adminClient.CreateTaskEvent(ctx, request)
			if err != nil {
				return errors.WrapError(err)
			}
		default:
			return fmt.Errorf("unknown event type [%+v]", eventMessage)
		}
	} else {
		return &errors.EventError{Code: errors.ResourceExhausted,
			Cause: fmt.Errorf("Admin EventSink throttling admin traffic"), Message: "Resource Exhausted"}
	}

	s.filter.Add(ctx, id)
	return nil
}

// Closes the gRPC client connection. This should be deferred on the client does shutdown cleanup.
func (s *adminEventSink) Close() error {
	return nil
}

// Generates an ID which uniquely represents the admin event entity and associated phase.
func IDFromMessage(message proto.Message) ([]byte, error) {
	var id string
	switch eventMessage := message.(type) {
	case *event.WorkflowExecutionEvent:
		wid := eventMessage.GetExecutionId()
		id = fmt.Sprintf("%s:%s:%s:%d", wid.GetProject(), wid.GetDomain(), wid.GetName(), eventMessage.GetPhase())
	case *event.NodeExecutionEvent:
		nid := eventMessage.GetId()
		wid := nid.GetExecutionId()
		id = fmt.Sprintf("%s:%s:%s:%s:%s:%d", wid.GetProject(), wid.GetDomain(), wid.GetName(), nid.GetNodeId(), eventMessage.GetRetryGroup(), eventMessage.GetPhase())
	case *event.TaskExecutionEvent:
		tid := eventMessage.GetTaskId()
		nid := eventMessage.GetParentNodeExecutionId()
		wid := nid.GetExecutionId()
		id = fmt.Sprintf("%s:%s:%s:%s:%s:%s:%d:%d:%d", wid.GetProject(), wid.GetDomain(), wid.GetName(), nid.GetNodeId(), tid.GetName(), tid.GetVersion(), eventMessage.GetRetryAttempt(), eventMessage.GetPhase(), eventMessage.GetPhaseVersion())
	default:
		return nil, fmt.Errorf("unknown event type [%+v]", eventMessage)
	}

	return []byte(id), nil
}

func initializeAdminClientFromConfig(ctx context.Context, config *Config) (client service.AdminServiceClient, err error) {
	cfg := admin2.GetConfig(ctx)
	tracerProvider := otelutils.GetTracerProvider(otelutils.AdminClientTracer)

	grpcOptions := []grpcRetry.CallOption{
		grpcRetry.WithBackoff(grpcRetry.BackoffExponentialWithJitter(time.Duration(config.BackoffScalar)*time.Millisecond, config.GetBackoffJitter(ctx))),
		grpcRetry.WithMax(uint(config.MaxRetries)), // #nosec G115
	}

	opt := grpc.WithChainUnaryInterceptor(
		otelgrpc.UnaryClientInterceptor(
			otelgrpc.WithTracerProvider(tracerProvider),
			otelgrpc.WithPropagators(propagation.TraceContext{}),
		),
		grpcRetry.UnaryClientInterceptor(grpcOptions...),
	)

	clients, err := admin2.NewClientsetBuilder().WithDialOptions(opt).WithConfig(cfg).Build(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize clientset. Error: %w", err)
	}

	return clients.AdminClient(), nil
}

func ConstructEventSink(ctx context.Context, config *Config, scope promutils.Scope) (EventSink, error) {
	switch config.Type {
	case EventSinkLog:
		return NewLogSink()
	case EventSinkFile:
		return NewFileSink(config.FilePath)
	case EventSinkAdmin:
		adminClient, err := initializeAdminClientFromConfig(ctx, config)
		if err != nil {
			return nil, err
		}

		filter, err := fastcheck.NewOppoBloomFilter(50000, scope.NewSubScope("admin").NewSubScope("filter"))
		if err != nil {
			return nil, err
		}

		return NewAdminEventSink(ctx, adminClient, config, filter)
	default:
		return NewStdoutSink()
	}
}
