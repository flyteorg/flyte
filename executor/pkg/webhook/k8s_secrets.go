package webhook

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	corev1 "k8s.io/api/core/v1"

	coreIdl "github.com/flyteorg/flyte/v2/gen/go/flyteidl2/core"
	"github.com/flyteorg/flyte/v2/executor/pkg/webhook/config"
	"github.com/flyteorg/flyte/v2/flytestdlib/logger"
)

const (
	K8sDefaultEnvVarPrefix  = "_FSEC_"
	EnvVarGroupKeySeparator = "_"
)

var (
	K8sSecretPathPrefix = []string{string(os.PathSeparator), "etc", "flyte", "secrets"}
)

// K8sSecretInjector injects secrets into pods by specifying either EnvVarSource or SecretVolumeSource in the Pod Spec.
type K8sSecretInjector struct{}

func (i K8sSecretInjector) Type() config.SecretManagerType {
	return config.SecretManagerTypeK8s
}

func (i K8sSecretInjector) Inject(ctx context.Context, secret *coreIdl.Secret, p *corev1.Pod) (newP *corev1.Pod, injected bool, err error) {
	if len(secret.GetGroup()) == 0 || len(secret.GetKey()) == 0 {
		return nil, false, fmt.Errorf("k8s Secrets Webhook require both key and group to be set. Secret: [%v]", secret)
	}

	switch secret.GetMountRequirement() {
	case coreIdl.Secret_ANY:
		fallthrough
	case coreIdl.Secret_FILE:
		volume := CreateVolumeForSecret(secret)
		p.Spec.Volumes = AppendVolume(p.Spec.Volumes, volume)

		mount := CreateVolumeMountForSecret(volume.Name, secret)
		p.Spec.InitContainers = AppendVolumeMounts(p.Spec.InitContainers, mount)
		p.Spec.Containers = AppendVolumeMounts(p.Spec.Containers, mount)

		defaultDirEnvVar := corev1.EnvVar{
			Name:  SecretPathDefaultDirEnvVar,
			Value: filepath.Join(K8sSecretPathPrefix...),
		}
		p.Spec.InitContainers = AppendEnvVars(p.Spec.InitContainers, defaultDirEnvVar)
		p.Spec.Containers = AppendEnvVars(p.Spec.Containers, defaultDirEnvVar)

		prefixEnvVar := corev1.EnvVar{
			Name:  SecretPathFilePrefixEnvVar,
			Value: "",
		}
		p.Spec.InitContainers = AppendEnvVars(p.Spec.InitContainers, prefixEnvVar)
		p.Spec.Containers = AppendEnvVars(p.Spec.Containers, prefixEnvVar)

		if secret.GetEnvVar() != "" {
			extraEnvVar := CreateVolumeMountEnvVarForSecretWithEnvName(secret)
			p.Spec.InitContainers = AppendEnvVars(p.Spec.InitContainers, extraEnvVar)
			p.Spec.Containers = AppendEnvVars(p.Spec.Containers, extraEnvVar)
		}

	case coreIdl.Secret_ENV_VAR:
		envVar := CreateEnvVarForSecret(secret)
		p.Spec.InitContainers = AppendEnvVars(p.Spec.InitContainers, envVar)
		p.Spec.Containers = AppendEnvVars(p.Spec.Containers, envVar)

		if secret.GetEnvVar() != "" {
			extraEnvVar := *envVar.DeepCopy()
			extraEnvVar.Name = secret.GetEnvVar()
			p.Spec.InitContainers = AppendEnvVars(p.Spec.InitContainers, extraEnvVar)
			p.Spec.Containers = AppendEnvVars(p.Spec.Containers, extraEnvVar)
		}

		prefixEnvVar := corev1.EnvVar{
			Name:  SecretEnvVarPrefix,
			Value: K8sDefaultEnvVarPrefix,
		}
		p.Spec.InitContainers = AppendEnvVars(p.Spec.InitContainers, prefixEnvVar)
		p.Spec.Containers = AppendEnvVars(p.Spec.Containers, prefixEnvVar)

	default:
		err := fmt.Errorf("unrecognized mount requirement [%v] for secret [%v]", secret.GetMountRequirement().String(), secret.GetKey())
		logger.Error(ctx, err)
		return p, false, err
	}

	return p, true, nil
}

func NewK8sSecretsInjector() K8sSecretInjector {
	return K8sSecretInjector{}
}
