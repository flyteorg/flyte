package webhook

import (
	"context"

	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/secret"
	webhookConfig "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/secret/config"
	"github.com/flyteorg/flyte/v2/flytestdlib/promutils"
)

func newSecretsMutator(ctx context.Context, cfg *webhookConfig.Config, podNamespace, limitNamespace string,
	scope promutils.Scope) (*secret.SecretsPodMutator, error) {
	return secret.NewSecretsMutatorWithOptions(ctx, cfg, podNamespace, scope, secret.WithLimitNamespace(limitNamespace))
}
