package interfaces

import (
	"context"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
)

//go:generate mockery -name=Webhook -output=../mocks -case=underscore

type Payload struct {
	Value string `protobuf:"bytes,1,opt,value=value"`
}

func (p Payload) Reset() {
	//TODO implement me
	panic("implement me")
}

func (p Payload) String() string {
	//TODO implement me
	panic("implement me")
}

func (p Payload) ProtoMessage() {
	//TODO implement me
	panic("implement me")
}

// Webhook Defines the interface for Publishing execution event to other services (Slack).
type Webhook interface {
	// Post The notificationType is inferred from the Notification object in the Execution Spec.
	Post(ctx context.Context, payload admin.WebhookPayload) error
}
