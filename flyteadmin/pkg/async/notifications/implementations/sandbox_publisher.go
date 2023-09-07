package implementations

import (
	"context"

	"github.com/flyteorg/flytestdlib/logger"
	"github.com/golang/protobuf/proto"
)

type SandboxPublisher struct {
	pubChan chan<- []byte
}

func (p *SandboxPublisher) Publish(ctx context.Context, notificationType string, msg proto.Message) error {
	logger.Debugf(ctx, "Publishing the following message [%s]", msg.String())

	data, err := proto.Marshal(msg)

	if err != nil {
		logger.Errorf(ctx, "Failed to publish a message with key [%s] and message [%s] and error: %v", notificationType, msg.String(), err)
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
