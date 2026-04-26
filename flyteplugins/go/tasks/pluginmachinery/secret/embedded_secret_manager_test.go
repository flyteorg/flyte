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
	"github.com/stretchr/testify/mock"
	corev1 "k8s.io/api/core/v1"
	k8sError "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/secret/config"
	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/secret/mocks"
	cacheMocks "github.com/flyteorg/flyte/v2/flytestdlib/cache/gocachemocks"
	stdlibErrors "github.com/flyteorg/flyte/v2/flytestdlib/errors"
	"github.com/flyteorg/flyte/v2/flytestdlib/logger"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/core"
)

const (
	testReferenceNamespace = "test-reference-namespace"
)

func newAlwaysMissCache(t *testing.T) *cacheMocks.CacheInterface[SecretValue] {
	t.Helper()
	m := cacheMocks.NewCacheInterface[SecretValue](t)
	m.EXPECT().Get(mock.Anything, mock.Anything).Return(SecretValue{}, fmt.Errorf("cache miss")).Maybe()
	m.EXPECT().Set(mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()
	m.EXPECT().Delete(mock.Anything, mock.Anything).Return(nil).Maybe()
	m.EXPECT().Invalidate(mock.Anything).Return(nil).Maybe()
	m.EXPECT().Clear(mock.Anything).Return(nil).Maybe()
	m.EXPECT().GetType().Return("mock").Maybe()
	return m
}

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
	}).Return(nil, stdlibErrors.Errorf(ErrCodeSecretNotFound, SecretNotFoundErrorFormat, projectSecretID))
	gcpClient.On("AccessSecretVersion", ctx, &secretmanagerpb.AccessSecretVersionRequest{
		Name: fmt.Sprintf(GCPSecretNameFormat, gcpProject, domainSecretID),
	}).Return(nil, stdlibErrors.Errorf(ErrCodeSecretNotFound, SecretNotFoundErrorFormat, projectSecretID))
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

	inputSecret := &core.Secret{
		Key: secretIDKey,
	}
	type test struct {
		name                  string
		pod                   *corev1.Pod
		expectedPod           *corev1.Pod
		expectedK8sSecretName string
		expectedInjected      bool
		expectedError         error
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
			expectedK8sSecretName: "",
			expectedInjected:      false,
			expectedError:         stdlibErrors.Errorf(ErrCodeSecretRequirementsError, SecretRequirementsErrorFormat, OrganizationLabel),
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
			expectedK8sSecretName: "",
			expectedInjected:      false,
			expectedError:         stdlibErrors.Errorf(ErrCodeSecretRequirementsError, SecretRequirementsErrorFormat, ProjectLabel),
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
			expectedK8sSecretName: "",
			expectedInjected:      false,
			expectedError:         stdlibErrors.Errorf(ErrCodeSecretRequirementsError, SecretRequirementsErrorFormat, DomainLabel),
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
									Value: config.DefaultSecretEnvVarPrefix,
								},
							},
						},
					},
					InitContainers: []corev1.Container{},
				},
			},
			expectedK8sSecretName: ToImagePullK8sName(SecretNameComponents{
				Org:     OrganizationLabel,
				Domain:  "",
				Project: "",
				Name:    secretIDKey,
			}),
			expectedInjected: true,
			expectedError:    nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &mocks.MockableControllerRuntimeClient{}

			if tt.expectedInjected {
				mockClient.
					On("Get", ctx, types.NamespacedName{Name: tt.expectedK8sSecretName, Namespace: testReferenceNamespace}, &corev1.Secret{}).
					Return(k8sError.NewNotFound(corev1.Resource("secret"), tt.expectedK8sSecretName))
			}

			secretCache := newAlwaysMissCache(t)
			parentCfg := &config.Config{SecretEnvVarPrefix: config.DefaultSecretEnvVarPrefix}
			injector := NewEmbeddedSecretManagerInjector(config.EmbeddedSecretManagerConfig{}, []SecretFetcher{gcpSecretsFetcher}, mockClient, testReferenceNamespace, secretCache, parentCfg)

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

			mockClient := &mocks.MockableControllerRuntimeClient{}
			kubernetesSecretName := ToImagePullK8sName(SecretNameComponents{
				Org:     "organization",
				Domain:  "domain",
				Project: "project",
				Name:    "secret1",
			})
			mockClient.
				On("Get", ctx, types.NamespacedName{Name: kubernetesSecretName, Namespace: testReferenceNamespace}, &corev1.Secret{}).
				Return(k8sError.NewNotFound(corev1.Resource("secret"), kubernetesSecretName))

			secretCache := newAlwaysMissCache(t)
			injector := NewEmbeddedSecretManagerInjector(
				config.EmbeddedSecretManagerConfig{},
				[]SecretFetcher{secretFetcherMock{
					Secrets: map[string]SecretValue{
						"u__org__organization__domain__domain__project__project__key__secret1": {
							BinaryValue: []byte("banana"),
						},
					},
				}}, mockClient, testReferenceNamespace, secretCache, &config.Config{SecretEnvVarPrefix: config.DefaultSecretEnvVarPrefix})

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

			mockClient := &mocks.MockableControllerRuntimeClient{}
			kubernetesSecretName := ToImagePullK8sName(SecretNameComponents{
				Org:     "o-apple",
				Domain:  "",
				Project: "",
				Name:    "secret1",
			})
			mockClient.
				On("Get", ctx, types.NamespacedName{Name: kubernetesSecretName, Namespace: testReferenceNamespace}, &corev1.Secret{}).
				Return(k8sError.NewNotFound(corev1.Resource("secret"), kubernetesSecretName))

			secretCache := newAlwaysMissCache(t)
			injector := NewEmbeddedSecretManagerInjector(
				config.EmbeddedSecretManagerConfig{},
				[]SecretFetcher{secretFetcherMock{
					Secrets: map[string]SecretValue{
						"u__org__o-apple__domain____project____key__secret1": {
							StringValue: "fruits",
						},
					},
				}}, mockClient, testReferenceNamespace, secretCache, &config.Config{SecretEnvVarPrefix: config.DefaultSecretEnvVarPrefix})

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

	mockClient := &mocks.MockableControllerRuntimeClient{}
	kubernetesSecretName1 := ToImagePullK8sName(SecretNameComponents{
		Org:     "o-apple",
		Domain:  "",
		Project: "",
		Name:    "secret1",
	})
	kubernetesSecretName2 := ToImagePullK8sName(SecretNameComponents{
		Org:     "o-apple",
		Domain:  "d-cherry",
		Project: "",
		Name:    "secret1",
	})
	mockClient.
		On("Get", ctx, types.NamespacedName{Name: kubernetesSecretName1, Namespace: testReferenceNamespace}, &corev1.Secret{}).
		Return(k8sError.NewNotFound(corev1.Resource("secret"), kubernetesSecretName1))
	mockClient.
		On("Get", ctx, types.NamespacedName{Name: kubernetesSecretName2, Namespace: testReferenceNamespace}, &corev1.Secret{}).
		Return(k8sError.NewNotFound(corev1.Resource("secret"), kubernetesSecretName2))

	secretCache := newAlwaysMissCache(t)
	injector := NewEmbeddedSecretManagerInjector(
		config.EmbeddedSecretManagerConfig{},
		[]SecretFetcher{secretFetcherMock{
			Secrets: map[string]SecretValue{
				"u__org__o-apple__domain____project____key__secret1": {
					StringValue: "fruits @ org",
				},
				"u__org__o-apple__domain__d-cherry__project____key__secret1": {
					StringValue: "fruits @ domain",
				},
			},
		}}, mockClient, testReferenceNamespace, secretCache, &config.Config{SecretEnvVarPrefix: config.DefaultSecretEnvVarPrefix})

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

	mockClient := &mocks.MockableControllerRuntimeClient{}
	kubernetesSecretName1 := ToImagePullK8sName(SecretNameComponents{
		Org:     "o-apple",
		Domain:  "",
		Project: "",
		Name:    "secret1",
	})
	kubernetesSecretName2 := ToImagePullK8sName(SecretNameComponents{
		Org:     "o-apple",
		Domain:  "d-cherry",
		Project: "",
		Name:    "secret1",
	})
	kubernetesSecretName3 := ToImagePullK8sName(SecretNameComponents{
		Org:     "o-apple",
		Domain:  "d-cherry",
		Project: "p-banana",
		Name:    "secret1",
	})
	mockClient.
		On("Get", ctx, types.NamespacedName{Name: kubernetesSecretName1, Namespace: testReferenceNamespace}, &corev1.Secret{}).
		Return(k8sError.NewNotFound(corev1.Resource("secret"), kubernetesSecretName1))
	mockClient.
		On("Get", ctx, types.NamespacedName{Name: kubernetesSecretName2, Namespace: testReferenceNamespace}, &corev1.Secret{}).
		Return(k8sError.NewNotFound(corev1.Resource("secret"), kubernetesSecretName2))
	mockClient.
		On("Get", ctx, types.NamespacedName{Name: kubernetesSecretName3, Namespace: testReferenceNamespace}, &corev1.Secret{}).
		Return(k8sError.NewNotFound(corev1.Resource("secret"), kubernetesSecretName3))

	secretCache := newAlwaysMissCache(t)
	injector := NewEmbeddedSecretManagerInjector(
		config.EmbeddedSecretManagerConfig{},
		[]SecretFetcher{secretFetcherMock{
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
		}}, mockClient, testReferenceNamespace, secretCache, &config.Config{SecretEnvVarPrefix: config.DefaultSecretEnvVarPrefix})

	pod, injected, err := injector.Inject(ctx, secret, pod)
	assert.NoError(t, err)
	assert.True(t, injected)
	assert.True(t, podHasSecretInjected(pod, "secret1", "fruits @ project", ""))
}

