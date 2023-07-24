package webhook

import (
	"context"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"

	secretUtils "github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/utils/secrets"

	"github.com/flyteorg/flytepropeller/pkg/controller/nodes/task/secretmanager"
	"github.com/flyteorg/flytepropeller/pkg/webhook/config"

	"github.com/flyteorg/flytestdlib/logger"
	"github.com/flyteorg/flytestdlib/promutils"

	corev1 "k8s.io/api/core/v1"
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
	Inject(ctx context.Context, secrets *core.Secret, p *corev1.Pod) (newP *corev1.Pod, injected bool, err error)
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

// NewSecretsMutator creates a new SecretsMutator with all available plugins. Depending on the selected plugins in the
// config, only the global plugin and one other plugin can be enabled.
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
