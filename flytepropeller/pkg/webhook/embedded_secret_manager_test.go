package webhook

import (
	"context"
	"fmt"
	"testing"

	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
	"github.com/go-test/deep"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flytepropeller/pkg/webhook/config"
	"github.com/flyteorg/flyte/flytepropeller/pkg/webhook/mocks"
	stdlibErrors "github.com/flyteorg/flyte/flytestdlib/errors"
	"github.com/flyteorg/flyte/flytestdlib/logger"
)

func TestEmbeddedSecretManagerInjector_Inject(t *testing.T) {
	ctx = context.Background()
	gcpClient = &mocks.GCPSecretsIface{}
	gcpProject = "project"
	secretIDKey := "secretID"
	secretValue := "secretValue"

	secretID := fmt.Sprintf(SecretsStorageFormat, OrganizationLabel, DomainLabel, ProjectLabel, secretIDKey)
	gcpClient.OnAccessSecretVersionMatch(ctx, &secretmanagerpb.AccessSecretVersionRequest{
		Name: fmt.Sprintf(GCPSecretNameFormat, gcpProject, secretID),
	}).Return(&secretmanagerpb.AccessSecretVersionResponse{
		Payload: &secretmanagerpb.SecretPayload{
			Data: []byte(secretValue),
		},
	}, nil)

	gcpSecretsFetcher := NewGCPSecretFetcher(config.GCPConfig{
		Project: gcpProject,
	}, gcpClient)

	injector := NewEmbeddedSecretManagerInjector(config.EmbeddedSecretManagerConfig{
		Enabled: true,
	}, gcpSecretsFetcher)

	inputSecret := &core.Secret{
		Key: secretIDKey,
	}
	type test struct {
		name             string
		pod              *corev1.Pod
		expectedPod      *corev1.Pod
		expectedInjected bool
		expectedError    error
	}

	tests := []test{
		{
			name: "empty organization",
			pod: &corev1.Pod{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{},
				},
			},
			expectedPod: &corev1.Pod{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{},
				},
			},
			expectedInjected: false,
			expectedError:    stdlibErrors.Errorf(ErrCodeSecretRequirementsError, fmt.Sprintf(SecretRequirementsErrorFormat, OrganizationLabel)),
		},
		{
			name: "empty project",
			pod: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						OrganizationLabel: OrganizationLabel,
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{},
				},
			},
			expectedPod: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						OrganizationLabel: OrganizationLabel,
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{},
				},
			},
			expectedInjected: false,
			expectedError:    stdlibErrors.Errorf(ErrCodeSecretRequirementsError, fmt.Sprintf(SecretRequirementsErrorFormat, ProjectLabel)),
		},
		{
			name: "empty domain",
			pod: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						OrganizationLabel: OrganizationLabel,
						ProjectLabel:      ProjectLabel,
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{},
				},
			},
			expectedPod: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						OrganizationLabel: OrganizationLabel,
						ProjectLabel:      ProjectLabel,
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{},
				},
			},
			expectedInjected: false,
			expectedError:    stdlibErrors.Errorf(ErrCodeSecretRequirementsError, fmt.Sprintf(SecretRequirementsErrorFormat, DomainLabel)),
		},
		{
			name: "all labels",
			pod: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						OrganizationLabel: OrganizationLabel,
						ProjectLabel:      ProjectLabel,
						DomainLabel:       DomainLabel,
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{},
					},
				},
			},
			expectedPod: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						OrganizationLabel: OrganizationLabel,
						ProjectLabel:      ProjectLabel,
						DomainLabel:       DomainLabel,
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Env: []corev1.EnvVar{
								{
									Name:  SecretEnvVarPrefix,
									Value: UnionSecretEnvVarPrefix,
								},
								{
									Name:  "_UNION_SECRETID",
									Value: secretValue,
								},
							},
						},
					},
					InitContainers: []corev1.Container{},
				},
			},
			expectedInjected: true,
			expectedError:    nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			actualP, injected, err := injector.Inject(ctx, inputSecret, tt.pod)
			assert.Equal(t, tt.expectedInjected, injected)
			assert.Equal(t, tt.expectedError, err)
			if diff := deep.Equal(actualP, tt.expectedPod); diff != nil {
				logger.Info(ctx, actualP)
				assert.Fail(t, "actual != expected", "Diff: %v", diff)
			}
		})
	}
}
