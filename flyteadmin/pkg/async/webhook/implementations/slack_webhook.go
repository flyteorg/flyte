package implementations

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flytepropeller/pkg/controller/nodes/task/secretmanager"

	"github.com/flyteorg/flytestdlib/logger"

	"github.com/flyteorg/flyteadmin/pkg/async/webhook/interfaces"
	runtimeInterfaces "github.com/flyteorg/flyteadmin/pkg/runtime/interfaces"
	"github.com/flyteorg/flytestdlib/promutils"
)

const Slack = "slack"

type SlackWebhook struct {
	Config        runtimeInterfaces.WebHookConfig
	systemMetrics webhookMetrics
}

func (s *SlackWebhook) GetConfig() runtimeInterfaces.WebHookConfig {
	return s.Config
}

func (s *SlackWebhook) Post(ctx context.Context, payload admin.WebhookPayload) error {
	sm := secretmanager.NewFileEnvSecretManager(secretmanager.GetConfig())
	webhookURL, err := sm.Get(ctx, s.Config.URL)
	if err != nil {
		logger.Errorf(ctx, "Failed to get url from secret manager with error: %v", err)
		return err
	}
	data := []byte(fmt.Sprintf("{'text': '%s'}", payload.Message))
	request, err := http.NewRequest("POST", webhookURL, bytes.NewBuffer(data))
	if err != nil {
		logger.Errorf(ctx, "Failed to create request to Slack webhook with error: %v", err)
		return err
	}
	request.Header.Add("Content-Type", "application/json")
	if len(s.Config.Token) != 0 {
		token, err := sm.Get(ctx, s.Config.Token)
		if err != nil {
			logger.Errorf(ctx, "Failed to get bearer token from secret manager with error: %v", err)
			return err
		}
		request.Header.Add("Authorization", "Bearer "+token)
	}

	client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		logger.Errorf(ctx, "Failed to post to Slack webhook with error: %v", err)
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		respBody, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("received an error response (%d): %s",
			resp.StatusCode,
			string(respBody),
		)
	}

	return nil
}

func NewSlackWebhook(config runtimeInterfaces.WebHookConfig, scope promutils.Scope) interfaces.Webhook {

	return &SlackWebhook{
		Config:        config,
		systemMetrics: newWebhookMetrics(scope.NewSubScope("slack_webhook")),
	}
}
