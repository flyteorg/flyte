package events

import (
	"context"

	"github.com/golang/protobuf/proto"
)

type EventSinkType = string

// EventSink determines how/where Events are emitted to. The type of EventSink the operator wants should be configurable.
// In Flyte, we also have local implementations so that operators can emit events without dependency on other services.
type EventSink interface {

	// Send the Event to this EventSink. The EventSink will identify the type of message through the
	// specified eventType and sink it appropriately based on the type.
	Sink(ctx context.Context, message proto.Message) error

	// Callers should close the EventSink when it is no longer being used as there may be long living
	// connections.
	Close() error
}
