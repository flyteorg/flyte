package secret

import (
	"fmt"
	"path/filepath"
	"strings"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/uuid"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/encoding"
	"github.com/flyteorg/flyte/flytepropeller/pkg/secret/config"
)

const (
	SecretFieldSeparator              = "__"
	ValueFormatter                    = "%s"
	SecretsStorageUnionPrefix         = "u"
	SecretsOrgDelimiter               = "org"
	SecretsDomainDelimiter            = "domain"
	SecretsProjectDelimiter           = "project"
	SecretsKeyDelimiter               = "key"
	SecretsStorageOrgPrefixFormat     = SecretsStorageUnionPrefix + SecretFieldSeparator + SecretsOrgDelimiter + SecretFieldSeparator + ValueFormatter
	SecretsStorageDomainPrefixFormat  = SecretsStorageOrgPrefixFormat + SecretFieldSeparator + SecretsDomainDelimiter + SecretFieldSeparator + ValueFormatter
	SecretsStorageProjectPrefixFormat = SecretsStorageDomainPrefixFormat + SecretFieldSeparator + SecretsProjectDelimiter + SecretFieldSeparator + ValueFormatter
	SecretsStorageFormat              = SecretsStorageProjectPrefixFormat + SecretFieldSeparator + SecretsKeyDelimiter + SecretFieldSeparator + ValueFormatter
)

// If env var exists in the existing list of envVars then return the index for it or else return -1
func hasEnvVar(envVars []corev1.EnvVar, envVarKey string) int {
	for index, e := range envVars {
		if e.Name == envVarKey {
			return index
		}
	}

	return -1
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

func AppendEnvVars(containers []corev1.Container, envVars ...corev1.EnvVar) []corev1.Container {
	res := make([]corev1.Container, 0, len(containers))
	for _, c := range containers {
		for _, envVar := range envVars {
			if foundIndex := hasEnvVar(c.Env, envVar.Name); foundIndex >= 0 {
				// This would be someone adding a duplicate key to what the webhook is trying to add.
				// We should delete the existing one and then add the new at the beginning
				c.Env = append(c.Env[:foundIndex], c.Env[foundIndex+1:]...)
			}

			// Append the passed in environment variable to the start of the list.
			c.Env = append([]corev1.EnvVar{envVar}, c.Env...)
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

type SecretNameComponents struct {
	Org     string
	Domain  string
	Project string
	Name    string // Secret name
}

func EncodeSecretName(components SecretNameComponents) string {
	return fmt.Sprintf(SecretsStorageFormat, components.Org, components.Domain, components.Project, components.Name)
}

func DecodeSecretName(encodedSecretName string) (*SecretNameComponents, error) {
	parts := strings.Split(encodedSecretName, SecretFieldSeparator)

	// We need at least 5 parts: u, org, <org>, domain, <secret-name>
	if len(parts) < 9 {
		return nil, fmt.Errorf("name %s has an invalid format: not enough parts", encodedSecretName)
	}

	if parts[0] != SecretsStorageUnionPrefix || parts[1] != SecretsOrgDelimiter || parts[3] != SecretsDomainDelimiter || parts[5] != SecretsProjectDelimiter || parts[7] != SecretsKeyDelimiter {
		return nil, fmt.Errorf("name %s has an invalid format: unexpected parts", encodedSecretName)
	}

	result := &SecretNameComponents{
		Org:     parts[2],
		Domain:  parts[4],
		Project: parts[6],
		Name:    strings.Join(parts[8:], SecretFieldSeparator),
	}

	return result, nil
}
