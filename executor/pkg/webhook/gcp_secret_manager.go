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
	GCPSecretsVolumeName = "gcp-secret-vol" // #nosec
)

var (
	GCPSecretMountPath = filepath.Join(string(os.PathSeparator), "etc", "flyte", "secrets")
)

// GCPSecretManagerInjector injects secrets from GCP Secret Manager as files via a Google Cloud SDK sidecar init container.
type GCPSecretManagerInjector struct {
	cfg config.GCPSecretManagerConfig
}

func formatGCPSecretAccessCommand(secret *coreIdl.Secret) []string {
	secretDir := strings.ToLower(filepath.Join(GCPSecretMountPath, secret.GetGroup()))
	secretPath := strings.ToLower(filepath.Join(secretDir, secret.GetGroupVersion()))
	args := fmt.Sprintf(
		"gcloud secrets versions access %[1]s/versions/%[2]s --out-file=%[4]s || gcloud secrets versions access %[2]s --secret=%[1]s --out-file=%[4]s; chmod +rX %[3]s %[4]s",
		secret.GetGroup(), secret.GetGroupVersion(), secretDir, secretPath,
	)
	return []string{"sh", "-ec", args}
}

func formatGCPInitContainerName(index int) string {
	return fmt.Sprintf("gcp-pull-secret-%v", index)
}

func (i GCPSecretManagerInjector) Type() config.SecretManagerType {
	return config.SecretManagerTypeGCP
}

func (i GCPSecretManagerInjector) Inject(ctx context.Context, secret *coreIdl.Secret, p *corev1.Pod) (newP *corev1.Pod, injected bool, err error) {
	if len(secret.GetGroup()) == 0 || len(secret.GetGroupVersion()) == 0 {
		return nil, false, fmt.Errorf("GCP Secrets Webhook require both group and group version to be set. Secret: [%v]", secret)
	}

	switch secret.GetMountRequirement() {
	case coreIdl.Secret_ANY:
		fallthrough
	case coreIdl.Secret_FILE:
		vol := corev1.Volume{
			Name: GCPSecretsVolumeName,
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{Medium: corev1.StorageMediumMemory},
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

		envVars := []corev1.EnvVar{
			{Name: SecretPathDefaultDirEnvVar, Value: GCPSecretMountPath},
			{Name: SecretPathFilePrefixEnvVar, Value: ""},
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

func createGCPSidecarContainer(cfg config.GCPSecretManagerConfig, p *corev1.Pod, secret *coreIdl.Secret) corev1.Container {
	return corev1.Container{
		Image:   cfg.SidecarImage,
		Name:    formatGCPInitContainerName(len(p.Spec.InitContainers)),
		Command: formatGCPSecretAccessCommand(secret),
		VolumeMounts: []corev1.VolumeMount{
			{Name: GCPSecretsVolumeName, MountPath: GCPSecretMountPath},
		},
		Resources: cfg.Resources,
	}
}

func NewGCPSecretManagerInjector(cfg config.GCPSecretManagerConfig) GCPSecretManagerInjector {
	return GCPSecretManagerInjector{cfg: cfg}
}
