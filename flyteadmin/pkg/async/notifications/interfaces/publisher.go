package interfaces

import (
	"context"

	"github.com/golang/protobuf/proto"
)

// Note on Notifications

// Notifications are handled in two steps.
// 1. Publishing a notification
// 2. Processing a notification

// Publishing a notification enqueues a notification message to be processed. Currently there is only
// one publisher for all notification types with the type differing based on the key.
// The notification hasn't been delivered at this stage.
// Processing a notification takes a notification message from the publisher and will pass
// the notification using the desired delivery method (ex: email). There is one processor per
// notification type.

// Publish a notification will differ between different types of notifications using the key
// The contract requires one subscription per type i.e. one for email one for slack, etc...
type Publisher interface {
	// The notification type is inferred from the Notification object in the Execution Spec.
	Publish(ctx context.Context, notificationType string, msg proto.Message) error
}
