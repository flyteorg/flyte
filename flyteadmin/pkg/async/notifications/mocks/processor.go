package mocks

import (
	"context"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
)

type RunFunc func() error

type StopFunc func() error

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

type SendEmailFunc func(ctx context.Context, email admin.EmailMessage) error

type MockEmailer struct {
	sendEmailFunc SendEmailFunc
}

func (m *MockEmailer) SetSendEmailFunc(sendEmail SendEmailFunc) {
	m.sendEmailFunc = sendEmail
}

func (m *MockEmailer) SendEmail(ctx context.Context, email admin.EmailMessage) error {
	if m.sendEmailFunc != nil {
		return m.sendEmailFunc(ctx, email)
	}
	return nil
}
