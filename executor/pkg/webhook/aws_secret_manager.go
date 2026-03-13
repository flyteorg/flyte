package webhook

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	corev1 "k8s.io/api/core/v1"

	coreIdl "github.com/flyteorg/flyte/v2/gen/go/flyteidl2/core"
	"github.com/flyteorg/flyte/v2/executor/pkg/webhook/config"
	"github.com/flyteorg/flyte/v2/flytestdlib/logger"
)

const (
	AWSSecretArnEnvVar        = "SECRET_ARN"
	AWSSecretFilenameEnvVar   = "SECRET_FILENAME"
	AWSSecretsVolumeName      = "aws-secret-vol" // #nosec
	AWSInitContainerMountPath = "/tmp"
)

var (
	AWSSecretMountPathPrefix = []string{string(os.PathSeparator), "etc", "flyte", "secrets"}
)

// AWSSecretManagerInjector injects secrets from AWS Secret Manager as files via a sidecar init container.
type AWSSecretManagerInjector struct {
	cfg config.AWSSecretManagerConfig
}

func formatAWSSecretArn(secret *coreIdl.Secret) string {
	return strings.TrimRight(secret.GetGroup(), ":") + ":" + strings.TrimLeft(secret.GetKey(), ":")
}

func formatAWSInitContainerName(index int) string {
	return fmt.Sprintf("aws-pull-secret-%v", index)
}

func (i AWSSecretManagerInjector) Type() config.SecretManagerType {
	return config.SecretManagerTypeAWS
}

func (i AWSSecretManagerInjector) Inject(ctx context.Context, secret *coreIdl.Secret, p *corev1.Pod) (newP *corev1.Pod, injected bool, err error) {
	if len(secret.GetGroup()) == 0 || len(secret.GetKey()) == 0 {
		return nil, false, fmt.Errorf("AWS Secrets Webhook require both key and group to be set. Secret: [%v]", secret)
	}

	switch secret.GetMountRequirement() {
	case coreIdl.Secret_ANY:
		fallthrough
	case coreIdl.Secret_FILE:
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

		envVars := []corev1.EnvVar{
			{Name: SecretPathDefaultDirEnvVar, Value: filepath.Join(AWSSecretMountPathPrefix...)},
			{Name: SecretPathFilePrefixEnvVar, Value: ""},
		}
		if secret.GetEnvVar() != "" {
			envVars = append(envVars, CreateVolumeMountEnvVarForSecretWithEnvName(secret))
		}
		for _, envVar := range envVars {
			p.Spec.InitContainers = AppendEnvVars(p.Spec.InitContainers, envVar)
			p.Spec.Containers = AppendEnvVars(p.Spec.Containers, envVar)
		}

	case coreIdl.Secret_ENV_VAR:
		fallthrough
	default:
		err := fmt.Errorf("unrecognized mount requirement [%v] for secret [%v]", secret.GetMountRequirement().String(), secret.GetKey())
		logger.Error(ctx, err)
		return p, false, err
	}

	return p, true, nil
}

func createAWSSidecarContainer(cfg config.AWSSecretManagerConfig, p *corev1.Pod, secret *coreIdl.Secret) corev1.Container {
	return corev1.Container{
		Image: cfg.SidecarImage,
		Name:  formatAWSInitContainerName(len(p.Spec.InitContainers)),
		VolumeMounts: []corev1.VolumeMount{
			{Name: AWSSecretsVolumeName, MountPath: AWSInitContainerMountPath},
		},
		Env: []corev1.EnvVar{
			{Name: AWSSecretArnEnvVar, Value: formatAWSSecretArn(secret)},
			{Name: AWSSecretFilenameEnvVar, Value: filepath.Join(string(filepath.Separator), strings.ToLower(secret.GetGroup()), strings.ToLower(secret.GetKey()))},
		},
		Resources: cfg.Resources,
	}
}

func NewAWSSecretManagerInjector(cfg config.AWSSecretManagerConfig) AWSSecretManagerInjector {
	return AWSSecretManagerInjector{cfg: cfg}
}
