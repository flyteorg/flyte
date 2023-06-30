package webhook

import (
	"context"
	"testing"

	"github.com/flyteorg/flytepropeller/pkg/webhook/config"

	"github.com/go-test/deep"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
)

func TestGCPSecretManagerInjector_Inject(t *testing.T) {
	injector := NewGCPSecretManagerInjector(config.DefaultConfig.GCPSecretManagerConfig)
	inputSecret := &core.Secret{
		Group:        "TestSecret",
		GroupVersion: "2",
	}

	expected := &corev1.Pod{
		Spec: corev1.PodSpec{
			Volumes: []corev1.Volume{
				{
					Name: "gcp-secret-vol",
					VolumeSource: corev1.VolumeSource{
						EmptyDir: &corev1.EmptyDirVolumeSource{
							Medium: corev1.StorageMediumMemory,
						},
					},
				},
			},

			InitContainers: []corev1.Container{
				{
					Name:  "gcp-pull-secret-0",
					Image: "gcr.io/google.com/cloudsdktool/cloud-sdk:alpine",
					Command: []string{
						"sh",
						"-ec",
						"gcloud secrets versions access TestSecret/versions/2 --out-file=/etc/flyte/secrets/testsecret/2 || gcloud secrets versions access 2 --secret=TestSecret --out-file=/etc/flyte/secrets/testsecret/2; chmod +rX /etc/flyte/secrets/testsecret /etc/flyte/secrets/testsecret/2",
					},
					Env: []corev1.EnvVar{
						{
							Name:  "FLYTE_SECRETS_DEFAULT_DIR",
							Value: "/etc/flyte/secrets",
						},
						{
							Name:  "FLYTE_SECRETS_FILE_PREFIX",
							Value: "",
						},
					},
					VolumeMounts: []corev1.VolumeMount{
						{
							Name:      "gcp-secret-vol",
							MountPath: "/etc/flyte/secrets",
						},
					},
					Resources: config.DefaultConfig.GCPSecretManagerConfig.Resources,
				},
			},
			Containers: []corev1.Container{},
		},
	}

	p := &corev1.Pod{
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{},
		},
	}
	actualP, injected, err := injector.Inject(context.Background(), inputSecret, p)
	assert.NoError(t, err)
	assert.True(t, injected)
	if diff := deep.Equal(actualP, expected); diff != nil {
		assert.Fail(t, "actual != expected", "Diff: %v", diff)
	}
}
