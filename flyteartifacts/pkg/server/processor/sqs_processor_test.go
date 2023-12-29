package processor

import (
	"context"
	"fmt"
	"testing"

	gizmoAWS "github.com/NYTimes/gizmo/pubsub/aws"
	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/golang/protobuf/proto"

	"github.com/flyteorg/flyte/flytestdlib/logger"
	"github.com/flyteorg/flyte/flytestdlib/promutils"
)

type EchoEventsHandler struct{}

func (e *EchoEventsHandler) HandleEvent(ctx context.Context, cloudEvent *event.Event, msg proto.Message) error {
	logger.Infof(ctx, "echo event handler received event [%s] %v", cloudEvent.Source(), msg)
	return nil
}

func TestRead(t *testing.T) {
	tt := false
	sqsConfig := gizmoAWS.SQSConfig{
		QueueName:           "staging-artifacts-cloudEventsQueue",
		QueueOwnerAccountID: "479331373192",
		ConsumeBase64:       &tt,
	}
	sqsConfig.Region = "us-east-2"
	var err error
	sub, err := gizmoAWS.NewSubscriber(sqsConfig)
	if err != nil {
		logger.Warnf(context.TODO(), "Failed to initialize new gizmo aws subscriber with config [%+v] and err: %v", sqsConfig, err)
	}

	if err != nil {
		panic(err)
	}
	scope := promutils.NewTestScope()
	echoer := EchoEventsHandler{}
	p := NewPubSubProcessor(sub, scope)
	p.StartProcessing(context.Background(), &echoer)

	fmt.Println("done")
}
