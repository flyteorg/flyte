package webhook

import (
	"context"

	"github.com/flyteorg/flyteadmin/pkg/async/webhook/implementations"
	"github.com/flyteorg/flyteadmin/pkg/async/webhook/interfaces"
	runtimeInterfaces "github.com/flyteorg/flyteadmin/pkg/runtime/interfaces"
	"github.com/flyteorg/flytestdlib/promutils"
)

func NewWebhooks(ctx context.Context, config runtimeInterfaces.WebhookConfig, scope promutils.Scope) []interfaces.Webhook {

	return []interfaces.Webhook{implementations.NewSlackWebhook(config, scope)}
}
