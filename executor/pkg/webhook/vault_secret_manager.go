package webhook

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	corev1 "k8s.io/api/core/v1"

	coreIdl "github.com/flyteorg/flyte/v2/gen/go/flyteidl2/core"
	"github.com/flyteorg/flyte/v2/executor/pkg/webhook/config"
	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/utils"
	"github.com/flyteorg/flyte/v2/flytestdlib/logger"
)

var (
	VaultSecretPathPrefix = []string{string(os.PathSeparator), "etc", "flyte", "secrets"}
)

// VaultSecretManagerInjector injects secrets from Hashicorp Vault by adding Vault Agent annotations to pods.
type VaultSecretManagerInjector struct {
	cfg config.VaultSecretManagerConfig
}

func (i VaultSecretManagerInjector) Type() config.SecretManagerType {
	return config.SecretManagerTypeVault
}

func (i VaultSecretManagerInjector) Inject(ctx context.Context, secret *coreIdl.Secret, p *corev1.Pod) (newP *corev1.Pod, injected bool, err error) {
	if len(secret.GetGroup()) == 0 || len(secret.GetKey()) == 0 {
		return nil, false, fmt.Errorf("Vault Secrets Webhook requires both key and group to be set. Secret: [%v]", secret)
	}

	switch secret.GetMountRequirement() {
	case coreIdl.Secret_ANY:
		fallthrough
	case coreIdl.Secret_FILE:
		defaultDirEnvVar := corev1.EnvVar{
			Name:  SecretPathDefaultDirEnvVar,
			Value: filepath.Join(VaultSecretPathPrefix...),
		}
		p.Spec.InitContainers = AppendEnvVars(p.Spec.InitContainers, defaultDirEnvVar)
		p.Spec.Containers = AppendEnvVars(p.Spec.Containers, defaultDirEnvVar)

		prefixEnvVar := corev1.EnvVar{
			Name:  SecretPathFilePrefixEnvVar,
			Value: "",
		}
		p.Spec.InitContainers = AppendEnvVars(p.Spec.InitContainers, prefixEnvVar)
		p.Spec.Containers = AppendEnvVars(p.Spec.Containers, prefixEnvVar)

		commonVaultAnnotations := map[string]string{
			"vault.hashicorp.com/agent-inject":            "true",
			"vault.hashicorp.com/secret-volume-path":      filepath.Join(VaultSecretPathPrefix...),
			"vault.hashicorp.com/role":                    i.cfg.Role,
			"vault.hashicorp.com/agent-pre-populate-only": "true",
		}

		secretVaultAnnotations := CreateVaultAnnotationsForSecret(secret, i.cfg.KVVersion)
		p.ObjectMeta.Annotations = utils.UnionMaps(secretVaultAnnotations, commonVaultAnnotations, i.cfg.Annotations, p.ObjectMeta.Annotations)

	case coreIdl.Secret_ENV_VAR:
		return p, false, fmt.Errorf("Env_Var is not a supported mount requirement for Vault Secret Manager")
	default:
		err := fmt.Errorf("unrecognized mount requirement [%v] for secret [%v]", secret.GetMountRequirement().String(), secret.GetKey())
		logger.Error(ctx, err)
		return p, false, err
	}

	return p, true, nil
}

func NewVaultSecretManagerInjector(cfg config.VaultSecretManagerConfig) VaultSecretManagerInjector {
	return VaultSecretManagerInjector{cfg: cfg}
}
