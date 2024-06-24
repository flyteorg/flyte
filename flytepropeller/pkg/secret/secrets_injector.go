package secret

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flytepropeller/pkg/controller/nodes/task/secretmanager"
	"github.com/flyteorg/flyte/flytepropeller/pkg/secret/config"
)

//go:generate mockery --output=./mocks --case=underscore --name=SecretsInjector
type SecretsInjector interface {
	Type() config.SecretManagerType
	Inject(ctx context.Context, secrets *core.Secret, p *corev1.Pod) (newP *corev1.Pod, injected bool, err error)
}

func newSecretsInjector(
	ctx context.Context,
	secretManagerType config.SecretManagerType,
	webhookConfig *config.Config,
	globalSecretManagerConfig *secretmanager.Config,
) (SecretsInjector, error) {
	switch secretManagerType {
	case config.SecretManagerTypeGlobal:
		return NewGlobalSecrets(secretmanager.NewFileEnvSecretManager(globalSecretManagerConfig)), nil
	case config.SecretManagerTypeK8s:
		return NewK8sSecretsInjector(), nil
	case config.SecretManagerTypeAWS:
		return NewAWSSecretManagerInjector(webhookConfig.AWSSecretManagerConfig), nil
	case config.SecretManagerTypeGCP:
		return NewGCPSecretManagerInjector(webhookConfig.GCPSecretManagerConfig), nil
	case config.SecretManagerTypeVault:
		return NewVaultSecretManagerInjector(webhookConfig.VaultSecretManagerConfig), nil
	case config.SecretManagerTypeEmbedded:
		secretFetcher, err := NewSecretFetcher(ctx, webhookConfig.EmbeddedSecretManagerConfig)
		if err != nil {
			return nil, err
		}
		return NewEmbeddedSecretManagerInjector(webhookConfig.EmbeddedSecretManagerConfig, secretFetcher), nil
	default:
		return nil, fmt.Errorf("unrecognized secret manager type [%v]", secretManagerType)
	}
}
