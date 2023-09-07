package implementations

import (
	"context"
	"time"

	"github.com/flyteorg/flyteadmin/pkg/async"
	"github.com/flyteorg/flyteadmin/pkg/async/notifications/interfaces"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flytestdlib/logger"
	"github.com/golang/protobuf/proto"
)

type SandboxProcessor struct {
	email   interfaces.Emailer
	subChan <-chan []byte
}

func (p *SandboxProcessor) StartProcessing() {
	for {
		logger.Warningf(context.Background(), "Starting SandBox notifications processor")
		err := p.run()
		logger.Errorf(context.Background(), "error with running processor err: [%v] ", err)
		time.Sleep(async.RetryDelay)
	}
}

func (p *SandboxProcessor) run() error {
	var emailMessage admin.EmailMessage

	for {
		select {
		case msg := <-p.subChan:
			err := proto.Unmarshal(msg, &emailMessage)
			if err != nil {
				logger.Errorf(context.Background(), "error with unmarshalling message [%v]", err)
				return err
			}

			err = p.email.SendEmail(context.Background(), emailMessage)
			if err != nil {
				logger.Errorf(context.Background(), "Error sending an email message for message [%s] with emailM with err: %v", emailMessage.String(), err)
				return err
			}
		default:
			logger.Debugf(context.Background(), "no message to process")
			return nil
		}
	}
}

func (p *SandboxProcessor) StopProcessing() error {
	logger.Debug(context.Background(), "call to sandbox stop processing.")
	return nil
}

func NewSandboxProcessor(subChan <-chan []byte, emailer interfaces.Emailer) interfaces.Processor {
	return &SandboxProcessor{
		subChan: subChan,
		email:   emailer,
	}
}
