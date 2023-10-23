package processor

import (
	"context"
	pbcloudevents "github.com/cloudevents/sdk-go/binding/format/protobuf/v2"
	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/flyteorg/flyte/flytestdlib/logger"
	"github.com/flyteorg/flyte/flytestdlib/sandbox_utils"
	"time"
)

type SandboxCloudEventsReceiver struct {
	subChan <-chan sandbox_utils.SandboxMessage
	Handler EventsHandlerInterface
}

func (p *SandboxCloudEventsReceiver) StartProcessing(ctx context.Context) {
	for {
		logger.Warningf(context.Background(), "Starting SandBox notifications processor")
		err := p.run(ctx)
		if err != nil {
			logger.Errorf(context.Background(), "error with running processor err: [%v], sleeping and restarting", err)
			time.Sleep(1000 * 1000 * 1000 * 1)
		}
		break
	}
	logger.Warning(context.Background(), "Sandbox cloud event processor has stopped because context cancelled")
}

func (p *SandboxCloudEventsReceiver) run(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			logger.Debug(context.Background(), "Context cancelled, stopping processing.")
			return nil

		case sandboxMsg := <-p.subChan:
			logger.Debugf(ctx, "received message [%v]", sandboxMsg)
			if sandboxMsg.Raw != nil {
				ce := &event.Event{}
				err := pbcloudevents.Protobuf.Unmarshal(sandboxMsg.Raw, ce)
				if err != nil {
					logger.Errorf(context.Background(), "error with unmarshalling message [%v]", err)
					return err
				}
				logger.Debugf(ctx, "Cloud event received message [%+v]", ce)
			}
			logger.Infof(ctx, "sandbox receiver ignoring message [%v]", sandboxMsg)
		}
	}
}

func (p *SandboxCloudEventsReceiver) StopProcessing() error {
	logger.Debug(context.Background(), "call to sandbox stop processing.")
	return nil
}

func NewSandboxCloudEventProcessor(eventsHandler EventsHandlerInterface) *SandboxCloudEventsReceiver {
	return &SandboxCloudEventsReceiver{
		Handler: eventsHandler,
		subChan: sandbox_utils.MsgChan,
	}
}
