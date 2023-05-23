package interfaces

import (
	"context"
)

//go:generate mockery -name=Webhook -output=../mocks -case=underscore

// Webhook Defines the interface for Publishing execution event to other services (Slack).
type Webhook interface {
	// Post The notificationType is inferred from the Notification object in the Execution Spec.
	Post(ctx context.Context, message string) error
}
