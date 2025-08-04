package secret

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	k8sRuntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flytepropeller/pkg/controller/nodes/task/secretmanager"
	"github.com/flyteorg/flyte/flytepropeller/pkg/secret/config"
	"github.com/flyteorg/flyte/flytestdlib/logger"
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
	podNamespace string,
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
		kubeConfig, err := rest.InClusterConfig()
		if err != nil {
			logger.Errorf(ctx, "Failed to get kubernetes config: %v", err)
			return nil, fmt.Errorf("failed to start secret manager service due to %v", err)
		}

		// Initialize controller-runtime client
		ctrlRuntimeScheme := k8sRuntime.NewScheme()
		if err := corev1.AddToScheme(ctrlRuntimeScheme); err != nil {
			logger.Errorf(ctx, "Failed to add core v1 to scheme: %v", err)
			return nil, fmt.Errorf("failed to add core v1 to scheme: %w", err)
		}

		ctrlRuntimeClient, err := client.New(kubeConfig, client.Options{
			Scheme: ctrlRuntimeScheme,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create controller-runtime client: %w", err)
		}

		var secretFetchers []SecretFetcher
		secretFetcher, err := NewSecretFetcher(ctx, webhookConfig.EmbeddedSecretManagerConfig)
		if err != nil {
			return nil, err
		}
		secretFetchers = append(secretFetchers, secretFetcher)
		return NewEmbeddedSecretManagerInjector(webhookConfig.EmbeddedSecretManagerConfig, secretFetchers, ctrlRuntimeClient, podNamespace), nil
	case config.SecretManagerTypeAzure:
		return NewAzureSecretManagerInjector(webhookConfig.AzureSecretManagerConfig), nil
	default:
		return nil, fmt.Errorf("unrecognized secret manager type [%v]", secretManagerType)
	}
}
