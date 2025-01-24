package secret

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
	"github.com/go-test/deep"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flytepropeller/pkg/secret/config"
	"github.com/flyteorg/flyte/flytepropeller/pkg/secret/mocks"
	stdlibErrors "github.com/flyteorg/flyte/flytestdlib/errors"
	"github.com/flyteorg/flyte/flytestdlib/logger"
)

func TestEmbeddedSecretManagerInjector_Inject(t *testing.T) {
	ctx = context.Background()
	gcpClient = &mocks.GCPSecretManagerClient{}
	gcpProject = "project"
	secretIDKey := "secretID"
	secretValue := "secretValue"

	projectSecretID := fmt.Sprintf(secretsStorageFormat, OrganizationLabel, DomainLabel, ProjectLabel, secretIDKey)
	domainSecretID := fmt.Sprintf(secretsStorageFormat, OrganizationLabel, DomainLabel, EmptySecretScope, secretIDKey)
	orgSecretID := fmt.Sprintf(secretsStorageFormat, OrganizationLabel, EmptySecretScope, EmptySecretScope, secretIDKey)

	gcpClient.On("AccessSecretVersion", ctx, &secretmanagerpb.AccessSecretVersionRequest{
		Name: fmt.Sprintf(GCPSecretNameFormat, gcpProject, projectSecretID),
	}).Return(nil, stdlibErrors.Errorf(ErrCodeSecretNotFound, fmt.Sprintf(SecretNotFoundErrorFormat, projectSecretID)))
	gcpClient.On("AccessSecretVersion", ctx, &secretmanagerpb.AccessSecretVersionRequest{
		Name: fmt.Sprintf(GCPSecretNameFormat, gcpProject, domainSecretID),
	}).Return(nil, stdlibErrors.Errorf(ErrCodeSecretNotFound, fmt.Sprintf(SecretNotFoundErrorFormat, projectSecretID)))
	gcpClient.On("AccessSecretVersion", ctx, &secretmanagerpb.AccessSecretVersionRequest{
		Name: fmt.Sprintf(GCPSecretNameFormat, gcpProject, orgSecretID),
	}).Return(&secretmanagerpb.AccessSecretVersionResponse{
		Payload: &secretmanagerpb.SecretPayload{
			Data: []byte(secretValue),
		},
	}, nil)

	gcpSecretsFetcher := NewGCPSecretFetcher(config.GCPConfig{
		Project: gcpProject,
	}, gcpClient)

	injector := NewEmbeddedSecretManagerInjector(config.EmbeddedSecretManagerConfig{}, gcpSecretsFetcher)

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
									Name:  "_UNION_SECRETID",
									Value: secretValue,
								},
								{
									Name:  SecretEnvVarPrefix,
									Value: UnionSecretEnvVarPrefix,
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

func TestEmbeddedSecretManagerInjector_InjectAsFile(t *testing.T) {
	ctx = context.Background()

	type test struct {
		name   string
		secret *core.Secret
	}
	tests := []test{
		{
			name: "Without envVar",
			secret: &core.Secret{
				Key:              "secret1",
				MountRequirement: core.Secret_FILE,
			},
		},
		{
			name: "With envVar",
			secret: &core.Secret{
				Key:              "secret1",
				MountRequirement: core.Secret_FILE,
				EnvVar:           "MY_ENV_VAR",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pod := &corev1.Pod{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{},
					},
				},
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"organization": "organization",
						"project":      "project",
						"domain":       "domain",
					},
				},
			}

			injector := NewEmbeddedSecretManagerInjector(
				config.EmbeddedSecretManagerConfig{},
				secretFetcherMock{
					Secrets: map[string]SecretValue{
						"u__org__organization__domain__domain__project__project__key__secret1": {
							BinaryValue: []byte("banana"),
						},
					},
				})

			pod, injected, err := injector.Inject(ctx, tt.secret, pod)
			assert.NoError(t, err)
			assert.True(t, injected)
			assert.Len(t, pod.Spec.InitContainers, 1)

			env, found := lo.Find(
				pod.Spec.InitContainers[0].Env,
				func(env corev1.EnvVar) bool { return env.Name == "SECRETS" })
			assert.True(t, found)
			assert.Equal(t, "secret1=YmFuYW5h\n", env.Value)

			if tt.secret.GetEnvVar() != "" {
				env, found = lo.Find(
					pod.Spec.Containers[0].Env,
					func(env corev1.EnvVar) bool { return env.Name == tt.secret.GetEnvVar() })
				assert.True(t, found)
				assert.Equal(t, "/etc/flyte/secrets/secret1", env.Value)
			}

		})
	}
}

