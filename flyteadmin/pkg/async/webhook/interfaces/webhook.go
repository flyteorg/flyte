package interfaces

import (
	"context"
	runtimeInterfaces "github.com/flyteorg/flyteadmin/pkg/runtime/interfaces"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
)

//go:generate mockery -name=Webhook -output=../mocks -case=underscore

type Payload struct {
	Value string `protobuf:"bytes,1,opt,value=value"`
}

// Webhook Defines the interface for Publishing execution event to other services (Slack).
type Webhook interface {
	// Post The notificationType is inferred from the Notification object in the Execution Spec.
	Post(ctx context.Context, payload admin.WebhookPayload) error
	GetConfig() runtimeInterfaces.WebHookConfig
}
