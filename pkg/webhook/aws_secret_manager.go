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
	// AWSSecretArnEnvVar defines the environment variable name to use to specify to the sidecar container which secret
	// to pull.
	AWSSecretArnEnvVar = "SECRET_ARN"

	// AWSSecretFilenameEnvVar defines the environment variable name to use to specify to the sidecar container where
	// to store the secret.
	AWSSecretFilenameEnvVar = "SECRET_FILENAME"

	// AWSSecretsVolumeName defines the static name of the volume used for mounting/sharing secrets between init-container
	// sidecar and the rest of the containers in the pod.
	AWSSecretsVolumeName = "aws-secret-vol" // #nosec

	// AWS SideCar Docker Container expects the mount to always be under /tmp
	AWSInitContainerMountPath = "/tmp"
)

var (
	// AWSSecretMountPathPrefix defins the default mount path for secrets
	AWSSecretMountPathPrefix = []string{string(os.PathSeparator), "etc", "flyte", "secrets"}
)

// AWSSecretManagerInjector allows injecting of secrets from AWS Secret Manager as files. It uses AWS-provided SideCar
// as an init-container to download the secret and save it to a local volume shared with all other containers in the pod.
// It supports multiple secrets to be mounted but that will result into adding an init container for each secret.
// The role/serviceaccount used to run the Pod must have permissions to pull the secret from AWS Secret Manager.
// Otherwise, the Pod will fail with an init-error.
// Files will be mounted on /etc/flyte/secrets/<SecretGroup>/<SecretKey>
type AWSSecretManagerInjector struct {
	cfg config.AWSSecretManagerConfig
}

func formatAWSSecretArn(secret *core.Secret) string {
	return strings.TrimRight(secret.Group, ":") + ":" + strings.TrimLeft(secret.Key, ":")
}

func formatAWSInitContainerName(index int) string {
	return fmt.Sprintf("aws-pull-secret-%v", index)
}

func (i AWSSecretManagerInjector) Type() config.SecretManagerType {
	return config.SecretManagerTypeAWS
}

func (i AWSSecretManagerInjector) Inject(ctx context.Context, secret *core.Secret, p *corev1.Pod) (newP *corev1.Pod, injected bool, err error) {
	if len(secret.Group) == 0 || len(secret.Key) == 0 {
		return nil, false, fmt.Errorf("AWS Secrets Webhook require both key and group to be set. "+
			"Secret: [%v]", secret)
	}

	switch secret.MountRequirement {
	case core.Secret_ANY:
		fallthrough
	case core.Secret_FILE:
		// A Volume with a static name so that if we try to inject multiple secrets, we won't mount multiple volumes.
		// We use Memory as the storage medium for volume source to avoid
		vol := corev1.Volume{
			Name: AWSSecretsVolumeName,
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{
					Medium: corev1.StorageMediumMemory,
				},
			},
		}

		p.Spec.Volumes = appendVolumeIfNotExists(p.Spec.Volumes, vol)
		p.Spec.InitContainers = append(p.Spec.InitContainers, createAWSSidecarContainer(i.cfg, p, secret))

		secretVolumeMount := corev1.VolumeMount{
			Name:      AWSSecretsVolumeName,
			ReadOnly:  true,
			MountPath: filepath.Join(AWSSecretMountPathPrefix...),
		}

		p.Spec.Containers = AppendVolumeMounts(p.Spec.Containers, secretVolumeMount)
		p.Spec.InitContainers = AppendVolumeMounts(p.Spec.InitContainers, secretVolumeMount)

		// Inject AWS secret-inject webhook annotations to mount the secret in a predictable location.
		envVars := []corev1.EnvVar{
			// Set environment variable to let the container know where to find the mounted files.
			{
				Name:  SecretPathDefaultDirEnvVar,
				Value: filepath.Join(AWSSecretMountPathPrefix...),
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

func createAWSSidecarContainer(cfg config.AWSSecretManagerConfig, p *corev1.Pod, secret *core.Secret) corev1.Container {
	return corev1.Container{
		Image: cfg.SidecarImage,
		// Create a unique name to allow multiple secrets to be mounted.
		Name: formatAWSInitContainerName(len(p.Spec.InitContainers)),
		VolumeMounts: []corev1.VolumeMount{
			{
				Name:      AWSSecretsVolumeName,
				MountPath: AWSInitContainerMountPath,
			},
		},
		Env: []corev1.EnvVar{
			{
				Name:  AWSSecretArnEnvVar,
				Value: formatAWSSecretArn(secret),
			},
			{
				Name:  AWSSecretFilenameEnvVar,
				Value: filepath.Join(string(filepath.Separator), strings.ToLower(secret.Group), strings.ToLower(secret.Key)),
			},
		},
		Resources: cfg.Resources,
	}
}

// NewAWSSecretManagerInjector creates a SecretInjector that's able to mount secrets from AWS Secret Manager.
func NewAWSSecretManagerInjector(cfg config.AWSSecretManagerConfig) AWSSecretManagerInjector {
	return AWSSecretManagerInjector{
		cfg: cfg,
	}
}
