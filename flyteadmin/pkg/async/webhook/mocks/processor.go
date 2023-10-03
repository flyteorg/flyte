package mocks

import (
	"context"

	runtimeInterfaces "github.com/flyteorg/flyteadmin/pkg/runtime/interfaces"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
)

type RunFunc func() error

type StopFunc func() error

type WebhookPostFunc func(ctx context.Context, payload admin.WebhookPayload) error

type MockSubscriber struct {
	runFunc  RunFunc
	stopFunc StopFunc
}

func (m *MockSubscriber) Run() error {
	if m.runFunc != nil {
		return m.runFunc()
	}
	return nil
}

func (m *MockSubscriber) Stop() error {
	if m.stopFunc != nil {
		return m.stopFunc()
	}
	return nil
}

type MockWebhook struct {
	post WebhookPostFunc
}

func (m *MockWebhook) GetConfig() runtimeInterfaces.WebHookConfig {
	return runtimeInterfaces.WebHookConfig{Payload: "hello world"}
}

func (m *MockWebhook) SetWebhookPostFunc(webhookPostFunc WebhookPostFunc) {
	m.post = webhookPostFunc
}

func (m *MockWebhook) Post(ctx context.Context, payload admin.WebhookPayload) error {
	if m.post != nil {
		return m.post(ctx, payload)
	}
	return nil
}