// TestEmbeddedSecretManagerInjector_LookUpSecret_StaleOrgCacheDoesNotMaskProjectScope
// locks in the cache-then-fetch-per-scope fix. Previously a cached org-scoped
// value would short-circuit the entire lookup when the project+domain cache
// missed, masking a newly created project+domain secret. The fix checks the
// cache and fetches per scope so lower (more specific) scopes win when they
// exist in the backend.
func TestEmbeddedSecretManagerInjector_LookUpSecret_StaleOrgCacheDoesNotMaskProjectScope(t *testing.T) {
	ctx := context.Background()
	secret := &core.Secret{Key: "secret1"}
	pod := &corev1.Pod{
		Spec: corev1.PodSpec{Containers: []corev1.Container{{}}},
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{
				"organization": "o-apple",
				"domain":       "d-cherry",
				"project":      "p-banana",
			},
		},
	}

	projectDomainID := "u__org__o-apple__domain__d-cherry__project__p-banana__key__secret1"
	orgID := "u__org__o-apple__domain____project____key__secret1"

	// Cache mock: project+domain MISS, but a stale org-scoped entry exists.
	// The fix must *not* return the org entry when a project+domain fetch
	// would succeed.
	secretCache := cacheMocks.NewCacheInterface[SecretValue](t)
	secretCache.EXPECT().Get(mock.Anything, projectDomainID).Return(SecretValue{}, fmt.Errorf("cache miss")).Maybe()
	secretCache.EXPECT().Get(mock.Anything, orgID).Return(SecretValue{StringValue: "stale org value"}, nil).Maybe()
	secretCache.EXPECT().Set(mock.Anything, projectDomainID, mock.Anything).Return(nil).Maybe()

	// image pull mirroring path calls Get on the controller-runtime client for
	// the org-scoped image-pull Secret; return NotFound so the happy path runs.
	mockClient := &mocks.MockableControllerRuntimeClient{}
	orgImagePullName := ToImagePullK8sName(SecretNameComponents{Org: "o-apple", Name: "secret1"})
	mockClient.
		On("Get", ctx, types.NamespacedName{Name: orgImagePullName, Namespace: testReferenceNamespace}, &corev1.Secret{}).
		Return(k8sError.NewNotFound(corev1.Resource("secret"), orgImagePullName)).Maybe()

	injector := NewEmbeddedSecretManagerInjector(
		config.EmbeddedSecretManagerConfig{},
		[]SecretFetcher{secretFetcherMock{
			Secrets: map[string]SecretValue{
				projectDomainID: {StringValue: "fresh project value"},
				orgID:           {StringValue: "stale org value"},
			},
		}},
		mockClient, testReferenceNamespace, secretCache,
		&config.Config{SecretEnvVarPrefix: config.DefaultSecretEnvVarPrefix},
	)

	pod, injected, err := injector.Inject(ctx, secret, pod)
	assert.NoError(t, err)
	assert.True(t, injected)
	// Must pick up the newly-created project+domain value, not the stale cached org value.
	assert.True(t, podHasSecretInjected(pod, "secret1", "fresh project value", ""))
}

