package implementations

import (
	"context"

	"google.golang.org/protobuf/proto"

	"github.com/flyteorg/flyte/flytestdlib/logger"
)

type SandboxPublisher struct {
	pubChan chan<- []byte
}

func (p *SandboxPublisher) Publish(ctx context.Context, notificationType string, msg proto.Message) error {
	logger.Debugf(ctx, "Publishing the following message [%+v]", msg)

	data, err := proto.Marshal(msg)

	if err != nil {
		logger.Errorf(ctx, "Failed to publish a message with key [%s] and message [%+v] and error: %v", notificationType, msg, err)
		return err
	}

	p.pubChan <- data

	return nil
}

func NewSandboxPublisher(pubChan chan<- []byte) *SandboxPublisher {
	return &SandboxPublisher{
		pubChan: pubChan,
	}
}
