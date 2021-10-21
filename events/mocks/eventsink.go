package mocks

import (
	"context"

	"github.com/golang/protobuf/proto"
)

type MockEventSink struct {
	SinkCb  func(ctx context.Context, message proto.Message) error
	CloseCb func() error
}

func (t *MockEventSink) Sink(ctx context.Context, message proto.Message) error {
	if t.SinkCb != nil {
		return t.SinkCb(ctx, message)
	}

	return nil
}

func (t *MockEventSink) Close() error {
	if t.CloseCb != nil {
		return t.CloseCb()
	}

	return nil
}

func NewMockEventSink() *MockEventSink {
	return &MockEventSink{
		SinkCb: func(ctx context.Context, message proto.Message) error {
			return nil
		},
		CloseCb: func() error {
			return nil
		},
	}
}
