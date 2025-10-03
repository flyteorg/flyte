package sandboxutils

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/golang/protobuf/proto"

	"github.com/flyteorg/flyte/flytestdlib/logger"
)

var MsgChan chan SandboxMessage
var once sync.Once

func init() {
	once.Do(func() {
		MsgChan = make(chan SandboxMessage, 1000)
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
	sm := SandboxMessage{
		Msg:   msg,
		Topic: topic,
	}
	select {
	case z.subChan <- sm:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// PublishRaw will publish a raw byte array as a message with context.
func (z *CloudEventsSandboxOnlyPublisher) PublishRaw(ctx context.Context, topic string, msgRaw []byte) error {
	logger.Debugf(ctx, "Sandbox cloud event publisher, sending raw to topic [%s]", topic)
	sm := SandboxMessage{
		Raw:   msgRaw,
		Topic: topic,
	}
	select {
	case z.subChan <- sm:
		// metric
		logger.Debugf(ctx, "Sandbox publisher sent message to %s", topic)
		return nil
	case <-time.After(10000 * time.Millisecond):
		// metric
		logger.Errorf(context.Background(), "CloudEventsSandboxOnlyPublisher PublishRaw timed out")
		return fmt.Errorf("CloudEventsSandboxOnlyPublisher PublishRaw timed out")
	}

}

func NewCloudEventsPublisher() *CloudEventsSandboxOnlyPublisher {
	return &CloudEventsSandboxOnlyPublisher{
		subChan: MsgChan,
	}
}