func TestEmbeddedSecretManagerInjector_InjectSecretScopedToOrganization(t *testing.T) {
	ctx = context.Background()

	type test struct {
		name   string
		secret *core.Secret
	}
	tests := []test{
		{
			name:   "Without envVar",
			secret: &core.Secret{Key: "secret1"},
		},
		{
			name:   "With envVar",
			secret: &core.Secret{Key: "secret1", EnvVar: "MY_VAR"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			pod := &corev1.Pod{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{},
					},
				},
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"organization": "o-apple",
						"domain":       "d-cherry",
						"project":      "p-banana",
					},
				},
			}

			injector := NewEmbeddedSecretManagerInjector(
				config.EmbeddedSecretManagerConfig{},
				secretFetcherMock{
					Secrets: map[string]SecretValue{
						"u__org__o-apple__domain____project____key__secret1": {
							StringValue: "fruits",
						},
					},
				})

			pod, injected, err := injector.Inject(ctx, tt.secret, pod)
			assert.NoError(t, err)
			assert.True(t, injected)
			assert.True(t, podHasSecretInjected(pod, "secret1", "fruits", tt.secret.GetEnvVar()))

		})
	}
}

func TestEmbeddedSecretManagerInjector_InjectSecretScopedToDomain(t *testing.T) {
	ctx = context.Background()
	secret := &core.Secret{Key: "secret1"}
	pod := &corev1.Pod{
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{},
			},
		},
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{
				"organization": "o-apple",
				"domain":       "d-cherry",
				"project":      "p-banana",
			},
		},
	}

	injector := NewEmbeddedSecretManagerInjector(
		config.EmbeddedSecretManagerConfig{},
		secretFetcherMock{
			Secrets: map[string]SecretValue{
				"u__org__o-apple__domain____project____key__secret1": {
					StringValue: "fruits @ org",
				},
				"u__org__o-apple__domain__d-cherry__project____key__secret1": {
					StringValue: "fruits @ domain",
				},
			},
		})

	pod, injected, err := injector.Inject(ctx, secret, pod)
	assert.NoError(t, err)
	assert.True(t, injected)
	assert.True(t, podHasSecretInjected(pod, "secret1", "fruits @ domain", ""))
}

func TestEmbeddedSecretManagerInjector_InjectSecretScopedToProject(t *testing.T) {
	ctx = context.Background()
	secret := &core.Secret{Key: "secret1"}
	pod := &corev1.Pod{
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{},
			},
		},
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{
				"organization": "o-apple",
				"domain":       "d-cherry",
				"project":      "p-banana",
			},
		},
	}

	injector := NewEmbeddedSecretManagerInjector(
		config.EmbeddedSecretManagerConfig{},
		secretFetcherMock{
			Secrets: map[string]SecretValue{
				"u__org__o-apple__domain____project____key__secret1": {
					StringValue: "fruits @ org",
				},
				"u__org__o-apple__domain__d-cherry__project____key__secret1": {
					StringValue: "fruits @ domain",
				},
				"u__org__o-apple__domain__d-cherry__project__p-banana__key__secret1": {
					StringValue: "fruits @ project",
				},
			},
		})

	pod, injected, err := injector.Inject(ctx, secret, pod)
	assert.NoError(t, err)
	assert.True(t, injected)
	assert.True(t, podHasSecretInjected(pod, "secret1", "fruits @ project", ""))
}

func podHasSecretInjected(pod *corev1.Pod, secretKey string, secretValue string, envVar string) bool {
	return lo.EveryBy(pod.Spec.Containers, func(container corev1.Container) bool {
		hasValueEnvVar := lo.ContainsBy(container.Env, func(env corev1.EnvVar) bool {
			return env.Name == ("_UNION_"+strings.ToUpper(secretKey)) &&
				env.Value == secretValue
		})
		hasPrefixEnvVar := lo.ContainsBy(container.Env, func(env corev1.EnvVar) bool {
			return env.Name == "FLYTE_SECRETS_ENV_PREFIX" && env.Value == "_UNION_"
		})
		hasCustomEnvVar := true
		if envVar != "" {
			hasCustomEnvVar = lo.ContainsBy(container.Env, func(env corev1.EnvVar) bool {
				return env.Name == envVar && env.Value == secretValue
			})
		}

		return hasValueEnvVar && hasPrefixEnvVar && hasCustomEnvVar
	})
}

type secretFetcherMock struct {
	Secrets map[string]SecretValue
}

func (f secretFetcherMock) GetSecretValue(ctx context.Context, secretID string) (*SecretValue, error) {
	v, ok := f.Secrets[secretID]
	if !ok {
		return nil, stdlibErrors.Errorf(ErrCodeSecretNotFound, "secret %q not found", secretID)
	}

	return &v, nil
}
