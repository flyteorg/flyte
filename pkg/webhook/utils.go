package webhook

import (
	"path/filepath"
	"strings"

	"github.com/flyteorg/flytepropeller/pkg/utils"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"

	corev1 "k8s.io/api/core/v1"
)

func hasEnvVar(envVars []corev1.EnvVar, envVarKey string) bool {
	for _, e := range envVars {
		if e.Name == envVarKey {
			return true
		}
	}

	return false
}

func CreateEnvVarForSecret(secret *core.Secret) corev1.EnvVar {
	return corev1.EnvVar{
		Name: strings.ToUpper(K8sDefaultEnvVarPrefix + secret.Group + EnvVarGroupKeySeparator + secret.Key),
		ValueFrom: &corev1.EnvVarSource{
			SecretKeyRef: &corev1.SecretKeySelector{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: secret.Group,
				},
				Key: secret.Key,
			},
		},
	}
}

func CreateVolumeForSecret(secret *core.Secret) corev1.Volume {
	return corev1.Volume{
		Name: utils.Base32Encoder.EncodeToString([]byte(secret.Group + EnvVarGroupKeySeparator + secret.Key + EnvVarGroupKeySeparator + secret.GroupVersion)),
		VolumeSource: corev1.VolumeSource{
			Secret: &corev1.SecretVolumeSource{
				SecretName: secret.Group,
				Items: []corev1.KeyToPath{
					{
						Key:  secret.Key,
						Path: secret.Key,
					},
				},
			},
		},
	}
}

func CreateVolumeMountForSecret(volumeName string, secret *core.Secret) corev1.VolumeMount {
	return corev1.VolumeMount{
		Name:      volumeName,
		ReadOnly:  true,
		MountPath: filepath.Join(filepath.Join(K8sSecretPathPrefix...), secret.Group),
	}
}

func AppendVolumeMounts(containers []corev1.Container, mount corev1.VolumeMount) []corev1.Container {
	res := make([]corev1.Container, 0, len(containers))
	for _, c := range containers {
		c.VolumeMounts = appendVolumeMountIfNotExists(c.VolumeMounts, mount)
		res = append(res, c)
	}

	return res
}

func AppendEnvVars(containers []corev1.Container, envVar corev1.EnvVar) []corev1.Container {
	res := make([]corev1.Container, 0, len(containers))
	for _, c := range containers {
		if !hasEnvVar(c.Env, envVar.Name) {
			c.Env = append(c.Env, envVar)
		}

		res = append(res, c)
	}

	return res
}

func appendVolumeIfNotExists(volumes []corev1.Volume, vol corev1.Volume) []corev1.Volume {
	for _, v := range volumes {
		if v.Name == vol.Name {
			return volumes
		}
	}

	return append(volumes, vol)
}

func appendVolumeMountIfNotExists(volumes []corev1.VolumeMount, vol corev1.VolumeMount) []corev1.VolumeMount {
	for _, v := range volumes {
		if v.Name == vol.Name {
			return volumes
		}
	}

	return append(volumes, vol)
}
