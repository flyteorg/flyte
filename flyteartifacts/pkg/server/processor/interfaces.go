package processor

import (
	"context"

	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/golang/protobuf/proto"
)

type EventsHandlerInterface interface {
	// HandleEvent The cloudEvent here is the original deserialized event and the proto msg is message
	// that's been unmarshalled already from the cloudEvent.Data() field.
	HandleEvent(ctx context.Context, cloudEvent *event.Event, msg proto.Message) error
}

// EventsProcessorInterface is a copy of the notifications processor in admin except that start takes a context
type EventsProcessorInterface interface {
	// StartProcessing whatever it is that needs to be processed.
	StartProcessing(ctx context.Context, handler EventsHandlerInterface)

	// StopProcessing is called when the server is shutting down.
	StopProcessing() error
}
