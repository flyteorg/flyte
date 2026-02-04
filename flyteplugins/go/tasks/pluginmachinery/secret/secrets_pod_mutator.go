package secret

import (
	"context"
	"fmt"
	"net/http"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/secret/config"
	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/secretmanager"
	secretUtils "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/utils/secrets"
	"github.com/flyteorg/flyte/v2/flytestdlib/logger"
	"github.com/flyteorg/flyte/v2/flytestdlib/promutils"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/core"
)

const (
	SecretPathDefaultDirEnvVar = "FLYTE_SECRETS_DEFAULT_DIR" // #nosec
	SecretPathFilePrefixEnvVar = "FLYTE_SECRETS_FILE_PREFIX" // #nosec
	SecretEnvVarPrefix         = "FLYTE_SECRETS_ENV_PREFIX"  // #nosec
	SecretsID                  = "secrets"
)

const (
	// NotFoundAcrossAllScopesMsg is the error message prefix returned when a secret is not found across all scopes,
	// and is used to match on errors.
	NotFoundAcrossAllScopesMsg = "none of the secret managers injected secret"
)

type SecretsPodMutator struct {
	// Secret manager types in order that they should be used.
	enabledSecretManagerTypes []config.SecretManagerType

	// It is expected that this map contains a key for every element in enabledSecretManagerTypes.
	injectors map[config.SecretManagerType]SecretsInjector
}

func (s SecretsPodMutator) ID() string {
	return SecretsID
}

func (s *SecretsPodMutator) Mutate(ctx context.Context, pod *corev1.Pod) (newP *corev1.Pod, podChanged bool, errResponse *admission.Response) {
	secrets, err := secretUtils.UnmarshalStringMapToSecrets(pod.GetAnnotations())
	if err != nil {
		admissionError := admission.Errored(http.StatusBadRequest, fmt.Errorf("failed to unmarshal secrets from pod annotations: %w", err))
		return pod, false, &admissionError
	}

	for _, secret := range secrets {
		mutatedPod, injected, err := s.injectSecret(ctx, secret, pod)
		if !injected {
			if err == nil {
				err = fmt.Errorf("%s [%v]", NotFoundAcrossAllScopesMsg, secret)
			} else {
				err = fmt.Errorf("%s [%v]: %w", NotFoundAcrossAllScopesMsg, secret, err)
			}
			admissionError := admission.Errored(http.StatusBadRequest, err)
			return pod, false, &admissionError
		}

		pod = mutatedPod
	}

	return pod, len(secrets) > 0, nil
}

func (s *SecretsPodMutator) LabelSelector() *metav1.LabelSelector {
	return &metav1.LabelSelector{
		MatchLabels: map[string]string{
			secretUtils.PodLabel: secretUtils.PodLabelValue,
		},
	}
}

func (s *SecretsPodMutator) injectSecret(ctx context.Context, secret *core.Secret, pod *corev1.Pod) (*corev1.Pod, bool /*injected*/, error) {
	logger.Debugf(ctx, "Injecting secret [%v].", secret)
	for _, secretManagerType := range s.enabledSecretManagerTypes {
		injector := s.injectors[secretManagerType]

		mutatedPod, injected, err := injector.Inject(ctx, secret, pod)
		logger.Debugf(
			ctx,
			"injection result with injector type [%v]: injected = [%v], error = [%v].",
			injector.Type(), injected, err)

		if err != nil {
			continue
		}
		if injected {
			return mutatedPod, true, nil
		}
	}

	err := fmt.Errorf("failed to inject secret [%v] using any of the enabled secret managers: %v. Possible reasons: the secret cannot be found or an internal error occurred", secret, s.enabledSecretManagerTypes)
	return pod, false, err
}

// NewSecretsMutator creates a new SecretsMutator with all available plugins.
func NewSecretsMutator(ctx context.Context, cfg *config.Config, podNamespace string, scope promutils.Scope) (*SecretsPodMutator, error) {
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
		injector, err := newSecretsInjector(ctx, secretManagerType, cfg, globalSecretManagerConfig, podNamespace,
			scope.NewSubScope("secret_injector"))
		if err != nil {
			return nil, err
		}
		injectors[secretManagerType] = injector
	}

	return &SecretsPodMutator{
		enabledSecretManagerTypes,
		injectors,
	}, nil
}
