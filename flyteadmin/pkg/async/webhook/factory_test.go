package webhook

import (
	"testing"

	runtimeInterfaces "github.com/flyteorg/flyte/flyteadmin/pkg/runtime/interfaces"

	"github.com/flyteorg/flyte/flyteadmin/pkg/async/webhook/implementations"
	"github.com/flyteorg/flyte/flytestdlib/promutils"
)

func TestGetWebhook(t *testing.T) {
	cfg := runtimeInterfaces.WebHookConfig{
		Name: implementations.Slack,
	}
	GetWebhook(cfg, promutils.NewTestScope())
}
