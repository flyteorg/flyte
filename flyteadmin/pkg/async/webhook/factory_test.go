package webhook

import (
	"testing"

	runtimeInterfaces "github.com/flyteorg/flyteadmin/pkg/runtime/interfaces"

	"github.com/flyteorg/flyteadmin/pkg/async/webhook/implementations"
	"github.com/flyteorg/flytestdlib/promutils"
)

func TestGetWebhook(t *testing.T) {
	cfg := runtimeInterfaces.WebHookConfig{
		Name: implementations.Slack,
	}
	GetWebhook(cfg, promutils.NewTestScope())
}
