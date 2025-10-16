package events

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/golang/protobuf/proto"
	grpcRetry "github.com/grpc-ecosystem/go-grpc-middleware/retry"
	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel/propagation"
	"golang.org/x/time/rate"
	"google.golang.org/grpc"

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

type queuedEvent struct {
	message proto.Message
	id      []byte
}

type asyncEventSinkMetrics struct {
	eventsQueued      prometheus.Counter
	eventsProcessed   prometheus.Counter
	eventsFailed      prometheus.Counter
	queueDepth        prometheus.Gauge
	processingLatency prometheus.Histogram
}

type adminEventSink struct {
	adminClient service.AdminServiceClient
	filter      fastcheck.Filter
	rateLimiter *rate.Limiter
	cfg         *Config

	eventQueue chan *queuedEvent
	stopCh     chan struct{}
	closed     chan struct{}
	wg         sync.WaitGroup
	metrics    *asyncEventSinkMetrics
	inFlight   sync.Map
}

func newAsyncEventSinkMetrics(scope promutils.Scope) *asyncEventSinkMetrics {
	return &asyncEventSinkMetrics{
		eventsQueued: scope.MustNewCounter("events_queued",
			"Total number of events queued for async processing"),
		eventsProcessed: scope.MustNewCounter("events_processed",
			"Total number of events successfully processed"),
		eventsFailed: scope.MustNewCounter("events_failed",
			"Total number of events that failed processing"),
		queueDepth: scope.MustNewGauge("queue_depth",
			"Current depth of the event queue"),
		processingLatency: scope.MustNewHistogram("processing_latency_ms",
			"Event processing latency in milliseconds"),
	}
}

// Constructs a new EventSink that sends events to FlyteAdmin through gRPC
func NewAdminEventSink(ctx context.Context, adminClient service.AdminServiceClient, config *Config, filter fastcheck.Filter, scope promutils.Scope) (EventSink, error) {
	rateLimiter := rate.NewLimiter(rate.Limit(config.Rate), config.Capacity)

	eventSink := &adminEventSink{
		adminClient: adminClient,
		filter:      filter,
		rateLimiter: rateLimiter,
		cfg:         config,
	}

	eventSink.eventQueue = make(chan *queuedEvent, config.EventQueueSize)
	eventSink.stopCh = make(chan struct{})
	eventSink.closed = make(chan struct{})
	eventSink.metrics = newAsyncEventSinkMetrics(scope.NewSubScope("async"))

	eventSink.wg.Add(1)
	go eventSink.worker()

	logger.Infof(ctx, "Created new AdminEventSink to Admin service with async processing enabled")
	return eventSink, nil
}

func (s *adminEventSink) Sink(ctx context.Context, message proto.Message) error {
	logger.Debugf(ctx, "AdminEventSink received a new event %s", message.String())

	id, err := IDFromMessage(message)
	if err != nil {
		return fmt.Errorf("failed to parse message id [%v]", message.String())
	}

	if s.filter.Contains(ctx, id) {
		logger.Debugf(ctx, "event '%s' has already been sent", string(id))
		return &errors.EventError{
			Code:    errors.AlreadyExists,
			Cause:   fmt.Errorf("event has already been sent"),
			Message: "Event Already Exists",
		}
	}

	// Check if event is currently in-flight to prevent duplicates in queue
	idStr := string(id)
	if _, exists := s.inFlight.LoadOrStore(idStr, true); exists {
		logger.Debugf(ctx, "event '%s' is already queued for processing", idStr)
		return &errors.EventError{
			Code:    errors.AlreadyExists,
			Cause:   fmt.Errorf("event is already queued"),
			Message: "Event Already Queued",
		}
	}

	// Check if sink is closed before enqueueing
	select {
	case <-s.closed:
		s.inFlight.Delete(idStr)
		return fmt.Errorf("event sink is closed")
	default:
	}

	qe := &queuedEvent{
		message: message,
		id:      id,
	}

	select {
	case s.eventQueue <- qe:
		s.metrics.eventsQueued.Inc()
		s.metrics.queueDepth.Set(float64(len(s.eventQueue)))
		return nil
	case <-s.closed:
		s.inFlight.Delete(idStr)
		return fmt.Errorf("event sink is closed")
	default:
		s.inFlight.Delete(idStr)
		logger.Warnf(ctx, "event queue is full (%d events), rejecting event '%s'", len(s.eventQueue), idStr)
		return &errors.EventError{
			Code:    errors.ResourceExhausted,
			Cause:   fmt.Errorf("event queue is full"),
			Message: "Event Queue Full",
		}
	}
}

func (s *adminEventSink) sendEvent(ctx context.Context, message proto.Message) error {
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
		return fmt.Errorf("unknown event type [%s]", eventMessage.String())
	}
	return nil
}

func (s *adminEventSink) worker() {
	ctx := context.Background()
	defer s.wg.Done()

	for {
		select {
		case <-s.stopCh:
			logger.Infof(ctx, "Draining event queue, %d events remaining", len(s.eventQueue))
			for {
				select {
				case qe := <-s.eventQueue:
					s.processEvent(ctx, qe, true)
				default:
					logger.Infof(ctx, "Event queue drained")
					return
				}
			}
		case qe := <-s.eventQueue:
			s.processEvent(ctx, qe, false)
		}
	}
}

func (s *adminEventSink) processEvent(ctx context.Context, qe *queuedEvent, draining bool) {
	startTime := time.Now()
	s.metrics.queueDepth.Set(float64(len(s.eventQueue)))
	idStr := string(qe.id)

	defer s.inFlight.Delete(idStr)

	if !draining {
		if err := s.rateLimiter.Wait(ctx); err != nil {
			logger.Warnf(ctx, "Rate limiter context cancelled: %v", err)
			s.metrics.eventsFailed.Inc()
			return
		}
	}

	if err := s.sendEvent(ctx, qe.message); err != nil {
		logger.Errorf(ctx, "Failed to send async event after retries: %v", err)
		s.metrics.eventsFailed.Inc()
	} else {
		s.filter.Add(ctx, qe.id)
		s.metrics.eventsProcessed.Inc()
		latency := time.Since(startTime).Milliseconds()
		s.metrics.processingLatency.Observe(float64(latency))
	}
}

func (s *adminEventSink) Close() error {
	logger.Info(context.Background(), "Shutting down async event sink")
	close(s.closed)
	close(s.stopCh)
	s.wg.Wait()
	logger.Info(context.Background(), "Async event sink shutdown complete")
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
		return nil, fmt.Errorf("unknown event type [%s]", eventMessage.String())
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

		return NewAdminEventSink(ctx, adminClient, config, filter, scope.NewSubScope("admin"))
	default:
		return NewStdoutSink()
	}
}
