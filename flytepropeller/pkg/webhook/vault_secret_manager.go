package webhook

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	coreIdl "github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/utils"
	"github.com/flyteorg/flytepropeller/pkg/webhook/config"
	"github.com/flyteorg/flytestdlib/logger"
	corev1 "k8s.io/api/core/v1"
)

var (
	VaultSecretPathPrefix = []string{string(os.PathSeparator), "etc", "flyte", "secrets"}
)

// VaultSecretManagerInjector allows injecting of secrets into pods by leveraging an existing deployment of Vault Agent
// Vault Agent functions as an additional webhook that is triggered through annotations and then retrieves and mounts
// the requested secrets from Vault. This injector parses a secret Request into vault annotations, interpreting the secret
// Group as the vault secret path and the secret Key as the key for which to extract a value from a Vault secret.
// It supports adding multiple secrets. (The common annotations will simply be overwritten if added several times)
// Note that you need to configure the Vault role that this injector will try to use and add Vault policies for
// the service account and namespaces that your workflows run under.
// Files will be mounted at /etc/flyte/secrets/<SecretGroup>/<SecretKey>
type VaultSecretManagerInjector struct {
	cfg config.VaultSecretManagerConfig
}

func (i VaultSecretManagerInjector) Type() config.SecretManagerType {
	return config.SecretManagerTypeVault
}

func (i VaultSecretManagerInjector) Inject(ctx context.Context, secret *coreIdl.Secret, p *corev1.Pod) (newP *corev1.Pod, injected bool, err error) {
	if len(secret.Group) == 0 || len(secret.Key) == 0 {
		return nil, false, fmt.Errorf("Vault Secrets Webhook requires both key and group to be set. "+
			"Secret: [%v]", secret)
	}

	switch secret.MountRequirement {
	case coreIdl.Secret_ANY:
		fallthrough
	case coreIdl.Secret_FILE:
		// Set environment variable to let the container know where to find the mounted files.
		defaultDirEnvVar := corev1.EnvVar{
			Name:  SecretPathDefaultDirEnvVar,
			Value: filepath.Join(VaultSecretPathPrefix...),
		}

		p.Spec.InitContainers = AppendEnvVars(p.Spec.InitContainers, defaultDirEnvVar)
		p.Spec.Containers = AppendEnvVars(p.Spec.Containers, defaultDirEnvVar)

		// Sets an empty prefix to let the containers know the file names will match the secret keys as-is.
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
		err := fmt.Errorf("unrecognized mount requirement [%v] for secret [%v]", secret.MountRequirement.String(), secret.Key)
		logger.Error(ctx, err)
		return p, false, err
	}

	return p, true, nil
}

func NewVaultSecretManagerInjector(cfg config.VaultSecretManagerConfig) VaultSecretManagerInjector {
	return VaultSecretManagerInjector{
		cfg: cfg,
	}
}
