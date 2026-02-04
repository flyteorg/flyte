package secret

import (
	"context"
	"testing"

	"github.com/go-test/deep"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"

	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/secret/config"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/core"
)

func TestAzureSecretManagerInjector_Inject(t *testing.T) {
	injector := NewAzureSecretManagerInjector(config.DefaultConfig.AzureSecretManagerConfig)

	tests := []struct {
		name            string
		secret          *core.Secret
		expectedCommand string
	}{
		{
			"test secret with explicit version",
			&core.Secret{
				Group:        "https://test-storage-account.vault.azure.net/secrets/test-secret-name",
				GroupVersion: "test-version",
			},
			"az login --service-principal -u $AZURE_CLIENT_ID -t $AZURE_TENANT_ID --federated-token \"$(cat $AZURE_FEDERATED_TOKEN_FILE)\"; " +
				"mkdir -p /etc/flyte/secrets/test-secret-name; az keyvault secret show --id \"https://test-storage-account.vault.azure.net/secrets/test-secret-name/test-version\" --query \"value\" -o tsv > /etc/flyte/secrets/test-secret-name/test-version",
		},
		{
			"test secret with ending slash",
			&core.Secret{
				Group:        "https://test-storage-account.vault.azure.net/secrets/test-secret-name/",
				GroupVersion: "test-version",
			},
			"az login --service-principal -u $AZURE_CLIENT_ID -t $AZURE_TENANT_ID --federated-token \"$(cat $AZURE_FEDERATED_TOKEN_FILE)\"; " +
				"mkdir -p /etc/flyte/secrets/test-secret-name; az keyvault secret show --id \"https://test-storage-account.vault.azure.net/secrets/test-secret-name/test-version\" --query \"value\" -o tsv > /etc/flyte/secrets/test-secret-name/test-version",
		},
		{
			"test secret with no version defaults to latest",
			&core.Secret{
				Group: "https://test-storage-account.vault.azure.net/secrets/test-secret-name",
			},
			"az login --service-principal -u $AZURE_CLIENT_ID -t $AZURE_TENANT_ID --federated-token \"$(cat $AZURE_FEDERATED_TOKEN_FILE)\"; " +
				"az keyvault secret show --id \"https://test-storage-account.vault.azure.net/secrets/test-secret-name/\" --query \"value\" -o tsv > /etc/flyte/secrets/test-secret-name",
		},
	}

	for _, tt := range tests {
		p := &corev1.Pod{
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{},
			},
		}

		t.Run(tt.name, func(t *testing.T) {
			expectedPod := expectedInitPod(tt.expectedCommand)
			actualPod, injected, err := injector.Inject(context.Background(), tt.secret, p)
			assert.NoError(t, err)
			assert.True(t, injected)
			if diff := deep.Equal(actualPod, expectedPod); diff != nil {
				assert.Fail(t, "actual != expected", "Diff: %v", diff)
			}
		})
	}
}

func expectedInitPod(command string) *corev1.Pod {
	return &corev1.Pod{
		Spec: corev1.PodSpec{
			Volumes: []corev1.Volume{
				{
					Name: "azure-secret-vol",
					VolumeSource: corev1.VolumeSource{
						EmptyDir: &corev1.EmptyDirVolumeSource{
							Medium: corev1.StorageMediumMemory,
						},
					},
				},
			},

			InitContainers: []corev1.Container{
				{
					Name:  "azure-pull-secret-0",
					Image: "mcr.microsoft.com/azure-cli:cbl-mariner2.0",
					Command: []string{
						"sh",
						"-ec",
						command,
					},
					Env: []corev1.EnvVar{
						{
							Name:  "FLYTE_SECRETS_FILE_PREFIX",
							Value: "",
						},
						{
							Name:  "FLYTE_SECRETS_DEFAULT_DIR",
							Value: "/etc/flyte/secrets",
						},
					},
					VolumeMounts: []corev1.VolumeMount{
						{
							Name:      "azure-secret-vol",
							MountPath: "/etc/flyte/secrets",
						},
					},
					Resources: config.DefaultConfig.AzureSecretManagerConfig.Resources,
				},
			},
			Containers: []corev1.Container{},
		},
	}
}
