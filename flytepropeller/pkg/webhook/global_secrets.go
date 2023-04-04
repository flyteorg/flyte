package webhook

import (
	"context"
	"fmt"
	"strings"

	"github.com/flyteorg/flytepropeller/pkg/webhook/config"

	coreIdl "github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flytestdlib/logger"
	corev1 "k8s.io/api/core/v1"
)

//go:generate mockery -all -case=underscore

type GlobalSecretProvider interface {
	GetForSecret(ctx context.Context, secret *coreIdl.Secret) (string, error)
}

// GlobalSecrets allows the injection of secrets from the process memory space (env vars) or mounted files into pods
// intercepted through this admission webhook. Secrets injected through this type will be mounted as environment
// variables. If a secret has a mounting requirement that does not allow Env Vars, it'll fail to inject the secret.
type GlobalSecrets struct {
	envSecretManager GlobalSecretProvider
}

func (g GlobalSecrets) Type() config.SecretManagerType {
	return config.SecretManagerTypeGlobal
}

func (g GlobalSecrets) Inject(ctx context.Context, secret *coreIdl.Secret, p *corev1.Pod) (newP *corev1.Pod, injected bool, err error) {
	v, err := g.envSecretManager.GetForSecret(ctx, secret)
	if err != nil {
		return p, false, err
	}

	switch secret.MountRequirement {
	case coreIdl.Secret_FILE:
		return nil, false, fmt.Errorf("global secrets can only be injected as environment "+
			"variables [%v/%v]", secret.Group, secret.Key)
	case coreIdl.Secret_ANY:
		fallthrough
	case coreIdl.Secret_ENV_VAR:
		if len(secret.Group) == 0 {
			return nil, false, fmt.Errorf("mounting a secret to env var requires selecting the "+
				"secret and a single key within. Key [%v]", secret.Key)
		}

		envVar := corev1.EnvVar{
			Name:  strings.ToUpper(K8sDefaultEnvVarPrefix + secret.Group + EnvVarGroupKeySeparator + secret.Key),
			Value: v,
		}

		prefixEnvVar := corev1.EnvVar{
			Name:  SecretEnvVarPrefix,
			Value: K8sDefaultEnvVarPrefix,
		}

		p.Spec.InitContainers = AppendEnvVars(p.Spec.InitContainers, prefixEnvVar)
		p.Spec.Containers = AppendEnvVars(p.Spec.Containers, prefixEnvVar)

		p.Spec.InitContainers = AppendEnvVars(p.Spec.InitContainers, envVar)
		p.Spec.Containers = AppendEnvVars(p.Spec.Containers, envVar)
	default:
		err := fmt.Errorf("unrecognized mount requirement [%v] for secret [%v]", secret.MountRequirement.String(), secret.Key)
		logger.Error(ctx, err)
		return p, false, err
	}

	return p, true, nil
}

func NewGlobalSecrets(provider GlobalSecretProvider) GlobalSecrets {
	return GlobalSecrets{
		envSecretManager: provider,
	}
}
