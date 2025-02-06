package interfaces

import (
	"context"

	"github.com/golang/protobuf/proto"
)

//go:generate mockery-v2 --name=Publisher --output=../mocks --case=underscore --with-expecter

// Publisher Defines the interface for Publishing execution event to other services (AWS pub/sub, Kafka).
type Publisher interface {
	// Publish The notificationType is inferred from the Notification object in the Execution Spec.
	Publish(ctx context.Context, notificationType string, msg proto.Message) error
}
