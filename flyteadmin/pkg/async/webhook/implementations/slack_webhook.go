package implementations

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyte/flytepropeller/pkg/controller/nodes/task/secretmanager"

	"github.com/flyteorg/flyte/flytestdlib/logger"

	"github.com/flyteorg/flyte/flyteadmin/pkg/async/webhook/interfaces"
	runtimeInterfaces "github.com/flyteorg/flyte/flyteadmin/pkg/runtime/interfaces"
	"github.com/flyteorg/flyte/flytestdlib/promutils"
)

const Slack = "slack"

type SlackWebhook struct {
	Config        runtimeInterfaces.WebHookConfig
	systemMetrics webhookMetrics
}

func (s *SlackWebhook) GetConfig() runtimeInterfaces.WebHookConfig {
	return s.Config
}

func (s *SlackWebhook) Post(ctx context.Context, payload admin.WebhookMessage, webhookName string) error {
	sm := secretmanager.NewFileEnvSecretManager(secretmanager.GetConfig())
	webhookURL, err := sm.Get(ctx, s.Config.URLSecretName)
	webhookURL = strings.TrimSpace(webhookURL)
	if err != nil {
		logger.Errorf(ctx, "Failed to get url from secret manager with error: %v", err)
		return err
	}

	messageKey := "text"
	data := map[string]string{
		messageKey: payload.Body,
		"channel":  "channel-name",
	}

	jsonData, err := json.Marshal(data)

	request, err := http.NewRequest("POST", webhookURL, bytes.NewBuffer(jsonData))
	if err != nil {
		logger.Errorf(ctx, "Failed to create request to Slack webhook with error: %v", err)
		return err
	}
	request.Header.Add("Content-Type", "application/json")
	if len(s.Config.TokenSecretName) != 0 {
		token, err := sm.Get(ctx, s.Config.TokenSecretName)
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