func TestEmbeddedSecretManagerInjector_InjectImagePullSecret(t *testing.T) {
	ctx = context.Background()

	testOrganization := "test-organization"
	testDomain := "test-domain"
	testProject := "test-project"
	testSecretName := "test-secret"
	testNamespace := "test-pod-namespace"

	secretName := "u__org__test-organization__domain__test-domain__project__test-project__key__test-secret" //nolint:gosec
	kubernetesSecretName := ToImagePullK8sName(SecretNameComponents{
		Org:     testOrganization,
		Domain:  testDomain,
		Project: testProject,
		Name:    testSecretName,
	})

	referenceImagePullSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      kubernetesSecretName,
			Namespace: testReferenceNamespace,
		},
		Data: map[string][]byte{
			".dockerconfigjson": []byte("test-credentials"),
		},
	}

	existingMirrorImagePullSecret := referenceImagePullSecret.DeepCopy()
	existingMirrorImagePullSecret.SetNamespace(testNamespace)

	secret := &core.Secret{Key: testSecretName}
	pod := &corev1.Pod{
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{},
			},
		},
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{
				"organization": testOrganization,
				"domain":       testDomain,
				"project":      testProject,
			},
			Namespace: testNamespace,
		},
	}

	enabledConfig := config.EmbeddedSecretManagerConfig{
		ImagePullSecrets: config.ImagePullSecretsConfig{
			Enabled: true,
		},
	}

	t.Run("existing image pull secret", func(t *testing.T) {
		testPod := pod.DeepCopy()

		mockClient := &mocks.MockableControllerRuntimeClient{}
		mockClient.
			On("Get", ctx, types.NamespacedName{Name: kubernetesSecretName, Namespace: testReferenceNamespace}, &corev1.Secret{}).
			Run(func(args mock.Arguments) {
				secret := args.Get(2).(*corev1.Secret)
				*secret = *referenceImagePullSecret
			}).
			Return(nil)

		mockClient.
			On("Get", ctx, types.NamespacedName{Name: kubernetesSecretName, Namespace: testNamespace}, &corev1.Secret{}).
			Run(func(args mock.Arguments) {
				secret := args.Get(2).(*corev1.Secret)
				*secret = *existingMirrorImagePullSecret
			}).
			Return(nil)

		secretCache := newAlwaysMissCache(t)
		injector := NewEmbeddedSecretManagerInjector(
			enabledConfig,
			[]SecretFetcher{secretFetcherMock{
				Secrets: map[string]SecretValue{
					secretName: {
						BinaryValue: []byte("test-credentials"),
					},
				},
			}}, mockClient, testReferenceNamespace, secretCache, &config.Config{SecretEnvVarPrefix: config.DefaultSecretEnvVarPrefix})

		resultPod, injected, err := injector.Inject(ctx, secret, testPod)
		assert.NoError(t, err)
		assert.True(t, injected)
		assert.Equal(t, []corev1.LocalObjectReference{{Name: kubernetesSecretName}}, resultPod.Spec.ImagePullSecrets)
	})

	t.Run("missing image pull secret", func(t *testing.T) {
		testPod := pod.DeepCopy()

		mockClient := &mocks.MockableControllerRuntimeClient{}
		mockClient.
			On("Get", ctx, types.NamespacedName{Name: kubernetesSecretName, Namespace: testReferenceNamespace}, &corev1.Secret{}).
			Run(func(args mock.Arguments) {
				secret := args.Get(2).(*corev1.Secret)
				*secret = *referenceImagePullSecret
			}).
			Return(nil)

		mockClient.
			On("Get", ctx, types.NamespacedName{Name: kubernetesSecretName, Namespace: testNamespace}, &corev1.Secret{}).
			Return(k8sError.NewNotFound(corev1.Resource("secret"), kubernetesSecretName))

		mockClient.
			On("Create", ctx, existingMirrorImagePullSecret).
			Return(nil)

		secretCache := newAlwaysMissCache(t)
		injector := NewEmbeddedSecretManagerInjector(
			enabledConfig,
			[]SecretFetcher{secretFetcherMock{
				Secrets: map[string]SecretValue{
					secretName: {
						BinaryValue: []byte("test-credentials"),
					},
				},
			}}, mockClient, testReferenceNamespace, secretCache, &config.Config{SecretEnvVarPrefix: config.DefaultSecretEnvVarPrefix})

		resultPod, injected, err := injector.Inject(ctx, secret, testPod)
		assert.NoError(t, err)
		assert.True(t, injected)
		assert.Equal(t, []corev1.LocalObjectReference{{Name: kubernetesSecretName}}, resultPod.Spec.ImagePullSecrets)
	})

	t.Run("image pull secrets disabled", func(t *testing.T) {
		testPod := pod.DeepCopy()

		mockClient := &mocks.MockableControllerRuntimeClient{}

		secretCache := newAlwaysMissCache(t)
		injector := NewEmbeddedSecretManagerInjector(
			config.EmbeddedSecretManagerConfig{},
			[]SecretFetcher{secretFetcherMock{
				Secrets: map[string]SecretValue{
					secretName: {
						BinaryValue: []byte("test-credentials"),
					},
				},
			}}, mockClient, testReferenceNamespace, secretCache, &config.Config{SecretEnvVarPrefix: config.DefaultSecretEnvVarPrefix})

		resultPod, injected, err := injector.Inject(ctx, secret, testPod)
		assert.NoError(t, err)
		assert.True(t, injected)
		assert.Nil(t, resultPod.Spec.ImagePullSecrets)

		mockClient.AssertNotCalled(t, "Get", mock.Anything, mock.Anything, mock.Anything)
		mockClient.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
	})

	t.Run("image pull secret is not scoped to project or domain", func(t *testing.T) {
		testPod := pod.DeepCopy()
		orgBasedSecretName := "u__org__test-organization__domain____project____key__test-secret" //nolint:gosec
		orgBasedKubernetesSecretName := ToImagePullK8sName(SecretNameComponents{
			Org:     testOrganization,
			Domain:  "",
			Project: "",
			Name:    testSecretName,
		})

		orgBasedReferenceImagePullSecret := referenceImagePullSecret.DeepCopy()
		orgBasedReferenceImagePullSecret.SetName(orgBasedKubernetesSecretName)
		mockClient := &mocks.MockableControllerRuntimeClient{}
		mockClient.
			On("Get", ctx, types.NamespacedName{Name: orgBasedKubernetesSecretName, Namespace: testReferenceNamespace}, &corev1.Secret{}).
			Run(func(args mock.Arguments) {
				secret := args.Get(2).(*corev1.Secret)
				*secret = *orgBasedReferenceImagePullSecret
			}).
			Return(nil)

		orgBasedExistingMirrorSecret := existingMirrorImagePullSecret.DeepCopy()
		orgBasedExistingMirrorSecret.SetName(orgBasedKubernetesSecretName)
		mockClient.
			On("Get", ctx, types.NamespacedName{Name: orgBasedKubernetesSecretName, Namespace: testNamespace}, &corev1.Secret{}).
			Run(func(args mock.Arguments) {
				secret := args.Get(2).(*corev1.Secret)
				*secret = *orgBasedExistingMirrorSecret
			}).
			Return(nil)

		secretCache := newAlwaysMissCache(t)
		injector := NewEmbeddedSecretManagerInjector(
			enabledConfig,
			[]SecretFetcher{secretFetcherMock{
				Secrets: map[string]SecretValue{
					orgBasedSecretName: {
						BinaryValue: []byte("test-credentials"),
					},
				},
			}}, mockClient, testReferenceNamespace, secretCache, &config.Config{SecretEnvVarPrefix: config.DefaultSecretEnvVarPrefix})

		resultPod, injected, err := injector.Inject(ctx, secret, testPod)
		assert.NoError(t, err)
		assert.True(t, injected)
		assert.Equal(t, []corev1.LocalObjectReference{{Name: orgBasedKubernetesSecretName}}, resultPod.Spec.ImagePullSecrets)
	})
}

func podHasSecretInjected(pod *corev1.Pod, secretKey string, secretValue string, envVar string) bool {
	// When Secret.EnvVar is set the injector emits a single value env var with
	// that name; otherwise it emits the auto-generated <prefix><KEY> name.
	expectedName := config.DefaultSecretEnvVarPrefix + strings.ToUpper(secretKey)
	if envVar != "" {
		expectedName = envVar
	}
	return lo.EveryBy(pod.Spec.Containers, func(container corev1.Container) bool {
		hasValueEnvVar := lo.ContainsBy(container.Env, func(env corev1.EnvVar) bool {
			return env.Name == expectedName && env.Value == secretValue
		})
		hasPrefixEnvVar := lo.ContainsBy(container.Env, func(env corev1.EnvVar) bool {
			return env.Name == SecretEnvVarPrefix && env.Value == config.DefaultSecretEnvVarPrefix
		})
		return hasValueEnvVar && hasPrefixEnvVar
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
