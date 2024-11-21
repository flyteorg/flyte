package webhook

import (
	"context"
	"testing"

	"github.com/go-test/deep"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flytepropeller/pkg/webhook/config"
)

func TestAWSSecretManagerInjector_Inject(t *testing.T) {
	injector := NewAWSSecretManagerInjector(config.DefaultConfig.AWSSecretManagerConfig)
	p := &corev1.Pod{
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{},
		},
	}
	inputSecret := &core.Secret{
		Group: "arn",
		Key:   "name",
	}

	expected := &corev1.Pod{
		Spec: corev1.PodSpec{
			Volumes: []corev1.Volume{
				{
					Name: "aws-secret-vol",
					VolumeSource: corev1.VolumeSource{
						EmptyDir: &corev1.EmptyDirVolumeSource{
							Medium: corev1.StorageMediumMemory,
						},
					},
				},
			},

			InitContainers: []corev1.Container{
				{
					Name:  "aws-pull-secret-0",
					Image: "docker.io/amazon/aws-secrets-manager-secret-uploader:v0.1.4",
					Env: []corev1.EnvVar{
						{
							Name:  "SECRET_ARN",
							Value: inputSecret.GetGroup() + ":" + inputSecret.GetKey(),
						},
						{
							Name:  "SECRET_FILENAME",
							Value: "/" + inputSecret.GetGroup() + "/" + inputSecret.GetKey(),
						},
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
							Name:      "aws-secret-vol",
							MountPath: "/tmp",
						},
					},
					Resources: config.DefaultConfig.AWSSecretManagerConfig.Resources,
				},
			},
			Containers: []corev1.Container{},
		},
	}

	actualP, injected, err := injector.Inject(context.Background(), inputSecret, p.DeepCopy())
	assert.NoError(t, err)
	assert.True(t, injected)
	if diff := deep.Equal(actualP, expected); diff != nil {
		assert.Fail(t, "actual != expected", "Diff: %v", diff)
	}
}
