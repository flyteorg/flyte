package processor

import (
	"context"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/event"
)

type EventsHandlerInterface interface {
	HandleEventExecStart(context.Context, *event.CloudEventExecutionStart) error
	HandleEventWorkflowExec(context.Context, *event.CloudEventWorkflowExecution) error
	HandleEventTaskExec(context.Context, *event.CloudEventTaskExecution) error
	HandleEventNodeExec(context.Context, *event.CloudEventNodeExecution) error
}

// EventsProcessorInterface is a copy of the notifications processor in admin except that start takes a context
type EventsProcessorInterface interface {
	// StartProcessing whatever it is that needs to be processed.
	StartProcessing(ctx context.Context)

	// StopProcessing is called when the server is shutting down.
	StopProcessing() error
}
