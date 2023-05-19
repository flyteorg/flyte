package implementations

import (
	"bytes"
	"context"
	"fmt"
	"net/http"

	"github.com/flyteorg/flyteadmin/pkg/async/webhook/interfaces"
	runtimeInterfaces "github.com/flyteorg/flyteadmin/pkg/runtime/interfaces"
	"github.com/flyteorg/flytestdlib/promutils"
	"github.com/golang/protobuf/proto"
)

type SlackWebhook struct {
	config        runtimeInterfaces.WebhookConfig
	systemMetrics webhookMetrics
}

func (s *SlackWebhook) Post(ctx context.Context, notificationType string, msg proto.Message) error {
	// curl -X POST -H 'Content-type: application/json' --data '{"text":"Hello, World!"}' https://hooks.slack.com/services/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
	webhookURL := "https://hooks.slack.com/services/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"
	data := []byte(fmt.Sprintf("{ hello: world }"))
	request, err := http.NewRequest("POST", webhookURL, bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	request.Header.Add("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	// TODO: Check response status code
	return nil
}

func NewSlackWebhook(config runtimeInterfaces.WebhookConfig, scope promutils.Scope) interfaces.Webhook {

	return &SlackWebhook{
		config:        config,
		systemMetrics: newWebhookMetrics(scope.NewSubScope("slack_webhook")),
	}
}
