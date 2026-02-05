package secret

import (
	"context"
	"fmt"
	"path"
	"path/filepath"
	"strings"

	corev1 "k8s.io/api/core/v1"

	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/secret/config"
	"github.com/flyteorg/flyte/v2/flytestdlib/logger"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/core"
)

const (
	AzureSecretsVolumeName = "azure-secret-vol" // #nosec G101

	AzureSecretMountPath = "/etc/flyte/secrets" // #nosec G101
)

// AzureSecretManagerInjector allows injecting of secrets from Azure Key Vault as files. It uses a Azure az-cli
// SDK SideCar as an init-container to download the secret and save it to a local volume shared with all other
// containers in the pod. It supports multiple secrets to be mounted but that will result into adding an init
// container for each secret. The Azure user-assigned managed identity associated with the Pod via
// Workload Identity Federation, must have permissions to pull
// the secret from Azure Key Vault.
//
// Files will be mounted on
// - /etc/flyte/secrets/<SecretGroup>/<SecretGroupVersion> when GroupVersion is set
// - /etc/flyte/secrets/<SecretGroup> when GroupVersion is not set, retrieving the latest version.
type AzureSecretManagerInjector struct {
	cfg config.AzureSecretManagerConfig
}

func (a AzureSecretManagerInjector) Type() config.SecretManagerType {
	return config.SecretManagerTypeAzure
}

func (a AzureSecretManagerInjector) Inject(ctx context.Context, secret *core.Secret, p *corev1.Pod) (newP *corev1.Pod, injected bool, err error) {
	if len(secret.Group) == 0 { // Group version allowed to be empty to retrieve latest secret
		return p, false, fmt.Errorf("Azure Secrets Webhook require both group and group version to be set. "+
			"Secret: [%v]", secret)
	}

	switch secret.MountRequirement {
	case core.Secret_ANY:
		fallthrough
	case core.Secret_FILE:
		// A Volume with a static name so that if we try to inject multiple secrets, we won't mount multiple volumes.
		// We use Memory as the storage medium for volume source to avoid
		vol := corev1.Volume{
			Name: AzureSecretsVolumeName,
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{
					Medium: corev1.StorageMediumMemory,
				},
			},
		}

		p.Spec.Volumes = appendVolumeIfNotExists(p.Spec.Volumes, vol)
		p.Spec.InitContainers = append(p.Spec.InitContainers, createAzureSidecarContainer(a.cfg, p, secret))

		secretVolumeMount := corev1.VolumeMount{
			Name:      AzureSecretsVolumeName,
			ReadOnly:  true,
			MountPath: AzureSecretMountPath,
		}

		p.Spec.Containers = AppendVolumeMounts(p.Spec.Containers, secretVolumeMount)
		p.Spec.InitContainers = AppendVolumeMounts(p.Spec.InitContainers, secretVolumeMount)

		// Inject Azure secret-inject webhook annotations to mount the secret in a predictable location.
		envVars := []corev1.EnvVar{
			// Set environment variable to let the container know where to find the mounted files.
			{
				Name:  SecretPathDefaultDirEnvVar,
				Value: AzureSecretMountPath,
			},
			// Sets an empty prefix to let the containers know the file names will match the secret keys as-is.
			{
				Name:  SecretPathFilePrefixEnvVar,
				Value: "",
			},
		}

		for _, envVar := range envVars {
			p.Spec.InitContainers = AppendEnvVars(p.Spec.InitContainers, envVar)
			p.Spec.Containers = AppendEnvVars(p.Spec.Containers, envVar)
		}
	case core.Secret_ENV_VAR:
		fallthrough
	default:
		err := fmt.Errorf("unrecognized mount requirement [%v] for secret [%v]", secret.MountRequirement.String(), secret.Key)
		logger.Error(ctx, err)
		return p, false, err
	}

	return p, true, nil
}

func createAzureSidecarContainer(cfg config.AzureSecretManagerConfig, p *corev1.Pod, secret *core.Secret) corev1.Container {
	return corev1.Container{
		Image: cfg.SidecarImage,
		// Create a unique name to allow multiple secrets to be mounted.
		Name:    formatAzureInitContainerName(len(p.Spec.InitContainers)),
		Command: formatAzureSecretAccessCommand(secret),
		VolumeMounts: []corev1.VolumeMount{
			{
				Name:      AzureSecretsVolumeName,
				MountPath: AzureSecretMountPath,
			},
		},
		Resources: cfg.Resources,
	}
}

func formatAzureInitContainerName(index int) string {
	return fmt.Sprintf("azure-pull-secret-%v", index)
}

func formatAzureSecretAccessCommand(secret *core.Secret) []string {
	// Azure provides the entire Secret URI the Key Vault Secret.
	trimmedUri := strings.TrimRight(secret.Group, "/")
	_, path := path.Split(trimmedUri)
	segments := strings.Split(path, "/")
	secretName := segments[len(segments)-1]

	// Initialize as if GroupVersion is not set
	secretDir := AzureSecretMountPath
	secretPath := strings.ToLower(filepath.Join(secretDir, secretName))
	mkdirCmd := ""
	if len(secret.GroupVersion) > 0 { // Apply GroupVersion specific configuration
		secretDir = strings.ToLower(filepath.Join(AzureSecretMountPath, secretName))
		secretPath = strings.ToLower(filepath.Join(secretDir, secret.GroupVersion))
		mkdirCmd = fmt.Sprintf("mkdir -p %[1]s; ", secretDir)
	}

	command := "az login --service-principal -u $AZURE_CLIENT_ID -t $AZURE_TENANT_ID --federated-token \"$(cat $AZURE_FEDERATED_TOKEN_FILE)\"; " +
		mkdirCmd + "az keyvault secret show --id \"%[1]s/%[2]s\" --query \"value\" -o tsv > %[3]s"

	args := fmt.Sprintf(
		command,
		trimmedUri,
		secret.GroupVersion,
		secretPath,
	)
	return []string{"sh", "-ec", args}
}

func NewAzureSecretManagerInjector(cfg config.AzureSecretManagerConfig) AzureSecretManagerInjector {
	return AzureSecretManagerInjector{
		cfg: cfg,
	}
}
