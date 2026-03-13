package webhook

import (
	"fmt"
	"path/filepath"
	"strings"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/uuid"

	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/encoding"
	coreIdl "github.com/flyteorg/flyte/v2/gen/go/flyteidl2/core"
	"github.com/flyteorg/flyte/v2/executor/pkg/webhook/config"
)

func hasEnvVar(envVars []corev1.EnvVar, envVarKey string) bool {
	for _, e := range envVars {
		if e.Name == envVarKey {
			return true
		}
	}
	return false
}

func CreateEnvVarForSecret(secret *coreIdl.Secret) corev1.EnvVar {
	optional := true
	return corev1.EnvVar{
		Name: strings.ToUpper(K8sDefaultEnvVarPrefix + secret.GetGroup() + EnvVarGroupKeySeparator + secret.GetKey()),
		ValueFrom: &corev1.EnvVarSource{
			SecretKeyRef: &corev1.SecretKeySelector{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: secret.GetGroup(),
				},
				Key:      secret.GetKey(),
				Optional: &optional,
			},
		},
	}
}

func CreateVolumeForSecret(secret *coreIdl.Secret) corev1.Volume {
	optional := true
	return corev1.Volume{
		Name: encoding.Base32Encoder.EncodeToString([]byte(secret.GetGroup() + EnvVarGroupKeySeparator + secret.GetGroupVersion())),
		VolumeSource: corev1.VolumeSource{
			Secret: &corev1.SecretVolumeSource{
				SecretName: secret.GetGroup(),
				Items: []corev1.KeyToPath{
					{
						Key:  secret.GetKey(),
						Path: strings.ToLower(secret.GetKey()),
					},
				},
				Optional: &optional,
			},
		},
	}
}

func CreateVolumeMountForSecret(volumeName string, secret *coreIdl.Secret) corev1.VolumeMount {
	return corev1.VolumeMount{
		Name:      volumeName,
		ReadOnly:  true,
		MountPath: filepath.Join(filepath.Join(K8sSecretPathPrefix...), strings.ToLower(secret.GetGroup())),
	}
}

func CreateVolumeMountEnvVarForSecretWithEnvName(secret *coreIdl.Secret) corev1.EnvVar {
	return corev1.EnvVar{
		Name:  secret.GetEnvVar(),
		Value: filepath.Join(filepath.Join(K8sSecretPathPrefix...), strings.ToLower(secret.GetGroup()), strings.ToLower(secret.GetKey())),
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

func AppendVolume(volumes []corev1.Volume, volume corev1.Volume) []corev1.Volume {
	for _, v := range volumes {
		if v.Secret != nil && v.Secret.SecretName == volume.Secret.SecretName {
			v.Secret.Items = append(v.Secret.Items, volume.Secret.Items...)
			return volumes
		}
	}
	return append(volumes, volume)
}

func CreateVaultAnnotationsForSecret(secret *coreIdl.Secret, kvversion config.KVVersion) map[string]string {
	id := string(uuid.NewUUID())

	secretVaultAnnotations := map[string]string{
		fmt.Sprintf("vault.hashicorp.com/agent-inject-secret-%s", id): secret.GetGroup(),
		fmt.Sprintf("vault.hashicorp.com/agent-inject-file-%s", id):   fmt.Sprintf("%s/%s", secret.GetGroup(), secret.GetKey()),
	}

	var query string
	switch secret.GetGroupVersion() {
	case "kv1":
		query = ".Data"
	case "kv2":
		query = ".Data.data"
	case "db":
		// database secrets engine — no templating
	default:
		switch kvversion {
		case config.KVVersion1:
			query = ".Data"
		case config.KVVersion2:
			query = ".Data.data"
		}
	}
	if query != "" {
		template := fmt.Sprintf(`{{- with secret "%s" -}}{{ %s.%s }}{{- end -}}`, secret.GetGroup(), query, secret.GetKey())
		secretVaultAnnotations[fmt.Sprintf("vault.hashicorp.com/agent-inject-template-%s", id)] = template
	}
	return secretVaultAnnotations
}
