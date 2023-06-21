package webhook

import (
	"testing"

	runtimeInterfaces "github.com/flyteorg/flyteadmin/pkg/runtime/interfaces"
	"github.com/flyteorg/flytestdlib/promutils"
)

func TestGetWebhook(t *testing.T) {
	cfg := runtimeInterfaces.WebHookConfig{
		Name: "unsupported",
	}
	GetWebhook(cfg, promutils.NewTestScope())
	cfg.Name = Slack
	GetWebhook(cfg, promutils.NewTestScope())
}
