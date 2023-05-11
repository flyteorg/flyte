package webhook

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/encoding"
	"github.com/flyteorg/flytepropeller/pkg/webhook/config"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/uuid"
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
	optional := true
	return corev1.EnvVar{
		Name: strings.ToUpper(K8sDefaultEnvVarPrefix + secret.Group + EnvVarGroupKeySeparator + secret.Key),
		ValueFrom: &corev1.EnvVarSource{
			SecretKeyRef: &corev1.SecretKeySelector{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: secret.Group,
				},
				Key:      secret.Key,
				Optional: &optional,
			},
		},
	}
}

func CreateVolumeForSecret(secret *core.Secret) corev1.Volume {
	optional := true
	return corev1.Volume{
		// we don't want to create different volume for the same secret group
		Name: encoding.Base32Encoder.EncodeToString([]byte(secret.Group + EnvVarGroupKeySeparator + secret.GroupVersion)),
		VolumeSource: corev1.VolumeSource{
			Secret: &corev1.SecretVolumeSource{
				SecretName: secret.Group,
				Items: []corev1.KeyToPath{
					{
						Key:  secret.Key,
						Path: strings.ToLower(secret.Key),
					},
				},
				Optional: &optional,
			},
		},
	}
}

func CreateVolumeMountForSecret(volumeName string, secret *core.Secret) corev1.VolumeMount {
	return corev1.VolumeMount{
		Name:      volumeName,
		ReadOnly:  true,
		MountPath: filepath.Join(filepath.Join(K8sSecretPathPrefix...), strings.ToLower(secret.Group)),
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
		// append secret items to existing volume for secret within same secret group
		if v.Secret != nil && v.Secret.SecretName == volume.Secret.SecretName {
			v.Secret.Items = append(v.Secret.Items, volume.Secret.Items...)
			return volumes
		}
	}

	return append(volumes, volume)
}

func CreateVaultAnnotationsForSecret(secret *core.Secret, kvversion config.KVVersion) map[string]string {
	// Creates three grouped annotations "agent-inject-secret", "agent-inject-file" and "agent-inject-template"
	// for a given secret request and KV engine version. The annotations respectively handle: 1. retrieving the
	// secret from a vault path specified in secret.Group, 2. storing it in a file named after secret.Group/secret.Key
	// and 3. creating a template that retrieves only secret.Key from the multiple k:v pairs present in a vault secret.
	id := string(uuid.NewUUID())

	secretVaultAnnotations := map[string]string{
		fmt.Sprintf("vault.hashicorp.com/agent-inject-secret-%s", id): secret.Group,
		fmt.Sprintf("vault.hashicorp.com/agent-inject-file-%s", id):   fmt.Sprintf("%s/%s", secret.Group, secret.Key),
	}

	// Set the consul template language query depending on the KV Secrets Engine version.
	// Version 1 stores plain k:v pairs under .Data, version 2 supports versioned secrets
	// and wraps the k:v pairs into an additional subfield.
	var query string
	switch secret.GroupVersion {
	case "kv1":
		query = ".Data"
	case "kv2":
		query = ".Data.data"
	case "db":
		// For the database secrets engine backend we do not want to use the templating
	default:
		// Deprecated: The config setting KVVersion is deprecated and will be removed in a future release.
		// You should instead use the GroupVersion field in the secret definition.
		// Support using the legacy KVVersion config if GroupVersion is not set
		switch kvversion {
		case config.KVVersion1:
			query = ".Data"
		case config.KVVersion2:
			query = ".Data.data"
		}
	}
	if query != "" {
		template := fmt.Sprintf(`{{- with secret "%s" -}}{{ %s.%s }}{{- end -}}`, secret.Group, query, secret.Key)
		secretVaultAnnotations[fmt.Sprintf("vault.hashicorp.com/agent-inject-template-%s", id)] = template

	}
	return secretVaultAnnotations
}
