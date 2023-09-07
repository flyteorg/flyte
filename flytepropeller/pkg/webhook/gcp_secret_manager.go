package webhook

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flytepropeller/pkg/webhook/config"
	"github.com/flyteorg/flytestdlib/logger"
	corev1 "k8s.io/api/core/v1"
)

const (
	// GCPSecretsVolumeName defines the static name of the volume used for mounting/sharing secrets between init-container
	// sidecar and the rest of the containers in the pod.
	GCPSecretsVolumeName = "gcp-secret-vol" // #nosec
)

var (
	// GCPSecretMountPath defines the default mount path for secrets
	GCPSecretMountPath = filepath.Join(string(os.PathSeparator), "etc", "flyte", "secrets")
)

// GCPSecretManagerInjector allows injecting of secrets from GCP Secret Manager as files. It uses a Google Cloud
// SDK SideCar as an init-container to download the secret and save it to a local volume shared with all other
// containers in the pod. It supports multiple secrets to be mounted but that will result into adding an init
// container for each secret. The Google serviceaccount (GSA) associated with the Pod, either via Workload
// Identity (recommended) or the underlying node's serviceacccount, must have permissions to pull the secret
// from GCP Secret Manager. Currently, the secret must also exist in the same project. Otherwise, the Pod will
// fail with an init-error.
// Files will be mounted on /etc/flyte/secrets/<SecretGroup>/<SecretGroupVersion>
type GCPSecretManagerInjector struct {
	cfg config.GCPSecretManagerConfig
}

func formatGCPSecretAccessCommand(secret *core.Secret) []string {
	// `gcloud` writes this file with permission 0600.
	// This will cause permission issues in the main container when using non-root
	// users, so we fix the file permissions with `chmod`.
	secretDir := strings.ToLower(filepath.Join(GCPSecretMountPath, secret.Group))
	secretPath := strings.ToLower(filepath.Join(secretDir, secret.GroupVersion))
	args := fmt.Sprintf(
		"gcloud secrets versions access %[1]s/versions/%[2]s --out-file=%[4]s || gcloud secrets versions access %[2]s --secret=%[1]s --out-file=%[4]s; chmod +rX %[3]s %[4]s",
		secret.Group,
		secret.GroupVersion,
		secretDir,
		secretPath,
	)
	return []string{"sh", "-ec", args}
}

func formatGCPInitContainerName(index int) string {
	return fmt.Sprintf("gcp-pull-secret-%v", index)
}

func (i GCPSecretManagerInjector) Type() config.SecretManagerType {
	return config.SecretManagerTypeGCP
}

func (i GCPSecretManagerInjector) Inject(ctx context.Context, secret *core.Secret, p *corev1.Pod) (newP *corev1.Pod, injected bool, err error) {
	if len(secret.Group) == 0 || len(secret.GroupVersion) == 0 {
		return nil, false, fmt.Errorf("GCP Secrets Webhook require both group and group version to be set. "+
			"Secret: [%v]", secret)
	}

	switch secret.MountRequirement {
	case core.Secret_ANY:
		fallthrough
	case core.Secret_FILE:
		// A Volume with a static name so that if we try to inject multiple secrets, we won't mount multiple volumes.
		// We use Memory as the storage medium for volume source to avoid
		vol := corev1.Volume{
			Name: GCPSecretsVolumeName,
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{
					Medium: corev1.StorageMediumMemory,
				},
			},
		}

		p.Spec.Volumes = appendVolumeIfNotExists(p.Spec.Volumes, vol)
		p.Spec.InitContainers = append(p.Spec.InitContainers, createGCPSidecarContainer(i.cfg, p, secret))

		secretVolumeMount := corev1.VolumeMount{
			Name:      GCPSecretsVolumeName,
			ReadOnly:  true,
			MountPath: GCPSecretMountPath,
		}

		p.Spec.Containers = AppendVolumeMounts(p.Spec.Containers, secretVolumeMount)
		p.Spec.InitContainers = AppendVolumeMounts(p.Spec.InitContainers, secretVolumeMount)

		// Inject GCP secret-inject webhook annotations to mount the secret in a predictable location.
		envVars := []corev1.EnvVar{
			// Set environment variable to let the container know where to find the mounted files.
			{
				Name:  SecretPathDefaultDirEnvVar,
				Value: GCPSecretMountPath,
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

func createGCPSidecarContainer(cfg config.GCPSecretManagerConfig, p *corev1.Pod, secret *core.Secret) corev1.Container {
	return corev1.Container{
		Image: cfg.SidecarImage,
		// Create a unique name to allow multiple secrets to be mounted.
		Name:    formatGCPInitContainerName(len(p.Spec.InitContainers)),
		Command: formatGCPSecretAccessCommand(secret),
		VolumeMounts: []corev1.VolumeMount{
			{
				Name:      GCPSecretsVolumeName,
				MountPath: GCPSecretMountPath,
			},
		},
		Resources: cfg.Resources,
	}
}

// NewGCPSecretManagerInjector creates a SecretInjector that's able to mount secrets from GCP Secret Manager.
func NewGCPSecretManagerInjector(cfg config.GCPSecretManagerConfig) GCPSecretManagerInjector {
	return GCPSecretManagerInjector{
		cfg: cfg,
	}
}
