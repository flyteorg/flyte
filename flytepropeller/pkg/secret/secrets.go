package secret

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	secretUtils "github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/utils/secrets"
	"github.com/flyteorg/flyte/flytepropeller/pkg/controller/nodes/task/secretmanager"
	"github.com/flyteorg/flyte/flytepropeller/pkg/secret/config"
	"github.com/flyteorg/flyte/flytestdlib/logger"
	"github.com/flyteorg/flyte/flytestdlib/promutils"
)

const (
	SecretPathDefaultDirEnvVar = "FLYTE_SECRETS_DEFAULT_DIR" // #nosec
	SecretPathFilePrefixEnvVar = "FLYTE_SECRETS_FILE_PREFIX" // #nosec
	SecretEnvVarPrefix         = "FLYTE_SECRETS_ENV_PREFIX"  // #nosec
	SecretsID                  = "secrets"
)

type SecretsMutator struct {
	// Secret manager types in order that they should be used.
	enabledSecretManagerTypes []config.SecretManagerType

	// It is expected that this map contains a key for every element in enabledSecretManagerTypes.
	injectors map[config.SecretManagerType]SecretsInjector
}

type SecretsInjector interface {
	Type() config.SecretManagerType
	Inject(ctx context.Context, secrets *core.Secret, p *corev1.Pod) (newP *corev1.Pod, injected bool, err error)
}

func (s SecretsMutator) ID() string {
	return SecretsID
}

func (s *SecretsMutator) Mutate(ctx context.Context, pod *corev1.Pod) (newP *corev1.Pod, podChanged bool, errResponse *admission.Response) {
	secrets, err := secretUtils.UnmarshalStringMapToSecrets(pod.GetAnnotations())
	if err != nil {
		admissionError := admission.Errored(http.StatusBadRequest, fmt.Errorf("failed to unmarshal secrets from pod annotations: %w", err))
		return pod, false, &admissionError
	}

	for _, secret := range secrets {
		mutatedPod, injected, err := s.injectSecret(ctx, secret, pod)
		if !injected {
			if err == nil {
				err = fmt.Errorf("none of the secret managers injected secret [%v]", secret)
			} else {
				err = fmt.Errorf("none of the secret managers injected secret [%v]: %w", secret, err)
			}
			admissionError := admission.Errored(http.StatusBadRequest, err)
			return pod, false, &admissionError
		}

		pod = mutatedPod
	}

	return pod, len(secrets) > 0, nil
}

func (s *SecretsMutator) LabelSelector() *metav1.LabelSelector {
	return &metav1.LabelSelector{
		MatchLabels: map[string]string{
			secretUtils.PodLabel: secretUtils.PodLabelValue,
		},
	}
}

func (s *SecretsMutator) injectSecret(ctx context.Context, secret *core.Secret, pod *corev1.Pod) (*corev1.Pod, bool /*injected*/, error) {
	errs := make([]error, 0)

	logger.Debugf(ctx, "Injecting secret [%v].", secret)

	for _, secretManagerType := range s.enabledSecretManagerTypes {
		injector := s.injectors[secretManagerType]

		mutatedPod, injected, err := injector.Inject(ctx, secret, pod)
		logger.Debugf(
			ctx,
			"injection result with injector type [%v]: injected = [%v], error = [%v].",
			injector.Type(), injected, err)

		if err != nil {
			errs = append(errs, err)
			continue
		}
		if injected {
			return mutatedPod, true, nil
		}
	}

	return pod, false, errors.Join(errs...)
}

// NewSecretsMutator creates a new SecretsMutator with all available plugins.
func NewSecretsMutator(ctx context.Context, cfg *config.Config, _ promutils.Scope) (*SecretsMutator, error) {
	enabledSecretManagerTypes := []config.SecretManagerType{
		config.SecretManagerTypeGlobal,
	}
	if len(cfg.SecretManagerTypes) != 0 {
		enabledSecretManagerTypes = append(enabledSecretManagerTypes, cfg.SecretManagerTypes...)
	} else {
		enabledSecretManagerTypes = append(enabledSecretManagerTypes, cfg.SecretManagerType)
	}

	injectors := make(map[config.SecretManagerType]SecretsInjector, len(enabledSecretManagerTypes))
	globalSecretManagerConfig := secretmanager.GetConfig()
	for _, secretManagerType := range enabledSecretManagerTypes {
		injector, err := newSecretManager(ctx, secretManagerType, cfg, globalSecretManagerConfig)
		if err != nil {
			return nil, err
		}
		injectors[secretManagerType] = injector
	}

	return &SecretsMutator{
		enabledSecretManagerTypes,
		injectors,
	}, nil
}

func newSecretManager(
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
		secretFetcher, err := NewSecretFetcherManager(ctx, webhookConfig.EmbeddedSecretManagerConfig)
		if err != nil {
			return nil, err
		}
		return NewEmbeddedSecretManagerInjector(webhookConfig.EmbeddedSecretManagerConfig, secretFetcher), nil
	default:
		return nil, fmt.Errorf("unrecognized secret manager type [%v]", secretManagerType)
	}
}
