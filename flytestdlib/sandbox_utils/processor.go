package sandbox_utils

import (
	"context"
	"fmt"
	"github.com/flyteorg/flyte/flytestdlib/logger"
	"github.com/golang/protobuf/proto"
	"sync"
)

var MsgChan chan SandboxMessage
var once sync.Once

func init() {
	once.Do(func() {
		MsgChan = make(chan SandboxMessage)
	})
}

type SandboxMessage struct {
	Msg   proto.Message
	Raw   []byte
	Topic string
}

type CloudEventsSandboxOnlyPublisher struct {
	subChan chan<- SandboxMessage
}

func (z *CloudEventsSandboxOnlyPublisher) Publish(ctx context.Context, topic string, msg proto.Message) error {
	logger.Debugf(ctx, "Sandbox cloud event publisher, sending to topic [%s]", topic)

	select {
	case z.subChan <- SandboxMessage{
		Msg:   msg,
		Topic: topic,
	}:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	default:
		return fmt.Errorf("channel has been closed")
	}
}

// PublishRaw will publish a raw byte array as a message with context.
func (z *CloudEventsSandboxOnlyPublisher) PublishRaw(ctx context.Context, topic string, msgRaw []byte) error {
	select {
	case z.subChan <- SandboxMessage{
		Raw:   msgRaw,
		Topic: topic,
	}:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	default:
		return fmt.Errorf("channel has been closed, can't publish raw")
	}

}

var CloudEventsPublisher = &CloudEventsSandboxOnlyPublisher{
	subChan: MsgChan,
}
