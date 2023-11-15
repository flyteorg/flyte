package implementations

import (
	"testing"

	runtimeInterfaces "github.com/flyteorg/flyte/flyteadmin/pkg/runtime/interfaces"
	"github.com/flyteorg/flyte/flytestdlib/promutils"
	"github.com/stretchr/testify/assert"
)

func TestSlackWebhook(t *testing.T) {
	cfg := runtimeInterfaces.WebHookConfig{Name: Slack}
	webhook := NewSlackWebhook(cfg, promutils.NewTestScope())
	assert.Equal(t, webhook.GetConfig().Name, cfg.Name)
}
