package webhook

import (
	"context"

	corev1 "k8s.io/api/core/v1"

	"github.com/flyteorg/flyte/v2/executor/pkg/webhook/config"
	secretUtils "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/utils/secrets"
	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/secretmanager"
	"github.com/flyteorg/flyte/v2/flytestdlib/logger"
	"github.com/flyteorg/flyte/v2/flytestdlib/promutils"
	coreIdl "github.com/flyteorg/flyte/v2/gen/go/flyteidl2/core"
)

const (
	SecretPathDefaultDirEnvVar = "FLYTE_SECRETS_DEFAULT_DIR" // #nosec
	SecretPathFilePrefixEnvVar = "FLYTE_SECRETS_FILE_PREFIX" // #nosec
	SecretEnvVarPrefix         = "FLYTE_SECRETS_ENV_PREFIX"  // #nosec
)

type SecretsMutator struct {
	cfg       *config.Config
	injectors []SecretsInjector
}

type SecretsInjector interface {
	Type() config.SecretManagerType
	Inject(ctx context.Context, secrets *coreIdl.Secret, p *corev1.Pod) (newP *corev1.Pod, injected bool, err error)
}

func (s SecretsMutator) ID() string {
	return "secrets"
}

func (s *SecretsMutator) Mutate(ctx context.Context, p *corev1.Pod) (newP *corev1.Pod, injected bool, err error) {
	secrets, err := secretUtils.UnmarshalStringMapToSecrets(p.GetAnnotations())
	if err != nil {
		return p, false, err
	}

	for _, secret := range secrets {
		for _, injector := range s.injectors {
			if injector.Type() != config.SecretManagerTypeGlobal && injector.Type() != s.cfg.SecretManagerType {
				logger.Infof(ctx, "Skipping SecretManager [%v] since it's not enabled.", injector.Type())
				continue
			}

			p, injected, err = injector.Inject(ctx, secret, p)
			if err != nil {
				logger.Infof(ctx, "Failed to inject a secret using injector [%v]. Error: %v", injector.Type(), err)
			} else if injected {
				break
			}
		}

		if err != nil {
			return p, false, err
		}
	}

	return p, injected, nil
}

// NewSecretsMutator creates a new SecretsMutator with all available injectors.
func NewSecretsMutator(cfg *config.Config, _ promutils.Scope) *SecretsMutator {
	return &SecretsMutator{
		cfg: cfg,
		injectors: []SecretsInjector{
			NewGlobalSecrets(secretmanager.NewFileEnvSecretManager(secretmanager.GetConfig())),
			NewK8sSecretsInjector(),
			NewAWSSecretManagerInjector(cfg.AWSSecretManagerConfig),
			NewGCPSecretManagerInjector(cfg.GCPSecretManagerConfig),
			NewVaultSecretManagerInjector(cfg.VaultSecretManagerConfig),
		},
	}
}
