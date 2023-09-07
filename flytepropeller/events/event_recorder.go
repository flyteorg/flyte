package events

import (
	"context"
	"fmt"
	"time"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/event"
	"github.com/flyteorg/flytepropeller/events/errors"
	"github.com/flyteorg/flytestdlib/promutils"
	"github.com/flyteorg/flytestdlib/promutils/labeled"
	"github.com/golang/protobuf/proto"
)

const MaxErrorMessageLength = 104857600 //100KB
const truncationIndicator = "... <Message Truncated> ..."

type recordingMetrics struct {
	EventRecordingFailure           labeled.StopWatch
	EventRecordingSuccess           labeled.StopWatch
	EventRecordingAlreadyExists     labeled.Counter
	EventRecordingExecutionNotFound labeled.Counter
	EventRecordingResourceExhausted labeled.Counter
	EventRecordingEventSinkError    labeled.Counter
	EventRecordingInvalidArgument   labeled.Counter
}

// Recorder for Workflow, Node, and Task events
type EventRecorder interface {
	RecordNodeEvent(ctx context.Context, event *event.NodeExecutionEvent) error

	RecordTaskEvent(ctx context.Context, event *event.TaskExecutionEvent) error

	RecordWorkflowEvent(ctx context.Context, event *event.WorkflowExecutionEvent) error
}

// EventRecorder records workflow, node and task events to the eventSink it is configured with.
type eventRecorder struct {
	eventSink EventSink
	metrics   *recordingMetrics
}

func (r *eventRecorder) sinkEvent(ctx context.Context, event proto.Message) error {
	startTime := time.Now()

	err := r.eventSink.Sink(ctx, event)
	if errors.IsResourceExhausted(err) {
		r.metrics.EventRecordingResourceExhausted.Inc(ctx)
	}

	if err != nil {
		r.metrics.EventRecordingFailure.Observe(ctx, startTime, time.Now())
		return err
	}

	r.metrics.EventRecordingSuccess.Observe(ctx, startTime, time.Now())
	return nil
}

func (r *eventRecorder) RecordNodeEvent(ctx context.Context, e *event.NodeExecutionEvent) error {
	if err, ok := e.GetOutputResult().(*event.NodeExecutionEvent_Error); ok {
		truncateErrorMessage(err.Error, MaxErrorMessageLength)
	}

	return r.sinkEvent(ctx, e)
}

func (r *eventRecorder) RecordTaskEvent(ctx context.Context, e *event.TaskExecutionEvent) error {
	if err, ok := e.GetOutputResult().(*event.TaskExecutionEvent_Error); ok {
		truncateErrorMessage(err.Error, MaxErrorMessageLength)
	}

	return r.sinkEvent(ctx, e)
}

func (r *eventRecorder) RecordWorkflowEvent(ctx context.Context, e *event.WorkflowExecutionEvent) error {
	if err, ok := e.GetOutputResult().(*event.WorkflowExecutionEvent_Error); ok {
		truncateErrorMessage(err.Error, MaxErrorMessageLength)
	}

	return r.sinkEvent(ctx, e)
}

// If error message too large, truncate to mitigate grpc message size limit. Split the truncated size equally between
// the beginning and the end of the message to capture the most relevant information.
func truncateErrorMessage(err *core.ExecutionError, length int) {
	if len(err.Message) > length {
		err.Message = fmt.Sprintf("%s\n%s\n%s", err.Message[:length/2], truncationIndicator, err.Message[(len(err.Message)-length/2):])
	}
}

// Construct a new Event Recorder
func NewEventRecorder(eventSink EventSink, scope promutils.Scope) EventRecorder {
	recordingScope := scope.NewSubScope("event_recording")
	return &eventRecorder{
		eventSink: eventSink,
		metrics: &recordingMetrics{
			EventRecordingFailure:           labeled.NewStopWatch("failure_duration", "The time it took the failed event recording to occur", time.Millisecond, recordingScope),
			EventRecordingSuccess:           labeled.NewStopWatch("success_duration", "The time it took for a successful event recording to occur", time.Millisecond, recordingScope),
			EventRecordingAlreadyExists:     labeled.NewCounter("already_exists", "The count that a recorded event already exists", recordingScope),
			EventRecordingResourceExhausted: labeled.NewCounter("resource_exhausted", "The count that recording events was throttled", recordingScope),
			EventRecordingInvalidArgument:   labeled.NewCounter("invalid_argument", "The count for invalid argument errors", recordingScope),
			EventRecordingEventSinkError:    labeled.NewCounter("unexpected_err", "The count of event recording failures for unexpected reasons", recordingScope),
		},
	}
}
