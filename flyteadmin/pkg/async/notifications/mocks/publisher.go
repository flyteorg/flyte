package mocks

import (
	"context"

	"github.com/golang/protobuf/proto"
)

type PublishFunc func(ctx context.Context, key string, msg proto.Message) error

type MockPublisher struct {
	publishFunc PublishFunc
}

func (m *MockPublisher) SetPublishCallback(publishFunction PublishFunc) {
	m.publishFunc = publishFunction
}

func (m *MockPublisher) Publish(ctx context.Context, notificationType string, msg proto.Message) error {
	if m.publishFunc != nil {
		return m.publishFunc(ctx, notificationType, msg)
	}
	return nil
}
