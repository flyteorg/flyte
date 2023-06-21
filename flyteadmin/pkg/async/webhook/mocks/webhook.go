package mocks

import (
	"context"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
)

type WebhookPostFunc func(ctx context.Context, payload admin.WebhookPayload) error

type Webhook struct {
	post WebhookPostFunc
}

func (m *Webhook) SetWebhookPostFunc(webhookPostFunc WebhookPostFunc) {
	m.post = webhookPostFunc
}

func (m *Webhook) Post(ctx context.Context, payload admin.WebhookPayload) error {
	if m.post != nil {
		return m.post(ctx, payload)
	}
	return nil
}
